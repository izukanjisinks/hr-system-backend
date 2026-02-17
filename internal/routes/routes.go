package routes

import (
	"hr-system/internal/handlers"
)

func RegisterRoutes(
	authHandler *handlers.AuthHandler,
	deptHandler *handlers.DepartmentHandler,
	posHandler *handlers.PositionHandler,
	empHandler *handlers.EmployeeHandler,
	docHandler *handlers.EmployeeDocumentHandler,
	ecHandler *handlers.EmergencyContactHandler,
) {
	RegisterPublicRoutes(authHandler)
	RegisterAuthRoutes(authHandler)
	RegisterDepartmentRoutes(deptHandler)
	RegisterPositionRoutes(posHandler)
	RegisterEmployeeRoutes(empHandler, docHandler, ecHandler)
}
