package services

import (
	"encoding/base64"
	"fmt"
	"log"
	"sync"
	"webserver/internal/services/ai"
	"webserver/internal/services/storage"
	"webserver/internal/services/websocket"
)

type Manager struct {
	bufferService    *storage.BufferService
	detectorService  *ai.DetectorService
	websocketService *websocket.HubService
	processingQueue  chan ImageProcessingTask
	numWorkers       int
	wg               sync.WaitGroup
	frameCounters    map[string]int // Licznik klatek dla każdej kamery
	frameCounterMu   sync.Mutex     // Mutex do ochrony frameCounters
	processEveryNth  int            // Przetwarzaj co N-tą klatkę
}

type ImageProcessingTask struct {
	Image  []byte
	Camera string
}

func NewManager(detectorService *ai.DetectorService, bufferService *storage.BufferService, websocketService *websocket.HubService, numWorkers int, processEveryNth int) *Manager {
	manager := &Manager{
		detectorService:  detectorService,
		bufferService:    bufferService,
		websocketService: websocketService,
		numWorkers:       numWorkers,                          // Liczba workerów do przetwarzania obrazów
		processingQueue:  make(chan ImageProcessingTask, 100), // Buffer dla 100 zadań
		frameCounters:    make(map[string]int),                // Liczniki klatek dla każdej kamery
		processEveryNth:  processEveryNth,                     // Przetwarzaj co N-tą klatkę
	}

	for i := 0; i < manager.numWorkers; i++ {
		manager.wg.Add(1)
		go manager.processingWorker(i)
	}

	log.Printf("🎬 Manager started - processing every %d frame(s)", manager.processEveryNth)
	return manager
}

func (m *Manager) HandleCameraImage(image []byte, camera string) {
	// 🚀 SZYBKIE: Natychmiast wyślij obraz do widzów (bez opóźnień)
	m.SendToViewers(image, camera)

	m.frameCounterMu.Lock()
	m.frameCounters[camera]++
	frameCount := m.frameCounters[camera]
	m.frameCounterMu.Unlock()

	// 🎯 Przetwarzaj tylko co N-tą klatkę
	if frameCount%m.processEveryNth != 0 {
		return // Pomijamy tę klatkę
	}
	m.ResetFrameCounter(camera)
	motionDetected, err := m.detectorService.DetectMotion(image, camera)
	if err != nil {
		log.Printf("Błąd rozpoznawania ruchu: %v", err)
		return
	}

	if !motionDetected {
		return
	}

	select {
	case m.processingQueue <- ImageProcessingTask{Image: image, Camera: camera}:
		log.Printf("📹 Camera %s: Frame %d queued for processing", camera, frameCount)
	default:
		log.Printf("⚠️  Processing queue full for camera %s (frame %d) - skipping AI detection", camera, frameCount)
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
func (m *Manager) GetDetectorService() *ai.DetectorService {
	return m.detectorService
}

// processingWorker przetwarza obrazy w osobnym wątku
func (m *Manager) processingWorker(workerID int) {
	defer m.wg.Done()

	log.Printf("🔧 Processing worker %d started", workerID)

	for task := range m.processingQueue {
		m.processImageAsync(task.Image, task.Camera, workerID)
	}

	log.Printf("🔧 Processing worker %d stopped", workerID)
}

// processImageAsync przetwarza obraz asynchronicznie
func (m *Manager) processImageAsync(image []byte, camera string, workerID int) {

	detections, err := m.detectorService.DetectObjects(image)
	if err != nil {
		log.Printf("Błąd detekcji obiektów: %v", err)
		return
	}

	if len(detections) > 0 {
		// Narysuj prostokąty na obrazie
		imageWithDetections, err := m.detectorService.DrawRectangle(detections, image)
		if err != nil {
			log.Printf("⚠️  Worker %d: Failed to draw rectangles: %v", workerID, err)
			imageWithDetections = image // Użyj oryginalnego obrazu
		}

		m.bufferService.AddImage(imageWithDetections, camera, detections[0].Label)
	}
}

// Stop zatrzymuje wszystkie workery
func (m *Manager) Stop() {
	close(m.processingQueue)
	m.wg.Wait()
	log.Printf("🛑 All processing workers stopped")
}

func (m *Manager) ResetFrameCounter(cameraId string) {
	m.frameCounterMu.Lock()
	m.frameCounters[cameraId] = 0
	m.frameCounterMu.Unlock()
	log.Printf("🔄 Frame counter reset for camera: %s", cameraId)
}
