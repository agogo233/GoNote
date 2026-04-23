package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	"gonote/internal/models/config"
	"gonote/internal/models"
	"gonote/internal/services"
)

func TestNoteList(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Storage: config.StorageConfig{
			NotesDir: tmpDir,
		},
	}
	svc := services.NewNoteService(tmpDir)
	handler := NewNoteHandler(svc, cfg)

	app := fiber.New()
	app.Get("/api/notes", handler.List)

	t.Run("returns empty list with default pagination", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/notes", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), `"notes":[]`)
		assert.Contains(t, string(body), `"pagination"`)
		assert.Contains(t, string(body), `"page":1`)
		assert.Contains(t, string(body), `"limit":50`)
	})

	t.Run("respects page parameter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/notes?page=2", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), `"page":2`)
	})

	t.Run("respects limit parameter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/notes?limit=10", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), `"limit":10`)
	})

	t.Run("uses default values when no params", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/notes", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), `"page":1`)
		assert.Contains(t, string(body), `"limit":50`)
	})

	t.Run("returns pagination metadata", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			path := filepath.Join(tmpDir, "note"+string(rune('a'+i))+".md")
			os.WriteFile(path, []byte("# Test Note"), 0644)
		}
		svc.InvalidateCache()

		req := httptest.NewRequest("GET", "/api/notes?page=1&limit=2", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), `"total":5`)
		assert.Contains(t, string(body), `"total_pages":3`)
		assert.Contains(t, string(body), `"has_next":true`)
		assert.Contains(t, string(body), `"has_prev":false`)
	})
}

func TestNoteHandler_Get(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Storage: config.StorageConfig{
			NotesDir: tmpDir,
		},
	}
	svc := services.NewNoteService(tmpDir)
	handler := NewNoteHandler(svc, cfg)

	app := fiber.New()
	app.Get("/api/notes/*", handler.Get)

	t.Run("returns empty note for non-existent file", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/notes/nonexistent.md", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var note models.NoteContent
		err = json.Unmarshal(body, &note)
		assert.NoError(t, err)
		assert.Equal(t, "nonexistent.md", note.Path)
		assert.Equal(t, "", note.Content)
	})

	t.Run("returns existing note with content", func(t *testing.T) {
		notePath := filepath.Join(tmpDir, "test.md")
		os.WriteFile(notePath, []byte("# Hello World"), 0644)

		req := httptest.NewRequest("GET", "/api/notes/test.md", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var note models.NoteContent
		err = json.Unmarshal(body, &note)
		assert.NoError(t, err)
		assert.Equal(t, "test.md", note.Path)
		assert.Equal(t, "# Hello World", note.Content)
	})
}

func TestNoteHandler_CreateOrUpdate(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Storage: config.StorageConfig{
			NotesDir: tmpDir,
		},
	}
	svc := services.NewNoteService(tmpDir)
	handler := NewNoteHandler(svc, cfg)

	app := fiber.New()
	app.Put("/api/notes/*", handler.CreateOrUpdate)

	t.Run("creates new note", func(t *testing.T) {
		body := map[string]string{"content": "# New Note\n\nThis is new."}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest("PUT", "/api/notes/new-note.md", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		var result models.NoteSaveResponse
		err = json.Unmarshal(respBody, &result)
		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, "new-note.md", result.Path)

		// Verify file was created
		_, err = os.Stat(filepath.Join(tmpDir, "new-note.md"))
		assert.NoError(t, err)
	})

	t.Run("updates existing note", func(t *testing.T) {
		notePath := filepath.Join(tmpDir, "existing.md")
		os.WriteFile(notePath, []byte("# Old"), 0644)

		body := map[string]string{"content": "# Updated"}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest("PUT", "/api/notes/existing.md", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		var result models.NoteSaveResponse
		err = json.Unmarshal(respBody, &result)
		assert.NoError(t, err)
		assert.True(t, result.Success)

		// Verify content was updated
		content, err := os.ReadFile(notePath)
		assert.NoError(t, err)
		assert.Equal(t, "# Updated", string(content))
	})
}

func TestNoteHandler_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Storage: config.StorageConfig{
			NotesDir: tmpDir,
		},
	}
	svc := services.NewNoteService(tmpDir)
	handler := NewNoteHandler(svc, cfg)

	app := fiber.New()
	app.Delete("/api/notes/*", handler.Delete)

	t.Run("deletes existing note", func(t *testing.T) {
		notePath := filepath.Join(tmpDir, "to-delete.md")
		os.WriteFile(notePath, []byte("# Delete Me"), 0644)

		req := httptest.NewRequest("DELETE", "/api/notes/to-delete.md", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		var result models.APIResponse
		err = json.Unmarshal(respBody, &result)
		assert.NoError(t, err)
		assert.True(t, result.Success)

		// Verify file was deleted
		_, err = os.Stat(notePath)
		assert.True(t, os.IsNotExist(err))
	})
}

func TestNoteHandler_Move(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Storage: config.StorageConfig{
			NotesDir: tmpDir,
		},
	}
	svc := services.NewNoteService(tmpDir)
	handler := NewNoteHandler(svc, cfg)

	app := fiber.New()
	app.Post("/api/notes/move", handler.Move)

	t.Run("moves note to new location", func(t *testing.T) {
		oldPath := filepath.Join(tmpDir, "old.md")
		os.WriteFile(oldPath, []byte("# Move Me"), 0644)

		body := map[string]string{"oldPath": "old.md", "newPath": "new.md"}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/api/notes/move", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		var result models.NoteMoveResponse
		err = json.Unmarshal(respBody, &result)
		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, "old.md", result.OldPath)
		assert.Equal(t, "new.md", result.NewPath)

		// Verify old file is gone
		_, err = os.Stat(oldPath)
		assert.True(t, os.IsNotExist(err))

		// Verify new file exists
		newPath := filepath.Join(tmpDir, "new.md")
		_, err = os.Stat(newPath)
		assert.NoError(t, err)
	})
}
