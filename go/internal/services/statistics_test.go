package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatisticsService_CalculateStatistics_EmptyContent(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewStatisticsService(tmpDir)

	stats := svc.CalculateStatisticsFromContent("")

	assert.Equal(t, 0, stats.Words)
	assert.Equal(t, 0, stats.Sentences)
	assert.Equal(t, 0, stats.Characters)
	assert.Equal(t, 0, stats.TotalCharacters)
	assert.Equal(t, 0, stats.ReadingTimeMinutes)
}

func TestStatisticsService_CharacterCounts(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewStatisticsService(tmpDir)

	content := "Hello World"
	stats := svc.CalculateStatisticsFromContent(content)

	assert.Equal(t, 11, stats.TotalCharacters)
	assert.Equal(t, 10, stats.Characters)
	assert.Equal(t, 2, stats.Words)
}

func TestStatisticsService_WordCount(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewStatisticsService(tmpDir)

	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{"single word", "Hello", 1},
		{"multiple words", "Hello World Test", 3},
		{"words with punctuation", "Hello, World! Test.", 3},
		{"empty string", "", 0},
		{"only whitespace", "   \n\t  ", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := svc.CalculateStatisticsFromContent(tt.content)
			assert.Equal(t, tt.expected, stats.Words)
		})
	}
}

func TestStatisticsService_SentenceCount(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewStatisticsService(tmpDir)

	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{"single sentence", "Hello world.", 1},
		{"multiple sentences", "Hello world. How are you? I'm fine!", 3},
		{"sentences with exclamation", "Wow! Amazing!", 2},
		{"no punctuation", "Hello world", 0},
		{"empty string", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := svc.CalculateStatisticsFromContent(tt.content)
			assert.Equal(t, tt.expected, stats.Sentences)
		})
	}
}

func TestStatisticsService_LineCount(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewStatisticsService(tmpDir)

	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{"single line", "Hello", 1},
		{"two lines", "Hello\nWorld", 2},
		{"three lines", "Hello\nWorld\nTest", 3},
		{"empty string", "", 1},
		{"trailing newline", "Hello\n", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := svc.CalculateStatisticsFromContent(tt.content)
			assert.Equal(t, tt.expected, stats.Lines)
		})
	}
}

func TestStatisticsService_ParagraphCount(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewStatisticsService(tmpDir)

	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{"single paragraph", "Hello world", 1},
		{"two paragraphs", "Hello\n\nWorld", 2},
		{"three paragraphs", "First\n\nSecond\n\nThird", 3},
		{"empty string", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := svc.CalculateStatisticsFromContent(tt.content)
			assert.Equal(t, tt.expected, stats.Paragraphs)
		})
	}
}

func TestStatisticsService_ReadingTime(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewStatisticsService(tmpDir)

	t.Run("zero for empty content", func(t *testing.T) {
		stats := svc.CalculateStatisticsFromContent("")
		assert.Equal(t, 0, stats.ReadingTimeMinutes)
	})

	t.Run("minimum 1 minute for short content", func(t *testing.T) {
		stats := svc.CalculateStatisticsFromContent("Hello world")
		assert.Equal(t, 1, stats.ReadingTimeMinutes)
	})

	t.Run("calculates for longer content", func(t *testing.T) {
		words := ""
		for i := 0; i < 500; i++ {
			words += "word "
		}
		stats := svc.CalculateStatisticsFromContent(words)
		assert.GreaterOrEqual(t, stats.ReadingTimeMinutes, 2)
	})
}

func TestStatisticsService_ListItems(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewStatisticsService(tmpDir)

	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{"unordered list with dash", "- item1\n- item2\n- item3", 3},
		{"unordered list with asterisk", "* item1\n* item2", 2},
		{"ordered list", "1. item1\n2. item2\n3. item3", 3},
		{"task list not counted", "- [ ] task1\n- [x] task2", 0},
		{"empty string", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := svc.CalculateStatisticsFromContent(tt.content)
			assert.Equal(t, tt.expected, stats.ListItems)
		})
	}
}

func TestStatisticsService_Tables(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewStatisticsService(tmpDir)

	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{"no table", "Just text", 0},
		{"empty string", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := svc.CalculateStatisticsFromContent(tt.content)
			assert.Equal(t, tt.expected, stats.Tables)
		})
	}
}

func TestStatisticsService_Links(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewStatisticsService(tmpDir)

	content := `
# Links Test

Standard markdown: [Google](https://google.com)
Internal link: [Note](other-note.md)
Wikilink: [[another-note]]
Wikilink with text: [[yet-another|Display Text]]
`
	stats := svc.CalculateStatisticsFromContent(content)

	assert.Equal(t, 2, stats.Links)
	assert.Equal(t, 2, stats.Wikilinks)
	assert.GreaterOrEqual(t, stats.InternalLinks, 3)
}

