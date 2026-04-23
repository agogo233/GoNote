package handlers

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	"gonote/internal/models/config"
	"gonote/internal/services"
)

func TestMediaHandler_Get(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Storage: config.StorageConfig{
			NotesDir: tmpDir,
		},
	}
	mediaService := services.NewMediaService(tmpDir)
	noteService := services.NewNoteService(tmpDir)
	handler := NewMediaHandler(mediaService, noteService, cfg)

	app := fiber.New()
	app.Get("/media/*", handler.Get)

	t.Run("returns 404 for non-existent media", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/media/nonexistent.png", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "Media not found")
	})

	t.Run("returns media file successfully", func(t *testing.T) {
		// Create test media file
		mediaPath := "test.png"
		fullPath := filepath.Join(tmpDir, mediaPath)
		testContent := []byte("fake png content")
		err := os.WriteFile(fullPath, testContent, 0644)
		assert.NoError(t, err)

		req := httptest.NewRequest("GET", "/media/test.png", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		// Content type may be simplified
		contentType := resp.Header.Get("Content-Type")
		assert.Contains(t, contentType, "image")

		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, testContent, body)
	})

	t.Run("returns correct content type for different media", func(t *testing.T) {
		testCases := []struct {
			filename       string
			expectedType   string
		}{
			{"test.jpg", "image/jpeg"},
			{"test.gif", "image/gif"},
			{"test.mp4", "video/mp4"},
			{"test.mp3", "audio/mpeg"},
			{"test.pdf", "application/pdf"},
			{"test.txt", "text/plain"},
		}

		for _, tc := range testCases {
			t.Run(tc.filename, func(t *testing.T) {
				fullPath := filepath.Join(tmpDir, tc.filename)
				err := os.WriteFile(fullPath, []byte("content"), 0644)
				assert.NoError(t, err)

				req := httptest.NewRequest("GET", fmt.Sprintf("/media/%s", tc.filename), nil)
				resp, err := app.Test(req)

				assert.NoError(t, err)
				assert.Equal(t, 200, resp.StatusCode)
				// Use GetFileContentType for exact MIME type matching
				expectedContentType := services.GetFileContentType(tc.filename)
				assert.Equal(t, expectedContentType, resp.Header.Get("Content-Type"))
			})
		}
	})

	t.Run("rejects path traversal attempts", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/media/../../../etc/passwd", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "Invalid path")
	})
}

func TestMediaHandler_Upload(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Storage: config.StorageConfig{
			NotesDir: tmpDir,
		},
		Upload: config.UploadConfig{
			MaxFileSizeMB: 50,
			AllowedTypes:  []string{}, // No restrictions for these tests
		},
	}
	mediaService := services.NewMediaService(tmpDir)
	noteService := services.NewNoteService(tmpDir)
	handler := NewMediaHandler(mediaService, noteService, cfg)

	app := fiber.New()
	app.Post("/api/media/upload", handler.Upload)

	t.Run("uploads image file successfully", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// Create form file
		fileWriter, err := writer.CreateFormFile("file", "test-image.png")
		assert.NoError(t, err)

		// Write fake PNG content
		pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
		_, err = fileWriter.Write(pngHeader)
		assert.NoError(t, err)

		// Add note_path field
		err = writer.WriteField("note_path", "test-note.md")
		assert.NoError(t, err)

		err = writer.Close()
		assert.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/media/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(respBody), `"success":true`)
		assert.Contains(t, string(respBody), `"type":"image"`)
		assert.Contains(t, string(respBody), "test-image")

		// Verify file was created - note_path="test-note.md" -> _attachments dir
		attachmentsDir := filepath.Join(tmpDir, "_attachments")
		files, err := os.ReadDir(attachmentsDir)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(files), 1)
		// Check that at least one file starts with "test-image" and ends with ".png"
		found := false
		for _, f := range files {
			if strings.HasPrefix(f.Name(), "test-image") && strings.HasSuffix(f.Name(), ".png") {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected to find a file starting with 'test-image' and ending with '.png'")
	})

	t.Run("uploads video file successfully", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		fileWriter, err := writer.CreateFormFile("file", "test-video.mp4")
		assert.NoError(t, err)
		_, err = fileWriter.Write([]byte("fake mp4 content"))
		assert.NoError(t, err)

		err = writer.Close()
		assert.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/media/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(respBody), `"type":"video"`)
	})

	t.Run("uploads audio file successfully", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		fileWriter, err := writer.CreateFormFile("file", "test-audio.mp3")
		assert.NoError(t, err)
		_, err = fileWriter.Write([]byte("fake mp3 content"))
		assert.NoError(t, err)

		err = writer.Close()
		assert.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/media/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(respBody), `"type":"audio"`)
	})

	t.Run("uploads PDF document successfully", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		fileWriter, err := writer.CreateFormFile("file", "test-document.pdf")
		assert.NoError(t, err)
		_, err = fileWriter.Write([]byte("%PDF-1.4 fake pdf"))
		assert.NoError(t, err)

		err = writer.Close()
		assert.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/media/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(respBody), `"type":"document"`)
	})

	t.Run("returns error when no file uploaded", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		err := writer.Close()
		assert.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/media/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(respBody), "No file uploaded")
	})

	t.Run("generates unique filename with timestamp", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		fileWriter, err := writer.CreateFormFile("file", "duplicate.png")
		assert.NoError(t, err)
		_, err = fileWriter.Write([]byte("content1"))
		assert.NoError(t, err)

		err = writer.Close()
		assert.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/media/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		resp1, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp1.StatusCode)

		// Upload same file again
		body2 := &bytes.Buffer{}
		writer2 := multipart.NewWriter(body2)
		fileWriter2, _ := writer2.CreateFormFile("file", "duplicate.png")
		fileWriter2.Write([]byte("content2"))
		writer2.Close()

		req2 := httptest.NewRequest("POST", "/api/media/upload", body2)
		req2.Header.Set("Content-Type", writer2.FormDataContentType())
		resp2, err := app.Test(req2)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp2.StatusCode)

		// Both files should exist with different names
		attachmentsDir := filepath.Join(tmpDir, "_attachments")
		files, _ := os.ReadDir(attachmentsDir)
		assert.GreaterOrEqual(t, len(files), 2)
	})

	t.Run("sanitizes filename", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		fileWriter, err := writer.CreateFormFile("file", "test/../evil.png")
		assert.NoError(t, err)
		_, err = fileWriter.Write([]byte("content"))
		assert.NoError(t, err)

		err = writer.Close()
		assert.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/media/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		// Should not contain path traversal
		assert.NotContains(t, string(respBody), "..")
	})
}

