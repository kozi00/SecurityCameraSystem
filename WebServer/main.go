package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"webserver/internal/handlers"
	"webserver/internal/middleware"
)

func dynamicHTMLHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// jeśli root "/", to zmapuj na "index.html"
	if path == "/" {
		path = "/index"
	}

	// Dodaj .html
	filePath := filepath.Join("static", path+".html")

	// Sprawdź, czy plik istnieje
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	// Serwuj plik
	http.ServeFile(w, r, filePath)
}

func main() {
	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	// ✅ API ENDPOINTS (logika w Go)
	mux.HandleFunc("/api/view", handlers.ViewWebsocketHandler)
	mux.HandleFunc("/api/camera", handlers.CameraWebsocketHandler)
	mux.HandleFunc("/api/pictures", handlers.DisplayPicturesHandler)
	mux.HandleFunc("/api/pictures/view", handlers.ViewPictureHandler)
	//mux.HandleFunc("/api/settings", handlers.SettingsHandler)

	mux.HandleFunc("/auth/login", handlers.LoginHandler)
	mux.HandleFunc("/auth/logout", handlers.LogoutHandler)

	mux.HandleFunc("/", dynamicHTMLHandler)

	// Owiń mux w AuthMiddleware
	handler := middleware.AuthMiddleware(mux)

	fmt.Println("Server running on http://localhost:8080")
	fmt.Println("Access requires login with password: sienkiewicza2")
	http.ListenAndServe("0.0.0.0:8080", handler)
}
