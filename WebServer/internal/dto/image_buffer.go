package dto

// BufferedImage holds image data and detection results before flushing to disk.

type BufferedImage struct {
	Timestamp  string
	Camera     string
	Detections []DetectionResult
	Data       []byte
}
