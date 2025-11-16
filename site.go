package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Photo struct {
	Filename string    `json:"filename"`
	Date     time.Time `json:"date"`
	Location string    `json:"location"`
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	// Add cache headers for static HTML
	w.Header().Set("Cache-Control", "public, max-age=3600")
	http.ServeFile(w, r, "index.html")
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form with max size
	err := r.ParseMultipartForm(config.MaxUploadSize)
	if err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	username := r.FormValue("username")
	listName := r.FormValue("listname")

	if username == "" || listName == "" {
		http.Error(w, "Username and list name are required", http.StatusBadRequest)
		return
	}

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		http.Error(w, "File must be an image", http.StatusBadRequest)
		return
	}

	err = saveImage(username, listName, file, header)
	if err != nil {
		logError("Error saving image: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Image uploaded successfully",
	})
}

func handleUserLists(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username required", http.StatusBadRequest)
		return
	}

	lists, err := getUserLists(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"username": username,
		"lists":    lists,
	})
}

func handleViewList(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/view/"), "/")
	if len(parts) != 2 {
		http.Error(w, "Invalid URL format. Use /api/view/{username}/{listname}", http.StatusBadRequest)
		return
	}

	username := parts[0]
	listName := parts[1]

	photos, err := getListPhotos(username, listName)
	if err != nil {
		http.Error(w, "List not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"username": username,
		"listName": listName,
		"photos":   photos,
	})
}

func handleServeImage(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/image/"), "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	username := parts[0]
	listName := parts[1]
	filename := strings.Join(parts[2:], "/")

	// Get size parameter for responsive images
	size := r.URL.Query().Get("size")
	if size == "" {
		size = "large"
	}

	basePath := filepath.Join(config.UploadDir, username, listName, filename)
	imagePath := getImageVariant(basePath, size)

	// Check if file exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	// Set cache headers for images (1 year)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")

	// Set content type based on file extension
	ext := strings.ToLower(filepath.Ext(imagePath))
	switch ext {
	case ".webp":
		w.Header().Set("Content-Type", "image/webp")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	}

	http.ServeFile(w, r, imagePath)
}

func saveImage(username, listName string, file multipart.File, header *multipart.FileHeader) error {
	// Get or create user
	userID, err := getOrCreateUser(username)
	if err != nil {
		return fmt.Errorf("failed to get/create user: %w", err)
	}

	// Get or create list
	listID, err := getOrCreateList(userID, listName)
	if err != nil {
		return fmt.Errorf("failed to get/create list: %w", err)
	}

	// Create directory
	dir := filepath.Join(config.UploadDir, username, listName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Generate unique filename
	randBytes := make([]byte, 8)
	rand.Read(randBytes)
	uniqueID := hex.EncodeToString(randBytes)
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%d_%s%s", time.Now().Unix(), uniqueID, ext)
	fullPath := filepath.Join(dir, filename)

	// Save original file
	dst, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Read file into buffer for multiple operations
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	// Extract EXIF metadata
	file.Seek(0, 0)
	date, location, _ := extractEXIFMetadata(file)

	// Write file
	_, err = dst.Write(fileBytes)
	if err != nil {
		return err
	}

	// Get image dimensions
	file.Seek(0, 0)
	cfg, _, err := image.DecodeConfig(file)
	width, height := 0, 0
	if err == nil {
		width = cfg.Width
		height = cfg.Height
	}

	// Save to database
	photo := Photo{
		Filename: filename,
		Date:     date,
		Location: location,
	}

	if err := savePhotoToDB(listID, photo, width, height, len(fileBytes)); err != nil {
		return err
	}

	// Queue image processing job
	imageJobQueue <- ImageJob{
		FilePath: fullPath,
		Username: username,
		ListName: listName,
		OrigName: header.Filename,
	}

	return nil
}

