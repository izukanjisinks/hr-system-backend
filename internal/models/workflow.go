package models

import "time"

// Workflow represents a workflow template
type Workflow struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// WorkflowStep represents a step/stage in a workflow
type WorkflowStep struct {
	ID           string    `json:"id"`
	WorkflowID   string    `json:"workflow_id"`
	StepName     string    `json:"step_name"`
	StepOrder    int       `json:"step_order"`
	Initial      bool      `json:"initial"`
	Final        bool      `json:"final"`
	AllowedRoles []string  `json:"allowed_roles"` // Will be stored as JSON in DB
	CreatedAt    time.Time `json:"created_at"`
}

// WorkflowTransition represents a transition between steps
type WorkflowTransition struct {
	ID             string    `json:"id"`
	WorkflowID     string    `json:"workflow_id"`
	FromStepID     string    `json:"from_step_id"`
	ToStepID       string    `json:"to_step_id"`
	ActionName     string    `json:"action_name"`     //review, approve, reject
	ConditionType  string    `json:"condition_type"`  // e.g., "user_role", "assigned_user_only"
	ConditionValue string    `json:"condition_value"` // JSON for complex conditions
	CreatedAt      time.Time `json:"created_at"`
}

//this basically refers to an instance in the workflow
type AssignedTodo struct {
	ID            string    `json:"id"`
	WorkflowId    string    `json:"workflow_id"`
	CurrentStepId string    `json:"current_step_id"`
	TodoId        string    `json:"todo_id"`
	AssignedTo    string    `json:"assigned_to"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// WorkflowHistory represents the audit trail of a workflow instance
type WorkflowHistory struct {
	ID          string    `json:"id"`
	InstanceID  string    `json:"instance_id"`
	FromStepID  *string   `json:"from_step_id"` // Nullable for initial creation
	ToStepID    string    `json:"to_step_id"`
	ActionTaken string    `json:"action_taken"`
	PerformedBy string    `json:"performed_by"`
	Comments    string    `json:"comments"`
	Timestamp   time.Time `json:"timestamp"`
}

// AvailableAction represents an action a user can take on an instance
type AvailableAction struct {
	ActionName   string `json:"action_name"`
	ToStepName   string `json:"to_step_name"`
	TransitionID string `json:"transition_id"`
}

// WorkflowInstanceWithDetails includes current step information
type WorkflowInstanceWithDetails struct {
	AssignedTodo
	CurrentStepName  string            `json:"current_step_name"`
	WorkflowName     string            `json:"workflow_name"`
	AvailableActions []AvailableAction `json:"available_actions"`
}
