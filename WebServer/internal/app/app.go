package app

import (
	"fmt"
	"net/http"
	"webserver/internal/config"
	"webserver/internal/routes"
	"webserver/internal/services"
	"webserver/internal/services/ai"
	"webserver/internal/services/storage"
	"webserver/internal/services/websocket"
)

type App struct {
	config           *config.Config
	detectorServices []*ai.DetectorService
	bufferService    *storage.BufferService
	hubService       *websocket.HubService
	manager          *services.Manager
}

func NewApp() *App {
	cfg := config.Load()

	detectors := make([]*ai.DetectorService, 0, cfg.ProcessingWorkers)
	for i := 0; i < cfg.ProcessingWorkers; i++ {
		ds := ai.NewDetectorService(cfg.ModelPath, cfg.ConfigPath, cfg.MotionThreshold) // załaduj model osobno
		detectors = append(detectors, ds)
	}
	buffer := storage.NewBufferService(cfg.ImageDirectory, cfg.ImageBufferLimit)
	hub := websocket.NewHubService()

	mng := services.NewManager(detectors, buffer, hub, cfg.ProcessingWorkers, cfg.ProcessingInterval)

	return &App{
		config:           cfg,
		detectorServices: detectors,
		bufferService:    buffer,
		hubService:       hub,
		manager:          mng,
	}
}

func (a *App) Run() error {
	// Start background services
	go a.bufferService.Run(a.config.ImageBufferFlushInterval)
	go a.hubService.Run()

	// Setup routes
	router := routes.SetupRoutes(a.manager, a.config)

	fmt.Printf("🚀 Security Camera Server\n")
	fmt.Printf("📍 URL: http://localhost:%d\n", a.config.Port)
	fmt.Printf("🔑 Password: %s\n", a.config.Password)
	fmt.Printf("📁 Images: %s\n", a.config.ImageDirectory)
	fmt.Printf("🤖 AI Model: %s\n", a.config.ModelPath)

	return http.ListenAndServe(fmt.Sprintf(":%d", a.config.Port), router)
}
