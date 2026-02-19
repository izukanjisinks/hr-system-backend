package routes

import (
	"net/http"

	"hr-system/internal/handlers"
)

func RegisterDashboardRoutes(h *handlers.DashboardHandler) {
	http.HandleFunc("GET /api/v1/hr/dashboard/me", withAuth(h.GetMyDashboard))
}
