package ai

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"sync"
	"webserver/internal/config"
	"webserver/internal/logger"

	"gocv.io/x/gocv"
)

const (
	MotionThreshold    = 10000 // Default threshold for motion detection
	DetectionThreshold = 0.5   // Default threshold for object detection confidence
)

// DetectionResult reprezentuje wynik detekcji obiektu
type DetectionResult struct {
	Label      string
	Confidence float64
	X          int
	Y          int
	Width      int
	Height     int
}

// CameraState przechowuje stan detekcji ruchu dla każdej kamery, aby porównać klatki
type CameraState struct {
	previousMat gocv.Mat
	hasPrevious bool
	mutex       sync.Mutex
}

// DetectorService serwis do rozpoznawania obiektów
type DetectorService struct {
	cameraStates map[string]*CameraState // State dla każdej kamery osobno
	statesMutex  sync.RWMutex            // Mutex do mapy states
	net          gocv.Net
	modelPath    string
	configPath   string
	logger       *logger.Logger
}

// NewService tworzy nowy serwis detekcji
func NewDetectorService(config *config.Config, logger *logger.Logger) *DetectorService {
	service := &DetectorService{
		cameraStates: make(map[string]*CameraState),
		modelPath:    config.ModelPath,
		configPath:   config.ConfigPath,
		logger:       logger,
	}

	if err := service.initializeNet(); err != nil {
		service.logger.Warning("Could not initialize detection network: %v", err)
		return service
	}

	return service
}

// initializeNet inicjalizuje sieć detekcji z plików modelu i konfiguracji
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
	errBackend := net.SetPreferableBackend(gocv.NetBackendDefault)
	errTarget := net.SetPreferableTarget(gocv.NetTargetCPU)

	if errBackend != nil || errTarget != nil {
		return fmt.Errorf("failed to set preferable backend or target")
	}

	s.net = net
	s.logger.Info("Detection network initialized successfully")
	return nil
}

func (s *DetectorService) getCameraState(cameraID string) *CameraState {
	s.statesMutex.RLock()
	state, exists := s.cameraStates[cameraID]
	s.statesMutex.RUnlock()

	if exists {
		return state
	}

	// Tworzymy nowy state
	s.statesMutex.Lock()
	defer s.statesMutex.Unlock()

	// Double-check (może zostać utworzony przez inny wątek)
	if state, exists := s.cameraStates[cameraID]; exists {
		return state
	}

	state = &CameraState{
		hasPrevious: false,
	}
	s.cameraStates[cameraID] = state
	s.logger.Info("Created motion detection state for camera: %s", cameraID)

	return state
}

// DetectMotion wykrywa ruch między klatkami
func (s *DetectorService) DetectMotion(imageBytes []byte, cameraID string) (bool, error) {
	// Convert bytes to Mat
	state := s.getCameraState(cameraID)
	state.mutex.Lock()
	defer state.mutex.Unlock()

	mat, err := gocv.IMDecode(imageBytes, gocv.IMReadColor)
	if err != nil {
		return false, fmt.Errorf("failed to decode image: %v", err)
	}
	defer mat.Close()

	if mat.Empty() {
		return false, fmt.Errorf("decoded image is empty")
	}

	if !state.hasPrevious {
		state.previousMat = mat.Clone()
		state.hasPrevious = true
		s.logger.Info("Initialized motion detection for camera: %s", cameraID)
		return false, nil
	}

	// Calculate difference
	diff := gocv.NewMat()
	defer diff.Close()
	err = gocv.AbsDiff(state.previousMat, mat, &diff)
	if err != nil {
		return false, fmt.Errorf("failed to compute absolute difference: %v", err)
	}
	// Convert to grayscale
	gray := gocv.NewMat()
	defer gray.Close()
	err = gocv.CvtColor(diff, &gray, gocv.ColorBGRToGray)
	if err != nil {
		return false, fmt.Errorf("failed to convert image to grayscale: %v", err)
	}
	// Apply threshold
	thresh := gocv.NewMat()
	defer thresh.Close()
	gocv.Threshold(gray, &thresh, 30, 255, gocv.ThresholdBinary)

	// Count non-zero pixels
	nonZeroPixels := gocv.CountNonZero(thresh)

	// Update previous frame
	state.previousMat.Close()
	if mat.Empty() {
		return false, fmt.Errorf("decoded image is empty")
	}
	state.previousMat = mat.Clone()

	// Motion detected if more than motionThreshold pixels changed
	motionDetected := nonZeroPixels > MotionThreshold

	if motionDetected {
		s.logger.Info("Motion detected: %d pixels changed", nonZeroPixels)
	}

	return motionDetected, nil
}

