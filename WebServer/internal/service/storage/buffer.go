package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
	"webserver/internal/config"
	"webserver/internal/dto"
	"webserver/internal/logger"
	"webserver/internal/model"
	"webserver/internal/repository"
)

const (
	// ImageBufferLimit limits how many images per camera are buffered before flushing.
	ImageBufferLimit = 10
	// ImageBufferFlushInterval defines how often (seconds) buffered images are flushed to disk.
	ImageBufferFlushInterval = 30
)

// BufferService buffers images in memory and periodically flushes them to disk.
type BufferService struct {
	imagesDir     string
	images        []dto.BufferedImage
	bufferCount   map[string]int
	mu            sync.Mutex
	logger        *logger.Logger
	imageRepo     repository.ImageRepository
	detectionRepo repository.DetectionRepository
}

// NewBufferService creates a new BufferService with the target directory and logger.
func NewBufferService(config *config.Config, logger *logger.Logger, imageRepo repository.ImageRepository, detectionRepo repository.DetectionRepository) *BufferService {
	return &BufferService{
		imagesDir:     config.ImageDirectory,
		images:        make([]dto.BufferedImage, 0),
		bufferCount:   make(map[string]int),
		logger:        logger,
		imageRepo:     imageRepo,
		detectionRepo: detectionRepo,
		mu:            sync.Mutex{},
	}
}

// Run starts a ticker loop that periodically flushes images to disk.
func (s *BufferService) Run() {
	ticker := time.NewTicker(time.Duration(ImageBufferFlushInterval) * time.Second)

	defer ticker.Stop()
	for {
		<-ticker.C
		s.FlushImages()
	}
}

// AddImage appends an image to the in-memory buffer for a given camera.
func (s *BufferService) AddImage(imageData []byte, cameraId string, detections []dto.DetectionResult) {
	s.mu.Lock()
	defer s.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02_15-04_05.000")
	image := dto.BufferedImage{
		Timestamp:  timestamp,
		Camera:     cameraId,
		Detections: detections,
		Data:       imageData,
	}

	if s.bufferCount[cameraId] < ImageBufferLimit {
		s.logger.Info("Buffer size for camera %s: %d/%d", cameraId, len(s.images), ImageBufferLimit)
		s.images = append(s.images, image)
		s.bufferCount[cameraId]++
	}
}

// FlushImages writes buffered images to disk and resets the buffer and per-camera counters.
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

	savedCount := 0
	for _, image := range s.images {
		// Build filename with detected object names
		objects := ""
		for _, det := range image.Detections {
			objects += det.Label + "_"
		}

		filename := fmt.Sprintf("%s_%s_%s.jpg", image.Timestamp, image.Camera, objects)
		fullpath := filepath.Join(s.imagesDir, filename)

		if err := os.WriteFile(fullpath, image.Data, 0644); err != nil {
			s.logger.Error("Error saving image %s: %v", filename, err)
			continue
		}

		// Save to database if repositories are available
		if s.imageRepo != nil {
			ts, err := time.Parse("2006-01-02_15-04_05.000", image.Timestamp)
			if err != nil {
				ts = time.Now()
			}

			dbImage := &model.Image{
				Filename:  filename,
				Camera:    image.Camera,
				Timestamp: ts,
				FilePath:  fullpath,
				FileSize:  int64(len(image.Data)),
			}

			imageID, err := s.imageRepo.Insert(dbImage)
			if err != nil {
				s.logger.Error("Error saving image to database %s: %v", filename, err)
				continue
			}

			// Insert detections for this image
			if s.detectionRepo != nil && len(image.Detections) > 0 {
				var dbDetections []model.Detection
				for _, det := range image.Detections {
					dbDetections = append(dbDetections, model.Detection{
						ImageID:    imageID,
						ObjectName: det.Label,
						X:          det.X,
						Y:          det.Y,
						Width:      det.Width,
						Height:     det.Height,
						Confidence: det.Confidence,
					})
				}
				if err := s.detectionRepo.InsertBatch(dbDetections); err != nil {
					s.logger.Error("Error saving detections to database: %v", err)
				}
			}
		}

		savedCount++
	}

	s.logger.Info("Flushed %d images to disk", savedCount)
	s.images = s.images[:0] // Clear buffer
	s.bufferCount = make(map[string]int)
}
