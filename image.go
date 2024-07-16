package main

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"time"
)

func extractMetadata(file io.Reader) (time.Time, string, error) {
	_, _, err := image.DecodeConfig(file)
	if err != nil {
		return time.Time{}, "", err
	}

	// TODO: Extract date and location from metadata
	// This requires parsing EXIF data, which is not available in the standard library
	// For now, we'll use the current time and an empty location
	return time.Now(), "", nil
}
