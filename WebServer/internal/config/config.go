package config

import (
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {
	Port                     int
	Password                 string
	ModelPath                string
	ConfigPath               string
	ImageDirectory           string
	ImageBufferLimit         int
	ImageBufferFlushInterval int
	MotionThreshold          int
	ProcessingInterval       int   // Co którą klatkę przetwarzać (1=każdą, 3=co trzecią)
	ProcessingWorkers        int   // Liczba worker threads do przetwarzania
	MaxImageDirectorySize    int64 // Maksymalny rozmiar katalogu z obrazami w GB
	LogDirectory             string
}

func Load() *Config {
	return &Config{
		Port:                     getEnvAsInt("PORT", 8080),
		Password:                 getEnv("PASSWORD", "sienkiewicza2"),
		ModelPath:                getEnv("MODEL_PATH", filepath.Join(".", "internal", "services", "ai", "frozen_inference_graph.pb")),
		ConfigPath:               getEnv("CONFIG_PATH", filepath.Join(".", "internal", "services", "ai", "ssd_mobilenet_v1_coco_2017_11_17.pbtxt")),
		ImageDirectory:           getEnv("IMAGE_DIR", filepath.Join(".", "images")),
		ImageBufferLimit:         getEnvAsInt("BUFFER_LIMIT", 7),
		ImageBufferFlushInterval: getEnvAsInt("FLUSH_INTERVAL", 30),
		MotionThreshold:          getEnvAsInt("MOTION_THRESHOLD", 10000),       // Default threshold for motion detection
		ProcessingInterval:       getEnvAsInt("PROCESSING_INTERVAL", 3),        // Przetwarzaj co 3. klatkę
		ProcessingWorkers:        getEnvAsInt("PROCESSING_WORKERS", 3),         // 3 worker threads
		MaxImageDirectorySize:    getEnvAsInt64("MAX_IMAGE_DIRECTORY_SIZE", 4), // Maksymalny rozmiar katalogu z obrazami w GB
		LogDirectory:             getEnv("LOG_DIR", filepath.Join(".", "logs")),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}
