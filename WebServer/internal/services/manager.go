package services

import (
	"encoding/base64"
	"fmt"
	"log"
	"webserver/internal/services/ai"
	"webserver/internal/services/storage"
	wshub "webserver/internal/services/websocket"
)

type Manager struct {
	bufferService    *storage.BufferService
	detectorService  *ai.DetectorService
	websocketService *wshub.HubService
}

func NewManager(detectorService *ai.DetectorService, bufferService *storage.BufferService, websocketService *wshub.HubService) *Manager {
	return &Manager{
		detectorService:  detectorService,
		bufferService:    bufferService,
		websocketService: websocketService,
	}
}

func (m *Manager) HandleCameraImage(image []byte, camera string) {
	// Process and send to viewers
	m.SendToViewers(image, camera)

	detectionsJSON := m.ProcessImage(image, camera)
	if detectionsJSON != "[]" {
		m.bufferService.SaveImage(image, camera)
	}
}

func (m *Manager) ProcessImage(image []byte, camera string) string {
	if m.detectorService == nil {
		return "[]"
	}

	motionDetected, err := m.detectorService.DetectMotion(image)
	if err != nil {
		log.Printf("Błąd rozpoznawania ruchu: %v", err)
		return "[]"
	}

	if !motionDetected {
		return "[]"
	}

	detections, err := m.detectorService.DetectObjects(image)
	if err != nil {
		log.Printf("Błąd detekcji obiektów: %v", err)
		return "[]"
	}

	detectionsJSON, err := m.detectorService.FormatDetectionsAsJSON(detections)
	if err != nil {
		log.Printf("Błąd formatowania detekcji do JSON: %v", err)
		return "[]"
	}

	return detectionsJSON
}

func (m *Manager) SendToViewers(image []byte, camera string) {
	detectionsJSON := m.ProcessImage(image, camera)

	encoded := base64.StdEncoding.EncodeToString(image)
	msg := fmt.Sprintf(`{"camera":"%s","image":"%s", "detections":%s}`,
		camera, encoded, detectionsJSON)

	m.websocketService.Broadcast([]byte(msg), camera)
}

func (m *Manager) GetWebsocketService() *wshub.HubService {
	return m.websocketService
}
func (m *Manager) GetBufferService() *storage.BufferService {
	return m.bufferService
}
func (m *Manager) GetDetectorService() *ai.DetectorService {
	return m.detectorService
}
