package database

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Image represents an image record in the database.
type Image struct {
	ID        int64     `json:"id"`
	Filename  string    `json:"filename"`
	Camera    string    `json:"camera"`
	Objects   []string  `json:"objects"`
	Timestamp time.Time `json:"timestamp"`
	FilePath  string    `json:"filepath"`
	FileSize  int64     `json:"filesize"`
}

// ImageFilter contains filtering options for querying images.
type ImageFilter struct {
	Camera     string
	Object     string
	StartDate  time.Time
	EndDate    time.Time
	TimeAfter  string
	TimeBefore string
	Limit      int
	Offset     int
}

// Database handles SQLite operations.
type Database struct {
	db *sql.DB
	mu sync.RWMutex
}

// New creates and initializes a new SQLite database.
func New(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	database := &Database{db: db}

	if err := database.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return database, nil
}

// migrate creates the necessary tables if they don't exist.
func (d *Database) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS images (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		filename TEXT NOT NULL UNIQUE,
		camera TEXT NOT NULL,
		timestamp DATETIME NOT NULL,
		filepath TEXT NOT NULL,
		filesize INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS image_objects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		image_id INTEGER NOT NULL,
		object_name TEXT NOT NULL,
		FOREIGN KEY (image_id) REFERENCES images(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_images_camera ON images(camera);
	CREATE INDEX IF NOT EXISTS idx_images_timestamp ON images(timestamp);
	CREATE INDEX IF NOT EXISTS idx_image_objects_name ON image_objects(object_name);
	CREATE INDEX IF NOT EXISTS idx_image_objects_image_id ON image_objects(image_id);
	`

	_, err := d.db.Exec(schema)
	return err
}

// InsertImage adds a new image record to the database.
func (d *Database) InsertImage(img *Image) (int64, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	tx, err := d.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert image
	result, err := tx.Exec(`
		INSERT INTO images (filename, camera, timestamp, filepath, filesize)
		VALUES (?, ?, ?, ?, ?)
	`, img.Filename, img.Camera, img.Timestamp, img.FilePath, img.FileSize)
	if err != nil {
		return 0, fmt.Errorf("failed to insert image: %w", err)
	}

	imageID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	// Insert objects
	for _, obj := range img.Objects {
		_, err := tx.Exec(`
			INSERT INTO image_objects (image_id, object_name)
			VALUES (?, ?)
		`, imageID, obj)
		if err != nil {
			return 0, fmt.Errorf("failed to insert object: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return imageID, nil
}

// GetImages retrieves images based on filter criteria.
func (d *Database) GetImages(filter *ImageFilter) ([]Image, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	query := `
		SELECT DISTINCT i.id, i.filename, i.camera, i.timestamp, i.filepath, i.filesize
		FROM images i
		LEFT JOIN image_objects o ON i.id = o.image_id
		WHERE 1=1
	`
	args := []interface{}{}

	if filter.Camera != "" {
		query += " AND i.camera = ?"
		args = append(args, filter.Camera)
	}

	if filter.Object != "" {
		query += " AND o.object_name = ?"
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

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query images: %w", err)
	}
	defer rows.Close()

	var images []Image
	for rows.Next() {
		var img Image
		err := rows.Scan(&img.ID, &img.Filename, &img.Camera, &img.Timestamp, &img.FilePath, &img.FileSize)
		if err != nil {
			return nil, fmt.Errorf("failed to scan image: %w", err)
		}

		// Get objects for this image
		objects, err := d.getImageObjects(img.ID)
		if err != nil {
			return nil, err
		}
		img.Objects = objects

		images = append(images, img)
	}

	return images, nil
}

// GetTotalCount returns the total count of images matching the filter (without limit/offset).
func (d *Database) GetTotalCount(filter *ImageFilter) (int, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	query := `
		SELECT COUNT(DISTINCT i.id)
		FROM images i
		LEFT JOIN image_objects o ON i.id = o.image_id
		WHERE 1=1
	`
	args := []interface{}{}

	if filter.Camera != "" {
		query += " AND i.camera = ?"
		args = append(args, filter.Camera)
	}

	if filter.Object != "" {
		query += " AND o.object_name = ?"
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
	err := d.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count images: %w", err)
	}

	return count, nil
}

// getImageObjects retrieves all objects for a given image ID.
func (d *Database) getImageObjects(imageID int64) ([]string, error) {
	rows, err := d.db.Query(`
		SELECT object_name FROM image_objects WHERE image_id = ?
	`, imageID)
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

// GetCameras returns a list of all unique camera names.
func (d *Database) GetCameras() ([]string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	rows, err := d.db.Query(`SELECT DISTINCT camera FROM images ORDER BY camera`)
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

// GetObjects returns a list of all unique detected objects.
func (d *Database) GetObjects() ([]string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	rows, err := d.db.Query(`SELECT DISTINCT object_name FROM image_objects ORDER BY object_name`)
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

// DeleteImage removes an image record from the database.
func (d *Database) DeleteImage(id int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// First delete related objects
	_, err := d.db.Exec(`DELETE FROM image_objects WHERE image_id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete image objects: %w", err)
	}

	_, err = d.db.Exec(`DELETE FROM images WHERE id = ?`, id)
	return err
}

