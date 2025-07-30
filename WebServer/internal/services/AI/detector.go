package ai

import (
	"encoding/json"
	"fmt"
	"image"
	"log"
	"os"

	"gocv.io/x/gocv"
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

// Service serwis do rozpoznawania obiektów
type DetectorService struct {
	previousMat gocv.Mat
	hasPrevious bool     // flaga do sprawdzania czy mamy poprzednią klatkę (false jesli zaczynamy program)
	net         gocv.Net // sieć do detekcji obiektów
	modelPath   string
	configPath  string
}

// NewService tworzy nowy serwis detekcji
func NewDetectorService(modelPath, configPath string) *DetectorService {
	service := &DetectorService{
		modelPath:  modelPath,
		configPath: configPath,
	}

	if err := service.initializeNet(); err != nil {
		log.Printf("Warning: Could not initialize detection network: %v", err)
		return service // Return service anyway, will work in fallback mode
	}

	return service
}

func (s *DetectorService) initializeNet() error {
	if _, err := os.Stat(s.modelPath); os.IsNotExist(err) {
		return fmt.Errorf("model file not found: %s", s.modelPath)
	}

	if _, err := os.Stat(s.configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", s.configPath)
	}

	net := gocv.ReadNet(s.modelPath, s.configPath)
	if net.Empty() {
		return fmt.Errorf("failed to load network")
	}

	s.net = net
	log.Printf("Detection network initialized successfully")
	return nil
}

// DetectMotion wykrywa ruch między klatkami
func (s *DetectorService) DetectMotion(imageBytes []byte) (bool, error) {
	// Convert bytes to Mat
	mat, err := gocv.IMDecode(imageBytes, gocv.IMReadColor)
	if err != nil {
		return false, fmt.Errorf("failed to decode image: %v", err)
	}
	defer mat.Close()

	if !s.hasPrevious {
		s.previousMat = mat.Clone()
		s.hasPrevious = true
		return false, nil // Pierwsza klatka - brak ruchu
	}

	// Calculate difference
	diff := gocv.NewMat()
	defer diff.Close()

	gocv.AbsDiff(s.previousMat, mat, &diff)

	// Convert to grayscale
	gray := gocv.NewMat()
	defer gray.Close()
	gocv.CvtColor(diff, &gray, gocv.ColorBGRToGray)

	// Apply threshold
	thresh := gocv.NewMat()
	defer thresh.Close()
	gocv.Threshold(gray, &thresh, 30, 255, gocv.ThresholdBinary)

	// Count non-zero pixels
	nonZeroPixels := gocv.CountNonZero(thresh)

	// Update previous frame
	s.previousMat.Close()
	s.previousMat = mat.Clone()

	// Motion detected if more than 1000 pixels changed
	motionDetected := nonZeroPixels > 1000

	if motionDetected {
		log.Printf("Motion detected: %d pixels changed", nonZeroPixels)
	}

	return motionDetected, nil
}

// DetectObjects wykrywa obiekty na obrazie
func (s *DetectorService) DetectObjects(imageBytes []byte) ([]DetectionResult, error) {
	if s.net.Empty() {
		return []DetectionResult{}, fmt.Errorf("detection network not initialized")
	}

	// Convert bytes to Mat
	mat, err := gocv.IMDecode(imageBytes, gocv.IMReadColor)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}
	defer mat.Close()

	// Create blob from image
	blob := gocv.BlobFromImage(mat, 1.0/127.5, image.Pt(300, 300), gocv.NewScalar(127.5, 127.5, 127.5, 0), true, false)
	defer blob.Close()

	// Set input to network
	s.net.SetInput(blob, "")

	// Run forward pass
	output := s.net.Forward("")
	defer output.Close()

	var results []DetectionResult

	// Process detections
	for i := 0; i < output.Total(); i += 7 {
		confidence := output.GetFloatAt(0, i+2)
		if confidence > 0.5 { // Confidence threshold
			classID := int(output.GetFloatAt(0, i+1))
			x := int(output.GetFloatAt(0, i+3) * float32(mat.Cols()))
			y := int(output.GetFloatAt(0, i+4) * float32(mat.Rows()))
			width := int(output.GetFloatAt(0, i+5)*float32(mat.Cols())) - x
			height := int(output.GetFloatAt(0, i+6)*float32(mat.Rows())) - y

			results = append(results, DetectionResult{
				Label:      getClassLabel(classID),
				Confidence: float64(confidence),
				X:          x,
				Y:          y,
				Width:      width,
				Height:     height,
			})
		}
	}

	return results, nil
}

// FormatDetectionsAsJSON formatuje wyniki detekcji do JSON
func (s *DetectorService) FormatDetectionsAsJSON(detections []DetectionResult) (string, error) {
	jsonBytes, err := json.Marshal(detections)
	if err != nil {
		return "", fmt.Errorf("failed to marshal detections: %v", err)
	}
	return string(jsonBytes), nil
}

// getClassLabel zwraca etykietę klasy dla danego ID
func getClassLabel(classID int) string {
	labels := map[int]string{
		1:  "person",
		2:  "bicycle",
		3:  "car",
		4:  "motorcycle",
		5:  "airplane",
		6:  "bus",
		7:  "train",
		8:  "truck",
		9:  "boat",
		10: "traffic light",
		// Add more labels as needed
	}

	if label, exists := labels[classID]; exists {
		return label
	}
	return fmt.Sprintf("unknown_%d", classID)
}

// Close zamyka serwis i zwalnia zasoby
func (s *DetectorService) Close() {
	if !s.net.Empty() {
		s.net.Close()
	}
	if s.hasPrevious {
		s.previousMat.Close()
	}
}
