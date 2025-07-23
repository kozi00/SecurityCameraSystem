package middleware

import (
	"net/http"
	"strings"
)

// AuthMiddleware sprawdza, czy użytkownik jest zalogowany (ma cookie 'authenticated=true')
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Pozwól na dostęp do strony logowania, api kamery,  i zasobów statycznych bez uwierzytelnienia
		if r.URL.Path == "/login" ||
			r.URL.Path == "/Login.html" ||
			strings.HasPrefix(r.URL.Path, "/css/") ||
			strings.HasPrefix(r.URL.Path, "/js/") ||
			strings.HasPrefix(r.URL.Path, "/camera") {
			next.ServeHTTP(w, r)
			return
		}

		// Sprawdź czy użytkownik jest zalogowany
		cookie, err := r.Cookie("authenticated")
		if err != nil || cookie.Value != "true" {
			// Jeśli to zapytanie AJAX/API, zwróć 401
			if r.Header.Get("X-Requested-With") == "XMLHttpRequest" ||
				r.Header.Get("Content-Type") == "application/json" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			// Dla zwykłych żądań przekieruj na login
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
