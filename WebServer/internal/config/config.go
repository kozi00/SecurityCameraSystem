package config

import (
	"os"
	"path/filepath"
	"strconv"
)

// Config holds application configuration loaded from environment variables
// or default values when env vars are not present.
type Config struct {
	Port              int
	Password          string
	ModelPath         string
	ConfigPath        string
	ImageDirectory    string
	ProcessingWorkers int
	LogDirectory      string
}

// Load reads configuration from environment variables and returns a Config instance.
func Load() *Config {
	return &Config{
		Port:              getEnvAsInt("PORT", 8080),
		Password:          getEnv("PASSWORD", "sienkiewicza2"),
		ModelPath:         getEnv("MODEL_PATH", filepath.Join(".", "internal", "services", "ai", "frozen_inference_graph.pb")),
		ConfigPath:        getEnv("CONFIG_PATH", filepath.Join(".", "internal", "services", "ai", "ssd_mobilenet_v1_coco_2017_11_17.pbtxt")),
		ImageDirectory:    getEnv("IMAGE_DIR", filepath.Join(".", "static", "images")),
		LogDirectory:      getEnv("LOG_DIR", filepath.Join(".", "logs")),
		ProcessingWorkers: getEnvAsInt("PROCESSING_WORKERS", 4), // 4 worker threads
	}
}

// getEnv returns the environment variable value or a default if empty.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt returns the integer value of an environment variable or a default value.
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
