package handlers

import (
	"net/http"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	password := r.FormValue("password")
	if password != "sienkiewicza2" { // Replace with your actual password check
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}
	// Ustaw cookie po poprawnym logowaniu
	http.SetCookie(w, &http.Cookie{
		Name:  "authenticated",
		Value: "true",
		Path:  "/",
		// Secure: true, // odkomentuj jeśli używasz HTTPS
		HttpOnly: true,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
