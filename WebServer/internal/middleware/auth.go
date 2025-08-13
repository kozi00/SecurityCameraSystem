package middleware

import (
	"net/http"
	"strings"
)

// AuthMiddleware validates the presence of an "authenticated" cookie for protected paths.
// It allows publicPaths without authentication and responds with 401 for API/AJAX
// requests or redirects to /login for regular requests.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		publicPaths := []string{ //endpoints that do not require authentication
			"/login",
			"/auth/login",
			"/api/camera",
			"/static/css/login.css",
		}

		for _, path := range publicPaths {
			if strings.HasPrefix(r.URL.Path, path) {
				next.ServeHTTP(w, r)
				return
			}
		}

		cookie, err := r.Cookie("authenticated")
		if err != nil || cookie.Value != "true" {
			// If this is an AJAX/API request, return 401
			if r.Header.Get("X-Requested-With") == "XMLHttpRequest" ||
				r.Header.Get("Content-Type") == "application/json" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			// For regular requests, redirect to login
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
