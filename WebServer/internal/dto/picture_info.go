package dto

import (
	"encoding/json"
	"time"
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
