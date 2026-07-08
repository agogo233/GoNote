package services

import (
	"html"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"gonote/internal/models"
)

// ContextWindowSize is the number of characters shown before and after a match in search results
const ContextWindowSize = 50

// SearchService handles search operations
type SearchService struct {
	notesDir    string
	noteService *NoteService
	regexCache  sync.Map
}

// NewSearchService creates a new SearchService.
// Pass a shared *NoteService to leverage caching; omit or pass nil to create a new one per search.
func NewSearchService(notesDir string, noteService ...*NoteService) *SearchService {
	s := &SearchService{notesDir: notesDir}
	if len(noteService) > 0 {
		s.noteService = noteService[0]
	}
	return s
}

// getOrCompileRegex 从缓存中获取已编译的正则，避免热路径上反复编译。
func (s *SearchService) getOrCompileRegex(pattern string) (*regexp.Regexp, error) {
	if cached, ok := s.regexCache.Load(pattern); ok {
		return cached.(*regexp.Regexp), nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	s.regexCache.Store(pattern, re)
	return re, nil
}

// Search performs full-text search through note contents
func (s *SearchService) Search(query string) ([]models.SearchResult, error) {
	results := []models.SearchResult{}

	ns := s.noteService
	if ns == nil {
		ns = NewNoteService(s.notesDir)
	}
	notes, _, err := ns.ScanNotes(false)
	if err != nil {
		return nil, err
	}

	// Escape the query for regex
	escapedQuery := regexp.QuoteMeta(query)

	// Case-insensitive pattern
	pattern, err := s.getOrCompileRegex("(?i)" + escapedQuery)
	if err != nil {
		return nil, err
	}

	for _, note := range notes {
		fullPath := filepath.Join(s.notesDir, note.Path)
		content, err := readFileContent(fullPath)
		if err != nil {
			continue
		}

		matches := pattern.FindAllStringIndex(content, -1)
		if len(matches) == 0 {
			continue
		}

		matchedLines := []models.MatchContext{}
		for i, match := range matches {
			if i >= 3 { // Limit to 3 matches per file
				break
			}

			startIndex := match[0]
			endIndex := match[1]
			matchedText := content[startIndex:endIndex]

			// Create slice window: ±ContextWindowSize characters around match
			contextStart := startIndex - ContextWindowSize
			if contextStart < 0 {
				contextStart = 0
			}
			contextEnd := endIndex + ContextWindowSize
			if contextEnd > len(content) {
				contextEnd = len(content)
			}

			// Extract and clean parts (newlines → spaces)
			before := strings.ReplaceAll(content[contextStart:startIndex], "\n", " ")
			after := strings.ReplaceAll(content[endIndex:contextEnd], "\n", " ")
			matchedClean := strings.ReplaceAll(matchedText, "\n", " ")

			// Escape HTML
			before = html.EscapeString(before)
			after = html.EscapeString(after)
			matchedClean = html.EscapeString(matchedClean)

			// Build snippet with <mark> highlight
			snippet := before + `<mark class="search-highlight">` + matchedClean + `</mark>` + after

			// Add ellipsis if truncated
			if contextStart > 0 {
				snippet = "..." + snippet
			}
			if contextEnd < len(content) {
				snippet = snippet + "..."
			}

			// Calculate line number
			lineNumber := strings.Count(content[:startIndex], "\n") + 1

			matchedLines = append(matchedLines, models.MatchContext{
				LineNumber: lineNumber,
				Context:    snippet,
			})
		}

		if len(matchedLines) > 0 {
			// Extract the actual title from content
			title := extractTitle(content, note.Path)
			results = append(results, models.SearchResult{
				Name:    title,
				Path:    note.Path,
				Folder:  note.Folder,
				Type:    note.Type,
				Matches: matchedLines,
			})
		}
	}

	return results, nil
}

// SearchFilename searches only note filenames and directory paths (legacy fallback).
// Matches against note.Path (relative) and the basename (with -/_ normalized to spaces).
func (s *SearchService) SearchFilename(query string) ([]models.SearchResult, error) {
	results := []models.SearchResult{}

	ns := s.noteService
	if ns == nil {
		ns = NewNoteService(s.notesDir)
	}
	notes, _, err := ns.ScanNotes(false)
	if err != nil {
		return nil, err
	}

	escapedQuery := regexp.QuoteMeta(query)
	pattern, err := s.getOrCompileRegex("(?i)" + escapedQuery)
	if err != nil {
		return nil, err
	}
	queryLower := strings.ToLower(query)

	for _, note := range notes {
		name := filepath.Base(note.Path)
		name = strings.TrimSuffix(name, filepath.Ext(name))
		name = strings.ReplaceAll(name, "-", " ")
		name = strings.ReplaceAll(name, "_", " ")

		if !pattern.MatchString(note.Path) && !strings.Contains(strings.ToLower(name), queryLower) {
			continue
		}

		results = append(results, models.SearchResult{
			Name:   name,
			Path:   note.Path,
			Folder: note.Folder,
			Type:   note.Type,
		})
	}

	return results, nil
}

// readFileContent reads file content safely
func readFileContent(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
