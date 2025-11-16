package main

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	Port           string
	DBPath         string
	UploadDir      string
	MaxUploadSize  int64
	ImageWorkers   int
	Debug          bool
	AllowedOrigins string
}

var config Config

func loadConfig() {
	config = Config{
		Port:           getEnv("PORT", "8080"),
		DBPath:         getEnv("DB_PATH", "./data/photos.db"),
		UploadDir:      getEnv("UPLOAD_DIR", "./uploads"),
		MaxUploadSize:  getEnvInt64("MAX_UPLOAD_SIZE", 50<<20), // 50MB default
		ImageWorkers:   getEnvInt("IMAGE_WORKERS", 4),
		Debug:          getEnvBool("DEBUG", false),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "*"),
	}

	// Create directories if they don't exist
	os.MkdirAll(config.UploadDir, 0755)
	os.MkdirAll("./data", 0755)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func logInfo(format string, v ...interface{}) {
	log.Printf("[INFO] "+format, v...)
}

func logError(format string, v ...interface{}) {
	log.Printf("[ERROR] "+format, v...)
}
