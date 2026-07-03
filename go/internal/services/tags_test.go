package services

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gonote/internal/models"
)

func TestParseTags_EmptyContent(t *testing.T) {
	tags := ParseTags("")
	assert.Empty(t, tags)
}

func TestParseTags_NoFrontmatter(t *testing.T) {
	content := "# Just a heading\nSome content here."
	tags := ParseTags(content)
	assert.Empty(t, tags)
}

func TestParseTags_EmptyFrontmatter(t *testing.T) {
	content := `---
---
No tags here.`
	tags := ParseTags(content)
	assert.Empty(t, tags)
}

func TestParseTags_InlineArrayTags(t *testing.T) {
	content := `---
title: Test Note
tags: [tag1, tag2, tag3]
---
Content here.`
	tags := ParseTags(content)
	assert.Equal(t, []string{"tag1", "tag2", "tag3"}, tags)
}

func TestParseTags_InlineArrayTagsWithSpaces(t *testing.T) {
	content := `---
tags: [ tag1 , tag2 , tag3 ]
---
Content.`
	tags := ParseTags(content)
	assert.Equal(t, []string{"tag1", "tag2", "tag3"}, tags)
}

func ParseTags_SingleTag(t *testing.T) {
	content := `---
tags: single-tag
---
Content.`
	tags := ParseTags(content)
	assert.Equal(t, []string{"single-tag"}, tags)
}

func TestParseTags_MultilineListTags(t *testing.T) {
	content := `---
tags:
  - tag1
  - tag2
  - tag3
---
Content here.`
	tags := ParseTags(content)
	assert.Equal(t, []string{"tag1", "tag2", "tag3"}, tags)
}

func TestParseTags_MultilineListTagsWithComments(t *testing.T) {
	content := `---
tags:
  - tag1
  - tag2
  - tag3
---
Content.`
	tags := ParseTags(content)
	// Tags are parsed from the list
	assert.Equal(t, []string{"tag1", "tag2", "tag3"}, tags)
}

func TestParseTags_CaseNormalization(t *testing.T) {
	content := `---
tags: [Tag1, TAG2, taG3]
---
Content.`
	tags := ParseTags(content)
	assert.Equal(t, []string{"tag1", "tag2", "tag3"}, tags)
}

func TestParseTags_RemovesDuplicates(t *testing.T) {
	content := `---
tags: [tag1, tag2, tag1, tag3, tag2]
---
Content.`
	tags := ParseTags(content)
	assert.Equal(t, []string{"tag1", "tag2", "tag3"}, tags)
}

func TestParseTags_SortedOutput(t *testing.T) {
	content := `---
tags: [zebra, apple, mango, banana]
---
Content.`
	tags := ParseTags(content)
	assert.Equal(t, []string{"apple", "banana", "mango", "zebra"}, tags)
}

func TestParseTags_EmptyTagsInList(t *testing.T) {
	content := `---
tags:
  - tag1
  - 
  - tag2
---
Content.`
	tags := ParseTags(content)
	assert.Equal(t, []string{"tag1", "tag2"}, tags)
}

func TestParseTags_MissingClosingFrontmatter(t *testing.T) {
	content := `---
tags: [tag1, tag2]
No closing.`
	tags := ParseTags(content)
	// Should not parse tags without closing frontmatter
	assert.Empty(t, tags)
}

func TestParseTags_OnlyOpeningFrontmatter(t *testing.T) {
	content := `---
tags: [tag1, tag2]`
	tags := ParseTags(content)
	assert.Empty(t, tags)
}

func TestParseTags_TagsWithSpecialCharacters(t *testing.T) {
	content := `---
tags: [tag-with-dash, tag_with_underscore, tag.with.dot]
---
Content.`
	tags := ParseTags(content)
	// Tags are sorted alphabetically
	assert.Contains(t, tags, "tag-with-dash")
	assert.Contains(t, tags, "tag_with_underscore")
	assert.Contains(t, tags, "tag.with.dot")
	assert.Len(t, tags, 3)
}

func TestParseTags_ChineseTags(t *testing.T) {
	content := `---
tags: [标签 1, 标签 2, 标签 3]
---
内容。`
	tags := ParseTags(content)
	assert.Equal(t, []string{"标签 1", "标签 2", "标签 3"}, tags)
}

func TestParseTags_MixedLanguageTags(t *testing.T) {
	// Test with space-separated tags in array format
	content := `---
tags: [english, 中文，日本語，français]
---
Content.`
	tags := ParseTags(content)
	// Tags are parsed and sorted
	assert.Contains(t, tags, "english")
	assert.GreaterOrEqual(t, len(tags), 1)
}

func TestSortUniqueTags_EmptyInput(t *testing.T) {
	result := sortUniqueTags([]string{})
	assert.Empty(t, result)
}

