package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"time"
)

// DetectionResult reprezentuje wynik detekcji obiektu
type DetectionResult struct {
	Label      string  `json:"label"`
	Confidence float64 `json:"confidence"`
	X          int     `json:"x"`
	Y          int     `json:"y"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
}

// ObjectDetectionService serwis do rozpoznawania obiektów
type ObjectDetectionService struct {
	enabled bool
	// Tutaj można dodać modele AI
}

// NewObjectDetectionService tworzy nowy serwis detekcji
func NewObjectDetectionService() *ObjectDetectionService {
	return &ObjectDetectionService{
		enabled: true,
	}
}

// DetectObjects wykrywa obiekty na obrazie
func (ods *ObjectDetectionService) DetectObjects(imageData []byte) ([]DetectionResult, error) {
	if !ods.enabled {
		return nil, nil
	}

	// Dekoduj obraz
	img, err := ods.decodeImage(imageData)
	if err != nil {
		return nil, fmt.Errorf("błąd dekodowania obrazu: %v", err)
	}

	// Tutaj można dodać prawdziwą detekcję z OpenCV lub AI
	// Na razie symulujemy detekcję
	detections := ods.simulateDetection(img)

	log.Printf("Wykryto %d obiektów na obrazie", len(detections))
	return detections, nil
}

// decodeImage dekoduje dane obrazu do image.Image
func (ods *ObjectDetectionService) decodeImage(data []byte) (image.Image, error) {
	reader := bytes.NewReader(data)
	img, err := jpeg.Decode(reader)
	if err != nil {
		return nil, err
	}
	return img, nil
}

// simulateDetection symuluje detekcję obiektów (do testów)
func (ods *ObjectDetectionService) simulateDetection(img image.Image) []DetectionResult {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Symuluj wykrycie osoby w centrum obrazu
	detections := []DetectionResult{
		{
			Label:      "person",
			Confidence: 0.85,
			X:          width / 4,
			Y:          height / 4,
			Width:      width / 2,
			Height:     height / 2,
		},
	}

	// Czasami wykryj samochód
	if time.Now().Second()%10 == 0 {
		detections = append(detections, DetectionResult{
			Label:      "car",
			Confidence: 0.72,
			X:          10,
			Y:          height - 100,
			Width:      200,
			Height:     80,
		})
	}

	return detections
}

// AnalyzeImageForMotion analizuje obraz pod kątem ruchu
func (ods *ObjectDetectionService) AnalyzeImageForMotion(imageData []byte, previousImage []byte) (bool, error) {
	if previousImage == nil {
		return false, nil
	}

	// Tutaj można implementować detekcję ruchu
	// Na razie zawsze zwracamy true dla testów
	return true, nil
}

// FormatDetectionsAsJSON formatuje wyniki detekcji do JSON
func (ods *ObjectDetectionService) FormatDetectionsAsJSON(detections []DetectionResult) (string, error) {
	jsonData, err := json.Marshal(detections)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
