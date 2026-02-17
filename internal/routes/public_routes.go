package routes

import (
	"net/http"

	"hr-system/internal/handlers"
)

func RegisterPublicRoutes(authHandler *handlers.AuthHandler) {
	http.HandleFunc("/health", withPublic(handlers.HealthHandler))
	http.HandleFunc("/api/v1/auth/login", withPublic(authHandler.Login))
}
