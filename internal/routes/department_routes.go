package routes

import (
	"net/http"

	"hr-system/internal/handlers"
	"hr-system/internal/models"
)

func RegisterDepartmentRoutes(h *handlers.DepartmentHandler) {
	http.HandleFunc("/api/v1/hr/departments/tree",
		withAuth(h.GetTree))

	http.HandleFunc("/api/v1/hr/departments",
		withAuth(h.List))

	http.HandleFunc("/api/v1/hr/departments",
		withAuthAndRole(h.Create, models.RoleSuperAdmin, models.RoleHRManager))

	http.HandleFunc("/api/v1/hr/departments/{id}",
		withAuth(h.GetByID))

	http.HandleFunc("/api/v1/hr/departments/{id}",
		withAuthAndRole(h.Update, models.RoleSuperAdmin, models.RoleHRManager))

	http.HandleFunc("/api/v1/hr/departments/{id}",
		withAuthAndRole(h.Delete, models.RoleSuperAdmin))
}