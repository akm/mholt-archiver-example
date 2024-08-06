package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func zipFilesHandler(w http.ResponseWriter, r *http.Request) {
	filesParam := r.URL.Query().Get("files")
	if filesParam == "" {
		http.Error(w, "No files specified", http.StatusBadRequest)
		return
	}

	files := strings.Split(filesParam, ",")
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	baseDir := os.Getenv("BASE_DIR")
	if baseDir == "" {
		baseDir = "."
	}

	for _, filename := range files {
		filename := filepath.Join(baseDir, filename)

		file, err := os.Open(filename)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to open file: %s", filename), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		wr, err := zipWriter.Create(filename)
		if err != nil {
			http.Error(w, "Failed to create zip entry", http.StatusInternalServerError)
			return
		}

		if _, err := io.Copy(wr, file); err != nil {
			http.Error(w, "Failed to write file to zip", http.StatusInternalServerError)
			return
		}
	}

	if err := zipWriter.Close(); err != nil {
		http.Error(w, "Failed to close zip writer", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\"files.zip\"")
	w.Write(buf.Bytes())
}

func main() {
	http.HandleFunc("/zip", zipFilesHandler)
	fmt.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
