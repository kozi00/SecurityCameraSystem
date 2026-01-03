// ImagesData is a paginated response payload for the images gallery.
package dto

type ImagesData struct {
	Images      []ImageInfo `json:"images"`
	ImagesDir   string      `json:"imagesDir"`
	Size        int64       `json:"size"`
	MaxSize     int64       `json:"maxSize"`
	Length      int         `json:"length"`
	TotalPages  int         `json:"totalPages"`
	CurrentPage int         `json:"currentPage"`
	Limit       int         `json:"pageSize"`
}
