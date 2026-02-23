package services

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"hr-system/internal/models"
	"hr-system/internal/repository"
)

type WorkflowService struct {
	workflowRepo  *repository.WorkflowRepository
	instanceRepo  *repository.WorkflowInstanceRepository
	taskRepo      *repository.AssignedTaskRepository
	historyRepo   *repository.WorkflowHistoryRepository
	userRepo      *repository.UserRepository
	employeeRepo  *repository.EmployeeRepository
}

func NewWorkflowService(
	workflowRepo *repository.WorkflowRepository,
	instanceRepo *repository.WorkflowInstanceRepository,
	taskRepo *repository.AssignedTaskRepository,
	historyRepo *repository.WorkflowHistoryRepository,
	userRepo *repository.UserRepository,
	employeeRepo *repository.EmployeeRepository,
) *WorkflowService {
	return &WorkflowService{
		workflowRepo:  workflowRepo,
		instanceRepo:  instanceRepo,
		taskRepo:      taskRepo,
		historyRepo:   historyRepo,
		userRepo:      userRepo,
		employeeRepo:  employeeRepo,
	}
}

// InitiateWorkflow starts a new workflow instance using workflow type
func (s *WorkflowService) InitiateWorkflow(
	workflowType models.WorkflowType,
	taskDetails models.TaskDetails,
	initiatorID string,
	priority string,
	dueDate *time.Time,
) (*models.WorkflowInstance, error) {
	// Get workflow template by type
	workflow, err := s.workflowRepo.GetByType(workflowType)
	if err != nil {
		return nil, fmt.Errorf("workflow not found for type %s: %w", workflowType, err)
	}

	// Get the first action step (the step after submission, e.g., "Pending Review")
	// This also returns the initial step ID for history tracking
	// This is optimized to do in one query instead of: get initial -> get transitions -> find submit -> get next step
	firstActionStep, initialStepID, err := s.workflowRepo.GetFirstActionStep(workflow.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get first action step: %w", err)
	}

	// Create workflow instance at the first action step (not the initial step)
	// The instance starts at "Pending Review" or whatever the first action step is
	instance := &models.WorkflowInstance{
		WorkflowID:    workflow.ID,
		CurrentStepID: firstActionStep.ID,
		Status:        "in_progress",
		TaskDetails:   taskDetails,
		CreatedBy:     initiatorID,
		Priority:      priority,
		DueDate:       dueDate,
	}

	if err := s.instanceRepo.Create(instance); err != nil {
		return nil, fmt.Errorf("failed to create instance: %w", err)
	}

	// Determine who to assign the task to (for the first action step)
	assigneeID, err := s.determineAssignee(firstActionStep, taskDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to determine assignee: %w", err)
	}

	// Create assigned task for the first action step
	task := &models.AssignedTask{
		InstanceID: instance.ID,
		StepID:     firstActionStep.ID,
		StepName:   firstActionStep.StepName,
		AssignedTo: assigneeID,
		AssignedBy: initiatorID,
		Status:     "pending",
		DueDate:    dueDate,
	}

	if err := s.taskRepo.Create(task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Get initiator name for history
	initiatorUUID, err := uuid.Parse(initiatorID)
	if err != nil {
		return nil, fmt.Errorf("invalid initiator ID: %w", err)
	}

	initiator, err := s.employeeRepo.GetByUserID(initiatorUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get initiator: %w", err)
	}

	// Create history entry (from initial step -> first action step via "submit")
	history := &models.WorkflowHistory{
		InstanceID:      instance.ID,
		FromStepID:      &initialStepID,
		ToStepID:        firstActionStep.ID,
		ActionTaken:     "submit",
		PerformedBy:     initiatorID,
		PerformedByName: initiator.FirstName + " " + initiator.LastName,
		Comments:        fmt.Sprintf("Initiated %s workflow", workflow.Name),
	}

	if err := s.historyRepo.Create(history); err != nil {
		return nil, fmt.Errorf("failed to create history: %w", err)
	}

	return instance, nil
}

// ProcessAction processes an action on a workflow instance
func (s *WorkflowService) ProcessAction(
	instanceID string,
	action string,
	performedByID string,
	comments string,
) error {
	// Get the workflow instance
	instance, err := s.instanceRepo.GetByID(instanceID)
	if err != nil {
		return fmt.Errorf("instance not found: %w", err)
	}

	// Check if instance is still in progress
	if instance.Status == "completed" || instance.Status == "cancelled" {
		return errors.New("workflow instance is already closed")
	}

	// Get current step
	currentStep, err := s.workflowRepo.GetStepByID(instance.CurrentStepID)
	if err != nil {
		return fmt.Errorf("current step not found: %w", err)
	}

	// Check if user has permission to perform this action
	if err := s.checkPermission(performedByID, currentStep); err != nil {
		return err
	}

	// Get valid transition
	transition, err := s.workflowRepo.GetTransitionByAction(currentStep.ID, action)
	if err == sql.ErrNoRows {
		return fmt.Errorf("action '%s' is not valid from current step", action)
	}
	if err != nil {
		return fmt.Errorf("failed to get transition: %w", err)
	}

	// Get next step
	nextStep, err := s.workflowRepo.GetStepByID(transition.ToStepID)
	if err != nil {
		return fmt.Errorf("next step not found: %w", err)
	}

	// Update instance
	newStatus := "in_progress"
	if nextStep.Final {
		newStatus = "completed"
	}

	if err := s.instanceRepo.UpdateStep(instanceID, nextStep.ID, newStatus); err != nil {
		return fmt.Errorf("failed to update instance: %w", err)
	}

	// Complete current task
	activeTask, err := s.taskRepo.GetActiveTaskForInstance(instanceID)
	if err == nil && activeTask != nil {
		if err := s.taskRepo.Complete(activeTask.ID); err != nil {
			return fmt.Errorf("failed to complete task: %w", err)
		}
	}

	// Create new task if not final step
	if !nextStep.Final {
		assigneeID, err := s.determineAssignee(nextStep, instance.TaskDetails)
		if err != nil {
			return fmt.Errorf("failed to determine assignee: %w", err)
		}

		newTask := &models.AssignedTask{
			InstanceID: instanceID,
			StepID:     nextStep.ID,
			StepName:   nextStep.StepName,
			AssignedTo: assigneeID,
			AssignedBy: performedByID,
			Status:     "pending",
			DueDate:    instance.DueDate,
		}

		if err := s.taskRepo.Create(newTask); err != nil {
			return fmt.Errorf("failed to create new task: %w", err)
		}
	} else {
		// Mark instance as completed
		if err := s.instanceRepo.Complete(instanceID); err != nil {
			return fmt.Errorf("failed to complete instance: %w", err)
		}
	}

	// Get performer name for history
	performerUUID, err := uuid.Parse(performedByID)
	if err != nil {
		return fmt.Errorf("invalid performer ID: %w", err)
	}

	performer, err := s.employeeRepo.GetByUserID(performerUUID)
	if err != nil {
		return fmt.Errorf("failed to get performer: %w", err)
	}

	// Create history entry
	history := &models.WorkflowHistory{
		InstanceID:      instanceID,
		FromStepID:      &currentStep.ID,
		ToStepID:        nextStep.ID,
		ActionTaken:     action,
		PerformedBy:     performedByID,
		PerformedByName: performer.FirstName + " " + performer.LastName,
		Comments:        comments,
	}

	if err := s.historyRepo.Create(history); err != nil {
		return fmt.Errorf("failed to create history: %w", err)
	}

	return nil
}

// GetMyTasks retrieves all tasks assigned to a user
func (s *WorkflowService) GetMyTasks(userID string, statusFilter string) ([]models.AssignedTask, error) {
	if statusFilter != "" {
		return s.taskRepo.GetByAssignee(userID, statusFilter)
	}
	return s.taskRepo.GetByAssignee(userID)
}

// GetInstanceHistory retrieves the complete history of a workflow instance
func (s *WorkflowService) GetInstanceHistory(instanceID string) ([]models.WorkflowHistory, error) {
	return s.historyRepo.GetByInstanceID(instanceID)
}

// GetInstanceByTaskID retrieves a workflow instance by the associated task ID
func (s *WorkflowService) GetInstanceByTaskID(taskID string) (*models.WorkflowInstance, error) {
	return s.instanceRepo.GetByTaskID(taskID)
}

// Helper: Determine who to assign the task to based on step configuration
// Uses intelligent load balancing - assigns to the user with the required role who has the fewest pending tasks
func (s *WorkflowService) determineAssignee(step *models.WorkflowStep, taskDetails models.TaskDetails) (string, error) {
	if len(step.AllowedRoles) == 0 {
		return "", errors.New("no allowed roles defined for step")
	}

	// Try each allowed role and find the user with the fewest pending tasks
	for _, roleName := range step.AllowedRoles {
		user, err := s.userRepo.GetUserWithFewestTasksByRole(roleName)
		if err == nil && user != nil {
			// Found a user with this role - return their user_id as string
			return user.UserID.String(), nil
		}
	}

	return "", fmt.Errorf("no active user found with any of the allowed roles: %v", step.AllowedRoles)
}

// Helper: Check if user has permission to perform action on a step
func (s *WorkflowService) checkPermission(userID string, step *models.WorkflowStep) error {
	// Get user
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	user, err := s.userRepo.GetUserByID(userUUID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Check if user has a role
	if user.Role == nil {
		return fmt.Errorf("user %s has no role assigned", user.Email)
	}

	// Check if user's role is in allowed roles
	for _, allowedRole := range step.AllowedRoles {
		if user.Role.Name == allowedRole {
			return nil
		}
	}

	return fmt.Errorf("user %s does not have permission to perform this action", user.Email)
}