package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"webserver/internal/config"
	"webserver/internal/dto"
	"webserver/internal/logger"
	"webserver/internal/repository"
	"webserver/internal/service"
)

const (
	// MaxImageDirectorySize defines the maximum size of the image directory in GB.
	MaxImageDirectorySize = 2
)

// GetPicturesFromDBHandler returns filtered list of images from database.
func GetPicturesFromDBHandler(manager *service.Manager, cfg *config.Config, logger *logger.Logger,
	imageRepo repository.ImageRepository, detectionRepo repository.DetectionRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		q := r.URL.Query()
		page := atoiDefault(q.Get("page"), 1)
		limit := atoiDefault(q.Get("limit"), 24)

		filter := &dto.ImageFilters{
			Camera:     q.Get("camera"),
			Object:     q.Get("object"),
			DateAfter:  parseDate(q.Get("dateAfter")),
			DateBefore: parseDate(q.Get("dateBefore")),
			TimeAfter:  parseTimeOfDay(q.Get("timeAfter")),
			TimeBefore: parseTimeOfDay(q.Get("timeBefore")),
		}

		images, err := imageRepo.GetAll(filter)
		if err != nil {
			logger.Error("Error querying images from database: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		totalSize, err := imageRepo.GetDirectorySize()
		if err != nil {
			logger.Error("Error getting image directory size: %v", err)
			totalSize = 0
		}

		totalCount, err := imageRepo.GetTotalCount(filter)
		if err != nil {
			logger.Error("Error counting images: %v", err)
			totalCount = len(images)
		}

		// Convert to ImageInfo format for response
		var pictures []dto.ImageInfo
		for _, img := range images {
			// Get object names for this image
			var objects []string
			if detectionRepo != nil {
				objects, err = detectionRepo.GetObjectNamesByImageID(img.ID)
				if err != nil {
					logger.Error("Error getting objects for image %d: %v", img.ID, err)
					objects = []string{}
				}
			}

			pictures = append(pictures, dto.ImageInfo{
				Name:      img.Filename,
				Date:      img.Timestamp,
				TimeOfDay: img.Timestamp,
				Camera:    img.Camera,
				Objects:   objects,
			})
		}

		data := dto.ImagesData{
			Images:      pictures,
			ImagesDir:   cfg.ImageDirectory,
			Size:        totalSize,
			MaxSize:     MaxImageDirectorySize,
			Length:      totalCount,
			TotalPages:  (totalCount + limit - 1) / limit,
			CurrentPage: page,
			Limit:       limit,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(data); err != nil {
			logger.Error("Error encoding JSON response: %v", err)
		}
	}
}

// DeletePictureHandler removes an image from disk and database.
func DeletePictureHandler(manager *service.Manager, cfg *config.Config, logger *logger.Logger,
	imageRepo repository.ImageRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filename := r.URL.Query().Get("filename")
		if filename == "" {
			http.Error(w, "Filename required", http.StatusBadRequest)
			return
		}

		// Delete file from disk
		filePath := filepath.Join(cfg.ImageDirectory, filename)
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			logger.Error("Failed to delete file %s: %v", filePath, err)
		}

		// Delete from database if available
		if imageRepo != nil {
			if err := imageRepo.DeleteByFilename(filename); err != nil {
				logger.Error("Failed to delete from database: %v", err)
			}
		}

		logger.Info("Deleted picture: %s", filename)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "deleted", "filename": filename})
	}
}

// ClearPicturesWithDBHandler deletes all files from the image directory and clears the database.
func ClearPicturesWithDBHandler(manager *service.Manager, cfg *config.Config, logger *logger.Logger,
	imageRepo repository.ImageRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		files, err := os.ReadDir(cfg.ImageDirectory)
		if err != nil {
			logger.Error("Error reading pictures directory: %v", err)
			http.Error(w, "Unable to read pictures directory", http.StatusInternalServerError)
			return
		}

		for _, file := range files {
			if !file.IsDir() {
				filePath := filepath.Join(cfg.ImageDirectory, file.Name())
				err := os.Remove(filePath)
				if err != nil {
					logger.Error("Error deleting file %s: %v", file.Name(), err)
				}
			}
		}

		if imageRepo != nil {
			if err := imageRepo.DeleteAll(); err != nil {
				logger.Error("Error clearing database: %v", err)
			}
		}

		logger.Info("All pictures cleared from directory: %s", cfg.ImageDirectory)
		w.WriteHeader(http.StatusNoContent)
	}
}

// ViewPictureHandler serves a single image file specified via the "image" query parameter.
func ViewPictureHandler(config *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		image := r.URL.Query().Get("image")
		if image == "" {
			http.Error(w, "Image parameter is required", http.StatusBadRequest)
			return
		}
		filePath := filepath.Join(config.ImageDirectory, image)
		http.ServeFile(w, r, filePath)
	}
}

// atoiDefault converts string to int or returns a default when conversion fails or value <= 0.
func atoiDefault(s string, def int) int {
	if v, err := strconv.Atoi(s); err == nil && v > 0 {
		return v
	}
	return def
}

// parseDate parses a date string in the format "2006-01-02" from the request (HTML input format).
func parseDate(v string) time.Time {
	if v == "" {
		return time.Time{}
	}
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		return time.Time{}
	}
	return t
}

// parseTimeOfDay parses a time-of-day string in the format "15:04" from the request (HTML input format).
func parseTimeOfDay(v string) time.Time {
	if v == "" {
		return time.Time{}
	}
	t, err := time.Parse("15:04", v)
	if err != nil {
		return time.Time{}
	}
	return t
}