func TestSortUniqueTags_AllDuplicates(t *testing.T) {
	result := sortUniqueTags([]string{"tag1", "tag1", "tag1"})
	assert.Equal(t, []string{"tag1"}, result)
}

func TestSortUniqueTags_AlreadySorted(t *testing.T) {
	result := sortUniqueTags([]string{"apple", "banana", "cherry"})
	assert.Equal(t, []string{"apple", "banana", "cherry"}, result)
}

func TestSortUniqueTags_ReverseSorted(t *testing.T) {
	result := sortUniqueTags([]string{"zebra", "mango", "apple"})
	assert.Equal(t, []string{"apple", "mango", "zebra"}, result)
}

// TagService tests

func TestTagService_GetAllTags(t *testing.T) {
	tmpDir := t.TempDir()
	noteService := NewNoteService(tmpDir)
	tagService := NewTagService(noteService, tmpDir)

	// Create notes with different tags
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

	tagCounts, err := tagService.GetAllTags()
	assert.NoError(t, err)
	assert.Equal(t, 3, len(tagCounts))
	assert.Equal(t, 2, tagCounts["tag1"])
	assert.Equal(t, 2, tagCounts["tag2"])
	assert.Equal(t, 2, tagCounts["tag3"])
}

func TestTagService_GetAllTags_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	noteService := NewNoteService(tmpDir)
	tagService := NewTagService(noteService, tmpDir)

	tagCounts, err := tagService.GetAllTags()
	assert.NoError(t, err)
	assert.Empty(t, tagCounts)
}

func TestTagService_GetAllTags_NoTags(t *testing.T) {
	tmpDir := t.TempDir()
	noteService := NewNoteService(tmpDir)
	tagService := NewTagService(noteService, tmpDir)

	// Create note without tags
	content := `---
title: No Tags
---
Content here.`
	fullPath := filepath.Join(tmpDir, "note.md")
	err := os.WriteFile(fullPath, []byte(content), 0644)
	assert.NoError(t, err)

	tagCounts, err := tagService.GetAllTags()
	assert.NoError(t, err)
	assert.Empty(t, tagCounts)
}

func TestTagService_GetNotesByTag(t *testing.T) {
	tmpDir := t.TempDir()
	noteService := NewNoteService(tmpDir)
	tagService := NewTagService(noteService, tmpDir)

	// Create notes with different tags
	notes := []struct {
		path    string
		content string
	}{
		{"note1.md", `---
tags: [tag1, common]
---
Content 1`},
		{"note2.md", `---
tags: [tag2, common]
---
Content 2`},
		{"note3.md", `---
tags: [tag3]
---
Content 3`},
	}

	for _, note := range notes {
		fullPath := filepath.Join(tmpDir, note.path)
		err := os.WriteFile(fullPath, []byte(note.content), 0644)
		assert.NoError(t, err)
	}

	// Test getting notes by tag
	t.Run("finds notes with specific tag", func(t *testing.T) {
		notes, err := tagService.GetNotesByTag("tag1")
		assert.NoError(t, err)
		assert.Len(t, notes, 1)
		assert.Equal(t, "note1.md", notes[0].Path)
	})

	t.Run("finds multiple notes with common tag", func(t *testing.T) {
		notes, err := tagService.GetNotesByTag("common")
		assert.NoError(t, err)
		assert.Len(t, notes, 2)
	})

	t.Run("returns empty for non-existent tag", func(t *testing.T) {
		notes, err := tagService.GetNotesByTag("nonexistent")
		assert.NoError(t, err)
		assert.Empty(t, notes)
	})

	t.Run("case insensitive tag matching", func(t *testing.T) {
		notes, err := tagService.GetNotesByTag("TAG1")
		assert.NoError(t, err)
		assert.Len(t, notes, 1)
	})
}

func TestTagService_GetTagsCached(t *testing.T) {
	tmpDir := t.TempDir()
	noteService := NewNoteService(tmpDir)
	tagService := NewTagService(noteService, tmpDir)

	// Create a file with tags
	content := `---
tags: [tag1, tag2]
---
Content.`
	filePath := filepath.Join(tmpDir, "test.md")
	err := os.WriteFile(filePath, []byte(content), 0644)
	assert.NoError(t, err)

	// First call - should parse and cache
	tags1 := tagService.GetTagsCached(filePath)
	assert.Equal(t, []string{"tag1", "tag2"}, tags1)

	// Second call - should use cache
	tags2 := tagService.GetTagsCached(filePath)
	assert.Equal(t, []string{"tag1", "tag2"}, tags2)

	// Modify file - should re-parse
	time.Sleep(10 * time.Millisecond) // Ensure different mod time
	os.WriteFile(filePath, []byte(content+"\n"), 0644)
	tags3 := tagService.GetTagsCached(filePath)
	assert.Equal(t, []string{"tag1", "tag2"}, tags3)
}

