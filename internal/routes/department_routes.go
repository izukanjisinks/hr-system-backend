package routes

import (
	"net/http"
	"strings"

	"hr-system/internal/handlers"
	"hr-system/internal/models"
)

func RegisterDepartmentRoutes(h *handlers.DepartmentHandler) {
	// GET /api/v1/hr/departments/tree  — must be registered before /:id
	http.HandleFunc("/api/v1/hr/departments/tree",
		withAuth(h.GetTree))

	// Collection routes
	http.HandleFunc("/api/v1/hr/departments", withAuth(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.List(w, r)
		case http.MethodPost:
			withAuthAndRole(h.Create, models.RoleSuperAdmin, models.RoleHRManager)(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Item routes: /api/v1/hr/departments/:id
	http.HandleFunc("/api/v1/hr/departments/", withAuth(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// Skip tree (handled above)
		if strings.HasSuffix(path, "/tree") {
			return
		}

		switch r.Method {
		case http.MethodGet:
			h.GetByID(w, r)
		case http.MethodPut:
			withAuthAndRole(h.Update, models.RoleSuperAdmin, models.RoleHRManager)(w, r)
		case http.MethodDelete:
			withAuthAndRole(h.Delete, models.RoleSuperAdmin)(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
}