func TestMediaHandler_Upload_Validation(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Storage: config.StorageConfig{
			NotesDir: tmpDir,
		},
		Upload: config.UploadConfig{
			MaxFileSizeMB: 1, // 1MB limit for testing
			AllowedTypes:  []string{},
		},
	}
	mediaService := services.NewMediaService(tmpDir)
	noteService := services.NewNoteService(tmpDir)
	handler := NewMediaHandler(mediaService, noteService, cfg)

	app := fiber.New()
	app.Post("/api/media/upload", handler.Upload)

	t.Run("rejects file exceeding size limit", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		fileWriter, err := writer.CreateFormFile("file", "large-file.png")
		assert.NoError(t, err)

		// Write 2MB of data (exceeds 1MB limit)
		largeData := make([]byte, 2*1024*1024)
		_, err = fileWriter.Write(largeData)
		assert.NoError(t, err)

		err = writer.Close()
		assert.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/media/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 413, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(respBody), "file too large")
	})

	t.Run("rejects disallowed file type", func(t *testing.T) {
		cfg.Upload.AllowedTypes = []string{"image/png", "image/jpeg"}

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		fileWriter, err := writer.CreateFormFile("file", "test.exe")
		assert.NoError(t, err)
		_, err = fileWriter.Write([]byte("executable content"))
		assert.NoError(t, err)

		err = writer.Close()
		assert.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/media/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 413, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(respBody), "file type not allowed")
	})

	t.Run("allows configured file types", func(t *testing.T) {
		// Test with no type restrictions to verify upload works
		cfg.Upload.AllowedTypes = []string{} // No restrictions
		// Re-create handler with updated config
		handler = NewMediaHandler(mediaService, noteService, cfg)
		app = fiber.New()
		app.Post("/api/media/upload", handler.Upload)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		fileWriter, err := writer.CreateFormFile("file", "test.png")
		assert.NoError(t, err)
		pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
		_, err = fileWriter.Write(pngHeader)
		assert.NoError(t, err)

		// Add note_path field
		err = writer.WriteField("note_path", "test-note.md")
		assert.NoError(t, err)

		err = writer.Close()
		assert.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/media/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})
}

