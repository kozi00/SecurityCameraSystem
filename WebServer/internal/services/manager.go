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

const (
	// ProcessingQueueSize is the capacity of the image processing channel.
	ProcessingQueueSize = 50
	// MotionDetectionWorkerId selects which worker performs motion detection for gating.
	MotionDetectionWorkerId = 0
)

// Manager orchestrates camera frame handling, motion gating, AI detection,
// buffering to disk, and broadcasting frames to viewers.
type Manager struct {
	bufferService    *storage.BufferService
	detectorServices []*ai.DetectorService
	websocketService *websocket.HubService
	logger           *logger.Logger

	processingQueue chan ImageProcessingTask
	frameCounters   map[string]int
	numWorkers      int

	frameCounterMu sync.Mutex
	wg             sync.WaitGroup
}

type ImageProcessingTask struct {
	Image  []byte
	Camera string
}

// NewManager constructs a Manager and starts processing worker goroutines.
func NewManager(detectorServices []*ai.DetectorService, bufferService *storage.BufferService, websocketService *websocket.HubService, config *config.Config, logger *logger.Logger) *Manager {
	manager := &Manager{
		detectorServices: detectorServices,
		bufferService:    bufferService,
		websocketService: websocketService,
		numWorkers:       config.ProcessingWorkers,
		processingQueue:  make(chan ImageProcessingTask, ProcessingQueueSize),
		frameCounters:    make(map[string]int),
		logger:           logger,
	}

	for i := 0; i < manager.numWorkers; i++ {
		manager.wg.Add(1)
		go manager.processingWorker(i)
	}

	manager.logger.Info("🎬 Manager started")
	return manager
}

// HandleCameraImage broadcasts to viewers (if any), checks motion gating,
// and enqueues the frame for detection when appropriate.
func (m *Manager) HandleCameraImage(image []byte, camera string) {
	if m.websocketService.GetClientCount() > 0 {
		m.sendToViewers(image, camera)
	}

	motionDetected, err := m.detectorServices[MotionDetectionWorkerId].DetectMotion(image, camera)
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

// GetWebsocketService returns the HubService responsible for viewer connections.
func (m *Manager) GetWebsocketService() *websocket.HubService {
	return m.websocketService
}

// GetBufferService returns the BufferService used for saving images.
func (m *Manager) GetBufferService() *storage.BufferService {
	return m.bufferService
}

// GetDetectorService returns the list of DetectorService workers.
func (m *Manager) GetDetectorService() []*ai.DetectorService {
	return m.detectorServices
}

// Stop gracefully stops workers by closing the queue and waiting for completion.
func (m *Manager) Stop() {
	close(m.processingQueue)
	m.wg.Wait()
	m.logger.Info("🛑 All processing workers stopped")
}

// sendToViewers encodes the frame as base64 within a small JSON and broadcasts it.
func (m *Manager) sendToViewers(image []byte, camera string) {

	encoded := base64.StdEncoding.EncodeToString(image)
	msg := fmt.Sprintf(`{"camera":"%s","image":"%s"}`,
		camera, encoded)

	m.websocketService.Broadcast([]byte(msg), camera)
}

// processingWorker consumes tasks from the queue and calls processImageAsync.
func (m *Manager) processingWorker(workerID int) {
	defer m.wg.Done()

	m.logger.Info("🔧 Processing worker %d started", workerID)

	for task := range m.processingQueue {
		m.processImageAsync(task.Image, task.Camera, workerID)
	}

	m.logger.Info("🔧 Processing worker %d stopped", workerID)
}

// processImageAsync performs object detection, draws annotations, and buffers the result.
func (m *Manager) processImageAsync(image []byte, camera string, workerID int) {

	detections, err := m.detectorServices[workerID].DetectObjects(image)
	if err != nil {
		m.logger.Error("Błąd detekcji obiektów: %v", err)
		return
	}

	if len(detections) > 0 {
		imageWithDetections, err := m.detectorServices[workerID].DrawRectangle(detections, image)
		if err != nil {
			m.logger.Error("Failed to draw rectangles: %v", err)
			imageWithDetections = image
		}

		m.bufferService.AddImage(imageWithDetections, camera, detections[0].Label)
	}
}
