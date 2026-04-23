package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSearchService(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewSearchService(tmpDir)

	assert.NotNil(t, svc)
	assert.Equal(t, tmpDir, svc.notesDir)
}

func TestSearchEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewSearchService(tmpDir)

	results, err := svc.Search("test")

	assert.NoError(t, err)
	assert.Empty(t, results)
}

func TestSearchNoMatch(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewSearchService(tmpDir)

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note.md"), []byte("Hello World"), 0644))

	results, err := svc.Search("nonexistent")

	assert.NoError(t, err)
	assert.Empty(t, results)
}

func TestSearchSingleMatch(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewSearchService(tmpDir)

	content := "# My Note\n\nThis is a test note with some content."
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note.md"), []byte(content), 0644))

	results, err := svc.Search("test")

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "My Note", results[0].Name)
	assert.Equal(t, "note.md", results[0].Path)
	assert.Len(t, results[0].Matches, 1)
	assert.Contains(t, results[0].Matches[0].Context, "test")
}

func TestSearchMultipleMatches(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewSearchService(tmpDir)

	content := "test test test test test"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note.md"), []byte(content), 0644))

	results, err := svc.Search("test")

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	// Should limit to 3 matches per file
	assert.LessOrEqual(t, len(results[0].Matches), 3)
}

func TestSearchMultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewSearchService(tmpDir)

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note1.md"), []byte("# Note One\n\nkeyword found here"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note2.md"), []byte("# Note Two\n\nno match here"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note3.md"), []byte("# Note Three\n\nalso has keyword"), 0644))

	results, err := svc.Search("keyword")

	assert.NoError(t, err)
	assert.Len(t, results, 2)

	names := make(map[string]bool)
	for _, r := range results {
		names[r.Name] = true
	}
	assert.True(t, names["Note One"])
	assert.True(t, names["Note Three"])
}

func TestSearchCaseInsensitive(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewSearchService(tmpDir)

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note.md"), []byte("Hello WORLD"), 0644))

	results, err := svc.Search("world")

	assert.NoError(t, err)
	assert.Len(t, results, 1)

	results2, err := svc.Search("HELLO")

	assert.NoError(t, err)
	assert.Len(t, results2, 1)
}

func TestSearchSpecialCharacters(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewSearchService(tmpDir)

	content := "function test() { return (a + b) * c; }"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "code.md"), []byte(content), 0644))

	results, err := svc.Search("(a + b)")

	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestSearchMatchContext(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewSearchService(tmpDir)

	content := "prefix text SEARCHTERM suffix text"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note.md"), []byte(content), 0644))

	results, err := svc.Search("SEARCHTERM")

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Contains(t, results[0].Matches[0].Context, "prefix text")
	assert.Contains(t, results[0].Matches[0].Context, "suffix text")
}

func TestSearchHighlight(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewSearchService(tmpDir)

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note.md"), []byte("find the keyword here"), 0644))

	results, err := svc.Search("keyword")

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Contains(t, results[0].Matches[0].Context, "<mark")
	assert.Contains(t, results[0].Matches[0].Context, "</mark>")
}

func TestSearchLineNumber(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewSearchService(tmpDir)

	content := "line 1\nline 2\nline 3 has the keyword\nline 4"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note.md"), []byte(content), 0644))

	results, err := svc.Search("keyword")

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, 3, results[0].Matches[0].LineNumber)
}

func TestSearchInSubdirectory(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewSearchService(tmpDir)

	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "subdir", "note.md"), []byte("search target"), 0644))

	results, err := svc.Search("target")

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "subdir/note.md", results[0].Path)
	assert.Equal(t, "subdir", results[0].Folder)
}

func TestSearchMultilineContent(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewSearchService(tmpDir)

	content := "# Title\n\nParagraph one\nwith keyword\nin the middle.\n\nParagraph two."
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note.md"), []byte(content), 0644))

	results, err := svc.Search("keyword")

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	// Newlines in context should be replaced with spaces
	assert.NotContains(t, results[0].Matches[0].Context, "\n")
}

func TestSearchHTMLEscaping(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewSearchService(tmpDir)

	content := "Find <script>alert('xss')</script> here"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note.md"), []byte(content), 0644))

	results, err := svc.Search("script")

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	// HTML should be escaped in the context - the < and > become &lt; and &gt;
	assert.NotContains(t, results[0].Matches[0].Context, "<script>")
	// The angle brackets are escaped
	assert.Contains(t, results[0].Matches[0].Context, "&lt;")
	assert.Contains(t, results[0].Matches[0].Context, "&gt;")
}

func TestSearchEllipsis(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewSearchService(tmpDir)

	// Content at the start - no leading ellipsis for short content
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "start.md"), []byte("keyword at start of content"), 0644))

	// Content at the end - no trailing ellipsis for short content
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "end.md"), []byte("content with keyword"), 0644))

	// Long content where keyword is in the middle - should have ellipsis
	longContent := "# Middle Note\n\nthis is a very long prefix text that should be truncated keyword and this is suffix text that should also be truncated"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "middle.md"), []byte(longContent), 0644))

	// Test that we find results
	results, err := svc.Search("keyword")
	assert.NoError(t, err)
	assert.Len(t, results, 3)

	// Find the result from middle.md
	var foundMiddle bool
	for _, r := range results {
		if r.Name == "Middle Note" {
			foundMiddle = true
			// For the long content, there should be ellipsis
			assert.Contains(t, r.Matches[0].Context, "...")
		}
	}
	assert.True(t, foundMiddle, "Should find middle.md result")
}

func TestReadFileContent(t *testing.T) {
	tmpDir := t.TempDir()

	content := "test file content"
	path := filepath.Join(tmpDir, "test.txt")
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	result, err := readFileContent(path)

	assert.NoError(t, err)
	assert.Equal(t, content, result)
}

func TestReadFileContentNotFound(t *testing.T) {
	_, err := readFileContent("/nonexistent/path/file.txt")

	assert.Error(t, err)
}
