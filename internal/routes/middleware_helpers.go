package routes

import (
	"net/http"

	"hr-system/internal/middleware"
)

func withAuth(handler http.HandlerFunc) http.HandlerFunc {
	return middleware.CORS(func(w http.ResponseWriter, r *http.Request) {
		middleware.JWTAuth(http.HandlerFunc(handler)).ServeHTTP(w, r)
	})
}

func withAuthAndRole(handler http.HandlerFunc, roles ...string) http.HandlerFunc {
	return middleware.CORS(func(w http.ResponseWriter, r *http.Request) {
		middleware.JWTAuth(
			middleware.RequireAnyRole(roles...)(http.HandlerFunc(handler)),
		).ServeHTTP(w, r)
	})
}

func withPublic(handler http.HandlerFunc) http.HandlerFunc {
	return middleware.CORS(handler)
}
