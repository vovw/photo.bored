package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/upload", handleUpload)
	http.HandleFunc("/view/", handleViewList)
	http.HandleFunc("/image/", handleServeImage)

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

