package handlers

import (
	"encoding/json"
	"net/http"

	"hr-system/internal/models"
	"hr-system/internal/repository"
)

type WorkflowAdminHandler struct {
	workflowRepo *repository.WorkflowRepository
}

func NewWorkflowAdminHandler(workflowRepo *repository.WorkflowRepository) *WorkflowAdminHandler {
	return &WorkflowAdminHandler{
		workflowRepo: workflowRepo,
	}
}

// ========== Workflow Template Management ==========

// GetAllWorkflows retrieves all active workflow templates
func (h *WorkflowAdminHandler) GetAllWorkflows(w http.ResponseWriter, r *http.Request) {
	workflows, err := h.workflowRepo.GetAllActive()
	if err != nil {
		http.Error(w, "Failed to retrieve workflows", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"workflows": workflows,
		"count":     len(workflows),
	})
}

// GetWorkflowByID retrieves a specific workflow template
func (h *WorkflowAdminHandler) GetWorkflowByID(w http.ResponseWriter, r *http.Request) {
	workflowID := r.PathValue("id")
	if workflowID == "" {
		http.Error(w, "Workflow ID is required", http.StatusBadRequest)
		return
	}

	workflow, err := h.workflowRepo.GetByID(workflowID)
	if err != nil {
		http.Error(w, "Workflow not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workflow)
}

// CreateWorkflow creates a new workflow template
func (h *WorkflowAdminHandler) CreateWorkflow(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var workflow models.Workflow
	if err := json.NewDecoder(r.Body).Decode(&workflow); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if workflow.Name == "" {
		http.Error(w, "Workflow name is required", http.StatusBadRequest)
		return
	}

	workflow.CreatedBy = userID
	workflow.IsActive = true

	if err := h.workflowRepo.Create(&workflow); err != nil {
		http.Error(w, "Failed to create workflow", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Workflow created successfully",
		"workflow": workflow,
	})
}

// ========== Workflow Steps Management ==========

// GetWorkflowSteps retrieves all steps for a workflow
func (h *WorkflowAdminHandler) GetWorkflowSteps(w http.ResponseWriter, r *http.Request) {
	workflowID := r.PathValue("id")
	if workflowID == "" {
		http.Error(w, "Workflow ID is required", http.StatusBadRequest)
		return
	}

	steps, err := h.workflowRepo.GetStepsByWorkflowID(workflowID)
	if err != nil {
		http.Error(w, "Failed to retrieve steps", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"steps": steps,
		"count": len(steps),
	})
}

// GetStepByID retrieves a specific workflow step
func (h *WorkflowAdminHandler) GetStepByID(w http.ResponseWriter, r *http.Request) {
	stepID := r.PathValue("step_id")
	if stepID == "" {
		http.Error(w, "Step ID is required", http.StatusBadRequest)
		return
	}

	step, err := h.workflowRepo.GetStepByID(stepID)
	if err != nil {
		http.Error(w, "Step not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(step)
}

// CreateWorkflowStepRequest represents the request body for creating a step
type CreateWorkflowStepRequest struct {
	WorkflowID           string   `json:"workflow_id"`
	StepName             string   `json:"step_name"`
	StepOrder            int      `json:"step_order"`
	Initial              bool     `json:"initial"`
	Final                bool     `json:"final"`
	AllowedRoles         []string `json:"allowed_roles"`
	RequiresAllApprovers bool     `json:"requires_all_approvers"`
	MinApprovals         int      `json:"min_approvals"`
}

// CreateWorkflowStep creates a new workflow step
func (h *WorkflowAdminHandler) CreateWorkflowStep(w http.ResponseWriter, r *http.Request) {
	var req CreateWorkflowStepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.WorkflowID == "" || req.StepName == "" {
		http.Error(w, "Workflow ID and step name are required", http.StatusBadRequest)
		return
	}

	// Note: You'll need to implement CreateStep in WorkflowRepository
	// For now, return a placeholder response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Step creation endpoint - implement CreateStep in repository",
		"request": req,
	})
}

// ========== Workflow Transitions Management ==========

// GetWorkflowTransitions retrieves all transitions for a workflow
func (h *WorkflowAdminHandler) GetWorkflowTransitions(w http.ResponseWriter, r *http.Request) {
	workflowID := r.PathValue("id")
	if workflowID == "" {
		http.Error(w, "Workflow ID is required", http.StatusBadRequest)
		return
	}

	transitions, err := h.workflowRepo.GetTransitionsByWorkflowID(workflowID)
	if err != nil {
		http.Error(w, "Failed to retrieve transitions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"transitions": transitions,
		"count":       len(transitions),
	})
}

// GetValidTransitions retrieves valid transitions from a specific step
func (h *WorkflowAdminHandler) GetValidTransitions(w http.ResponseWriter, r *http.Request) {
	stepID := r.PathValue("step_id")
	if stepID == "" {
		http.Error(w, "Step ID is required", http.StatusBadRequest)
		return
	}

	transitions, err := h.workflowRepo.GetValidTransitions(stepID)
	if err != nil {
		http.Error(w, "Failed to retrieve transitions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"transitions": transitions,
		"count":       len(transitions),
	})
}

// CreateWorkflowTransitionRequest represents the request body for creating a transition
type CreateWorkflowTransitionRequest struct {
	WorkflowID     string `json:"workflow_id"`
	FromStepID     string `json:"from_step_id"`
	ToStepID       string `json:"to_step_id"`
	ActionName     string `json:"action_name"`
	ConditionType  string `json:"condition_type"`
	ConditionValue string `json:"condition_value"`
}

// CreateWorkflowTransition creates a new workflow transition
func (h *WorkflowAdminHandler) CreateWorkflowTransition(w http.ResponseWriter, r *http.Request) {
	var req CreateWorkflowTransitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.WorkflowID == "" || req.FromStepID == "" || req.ToStepID == "" || req.ActionName == "" {
		http.Error(w, "Workflow ID, from step, to step, and action name are required", http.StatusBadRequest)
		return
	}

	// Note: You'll need to implement CreateTransition in WorkflowRepository
	// For now, return a placeholder response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Transition creation endpoint - implement CreateTransition in repository",
		"request": req,
	})
}

// ========== Workflow Structure Overview ==========

// GetWorkflowStructure retrieves complete workflow structure (steps + transitions)
func (h *WorkflowAdminHandler) GetWorkflowStructure(w http.ResponseWriter, r *http.Request) {
	workflowID := r.PathValue("id")
	if workflowID == "" {
		http.Error(w, "Workflow ID is required", http.StatusBadRequest)
		return
	}

	// Get workflow
	workflow, err := h.workflowRepo.GetByID(workflowID)
	if err != nil {
		http.Error(w, "Workflow not found", http.StatusNotFound)
		return
	}

	// Get steps
	steps, err := h.workflowRepo.GetStepsByWorkflowID(workflowID)
	if err != nil {
		http.Error(w, "Failed to retrieve steps", http.StatusInternalServerError)
		return
	}

	// Get transitions
	transitions, err := h.workflowRepo.GetTransitionsByWorkflowID(workflowID)
	if err != nil {
		http.Error(w, "Failed to retrieve transitions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"workflow":    workflow,
		"steps":       steps,
		"transitions": transitions,
	})
}