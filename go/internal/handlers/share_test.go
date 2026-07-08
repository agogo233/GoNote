package handlers

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	"gonote/internal/models/config"
	"gonote/internal/services"
)

func TestShareHandler_Create(t *testing.T) {
	t.Run("creates share token for note", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg := &config.Config{
			Storage: config.StorageConfig{
				NotesDir: tmpDir,
			},
		}
		shareSvc := services.NewShareService(tmpDir)
		exportSvc := services.NewExportService(tmpDir, "themes")
		handler := NewShareHandler(shareSvc, exportSvc, cfg, "themes")

		app := fiber.New()
		app.Post("/api/share/*", handler.Create)

		req := httptest.NewRequest("POST", "/api/share/notes/test.md", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.True(t, result["success"].(bool))
		assert.NotEmpty(t, result["token"])
		assert.Equal(t, "notes/test.md", result["path"])
	})

	t.Run("creates share token with theme", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg := &config.Config{
			Storage: config.StorageConfig{
				NotesDir: tmpDir,
			},
		}
		shareSvc := services.NewShareService(tmpDir)
		exportSvc := services.NewExportService(tmpDir, "themes")
		handler := NewShareHandler(shareSvc, exportSvc, cfg, "themes")

		app := fiber.New()
		app.Post("/api/share/*", handler.Create)

		req := httptest.NewRequest("POST", "/api/share/notes/test.md?theme=dark", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, "dark", result["theme"])
	})

	t.Run("handles URL-encoded paths", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg := &config.Config{
			Storage: config.StorageConfig{
				NotesDir: tmpDir,
			},
		}
		shareSvc := services.NewShareService(tmpDir)
		exportSvc := services.NewExportService(tmpDir, "themes")
		handler := NewShareHandler(shareSvc, exportSvc, cfg, "themes")

		app := fiber.New()
		app.Post("/api/share/*", handler.Create)

		// URL-encoded Chinese characters
		encodedPath := url.PathEscape("笔记/测试.md")
		req := httptest.NewRequest("POST", "/api/share/"+encodedPath, nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, "笔记/测试.md", result["path"])
	})
}

func TestShareHandler_GetStatus(t *testing.T) {
	t.Run("returns not shared for new note", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg := &config.Config{
			Storage: config.StorageConfig{
				NotesDir: tmpDir,
			},
		}
		shareSvc := services.NewShareService(tmpDir)
		exportSvc := services.NewExportService(tmpDir, "themes")
		handler := NewShareHandler(shareSvc, exportSvc, cfg, "themes")

		app := fiber.New()
		app.Get("/api/share/*", handler.GetStatus)

		req := httptest.NewRequest("GET", "/api/share/notes/test.md", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.False(t, result["shared"].(bool))
	})

	t.Run("returns shared status after creating token", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg := &config.Config{
			Storage: config.StorageConfig{
				NotesDir: tmpDir,
			},
		}
		shareSvc := services.NewShareService(tmpDir)
		exportSvc := services.NewExportService(tmpDir, "themes")
		handler := NewShareHandler(shareSvc, exportSvc, cfg, "themes")

		// First create a token
		_, err := shareSvc.CreateShareToken("notes/test.md", "dark")
		assert.NoError(t, err)

		app := fiber.New()
		app.Get("/api/share/*", handler.GetStatus)

		req := httptest.NewRequest("GET", "/api/share/notes/test.md", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.True(t, result["shared"].(bool))
		assert.NotEmpty(t, result["token"])
	})
}

func TestShareHandler_Revoke(t *testing.T) {
	t.Run("revokes share token", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg := &config.Config{
			Storage: config.StorageConfig{
				NotesDir: tmpDir,
			},
		}
		shareSvc := services.NewShareService(tmpDir)
		exportSvc := services.NewExportService(tmpDir, "themes")
		handler := NewShareHandler(shareSvc, exportSvc, cfg, "themes")

		// First create a token
		_, err := shareSvc.CreateShareToken("notes/test.md", "dark")
		assert.NoError(t, err)

		app := fiber.New()
		app.Delete("/api/share/*", handler.Revoke)

		req := httptest.NewRequest("DELETE", "/api/share/notes/test.md", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.True(t, result["success"].(bool))
	})
}

