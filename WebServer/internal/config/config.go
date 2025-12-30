package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config holds application configuration loaded from environment variables
// or default values when env vars are not present.

type CameraID string

type Config struct {
	Port              int
	Password          string
	ModelPath         string
	ConfigPath        string
	ImageDirectory    string
	ProcessingWorkers int
	LogDirectory      string
	DatabasePath      string
	CamerasPort       int
	CameraNames       map[string]string
}

// Load reads configuration from environment variables and returns a Config instance.
func Load() *Config {
	return &Config{
		Port:              getEnvAsInt("PORT", 80),
		Password:          getEnv("PASSWORD", ""),
		ModelPath:         getEnv("MODEL_PATH", filepath.Join(".", "internal", "services", "ai", "frozen_inference_graph.pb")),
		ConfigPath:        getEnv("CONFIG_PATH", filepath.Join(".", "internal", "services", "ai", "ssd_mobilenet_v1_coco_2017_11_17.pbtxt")),
		ImageDirectory:    getEnv("IMAGE_DIR", filepath.Join(".", "static", "images")),
		LogDirectory:      getEnv("LOG_DIR", filepath.Join(".", "logs")),
		DatabasePath:      getEnv("DATABASE_PATH", filepath.Join(".", "data", "images.db")),
		ProcessingWorkers: getEnvAsInt("PROCESSING_WORKERS", 4), // 4 worker threads of ai processing
		CamerasPort:       getEnvAsInt("CAMERAS_PORT", 81),
		CameraNames:       parseCameraEnv(getEnv("CAMERAS", "")),
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

func parseCameraEnv(envValue string) map[string]string {
	cameras := make(map[string]string)

	if envValue == "" {
		return map[string]string{
			"192.168.1.32": "drzwi",
			"192.168.1.29": "brama",
		}
	}

	pairs := strings.Split(envValue, ",")
	for _, pair := range pairs {
		parts := strings.Split(strings.TrimSpace(pair), ":")
		if len(parts) == 2 {
			ip := strings.TrimSpace(parts[0])
			name := strings.TrimSpace(parts[1])
			cameras[ip] = name
		}
	}
	return cameras
}
