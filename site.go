package main

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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
func handleDeletePhoto(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodDelete {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    err := r.ParseForm()
    if err != nil {
        http.Error(w, "Error parsing form", http.StatusBadRequest)
        return
    }
    username := r.Form.Get("username")
    listName := r.Form.Get("listname")
    filename := r.Form.Get("filename")

    if username == "" || listName == "" || filename == "" {
        http.Error(w, "Missing required parameters", http.StatusBadRequest)
        return
    }
    userLists, ok := userLists[username]
    if !ok {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }
    list, ok := userLists[listName]
    if !ok {
        http.Error(w, "List not found", http.StatusNotFound)
        return
    }
    for i, photo := range list.Photos {
        if photo.Filename == filename {
            
            list.Photos = append(list.Photos[:i], list.Photos[i+1:]...)
            
            // Delete the file from the filesystem
            err := os.Remove(filepath.Join("uploads", username, listName, filename))
            if err != nil {
                log.Printf("Error deleting file: %v", err)
                http.Error(w, "Error deleting file", http.StatusInternalServerError)
                return
            }

            fmt.Fprintf(w, "Photo deleted successfully")
            return
        }
    }
    http.Error(w, "Photo not found", http.StatusNotFound)
}
func handleMovePhoto(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    err := r.ParseForm()
    if err != nil {
        http.Error(w, "Error parsing form", http.StatusBadRequest)
        return
    }

    username := r.Form.Get("username")
    sourceList := r.Form.Get("source_list")
    destList := r.Form.Get("dest_list")
    photoCount, err := strconv.Atoi(r.Form.Get("photo_count"))
    if err != nil || photoCount <= 0 {
        http.Error(w, "Invalid photo count", http.StatusBadRequest)
        return
    }
    userLists, ok := userLists[username]
    if !ok {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }
    sourcePhotos, ok := userLists[sourceList]
    if !ok {
        http.Error(w, "Source list not found", http.StatusNotFound)
        return
    }

    destPhotos, ok := userLists[destList]
    if !ok {
        http.Error(w, "Destination list not found", http.StatusNotFound)
        return
    }

    movedPhotos := 0
    for i := 0; i < photoCount; i++ {
        filename := r.Form.Get(fmt.Sprintf("photo_%d", i))
        if filename == "" {
            continue
        }
        for j, photo := range sourcePhotos.Photos {
            if photo.Filename == filename {
                destPhotos.Photos = append(destPhotos.Photos, photo)
                sourcePhotos.Photos = append(sourcePhotos.Photos[:j], sourcePhotos.Photos[j+1:]...)
                
                // Move the actual file
                oldPath := filepath.Join("uploads", username, sourceList, filename)
                newPath := filepath.Join("uploads", username, destList, filename)
                err := os.Rename(oldPath, newPath)
                if err != nil {
                    log.Printf("Error moving file %s: %v", filename, err)
                    continue
                }
                
                movedPhotos++
                break
            }
        }
    }

    fmt.Fprintf(w, "Moved %d photos successfully", movedPhotos)
}