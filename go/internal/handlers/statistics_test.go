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

func TestStatisticsHandler_GetStatistics(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Storage: config.StorageConfig{
			NotesDir: tmpDir,
		},
	}
	statsService := services.NewStatisticsService(tmpDir)
	handler := NewStatisticsHandler(statsService, cfg)

	app := fiber.New()
	app.Get("/api/statistics/*", handler.GetStatistics)

	t.Run("returns 404 for non-existent note", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/statistics/nonexistent.md", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 500, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "detail")
	})

	t.Run("returns statistics for existing note", func(t *testing.T) {
		content := `---
title: Test Note
---

# Hello World

This is a test note with some content.
It has multiple sentences. And words!

- List item 1
- List item 2

[Link](https://example.com)
`
		fullPath := filepath.Join(tmpDir, "test.md")
		err := os.WriteFile(fullPath, []byte(content), 0644)
		assert.NoError(t, err)

		req := httptest.NewRequest("GET", "/api/statistics/test.md", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), `"success":true`)
		assert.Contains(t, string(body), `"data"`)
		assert.Contains(t, string(body), `"words"`)
		assert.Contains(t, string(body), `"sentences"`)
	})

	t.Run("returns comprehensive statistics", func(t *testing.T) {
		content := `---
title: Comprehensive Note
tags: [test]
---

# Main Heading

This is a comprehensive test with various elements.

## Code

` + "```go" + `
fmt.Println("Hello")
` + "```" + `

## Tasks

- [ ] Task 1
- [x] Task 2

## Image

![Image](test.png)
`
		fullPath := filepath.Join(tmpDir, "comprehensive.md")
		err := os.WriteFile(fullPath, []byte(content), 0644)
		assert.NoError(t, err)

		req := httptest.NewRequest("GET", "/api/statistics/comprehensive.md", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), `"headings"`)
		assert.Contains(t, string(body), `"tasks"`)
		assert.Contains(t, string(body), `"code_blocks"`)
		assert.Contains(t, string(body), `"images"`)
	})
}

func TestStatisticsHandler_GetStatistics_PathHandling(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Storage: config.StorageConfig{
			NotesDir: tmpDir,
		},
	}
	statsService := services.NewStatisticsService(tmpDir)
	handler := NewStatisticsHandler(statsService, cfg)

	app := fiber.New()
	app.Get("/api/statistics/*", handler.GetStatistics)

	t.Run("handles note in subdirectory", func(t *testing.T) {
		subDir := filepath.Join(tmpDir, "subdir")
		err := os.MkdirAll(subDir, 0755)
		assert.NoError(t, err)

		content := "# Subdir Note\nContent here."
		fullPath := filepath.Join(subDir, "note.md")
		err = os.WriteFile(fullPath, []byte(content), 0644)
		assert.NoError(t, err)

		req := httptest.NewRequest("GET", "/api/statistics/subdir/note.md", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), `"success":true`)
	})

	t.Run("handles note with special characters in name", func(t *testing.T) {
		content := "# Special Note"
		fullPath := filepath.Join(tmpDir, "special-note.md")
		err := os.WriteFile(fullPath, []byte(content), 0644)
		assert.NoError(t, err)

		req := httptest.NewRequest("GET", "/api/statistics/special-note.md", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("rejects path traversal", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/statistics/../../../etc/passwd", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "Invalid path")
	})
}

func TestStatisticsHandler_GetStatistics_URLEncoding(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Storage: config.StorageConfig{
			NotesDir: tmpDir,
		},
	}
	statsService := services.NewStatisticsService(tmpDir)
	handler := NewStatisticsHandler(statsService, cfg)

	app := fiber.New()
	app.Get("/api/statistics/*", handler.GetStatistics)

	t.Run("handles URL encoded path", func(t *testing.T) {
		content := "# Test Note"
		fullPath := filepath.Join(tmpDir, "test note.md")
		err := os.WriteFile(fullPath, []byte(content), 0644)
		assert.NoError(t, err)

		req := httptest.NewRequest("GET", "/api/statistics/test%20note.md", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), `"success":true`)
	})

	t.Run("handles Chinese characters in path", func(t *testing.T) {
		content := "# 测试笔记"
		fullPath := filepath.Join(tmpDir, "测试.md")
		err := os.WriteFile(fullPath, []byte(content), 0644)
		assert.NoError(t, err)

		req := httptest.NewRequest("GET", "/api/statistics/%E6%B5%8B%E8%AF%95.md", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), `"success":true`)
	})
}

