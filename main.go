package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Load configuration
	loadConfig()
	logInfo("Starting photo.bored server")
	logInfo("Port: %s", config.Port)
	logInfo("Upload directory: %s", config.UploadDir)
	logInfo("Database: %s", config.DBPath)

	// Initialize database
	if err := initDB(config.DBPath); err != nil {
		logError("Failed to initialize database: %v", err)
		os.Exit(1)
	}
	defer closeDB()
	logInfo("Database initialized")

	// Initialize image processing workers
	initImageWorkers(config.ImageWorkers)
	logInfo("Started %d image processing workers", config.ImageWorkers)
	defer stopImageWorkers()

	// Setup routes with middlewares
	http.HandleFunc("/", chain(handleRoot, securityHeadersMiddleware, loggingMiddleware))
	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/api/upload", chain(handleUpload, securityHeadersMiddleware, loggingMiddleware))
	http.HandleFunc("/api/lists", chain(handleUserLists, gzipMiddleware, securityHeadersMiddleware, loggingMiddleware))
	http.HandleFunc("/api/view/", chain(handleViewList, gzipMiddleware, securityHeadersMiddleware, loggingMiddleware))
	http.HandleFunc("/image/", chain(handleServeImage, securityHeadersMiddleware, loggingMiddleware))

	// Create server with timeouts
	server := &http.Server{
		Addr:         ":" + config.Port,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		logInfo("Server listening on :%s", config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logError("Server error: %v", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logInfo("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logError("Server forced to shutdown: %v", err)
	}

	logInfo("Server stopped")
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

