package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
	"webserver/internal/config"
	"webserver/internal/logger"
)

const (
	ImageBufferLimit         = 7
	ImageBufferFlushInterval = 30
)

type Image struct {
	Timestamp string
	Camera    string
	Object    string
	Data      []byte
}

type BufferService struct {
	imagesDir   string
	images      []Image
	bufferCount map[string]int // Limit for each camera
	mu          sync.Mutex
	logger      *logger.Logger
}

func NewBufferService(config *config.Config, logger *logger.Logger) *BufferService {
	return &BufferService{
		imagesDir:   config.ImageDirectory,
		images:      make([]Image, 0),
		bufferCount: make(map[string]int),
		logger:      logger,
		mu:          sync.Mutex{},
	}
}

func (s *BufferService) Run() {
	ticker := time.NewTicker(time.Duration(ImageBufferFlushInterval) * time.Second)

	defer ticker.Stop()
	for {
		<-ticker.C
		s.FlushImages()
	}
}

func (s *BufferService) AddImage(imageData []byte, cameraId, object string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02_15-04_05.000")
	image := Image{
		Timestamp: timestamp,
		Camera:    cameraId,
		Object:    object,
		Data:      imageData,
	}

	if s.bufferCount[cameraId] < ImageBufferLimit {
		s.logger.Info("Buffer size for camera %s: %d/%d", cameraId, len(s.images), ImageBufferLimit)
		s.images = append(s.images, image)
		s.bufferCount[cameraId]++
	}
}

func (s *BufferService) FlushImages() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.images) == 0 {
		return
	}

	if err := os.MkdirAll(s.imagesDir, 0755); err != nil {
		s.logger.Error("Error creating directory: %v", err)
		return
	}

	for _, image := range s.images {
		filename := fmt.Sprintf("%s_%s_%s.jpg", image.Timestamp, image.Camera, image.Object)
		fullpath := filepath.Join(s.imagesDir, filename)

		if err := os.WriteFile(fullpath, image.Data, 0644); err != nil {
			s.logger.Error("Error saving image %s: %v", filename, err)
			continue
		}
	}

	s.logger.Info("Flushed %d images to disk", len(s.images))
	s.images = s.images[:0] // Clear buffer
	s.bufferCount = make(map[string]int)
}
