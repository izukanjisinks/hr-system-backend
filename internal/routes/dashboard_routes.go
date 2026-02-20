package routes

import (
	"net/http"

	"hr-system/internal/handlers"
)

func RegisterDashboardRoutes(h *handlers.DashboardHandler) {
	// Register without method prefix to allow OPTIONS through middleware
	http.HandleFunc("/api/v1/hr/dashboard/me", withAuth(h.GetMyDashboard))
}