func TestStatisticsHandler_GetStatistics_StatisticsData(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Storage: config.StorageConfig{
			NotesDir: tmpDir,
		},
	}
	statsService := services.NewStatisticsService(tmpDir)
	handler := NewStatisticsHandler(statsService, cfg)

	app := fiber.New()
	app.Get("/api/statistics/*", handler.GetStatistics)

	t.Run("returns word count", func(t *testing.T) {
		content := "Hello World Test"
		fullPath := filepath.Join(tmpDir, "wordcount.md")
		os.WriteFile(fullPath, []byte(content), 0644)

		req := httptest.NewRequest("GET", "/api/statistics/wordcount.md", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), `"words":3`)
	})

	t.Run("returns line count", func(t *testing.T) {
		content := "Line 1\nLine 2\nLine 3"
		fullPath := filepath.Join(tmpDir, "linecount.md")
		os.WriteFile(fullPath, []byte(content), 0644)

		req := httptest.NewRequest("GET", "/api/statistics/linecount.md", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), `"lines":3`)
	})

	t.Run("returns heading counts", func(t *testing.T) {
		content := "# H1\n## H2\n### H3"
		fullPath := filepath.Join(tmpDir, "headings.md")
		os.WriteFile(fullPath, []byte(content), 0644)

		req := httptest.NewRequest("GET", "/api/statistics/headings.md", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), `"headings"`)
		assert.Contains(t, string(body), `"h1":1`)
		assert.Contains(t, string(body), `"h2":1`)
		assert.Contains(t, string(body), `"h3":1`)
	})

	t.Run("returns task counts", func(t *testing.T) {
		content := "- [ ] Task 1\n- [x] Task 2\n- [ ] Task 3"
		fullPath := filepath.Join(tmpDir, "tasks.md")
		os.WriteFile(fullPath, []byte(content), 0644)

		req := httptest.NewRequest("GET", "/api/statistics/tasks.md", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), `"tasks"`)
		assert.Contains(t, string(body), `"total":3`)
		assert.Contains(t, string(body), `"completed":1`)
		assert.Contains(t, string(body), `"pending":2`)
	})

	t.Run("returns link counts", func(t *testing.T) {
		content := "[Link](https://example.com)\n[[Wikilink]]"
		fullPath := filepath.Join(tmpDir, "links.md")
		os.WriteFile(fullPath, []byte(content), 0644)

		req := httptest.NewRequest("GET", "/api/statistics/links.md", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), `"wikilinks":1`)
	})

	t.Run("returns code block count", func(t *testing.T) {
		content := "```\ncode\n```\n\n```\nmore code\n```"
		fullPath := filepath.Join(tmpDir, "code.md")
		os.WriteFile(fullPath, []byte(content), 0644)

		req := httptest.NewRequest("GET", "/api/statistics/code.md", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), `"code_blocks":2`)
	})

	t.Run("returns image count", func(t *testing.T) {
		content := "![Image1](img1.png)\n![Image2](img2.png)"
		fullPath := filepath.Join(tmpDir, "images.md")
		os.WriteFile(fullPath, []byte(content), 0644)

		req := httptest.NewRequest("GET", "/api/statistics/images.md", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), `"images":2`)
	})
}

func TestStatisticsHandler_GetStatistics_EmptyContent(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Storage: config.StorageConfig{
			NotesDir: tmpDir,
		},
	}
	statsService := services.NewStatisticsService(tmpDir)
	handler := NewStatisticsHandler(statsService, cfg)

	app := fiber.New()
	app.Get("/api/statistics/*", handler.GetStatistics)

	t.Run("handles empty note", func(t *testing.T) {
		content := ""
		fullPath := filepath.Join(tmpDir, "empty.md")
		os.WriteFile(fullPath, []byte(content), 0644)

		req := httptest.NewRequest("GET", "/api/statistics/empty.md", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), `"success":true`)
		assert.Contains(t, string(body), `"words":0`)
	})

	t.Run("handles note with only frontmatter", func(t *testing.T) {
		content := `---
title: Empty
---`
		fullPath := filepath.Join(tmpDir, "frontmatter-only.md")
		os.WriteFile(fullPath, []byte(content), 0644)

		req := httptest.NewRequest("GET", "/api/statistics/frontmatter-only.md", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), `"success":true`)
	})
}

func TestStatisticsHandler_Integration(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Storage: config.StorageConfig{
			NotesDir: tmpDir,
		},
	}
	statsService := services.NewStatisticsService(tmpDir)
	handler := NewStatisticsHandler(statsService, cfg)

	app := fiber.New()
	app.Get("/api/statistics/*", handler.GetStatistics)

	// Create a realistic note
	scenario := []struct {
		path    string
		content string
	}{
		{"programming/golang.md", `---
title: Go Programming
tags: [programming, go, backend]
---

# Go Programming Language

Go is a statically typed, compiled programming language.

## Features

- Fast compilation
- Concurrent execution
- Strong typing

## Code Example

` + "```go" + `
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
` + "```" + `

## Tasks

- [ ] Learn Go basics
- [x] Write first program
- [ ] Build production app

## Links

- [Official Docs](https://golang.org)
- [[Python Comparison]]
`},
	}

	for _, s := range scenario {
		fullPath := filepath.Join(tmpDir, s.path)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		assert.NoError(t, err)
		err = os.WriteFile(fullPath, []byte(s.content), 0644)
		assert.NoError(t, err)
	}

	t.Run("returns comprehensive statistics", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/statistics/programming/golang.md", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), `"success":true`)
		assert.Contains(t, string(body), `"data"`)
		// Check various statistics
		assert.Contains(t, string(body), `"words"`)
		assert.Contains(t, string(body), `"sentences"`)
		assert.Contains(t, string(body), `"headings"`)
		assert.Contains(t, string(body), `"tasks"`)
		assert.Contains(t, string(body), `"code_blocks":1`)
		assert.Contains(t, string(body), `"wikilinks":1`)
	})
}
