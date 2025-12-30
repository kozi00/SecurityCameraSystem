package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
	"webserver/internal/config"
	"webserver/internal/database"
	"webserver/internal/logger"
	"webserver/internal/services"
)

const (
	// MaxImageDirectorySize defines the maximum size of the image directory in GB.
	MaxImageDirectorySize = 2
)

// PictureInfo represents parsed metadata about a stored picture.
type PictureInfo struct {
	Name      string    `json:"name"`
	Date      time.Time `json:"date"`
	TimeOfDay time.Time `json:"timeOfDay"`
	Camera    string    `json:"camera"`
	Objects   []string  `json:"objects"` // Multiple detected objects
}

// MarshalJSON customizes JSON output for PictureInfo to format date and time-of-day.
func (p PictureInfo) MarshalJSON() ([]byte, error) {
	type Alias PictureInfo
	return json.Marshal(&struct {
		Date      string `json:"date"`
		TimeOfDay string `json:"timeOfDay"`
		Alias
	}{
		Date:      p.Date.Format("02-01-2006"),
		TimeOfDay: p.TimeOfDay.Format("15:04"),
		Alias:     (Alias)(p),
	})
}

// PicturesData is a paginated response payload for the pictures gallery.
type PicturesData struct {
	Pictures    []PictureInfo `json:"pictures"`
	ImagesDir   string        `json:"imagesDir"`
	Size        int64         `json:"size"`
	MaxSize     int64         `json:"maxSize"`
	Length      int           `json:"length"`
	TotalPages  int           `json:"totalPages"`
	CurrentPage int           `json:"currentPage"`
	Limit       int           `json:"pageSize"`
}

// PictureFilters describe user-provided filters to narrow the picture list.
type PictureFilters struct {
	Camera     string
	Object     string
	DateAfter  time.Time
	DateBefore time.Time
	TimeAfter  time.Time
	TimeBefore time.Time
}

