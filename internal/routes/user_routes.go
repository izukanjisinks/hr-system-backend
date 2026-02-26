package routes

import (
	"net/http"

	"hr-system/internal/handlers"
	"hr-system/internal/models"
)

func RegisterUserRoutes(h *handlers.UserHandler) {
	// Get current user's profile - any authenticated user
	http.HandleFunc("GET /api/v1/users/profile",
		withAuth(h.GetProfile))

	// List all users - requires SuperAdmin or HRManager
	http.HandleFunc("GET /api/v1/users",
		withAuthAndRole(h.GetAll, models.RoleSuperAdmin, models.RoleManager, models.RoleHRManager))

	// Get user by ID - requires SuperAdmin or HRManager
	http.HandleFunc("GET /api/v1/users/{id}",
		withAuthAndRole(h.GetByID, models.RoleSuperAdmin, models.RoleManager, models.RoleHRManager))
}
