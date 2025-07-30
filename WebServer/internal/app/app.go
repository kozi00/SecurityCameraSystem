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
	config          *config.Config
	detectorService *ai.DetectorService
	bufferService   *storage.BufferService
	hubService      *websocket.HubService
	manager         *services.Manager
}

func NewApp() *App {
	cfg := config.Load()

	detector := ai.NewDetectorService(cfg.ModelPath, cfg.ConfigPath, cfg.MotionThreshold)
	buffer := storage.NewBufferService(cfg.ImageDirectory, cfg.ImageBufferLimit)
	hub := websocket.NewHubService()

	mng := services.NewManager(detector, buffer, hub)

	return &App{
		config:          cfg,
		detectorService: detector,
		bufferService:   buffer,
		hubService:      hub,
		manager:         mng,
	}
}

func (a *App) Run() error {
	// Start background services
	go a.bufferService.Run(a.config.ImageBufferFlushInterval)
	go a.hubService.Run()

	// Setup routes
	router := routes.SetupRoutes(a.manager)

	fmt.Printf("🚀 Security Camera Server\n")
	fmt.Printf("📍 URL: http://localhost:%d\n", a.config.Port)
	fmt.Printf("🔑 Password: %s\n", a.config.Password)
	fmt.Printf("📁 Images: %s\n", a.config.ImageDirectory)
	fmt.Printf("🤖 AI Model: %s\n", a.config.ModelPath)

	return http.ListenAndServe(fmt.Sprintf(":%d", a.config.Port), router)
}
