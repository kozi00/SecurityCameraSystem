package sqlite

import (
	"database/sql"
	"fmt"

	"webserver/internal/dto"
	"webserver/internal/model"
)

// ImageRepository implements repository.ImageRepository for SQLite.
type ImageRepository struct {
	db *DB
}

// NewImageRepository creates a new SQLite image repository.
func NewImageRepository(db *DB) *ImageRepository {
	return &ImageRepository{db: db}
}

// Insert adds a new image record to the database.
func (r *ImageRepository) Insert(img *model.Image) (int64, error) {
	r.db.Lock()
	defer r.db.Unlock()

	result, err := r.db.Conn().Exec(`
		INSERT INTO images (filename, camera, timestamp, filepath, filesize)
		VALUES (?, ?, ?, ?, ?)
	`, img.Filename, img.Camera, img.Timestamp, img.FilePath, img.FileSize)
	if err != nil {
		return 0, fmt.Errorf("failed to insert image: %w", err)
	}

	return result.LastInsertId()
}

// GetByID retrieves an image by its ID.
func (r *ImageRepository) GetByID(id int64) (*model.Image, error) {
	r.db.RLock()
	defer r.db.RUnlock()

	var img model.Image
	err := r.db.Conn().QueryRow(`
		SELECT id, filename, camera, timestamp, filepath, filesize 
		FROM images WHERE id = ?
	`, id).Scan(&img.ID, &img.Filename, &img.Camera, &img.Timestamp, &img.FilePath, &img.FileSize)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %w", err)
	}
	return &img, nil
}

// GetByFilename retrieves an image by its filename.
func (r *ImageRepository) GetByFilename(filename string) (*model.Image, error) {
	r.db.RLock()
	defer r.db.RUnlock()

	var img model.Image
	err := r.db.Conn().QueryRow(`
		SELECT id, filename, camera, timestamp, filepath, filesize 
		FROM images WHERE filename = ?
	`, filename).Scan(&img.ID, &img.Filename, &img.Camera, &img.Timestamp, &img.FilePath, &img.FileSize)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %w", err)
	}
	return &img, nil
}

// GetAll retrieves images based on filter criteria.
func (r *ImageRepository) GetAll(filter *dto.ImageFilters) ([]model.Image, error) {
	r.db.RLock()
	defer r.db.RUnlock()

	query := `
		SELECT DISTINCT i.id, i.filename, i.camera, i.timestamp, i.filepath, i.filesize
		FROM images i
		LEFT JOIN detections d ON i.id = d.image_id
		WHERE 1=1
	`
	args := []interface{}{}

	if filter.Camera != "" {
		query += " AND i.camera = ?"
		args = append(args, filter.Camera)
	}

	if filter.Object != "" {
		query += " AND d.object_name = ?"
		args = append(args, filter.Object)
	}

	if !filter.DateAfter.IsZero() {
		query += " AND DATE(i.timestamp) >= DATE(?)"
		args = append(args, filter.DateAfter)
	}

	if !filter.DateBefore.IsZero() {
		query += " AND DATE(i.timestamp) <= DATE(?)"
		args = append(args, filter.DateBefore)
	}

	if !filter.TimeAfter.IsZero() {
		query += " AND TIME(i.timestamp) >= TIME(?)"
		args = append(args, filter.TimeAfter.Format("15:04:05"))
	}

	if !filter.TimeBefore.IsZero() {
		query += " AND TIME(i.timestamp) <= TIME(?)"
		args = append(args, filter.TimeBefore.Format("15:04:05"))
	}

	query += " ORDER BY i.timestamp DESC"

	rows, err := r.db.Conn().Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query images: %w", err)
	}
	defer rows.Close()

	var images []model.Image
	for rows.Next() {
		var img model.Image
		if err := rows.Scan(&img.ID, &img.Filename, &img.Camera, &img.Timestamp, &img.FilePath, &img.FileSize); err != nil {
			return nil, fmt.Errorf("failed to scan image: %w", err)
		}
		images = append(images, img)
	}

	return images, nil
}

// GetTotalCount returns the total number of images matching the filter criteria.
func (r *ImageRepository) GetTotalCount(filter *dto.ImageFilters) (int, error) {
	r.db.RLock()
	defer r.db.RUnlock()

	query := `
		SELECT COUNT(DISTINCT i.id)
		FROM images i
		LEFT JOIN detections d ON i.id = d.image_id
		WHERE 1=1
	`
	args := []interface{}{}

	if filter.Camera != "" {
		query += " AND i.camera = ?"
		args = append(args, filter.Camera)
	}

	if filter.Object != "" {
		query += " AND d.object_name = ?"
		args = append(args, filter.Object)
	}

	if !filter.DateAfter.IsZero() {
		query += " AND DATE(i.timestamp) >= DATE(?)"
		args = append(args, filter.DateAfter)
	}

	if !filter.DateBefore.IsZero() {
		query += " AND DATE(i.timestamp) <= DATE(?)"
		args = append(args, filter.DateBefore)
	}

	if !filter.TimeAfter.IsZero() {
		query += " AND TIME(i.timestamp) >= TIME(?)"
		args = append(args, filter.TimeAfter.Format("15:04:05"))
	}

	if !filter.TimeBefore.IsZero() {
		query += " AND TIME(i.timestamp) <= TIME(?)"
		args = append(args, filter.TimeBefore.Format("15:04:05"))
	}

	var count int
	err := r.db.Conn().QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count images: %w", err)
	}

	return count, nil
}

func (r *ImageRepository) GetDirectorySize() (int64, error) {
	r.db.RLock()
	defer r.db.RUnlock()

	var totalSize int64
	err := r.db.Conn().QueryRow(`SELECT SUM(filesize) FROM images`).Scan(&totalSize)
	if err != nil {
		return 0, fmt.Errorf("failed to get directory size: %w", err)
	}
	return totalSize, nil
}

// Delete removes an image by its ID.
func (r *ImageRepository) Delete(id int64) error {
	r.db.Lock()
	defer r.db.Unlock()

	// First delete related detections
	if _, err := r.db.Conn().Exec(`DELETE FROM detections WHERE image_id = ?`, id); err != nil {
		return fmt.Errorf("failed to delete detections: %w", err)
	}

	if _, err := r.db.Conn().Exec(`DELETE FROM images WHERE id = ?`, id); err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}
	return nil
}

// DeleteByFilename removes an image by its filename.
func (r *ImageRepository) DeleteByFilename(filename string) error {
	r.db.Lock()
	defer r.db.Unlock()

	// Get image ID first
	var imageID int64
	err := r.db.Conn().QueryRow(`SELECT id FROM images WHERE filename = ?`, filename).Scan(&imageID)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to get image id: %w", err)
	}

	// Delete related detections
	if _, err := r.db.Conn().Exec(`DELETE FROM detections WHERE image_id = ?`, imageID); err != nil {
		return fmt.Errorf("failed to delete detections: %w", err)
	}

	if _, err := r.db.Conn().Exec(`DELETE FROM images WHERE id = ?`, imageID); err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}
	return nil
}

// DeleteAll removes all images and their detections.
func (r *ImageRepository) DeleteAll() error {
	r.db.Lock()
	defer r.db.Unlock()

	if _, err := r.db.Conn().Exec(`DELETE FROM detections`); err != nil {
		return fmt.Errorf("failed to delete detections: %w", err)
	}

	if _, err := r.db.Conn().Exec(`DELETE FROM images`); err != nil {
		return fmt.Errorf("failed to delete images: %w", err)
	}

	return nil
}
