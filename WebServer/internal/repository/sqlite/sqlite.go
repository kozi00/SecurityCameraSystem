package sqlite

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps the SQLite database connection with thread-safe access.
type DB struct {
	conn *sql.DB
	mu   sync.RWMutex
}

// New creates and initializes a new SQLite database connection.
func New(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	conn.SetMaxOpenConns(1)
	conn.SetMaxIdleConns(1)
	conn.SetConnMaxLifetime(0)

	db := &DB{conn: conn}

	if err := db.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

// migrate creates the necessary tables if they don't exist.
func (db *DB) migrate() error {
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

	CREATE TABLE IF NOT EXISTS detections (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		image_id INTEGER NOT NULL,
		object_name TEXT NOT NULL,
		x INTEGER DEFAULT 0,
		y INTEGER DEFAULT 0,
		width INTEGER DEFAULT 0,
		height INTEGER DEFAULT 0,
		confidence REAL DEFAULT 0,
		FOREIGN KEY (image_id) REFERENCES images(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_images_camera ON images(camera);
	CREATE INDEX IF NOT EXISTS idx_images_timestamp ON images(timestamp);
	CREATE INDEX IF NOT EXISTS idx_detections_object_name ON detections(object_name);
	CREATE INDEX IF NOT EXISTS idx_detections_image_id ON detections(image_id);
	`

	_, err := db.conn.Exec(schema)
	return err
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// Conn returns the underlying database connection for use by repositories.
func (db *DB) Conn() *sql.DB {
	return db.conn
}

// Lock acquires a write lock.
func (db *DB) Lock() {
	db.mu.Lock()
}

// Unlock releases the write lock.
func (db *DB) Unlock() {
	db.mu.Unlock()
}

// RLock acquires a read lock.
func (db *DB) RLock() {
	db.mu.RLock()
}

// RUnlock releases the read lock.
func (db *DB) RUnlock() {
	db.mu.RUnlock()
}
