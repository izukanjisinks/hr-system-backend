package routes

import (
	"net/http"
	"strings"

	"hr-system/internal/handlers"
	"hr-system/internal/models"
)

func RegisterEmployeeRoutes(
	empH *handlers.EmployeeHandler,
	docH *handlers.EmployeeDocumentHandler,
	ecH *handlers.EmergencyContactHandler,
) {
	// /api/v1/hr/employees
	http.HandleFunc("/api/v1/hr/employees", withAuth(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			withAuthAndRole(empH.List, models.RoleSuperAdmin, models.RoleHRManager, models.RoleManager)(w, r)
		case http.MethodPost:
			withAuthAndRole(empH.Create, models.RoleSuperAdmin, models.RoleHRManager)(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// /api/v1/hr/employees/  (sub-routes)
	http.HandleFunc("/api/v1/hr/employees/", withAuth(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		switch {
		case strings.HasSuffix(path, "/direct-reports"):
			empH.GetDirectReports(w, r)

		case strings.Contains(path, "/documents/") && strings.HasSuffix(path, "/verify"):
			if r.Method == http.MethodPut {
				withAuthAndRole(docH.Verify, models.RoleSuperAdmin, models.RoleHRManager)(w, r)
			}

		case strings.HasSuffix(path, "/documents"):
			switch r.Method {
			case http.MethodGet:
				docH.ListByEmployee(w, r)
			case http.MethodPost:
				docH.Create(w, r)
			}

		case strings.HasSuffix(path, "/emergency-contacts"):
			switch r.Method {
			case http.MethodGet:
				ecH.ListByEmployee(w, r)
			case http.MethodPost:
				ecH.Create(w, r)
			}

		default:
			// /api/v1/hr/employees/:id
			switch r.Method {
			case http.MethodGet:
				empH.GetByID(w, r)
			case http.MethodPut:
				withAuthAndRole(empH.Update, models.RoleSuperAdmin, models.RoleHRManager)(w, r)
			case http.MethodDelete:
				withAuthAndRole(empH.Delete, models.RoleSuperAdmin)(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}
	}))
}
