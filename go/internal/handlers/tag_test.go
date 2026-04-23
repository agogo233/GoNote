package handlers

import (
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	"gonote/internal/models/config"
	"gonote/internal/services"
)

func TestTagHandler_List(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Storage: config.StorageConfig{
			NotesDir: tmpDir,
		},
	}
	noteService := services.NewNoteService(tmpDir)
	tagService := services.NewTagService(noteService, tmpDir)
	handler := NewTagHandler(tagService, cfg)

	app := fiber.New()
	app.Get("/api/tags", handler.List)

	t.Run("returns empty tags when no notes", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/tags", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), `"tags":{}`)
	})

	t.Run("returns tags with counts", func(t *testing.T) {
		// Create notes with tags
		notes := []struct {
			path    string
			content string
		}{
			{"note1.md", `---
tags: [tag1, tag2]
---
Content 1`},
			{"note2.md", `---
tags: [tag2, tag3]
---
Content 2`},
			{"note3.md", `---
tags: [tag1, tag3]
---
Content 3`},
		}

		for _, note := range notes {
			fullPath := filepath.Join(tmpDir, note.path)
			err := os.WriteFile(fullPath, []byte(note.content), 0644)
			assert.NoError(t, err)
		}

		// Clear caches to ensure fresh data is read
		noteService.ClearCache()
		tagService.ClearCache()

		req := httptest.NewRequest("GET", "/api/tags", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), `"tag1":2`)
		assert.Contains(t, string(body), `"tag2":2`)
		assert.Contains(t, string(body), `"tag3":2`)
	})

	t.Run("handles notes without tags", func(t *testing.T) {
		content := `---
title: No Tags
---
Content without tags.`
		fullPath := filepath.Join(tmpDir, "no-tags.md")
		err := os.WriteFile(fullPath, []byte(content), 0644)
		assert.NoError(t, err)

		// Clear caches to ensure fresh data is read
		noteService.ClearCache()
		tagService.ClearCache()

		req := httptest.NewRequest("GET", "/api/tags", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		// Should still have the previous tags
		assert.Contains(t, string(body), `"tag1"`)
	})
}

func TestTagHandler_GetNotesByTag(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Storage: config.StorageConfig{
			NotesDir: tmpDir,
		},
	}
	noteService := services.NewNoteService(tmpDir)
	tagService := services.NewTagService(noteService, tmpDir)
	handler := NewTagHandler(tagService, cfg)

	app := fiber.New()
	app.Get("/api/tags/*", handler.GetNotesByTag)

	t.Run("returns empty list for non-existent tag", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/tags/nonexistent", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), `"count":0`)
		assert.Contains(t, string(body), `"notes":[]`)
	})

	t.Run("returns notes with specific tag", func(t *testing.T) {
		// Create notes with different tags
		notes := []struct {
			path    string
			content string
		}{
			{"note1.md", `---
tags: [programming, go]
---
# Go Programming`},
			{"note2.md", `---
tags: [programming, python]
---
# Python Programming`},
			{"note3.md", `---
tags: [personal]
---
# Personal Note`},
		}

		for _, note := range notes {
			fullPath := filepath.Join(tmpDir, note.path)
			err := os.WriteFile(fullPath, []byte(note.content), 0644)
			assert.NoError(t, err)
		}

		// Clear caches to ensure fresh data is read
		noteService.ClearCache()
		tagService.ClearCache()

		req := httptest.NewRequest("GET", "/api/tags/programming", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), `"tag":"programming"`)
		assert.Contains(t, string(body), `"count":2`)
		assert.Contains(t, string(body), "note1.md")
		assert.Contains(t, string(body), "note2.md")
		assert.NotContains(t, string(body), "note3.md")
	})

	t.Run("case insensitive tag matching", func(t *testing.T) {
		// Clear caches to ensure fresh data is read
		noteService.ClearCache()
		tagService.ClearCache()

		req := httptest.NewRequest("GET", "/api/tags/PROGRAMMING", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), `"count":2`)
	})

	t.Run("returns notes in subdirectories", func(t *testing.T) {
		// Create note in subdirectory
		subDir := filepath.Join(tmpDir, "subdir")
		err := os.MkdirAll(subDir, 0755)
		assert.NoError(t, err)

		content := `---
tags: [programming]
---
# Subdir Note`
		fullPath := filepath.Join(subDir, "subdir-note.md")
		err = os.WriteFile(fullPath, []byte(content), 0644)
		assert.NoError(t, err)

		// Clear caches to ensure fresh data is read
		noteService.ClearCache()
		tagService.ClearCache()

		req := httptest.NewRequest("GET", "/api/tags/programming", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), `"count":3`)
		assert.Contains(t, string(body), "subdir-note.md")
	})
}

func TestTagHandler_GetNotesByTag_PathParams(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Storage: config.StorageConfig{
			NotesDir: tmpDir,
		},
	}
	noteService := services.NewNoteService(tmpDir)
	tagService := services.NewTagService(noteService, tmpDir)
	handler := NewTagHandler(tagService, cfg)

	app := fiber.New()
	app.Get("/api/tags/*", handler.GetNotesByTag)

	t.Run("handles tag with hyphens", func(t *testing.T) {
		content := `---
tags: [my-tag]
---
Content`
		fullPath := filepath.Join(tmpDir, "note.md")
		os.WriteFile(fullPath, []byte(content), 0644)

		req := httptest.NewRequest("GET", "/api/tags/my-tag", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), `"tag":"my-tag"`)
	})

	t.Run("handles tag with underscores", func(t *testing.T) {
		content := `---
tags: [my_tag]
---
Content`
		fullPath := filepath.Join(tmpDir, "note2.md")
		os.WriteFile(fullPath, []byte(content), 0644)

		req := httptest.NewRequest("GET", "/api/tags/my_tag", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), `"tag":"my_tag"`)
	})
}

func TestTagHandler_Integration(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Storage: config.StorageConfig{
			NotesDir: tmpDir,
		},
	}
	noteService := services.NewNoteService(tmpDir)
	tagService := services.NewTagService(noteService, tmpDir)
	handler := NewTagHandler(tagService, cfg)

	app := fiber.New()
	app.Get("/api/tags", handler.List)
	app.Get("/api/tags/*", handler.GetNotesByTag)

	// Create a realistic set of notes
	scenario := []struct {
		path    string
		content string
	}{
		{"programming/golang.md", `---
title: Go Programming
tags: [programming, go, backend]
---
# Go Programming Language`},
		{"programming/python.md", `---
title: Python Programming
tags: [programming, python, backend]
---
# Python Programming Language`},
		{"personal/todo.md", `---
title: Todo
tags: [personal, todo]
---
# Todo List`},
		{"docs/api.md", `---
title: API Docs
tags: [docs, api]
---
# API Documentation`},
	}

	for _, s := range scenario {
		fullPath := filepath.Join(tmpDir, s.path)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		assert.NoError(t, err)
		err = os.WriteFile(fullPath, []byte(s.content), 0644)
		assert.NoError(t, err)
	}

	t.Run("list all tags", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/tags", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), `"programming":2`)
		assert.Contains(t, string(body), `"go":1`)
		assert.Contains(t, string(body), `"python":1`)
		assert.Contains(t, string(body), `"personal":1`)
		assert.Contains(t, string(body), `"docs":1`)
		assert.Contains(t, string(body), `"api":1`)
	})

	t.Run("get notes by tag", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/tags/backend", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), `"count":2`)
		assert.Contains(t, string(body), "golang.md")
		assert.Contains(t, string(body), "python.md")
	})
}
