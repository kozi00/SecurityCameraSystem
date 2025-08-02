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
	ProcessingInterval       int // Co którą klatkę przetwarzać (1=każdą, 3=co trzecią)
	ProcessingWorkers        int // Liczba worker threads do przetwarzania
}

func Load() *Config {
	return &Config{
		Port:                     getEnvAsInt("PORT", 8080),
		Password:                 getEnv("PASSWORD", "sienkiewicza2"),
		ModelPath:                getEnv("MODEL_PATH", filepath.Join(".", "internal", "services", "ai", "mobilenet_iter_73000.caffemodel")),
		ConfigPath:               getEnv("CONFIG_PATH", filepath.Join(".", "internal", "services", "ai", "deploy.prototxt")),
		ImageDirectory:           getEnv("IMAGE_DIR", filepath.Join(".", "static", "images")),
		ImageBufferLimit:         getEnvAsInt("BUFFER_LIMIT", 7),
		ImageBufferFlushInterval: getEnvAsInt("FLUSH_INTERVAL", 30),
		MotionThreshold:          getEnvAsInt("MOTION_THRESHOLD", 10000), // Default threshold for motion detection
		ProcessingInterval:       getEnvAsInt("PROCESSING_INTERVAL", 3),  // Przetwarzaj co 3. klatkę
		ProcessingWorkers:        getEnvAsInt("PROCESSING_WORKERS", 5),   // 5 worker threads
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
