package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
	"webserver/internal/config"
	"webserver/internal/logger"
)

const (
	MaxImageDirectorySize = 2 // Maksymalny rozmiar katalogu z obrazami w GB
)

type PicturesData struct {
	Pictures    []string          `json:"pictures"`
	ImagesDir   string            `json:"imagesDir"`
	Size        int64             `json:"size"`
	MaxSize     int64             `json:"maxSize"`
	Length      int               `json:"length"`
	TotalPages  int               `json:"totalPages"`
	CurrentPage int               `json:"currentPage"`
	Limit       int               `json:"pageSize"`
	Filters     map[string]string `json:"filters,omitempty"`
}

func DisplayPicturesHandler(config *config.Config, logger *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		pageString := q.Get("page")
		limitString := q.Get("limit")
		// filters
		cameraFilter := q.Get("camera")
		objectFilter := q.Get("object")
		afterStr := q.Get("after") // 2025-08-08 lub RFC3339
		beforeStr := q.Get("before")

		page, err := strconv.Atoi(pageString)
		if page <= 0 || err != nil {
			page = 1
		}
		limit, err := strconv.Atoi(limitString)
		if limit <= 0 || err != nil {
			limit = 24
		}

		var afterTime, beforeTime time.Time
		if afterStr != "" {
			if t, err := parseFlexibleDate(afterStr); err == nil {
				afterTime = t
			}
		}
		if beforeStr != "" {
			if t, err := parseFlexibleDate(beforeStr); err == nil {
				beforeTime = t
			}
		}

		entries, err := os.ReadDir(config.ImageDirectory)
		if err != nil {
			logger.Error("Error reading pictures directory: %v", err)
			http.Error(w, "Unable to read pictures directory", http.StatusInternalServerError)
			return
		}

		var (
			totalSize int64
			filtered  []string
		)

		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			info, err := e.Info()
			if err != nil {
				continue
			}

			// parse filename pattern: TIMESTAMP_CAMERA_OBJECT.jpg
			// timestamp layout: 2006-01-02_15-04-05.000
			parts := strings.SplitN(strings.TrimSuffix(name, filepath.Ext(name)), "_", 3)
			var ts time.Time
			if len(parts) >= 2 { // at least timestamp + camera
				// join first two parts if object exists could have third
				ts, _ = time.Parse("2006-01-02_15-04-05.000", parts[0]+"_"+strings.Split(parts[1], "_")[0])
			}
			cameraVal := ""
			objectVal := ""
			if len(parts) >= 2 {
				cameraVal = parts[1]
			}
			if len(parts) == 3 {
				objectVal = parts[2]
			}

			if cameraFilter != "" && !strings.EqualFold(cameraVal, cameraFilter) {
				continue
			}
			if objectFilter != "" && !strings.Contains(strings.ToLower(objectVal), strings.ToLower(objectFilter)) {
				continue
			}
			if !afterTime.IsZero() && ts.Before(afterTime) {
				continue
			}
			if !beforeTime.IsZero() && ts.After(beforeTime) {
				continue
			}

			filtered = append(filtered, name)
			totalSize += info.Size()
		}

		// sort newest first (by filename timestamp prefix) -> filenames already contain timestamp increasing lexicographically
		slices.Sort(filtered)
		slices.Reverse(filtered)

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
			Filters: map[string]string{
				"camera": cameraFilter,
				"object": objectFilter,
				"after":  afterStr,
				"before": beforeStr,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(data); err != nil {
			logger.Error("Error encoding JSON response: %v", err)
		}
	}
}

// parseFlexibleDate supports YYYY-MM-DD or RFC3339
func parseFlexibleDate(v string) (time.Time, error) {
	if t, err := time.Parse("2006-01-02", v); err == nil {
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339, v); err == nil {
		return t, nil
	}
	if sec, err := strconv.ParseInt(v, 10, 64); err == nil {
		return time.Unix(sec, 0), nil
	}
	return time.Time{}, fmt.Errorf("unsupported date format: %s", v)
}

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