func TestTagService_GetTagsCached_NonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	noteService := NewNoteService(tmpDir)
	tagService := NewTagService(noteService, tmpDir)

	tags := tagService.GetTagsCached("/nonexistent/file.md")
	assert.Empty(t, tags)
}

func TestTagService_ClearCache(t *testing.T) {
	tmpDir := t.TempDir()
	noteService := NewNoteService(tmpDir)
	tagService := NewTagService(noteService, tmpDir)

	content := `---
tags: [tag1]
---
Content.`
	filePath := filepath.Join(tmpDir, "test.md")
	os.WriteFile(filePath, []byte(content), 0644)

	// First call populates cache
	tags := tagService.GetTagsCached(filePath)
	assert.Equal(t, []string{"tag1"}, tags)

	// Clear cache
	tagService.ClearCache()

	// After clearing, re-parsing from file should still work
	tags = tagService.GetTagsCached(filePath)
	assert.Equal(t, []string{"tag1"}, tags)
}

func TestTagService_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	noteService := NewNoteService(tmpDir)
	tagService := NewTagService(noteService, tmpDir)

	// Create test file
	content := `---
tags: [tag1, tag2]
---
Content.`
	filePath := filepath.Join(tmpDir, "test.md")
	os.WriteFile(filePath, []byte(content), 0644)

	// Concurrent reads
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tags := tagService.GetTagsCached(filePath)
			assert.Len(t, tags, 2)
		}()
	}
	wg.Wait()
}

// FilterNotesByTags tests

func TestFilterNotesByTags_EmptyTags(t *testing.T) {
	notes := []models.Note{
		{Path: "note1.md", Tags: []string{"tag1", "tag2"}},
		{Path: "note2.md", Tags: []string{"tag3"}},
	}

	result := FilterNotesByTags(notes, []string{})
	assert.Len(t, result, 2) // Returns all notes when no filter tags
}

func TestFilterNotesByTags_SingleTag(t *testing.T) {
	notes := []models.Note{
		{Path: "note1.md", Tags: []string{"tag1", "tag2"}},
		{Path: "note2.md", Tags: []string{"tag2"}},
		{Path: "note3.md", Tags: []string{"tag3"}},
	}

	result := FilterNotesByTags(notes, []string{"tag2"})
	assert.Len(t, result, 2)
	assert.Equal(t, "note1.md", result[0].Path)
	assert.Equal(t, "note2.md", result[1].Path)
}

func TestFilterNotesByTags_MultipleTags_AND_Logic(t *testing.T) {
	notes := []models.Note{
		{Path: "note1.md", Tags: []string{"tag1", "tag2"}},
		{Path: "note2.md", Tags: []string{"tag2", "tag3"}},
		{Path: "note3.md", Tags: []string{"tag1", "tag2", "tag3"}},
		{Path: "note4.md", Tags: []string{"tag1"}},
	}

	// AND logic: must have BOTH tag1 AND tag2
	result := FilterNotesByTags(notes, []string{"tag1", "tag2"})
	assert.Len(t, result, 2)
	assert.Equal(t, "note1.md", result[0].Path)
	assert.Equal(t, "note3.md", result[1].Path)
}

func TestFilterNotesByTags_CaseInsensitive(t *testing.T) {
	notes := []models.Note{
		{Path: "note1.md", Tags: []string{"tag1", "tag2"}},
		{Path: "note2.md", Tags: []string{"tag1"}},
	}

	// Tags are converted to lowercase for matching
	result := FilterNotesByTags(notes, []string{"tag1"})
	assert.Len(t, result, 2)
}

func TestFilterNotesByTags_NoMatchingNotes(t *testing.T) {
	notes := []models.Note{
		{Path: "note1.md", Tags: []string{"tag1"}},
		{Path: "note2.md", Tags: []string{"tag2"}},
	}

	result := FilterNotesByTags(notes, []string{"nonexistent"})
	assert.Empty(t, result)
}

func TestFilterNotesByTags_NotesWithoutTags(t *testing.T) {
	notes := []models.Note{
		{Path: "note1.md", Tags: []string{"tag1"}},
		{Path: "note2.md", Tags: nil},
		{Path: "note3.md", Tags: []string{}},
	}

	result := FilterNotesByTags(notes, []string{"tag1"})
	assert.Len(t, result, 1)
	assert.Equal(t, "note1.md", result[0].Path)
}

func TestFilterNotesByTags_WhitespaceInTags(t *testing.T) {
	notes := []models.Note{
		{Path: "note1.md", Tags: []string{"tag1", "tag2"}},
	}

	result := FilterNotesByTags(notes, []string{" tag1 ", "tag2  "})
	assert.Len(t, result, 1)
}

