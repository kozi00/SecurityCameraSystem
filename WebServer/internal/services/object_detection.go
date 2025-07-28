package services

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

// ObjectDetectionService serwis do rozpoznawania obiektów
type ObjectDetectionService struct {
	enabled     bool
	previousMat gocv.Mat
	hasPrevious bool     // flaga do sprawdzania czy mamy poprzednią klatkę (false jesli zaczynamy program)
	net         gocv.Net // sieć do detekcji obiektów
}

func (ods *ObjectDetectionService) Close() {
	if ods.hasPrevious {
		ods.previousMat.Close()
		ods.hasPrevious = false
	}
	ods.net.Close()
}

// NewObjectDetectionService tworzy nowy serwis detekcji
func NewObjectDetectionService() *ObjectDetectionService {
	service := &ObjectDetectionService{
		enabled:     true,
		hasPrevious: false,
	}

	// pelne sciezki
	modelPath := "D:\\2025Scripts\\SecurityCameraSystem\\WebServer\\internal\\services\\AI\\frozen_inference_graph.pb"
	configPath := "D:\\2025Scripts\\SecurityCameraSystem\\WebServer\\internal\\services\\AI\\ssd_mobilenet_v1_coco_2017_11_17.pbtxt"

	// Sprawdź czy pliki istnieją
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		log.Printf("uwaga: ścieżka do załadowania modelu jest nieprawidłowa: %s", modelPath)
		return nil
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("uwaga: ścieżka do załadowania sieci jest nieprawidłowa: %s", configPath)
		return nil
	}

	service.net = gocv.ReadNet(modelPath, configPath)
	if service.net.Empty() {
		log.Printf("nie można załadować modelu AI z plików: %s, %s", modelPath, configPath)
		return nil
	}

	log.Printf("Model detekcji obiektów załadowany pomyślnie")
	return service
}

// DetectObjects wykrywa obiekty na obrazie
func (ods *ObjectDetectionService) DetectObjects(imageData []byte) ([]DetectionResult, error) {
	if !ods.enabled {
		return nil, nil
	}
	// Dekodowanie obrazu do gocv.Mat
	mat, err := gocv.IMDecode(imageData, gocv.IMReadColor)
	if err != nil {
		return nil, fmt.Errorf("błąd dekodowania do Mat: %v", err)
	}
	defer mat.Close()

	if ods.net.Empty() {
		return nil, fmt.Errorf("model detekcji nie jest załadowany")
	}
	//blob := gocv.BlobFromImage(mat, 1.0, image.Pt(300, 300), gocv.NewScalar(104, 177, 123, 0), false, false)
	blob := gocv.BlobFromImage(mat, 1.0/127.5, image.Pt(300, 300), gocv.NewScalar(127.5, 127.5, 127.5, 0), true, false)
	defer blob.Close()

	ods.net.SetInput(blob, "")
	detections := ods.net.Forward("")
	defer detections.Close()

	// Mapowanie klas COCO (dla MobileNet SSD)
	classNames := map[int]string{
		1:  "osoba",
		2:  "rower",
		3:  "samochód",
		4:  "motocykl",
		6:  "autobus",
		7:  "pociąg",
		8:  "ciężarówka",
		15: "kot",
		16: "pies",
	}
	results := ods.PerformDetection(&mat, &detections, classNames)

	log.Printf("Wykryto %d obiektów na obrazie", len(results))
	return results, nil
}

func (ods *ObjectDetectionService) PerformDetection(mat *gocv.Mat, detections *gocv.Mat, classNames map[int]string) []DetectionResult {
	rows := detections.Total() / 7
	var results []DetectionResult

	for i := 0; i < int(rows); i++ {
		confidence := detections.GetFloatAt(0, i+2)
		if confidence > 0.5 { // Próg pewności 50%
			classID := int(detections.GetFloatAt(0, i+1))

			// Sprawdź czy znamy tę klasę
			className, exists := classNames[classID]
			if !exists {
				continue
			}

			// Pobierz współrzędne (normalizowane 0-1)
			left := detections.GetFloatAt(0, i+3)
			top := detections.GetFloatAt(0, i+4)
			right := detections.GetFloatAt(0, i+5)
			bottom := detections.GetFloatAt(0, i+6)

			// Konwertuj do pikseli (zakładając rozmiar obrazu)
			imgWidth := float32(mat.Cols())
			imgHeight := float32(mat.Rows())

			x := int(left * imgWidth)
			y := int(top * imgHeight)
			width := int((right - left) * imgWidth)
			height := int((bottom - top) * imgHeight)

			results = append(results, DetectionResult{
				Label:      className,
				Confidence: float64(confidence),
				X:          x,
				Y:          y,
				Width:      width,
				Height:     height,
			})

			log.Printf("Wykryto: %s (%.2f%%) na pozycji [%d,%d,%d,%d]",
				className, confidence*100, x, y, width, height)
		}
	}
	return results
}

// AnalyzeImageForMotion analizuje obraz pod kątem ruchu
func (ods *ObjectDetectionService) DetectMotion(imageData []byte) (bool, error) {
	currentMat, err := gocv.IMDecode(imageData, gocv.IMReadGrayScale)
	if err != nil {
		return false, fmt.Errorf("błąd dekodowania obrazu: %v", err)
	}
	defer currentMat.Close()

	// Jeśli to pierwsza klatka, skopiuj ją jako poprzednią
	if !ods.hasPrevious {
		ods.previousMat = gocv.NewMat()
		currentMat.CopyTo(&ods.previousMat)
		ods.hasPrevious = true
		return false, nil
	}

	// Oblicz różnicę między obrazami
	diff := gocv.NewMat()
	defer diff.Close()
	gocv.AbsDiff(ods.previousMat, currentMat, &diff) // ✅ Bez * przed ods.previousMat

	// Binaryzacja
	thresh := gocv.NewMat()
	defer thresh.Close()
	gocv.Threshold(diff, &thresh, 25, 255, gocv.ThresholdBinary)

	// Znajdź kontury
	contours := gocv.FindContours(thresh, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	defer contours.Close()

	motionDetected := false
	minContourArea := 5000.0

	for i := 0; i < contours.Size(); i++ {
		contour := contours.At(i)
		area := gocv.ContourArea(contour)

		if area > minContourArea {
			motionDetected = true
			log.Printf("Wykryto ruch o powierzchni: %.2f", area)
			break
		}
	}

	// Zapisz aktualną klatkę jako poprzednią (kopiuj dane)
	currentMat.CopyTo(&ods.previousMat)

	return motionDetected, nil
}

// FormatDetectionsAsJSON formatuje wyniki detekcji do JSON
func (ods *ObjectDetectionService) FormatDetectionsAsJSON(detections []DetectionResult) (string, error) {
	jsonData, err := json.Marshal(detections)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
