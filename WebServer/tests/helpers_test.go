package tests

import (
	"testing"
	"time"

	"webserver/internal/dto"
)

// ========================================
// Helper Function Tests
// ========================================

func TestAtoiDefault_ValidInput(t *testing.T) {
	tests := []struct {
		input    string
		def      int
		expected int
	}{
		{"10", 5, 10},
		{"1", 0, 1},
		{"100", 1, 100},
		{"999", 0, 999},
	}

	for _, tt := range tests {
		result := atoiDefault(tt.input, tt.def)
		if result != tt.expected {
			t.Errorf("atoiDefault(%q, %d) = %d, expected %d", tt.input, tt.def, result, tt.expected)
		}
	}
}

func TestAtoiDefault_InvalidInput(t *testing.T) {
	tests := []struct {
		input    string
		def      int
		expected int
	}{
		{"", 5, 5},
		{"abc", 10, 10},
		{"-1", 5, 5},
		{"0", 5, 5},
		{"12.5", 5, 5},
		{"12abc", 5, 5},
	}

	for _, tt := range tests {
		result := atoiDefault(tt.input, tt.def)
		if result != tt.expected {
			t.Errorf("atoiDefault(%q, %d) = %d, expected %d", tt.input, tt.def, result, tt.expected)
		}
	}
}

// ========================================
// DTO Tests
// ========================================

func TestImageFilters_Empty(t *testing.T) {
	filter := &dto.ImageFilters{}

	if filter.Camera != "" {
		t.Errorf("Expected empty camera, got %s", filter.Camera)
	}

	if filter.Object != "" {
		t.Errorf("Expected empty object, got %s", filter.Object)
	}

	if !filter.DateAfter.IsZero() {
		t.Error("Expected zero DateAfter")
	}
}

func TestImageFilters_WithValues(t *testing.T) {
	now := time.Now()
	filter := &dto.ImageFilters{
		Camera:     "cam1",
		Object:     "person",
		DateAfter:  now.AddDate(0, 0, -7),
		DateBefore: now,
		TimeAfter:  time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC),
		TimeBefore: time.Date(0, 1, 1, 18, 0, 0, 0, time.UTC),
	}

	if filter.Camera != "cam1" {
		t.Errorf("Expected camera 'cam1', got %s", filter.Camera)
	}

	if filter.Object != "person" {
		t.Errorf("Expected object 'person', got %s", filter.Object)
	}

	if filter.DateAfter.IsZero() {
		t.Error("DateAfter should not be zero")
	}
}

func TestImageInfo_MarshalJSON(t *testing.T) {
	info := dto.ImageInfo{
		Name:      "test.jpg",
		Date:      time.Date(2025, 6, 15, 14, 30, 0, 0, time.UTC),
		TimeOfDay: time.Date(2025, 6, 15, 14, 30, 0, 0, time.UTC),
		Camera:    "cam1",
		Objects:   []string{"person", "car"},
	}

	data, err := info.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	jsonStr := string(data)

	// Check date format (DD-MM-YYYY)
	if !contains(jsonStr, "15-06-2025") {
		t.Errorf("Expected date format DD-MM-YYYY, got: %s", jsonStr)
	}

	// Check time format (HH:MM)
	if !contains(jsonStr, "14:30") {
		t.Errorf("Expected time format HH:MM, got: %s", jsonStr)
	}
}

func TestImagesData_Pagination(t *testing.T) {
	tests := []struct {
		length   int
		limit    int
		expected int
		name     string
	}{
		{100, 10, 10, "even division"},
		{25, 10, 3, "with remainder"},
		{10, 10, 1, "exact match"},
		{0, 10, 0, "empty"},
		{5, 10, 1, "less than limit"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			totalPages := (tt.length + tt.limit - 1) / tt.limit
			if tt.length == 0 {
				totalPages = 0
			}

			if totalPages != tt.expected {
				t.Errorf("Expected %d pages, got %d", tt.expected, totalPages)
			}
		})
	}
}

// ========================================
// Data Validation Tests
// ========================================

func TestFilename_Validation(t *testing.T) {
	validFilenames := []string{
		"image.jpg",
		"photo_001.png",
		"cam1_2025-01-04_14-30-00.jpg",
		"test-file.jpeg",
	}

	for _, filename := range validFilenames {
		if !isValidFilename(filename) {
			t.Errorf("Expected %s to be valid", filename)
		}
	}
}

func TestFilename_Invalid(t *testing.T) {
	invalidFilenames := []string{
		"",
		"../secret.jpg",
		"/etc/passwd",
		"file\x00name.jpg",
	}

	for _, filename := range invalidFilenames {
		if isValidFilename(filename) {
			t.Errorf("Expected %s to be invalid", filename)
		}
	}
}

// ========================================
// Helper functions
// ========================================

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func isValidFilename(filename string) bool {
	if filename == "" {
		return false
	}

	// Check for null bytes
	for _, c := range filename {
		if c == 0 {
			return false
		}
	}

	// Check for path traversal
	if len(filename) >= 2 && filename[0:2] == ".." {
		return false
	}

	// Check for absolute paths
	if len(filename) > 0 && filename[0] == '/' {
		return false
	}

	return true
}
