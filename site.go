package main

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type PhotoList struct {
	Name   string
	Photos []Photo
}

type Photo struct {
	Filename string
	Date     time.Time
	Location string
}

var userLists = make(map[string]map[string]*PhotoList)

func handleRoot(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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

	err = saveImage(username, listName, file, header)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Image uploaded successfully")
}

func handleViewList(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	username := parts[2]
	listName := parts[3]

	list, ok := userLists[username][listName]
	if !ok {
		http.Error(w, "List not found", http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "Photos in %s's %s list:\n", username, listName)
	for _, photo := range list.Photos {
		fmt.Fprintf(w, "- %s (Date: %s, Location: %s)\n", photo.Filename, photo.Date, photo.Location)
	}
}

func handleServeImage(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 5 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	username := parts[2]
	listName := parts[3]
	filename := parts[4]

	path := filepath.Join("uploads", username, listName, filename)
	http.ServeFile(w, r, path)
}

func saveImage(username, listName string, file multipart.File, header *multipart.FileHeader) error {
	dir := filepath.Join("uploads", username, listName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	filename := filepath.Join(dir, header.Filename)
	dst, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err = file.Seek(0, 0); err != nil {
		return err
	}

	date, location, err := extractMetadata(file)
	if err != nil {
		return err
	}

	if _, err = file.Seek(0, 0); err != nil {
		return err
	}

	if _, err = io.Copy(dst, file); err != nil {
		return err
	}

	photo := Photo{
		Filename: header.Filename,
		Date:     date,
		Location: location,
	}

	addPhotoToList(username, listName, photo)
	return nil
}

func addPhotoToList(username, listName string, photo Photo) {
	if _, ok := userLists[username]; !ok {
		userLists[username] = make(map[string]*PhotoList)
	}
	if _, ok := userLists[username][listName]; !ok {
		userLists[username][listName] = &PhotoList{Name: listName}
	}
	userLists[username][listName].Photos = append(userLists[username][listName].Photos, photo)
}

