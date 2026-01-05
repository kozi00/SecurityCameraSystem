package handler

import (
	"net/http"
	"webserver/internal/config"
	"webserver/internal/logger"
)

// LoginHandler handles POST /auth/login by validating password and issuing an auth cookie.
func LoginHandler(config *config.Config, logger *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		password := r.FormValue("password")
		if password != config.Password {
			http.Error(w, "Invalid password", http.StatusUnauthorized)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "authenticated",
			Value:    "true",
			Path:     "/",
			MaxAge:   2592000, // 30 days
			HttpOnly: true,
		})
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
