package config

import (
	"os"
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
}

func Load() *Config {
	return &Config{
		Port:                     getEnvAsInt("PORT", 8080),
		Password:                 getEnv("PASSWORD", "sienkiewicza2"),
		ModelPath:                getEnv("MODEL_PATH", "D:\\2025Scripts\\SecurityCameraSystem\\WebServer\\internal\\services\\AI\\frozen_inference_graph.pb"),
		ConfigPath:               getEnv("CONFIG_PATH", "D:\\2025Scripts\\SecurityCameraSystem\\WebServer\\internal\\services\\AI\\ssd_mobilenet_v1_coco_2017_11_17.pbtxt"),
		ImageDirectory:           getEnv("IMAGE_DIR", "D:\\2025Scripts\\SecurityCameraSystem\\WebServer\\static\\images"),
		ImageBufferLimit:         getEnvAsInt("BUFFER_LIMIT", 20),
		ImageBufferFlushInterval: getEnvAsInt("FLUSH_INTERVAL", 5),
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
