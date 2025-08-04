package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
)

type PicturesData struct {
	Pictures    []string `json:"pictures"`
	Size        int64    `json:"size"`
	Length      int      `json:"length"`
	TotalPages  int      `json:"totalPages"`
	CurrentPage int      `json:"currentPage"`
	Limit       int      `json:"pageSize"`
}

func DisplayPicturesHandler(imagesDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		pageString := r.URL.Query().Get("page")
		limitString := r.URL.Query().Get("limit")

		page, err := strconv.Atoi(pageString)
		if page <= 0 || err != nil {
			page = 1
		}
		limit, err := strconv.Atoi(limitString)
		if limit <= 0 || err != nil {
			limit = 10
		}

		files, err := os.ReadDir(imagesDir)
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
			Length:      len(pictureFiles),
			TotalPages:  (len(pictureFiles) + limit - 1) / limit,
			CurrentPage: page,
			Limit:       limit,
		}
		json.NewEncoder(w).Encode(data)
	}
}

func ViewPictureHandler(imagesDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		image := r.URL.Query().Get("image")
		http.ServeFile(w, r, imagesDir+image)
	}
}
