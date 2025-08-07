package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"webserver/internal/config"
	"webserver/internal/logger"
)

type PicturesData struct {
	Pictures    []string `json:"pictures"`
	Size        int64    `json:"size"`
	MaxSize     int64    `json:"maxSize"`
	Length      int      `json:"length"`
	TotalPages  int      `json:"totalPages"`
	CurrentPage int      `json:"currentPage"`
	Limit       int      `json:"pageSize"`
}

func DisplayPicturesHandler(config *config.Config, logger *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		pageString := r.URL.Query().Get("page")
		limitString := r.URL.Query().Get("limit")

		page, err := strconv.Atoi(pageString)
		if page <= 0 || err != nil { //przykladowe wartosci domyslne w przypadku bledow
			page = 1
		}
		limit, err := strconv.Atoi(limitString)
		if limit <= 0 || err != nil {
			limit = 10
		}

		files, err := os.ReadDir(config.ImageDirectory)
		if err != nil {
			http.Error(w, "Unable to read pictures directory", http.StatusInternalServerError)
			return
		}
		var totalSize int64 = 0
		var pictureFiles []string

		for _, file := range files {
			if !file.IsDir() {
				info, err := file.Info()
				if err == nil {
					pictureFiles = append(pictureFiles, file.Name())
					totalSize += info.Size()
				}
			}
		}

		start := (page - 1) * limit
		end := start + limit
		if start > len(pictureFiles) {
			start = len(pictureFiles)
		}
		if end > len(pictureFiles) {
			end = len(pictureFiles)
		}

		paginated := pictureFiles[start:end]
		data := PicturesData{
			Pictures:    paginated,
			Size:        totalSize,
			MaxSize:     config.MaxImageDirectorySize,
			Length:      len(pictureFiles),
			TotalPages:  (len(pictureFiles) + limit - 1) / limit,
			CurrentPage: page,
			Limit:       limit,
		}
		err = json.NewEncoder(w).Encode(data)
		if err != nil {
			logger.Error("Error encoding JSON response: %v", err)
		}
	}
}

func ViewPictureHandler(config *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		image := r.URL.Query().Get("image")
		http.ServeFile(w, r, config.ImageDirectory+image)
	}
}
