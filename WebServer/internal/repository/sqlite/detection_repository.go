package sqlite

import (
	"fmt"

	"webserver/internal/models"
)

// DetectionRepository implements repository.DetectionRepository for SQLite.
type DetectionRepository struct {
	db *DB
}

// NewDetectionRepository creates a new SQLite detection repository.
func NewDetectionRepository(db *DB) *DetectionRepository {
	return &DetectionRepository{db: db}
}

// Insert adds a new detection record to the database.
func (r *DetectionRepository) Insert(det *models.Detection) (int64, error) {
	r.db.Lock()
	defer r.db.Unlock()

	result, err := r.db.Conn().Exec(`
		INSERT INTO detections (image_id, object_name, x, y, width, height, confidence)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, det.ImageID, det.ObjectName, det.X, det.Y, det.Width, det.Height, det.Confidence)
	if err != nil {
		return 0, fmt.Errorf("failed to insert detection: %w", err)
	}

	return result.LastInsertId()
}

// InsertBatch adds multiple detections in a single transaction.
func (r *DetectionRepository) InsertBatch(detections []models.Detection) error {
	r.db.Lock()
	defer r.db.Unlock()

	tx, err := r.db.Conn().Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO detections (image_id, object_name, x, y, width, height, confidence)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, det := range detections {
		if _, err := stmt.Exec(det.ImageID, det.ObjectName, det.X, det.Y, det.Width, det.Height, det.Confidence); err != nil {
			return fmt.Errorf("failed to insert detection: %w", err)
		}
	}

	return tx.Commit()
}

// GetByImageID retrieves all detections for an image.
func (r *DetectionRepository) GetByImageID(imageID int64) ([]models.Detection, error) {
	r.db.RLock()
	defer r.db.RUnlock()

	rows, err := r.db.Conn().Query(`
		SELECT id, image_id, object_name, x, y, width, height, confidence 
		FROM detections WHERE image_id = ?
	`, imageID)
	if err != nil {
		return nil, fmt.Errorf("failed to query detections: %w", err)
	}
	defer rows.Close()

	var detections []models.Detection
	for rows.Next() {
		var det models.Detection
		if err := rows.Scan(&det.ID, &det.ImageID, &det.ObjectName, &det.X, &det.Y, &det.Width, &det.Height, &det.Confidence); err != nil {
			return nil, fmt.Errorf("failed to scan detection: %w", err)
		}
		detections = append(detections, det)
	}

	return detections, nil
}

// GetObjectNamesByImageID returns just the object names for an image.
func (r *DetectionRepository) GetObjectNamesByImageID(imageID int64) ([]string, error) {
	r.db.RLock()
	defer r.db.RUnlock()

	rows, err := r.db.Conn().Query(`SELECT DISTINCT object_name FROM detections WHERE image_id = ?`, imageID)
	if err != nil {
		return nil, fmt.Errorf("failed to query object names: %w", err)
	}
	defer rows.Close()

	var objects []string
	for rows.Next() {
		var obj string
		if err := rows.Scan(&obj); err != nil {
			return nil, fmt.Errorf("failed to scan object name: %w", err)
		}
		objects = append(objects, obj)
	}

	return objects, nil
}

// GetAllObjectNames returns a list of all unique detected object names.
func (r *DetectionRepository) GetAllObjectNames() ([]string, error) {
	r.db.RLock()
	defer r.db.RUnlock()

	rows, err := r.db.Conn().Query(`SELECT DISTINCT object_name FROM detections ORDER BY object_name`)
	if err != nil {
		return nil, fmt.Errorf("failed to query objects: %w", err)
	}
	defer rows.Close()

	var objects []string
	for rows.Next() {
		var obj string
		if err := rows.Scan(&obj); err != nil {
			return nil, fmt.Errorf("failed to scan object: %w", err)
		}
		objects = append(objects, obj)
	}

	return objects, nil
}

// DeleteByImageID removes all detections for a specific image.
func (r *DetectionRepository) DeleteByImageID(imageID int64) error {
	r.db.Lock()
	defer r.db.Unlock()

	if _, err := r.db.Conn().Exec(`DELETE FROM detections WHERE image_id = ?`, imageID); err != nil {
		return fmt.Errorf("failed to delete detections: %w", err)
	}
	return nil
}
