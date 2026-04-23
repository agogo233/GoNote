package services

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPrefixSearch(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 创建包含 "register" 的测试笔记
	testNote := filepath.Join(tempDir, "test-register.md")
	content := `# Register Test

This is a register test. The register function is important.
Registration is required for all users.
`
	if err := os.WriteFile(testNote, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// 创建服务
	noteService := NewNoteService(tempDir)
	searchIndex := NewSearchIndex(tempDir, noteService)

	// 构建索引
	if err := searchIndex.BuildIndex(); err != nil {
		t.Fatalf("BuildIndex failed: %v", err)
	}

	// 打印索引中的所有词
	terms := searchIndex.GetIndexedTerms()
	t.Log("=== Indexed Terms ===")
	for _, term := range terms {
		t.Logf("  %s", term)
	}
	t.Logf("Total: %d terms", len(terms))

	// 测试搜索
	tests := []struct {
		query string
		desc  string
	}{
		{"reg", "prefix search for 'reg'"},
		{"register", "exact search for 'register'"},
		{"regist", "prefix search for 'regist'"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			results, err := searchIndex.Search(tt.query)
			if err != nil {
				t.Fatalf("Search failed: %v", err)
			}
			t.Logf("Query '%s' returned %d results", tt.query, len(results))
			for _, r := range results {
				t.Logf("  - %s: %s", r.Path, r.Name)
			}
			
			if len(results) == 0 {
				t.Errorf("Expected at least 1 result for query '%s', got 0", tt.query)
			}
		})
	}
}
