package main

import (
	"compress/flate"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver/v3"
)

func zipFilesHandler(w http.ResponseWriter, r *http.Request) {
	filesParam := r.URL.Query().Get("files")
	if filesParam == "" {
		http.Error(w, "No files specified", http.StatusBadRequest)
		return
	}

	files := strings.Split(filesParam, ",")

	baseDir := os.Getenv("BASE_DIR")
	if baseDir == "" {
		baseDir = "."
	}

	z := archiver.Zip{
		CompressionLevel:       flate.DefaultCompression,
		MkdirAll:               true,
		SelectiveCompression:   true,
		ContinueOnError:        false,
		OverwriteExisting:      false,
		ImplicitTopLevelFolder: false,
	}

	if err := z.Create(w); err != nil {
		http.Error(w, "Failed to create zip archive", http.StatusInternalServerError)
		return
	}
	defer z.Close()

	for _, origFilename := range files {
		filename := filepath.Join(baseDir, origFilename)

		info, err := os.Stat(filename)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to stat file: %s", filename), http.StatusInternalServerError)
			return
		}

		// get file's name for the inside of the archive
		internalName, err := archiver.NameInArchive(info, filename, origFilename)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get internal name for file: %s", filename), http.StatusInternalServerError)
			return
		}

		file, err := os.Open(filename)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to open file: %s", filename), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		// write it to the archive
		if err := z.Write(archiver.File{
			FileInfo: archiver.FileInfo{
				FileInfo:   info,
				CustomName: internalName,
			},
			ReadCloser: file,
		}); err != nil {
			http.Error(w, fmt.Sprintf("Failed to write file to zip: %s", filename), http.StatusInternalServerError)
			return
		}
	}

	// w.Header().Set("Content-Type", "application/zip")
	// w.Header().Set("Content-Disposition", "attachment; filename=\"files.zip\"")
	// w.Write(buf.Bytes())
}

func main() {
	http.HandleFunc("/zip", zipFilesHandler)
	fmt.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
