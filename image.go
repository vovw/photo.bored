package main

import (
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

func extractMetadata(file io.Reader) (time.Time, string, error) {
    // Decode the image to verify it's valid
    _, _, err := image.DecodeConfig(file)
    if err != nil {
        log.Printf("Error decoding image: %v", err)
        return time.Time{}, "", err
    }

    // Reset the reader to the beginning of the file
    seeker, ok := file.(io.ReadSeeker)
    if !ok {
        log.Print("File is not seekable")
        return time.Time{}, "", errors.New("file is not seekable")
    }
    _, err = seeker.Seek(0, io.SeekStart)
    if err != nil {
        log.Printf("Error seeking file: %v", err)
        return time.Time{}, "", err
    }

    // Extract EXIF data
    x, err := exif.Decode(file)
    if err != nil {
        log.Printf("Error decoding EXIF data: %v", err)
        return time.Time{}, "", err
    }

    // Get the date
    dateTime, err := x.DateTime()
    if err != nil {
        log.Printf("Error getting DateTime from EXIF: %v", err)
        dateTime = time.Now() // Fallback to current time if not available
    }

    // Get the GPS info
    lat, long, err := x.LatLong()
    location := ""
    if err == nil {
		location = fmt.Sprintf("%.6f,%.6f", lat, long)
    } else {
        log.Printf("Error getting GPS data from EXIF: %v", err)
    }

    return dateTime, location, nil
}