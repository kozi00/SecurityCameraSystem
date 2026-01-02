package models

import "time"

// Image represents an image record.
type Image struct {
	ID        int64     `json:"id"`
	Filename  string    `json:"filename"`
	Camera    string    `json:"camera"`
	Timestamp time.Time `json:"timestamp"`
	FilePath  string    `json:"filepath"`
	FileSize  int64     `json:"filesize"`
}

// ImageFilter contains filtering options for querying images.
type ImageFilter struct {
	Camera     string
	Object     string
	StartDate  time.Time
	EndDate    time.Time
	TimeAfter  string
	TimeBefore string
	Limit      int
	Offset     int
}

// ImageStats contains statistics about stored images.
type ImageStats struct {
	TotalImages    int            `json:"total_images"`
	TotalSizeBytes int64          `json:"total_size_bytes"`
	PerCamera      map[string]int `json:"per_camera"`
	ObjectCounts   map[string]int `json:"object_counts"`
}
