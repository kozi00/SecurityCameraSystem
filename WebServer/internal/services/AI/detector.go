package ai

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"os"

	"gocv.io/x/gocv"
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

// Service serwis do rozpoznawania obiektów
type DetectorService struct {
	previousMat     gocv.Mat
	hasPrevious     bool     // flaga do sprawdzania czy mamy poprzednią klatkę (false jesli zaczynamy program)
	net             gocv.Net // sieć do detekcji obiektów
	modelPath       string
	configPath      string
	motionThreshold int
}

// NewService tworzy nowy serwis detekcji
func NewDetectorService(modelPath, configPath string, motionThreshold int) *DetectorService {
	service := &DetectorService{
		modelPath:       modelPath,
		configPath:      configPath,
		motionThreshold: motionThreshold,
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

	//net := gocv.ReadNetFromONNX(s.modelPath)
	net := gocv.ReadNet(s.modelPath, s.configPath)

	if net.Empty() {
		return fmt.Errorf("failed to load network")
	}
	net.SetPreferableBackend(gocv.NetBackendDefault)
	net.SetPreferableTarget(gocv.NetTargetCPU)

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

	// Motion detected if more than motionThreshold pixels changed
	motionDetected := nonZeroPixels > s.motionThreshold

	if motionDetected {
		log.Printf("Motion detected: %d pixels changed", nonZeroPixels)
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

	originalWidth := mat.Cols()
	originalHeight := mat.Rows()

	// Tworzenie blob z obrazu - zoptymalizowane parametry
	blob := gocv.BlobFromImage(
		mat,
		1.0/255.0,                  // scalefactor
		image.Pt(320, 320),         // size - mniejszy dla Orange Pi
		gocv.NewScalar(0, 0, 0, 0), // mean
		true,                       // swapRB
		false,                      // crop
	)
	defer blob.Close()

	// Ustaw input dla sieci
	s.net.SetInput(blob, "")

	// Forward pass przez wszystkie warstwy wyjściowe
	outputLayers := getOutputLayers(s.net)
	outputs := s.net.ForwardLayers(outputLayers)
	defer func() {
		for i := range outputs {
			outputs[i].Close()
		}
	}()

	// Przetwórz wyniki z wszystkich warstw wyjściowych
	detections := s.processOutputs(outputs, originalWidth, originalHeight)
	return detections, nil
}

// processOutputs przetwarza wyniki z warstw wyjściowych YOLO
func (s *DetectorService) processOutputs(outputs []gocv.Mat, originalWidth, originalHeight int) []DetectionResult {
	var detections []DetectionResult

	for _, output := range outputs {
		rows := output.Size()[0]
		cols := output.Size()[1]

		for i := 0; i < rows; i++ {
			// Pobierz dane dla jednego wykrycia
			data := output.RowRange(i, i+1)
			defer data.Close()

			// Współrzędne centrum i rozmiary (znormalizowane 0-1)
			centerX := data.GetFloatAt(0, 0)
			centerY := data.GetFloatAt(0, 1)
			width := data.GetFloatAt(0, 2)
			height := data.GetFloatAt(0, 3)

			// Objectness score
			objectness := data.GetFloatAt(0, 4)

			// Sprawdź objectness threshold
			if objectness < 0.5 {
				continue
			}

			// Znajdź klasę z najwyższym score
			maxClassScore := float32(0.0)
			classID := 0

			// Iteruj przez class scores (zaczynając od indeksu 5)
			for j := 5; j < cols; j++ {
				classScore := data.GetFloatAt(0, j)
				if classScore > maxClassScore {
					maxClassScore = classScore
					classID = j - 5
				}
			}

			// Oblicz końcowy confidence score
			finalConfidence := objectness * maxClassScore

			if finalConfidence > 0.5 {
				// Przelicz współrzędne na pixele
				pixelCenterX := centerX * float32(originalWidth)
				pixelCenterY := centerY * float32(originalHeight)
				pixelWidth := width * float32(originalWidth)
				pixelHeight := height * float32(originalHeight)

				// Oblicz współrzędne lewego górnego rogu
				x := int(pixelCenterX - pixelWidth/2)
				y := int(pixelCenterY - pixelHeight/2)

				// Upewnij się, że współrzędne są w granicach obrazu
				x = max(0, min(x, originalWidth))
				y = max(0, min(y, originalHeight))
				w := max(0, min(int(pixelWidth), originalWidth-x))
				h := max(0, min(int(pixelHeight), originalHeight-y))

				detections = append(detections, DetectionResult{
					Label:      s.GetClassLabel(classID),
					Confidence: float64(finalConfidence),
					X:          x,
					Y:          y,
					Width:      w,
					Height:     h,
				})
			}
		}
	}

	return detections
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
		gocv.Rectangle(&mat, rect, red, 2)

		// Opcjonalnie: dodaj etykietę klasy
		label := fmt.Sprintf("%s (%.2f)", detection.Label, detection.Confidence)
		pt := image.Pt(detection.X, detection.Y-5)
		gocv.PutText(&mat, label, pt, gocv.FontHersheySimplex, 0.5, red, 1)
	}

	buf, err := gocv.IMEncode(".jpg", mat)
	if err != nil {
		log.Printf("Failed to encode image: %v", err)
		return nil, err
	}
	defer buf.Close()
	finalImage := make([]byte, len(buf.GetBytes()))
	copy(finalImage, buf.GetBytes())

	return finalImage, nil
}

// GetClassLabel zwraca etykietę klasy dla danego ID
func (s *DetectorService) GetClassLabel(classID int) string {
	labels := map[int]string{
		0:  "osoba",
		1:  "rower",
		2:  "samochód",
		3:  "motocykl",
		4:  "samolot",
		5:  "autobus",
		7:  "ciężarówka",
		14: "ptak",
		15: "kot",
		16: "pies",
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
func getOutputLayers(net gocv.Net) []string {
	layerNames := net.GetLayerNames()
	unconnectedOutLayers := net.GetUnconnectedOutLayers()

	var outputLayers []string
	for _, i := range unconnectedOutLayers {
		if i-1 < len(layerNames) {
			outputLayers = append(outputLayers, layerNames[i-1])
		}
	}

	return outputLayers
}