// DeleteImageByFilename removes an image by its filename.
func (d *Database) DeleteImageByFilename(filename string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Get image ID first
	var imageID int64
	err := d.db.QueryRow(`SELECT id FROM images WHERE filename = ?`, filename).Scan(&imageID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil // Image not found, nothing to delete
		}
		return fmt.Errorf("failed to get image id: %w", err)
	}

	// Delete related objects
	_, err = d.db.Exec(`DELETE FROM image_objects WHERE image_id = ?`, imageID)
	if err != nil {
		return fmt.Errorf("failed to delete image objects: %w", err)
	}

	_, err = d.db.Exec(`DELETE FROM images WHERE id = ?`, imageID)
	return err
}

// GetImageByFilename retrieves an image by its filename.
func (d *Database) GetImageByFilename(filename string) (*Image, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var img Image
	err := d.db.QueryRow(`
		SELECT id, filename, camera, timestamp, filepath, filesize
		FROM images WHERE filename = ?
	`, filename).Scan(&img.ID, &img.Filename, &img.Camera, &img.Timestamp, &img.FilePath, &img.FileSize)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query image: %w", err)
	}

	objects, err := d.getImageObjects(img.ID)
	if err != nil {
		return nil, err
	}
	img.Objects = objects

	return &img, nil
}

// GetStats returns statistics about stored images.
func (d *Database) GetStats() (map[string]interface{}, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	stats := make(map[string]interface{})

	// Total images count
	var totalImages int
	err := d.db.QueryRow(`SELECT COUNT(*) FROM images`).Scan(&totalImages)
	if err != nil {
		return nil, err
	}
	stats["total_images"] = totalImages

	// Total size
	var totalSize int64
	err = d.db.QueryRow(`SELECT COALESCE(SUM(filesize), 0) FROM images`).Scan(&totalSize)
	if err != nil {
		return nil, err
	}
	stats["total_size_bytes"] = totalSize

	// Images per camera
	rows, err := d.db.Query(`SELECT camera, COUNT(*) FROM images GROUP BY camera`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	perCamera := make(map[string]int)
	for rows.Next() {
		var camera string
		var count int
		if err := rows.Scan(&camera, &count); err != nil {
			return nil, err
		}
		perCamera[camera] = count
	}
	stats["per_camera"] = perCamera

	// Most detected objects
	objectRows, err := d.db.Query(`
		SELECT object_name, COUNT(*) as cnt 
		FROM image_objects 
		GROUP BY object_name 
		ORDER BY cnt DESC 
		LIMIT 10
	`)
	if err != nil {
		return nil, err
	}
	defer objectRows.Close()

	objectCounts := make(map[string]int)
	for objectRows.Next() {
		var obj string
		var count int
		if err := objectRows.Scan(&obj, &count); err != nil {
			return nil, err
		}
		objectCounts[obj] = count
	}
	stats["object_counts"] = objectCounts

	return stats, nil
}

// Close closes the database connection.
func (d *Database) Close() error {
	return d.db.Close()
}

// ImageExists checks if an image with the given filename already exists.
func (d *Database) ImageExists(filename string) (bool, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var count int
	err := d.db.QueryRow(`SELECT COUNT(*) FROM images WHERE filename = ?`, filename).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// BulkInsertImages inserts multiple images in a single transaction.
func (d *Database) BulkInsertImages(images []Image) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	imgStmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO images (filename, camera, timestamp, filepath, filesize)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare image statement: %w", err)
	}
	defer imgStmt.Close()

	objStmt, err := tx.Prepare(`
		INSERT INTO image_objects (image_id, object_name)
		VALUES (?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare object statement: %w", err)
	}
	defer objStmt.Close()

	for _, img := range images {
		result, err := imgStmt.Exec(img.Filename, img.Camera, img.Timestamp, img.FilePath, img.FileSize)
		if err != nil {
			continue // Skip duplicates
		}

		imageID, err := result.LastInsertId()
		if err != nil || imageID == 0 {
			continue
		}

		for _, obj := range img.Objects {
			objStmt.Exec(imageID, obj)
		}
	}

	return tx.Commit()
}

// ClearAll removes all images and objects from the database.
func (d *Database) ClearAll() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.db.Exec(`DELETE FROM image_objects`)
	if err != nil {
		return fmt.Errorf("failed to delete image objects: %w", err)
	}

	_, err = d.db.Exec(`DELETE FROM images`)
	if err != nil {
		return fmt.Errorf("failed to delete images: %w", err)
	}

	return nil
}

// ParseFilename extracts metadata from legacy filename format.
// Format: 2006-01-02_15-04_05.000_CameraName_Object1_Object2_.jpg
func ParseFilename(filename string) (timestamp time.Time, camera string, objects []string, err error) {
	// Remove .jpg extension
	name := strings.TrimSuffix(filename, ".jpg")
	parts := strings.Split(name, "_")

	if len(parts) < 4 {
		return time.Time{}, "", nil, fmt.Errorf("invalid filename format: %s", filename)
	}

	// Parse timestamp (first 3 parts: date, time, seconds.milliseconds)
	timeStr := parts[0] + "_" + parts[1] + "_" + parts[2]
	timestamp, err = time.Parse("2006-01-02_15-04_05.000", timeStr)
	if err != nil {
		return time.Time{}, "", nil, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	// Camera name is the 4th part
	camera = parts[3]

	// Objects are remaining parts (excluding last empty one if exists)
	for i := 4; i < len(parts); i++ {
		if parts[i] != "" {
			objects = append(objects, parts[i])
		}
	}

	return timestamp, camera, objects, nil
}
