package handlers

import (
	"net/http"
)

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Usuń cookie uwierzytelnienia
	http.SetCookie(w, &http.Cookie{
		Name:   "authenticated",
		Value:  "",
		Path:   "/",
		MaxAge: -1, // Usuwa cookie
	})

	// Przekieruj na stronę logowania
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
