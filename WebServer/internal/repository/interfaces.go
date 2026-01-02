package repository

import "webserver/internal/models"

// ImageRepository defines the interface for image data operations.
type ImageRepository interface {
	// Create operations
	Insert(img *models.Image) (int64, error)

	// Read operations
	GetByID(id int64) (*models.Image, error)
	GetByFilename(filename string) (*models.Image, error)
	GetAll(filter *models.ImageFilter) ([]models.Image, error)
	GetTotalCount(filter *models.ImageFilter) (int, error)
	Exists(filename string) (bool, error)
	GetCameras() ([]string, error)
	GetStats() (*models.ImageStats, error)

	// Delete operations
	Delete(id int64) error
	DeleteByFilename(filename string) error
	DeleteAll() error
}

// DetectionRepository defines the interface for detection data operations.
type DetectionRepository interface {
	// Create operations
	Insert(det *models.Detection) (int64, error)
	InsertBatch(detections []models.Detection) error

	// Read operations
	GetByImageID(imageID int64) ([]models.Detection, error)
	GetObjectNamesByImageID(imageID int64) ([]string, error)
	GetAllObjectNames() ([]string, error)

	// Delete operations
	DeleteByImageID(imageID int64) error
}
