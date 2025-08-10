package handlers

import (
	"encoding/json"
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

		cameraFilter := q.Get("camera")
		objectFilter := q.Get("object")
		timeAfterStr := q.Get("timeAfter")
		timeBeforeStr := q.Get("timeBefore")
		dateAfterStr := q.Get("dateAfter")
		dateBeforeStr := q.Get("dateBefore")

		page, err := strconv.Atoi(pageString)
		if page <= 0 || err != nil {
			page = 1
		}
		limit, err := strconv.Atoi(limitString)
		if limit <= 0 || err != nil {
			limit = 24
		}

		var dateAfter, dateBefore time.Time
		if dateAfterStr != "" {
			dateAfter, _ = time.Parse("2006-01-02", dateAfterStr)
		}
		if dateBeforeStr != "" {
			dateBefore, _ = time.Parse("2006-01-02", dateBeforeStr)
		}

		var timeAfter, timeBefore time.Time
		if timeAfterStr != "" {
			timeAfter, _ = time.Parse("15:04", timeAfterStr)
		}
		if timeBeforeStr != "" {
			timeBefore, _ = time.Parse("15:04", timeBeforeStr)
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
			// timestamp layout: 2006-01-02_15-04_05.000
			// 2025-08-07 _ 13-34 _ 01.681 _ drzwi _ osoba
			parts := strings.SplitN(strings.TrimSuffix(name, filepath.Ext(name)), "_", 5)
			if len(parts) != 5 {
				logger.Warning("Skipping file with unexpected name format: %s", name)
				continue
			}
			date, _ := time.Parse("2006-01-02", parts[0])
			time, _ := time.Parse("15-04", parts[1])
			cameraVal := parts[3]
			objectVal := parts[4]

			if cameraFilter != "" && !strings.EqualFold(cameraVal, cameraFilter) {
				continue
			}
			if objectFilter != "" && !strings.EqualFold(objectVal, objectFilter) {
				continue
			}
			if !dateAfter.IsZero() && date.Before(dateAfter) {
				continue
			}
			if !dateBefore.IsZero() && date.After(dateBefore) {
				continue
			}
			if !timeAfter.IsZero() && time.Before(timeAfter) {
				continue
			}
			if !timeBefore.IsZero() && time.After(timeBefore) {
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
				"camera":     cameraFilter,
				"object":     objectFilter,
				"dateAfter":  dateAfterStr,
				"dateBefore": dateBeforeStr,
				"timeAfter":  timeAfterStr,
				"timeBefore": timeBeforeStr,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(data); err != nil {
			logger.Error("Error encoding JSON response: %v", err)
		}
	}
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
