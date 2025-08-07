package services

import (
	"encoding/base64"
	"fmt"
	"sync"
	"webserver/internal/config"
	"webserver/internal/logger"
	"webserver/internal/services/ai"
	"webserver/internal/services/storage"
	"webserver/internal/services/websocket"
)

type Manager struct {
	bufferService    *storage.BufferService
	detectorServices []*ai.DetectorService
	websocketService *websocket.HubService
	logger           *logger.Logger

	processingQueue chan ImageProcessingTask
	frameCounters   map[string]int // Licznik klatek dla każdej kamery
	processEveryNth int            // Przetwarzaj co N-tą klatkę
	numWorkers      int

	frameCounterMu sync.Mutex // Mutex do ochrony frameCounters
	wg             sync.WaitGroup
}

type ImageProcessingTask struct {
	Image  []byte
	Camera string
}

func NewManager(detectorServices []*ai.DetectorService, bufferService *storage.BufferService, websocketService *websocket.HubService, config *config.Config, logger *logger.Logger) *Manager {
	manager := &Manager{
		detectorServices: detectorServices,
		bufferService:    bufferService,
		websocketService: websocketService,
		numWorkers:       config.ProcessingWorkers,            // Liczba workerów do przetwarzania obrazów
		processingQueue:  make(chan ImageProcessingTask, 100), // Buffer dla 100 zadań
		frameCounters:    make(map[string]int),                // Liczniki klatek dla każdej kamery
		processEveryNth:  config.ProcessingInterval,           // Przetwarzaj co N-tą klatkę
		logger:           logger,
	}

	for i := 0; i < manager.numWorkers; i++ {
		manager.wg.Add(1)
		go manager.processingWorker(i)
	}

	manager.logger.Info("🎬 Manager started - processing every %d frame(s)", manager.processEveryNth)
	return manager
}

func (m *Manager) HandleCameraImage(image []byte, camera string) {
	m.SendToViewers(image, camera)

	m.frameCounterMu.Lock()
	m.frameCounters[camera]++
	frameCount := m.frameCounters[camera]
	m.frameCounterMu.Unlock()

	// 🎯 Przetwarzaj tylko co N-tą klatkę
	if frameCount%m.processEveryNth != 0 {
		return
	}
	m.ResetFrameCounter(camera)

	motionDetected, err := m.detectorServices[0].DetectMotion(image, camera)

	if err != nil {
		m.logger.Error("Error detecting motion: %v", err)
		return
	}

	if !motionDetected {
		return
	}

	select {
	case m.processingQueue <- ImageProcessingTask{Image: image, Camera: camera}:
		m.logger.Info("📹 Camera %s: Frame queued for processing", camera)
	default:
		m.logger.Warning("⚠️  Processing queue full for camera %s - skipping AI detection", camera)
	}
}

func (m *Manager) SendToViewers(image []byte, camera string) {

	encoded := base64.StdEncoding.EncodeToString(image)
	msg := fmt.Sprintf(`{"camera":"%s","image":"%s"}`,
		camera, encoded)

	m.websocketService.Broadcast([]byte(msg), camera)
}

func (m *Manager) GetWebsocketService() *websocket.HubService {
	return m.websocketService
}
func (m *Manager) GetBufferService() *storage.BufferService {
	return m.bufferService
}
func (m *Manager) GetDetectorService() []*ai.DetectorService {
	return m.detectorServices
}

// processingWorker przetwarza obrazy w osobnym wątku
func (m *Manager) processingWorker(workerID int) {
	defer m.wg.Done()

	m.logger.Info("🔧 Processing worker %d started", workerID)

	for task := range m.processingQueue {
		m.processImageAsync(task.Image, task.Camera, workerID)
	}

	m.logger.Info("🔧 Processing worker %d stopped", workerID)
}

// processImageAsync przetwarza obraz asynchronicznie
func (m *Manager) processImageAsync(image []byte, camera string, workerID int) {

	detections, err := m.detectorServices[workerID].DetectObjects(image)
	if err != nil {
		m.logger.Error("Błąd detekcji obiektów: %v", err)
		return
	}

	if len(detections) > 0 {
		// Narysuj prostokąty na obrazie
		imageWithDetections, err := m.detectorServices[workerID].DrawRectangle(detections, image)
		if err != nil {
			m.logger.Error("Failed to draw rectangles: %v", err)
			imageWithDetections = image // Użyj oryginalnego obrazu
		}

		m.bufferService.AddImage(imageWithDetections, camera, detections[0].Label)
	}
}

// Stop zatrzymuje wszystkie workery
func (m *Manager) Stop() {
	close(m.processingQueue)
	m.wg.Wait()
	m.logger.Info("🛑 All processing workers stopped")
}

func (m *Manager) ResetFrameCounter(cameraId string) {
	m.frameCounterMu.Lock()
	m.frameCounters[cameraId] = 0
	m.frameCounterMu.Unlock()
}
