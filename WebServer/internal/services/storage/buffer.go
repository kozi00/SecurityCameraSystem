package storage

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
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
	bufferLimit int
	mu          sync.Mutex
}

func NewBufferService(imagesDir string, bufferLimit int) *BufferService {
	return &BufferService{
		imagesDir:   imagesDir,
		bufferLimit: bufferLimit,
		images:      make([]Image, 0),
		mu:          sync.Mutex{},
	}
}

func (s *BufferService) Run(flushInterval int) {
	ticker := time.NewTicker(time.Duration(flushInterval) * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.FlushImages()
	}
}

func (s *BufferService) AddImage(imageData []byte, camera, object string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	image := Image{
		Timestamp: timestamp,
		Camera:    camera,
		Object:    object,
		Data:      imageData,
	}

	s.images = append(s.images, image)

	if len(s.images) >= s.bufferLimit {
		go s.FlushImages()
	}
}

func (s *BufferService) SaveImage(imageData []byte, camera string) error {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("%s_%s.jpg", camera, timestamp)
	filepath := fmt.Sprintf("%s/%s", s.imagesDir, filename)

	if err := os.MkdirAll(s.imagesDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	if err := os.WriteFile(filepath, imageData, 0644); err != nil {
		return fmt.Errorf("failed to save image: %v", err)
	}

	log.Printf("Image saved: %s", filename)
	return nil
}

func (s *BufferService) FlushImages() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.images) == 0 {
		return
	}

	if err := os.MkdirAll(s.imagesDir, 0755); err != nil {
		log.Printf("Error creating directory: %v", err)
		return
	}

	for _, image := range s.images {
		filename := fmt.Sprintf("%s_%s_%s.jpg", image.Camera, image.Timestamp, image.Object)
		filepath := fmt.Sprintf("%s/%s", s.imagesDir, filename)

		if err := os.WriteFile(filepath, image.Data, 0644); err != nil {
			log.Printf("Error saving image %s: %v", filename, err)
			continue
		}
	}

	log.Printf("Flushed %d images to disk", len(s.images))
	s.images = s.images[:0] // Clear buffer
}