// DisplayPicturesHandler lists saved images, supports filtering and pagination.
// Response is JSON of type PicturesData.
func DisplayPicturesHandler(config *config.Config, logger *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		q := r.URL.Query()
		page := atoiDefault(q.Get("page"), 1)
		limit := atoiDefault(q.Get("limit"), 24)

		filters := PictureFilters{
			Camera:     q.Get("camera"),
			Object:     q.Get("object"),
			DateAfter:  parseDate(q.Get("dateAfter")),
			DateBefore: parseDate(q.Get("dateBefore")),
			TimeAfter:  parseTimeOfDay(q.Get("timeAfter")),
			TimeBefore: parseTimeOfDay(q.Get("timeBefore")),
		}

		files, err := os.ReadDir(config.ImageDirectory)
		if err != nil {
			logger.Error("Error reading image directory: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		filtered, totalSize := getFilteredPictures(filters, files, logger)

		slices.SortFunc(filtered, func(a, b PictureInfo) int {
			return strings.Compare(b.Name, a.Name) //Sorting from newest first
		})

		// Pagination
		start := (page - 1) * limit
		if start > len(filtered) {
			start = len(filtered)
		}
		end := start + limit
		if end > len(filtered) {
			end = len(filtered)
		}

		data := PicturesData{
			Pictures:    filtered[start:end],
			ImagesDir:   config.ImageDirectory,
			Size:        totalSize,
			MaxSize:     MaxImageDirectorySize,
			Length:      len(filtered),
			TotalPages:  (len(filtered) + limit - 1) / limit,
			CurrentPage: page,
			Limit:       limit,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(data); err != nil {
			logger.Error("Error encoding JSON response: %v", err)
		}
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

// ClearPicturesHandler deletes all files from the image directory.
func ClearPicturesHandler(config *config.Config, logger *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		files, err := os.ReadDir(config.ImageDirectory)
		if err != nil {
			logger.Error("Error reading pictures directory: %v", err)
			http.Error(w, "Unable to read pictures directory", http.StatusInternalServerError)
			return
		}

		for _, file := range files {
			if !file.IsDir() {
				filePath := filepath.Join(config.ImageDirectory, file.Name())
				err := os.Remove(filePath)
				if err != nil {
					logger.Error("Error deleting file %s: %v", file.Name(), err)
				}
			}
		}
		logger.Info("All pictures cleared from directory: %s", config.ImageDirectory)
		w.WriteHeader(http.StatusNoContent) // No Content response
	}
}

// helpers
// getFilteredPictures returns pictures that match provided filters and sums their total size.
func getFilteredPictures(filters PictureFilters, files []os.DirEntry, logger *logger.Logger) ([]PictureInfo, int64) {
	var (
		filtered  []PictureInfo
		totalSize int64
	)

	for _, e := range files {
		if e.IsDir() {
			continue
		}

		name := e.Name()
		info, err := e.Info()
		if err != nil {
			logger.Error("Error getting file info for %s: %v", name, err)
			continue
		}

		picture, err := parsePictureName(name)
		if err != nil {
			logger.Error("Failed to parse picture name %s: %v", name, err)
			continue
		}

		if !checkMatch(picture, filters) {
			continue
		}
		filtered = append(filtered, picture)
		totalSize += info.Size()
	}

	return filtered, totalSize
}

// checkMatch returns true if the picture matches the given filters.
func checkMatch(pic PictureInfo, filters PictureFilters) bool {
	if filters.Camera != "" && !strings.EqualFold(pic.Camera, filters.Camera) {
		return false
	}
	if filters.Object != "" {
		// Check if any of the detected objects matches the filter
		found := false
		for _, obj := range pic.Objects {
			if strings.EqualFold(obj, filters.Object) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if !filters.DateAfter.IsZero() && pic.Date.Before(filters.DateAfter) {
		return false
	}
	if !filters.DateBefore.IsZero() && pic.Date.After(filters.DateBefore) {
		return false
	}
	if !filters.TimeAfter.IsZero() && pic.TimeOfDay.Before(filters.TimeAfter) {
		return false
	}
	if !filters.TimeBefore.IsZero() && pic.TimeOfDay.After(filters.TimeBefore) {
		return false
	}
	return true
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

// parsePictureName parses files in the pattern
// 2006-01-02_15-04_05.000_camera_object1_object2_objectN.jpg
// The seconds.milliseconds segment is ignored for filtering; date and HH:MM are used.
func parsePictureName(filename string) (PictureInfo, error) {
	base := strings.TrimSuffix(filename, filepath.Ext(filename))
	parts := strings.Split(base, "_")

	// Format: [date, time, seconds.ms, camera, object1, object2, ...]
	if len(parts) >= 5 {
		d, err1 := time.Parse("2006-01-02", parts[0])
		hm, err2 := time.Parse("15-04", parts[1])
		if err1 == nil && err2 == nil {
			camera := parts[3]
			// All parts from index 4 onwards are detected objects
			objects := parts[4 : len(parts)-1]

			return PictureInfo{
				Name:      filename,
				Date:      d,
				TimeOfDay: hm,
				Camera:    camera,
				Objects:   objects,
			}, nil
		}
	}
	return PictureInfo{}, errors.New("invalid picture name format")
}

// =====================================================
// Database-based handlers
// =====================================================

// GetPicturesFromDBHandler returns filtered list of images from database.
func GetPicturesFromDBHandler(manager *services.Manager, cfg *config.Config, logger *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db := manager.GetBufferService().GetDatabase()
		if db == nil {
			// Fallback to file-based handler if database not available
			DisplayPicturesHandler(cfg, logger)(w, r)
			return
		}

		q := r.URL.Query()
		page := atoiDefault(q.Get("page"), 1)
		limit := atoiDefault(q.Get("limit"), 24)

		filter := &database.ImageFilter{
			Camera:     q.Get("camera"),
			Object:     q.Get("object"),
			StartDate:  parseDate(q.Get("dateAfter")),
			EndDate:    parseDate(q.Get("dateBefore")),
			TimeAfter:  q.Get("timeAfter"),
			TimeBefore: q.Get("timeBefore"),
			Limit:      limit,
			Offset:     (page - 1) * limit,
		}

		images, err := db.GetImages(filter)
		if err != nil {
			logger.Error("Error querying images from database: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		totalCount, err := db.GetTotalCount(filter)
		if err != nil {
			logger.Error("Error counting images: %v", err)
			totalCount = len(images)
		}

		// Calculate total size
		var totalSize int64
		for _, img := range images {
			totalSize += img.FileSize
		}

		// Convert to PictureInfo format for backwards compatibility
		var pictures []PictureInfo
		for _, img := range images {
			pictures = append(pictures, PictureInfo{
				Name:      img.Filename,
				Date:      img.Timestamp,
				TimeOfDay: img.Timestamp,
				Camera:    img.Camera,
				Objects:   img.Objects,
			})
		}

		data := PicturesData{
			Pictures:    pictures,
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

// GetFiltersHandler returns available cameras and objects for filtering.
func GetFiltersHandler(manager *services.Manager, logger *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db := manager.GetBufferService().GetDatabase()
		if db == nil {
			http.Error(w, "Database not available", http.StatusInternalServerError)
			return
		}

		cameras, err := db.GetCameras()
		if err != nil {
			logger.Error("Failed to get cameras: %v", err)
			cameras = []string{}
		}

		objects, err := db.GetObjects()
		if err != nil {
			logger.Error("Failed to get objects: %v", err)
			objects = []string{}
		}

		response := map[string]interface{}{
			"cameras": cameras,
			"objects": objects,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// GetStatsHandler returns image statistics.
func GetStatsHandler(manager *services.Manager, logger *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db := manager.GetBufferService().GetDatabase()
		if db == nil {
			http.Error(w, "Database not available", http.StatusInternalServerError)
			return
		}

		stats, err := db.GetStats()
		if err != nil {
			logger.Error("Failed to get stats: %v", err)
			http.Error(w, "Failed to retrieve stats", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}
}

// DeletePictureHandler removes an image from disk and database.
func DeletePictureHandler(manager *services.Manager, cfg *config.Config, logger *logger.Logger) http.HandlerFunc {
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
		db := manager.GetBufferService().GetDatabase()
		if db != nil {
			if err := db.DeleteImageByFilename(filename); err != nil {
				logger.Error("Failed to delete from database: %v", err)
			}
		}

		logger.Info("Deleted picture: %s", filename)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "deleted", "filename": filename})
	}
}

// ClearPicturesWithDBHandler deletes all files from the image directory and clears the database.
func ClearPicturesWithDBHandler(manager *services.Manager, cfg *config.Config, logger *logger.Logger) http.HandlerFunc {
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

		// Clear database if available
		db := manager.GetBufferService().GetDatabase()
		if db != nil {
			if err := db.ClearAll(); err != nil {
				logger.Error("Error clearing database: %v", err)
			}
		}

		logger.Info("All pictures cleared from directory: %s", cfg.ImageDirectory)
		w.WriteHeader(http.StatusNoContent)
	}
}
