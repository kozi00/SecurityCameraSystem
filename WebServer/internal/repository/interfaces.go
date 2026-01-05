package repository

import (
	"webserver/internal/dto"
	"webserver/internal/model"
)

// ImageRepository defines the interface for image data operations.
type ImageRepository interface {
	// Create operations
	Insert(img *model.Image) (int64, error)

	// Read operations
	GetByID(id int64) (*model.Image, error)
	GetByFilename(filename string) (*model.Image, error)
	GetAll(filter *dto.ImageFilters) ([]model.Image, error)
	GetTotalCount(filter *dto.ImageFilters) (int, error)
	GetDirectorySize() (int64, error)

	// Delete operations
	Delete(id int64) error
	DeleteByFilename(filename string) error
	DeleteAll() error
}

// DetectionRepository defines the interface for detection data operations.
type DetectionRepository interface {
	// Create operations
	Insert(det *model.Detection) (int64, error)
	InsertBatch(detections []model.Detection) error

	// Read operations
	GetByImageID(imageID int64) ([]model.Detection, error)
	GetObjectNamesByImageID(imageID int64) ([]string, error)
	GetAllObjectNames() ([]string, error)

	// Delete operations
	DeleteByImageID(imageID int64) error
}
