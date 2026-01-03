// PicturesData is a paginated response payload for the pictures gallery.
package dto

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
