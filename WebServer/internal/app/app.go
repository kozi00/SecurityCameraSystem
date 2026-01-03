package app

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"webserver/internal/config"
	"webserver/internal/logger"
	"webserver/internal/repository"
	"webserver/internal/repository/sqlite"
	"webserver/internal/route"
	"webserver/internal/service"
	"webserver/internal/service/ai"
	"webserver/internal/service/storage"
	"webserver/internal/service/websocket"
)

// App wires core components (config, logger, services) and runs the HTTP server.
type App struct {
	config           *config.Config
	logger           *logger.Logger
	detectorServices []*ai.DetectorService
	bufferService    *storage.BufferService
	hubService       *websocket.HubService
	manager          *service.Manager
	db               *sqlite.DB
	imageRepo        repository.ImageRepository
	detectionRepo    repository.DetectionRepository
}

// NewApp constructs the application, initializing all services and dependencies.
// It pre-allocates detector workers according to the configured ProcessingWorkers.
func NewApp() *App {
	cfg := config.Load()
	logger := logger.NewLogger(cfg)

	// Ensure database directory exists
	dbDir := filepath.Dir(cfg.DatabasePath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		logger.Error("Failed to create database directory: %v", err)
	}

	// Initialize database and repositories
	var db *sqlite.DB
	var imageRepo repository.ImageRepository
	var detectionRepo repository.DetectionRepository

	db, err := sqlite.New(cfg.DatabasePath)
	if err != nil {
		logger.Error("Failed to initialize database: %v", err)
		db = nil
	} else {
		logger.Info("📦 Database initialized at %s", cfg.DatabasePath)
		imageRepo = sqlite.NewImageRepository(db)
		detectionRepo = sqlite.NewDetectionRepository(db)
	}

	detectors := make([]*ai.DetectorService, 0, cfg.ProcessingWorkers)
	for i := 0; i < cfg.ProcessingWorkers; i++ {
		ds := ai.NewDetectorService(cfg, logger)
		detectors = append(detectors, ds)
	}
	buffer := storage.NewBufferService(cfg, logger, imageRepo, detectionRepo)
	hub := websocket.NewHubService(cfg, logger)

	mng := service.NewManager(detectors, buffer, hub, cfg, logger)

	return &App{
		config:           cfg,
		detectorServices: detectors,
		bufferService:    buffer,
		hubService:       hub,
		manager:          mng,
		logger:           logger,
		db:               db,
		imageRepo:        imageRepo,
		detectionRepo:    detectionRepo,
	}
}

// Run starts background services, sets up routes and blocks serving HTTP.
// Returns any error produced by http.ListenAndServe.
func (a *App) Run() error {
	// Close database on exit
	if a.db != nil {
		defer a.db.Close()
	}

	// Start background services
	go a.bufferService.Run()
	go a.hubService.Run()

	// Setup routes
	router := route.SetupRoutes(a.manager, a.config, a.logger, a.imageRepo, a.detectionRepo)

	a.logger.Info("🚀 Security Camera Server\n")
	a.logger.Info("📍 URL: http://localhost:%d\n", a.config.Port)
	a.logger.Info("🔑 Password: %s\n", a.config.Password)
	a.logger.Info("📁 Images: %s\n", a.config.ImageDirectory)
	a.logger.Info("🤖 AI Model: %s\n", a.config.ModelPath)
	a.logger.Info("📦 Database: %s\n", a.config.DatabasePath)

	return http.ListenAndServe(fmt.Sprintf(":%d", a.config.Port), router)
}