func TestShareHandler_ListSharedNotes(t *testing.T) {
	t.Run("returns empty list when no shares", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg := &config.Config{
			Storage: config.StorageConfig{
				NotesDir: tmpDir,
			},
		}
		shareSvc := services.NewShareService(tmpDir)
		exportSvc := services.NewExportService(tmpDir, "themes")
		handler := NewShareHandler(shareSvc, exportSvc, cfg, "themes")

		app := fiber.New()
		app.Get("/api/shared", handler.ListSharedNotes)

		req := httptest.NewRequest("GET", "/api/shared", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		paths := result["paths"].([]interface{})
		assert.Equal(t, 0, len(paths))
	})

	t.Run("returns list of shared notes", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg := &config.Config{
			Storage: config.StorageConfig{
				NotesDir: tmpDir,
			},
		}
		shareSvc := services.NewShareService(tmpDir)
		exportSvc := services.NewExportService(tmpDir, "themes")
		handler := NewShareHandler(shareSvc, exportSvc, cfg, "themes")

		// Create some tokens
		shareSvc.CreateShareToken("notes/note1.md", "light")
		shareSvc.CreateShareToken("notes/note2.md", "dark")

		app := fiber.New()
		app.Get("/api/shared", handler.ListSharedNotes)

		req := httptest.NewRequest("GET", "/api/shared", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		paths := result["paths"].([]interface{})
		assert.Equal(t, 2, len(paths))
	})
}

func TestShareHandler_ViewSharedNote(t *testing.T) {
	t.Run("returns 404 for non-existent token", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg := &config.Config{
			Storage: config.StorageConfig{
				NotesDir: tmpDir,
			},
		}
		shareSvc := services.NewShareService(tmpDir)
		exportSvc := services.NewExportService(tmpDir, "themes")
		handler := NewShareHandler(shareSvc, exportSvc, cfg, "themes")

		app := fiber.New()
		app.Get("/share/:token", handler.ViewSharedNote)

		req := httptest.NewRequest("GET", "/share/nonexistent-token", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode)
	})

	t.Run("returns HTML for valid token", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a test note
		noteDir := filepath.Join(tmpDir, "notes")
		os.MkdirAll(noteDir, 0755)
		notePath := filepath.Join(noteDir, "test.md")
		os.WriteFile(notePath, []byte("# Test Note\n\nThis is a test."), 0644)

		cfg := &config.Config{
			Storage: config.StorageConfig{
				NotesDir: tmpDir,
			},
		}
		shareSvc := services.NewShareService(tmpDir)
		exportSvc := services.NewExportService(tmpDir, "themes")
		handler := NewShareHandler(shareSvc, exportSvc, cfg, "themes")

		// Create a token
		token, err := shareSvc.CreateShareToken("notes/test.md", "light")
		assert.NoError(t, err)

		app := fiber.New()
		app.Get("/share/:token", handler.ViewSharedNote)

		req := httptest.NewRequest("GET", "/share/"+token, nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"))

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "Test Note")
	})

	t.Run("returns 400 for path traversal attempt", func(t *testing.T) {
		tmpDir := t.TempDir()

		cfg := &config.Config{
			Storage: config.StorageConfig{
				NotesDir: tmpDir,
			},
		}
		exportSvc := services.NewExportService(tmpDir, "themes")

		// Manually create a malicious token in the tokens file
		tokensFile := filepath.Join(tmpDir, ".share-tokens.json")
		maliciousContent := `{"evil-token": {"path": "../../../etc/passwd", "theme": "light", "created": "2024-01-01T00:00:00Z"}}`
		os.WriteFile(tokensFile, []byte(maliciousContent), 0644)

		// Re-create service to load the malicious token
		shareSvc := services.NewShareService(tmpDir)
		handler := NewShareHandler(shareSvc, exportSvc, cfg, "themes")

		app := fiber.New()
		app.Get("/share/:token", handler.ViewSharedNote)

		req := httptest.NewRequest("GET", "/share/evil-token", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		// Should return 400 for invalid path (security check)
		assert.Equal(t, 400, resp.StatusCode)
	})
}

func TestNewShareHandler(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Storage: config.StorageConfig{
			NotesDir: tmpDir,
		},
	}
	shareSvc := services.NewShareService(tmpDir)
	exportSvc := services.NewExportService(tmpDir, "themes")
	handler := NewShareHandler(shareSvc, exportSvc, cfg, "themes")

	assert.NotNil(t, handler)
	assert.Equal(t, cfg, handler.config)
}
