package services

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// NoteStatistics contains comprehensive statistics about a note
type NoteStatistics struct {
	Words            int            `json:"words"`              // Word count
	Sentences        int            `json:"sentences"`          // Sentence count
	Characters       int            `json:"characters"`         // Character count (excluding whitespace)
	TotalCharacters  int            `json:"total_characters"`   // Total character count (including whitespace)
	ReadingTimeMinutes int          `json:"reading_time_minutes"` // Estimated reading time in minutes
	Lines            int            `json:"lines"`              // Line count
	Paragraphs       int            `json:"paragraphs"`         // Paragraph count
	ListItems        int            `json:"list_items"`         // List item count
	Tables           int            `json:"tables"`             // Table count
	Links            int            `json:"links"`              // Total link count
	InternalLinks    int            `json:"internal_links"`     // Internal link count
	ExternalLinks    int            `json:"external_links"`     // External link count
	Wikilinks        int            `json:"wikilinks"`          // Wikilink count
	CodeBlocks       int            `json:"code_blocks"`        // Code block count
	InlineCode       int            `json:"inline_code"`        // Inline code count
	Headings         HeadingCounts  `json:"headings"`           // Heading counts by level
	Tasks            TaskCounts     `json:"tasks"`              // Task counts
	Images           int            `json:"images"`             // Image count
	Blockquotes      int            `json:"blockquotes"`        // Blockquote count
}

// HeadingCounts contains counts of headings by level
type HeadingCounts struct {
	H1 int `json:"h1"` // Level 1 headings
	H2 int `json:"h2"` // Level 2 headings
	H3 int `json:"h3"` // Level 3 headings
}

// TaskCounts contains task-related counts
type TaskCounts struct {
	Total     int `json:"total"`      // Total tasks
	Completed int `json:"completed"`  // Completed tasks
	Pending   int `json:"pending"`    // Pending tasks
}

// StatisticsService handles note statistics calculation
type StatisticsService struct {
	notesDir string
}

// NewStatisticsService creates a new StatisticsService
func NewStatisticsService(notesDir string) *StatisticsService {
	return &StatisticsService{notesDir: notesDir}
}

// CalculateStatistics calculates comprehensive statistics for a note
func (s *StatisticsService) CalculateStatistics(notePath string) (*NoteStatistics, error) {
	fullPath := filepath.Join(s.notesDir, notePath)

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}

	return s.CalculateStatisticsFromContent(string(content)), nil
}

// CalculateStatisticsFromContent calculates statistics from note content
func (s *StatisticsService) CalculateStatisticsFromContent(content string) *NoteStatistics {
	stats := &NoteStatistics{}

	// Total character count (including whitespace)
	stats.TotalCharacters = len(content)

	// Character count (excluding whitespace)
	charsNoSpace := strings.ReplaceAll(content, " ", "")
	charsNoSpace = strings.ReplaceAll(charsNoSpace, "\n", "")
	charsNoSpace = strings.ReplaceAll(charsNoSpace, "\t", "")
	charsNoSpace = strings.ReplaceAll(charsNoSpace, "\r", "")
	stats.Characters = len(charsNoSpace)

	// Word count (split by whitespace and filter empty)
	words := regexp.MustCompile(`\S+`).FindAllString(content, -1)
	stats.Words = len(words)

	// Reading time (average 200 words per minute, minimum 1 minute)
	stats.ReadingTimeMinutes = stats.Words / 200
	if stats.ReadingTimeMinutes < 1 && stats.Words > 0 {
		stats.ReadingTimeMinutes = 1
	}

	// Line count
	stats.Lines = len(strings.Split(content, "\n"))

	// Paragraph count (blocks separated by blank lines)
	paragraphs := strings.Split(content, "\n\n")
	paragraphCount := 0
	for _, p := range paragraphs {
		if strings.TrimSpace(p) != "" {
			paragraphCount++
		}
	}
	stats.Paragraphs = paragraphCount

	// Sentence count: punctuation [.!?]+ followed by space or end-of-string
	sentences := regexp.MustCompile(`[.!?]+(?:\s|$)`).FindAllString(content, -1)
	stats.Sentences = len(sentences)

	// List items: lines starting with -, *, + or a number (e.g. 1., 10.), excluding tasks [-]
	// Note: Using (?!) negative lookahead is not supported in Go regexp
	// Instead, match list items and filter out task lists in a separate pass
	listItemRegex := regexp.MustCompile(`(?m)^\s*(?:[-*+]|\d+\.)\s+(?:\[[ xX]\]\s+)?(.+)$`)
	matches := listItemRegex.FindAllStringSubmatch(content, -1)
	// Count only non-task list items (task lists have [ ] or [x] pattern)
	taskListRegex := regexp.MustCompile(`(?m)^\s*(?:[-*+]|\d+\.)\s+\[[ xX]\]`)
	for _, match := range matches {
		line := match[0]
		if !taskListRegex.MatchString(line) {
			stats.ListItems++
		}
	}

	// Tables: count markdown table separator rows (| --- | --- |)
	tables := regexp.MustCompile(`^\s*\|(?:\s*:?-+:?\s*\|){1,}\s*$`).FindAllString(content, -1)
	stats.Tables = len(tables)

	// Markdown link count (standard [text](url) format)
	markdownLinks := regexp.MustCompile(`\[([^\]]+)\]\(([^\)]+)\)`).FindAllString(content, -1)
	stats.Links = len(markdownLinks)

	// Internal link count (standard markdown links to .md files)
	markdownInternalLinks := regexp.MustCompile(`\[([^\]]+)\]\(([^\)]+\.md)\)`).FindAllString(content, -1)

	// Wikilink count ([[note]] or [[note|display text]] format - Obsidian style)
	wikilinks := regexp.MustCompile(`\[\[([^\]|]+)(?:\|[^\]]+)?\]\]`).FindAllString(content, -1)
	stats.Wikilinks = len(wikilinks)

	// Total links and internal links
	stats.InternalLinks = len(markdownInternalLinks) + stats.Wikilinks
	stats.ExternalLinks = stats.Links - len(markdownInternalLinks)

	// Code block count
	codeBlocks := regexp.MustCompile("```[\\s\\S]*?```").FindAllString(content, -1)
	stats.CodeBlocks = len(codeBlocks)

	// Inline code count
	inlineCode := regexp.MustCompile("`[^`]+`").FindAllString(content, -1)
	stats.InlineCode = len(inlineCode)

	// Heading count by level (using multiline mode with (?m))
	stats.Headings.H1 = len(regexp.MustCompile(`(?m)^# `).FindAllString(content, -1))
	stats.Headings.H2 = len(regexp.MustCompile(`(?m)^## `).FindAllString(content, -1))
	stats.Headings.H3 = len(regexp.MustCompile(`(?m)^### `).FindAllString(content, -1))

	// Task count (checkboxes)
	totalTasks := len(regexp.MustCompile(`- \[[ x]\]`).FindAllString(content, -1))
	completedTasks := len(regexp.MustCompile(`(?i)- \[x\]`).FindAllString(content, -1))
	stats.Tasks.Total = totalTasks
	stats.Tasks.Completed = completedTasks
	stats.Tasks.Pending = totalTasks - completedTasks

	// Image count
	images := regexp.MustCompile(`!\[([^\]]*)\]\(([^\)]+)\)`).FindAllString(content, -1)
	stats.Images = len(images)

	// Blockquote count
	blockquotes := regexp.MustCompile(`^> `).FindAllString(content, -1)
	stats.Blockquotes = len(blockquotes)

	return stats
}
