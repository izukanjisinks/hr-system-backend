package routes

import (
	"net/http"

	"hr-system/internal/handlers"
	"hr-system/internal/models"
)

func RegisterPositionRoutes(h *handlers.PositionHandler) {
	http.HandleFunc("/api/v1/hr/positions",
		withAuth(h.List))

	http.HandleFunc("/api/v1/hr/positions",
		withAuthAndRole(h.Create, models.RoleSuperAdmin, models.RoleHRManager))

	http.HandleFunc("/api/v1/hr/positions/{id}",
		withAuth(h.GetByID))

	http.HandleFunc("/api/v1/hr/positions/{id}",
		withAuthAndRole(h.Update, models.RoleSuperAdmin, models.RoleHRManager))

	http.HandleFunc("/api/v1/hr/positions/{id}",
		withAuthAndRole(h.Delete, models.RoleSuperAdmin))
}
