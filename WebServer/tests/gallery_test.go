package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"webserver/internal/config"
	"webserver/internal/dto"
	"webserver/internal/model"
	"webserver/internal/repository/sqlite"
)

// ========================================
// Test Setup Helpers
// ========================================

func setupTestDB(t *testing.T) (*sqlite.DB, func()) {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "gallery_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.New(dbPath)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create test database: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.RemoveAll(tempDir)
	}

	return db, cleanup
}

func setupTestConfig(t *testing.T) (*config.Config, string, func()) {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "gallery_images")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	cfg := &config.Config{
		ImageDirectory: tempDir,
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return cfg, tempDir, cleanup
}

func createTestImageFile(t *testing.T, dir, filename string, content []byte) string {
	t.Helper()

	if content == nil {
		content = []byte("fake image data for testing purposes")
	}

	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	return path
}

// ========================================
// Image Repository Tests
// ========================================

func TestImageRepository_Insert(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := sqlite.NewImageRepository(db)

	img := &model.Image{
		Filename:  "test_image.jpg",
		Camera:    "cam1",
		Timestamp: time.Now(),
		FilePath:  "/images/test_image.jpg",
		FileSize:  1024,
	}

	id, err := repo.Insert(img)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	if id <= 0 {
		t.Errorf("Expected positive ID, got %d", id)
	}
}

func TestImageRepository_Insert_DuplicateFilename(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := sqlite.NewImageRepository(db)

	img := &model.Image{
		Filename:  "duplicate.jpg",
		Camera:    "cam1",
		Timestamp: time.Now(),
		FilePath:  "/images/duplicate.jpg",
		FileSize:  1024,
	}

	_, err := repo.Insert(img)
	if err != nil {
		t.Fatalf("First insert failed: %v", err)
	}

	// Try to insert with same filename
	_, err = repo.Insert(img)
	if err == nil {
		t.Error("Expected error for duplicate filename, got nil")
	}
}

func TestImageRepository_GetByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := sqlite.NewImageRepository(db)

	timestamp := time.Now().Truncate(time.Second)
	img := &model.Image{
		Filename:  "getbyid_test.jpg",
		Camera:    "cam2",
		Timestamp: timestamp,
		FilePath:  "/images/getbyid_test.jpg",
		FileSize:  2048,
	}

	id, err := repo.Insert(img)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	retrieved, err := repo.GetByID(id)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected image, got nil")
	}

	if retrieved.Filename != img.Filename {
		t.Errorf("Filename mismatch: expected %s, got %s", img.Filename, retrieved.Filename)
	}

	if retrieved.Camera != img.Camera {
		t.Errorf("Camera mismatch: expected %s, got %s", img.Camera, retrieved.Camera)
	}

	if retrieved.FileSize != img.FileSize {
		t.Errorf("FileSize mismatch: expected %d, got %d", img.FileSize, retrieved.FileSize)
	}
}

func TestImageRepository_GetByID_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := sqlite.NewImageRepository(db)

	retrieved, err := repo.GetByID(99999)
	if err != nil {
		t.Fatalf("GetByID should not error for non-existent ID: %v", err)
	}

	if retrieved != nil {
		t.Error("Expected nil for non-existent image")
	}
}

func TestImageRepository_GetByFilename(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := sqlite.NewImageRepository(db)

	img := &model.Image{
		Filename:  "byfilename_test.jpg",
		Camera:    "cam3",
		Timestamp: time.Now(),
		FilePath:  "/images/byfilename_test.jpg",
		FileSize:  512,
	}

	_, err := repo.Insert(img)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	retrieved, err := repo.GetByFilename("byfilename_test.jpg")
	if err != nil {
		t.Fatalf("GetByFilename failed: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected image, got nil")
	}

	if retrieved.Camera != "cam3" {
		t.Errorf("Camera mismatch: expected cam3, got %s", retrieved.Camera)
	}
}

