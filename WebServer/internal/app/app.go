package app

import (
	"fmt"
	"net/http"
	"webserver/internal/config"
	"webserver/internal/logger"
	"webserver/internal/routes"
	"webserver/internal/services"
	"webserver/internal/services/ai"
	"webserver/internal/services/storage"
	"webserver/internal/services/websocket"
)

// App wires core components (config, logger, services) and runs the HTTP server.
type App struct {
	config           *config.Config
	logger           *logger.Logger
	detectorServices []*ai.DetectorService
	bufferService    *storage.BufferService
	hubService       *websocket.HubService
	manager          *services.Manager
}

// NewApp constructs the application, initializing all services and dependencies.
// It pre-allocates detector workers according to the configured ProcessingWorkers.
func NewApp() *App {
	cfg := config.Load()
	logger := logger.NewLogger(cfg)

	detectors := make([]*ai.DetectorService, 0, cfg.ProcessingWorkers)
	for i := 0; i < cfg.ProcessingWorkers; i++ { //creating a few detector services to handle image processing asynchronously
		ds := ai.NewDetectorService(cfg, logger)
		detectors = append(detectors, ds)
	}
	buffer := storage.NewBufferService(cfg, logger)
	hub := websocket.NewHubService(cfg, logger)

	mng := services.NewManager(detectors, buffer, hub, cfg, logger)

	return &App{
		config:           cfg,
		detectorServices: detectors,
		bufferService:    buffer,
		hubService:       hub,
		manager:          mng,
		logger:           logger,
	}
}

// Run starts background services, sets up routes and blocks serving HTTP.
// Returns any error produced by http.ListenAndServe.
func (a *App) Run() error {
	// Start background services
	go a.bufferService.Run()
	go a.hubService.Run()

	// Setup routes
	router := routes.SetupRoutes(a.manager, a.config, a.logger)

	a.logger.Info("🚀 Security Camera Server\n")
	a.logger.Info("📍 URL: http://localhost:%d\n", a.config.Port)
	a.logger.Info("🔑 Password: %s\n", a.config.Password)
	a.logger.Info("📁 Images: %s\n", a.config.ImageDirectory)
	a.logger.Info("🤖 AI Model: %s\n", a.config.ModelPath)

	return http.ListenAndServe(fmt.Sprintf(":%d", a.config.Port), router)
}