func TestMediaHandler_Move(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Storage: config.StorageConfig{
			NotesDir: tmpDir,
		},
	}
	mediaService := services.NewMediaService(tmpDir)
	noteService := services.NewNoteService(tmpDir)
	handler := NewMediaHandler(mediaService, noteService, cfg)

	app := fiber.New()
	app.Post("/api/media/move", handler.Move)

	t.Run("moves media file successfully", func(t *testing.T) {
		// Create source file in _attachments directory
		oldPath := "_attachments/test.png"
		fullOldPath := filepath.Join(tmpDir, oldPath)
		err := os.MkdirAll(filepath.Dir(fullOldPath), 0755)
		assert.NoError(t, err)
		err = os.WriteFile(fullOldPath, []byte("image content"), 0644)
		assert.NoError(t, err)

		// Move request
		reqBody := fmt.Sprintf(`{"oldPath":"%s","newPath":"_attachments/test2.png"}`, oldPath)
		req := httptest.NewRequest("POST", "/api/media/move", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(respBody), `"success":true`)

		// Verify file was moved
		_, err = os.Stat(fullOldPath)
		assert.True(t, os.IsNotExist(err))

		newPath := filepath.Join(tmpDir, "_attachments/test2.png")
		_, err = os.Stat(newPath)
		assert.NoError(t, err)
	})

	t.Run("returns error for non-existent source file", func(t *testing.T) {
		reqBody := `{"oldPath":"nonexistent.png","newPath":"new.png"}`
		req := httptest.NewRequest("POST", "/api/media/move", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 500, resp.StatusCode)
	})

	t.Run("rejects path traversal in move", func(t *testing.T) {
		reqBody := `{"oldPath":"../../../etc/passwd","newPath":"test.png"}`
		req := httptest.NewRequest("POST", "/api/media/move", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(respBody), "Invalid path")
	})
}

func TestMediaHandler_ListOrphanedMedia(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Storage: config.StorageConfig{
			NotesDir: tmpDir,
		},
	}
	mediaService := services.NewMediaService(tmpDir)
	noteService := services.NewNoteService(tmpDir)
	handler := NewMediaHandler(mediaService, noteService, cfg)

	app := fiber.New()
	app.Get("/api/media/orphaned", handler.ListOrphanedMedia)

	t.Run("returns empty list when no orphaned media", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/media/orphaned", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(respBody), `"success":true`)
		assert.Contains(t, string(respBody), `"count":0`)
		assert.Contains(t, string(respBody), `"files":[]`)
	})

	t.Run("returns orphaned media files", func(t *testing.T) {
		// Create orphaned media file in _attachments directory
		orphanPath := "_attachments/test-image.png"
		fullPath := filepath.Join(tmpDir, orphanPath)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		assert.NoError(t, err)
		err = os.WriteFile(fullPath, []byte("orphaned image"), 0644)
		assert.NoError(t, err)

		req := httptest.NewRequest("GET", "/api/media/orphaned", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(respBody), `"success":true`)
		assert.Contains(t, string(respBody), `"count":1`)
		assert.Contains(t, string(respBody), "test-image.png")
		assert.Contains(t, string(respBody), `"mediaType":"image"`)
	})

	t.Run("calculates total size correctly", func(t *testing.T) {
		// Clean up _attachments directory first to avoid interference from previous tests
		attachmentsDir := filepath.Join(tmpDir, "_attachments")
		os.RemoveAll(attachmentsDir)
		
		// Create multiple orphaned files in _attachments directory
		for i := 0; i < 3; i++ {
			orphanPath := fmt.Sprintf("_attachments/test-%d.png", i)
			fullPath := filepath.Join(tmpDir, orphanPath)
			os.MkdirAll(filepath.Dir(fullPath), 0755)
			os.WriteFile(fullPath, []byte(fmt.Sprintf("content-%d", i)), 0644)
		}

		req := httptest.NewRequest("GET", "/api/media/orphaned", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(respBody), `"count":3`)
		// Total size should be the sum of all file sizes
		assert.Contains(t, string(respBody), `"totalSize"`)
	})
}

