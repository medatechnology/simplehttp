package simplehttp

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// FileInfo represents uploaded file metadata
type FileInfo struct {
	Filename     string
	Size         int64
	ContentType  string
	LastModified time.Time
	Hash         string // MD5/SHA hash of file
}

// File handling utilities
type FileHandler struct {
	UploadDir    string
	MaxFileSize  int64
	AllowedTypes []string
}

func NewFileHandler(uploadDir string) *FileHandler {
	return &FileHandler{
		UploadDir:    uploadDir,
		MaxFileSize:  10 << 20, // 10MB default
		AllowedTypes: []string{"image/*", "application/pdf"},
	}
}

// This is independent of implementation
// Make sure the implementation context has .GetFile and .SaveFile
func (h *FileHandler) HandleUpload() MedaHandlerFunc {
	return func(c MedaContext) error {
		file, err := c.GetFile("file")
		if err != nil {
			return c.JSON(400, map[string]string{"error": "file required"})
		}

		// Validate file size
		if file.Size > h.MaxFileSize {
			return c.JSON(400, map[string]string{"error": "file too large"})
		}

		// Generate safe filename
		filename := generateSafeFilename(file.Filename)

		// Save file
		if err := c.SaveFile(file, filepath.Join(h.UploadDir, filename)); err != nil {
			return c.JSON(500, map[string]string{"error": "failed to save file"})
		}

		return c.JSON(200, FileInfo{
			Filename:    filename,
			Size:        file.Size,
			ContentType: file.Header.Get("Content-Type"),
		})
	}
}

func (h *FileHandler) HandleDownload(filepath string) MedaHandlerFunc {
	return func(c MedaContext) error {
		return c.SendFile(filepath, true)
	}
}

func generateSafeFilename(filename string) string {
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)

	// Replace unsafe characters
	safeName := strings.Map(func(r rune) rune {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, name)

	// Add timestamp to ensure uniqueness
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s_%d%s", safeName, timestamp, ext)
}