func TestStatisticsService_Wikilinks(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewStatisticsService(tmpDir)

	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{"simple wikilink", "[[note]]", 1},
		{"wikilink with display text", "[[note|Display]]", 1},
		{"multiple wikilinks", "[[note1]] and [[note2]]", 2},
		{"no wikilinks", "Just text", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := svc.CalculateStatisticsFromContent(tt.content)
			assert.Equal(t, tt.expected, stats.Wikilinks)
		})
	}
}

func TestStatisticsService_CodeBlocks(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewStatisticsService(tmpDir)

	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{"single code block", "```\ncode\n```", 1},
		{"code block with language", "```go\ncode\n```", 1},
		{"multiple code blocks", "```\ncode1\n```\n\n```\ncode2\n```", 2},
		{"no code", "Just text", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := svc.CalculateStatisticsFromContent(tt.content)
			assert.Equal(t, tt.expected, stats.CodeBlocks)
		})
	}
}

func TestStatisticsService_InlineCode(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewStatisticsService(tmpDir)

	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{"single inline", "`code`", 1},
		{"multiple inline", "`code1` and `code2`", 2},
		{"no inline", "Just text", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := svc.CalculateStatisticsFromContent(tt.content)
			assert.Equal(t, tt.expected, stats.InlineCode)
		})
	}
}

func TestStatisticsService_Headings(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewStatisticsService(tmpDir)

	content := "# Heading 1\n## Heading 2\n### Heading 3\n# Another H1\n## Another H2\n"
	stats := svc.CalculateStatisticsFromContent(content)

	assert.GreaterOrEqual(t, stats.Headings.H1, 1)
	assert.GreaterOrEqual(t, stats.Headings.H2, 1)
	assert.GreaterOrEqual(t, stats.Headings.H3, 1)
}

func TestStatisticsService_Tasks(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewStatisticsService(tmpDir)

	content := `
# Task List

- [ ] Pending task 1
- [ ] Pending task 2
- [x] Completed task 1
`
	stats := svc.CalculateStatisticsFromContent(content)

	assert.GreaterOrEqual(t, stats.Tasks.Total, 2)
	assert.GreaterOrEqual(t, stats.Tasks.Completed, 1)
	assert.GreaterOrEqual(t, stats.Tasks.Pending, 1)
}

func TestStatisticsService_Images(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewStatisticsService(tmpDir)

	content := `
# Images

![Alt text](image.png)
![Another](photo.jpg)
Regular link: [not an image](file.txt)
`
	stats := svc.CalculateStatisticsFromContent(content)

	assert.Equal(t, 2, stats.Images)
}

func TestStatisticsService_Blockquotes(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewStatisticsService(tmpDir)

	content := "> This is a quote\n"
	stats := svc.CalculateStatisticsFromContent(content)

	assert.GreaterOrEqual(t, stats.Blockquotes, 1)
}

func TestStatisticsService_CalculateStatistics_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewStatisticsService(tmpDir)

	_, err := svc.CalculateStatistics("nonexistent.md")
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestStatisticsService_CalculateStatistics_FromFile(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewStatisticsService(tmpDir)

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
	filePath := filepath.Join(tmpDir, "test.md")
	err := os.WriteFile(filePath, []byte(content), 0644)
	assert.NoError(t, err)

	stats, err := svc.CalculateStatistics("test.md")
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Greater(t, stats.Words, 0)
}

func TestStatisticsService_UnicodeContent(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewStatisticsService(tmpDir)

	content := "你好世界\nこんにちは\n안녕하세요"
	stats := svc.CalculateStatisticsFromContent(content)

	assert.Greater(t, stats.Words, 0)
	assert.Equal(t, 3, stats.Lines)
}

func TestStatisticsService_StatisticsStruct(t *testing.T) {
	stats := &NoteStatistics{
		Words:            100,
		Sentences:        10,
		Characters:       500,
		TotalCharacters:  600,
		ReadingTimeMinutes: 1,
		Lines:            20,
		Paragraphs:       5,
		ListItems:        8,
		Tables:           2,
		Links:            15,
		InternalLinks:    10,
		ExternalLinks:    5,
		Wikilinks:        3,
		CodeBlocks:       4,
		InlineCode:       6,
		Headings: HeadingCounts{H1: 2, H2: 5, H3: 3},
		Tasks: TaskCounts{Total: 10, Completed: 6, Pending: 4},
		Images:       7,
		Blockquotes:  3,
	}

	assert.Equal(t, 100, stats.Words)
	assert.Equal(t, 10, stats.Sentences)
	assert.Equal(t, 2, stats.Headings.H1)
	assert.Equal(t, 6, stats.Tasks.Completed)
	assert.Equal(t, 4, stats.Tasks.Pending)
}
