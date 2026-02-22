package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"hr-system/internal/database"
	"hr-system/internal/models"

	"github.com/google/uuid"
)

type WorkflowRepository struct {
	db *sql.DB
}

func NewWorkflowRepository() *WorkflowRepository {
	return &WorkflowRepository{
		db: database.DB,
	}
}

// GetByID retrieves a workflow template by ID
func (r *WorkflowRepository) GetByID(id string) (*models.Workflow, error) {
	query := `
		SELECT id, name, description, is_active, created_by, created_at, updated_at
		FROM workflows
		WHERE id = $1
	`

	var wf models.Workflow
	err := r.db.QueryRow(query, id).Scan(
		&wf.ID,
		&wf.Name,
		&wf.Description,
		&wf.IsActive,
		&wf.CreatedBy,
		&wf.CreatedAt,
		&wf.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &wf, nil
}

// GetByName retrieves a workflow template by name
func (r *WorkflowRepository) GetByName(name string) (*models.Workflow, error) {
	query := `
		SELECT id, name, description, is_active, created_by, created_at, updated_at
		FROM workflows
		WHERE name = $1 AND is_active = true
	`

	var wf models.Workflow
	err := r.db.QueryRow(query, name).Scan(
		&wf.ID,
		&wf.Name,
		&wf.Description,
		&wf.IsActive,
		&wf.CreatedBy,
		&wf.CreatedAt,
		&wf.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &wf, nil
}

// GetAllActive retrieves all active workflow templates
func (r *WorkflowRepository) GetAllActive() ([]models.Workflow, error) {
	query := `
		SELECT id, name, description, is_active, created_by, created_at, updated_at
		FROM workflows
		WHERE is_active = true
		ORDER BY name
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workflows []models.Workflow
	for rows.Next() {
		var wf models.Workflow
		err := rows.Scan(
			&wf.ID,
			&wf.Name,
			&wf.Description,
			&wf.IsActive,
			&wf.CreatedBy,
			&wf.CreatedAt,
			&wf.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		workflows = append(workflows, wf)
	}

	return workflows, nil
}

// GetStepsByWorkflowID retrieves all steps for a workflow
func (r *WorkflowRepository) GetStepsByWorkflowID(workflowID string) ([]models.WorkflowStep, error) {
	query := `
		SELECT id, workflow_id, step_name, step_order, initial, final,
		       allowed_roles, requires_all_approvers, min_approvals, created_at
		FROM workflow_steps
		WHERE workflow_id = $1
		ORDER BY step_order
	`

	rows, err := r.db.Query(query, workflowID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []models.WorkflowStep
	for rows.Next() {
		var step models.WorkflowStep
		var allowedRolesJSON []byte

		err := rows.Scan(
			&step.ID,
			&step.WorkflowID,
			&step.StepName,
			&step.StepOrder,
			&step.Initial,
			&step.Final,
			&allowedRolesJSON,
			&step.RequiresAllApprovers,
			&step.MinApprovals,
			&step.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse allowed_roles JSON
		if err := json.Unmarshal(allowedRolesJSON, &step.AllowedRoles); err != nil {
			return nil, fmt.Errorf("failed to parse allowed_roles: %w", err)
		}

		steps = append(steps, step)
	}

	return steps, nil
}

// GetStepByID retrieves a specific workflow step
func (r *WorkflowRepository) GetStepByID(stepID string) (*models.WorkflowStep, error) {
	query := `
		SELECT id, workflow_id, step_name, step_order, initial, final,
		       allowed_roles, requires_all_approvers, min_approvals, created_at
		FROM workflow_steps
		WHERE id = $1
	`

	var step models.WorkflowStep
	var allowedRolesJSON []byte

	err := r.db.QueryRow(query, stepID).Scan(
		&step.ID,
		&step.WorkflowID,
		&step.StepName,
		&step.StepOrder,
		&step.Initial,
		&step.Final,
		&allowedRolesJSON,
		&step.RequiresAllApprovers,
		&step.MinApprovals,
		&step.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse allowed_roles JSON
	if err := json.Unmarshal(allowedRolesJSON, &step.AllowedRoles); err != nil {
		return nil, fmt.Errorf("failed to parse allowed_roles: %w", err)
	}

	return &step, nil
}

// GetInitialStep retrieves the initial step of a workflow
func (r *WorkflowRepository) GetInitialStep(workflowID string) (*models.WorkflowStep, error) {
	query := `
		SELECT id, workflow_id, step_name, step_order, initial, final,
		       allowed_roles, requires_all_approvers, min_approvals, created_at
		FROM workflow_steps
		WHERE workflow_id = $1 AND initial = true
		LIMIT 1
	`

	var step models.WorkflowStep
	var allowedRolesJSON []byte

	err := r.db.QueryRow(query, workflowID).Scan(
		&step.ID,
		&step.WorkflowID,
		&step.StepName,
		&step.StepOrder,
		&step.Initial,
		&step.Final,
		&allowedRolesJSON,
		&step.RequiresAllApprovers,
		&step.MinApprovals,
		&step.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse allowed_roles JSON
	if err := json.Unmarshal(allowedRolesJSON, &step.AllowedRoles); err != nil {
		return nil, fmt.Errorf("failed to parse allowed_roles: %w", err)
	}

	return &step, nil
}

// GetTransitionsByWorkflowID retrieves all transitions for a workflow
func (r *WorkflowRepository) GetTransitionsByWorkflowID(workflowID string) ([]models.WorkflowTransition, error) {
	query := `
		SELECT id, workflow_id, from_step_id, to_step_id, action_name,
		       condition_type, condition_value, created_at
		FROM workflow_transitions
		WHERE workflow_id = $1
	`

	rows, err := r.db.Query(query, workflowID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transitions []models.WorkflowTransition
	for rows.Next() {
		var tr models.WorkflowTransition
		err := rows.Scan(
			&tr.ID,
			&tr.WorkflowID,
			&tr.FromStepID,
			&tr.ToStepID,
			&tr.ActionName,
			&tr.ConditionType,
			&tr.ConditionValue,
			&tr.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		transitions = append(transitions, tr)
	}

	return transitions, nil
}

// GetValidTransitions retrieves valid transitions from a specific step
func (r *WorkflowRepository) GetValidTransitions(fromStepID string) ([]models.WorkflowTransition, error) {
	query := `
		SELECT id, workflow_id, from_step_id, to_step_id, action_name,
		       condition_type, condition_value, created_at
		FROM workflow_transitions
		WHERE from_step_id = $1
	`

	rows, err := r.db.Query(query, fromStepID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transitions []models.WorkflowTransition
	for rows.Next() {
		var tr models.WorkflowTransition
		err := rows.Scan(
			&tr.ID,
			&tr.WorkflowID,
			&tr.FromStepID,
			&tr.ToStepID,
			&tr.ActionName,
			&tr.ConditionType,
			&tr.ConditionValue,
			&tr.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		transitions = append(transitions, tr)
	}

	return transitions, nil
}

// GetTransitionByAction retrieves a specific transition based on action
func (r *WorkflowRepository) GetTransitionByAction(fromStepID, action string) (*models.WorkflowTransition, error) {
	query := `
		SELECT id, workflow_id, from_step_id, to_step_id, action_name,
		       condition_type, condition_value, created_at
		FROM workflow_transitions
		WHERE from_step_id = $1 AND action_name = $2
		LIMIT 1
	`

	var tr models.WorkflowTransition
	err := r.db.QueryRow(query, fromStepID, action).Scan(
		&tr.ID,
		&tr.WorkflowID,
		&tr.FromStepID,
		&tr.ToStepID,
		&tr.ActionName,
		&tr.ConditionType,
		&tr.ConditionValue,
		&tr.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &tr, nil
}

// Create creates a new workflow template
func (r *WorkflowRepository) Create(workflow *models.Workflow) error {
	query := `
		INSERT INTO workflows (id, name, description, is_active, created_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at
	`

	workflow.ID = uuid.New().String()

	return r.db.QueryRow(
		query,
		workflow.ID,
		workflow.Name,
		workflow.Description,
		workflow.IsActive,
		workflow.CreatedBy,
	).Scan(&workflow.CreatedAt, &workflow.UpdatedAt)
}