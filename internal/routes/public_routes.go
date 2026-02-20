package routes

import (
	"net/http"

	"hr-system/internal/handlers"
)

func RegisterPublicRoutes(authHandler *handlers.AuthHandler) {
	// Health endpoint
	http.HandleFunc("GET /health", withPublic(handlers.HealthHandler))

	// Login endpoint
	http.HandleFunc("POST /api/v1/auth/login", withPublic(authHandler.Login))
}
