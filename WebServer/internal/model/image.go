package model

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
