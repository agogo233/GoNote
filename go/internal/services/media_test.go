package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMediaService(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewMediaService(tmpDir)

	assert.NotNil(t, service)
	assert.Equal(t, tmpDir, service.notesDir)
}

func TestGetMediaType(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{"PNG", "image.png", "image"},
		{"JPG", "photo.jpg", "image"},
		{"JPEG", "photo.jpeg", "image"},
		{"GIF", "animation.gif", "image"},
		{"WEBP", "image.webp", "image"},
		{"MP3", "audio.mp3", "audio"},
		{"WAV", "audio.wav", "audio"},
		{"OGG", "audio.ogg", "audio"},
		{"MP4", "video.mp4", "video"},
		{"WEBM", "video.webm", "video"},
		{"MOV", "video.mov", "video"},
		{"AVI", "video.avi", "video"},
		{"PDF", "document.pdf", "document"},
		{"Unknown", "file.xyz", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMediaType(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsMediaFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{"PNG", "image.png", true},
		{"JPG", "photo.jpg", true},
		{"MP3", "audio.mp3", true},
		{"MP4", "video.mp4", true},
		{"PDF", "document.pdf", true},
		{"Markdown", "note.md", false},
		{"Text", "file.txt", false},
		{"Unknown", "file.xyz", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsMediaFile(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsImageFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{"PNG", "image.png", true},
		{"JPG", "photo.jpg", true},
		{"GIF", "animation.gif", true},
		{"WEBP", "image.webp", true},
		{"MP3", "audio.mp3", false},
		{"MP4", "video.mp4", false},
		{"PDF", "document.pdf", false},
		{"Markdown", "note.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsImageFile(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMediaService_GetMedia(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewMediaService(tmpDir)

	// Test with non-existent file
	data, contentType, err := service.GetMedia("nonexistent.png")
	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Empty(t, contentType)
}

func TestMediaService_MoveMedia(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewMediaService(tmpDir)

	// Create source file
	srcContent := []byte("test content")
	srcPath := filepath.Join(tmpDir, "src.png")
	err := os.WriteFile(srcPath, srcContent, 0644)
	assert.NoError(t, err)

	// Create destination directory
	destDir := filepath.Join(tmpDir, "dest")
	err = os.MkdirAll(destDir, 0755)
	assert.NoError(t, err)

	destPath := filepath.Join(destDir, "dest.png")

	// Move the file
	err = service.MoveMedia("src.png", "dest/dest.png")
	assert.NoError(t, err)

	// Verify source is gone
	_, err = os.Stat(srcPath)
	assert.True(t, os.IsNotExist(err))

	// Verify destination exists
	_, err = os.Stat(destPath)
	assert.NoError(t, err)

	// Verify content
	content, err := os.ReadFile(destPath)
	assert.NoError(t, err)
	assert.Equal(t, srcContent, content)
}

func TestMediaService_MoveMedia_SourceNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewMediaService(tmpDir)

	err := service.MoveMedia("nonexistent.png", "dest.png")
	assert.Error(t, err)
}

func TestGetFileContentType(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{"PNG", "image.png", "image/png"},
		{"JPEG", "photo.jpg", "image/jpeg"},
		{"PDF", "document.pdf", "application/pdf"},
		{"MP3", "audio.mp3", "audio/mpeg"},
		{"MP4", "video.mp4", "video/mp4"},
		{"Unknown", "file.xyz", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetFileContentType(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}
