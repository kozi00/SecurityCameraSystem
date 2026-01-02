package sqlite

import (
	"database/sql"
	"fmt"

	"webserver/internal/models"
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
func (r *ImageRepository) Insert(img *models.Image) (int64, error) {
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
func (r *ImageRepository) GetByID(id int64) (*models.Image, error) {
	r.db.RLock()
	defer r.db.RUnlock()

	var img models.Image
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
func (r *ImageRepository) GetByFilename(filename string) (*models.Image, error) {
	r.db.RLock()
	defer r.db.RUnlock()

	var img models.Image
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
func (r *ImageRepository) GetAll(filter *models.ImageFilter) ([]models.Image, error) {
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

	if !filter.StartDate.IsZero() {
		query += " AND DATE(i.timestamp) >= DATE(?)"
		args = append(args, filter.StartDate)
	}

	if !filter.EndDate.IsZero() {
		query += " AND DATE(i.timestamp) <= DATE(?)"
		args = append(args, filter.EndDate)
	}

	if filter.TimeAfter != "" {
		query += " AND TIME(i.timestamp) >= TIME(?)"
		args = append(args, filter.TimeAfter)
	}

	if filter.TimeBefore != "" {
		query += " AND TIME(i.timestamp) <= TIME(?)"
		args = append(args, filter.TimeBefore)
	}

	query += " ORDER BY i.timestamp DESC"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := r.db.Conn().Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query images: %w", err)
	}
	defer rows.Close()

	var images []models.Image
	for rows.Next() {
		var img models.Image
		if err := rows.Scan(&img.ID, &img.Filename, &img.Camera, &img.Timestamp, &img.FilePath, &img.FileSize); err != nil {
			return nil, fmt.Errorf("failed to scan image: %w", err)
		}
		images = append(images, img)
	}

	return images, nil
}

// GetTotalCount returns the total count of images matching the filter.
func (r *ImageRepository) GetTotalCount(filter *models.ImageFilter) (int, error) {
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

	if !filter.StartDate.IsZero() {
		query += " AND DATE(i.timestamp) >= DATE(?)"
		args = append(args, filter.StartDate)
	}

	if !filter.EndDate.IsZero() {
		query += " AND DATE(i.timestamp) <= DATE(?)"
		args = append(args, filter.EndDate)
	}

	if filter.TimeAfter != "" {
		query += " AND TIME(i.timestamp) >= TIME(?)"
		args = append(args, filter.TimeAfter)
	}

	if filter.TimeBefore != "" {
		query += " AND TIME(i.timestamp) <= TIME(?)"
		args = append(args, filter.TimeBefore)
	}

	var count int
	if err := r.db.Conn().QueryRow(query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count images: %w", err)
	}

	return count, nil
}

// Exists checks if an image with the given filename exists.
func (r *ImageRepository) Exists(filename string) (bool, error) {
	r.db.RLock()
	defer r.db.RUnlock()

	var count int
	err := r.db.Conn().QueryRow(`SELECT COUNT(*) FROM images WHERE filename = ?`, filename).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check image existence: %w", err)
	}
	return count > 0, nil
}

// GetCameras returns a list of unique camera names.
func (r *ImageRepository) GetCameras() ([]string, error) {
	r.db.RLock()
	defer r.db.RUnlock()

	rows, err := r.db.Conn().Query(`SELECT DISTINCT camera FROM images ORDER BY camera`)
	if err != nil {
		return nil, fmt.Errorf("failed to query cameras: %w", err)
	}
	defer rows.Close()

	var cameras []string
	for rows.Next() {
		var camera string
		if err := rows.Scan(&camera); err != nil {
			return nil, fmt.Errorf("failed to scan camera: %w", err)
		}
		cameras = append(cameras, camera)
	}
	return cameras, nil
}

// GetStats returns statistics about stored images.
func (r *ImageRepository) GetStats() (*models.ImageStats, error) {
	r.db.RLock()
	defer r.db.RUnlock()

	stats := &models.ImageStats{
		PerCamera:    make(map[string]int),
		ObjectCounts: make(map[string]int),
	}

	// Total images count
	if err := r.db.Conn().QueryRow(`SELECT COUNT(*) FROM images`).Scan(&stats.TotalImages); err != nil {
		return nil, err
	}

	// Total size
	if err := r.db.Conn().QueryRow(`SELECT COALESCE(SUM(filesize), 0) FROM images`).Scan(&stats.TotalSizeBytes); err != nil {
		return nil, err
	}

	// Images per camera
	rows, err := r.db.Conn().Query(`SELECT camera, COUNT(*) FROM images GROUP BY camera`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var camera string
		var count int
		if err := rows.Scan(&camera, &count); err != nil {
			return nil, err
		}
		stats.PerCamera[camera] = count
	}

	// Most detected objects
	objectRows, err := r.db.Conn().Query(`
		SELECT object_name, COUNT(*) as cnt 
		FROM detections 
		GROUP BY object_name 
		ORDER BY cnt DESC 
		LIMIT 10
	`)
	if err != nil {
		return nil, err
	}
	defer objectRows.Close()

	for objectRows.Next() {
		var obj string
		var count int
		if err := objectRows.Scan(&obj, &count); err != nil {
			return nil, err
		}
		stats.ObjectCounts[obj] = count
	}

	return stats, nil
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
