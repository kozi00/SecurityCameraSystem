package handlers

import (
	"net/http"
)

// LogoutHandler clears the authentication cookie and redirects to the login page.
func LogoutHandler(w http.ResponseWriter, r *http.Request) {

	http.SetCookie(w, &http.Cookie{
		Name:   "authenticated",
		Value:  "",
		Path:   "/",
		MaxAge: -1, //Deleting cookie
	})

	// Redirect to login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
