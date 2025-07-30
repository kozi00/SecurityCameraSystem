package storage

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
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
	ticker := time.NewTicker(time.Duration(flushInterval) * time.Second)

	defer ticker.Stop()
	for {
		<-ticker.C
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

	if len(s.images) < s.bufferLimit {
		log.Printf("Buffer size: %d/%d", len(s.images), s.bufferLimit)
		s.images = append(s.images, image)
	}
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
		filename := fmt.Sprintf("%s_%s_%s.jpg", image.Timestamp, image.Camera, image.Object)
		fullpath := filepath.Join(s.imagesDir, filename)

		if err := os.WriteFile(fullpath, image.Data, 0644); err != nil {
			log.Printf("Error saving image %s: %v", filename, err)
			continue
		}
	}

	log.Printf("Flushed %d images to disk", len(s.images))
	s.images = s.images[:0] // Clear buffer
}