func TestImageRepository_GetAll_NoFilter(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := sqlite.NewImageRepository(db)

	// Insert multiple images
	for i := 0; i < 5; i++ {
		img := &model.Image{
			Filename:  "image_" + string(rune('a'+i)) + ".jpg",
			Camera:    "cam1",
			Timestamp: time.Now(),
			FilePath:  "/images/",
			FileSize:  int64(i * 100),
		}
		repo.Insert(img)
	}

	images, err := repo.GetAll(&dto.ImageFilters{})
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}

	if len(images) != 5 {
		t.Errorf("Expected 5 images, got %d", len(images))
	}
}

func TestImageRepository_GetAll_FilterByCamera(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := sqlite.NewImageRepository(db)

	// Insert images for different cameras
	cameras := []string{"cam1", "cam1", "cam2", "cam3", "cam1"}
	for i, cam := range cameras {
		img := &model.Image{
			Filename:  "filter_" + string(rune('a'+i)) + ".jpg",
			Camera:    cam,
			Timestamp: time.Now(),
			FilePath:  "/images/",
			FileSize:  100,
		}
		repo.Insert(img)
	}

	images, err := repo.GetAll(&dto.ImageFilters{Camera: "cam1"})
	if err != nil {
		t.Fatalf("GetAll with camera filter failed: %v", err)
	}

	if len(images) != 3 {
		t.Errorf("Expected 3 images for cam1, got %d", len(images))
	}
}

func TestImageRepository_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := sqlite.NewImageRepository(db)

	img := &model.Image{
		Filename:  "to_delete.jpg",
		Camera:    "cam1",
		Timestamp: time.Now(),
		FilePath:  "/images/to_delete.jpg",
		FileSize:  100,
	}

	id, err := repo.Insert(img)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	err = repo.Delete(id)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	retrieved, _ := repo.GetByID(id)
	if retrieved != nil {
		t.Error("Image should be deleted")
	}
}

func TestImageRepository_DeleteByFilename(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := sqlite.NewImageRepository(db)

	img := &model.Image{
		Filename:  "delete_by_name.jpg",
		Camera:    "cam1",
		Timestamp: time.Now(),
		FilePath:  "/images/delete_by_name.jpg",
		FileSize:  100,
	}

	repo.Insert(img)

	err := repo.DeleteByFilename("delete_by_name.jpg")
	if err != nil {
		t.Fatalf("DeleteByFilename failed: %v", err)
	}

	retrieved, _ := repo.GetByFilename("delete_by_name.jpg")
	if retrieved != nil {
		t.Error("Image should be deleted")
	}
}

func TestImageRepository_DeleteAll(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := sqlite.NewImageRepository(db)

	// Insert multiple images
	for i := 0; i < 3; i++ {
		img := &model.Image{
			Filename:  "deleteall_" + string(rune('a'+i)) + ".jpg",
			Camera:    "cam1",
			Timestamp: time.Now(),
			FilePath:  "/images/",
			FileSize:  100,
		}
		repo.Insert(img)
	}

	err := repo.DeleteAll()
	if err != nil {
		t.Fatalf("DeleteAll failed: %v", err)
	}

	count, _ := repo.GetTotalCount(&dto.ImageFilters{})
	if count != 0 {
		t.Errorf("Expected 0 images after DeleteAll, got %d", count)
	}
}

func TestImageRepository_GetTotalCount(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := sqlite.NewImageRepository(db)

	// Start with 0
	count, err := repo.GetTotalCount(&dto.ImageFilters{})
	if err != nil {
		t.Fatalf("GetTotalCount failed: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 images, got %d", count)
	}

	// Add images
	for i := 0; i < 7; i++ {
		img := &model.Image{
			Filename:  "count_" + string(rune('a'+i)) + ".jpg",
			Camera:    "cam1",
			Timestamp: time.Now(),
			FilePath:  "/images/",
			FileSize:  100,
		}
		repo.Insert(img)
	}

	count, err = repo.GetTotalCount(&dto.ImageFilters{})
	if err != nil {
		t.Fatalf("GetTotalCount failed: %v", err)
	}
	if count != 7 {
		t.Errorf("Expected 7 images, got %d", count)
	}
}

// ========================================
// Detection Repository Tests
// ========================================

