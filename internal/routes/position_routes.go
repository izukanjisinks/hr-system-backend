package routes

import (
	"net/http"

	"hr-system/internal/handlers"
	"hr-system/internal/models"
)

func RegisterPositionRoutes(h *handlers.PositionHandler) {
	http.HandleFunc("/api/v1/hr/positions", withAuth(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.List(w, r)
		case http.MethodPost:
			withAuthAndRole(h.Create, models.RoleSuperAdmin, models.RoleHRManager)(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	http.HandleFunc("/api/v1/hr/positions/", withAuth(func(w http.ResponseWriter, r *http.Request) {
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
