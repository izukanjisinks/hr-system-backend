package handlers

import (
	"net/http"

	"hr-system/internal/middleware"
	"hr-system/internal/services"
	"hr-system/pkg/utils"
)

type DashboardHandler struct {
	service *services.DashboardService
}

func NewDashboardHandler(svc *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{service: svc}
}

func (h *DashboardHandler) GetMyDashboard(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		utils.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		utils.RespondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	stats, err := h.service.GetEmployeeDashboard(userID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to load dashboard")
		return
	}
	if stats == nil {
		utils.RespondError(w, http.StatusNotFound, "No employee record linked to your account")
		return
	}

	utils.RespondJSON(w, http.StatusOK, stats)
}
