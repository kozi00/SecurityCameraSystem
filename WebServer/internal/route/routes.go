package route

import (
	"net/http"
	"os"
	"path/filepath"
	"webserver/internal/config"
	"webserver/internal/handler"
	"webserver/internal/logger"
	"webserver/internal/middleware"
	"webserver/internal/repository"
	"webserver/internal/service"
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
func SetupRoutes(manager *service.Manager, cfg *config.Config, logger *logger.Logger,
	imageRepo repository.ImageRepository, detectionRepo repository.DetectionRepository) http.Handler {
	mux := http.NewServeMux()

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Start UDP camera handler in a separate goroutine
	go handler.UDPCameraHandler(manager, logger, cfg)

	// API endpoints
	mux.HandleFunc("/api/view", handler.ViewWebsocketHandler(manager, logger))
	mux.HandleFunc("/api/pictures", handler.GetPicturesFromDBHandler(manager, cfg, logger, imageRepo, detectionRepo))
	mux.HandleFunc("/api/pictures/view", handler.ViewPictureHandler(cfg))
	mux.HandleFunc("/api/pictures/clear", handler.ClearPicturesWithDBHandler(manager, cfg, logger, imageRepo))
	mux.HandleFunc("/api/pictures/delete", handler.DeletePictureHandler(manager, cfg, logger, imageRepo))

	// Log endpoints
	mux.HandleFunc("/logs/info", handler.ShowInfoLogsHandler(cfg))
	mux.HandleFunc("/logs/warning", handler.ShowWarningLogsHandler(cfg))
	mux.HandleFunc("/logs/error", handler.ShowErrorLogsHandler(cfg))

	mux.HandleFunc("/logs/info/clear", handler.ClearInfoLogsHandler(logger))
	mux.HandleFunc("/logs/warning/clear", handler.ClearWarningLogsHandler(logger))
	mux.HandleFunc("/logs/error/clear", handler.ClearErrorLogsHandler(logger))

	// Auth endpoints
	mux.HandleFunc("/auth/login", handler.LoginHandler(cfg, logger))
	mux.HandleFunc("/auth/logout", handler.LogoutHandler)

	// Automatic HTML handler mapping for example: /settings -> /static/settings.html
	mux.HandleFunc("/", dynamicHTMLHandler)

	// Apply middleware
	return middleware.AuthMiddleware(mux)
}
