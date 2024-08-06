package main

import (
	"compress/flate"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mholt/archiver/v3"
	"github.com/shirou/gopsutil/process"
)

func zipFilesHandler(w http.ResponseWriter, r *http.Request) {
	filesParam := r.URL.Query().Get("files")
	if filesParam == "" {
		http.Error(w, "No files specified", http.StatusBadRequest)
		return
	}

	log.Printf("START Zipping files: %s\n", filesParam)
	defer log.Printf("  END Zipping files: %s\n", filesParam)

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

func printUsage() {
	pid := os.Getpid()
	proc, err := process.NewProcess(int32(pid))
	if err != nil {
		fmt.Printf("Failed to get process: %v\n", err)
		return
	}

	for {
		cpuPercent, err := proc.CPUPercent()
		if err != nil {
			fmt.Printf("Failed to get CPU usage: %v\n", err)
		}

		memInfo, err := proc.MemoryInfo()
		if err != nil {
			fmt.Printf("Failed to get memory usage: %v\n", err)
		}

		log.Printf("CPU Usage: %.2f%%, Memory Usage: %v bytes\n", cpuPercent, memInfo.RSS)
		time.Sleep(100 * time.Millisecond)
	}
}

func main() {
	go printUsage()

	http.HandleFunc("/zip", zipFilesHandler)
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