func TestDetectionRepository_Insert(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	imageRepo := sqlite.NewImageRepository(db)
	detectionRepo := sqlite.NewDetectionRepository(db)

	// First insert an image
	img := &model.Image{
		Filename:  "detection_test.jpg",
		Camera:    "cam1",
		Timestamp: time.Now(),
		FilePath:  "/images/detection_test.jpg",
		FileSize:  1024,
	}
	imageID, _ := imageRepo.Insert(img)

	// Insert detection
	det := &model.Detection{
		ImageID:    imageID,
		ObjectName: "person",
		X:          100,
		Y:          50,
		Width:      200,
		Height:     400,
		Confidence: 0.95,
	}

	detID, err := detectionRepo.Insert(det)
	if err != nil {
		t.Fatalf("Detection insert failed: %v", err)
	}

	if detID <= 0 {
		t.Errorf("Expected positive detection ID, got %d", detID)
	}
}

func TestDetectionRepository_InsertBatch(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	imageRepo := sqlite.NewImageRepository(db)
	detectionRepo := sqlite.NewDetectionRepository(db)

	img := &model.Image{
		Filename:  "batch_detection.jpg",
		Camera:    "cam1",
		Timestamp: time.Now(),
		FilePath:  "/images/batch_detection.jpg",
		FileSize:  1024,
	}
	imageID, _ := imageRepo.Insert(img)

	detections := []model.Detection{
		{ImageID: imageID, ObjectName: "person", Confidence: 0.9},
		{ImageID: imageID, ObjectName: "car", Confidence: 0.85},
		{ImageID: imageID, ObjectName: "dog", Confidence: 0.75},
	}

	err := detectionRepo.InsertBatch(detections)
	if err != nil {
		t.Fatalf("InsertBatch failed: %v", err)
	}

	retrieved, _ := detectionRepo.GetByImageID(imageID)
	if len(retrieved) != 3 {
		t.Errorf("Expected 3 detections, got %d", len(retrieved))
	}
}

func TestDetectionRepository_GetObjectNamesByImageID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	imageRepo := sqlite.NewImageRepository(db)
	detectionRepo := sqlite.NewDetectionRepository(db)

	img := &model.Image{
		Filename:  "objects_test.jpg",
		Camera:    "cam1",
		Timestamp: time.Now(),
		FilePath:  "/images/objects_test.jpg",
		FileSize:  1024,
	}
	imageID, _ := imageRepo.Insert(img)

	detections := []model.Detection{
		{ImageID: imageID, ObjectName: "person", Confidence: 0.9},
		{ImageID: imageID, ObjectName: "car", Confidence: 0.85},
		{ImageID: imageID, ObjectName: "person", Confidence: 0.7}, // Duplicate should be distinct
	}
	detectionRepo.InsertBatch(detections)

	objects, err := detectionRepo.GetObjectNamesByImageID(imageID)
	if err != nil {
		t.Fatalf("GetObjectNamesByImageID failed: %v", err)
	}

	// Should return distinct objects
	if len(objects) != 2 {
		t.Errorf("Expected 2 distinct objects, got %d: %v", len(objects), objects)
	}
}

func TestDetectionRepository_DeleteByImageID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	imageRepo := sqlite.NewImageRepository(db)
	detectionRepo := sqlite.NewDetectionRepository(db)

	img := &model.Image{
		Filename:  "delete_detections.jpg",
		Camera:    "cam1",
		Timestamp: time.Now(),
		FilePath:  "/images/delete_detections.jpg",
		FileSize:  1024,
	}
	imageID, _ := imageRepo.Insert(img)

	detections := []model.Detection{
		{ImageID: imageID, ObjectName: "person", Confidence: 0.9},
		{ImageID: imageID, ObjectName: "car", Confidence: 0.85},
	}
	detectionRepo.InsertBatch(detections)

	err := detectionRepo.DeleteByImageID(imageID)
	if err != nil {
		t.Fatalf("DeleteByImageID failed: %v", err)
	}

	retrieved, _ := detectionRepo.GetByImageID(imageID)
	if len(retrieved) != 0 {
		t.Errorf("Expected 0 detections after delete, got %d", len(retrieved))
	}
}

