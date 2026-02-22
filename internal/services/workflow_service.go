package services

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"hr-system/internal/models"
	"hr-system/internal/repository"
)

type WorkflowService struct {
	workflowRepo  *repository.WorkflowRepository
	instanceRepo  *repository.WorkflowInstanceRepository
	taskRepo      *repository.AssignedTaskRepository
	historyRepo   *repository.WorkflowHistoryRepository
	userRepo      *repository.UserRepository
}

func NewWorkflowService(
	workflowRepo *repository.WorkflowRepository,
	instanceRepo *repository.WorkflowInstanceRepository,
	taskRepo *repository.AssignedTaskRepository,
	historyRepo *repository.WorkflowHistoryRepository,
	userRepo *repository.UserRepository,
) *WorkflowService {
	return &WorkflowService{
		workflowRepo:  workflowRepo,
		instanceRepo:  instanceRepo,
		taskRepo:      taskRepo,
		historyRepo:   historyRepo,
		userRepo:      userRepo,
	}
}

// InitiateWorkflow starts a new workflow instance
func (s *WorkflowService) InitiateWorkflow(
	workflowName string,
	taskDetails models.TaskDetails,
	initiatorID string,
	priority string,
	dueDate *time.Time,
) (*models.WorkflowInstance, error) {
	// Get workflow template
	workflow, err := s.workflowRepo.GetByName(workflowName)
	if err != nil {
		return nil, fmt.Errorf("workflow not found: %w", err)
	}

	// Get initial step
	initialStep, err := s.workflowRepo.GetInitialStep(workflow.ID)
	if err != nil {
		return nil, fmt.Errorf("initial step not found: %w", err)
	}

	// Get transitions from initial step
	transitions, err := s.workflowRepo.GetValidTransitions(initialStep.ID)
	if err != nil {
		return nil, fmt.Errorf("no valid transitions found: %w", err)
	}

	// Find the "submit" transition
	var nextStepID string
	for _, tr := range transitions {
		if tr.ActionName == "submit" {
			nextStepID = tr.ToStepID
			break
		}
	}

	if nextStepID == "" {
		return nil, errors.New("no submit transition found from initial step")
	}

	// Create workflow instance
	instance := &models.WorkflowInstance{
		WorkflowID:    workflow.ID,
		CurrentStepID: nextStepID,
		Status:        "in_progress",
		TaskDetails:   taskDetails,
		CreatedBy:     initiatorID,
		Priority:      priority,
		DueDate:       dueDate,
	}

	if err := s.instanceRepo.Create(instance); err != nil {
		return nil, fmt.Errorf("failed to create instance: %w", err)
	}

	// Get the next step details
	nextStep, err := s.workflowRepo.GetStepByID(nextStepID)
	if err != nil {
		return nil, fmt.Errorf("failed to get next step: %w", err)
	}

	// Determine who to assign the task to
	assigneeID, err := s.determineAssignee(nextStep, taskDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to determine assignee: %w", err)
	}

	// Create assigned task
	task := &models.AssignedTask{
		InstanceID: instance.ID,
		StepID:     nextStep.ID,
		StepName:   nextStep.StepName,
		AssignedTo: assigneeID,
		AssignedBy: initiatorID,
		Status:     "pending",
		DueDate:    dueDate,
	}

	if err := s.taskRepo.Create(task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Get initiator name for history
	initiator, err := s.userRepo.GetByID(initiatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get initiator: %w", err)
	}

	// Create history entry
	history := &models.WorkflowHistory{
		InstanceID:      instance.ID,
		FromStepID:      &initialStep.ID,
		ToStepID:        nextStepID,
		ActionTaken:     "submit",
		PerformedBy:     initiatorID,
		PerformedByName: initiator.FirstName + " " + initiator.LastName,
		Comments:        fmt.Sprintf("Initiated %s workflow", workflowName),
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
	performer, err := s.userRepo.GetByID(performedByID)
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
func (s *WorkflowService) determineAssignee(step *models.WorkflowStep, taskDetails models.TaskDetails) (string, error) {
	// For now, find the first user with one of the allowed roles
	// In a production system, you might want more sophisticated logic
	// (e.g., department-based assignment, round-robin, etc.)

	if len(step.AllowedRoles) == 0 {
		return "", errors.New("no allowed roles defined for step")
	}

	// Get a user with one of the allowed roles
	// This is a simplified version - you may want to enhance this logic
	for _, roleCode := range step.AllowedRoles {
		// Try to find a user with this role
		// For now, we'll return the first HR Manager or Department Head
		// You should implement proper user selection logic here
		if roleCode == "HR_MANAGER" || roleCode == "DEPARTMENT_HEAD" {
			// This is a placeholder - implement proper user lookup
			// For example, you could look up the employee's department head
			// or the assigned HR manager from the task details
			return taskDetails.SenderDetails.SenderID, nil // Temporary
		}
	}

	return "", errors.New("no suitable assignee found for step")
}

// Helper: Check if user has permission to perform action on a step
func (s *WorkflowService) checkPermission(userID string, step *models.WorkflowStep) error {
	// Get user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Get user's role
	role, err := s.userRepo.GetRoleByUserID(userID)
	if err != nil {
		return fmt.Errorf("user role not found: %w", err)
	}

	// Check if user's role is in allowed roles
	for _, allowedRole := range step.AllowedRoles {
		if role.Code == allowedRole {
			return nil
		}
	}

	return fmt.Errorf("user %s does not have permission to perform this action", user.Email)
}