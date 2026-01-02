package models

// Detection represents a detected object in an image.
type Detection struct {
	ID         int64   `json:"id"`
	ImageID    int64   `json:"image_id"`
	ObjectName string  `json:"object_name"`
	X          int     `json:"x"`
	Y          int     `json:"y"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	Confidence float64 `json:"confidence"`
}