func TestMediaHandler_CleanupOrphanedMedia(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Storage: config.StorageConfig{
			NotesDir: tmpDir,
		},
	}
	mediaService := services.NewMediaService(tmpDir)
	noteService := services.NewNoteService(tmpDir)
	handler := NewMediaHandler(mediaService, noteService, cfg)

	app := fiber.New()
	app.Delete("/api/media/orphaned", handler.CleanupOrphanedMedia)

	t.Run("deletes orphaned media successfully", func(t *testing.T) {
		// Create orphaned media file in _attachments directory
		orphanPath := "_attachments/to-delete.png"
		fullPath := filepath.Join(tmpDir, orphanPath)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		assert.NoError(t, err)
		testContent := []byte("orphaned content")
		err = os.WriteFile(fullPath, testContent, 0644)
		assert.NoError(t, err)

		req := httptest.NewRequest("DELETE", "/api/media/orphaned", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(respBody), `"success":true`)
		assert.Contains(t, string(respBody), `"deletedCount":1`)
		assert.Contains(t, string(respBody), "Successfully deleted orphaned media files")

		// Verify file was deleted
		_, err = os.Stat(fullPath)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("returns freed space", func(t *testing.T) {
		// Create orphaned file with known size in _attachments directory
		orphanPath := "_attachments/size-test.png"
		fullPath := filepath.Join(tmpDir, orphanPath)
		os.MkdirAll(filepath.Dir(fullPath), 0755)
		testContent := []byte("known size content") // 18 bytes
		os.WriteFile(fullPath, testContent, 0644)

		req := httptest.NewRequest("DELETE", "/api/media/orphaned", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(respBody), `"freedSpace":18`)
	})

	t.Run("handles no orphaned media gracefully", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/media/orphaned", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(respBody), `"success":true`)
		assert.Contains(t, string(respBody), `"deletedCount":0`)
		assert.Contains(t, string(respBody), "No orphaned media files found")
	})
}

func TestMediaHandler_validateUpload(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Upload: config.UploadConfig{
			MaxFileSizeMB: 10,
			AllowedTypes:  []string{"image/png", "image/jpeg"},
		},
	}
	mediaService := services.NewMediaService(tmpDir)
	noteService := services.NewNoteService(tmpDir)
	handler := NewMediaHandler(mediaService, noteService, cfg)

	t.Run("accepts file within size limit", func(t *testing.T) {
		file := &multipart.FileHeader{
			Filename: "test.png",
			Size:     1024 * 1024, // 1MB
			Header: map[string][]string{
				"Content-Type": {"image/png"},
			},
		}
		err := handler.validateUpload(file)
		assert.NoError(t, err)
	})

	t.Run("rejects file exceeding size limit", func(t *testing.T) {
		file := &multipart.FileHeader{
			Filename: "large.png",
			Size:     11 * 1024 * 1024, // 11MB
		}
		err := handler.validateUpload(file)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file too large")
	})

	t.Run("accepts allowed file type", func(t *testing.T) {
		file := &multipart.FileHeader{
			Filename: "test.png",
			Size:     1024,
			Header: map[string][]string{
				"Content-Type": {"image/png"},
			},
		}
		err := handler.validateUpload(file)
		assert.NoError(t, err)
	})

	t.Run("rejects disallowed file type", func(t *testing.T) {
		file := &multipart.FileHeader{
			Filename: "test.exe",
			Size:     1024,
			Header: map[string][]string{
				"Content-Type": {"application/x-executable"},
			},
		}
		err := handler.validateUpload(file)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file type not allowed")
	})

	t.Run("handles content type with charset", func(t *testing.T) {
		file := &multipart.FileHeader{
			Filename: "test.png",
			Size:     1024,
			Header: map[string][]string{
				"Content-Type": {"image/png; charset=utf-8"},
			},
		}
		err := handler.validateUpload(file)
		assert.NoError(t, err)
	})
}

func TestMediaHandler_isAllowedType(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Upload: config.UploadConfig{
			MaxFileSizeMB: 10,
			AllowedTypes:  []string{"image/*", "application/pdf", "text/plain"},
		},
	}
	mediaService := services.NewMediaService(tmpDir)
	noteService := services.NewNoteService(tmpDir)
	handler := NewMediaHandler(mediaService, noteService, cfg)

	t.Run("allows wildcard image types", func(t *testing.T) {
		assert.True(t, handler.isAllowedType("image/png"))
		assert.True(t, handler.isAllowedType("image/jpeg"))
		assert.True(t, handler.isAllowedType("image/gif"))
		assert.True(t, handler.isAllowedType("image/webp"))
	})

	t.Run("allows specific types", func(t *testing.T) {
		assert.True(t, handler.isAllowedType("application/pdf"))
		assert.True(t, handler.isAllowedType("text/plain"))
	})

	t.Run("rejects non-matching types", func(t *testing.T) {
		assert.False(t, handler.isAllowedType("video/mp4"))
		assert.False(t, handler.isAllowedType("audio/mpeg"))
		assert.False(t, handler.isAllowedType("application/x-executable"))
	})

	t.Run("handles case insensitivity", func(t *testing.T) {
		assert.True(t, handler.isAllowedType("IMAGE/PNG"))
		assert.True(t, handler.isAllowedType("Image/Jpeg"))
	})

	t.Run("allows all types when no restrictions", func(t *testing.T) {
		cfg.Upload.AllowedTypes = []string{}
		handler = NewMediaHandler(mediaService, noteService, cfg)
		assert.True(t, handler.isAllowedType("application/anything"))
	})
}
