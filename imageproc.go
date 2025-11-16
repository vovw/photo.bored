package main

import (
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
	"github.com/rwcarlsen/goexif/exif"
)

// Image processing worker pool
var (
	imageJobQueue = make(chan ImageJob, 100)
	workerPool    *sync.WaitGroup
)

type ImageJob struct {
	FilePath string
	Username string
	ListName string
	OrigName string
}

// Thumbnail sizes for responsive images
const (
	ThumbnailSize = 300
	MediumSize    = 800
	LargeSize     = 1600
)

func initImageWorkers(numWorkers int) {
	workerPool = &sync.WaitGroup{}
	for i := 0; i < numWorkers; i++ {
		workerPool.Add(1)
		go imageWorker()
	}
}

func imageWorker() {
	defer workerPool.Done()
	for job := range imageJobQueue {
		processImageVariants(job)
	}
}

func stopImageWorkers() {
	close(imageJobQueue)
	workerPool.Wait()
}

// Extract EXIF metadata from image
func extractEXIFMetadata(file io.Reader) (date time.Time, location string, err error) {
	x, err := exif.Decode(file)
	if err != nil {
		// No EXIF data, return defaults
		return time.Now(), "", nil
	}

	// Extract date
	dateStr, err := x.Get(exif.DateTime)
	if err == nil {
		dateVal, _ := dateStr.StringVal()
		date, err = time.Parse("2006:01:02 15:04:05", dateVal)
		if err != nil {
			date = time.Now()
		}
	} else {
		date = time.Now()
	}

	// Extract GPS location
	lat, lon, err := x.LatLong()
	if err == nil {
		location = fmt.Sprintf("%.6f,%.6f", lat, lon)
	}

	return date, location, nil
}

// Process image and create optimized variants
func processImageVariants(job ImageJob) error {
	// Open original image
	src, err := imaging.Open(job.FilePath)
	if err != nil {
		return err
	}

	// Get original dimensions
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	baseDir := filepath.Dir(job.FilePath)
	baseName := strings.TrimSuffix(filepath.Base(job.FilePath), filepath.Ext(job.FilePath))

	// Generate thumbnail
	if width > ThumbnailSize || height > ThumbnailSize {
		thumb := imaging.Fit(src, ThumbnailSize, ThumbnailSize, imaging.Lanczos)
		thumbPath := filepath.Join(baseDir, baseName+"_thumb.jpg")
		saveOptimizedJPEG(thumb, thumbPath, 80)

		// Also save WebP version
		thumbWebP := filepath.Join(baseDir, baseName+"_thumb.webp")
		saveWebP(thumb, thumbWebP, 80)
	}

	// Generate medium size
	if width > MediumSize || height > MediumSize {
		medium := imaging.Fit(src, MediumSize, MediumSize, imaging.Lanczos)
		mediumPath := filepath.Join(baseDir, baseName+"_medium.jpg")
		saveOptimizedJPEG(medium, mediumPath, 85)

		mediumWebP := filepath.Join(baseDir, baseName+"_medium.webp")
		saveWebP(medium, mediumWebP, 85)
	}

	// Generate large size and optimize original
	if width > LargeSize || height > LargeSize {
		large := imaging.Fit(src, LargeSize, LargeSize, imaging.Lanczos)
		largePath := filepath.Join(baseDir, baseName+"_large.jpg")
		saveOptimizedJPEG(large, largePath, 90)

		largeWebP := filepath.Join(baseDir, baseName+"_large.webp")
		saveWebP(large, largeWebP, 90)
	} else {
		// Just optimize the original
		optimizedPath := filepath.Join(baseDir, baseName+"_large.jpg")
		saveOptimizedJPEG(src, optimizedPath, 90)

		optimizedWebP := filepath.Join(baseDir, baseName+"_large.webp")
		saveWebP(src, optimizedWebP, 90)
	}

	return nil
}

func saveOptimizedJPEG(img image.Image, path string, quality int) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	return jpeg.Encode(out, img, &jpeg.Options{Quality: quality})
}

func saveWebP(img image.Image, path string, quality float32) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	return webp.Encode(out, img, &webp.Options{Quality: quality})
}

// Get the appropriate image variant based on request
func getImageVariant(basePath, size string) string {
	dir := filepath.Dir(basePath)
	ext := filepath.Ext(basePath)
	name := strings.TrimSuffix(filepath.Base(basePath), ext)

	var variant string
	switch size {
	case "thumb":
		variant = name + "_thumb"
	case "medium":
		variant = name + "_medium"
	case "large":
		variant = name + "_large"
	default:
		return basePath
	}

	// Check for WebP version first (better compression)
	webpPath := filepath.Join(dir, variant+".webp")
	if _, err := os.Stat(webpPath); err == nil {
		return webpPath
	}

	// Fall back to JPEG
	jpegPath := filepath.Join(dir, variant+".jpg")
	if _, err := os.Stat(jpegPath); err == nil {
		return jpegPath
	}

	// Fall back to original
	return basePath
}

// Auto-orient image based on EXIF orientation
func autoOrient(img image.Image, orientation int) image.Image {
	switch orientation {
	case 2:
		return imaging.FlipH(img)
	case 3:
		return imaging.Rotate180(img)
	case 4:
		return imaging.FlipV(img)
	case 5:
		return imaging.Transpose(img)
	case 6:
		return imaging.Rotate270(img)
	case 7:
		return imaging.Transverse(img)
	case 8:
		return imaging.Rotate90(img)
	}
	return img
}
