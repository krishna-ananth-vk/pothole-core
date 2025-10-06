package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

type UploadResponse struct {
	FileName  string    `json:"file_name"`
	FilePath  string    `json:"file_path"`
	URL       string    `json:"url"`
	Timestamp time.Time `json:"timestamp"`
}

func UploadImage(w http.ResponseWriter, r *http.Request) {
	const uploadDir = "/var/www/public"

	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	file, header, err := r.FormFile("file")
	if err != nil {
		zap.L().Error("failed to read form file", zap.Error(err))
		http.Error(w, "Invalid file upload", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			zap.L().Error("failed to create upload directory", zap.Error(err))
			http.Error(w, "Failed to create directory", http.StatusInternalServerError)
			return
		}
	}

	ext := filepath.Ext(header.Filename)
	timestamp := time.Now().UnixNano()
	newFileName := fmt.Sprintf("%d%s", timestamp, ext)
	destPath := filepath.Join(uploadDir, newFileName)

	destFile, err := os.Create(destPath)
	if err != nil {
		zap.L().Error("failed to create destination file", zap.Error(err))
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, file); err != nil {
		zap.L().Error("failed to write file", zap.Error(err))
		http.Error(w, "Failed to write file", http.StatusInternalServerError)
		return
	}

	publicURL := fmt.Sprintf("/public/%s", newFileName)

	resp := UploadResponse{
		FileName:  header.Filename,
		FilePath:  destPath,
		URL:       publicURL,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
