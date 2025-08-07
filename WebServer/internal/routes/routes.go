package routes

import (
	"net/http"
	"os"
	"path/filepath"
	"webserver/internal/config"
	"webserver/internal/handlers"
	"webserver/internal/logger"
	"webserver/internal/middleware"
	"webserver/internal/services"
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

func SetupRoutes(manager *services.Manager, cfg *config.Config, logger *logger.Logger) http.Handler {
	mux := http.NewServeMux()

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// API endpoints - używaj Manager bezpośrednio
	mux.HandleFunc("/api/view", handlers.ViewWebsocketHandler(manager, logger))
	mux.HandleFunc("/api/camera", handlers.CameraWebsocketHandler(manager, logger))
	mux.HandleFunc("/api/pictures", handlers.DisplayPicturesHandler(cfg, logger))
	mux.HandleFunc("/api/pictures/view", handlers.ViewPictureHandler(cfg))

	// Auth endpoints
	mux.HandleFunc("/auth/login", handlers.LoginHandler)
	mux.HandleFunc("/auth/logout", handlers.LogoutHandler)

	// Dynamic HTML handler
	mux.HandleFunc("/", dynamicHTMLHandler)

	// Apply middleware
	return middleware.AuthMiddleware(mux)
}