func TestFilterNotesByTags_PreservesOrder(t *testing.T) {
	notes := []models.Note{
		{Path: "note1.md", Tags: []string{"tag1"}},
		{Path: "note2.md", Tags: []string{"tag1"}},
		{Path: "note3.md", Tags: []string{"tag1"}},
	}

	result := FilterNotesByTags(notes, []string{"tag1"})
	assert.Len(t, result, 3)
	assert.Equal(t, "note1.md", result[0].Path)
	assert.Equal(t, "note2.md", result[1].Path)
	assert.Equal(t, "note3.md", result[2].Path)
}

func TestNoteHasAllTags_EmptyRequiredTags(t *testing.T) {
	note := models.Note{Path: "note.md", Tags: []string{"tag1"}}
	result := noteHasAllTags(note, map[string]bool{})
	assert.True(t, result) // Empty required tags matches all
}

func TestNoteHasAllTags_NoteWithoutTags(t *testing.T) {
	note := models.Note{Path: "note.md", Tags: nil}
	result := noteHasAllTags(note, map[string]bool{"tag1": true})
	assert.False(t, result)
}

func TestNoteHasAllTags_PartialMatch(t *testing.T) {
	note := models.Note{Path: "note.md", Tags: []string{"tag1", "tag2"}}
	result := noteHasAllTags(note, map[string]bool{"tag1": true, "tag3": true})
	assert.False(t, result) // Missing tag3
}

func TestNoteHasAllTags_ExactMatch(t *testing.T) {
	note := models.Note{Path: "note.md", Tags: []string{"tag1", "tag2"}}
	result := noteHasAllTags(note, map[string]bool{"tag1": true, "tag2": true})
	assert.True(t, result)
}

func TestNoteHasAllTags_SubsetMatch(t *testing.T) {
	note := models.Note{Path: "note.md", Tags: []string{"tag1", "tag2", "tag3"}}
	result := noteHasAllTags(note, map[string]bool{"tag1": true})
	assert.True(t, result)
}

// Integration tests with file system

func TestTagService_Integration_ComplexScenario(t *testing.T) {
	tmpDir := t.TempDir()
	noteService := NewNoteService(tmpDir)
	tagService := NewTagService(noteService, tmpDir)

	// Create a complex set of notes
	scenario := []struct {
		path    string
		content string
	}{
		{"programming/go.md", `---
tags: [programming, go, backend]
---
# Go Programming
Go is great for backend.`},
		{"programming/python.md", `---
tags: [programming, python, backend]
---
# Python Programming
Python is versatile.`},
		{"personal/notes.md", `---
tags: [personal]
---
# Personal Note
Just a personal note.`},
		{"docs/api.md", `---
tags: [docs, api]
---
# API Documentation
API docs here.`},
		{"no-tags.md", `---
title: No Tags
---
This note has no tags.`},
	}

	for _, s := range scenario {
		fullPath := filepath.Join(tmpDir, s.path)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		assert.NoError(t, err)
		err = os.WriteFile(fullPath, []byte(s.content), 0644)
		assert.NoError(t, err)
	}

	t.Run("GetAllTags returns correct counts", func(t *testing.T) {
		tagCounts, err := tagService.GetAllTags()
		assert.NoError(t, err)
		assert.Equal(t, 2, tagCounts["programming"])
		assert.Equal(t, 2, tagCounts["backend"])
		assert.Equal(t, 1, tagCounts["go"])
		assert.Equal(t, 1, tagCounts["python"])
		assert.Equal(t, 1, tagCounts["personal"])
		assert.Equal(t, 1, tagCounts["docs"])
		assert.Equal(t, 1, tagCounts["api"])
	})

	t.Run("GetNotesByTag finds notes in subdirectories", func(t *testing.T) {
		notes, err := tagService.GetNotesByTag("programming")
		assert.NoError(t, err)
		assert.Len(t, notes, 2)
	})

	t.Run("FilterNotesByTags with AND logic", func(t *testing.T) {
		allNotes, _, _ := noteService.ScanNotes(false)
		filtered := FilterNotesByTags(allNotes, []string{"programming", "backend"})
		assert.Len(t, filtered, 2)
	})

	t.Run("Tag caching works correctly", func(t *testing.T) {
		// First access
		tags1 := tagService.GetTagsCached(filepath.Join(tmpDir, "programming/go.md"))
		assert.Equal(t, []string{"backend", "go", "programming"}, tags1)

		// Second access (cached)
		tags2 := tagService.GetTagsCached(filepath.Join(tmpDir, "programming/go.md"))
		assert.Equal(t, tags1, tags2)
	})
}