// ========================================
// Gallery Handler Tests (Integration-like)
// ========================================

func TestGalleryHandler_DeletePicture_Success(t *testing.T) {
	db, cleanupDB := setupTestDB(t)
	defer cleanupDB()

	cfg, imageDir, cleanupCfg := setupTestConfig(t)
	defer cleanupCfg()

	// Create test file and database record
	testFilename := "to_delete_handler.jpg"
	createTestImageFile(t, imageDir, testFilename, nil)

	imageRepo := sqlite.NewImageRepository(db)
	img := &model.Image{
		Filename:  testFilename,
		Camera:    "cam1",
		Timestamp: time.Now(),
		FilePath:  filepath.Join(imageDir, testFilename),
		FileSize:  100,
	}
	imageRepo.Insert(img)

	// Create handler
	handler := createDeleteHandler(cfg, imageRepo)

	req := httptest.NewRequest(http.MethodDelete, "/api/pictures/delete?filename="+testFilename, nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Verify file deleted from disk
	if _, err := os.Stat(filepath.Join(imageDir, testFilename)); !os.IsNotExist(err) {
		t.Error("File should be deleted from disk")
	}

	// Verify deleted from database
	retrieved, _ := imageRepo.GetByFilename(testFilename)
	if retrieved != nil {
		t.Error("Image should be deleted from database")
	}
}

func TestGalleryHandler_DeletePicture_MissingFilename(t *testing.T) {
	cfg := &config.Config{ImageDirectory: "."}
	handler := createDeleteHandler(cfg, nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/pictures/delete", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestGalleryHandler_ClearPictures_Success(t *testing.T) {
	db, cleanupDB := setupTestDB(t)
	defer cleanupDB()

	cfg, imageDir, cleanupCfg := setupTestConfig(t)
	defer cleanupCfg()

	// Create test files
	for i := 0; i < 3; i++ {
		filename := "clear_" + string(rune('a'+i)) + ".jpg"
		createTestImageFile(t, imageDir, filename, nil)
	}

	imageRepo := sqlite.NewImageRepository(db)
	handler := createClearHandler(cfg, imageRepo)

	req := httptest.NewRequest(http.MethodPost, "/api/pictures/clear", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, rr.Code)
	}

	// Verify all files deleted
	files, _ := os.ReadDir(imageDir)
	if len(files) != 0 {
		t.Errorf("Expected 0 files, got %d", len(files))
	}
}

func TestGalleryHandler_ViewPicture_MissingParam(t *testing.T) {
	cfg := &config.Config{ImageDirectory: "."}
	handler := createViewHandler(cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/pictures/view", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestGalleryHandler_ViewPicture_Success(t *testing.T) {
	cfg, imageDir, cleanup := setupTestConfig(t)
	defer cleanup()

	testFilename := "view_test.jpg"
	testContent := []byte("JPEG image content here")
	createTestImageFile(t, imageDir, testFilename, testContent)

	handler := createViewHandler(cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/pictures/view?image="+testFilename, nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if rr.Body.String() != string(testContent) {
		t.Error("Response body should match file content")
	}
}

func TestGalleryHandler_GetPictures_Success(t *testing.T) {
	db, cleanupDB := setupTestDB(t)
	defer cleanupDB()

	cfg, _, cleanupCfg := setupTestConfig(t)
	defer cleanupCfg()

	imageRepo := sqlite.NewImageRepository(db)
	detectionRepo := sqlite.NewDetectionRepository(db)

	// Insert test images
	for i := 0; i < 3; i++ {
		img := &model.Image{
			Filename:  "gallery_" + string(rune('a'+i)) + ".jpg",
			Camera:    "cam1",
			Timestamp: time.Now(),
			FilePath:  "/images/",
			FileSize:  int64((i + 1) * 1000),
		}
		imageRepo.Insert(img)
	}

	handler := createGetPicturesHandler(cfg, imageRepo, detectionRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/pictures?page=1&limit=10", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	length := int(response["length"].(float64))
	if length != 3 {
		t.Errorf("Expected 3 images, got %d", length)
	}

	page := int(response["currentPage"].(float64))
	if page != 1 {
		t.Errorf("Expected page 1, got %d", page)
	}
}

func TestGalleryHandler_GetPictures_WithPagination(t *testing.T) {
	db, cleanupDB := setupTestDB(t)
	defer cleanupDB()

	cfg, _, cleanupCfg := setupTestConfig(t)
	defer cleanupCfg()

	imageRepo := sqlite.NewImageRepository(db)
	detectionRepo := sqlite.NewDetectionRepository(db)

	// Insert 25 images
	for i := 0; i < 25; i++ {
		img := &model.Image{
			Filename:  "page_" + string(rune('a'+i)) + ".jpg",
			Camera:    "cam1",
			Timestamp: time.Now(),
			FilePath:  "/images/",
			FileSize:  100,
		}
		imageRepo.Insert(img)
	}

	handler := createGetPicturesHandler(cfg, imageRepo, detectionRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/pictures?page=1&limit=10", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	var response map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&response)

	totalPages := int(response["totalPages"].(float64))
	if totalPages != 3 {
		t.Errorf("Expected 3 total pages, got %d", totalPages)
	}

	length := int(response["length"].(float64))
	if length != 25 {
		t.Errorf("Expected 25 total images, got %d", length)
	}
}

// ========================================
// Test Helper Handler Creators
// ========================================

func createDeleteHandler(cfg *config.Config, imageRepo *sqlite.ImageRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filename := r.URL.Query().Get("filename")
		if filename == "" {
			http.Error(w, "Filename required", http.StatusBadRequest)
			return
		}

		filePath := filepath.Join(cfg.ImageDirectory, filename)
		os.Remove(filePath)

		if imageRepo != nil {
			imageRepo.DeleteByFilename(filename)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "deleted", "filename": filename})
	}
}

func createClearHandler(cfg *config.Config, imageRepo *sqlite.ImageRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		files, err := os.ReadDir(cfg.ImageDirectory)
		if err != nil {
			http.Error(w, "Unable to read directory", http.StatusInternalServerError)
			return
		}

		for _, file := range files {
			if !file.IsDir() {
				os.Remove(filepath.Join(cfg.ImageDirectory, file.Name()))
			}
		}

		if imageRepo != nil {
			imageRepo.DeleteAll()
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func createViewHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		image := r.URL.Query().Get("image")
		if image == "" {
			http.Error(w, "Image parameter required", http.StatusBadRequest)
			return
		}

		filePath := filepath.Join(cfg.ImageDirectory, image)
		http.ServeFile(w, r, filePath)
	}
}

func createGetPicturesHandler(cfg *config.Config, imageRepo *sqlite.ImageRepository, detectionRepo *sqlite.DetectionRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		page := atoiDefault(q.Get("page"), 1)
		limit := atoiDefault(q.Get("limit"), 24)

		filter := &dto.ImageFilters{
			Camera: q.Get("camera"),
			Object: q.Get("object"),
		}

		images, _ := imageRepo.GetAll(filter)
		totalCount, _ := imageRepo.GetTotalCount(filter)

		var totalSize int64
		for _, img := range images {
			totalSize += img.FileSize
		}

		var pictures []dto.ImageInfo
		for _, img := range images {
			var objects []string
			if detectionRepo != nil {
				objects, _ = detectionRepo.GetObjectNamesByImageID(img.ID)
			}

			pictures = append(pictures, dto.ImageInfo{
				Name:      img.Filename,
				Date:      img.Timestamp,
				TimeOfDay: img.Timestamp,
				Camera:    img.Camera,
				Objects:   objects,
			})
		}

		data := dto.ImagesData{
			Images:      pictures,
			ImagesDir:   cfg.ImageDirectory,
			Size:        totalSize,
			MaxSize:     2,
			Length:      totalCount,
			TotalPages:  (totalCount + limit - 1) / limit,
			CurrentPage: page,
			Limit:       limit,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	var result int
	for _, c := range s {
		if c < '0' || c > '9' {
			return def
		}
		result = result*10 + int(c-'0')
	}
	if result <= 0 {
		return def
	}
	return result
}
