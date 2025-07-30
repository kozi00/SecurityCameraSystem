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
	detections := m.ProcessImage(image, camera)

	if detections == nil {
		//log.Printf("No detections for camera %s", camera)
		m.SendToViewers(image, camera)
	} else {
		imageDetected, err := m.detectorService.DrawRectangle(detections, image)
		if err != nil {
			log.Printf("Failed to draw rectangles on image: %v", err)
		}
		m.SendToViewers(imageDetected, camera)
		m.bufferService.AddImage(imageDetected, camera, detections[0].Label)
		log.Printf("Processed image for camera %s with detections: %v", camera, detections)
	}
}

func (m *Manager) ProcessImage(image []byte, camera string) []ai.DetectionResult {
	if m.detectorService == nil {
		return nil
	}

	motionDetected, err := m.detectorService.DetectMotion(image)
	if err != nil {
		log.Printf("Błąd rozpoznawania ruchu: %v", err)
		return nil
	}

	if !motionDetected {
		return nil
	}

	detections, err := m.detectorService.DetectObjects(image)
	if err != nil {
		log.Printf("Błąd detekcji obiektów: %v", err)
		return nil
	}

	return detections
}

func (m *Manager) SendToViewers(image []byte, camera string) {

	encoded := base64.StdEncoding.EncodeToString(image)
	msg := fmt.Sprintf(`{"camera":"%s","image":"%s"}`,
		camera, encoded)

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
