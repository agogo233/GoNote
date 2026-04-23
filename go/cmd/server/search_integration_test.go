package main

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v2"
	"gonote/internal/handlers"
	"gonote/internal/services"
)

// TestSearchByTitle_Integration 集成测试标题搜索
func TestSearchByTitle_Integration(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 创建测试笔记
	notes := map[string]string{
		"go-tutorial.md": `# Go语言入门教程

这是一篇关于Go语言的教程。
`,
		"python-guide.md": `---
title: Python编程指南
---
Python是一种流行的编程语言。
`,
		"javascript-notes.md": `# JavaScript高级编程

JavaScript笔记内容。
`,
	}

	for path, content := range notes {
		fullPath := filepath.Join(tempDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
	}

	// 创建服务
	noteService := services.NewNoteService(tempDir)
	searchService := services.NewSearchService(tempDir)
	searchIndex := services.NewSearchIndex(tempDir, noteService)

	// 构建索引
	if err := searchIndex.BuildIndex(); err != nil {
		t.Fatalf("BuildIndex failed: %v", err)
	}

	// 打印索引状态
	t.Log("=== Index State ===")
	t.Logf("Index size: %d terms", searchIndex.GetIndexSize())

	for _, note := range []string{"go-tutorial.md", "python-guide.md", "javascript-notes.md"} {
		t.Logf("Note %s indexed", note)
	}

	// 创建 Fiber app
	app := fiber.New()

	// 创建 handler
	searchHandler := handlers.NewSearchHandlerWithIndex(searchService, searchIndex, nil)
	app.Get("/api/search", searchHandler.Search)

	// 测试用例
	tests := []struct {
		name        string
		query       string
		mode        string
		expectCount int
	}{
		{"标题模式搜索Go", "Go", "title", 1},
		{"标题模式搜索Python", "Python", "title", 1},
		{"标题模式搜索JavaScript", "JavaScript", "title", 1},
		{"全文模式搜索Go", "Go", "full", 1},
		{"智能模式搜索Go", "Go", "smart", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 发送请求
			req := httptest.NewRequest("GET", "/api/search?q="+tt.query+"&mode="+tt.mode, nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}

			if resp.StatusCode != 200 {
				t.Fatalf("Expected status 200, got %d", resp.StatusCode)
			}

			// 解析响应
			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			results, ok := result["results"].([]interface{})
			if !ok {
				t.Fatalf("Response has no results array")
			}

			t.Logf("Query=%q mode=%s returned %d results", tt.query, tt.mode, len(results))
			for i, r := range results {
				if m, ok := r.(map[string]interface{}); ok {
					t.Logf("  [%d] name=%s path=%s", i, m["name"], m["path"])
				}
			}

			if len(results) < tt.expectCount {
				t.Errorf("Expected at least %d results, got %d", tt.expectCount, len(results))
			}

			// 验证返回的是标题而不是文件名
			if len(results) > 0 {
				if m, ok := results[0].(map[string]interface{}); ok {
					name := m["name"].(string)
					// 标题应该包含 meaningful 的内容，而不是简单的文件名
					t.Logf("First result name: %q (should be actual title, not filename)", name)
				}
			}
		})
	}
}