func (s *DetectorService) DetectObjects(imageBytes []byte) ([]DetectionResult, error) {
	if s.net.Empty() {
		return []DetectionResult{}, fmt.Errorf("detection network not initialized")
	}

	// Dekoduj obraz
	mat, err := gocv.IMDecode(imageBytes, gocv.IMReadColor)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}
	defer mat.Close()

	// Sprawdź czy obraz nie jest pusty
	if mat.Empty() {
		return nil, fmt.Errorf("decoded image is empty")
	}

	blob := gocv.BlobFromImage(mat, 1.0/127.5, image.Pt(300, 300), gocv.NewScalar(127.5, 127.5, 127.5, 0), true, false)
	defer blob.Close()

	// Ustaw input dla sieci
	s.net.SetInput(blob, "")

	output := s.net.Forward("")
	defer output.Close()

	var results []DetectionResult

	// Process detections
	outputReshaped := output.Reshape(1, output.Total()/7)
	for i := 0; i < outputReshaped.Rows(); i++ {
		confidence := outputReshaped.GetFloatAt(i, 2)
		if confidence > DetectionThreshold {
			classID := int(outputReshaped.GetFloatAt(i, 1))
			x := int(outputReshaped.GetFloatAt(i, 3) * float32(mat.Cols()))
			y := int(outputReshaped.GetFloatAt(i, 4) * float32(mat.Rows()))
			width := int(outputReshaped.GetFloatAt(i, 5)*float32(mat.Cols())) - x
			height := int(outputReshaped.GetFloatAt(i, 6)*float32(mat.Rows())) - y

			results = append(results, DetectionResult{
				Label:      getClassLabel(classID),
				Confidence: float64(confidence),
				X:          x,
				Y:          y,
				Width:      width,
				Height:     height,
			})
			s.logger.Info("Detected %s", results[len(results)-1].Label)
		}
	}

	return results, nil
}

// DrawRectangle rysuje prostokąty na obrazie
func (s *DetectorService) DrawRectangle(detections []DetectionResult, img []byte) ([]byte, error) {
	red := color.RGBA{R: 255, G: 0, B: 0, A: 0}

	mat, err := gocv.IMDecode(img, gocv.IMReadColor)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}
	defer mat.Close()

	for _, detection := range detections {
		rect := image.Rect(detection.X, detection.Y, detection.X+detection.Width, detection.Y+detection.Height)
		err = gocv.Rectangle(&mat, rect, red, 2)
		if err != nil {
			return nil, fmt.Errorf("failed to draw rectangle: %v", err)
		}

		// Opcjonalnie: dodaj etykietę klasy
		label := fmt.Sprintf("%s (%.2f)", detection.Label, detection.Confidence)
		pt := image.Pt(detection.X, detection.Y-5)
		err = gocv.PutText(&mat, label, pt, gocv.FontHersheySimplex, 0.5, red, 1)
		if err != nil {
			return nil, fmt.Errorf("failed to draw text: %v", err)
		}
	}

	buf, err := gocv.IMEncode(".jpg", mat)
	if err != nil {
		s.logger.Error("Failed to encode image: %v", err)
		return nil, err
	}
	defer buf.Close()
	finalImage := make([]byte, len(buf.GetBytes()))
	copy(finalImage, buf.GetBytes())

	return finalImage, nil
}

func getClassLabel(classID int) string {
	labels := map[int]string{
		1:  "osoba",
		2:  "rower",
		3:  "samochód",
		4:  "motocykl",
		5:  "samolot",
		6:  "autobus",
		8:  "ciężarówka",
		16: "ptak",
		17: "kot",
		18: "pies",
	}

	if label, exists := labels[classID]; exists {
		return label
	}
	return fmt.Sprintf("nieznany_%d", classID)
}
