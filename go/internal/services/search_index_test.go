package services

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSearchIndex_NewSearchIndex(t *testing.T) {
	tempDir := t.TempDir()
	
	ns := NewNoteService(tempDir)
	si := NewSearchIndex(tempDir, ns)
	
	if si == nil {
		t.Fatal("NewSearchIndex returned nil")
	}
	if si.notesDir != tempDir {
		t.Errorf("Expected notesDir %s, got %s", tempDir, si.notesDir)
	}
	if si.index == nil {
		t.Error("Index map should not be nil")
	}
}

func TestSearchIndex_BuildIndex(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create test notes
	notes := map[string]string{
		"test1.md": "# Test Note One\nThis is a test note about golang.",
		"test2.md": "# Another Note\nThis note discusses programming.",
		"subdir/test3.md": "# Subdir Note\nGolang programming is fun.",
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
	
	ns := NewNoteService(tempDir)
	si := NewSearchIndex(tempDir, ns)
	err := si.BuildIndex()
	if err != nil {
		t.Fatalf("BuildIndex failed: %v", err)
	}
	
	// Verify index was built
	if si.GetIndexSize() == 0 {
		t.Error("Index should not be empty after building")
	}
}

func TestSearchIndex_Search(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create test notes
	notes := map[string]string{
		"golang.md": "# Golang Guide\nGolang is a programming language.",
		"python.md": "# Python Guide\nPython is another programming language.",
		"other.md":  "# Other\nThis note has no relevant content.",
	}
	
	for path, content := range notes {
		fullPath := filepath.Join(tempDir, path)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
	}
	
	ns := NewNoteService(tempDir)
	si := NewSearchIndex(tempDir, ns)
	if err := si.BuildIndex(); err != nil {
		t.Fatalf("BuildIndex failed: %v", err)
	}
	
	// Search for "golang"
	results, err := si.Search("golang")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	
	if len(results) == 0 {
		t.Error("Expected to find results for 'golang'")
	}
	
	// Search for "programming"
	results, err = si.Search("programming")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	
	if len(results) < 2 {
		t.Errorf("Expected at least 2 results for 'programming', got %d", len(results))
	}
	
	// Search for non-existent term
	results, err = si.Search("nonexistent")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	
	if len(results) != 0 {
		t.Errorf("Expected 0 results for 'nonexistent', got %d", len(results))
	}
}

func TestSearchIndex_EmptyQuery(t *testing.T) {
	tempDir := t.TempDir()
	ns := NewNoteService(tempDir)
	si := NewSearchIndex(tempDir, ns)
	
	results, err := si.Search("")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	
	if len(results) != 0 {
		t.Errorf("Expected 0 results for empty query, got %d", len(results))
	}
}

func TestSearchIndex_UpdateIndex(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create initial note
	notePath := filepath.Join(tempDir, "test.md")
	if err := os.WriteFile(notePath, []byte("Original content"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
	
	ns := NewNoteService(tempDir)
	si := NewSearchIndex(tempDir, ns)
	if err := si.BuildIndex(); err != nil {
		t.Fatalf("BuildIndex failed: %v", err)
	}
	
	// Verify original content is indexed
	results, _ := si.Search("original")
	if len(results) == 0 {
		t.Error("Expected to find 'original'")
	}
	
	// Update the note
	if err := os.WriteFile(notePath, []byte("Updated content with newword"), 0644); err != nil {
		t.Fatalf("Failed to update file: %v", err)
	}
	
	// Update index
	if err := si.UpdateIndex("test.md"); err != nil {
		t.Fatalf("UpdateIndex failed: %v", err)
	}
	
	// Verify new content is indexed
	results, _ = si.Search("newword")
	if len(results) == 0 {
		t.Error("Expected to find 'newword' after update")
	}
}

func TestSearchIndex_RemoveFromIndex(t *testing.T) {
	tempDir := t.TempDir()

	notePath := filepath.Join(tempDir, "test.md")
	if err := os.WriteFile(notePath, []byte("Unique content word"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	ns := NewNoteService(tempDir)
	si := NewSearchIndex(tempDir, ns)
	if err := si.BuildIndex(); err != nil {
		t.Fatalf("BuildIndex failed: %v", err)
	}

	results, _ := si.Search("unique")
	if len(results) == 0 {
		t.Error("Expected to find 'unique'")
	}

	si.RemoveFromIndex("test.md")
	os.Remove(notePath)

	ns.InvalidateCache()
	si.noteService = NewNoteService(tempDir)

	results, _ = si.Search("unique")
	if len(results) != 0 {
		t.Error("Expected no results after removal")
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		input string
		expected int
	}{
		{"Hello World", 2},
		{"This is a test sentence.", 4},
		{"Golang Programming Language", 3},
		{"Multiple spaces between words", 4},
		{"UPPERCASE WORDS", 2},
		{"a b c d e", 0},
		{"ab cd ef", 3},
	}

	for _, tt := range tests {
		terms := tokenize(tt.input)
		if len(terms) < tt.expected {
			t.Errorf("tokenize(%q) returned %d terms, expected at least %d", tt.input, len(terms), tt.expected)
		}
	}
}

func TestTokenize_Lowercase(t *testing.T) {
	terms := tokenize("HELLO World")
	for _, term := range terms {
		if term != "hello" && term != "world" {
			t.Errorf("Expected lowercase terms, got %q", term)
		}
	}
}

func TestTokenize_NoDuplicates(t *testing.T) {
	terms := tokenize("test test test testing")
	
	// Count occurrences
	counts := make(map[string]int)
	for _, term := range terms {
		counts[term]++
	}
	
	for term, count := range counts {
		if count > 1 {
			t.Errorf("Duplicate term %q found %d times", term, count)
		}
	}
}

func TestSearchIndex_GetIndexSize(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create note with known content
	if err := os.WriteFile(filepath.Join(tempDir, "test.md"), []byte("hello world"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
	
	ns := NewNoteService(tempDir)
	si := NewSearchIndex(tempDir, ns)
	
	// Before building
	if si.GetIndexSize() != 0 {
		t.Error("Index should be empty before building")
	}
	
	// After building
	si.BuildIndex()
	if si.GetIndexSize() < 2 {
		t.Errorf("Expected at least 2 terms (hello, world), got %d", si.GetIndexSize())
	}
}

func TestSearchIndex_GetIndexedTerms(t *testing.T) {
	tempDir := t.TempDir()

	// Create note
	if err := os.WriteFile(filepath.Join(tempDir, "test.md"), []byte("alpha beta gamma"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	ns := NewNoteService(tempDir)
	si := NewSearchIndex(tempDir, ns)
	si.BuildIndex()

	terms := si.GetIndexedTerms()

	// Check that expected terms are present
	termMap := make(map[string]bool)
	for _, term := range terms {
		termMap[term] = true
	}

	expectedTerms := []string{"alpha", "beta", "gamma"}
	for _, expected := range expectedTerms {
		if !termMap[expected] {
			t.Errorf("Expected term %q not found in indexed terms", expected)
		}
	}
}

func TestSearchIndex_SearchByTitle(t *testing.T) {
	tempDir := t.TempDir()

	// Create test notes with different titles
	notes := map[string]string{
		"golang-guide.md":     "# Golang Programming Guide\nThis is about Go language.",
		"python-tutorial.md":  "# Python Tutorial for Beginners\nLearn Python basics.",
		"database-notes.md":   "# Database Design Patterns\nSQL and NoSQL databases.",
		"subdir/algorithms.md": "# Algorithm Analysis\nSorting and searching algorithms.",
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

	ns := NewNoteService(tempDir)
	si := NewSearchIndex(tempDir, ns)
	if err := si.BuildIndex(); err != nil {
		t.Fatalf("BuildIndex failed: %v", err)
	}

	// Test 1: Search by exact title word
	results, err := si.SearchByTitle("Golang")
	if err != nil {
		t.Fatalf("SearchByTitle failed: %v", err)
	}
	if len(results) == 0 {
		t.Error("Expected to find results for 'Golang'")
	} else if results[0].Name != "Golang Programming Guide" {
		t.Errorf("Expected title 'Golang Programming Guide', got '%s'", results[0].Name)
	}

	// Test 2: Search by partial title (prefix match)
	results, err = si.SearchByTitle("Python")
	if err != nil {
		t.Fatalf("SearchByTitle failed: %v", err)
	}
	if len(results) == 0 {
		t.Error("Expected to find results for 'Python'")
	}

	// Test 3: Search with multiple words
	results, err = si.SearchByTitle("Database Design")
	if err != nil {
		t.Fatalf("SearchByTitle failed: %v", err)
	}
	if len(results) == 0 {
		t.Error("Expected to find results for 'Database Design'")
	}

	// Test 4: Case-insensitive search
	results, err = si.SearchByTitle("golang")
	if err != nil {
		t.Fatalf("SearchByTitle failed: %v", err)
	}
	if len(results) == 0 {
		t.Error("Expected case-insensitive search to work")
	}

	// Test 5: Search for non-existent title
	results, err = si.SearchByTitle("NonExistent")
	if err != nil {
		t.Fatalf("SearchByTitle failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results for non-existent title, got %d", len(results))
	}
}

func TestSearchIndex_SearchByTitle_EmptyQuery(t *testing.T) {
	tempDir := t.TempDir()
	ns := NewNoteService(tempDir)
	si := NewSearchIndex(tempDir, ns)

	results, err := si.SearchByTitle("")
	if err != nil {
		t.Fatalf("SearchByTitle failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results for empty query, got %d", len(results))
	}
}

func TestSearchIndex_SearchByTitle_FallbackToDisk(t *testing.T) {
	tempDir := t.TempDir()

	// Create a note
	notePath := filepath.Join(tempDir, "test-note.md")
	content := "# My Special Title\nThis is some content."
	if err := os.WriteFile(notePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	ns := NewNoteService(tempDir)
	si := NewSearchIndex(tempDir, ns)
	
	// Don't build index - test disk fallback
	// Build empty index
	si.BuildIndex()

	// Search should still work via disk fallback
	results, err := si.SearchByTitle("Special")
	if err != nil {
		t.Fatalf("SearchByTitle failed: %v", err)
	}
	if len(results) == 0 {
		t.Error("Expected disk fallback to find results")
	}
}

func TestSearchIndex_SearchSmart(t *testing.T) {
	tempDir := t.TempDir()

	notes := map[string]string{
		"title-match.md": "# Unique Title Word\nThis content has different words.",
		"content-match.md": "# Other Note\nThis has unique content keyword here.",
	}

	for path, content := range notes {
		fullPath := filepath.Join(tempDir, path)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
	}

	ns := NewNoteService(tempDir)
	si := NewSearchIndex(tempDir, ns)
	if err := si.BuildIndex(); err != nil {
		t.Fatalf("BuildIndex failed: %v", err)
	}

	results, err := si.SearchSmart("Title")
	if err != nil {
		t.Fatalf("SearchSmart failed: %v", err)
	}
	if len(results) == 0 {
		t.Error("Expected SearchSmart to find title matches")
	}

	results, err = si.SearchSmart("content keyword")
	if err != nil {
		t.Fatalf("SearchSmart failed: %v", err)
	}
	if len(results) == 0 {
		t.Error("Expected SearchSmart to find content matches")
	}
}

func TestSearchIndex_SearchCJK(t *testing.T) {
	tempDir := t.TempDir()

	notes := map[string]string{
		"shanghai.md": "# 上海市旅游攻略\n上海是中国的经济中心，有很多旅游景点。",
		"beijing.md": "# 北京旅游指南\n北京是中国的首都，有故宫和长城。",
		"other.md": "# 其他笔记\n这段内容没有相关关键词。",
	}

	for path, content := range notes {
		fullPath := filepath.Join(tempDir, path)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
	}

	ns := NewNoteService(tempDir)
	si := NewSearchIndex(tempDir, ns)
	if err := si.BuildIndex(); err != nil {
		t.Fatalf("BuildIndex failed: %v", err)
	}

	results, err := si.Search("上海")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) == 0 {
		t.Error("Expected to find results for CJK query '上海'")
	}

	results, err = si.Search("旅游")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) < 2 {
		t.Errorf("Expected at least 2 results for CJK query '旅游', got %d", len(results))
	}

	results, err = si.Search("不存在的词")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results for non-existent CJK query, got %d", len(results))
	}
}

func TestSearchIndex_SearchCJK_PrefixMatch(t *testing.T) {
	tempDir := t.TempDir()

	content := "# 笔记标题\n上海市旅游攻略非常实用，推荐给大家。"
	if err := os.WriteFile(filepath.Join(tempDir, "cjk-note.md"), []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	ns := NewNoteService(tempDir)
	si := NewSearchIndex(tempDir, ns)
	if err := si.BuildIndex(); err != nil {
		t.Fatalf("BuildIndex failed: %v", err)
	}

	results, err := si.Search("上海市")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) == 0 {
		t.Error("Expected prefix match for '上海市' to find note with '上海市旅游攻略'")
	}
}

func TestSearchIndex_MoveUpdatesIndex(t *testing.T) {
	tempDir := t.TempDir()

	oldPath := filepath.Join(tempDir, "old-note.md")
	if err := os.WriteFile(oldPath, []byte("# My Note\nUnique content about golang."), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	ns := NewNoteService(tempDir)
	si := NewSearchIndex(tempDir, ns)
	if err := si.BuildIndex(); err != nil {
		t.Fatalf("BuildIndex failed: %v", err)
	}

	results, _ := si.Search("unique")
	if len(results) == 0 {
		t.Error("Expected to find 'unique' in old note")
	}

	si.RemoveFromIndex("old-note.md")

	newPath := filepath.Join(tempDir, "new-note.md")
	if err := os.WriteFile(newPath, []byte("# Moved Note\nUnique content about golang."), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	if err := si.UpdateIndex("new-note.md"); err != nil {
		t.Fatalf("UpdateIndex failed: %v", err)
	}

	results, _ = si.Search("unique")
	if len(results) == 0 {
		t.Error("Expected to find 'unique' in new note after move")
	}
}

func TestSearchIndex_RemoveFromIndex_CleansTitleIndex(t *testing.T) {
	tempDir := t.TempDir()

	notePath := filepath.Join(tempDir, "test.md")
	if err := os.WriteFile(notePath, []byte("# Unique Title Here\nSome content."), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	ns := NewNoteService(tempDir)
	si := NewSearchIndex(tempDir, ns)
	if err := si.BuildIndex(); err != nil {
		t.Fatalf("BuildIndex failed: %v", err)
	}

	results, _ := si.SearchByTitle("Unique")
	if len(results) == 0 {
		t.Error("Expected to find title 'Unique'")
	}

	si.RemoveFromIndex("test.md")
	os.Remove(notePath)

	ns.InvalidateCache()
	si.noteService = NewNoteService(tempDir)

	results, _ = si.SearchByTitle("Unique")
	if len(results) != 0 {
		t.Error("Expected no title results after removal")
	}
}

func TestSearchIndex_SearchByTitle_CJK(t *testing.T) {
	tempDir := t.TempDir()

	notes := map[string]string{
		"shanghai.md": "# 上海市旅游攻略\n上海是经济中心。",
		"beijing.md": "# 北京旅游指南\n北京是首都。",
	}

	for path, content := range notes {
		fullPath := filepath.Join(tempDir, path)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
	}

	ns := NewNoteService(tempDir)
	si := NewSearchIndex(tempDir, ns)
	if err := si.BuildIndex(); err != nil {
		t.Fatalf("BuildIndex failed: %v", err)
	}

	results, err := si.SearchByTitle("上海")
	if err != nil {
		t.Fatalf("SearchByTitle failed: %v", err)
	}
	if len(results) == 0 {
		t.Error("Expected to find CJK title matching '上海'")
	}

	results, err = si.SearchByTitle("旅游")
	if err != nil {
		t.Fatalf("SearchByTitle failed: %v", err)
	}
	if len(results) < 2 {
		t.Errorf("Expected at least 2 CJK title results for '旅游', got %d", len(results))
	}
}

func TestSearchIndex_SearchSmart_CJK(t *testing.T) {
	tempDir := t.TempDir()

	notes := map[string]string{
		"cjk-title.md": "# 上海美食推荐\n这里介绍上海的特色美食。",
		"cjk-content.md": "# 其他笔记\n上海美食非常有名，值得一试。",
	}

	for path, content := range notes {
		fullPath := filepath.Join(tempDir, path)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
	}

	ns := NewNoteService(tempDir)
	si := NewSearchIndex(tempDir, ns)
	if err := si.BuildIndex(); err != nil {
		t.Fatalf("BuildIndex failed: %v", err)
	}

	results, err := si.SearchSmart("上海美食")
	if err != nil {
		t.Fatalf("SearchSmart failed: %v", err)
	}
	if len(results) == 0 {
		t.Error("Expected SearchSmart to find CJK results")
	}
}
