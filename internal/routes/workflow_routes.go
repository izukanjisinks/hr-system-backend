package routes

import (
	"net/http"

	"hr-system/internal/handlers"
)

func RegisterWorkflowRoutes(h *handlers.WorkflowHandler) {
	// Get my tasks (all or filtered by status)
	// Query param: ?status=pending
	http.HandleFunc("GET /api/v1/workflow/my-tasks", withAuth(h.GetMyTasks))

	// Get only pending tasks
	http.HandleFunc("GET /api/v1/workflow/my-tasks/pending", withAuth(h.GetMyPendingTasks))

	// Get specific task details with full context
	http.HandleFunc("GET /api/v1/workflow/tasks/{id}", withAuth(h.GetTaskDetails))

	// Initiate a new workflow (typically called by other services, not directly by users)
	http.HandleFunc("POST /api/v1/workflow/instances", withAuth(h.InitiateWorkflow))

	// Get workflow instance by task ID (e.g., leave request ID)
	// This needs to come before the {id} routes to avoid conflicts
	http.HandleFunc("GET /api/v1/workflow/task/{task_id}/instance", withAuth(h.GetInstanceByTaskID))

	// Process an action on a workflow instance
	// Body: {"action": "approve", "comments": "Looks good"}
	http.HandleFunc("POST /api/v1/workflow/instances/{id}/action", withAuth(h.ProcessAction))

	// Get workflow instance history
	http.HandleFunc("GET /api/v1/workflow/instances/{id}/history", withAuth(h.GetInstanceHistory))
}