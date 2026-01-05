package tests

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"webserver/internal/dto"
	"webserver/internal/model"
	"webserver/internal/repository/sqlite"
)

// ========================================
// Database Integration Tests
// ========================================

func TestDatabase_Connection(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "db_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Verify database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file should exist")
	}
}

func TestDatabase_Migration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "db_migrate_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Verify tables exist by inserting data
	imageRepo := sqlite.NewImageRepository(db)
	detectionRepo := sqlite.NewDetectionRepository(db)

	img := &model.Image{
		Filename:  "migration_test.jpg",
		Camera:    "cam1",
		Timestamp: time.Now(),
		FilePath:  "/images/migration_test.jpg",
		FileSize:  1024,
	}

	imageID, err := imageRepo.Insert(img)
	if err != nil {
		t.Fatalf("Failed to insert into images table: %v", err)
	}

	det := &model.Detection{
		ImageID:    imageID,
		ObjectName: "person",
		Confidence: 0.9,
	}

	_, err = detectionRepo.Insert(det)
	if err != nil {
		t.Fatalf("Failed to insert into detections table: %v", err)
	}
}

func TestDatabase_ConcurrentAccess(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "db_concurrent_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	imageRepo := sqlite.NewImageRepository(db)

	// Concurrent inserts
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			img := &model.Image{
				Filename:  "concurrent_" + string(rune('a'+idx)) + ".jpg",
				Camera:    "cam1",
				Timestamp: time.Now(),
				FilePath:  "/images/",
				FileSize:  100,
			}
			_, err := imageRepo.Insert(img)
			if err != nil {
				t.Errorf("Concurrent insert %d failed: %v", idx, err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all inserts succeeded
	count, _ := imageRepo.GetTotalCount(&dto.ImageFilters{})
	if count != 10 {
		t.Errorf("Expected 10 images, got %d", count)
	}
}

func TestDatabase_ForeignKeyConstraint(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "db_fk_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	imageRepo := sqlite.NewImageRepository(db)
	detectionRepo := sqlite.NewDetectionRepository(db)

	// Insert image with detections
	img := &model.Image{
		Filename:  "fk_test.jpg",
		Camera:    "cam1",
		Timestamp: time.Now(),
		FilePath:  "/images/fk_test.jpg",
		FileSize:  1024,
	}
	imageID, _ := imageRepo.Insert(img)

	detections := []model.Detection{
		{ImageID: imageID, ObjectName: "person", Confidence: 0.9},
		{ImageID: imageID, ObjectName: "car", Confidence: 0.85},
	}
	detectionRepo.InsertBatch(detections)

	// Delete image - detections should be cascade deleted
	imageRepo.Delete(imageID)

	// Verify detections are gone
	retrieved, _ := detectionRepo.GetByImageID(imageID)
	if len(retrieved) != 0 {
		t.Errorf("Expected 0 detections after cascade delete, got %d", len(retrieved))
	}
}

// ========================================
// Image with Detections Full Flow Test
// ========================================

func TestFullFlow_ImageWithDetections(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "full_flow_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	imageRepo := sqlite.NewImageRepository(db)
	detectionRepo := sqlite.NewDetectionRepository(db)

	// Step 1: Create image
	timestamp := time.Now()
	img := &model.Image{
		Filename:  "full_flow_test.jpg",
		Camera:    "front_door",
		Timestamp: timestamp,
		FilePath:  "/images/full_flow_test.jpg",
		FileSize:  2048,
	}

	imageID, err := imageRepo.Insert(img)
	if err != nil {
		t.Fatalf("Step 1 - Insert image failed: %v", err)
	}
	t.Logf("Step 1: Created image with ID %d", imageID)

	// Step 2: Add detections
	detections := []model.Detection{
		{ImageID: imageID, ObjectName: "person", X: 100, Y: 50, Width: 150, Height: 300, Confidence: 0.95},
		{ImageID: imageID, ObjectName: "car", X: 400, Y: 200, Width: 200, Height: 100, Confidence: 0.87},
		{ImageID: imageID, ObjectName: "dog", X: 250, Y: 350, Width: 80, Height: 60, Confidence: 0.72},
	}

	err = detectionRepo.InsertBatch(detections)
	if err != nil {
		t.Fatalf("Step 2 - Insert detections failed: %v", err)
	}
	t.Logf("Step 2: Added %d detections", len(detections))

	// Step 3: Retrieve image
	retrievedImg, err := imageRepo.GetByID(imageID)
	if err != nil {
		t.Fatalf("Step 3 - Get image failed: %v", err)
	}

	if retrievedImg.Camera != "front_door" {
		t.Errorf("Step 3 - Camera mismatch: expected front_door, got %s", retrievedImg.Camera)
	}
	t.Logf("Step 3: Retrieved image - Camera: %s, Size: %d bytes", retrievedImg.Camera, retrievedImg.FileSize)

	// Step 4: Get object names
	objects, err := detectionRepo.GetObjectNamesByImageID(imageID)
	if err != nil {
		t.Fatalf("Step 4 - Get objects failed: %v", err)
	}

	if len(objects) != 3 {
		t.Errorf("Step 4 - Expected 3 objects, got %d", len(objects))
	}
	t.Logf("Step 4: Retrieved objects: %v", objects)

	// Step 5: Get full detections
	fullDetections, err := detectionRepo.GetByImageID(imageID)
	if err != nil {
		t.Fatalf("Step 5 - Get full detections failed: %v", err)
	}

	for _, det := range fullDetections {
		t.Logf("Step 5: Detection - %s at (%d,%d) with confidence %.2f",
			det.ObjectName, det.X, det.Y, det.Confidence)
	}

	// Step 6: Filter by camera
	filtered, err := imageRepo.GetAll(&dto.ImageFilters{Camera: "front_door"})
	if err != nil {
		t.Fatalf("Step 6 - Filter by camera failed: %v", err)
	}

	if len(filtered) != 1 {
		t.Errorf("Step 6 - Expected 1 filtered image, got %d", len(filtered))
	}
	t.Logf("Step 6: Filter by camera returned %d images", len(filtered))

	// Step 7: Clean up
	err = imageRepo.Delete(imageID)
	if err != nil {
		t.Fatalf("Step 7 - Delete image failed: %v", err)
	}

	// Verify deletion
	count, _ := imageRepo.GetTotalCount(&dto.ImageFilters{})
	if count != 0 {
		t.Errorf("Step 7 - Expected 0 images after delete, got %d", count)
	}
	t.Logf("Step 7: Successfully deleted image and all related detections")
}
