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

// dynamicHTMLHandler serves /path as /static/path.html if the file exists; otherwise 404.
func dynamicHTMLHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/" {
		path = "/index"
	}

	filePath := filepath.Join("static", path+".html")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, filePath)
}

// SetupRoutes registers HTTP routes, static file serving, API endpoints,
// and wraps the mux with the authentication middleware.
func SetupRoutes(manager *services.Manager, cfg *config.Config, logger *logger.Logger) http.Handler {
	mux := http.NewServeMux()

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Start UDP camera handler in a separate goroutine
	go handlers.UDPCameraHandler(manager, logger, cfg)

	// API endpoints
	mux.HandleFunc("/api/view", handlers.ViewWebsocketHandler(manager, logger))
	mux.HandleFunc("/api/pictures", handlers.GetPicturesFromDBHandler(manager, cfg, logger))
	mux.HandleFunc("/api/pictures/view", handlers.ViewPictureHandler(cfg))
	mux.HandleFunc("/api/pictures/clear", handlers.ClearPicturesWithDBHandler(manager, cfg, logger))
	mux.HandleFunc("/api/pictures/delete", handlers.DeletePictureHandler(manager, cfg, logger))
	mux.HandleFunc("/api/pictures/filters", handlers.GetFiltersHandler(manager, logger))
	mux.HandleFunc("/api/pictures/stats", handlers.GetStatsHandler(manager, logger))

	// Log endpoints
	mux.HandleFunc("/logs/info", handlers.ShowInfoLogsHandler(cfg))
	mux.HandleFunc("/logs/warning", handlers.ShowWarningLogsHandler(cfg))
	mux.HandleFunc("/logs/error", handlers.ShowErrorLogsHandler(cfg))

	mux.HandleFunc("/logs/info/clear", handlers.ClearInfoLogsHandler(logger))
	mux.HandleFunc("/logs/warning/clear", handlers.ClearWarningLogsHandler(logger))
	mux.HandleFunc("/logs/error/clear", handlers.ClearErrorLogsHandler(logger))

	// Auth endpoints
	mux.HandleFunc("/auth/login", handlers.LoginHandler(cfg, logger))
	mux.HandleFunc("/auth/logout", handlers.LogoutHandler)

	// Automatic HTML handler mapping for example: /settings -> /static/settings.html
	mux.HandleFunc("/", dynamicHTMLHandler)

	// Apply middleware
	return middleware.AuthMiddleware(mux)
}
