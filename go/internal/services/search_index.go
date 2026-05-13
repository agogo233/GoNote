package services

import (
	"container/list"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"gonote/internal/models"
)

// Pre-compiled regex patterns for performance
var (
	cjkRegex    = regexp.MustCompile(`[\p{Han}\p{Hiragana}\p{Katakana}\p{Hangul}]`)
	asciiWordRE = regexp.MustCompile(`[a-z0-9]+`)
	cjkWordRE   = regexp.MustCompile(`[\p{Han}\p{Hiragana}\p{Katakana}\p{Hangul}]+`)
)

// SearchIndex provides fast full-text search using inverted index
// Uses double-buffering for non-blocking index rebuilds
type SearchIndex struct {
	mu            sync.RWMutex
	index         map[string]*list.List // term -> list of IndexEntry
	titleIndex    map[string]*list.List // term -> list of TitleEntry (title-only index)
	titleMap      map[string]string     // notePath -> title (for title lookup)
	notesDir      string
	cache         *Cache
	noteService   *NoteService   // Shared NoteService for reusing cache
	searchService *SearchService // SearchService for disk-scan fallback
}

// TitleEntry represents a title match with score
type TitleEntry struct {
	NotePath string
	Title    string
	Score    float64 // relevance score for ranking
}

// IndexEntry represents a single occurrence of a term in a note
type IndexEntry struct {
	NotePath string
	Position int // byte position in file
}

// NewSearchIndex creates a new search index with shared NoteService
func NewSearchIndex(notesDir string, noteService *NoteService) *SearchIndex {
	return &SearchIndex{
		index:         make(map[string]*list.List),
		titleIndex:    make(map[string]*list.List),
		titleMap:      make(map[string]string),
		notesDir:      notesDir,
		cache:         NewCache(10000, 15*time.Minute), // Cache index entries for 15 minutes
		noteService:   noteService,
		searchService: NewSearchService(notesDir),
	}
}

// BuildIndex builds the full search index from all notes
// Uses double-buffering: builds new index without lock, then swaps atomically
func (si *SearchIndex) BuildIndex() error {
	// Phase 1: Build new index without holding lock (allows concurrent searches)
	newIndex := make(map[string]*list.List)
	newTitleIndex := make(map[string]*list.List)
	newTitleMap := make(map[string]string)

	// Use shared NoteService to leverage its cache
	if si.noteService == nil {
		si.noteService = NewNoteService(si.notesDir)
	}
	notes, _, err := si.noteService.ScanNotes(false)
	if err != nil {
		return err
	}

	// Index each note into the new index
	for _, note := range notes {
		if err := si.indexNoteTo(note.Path, newIndex, newTitleIndex, newTitleMap); err != nil {
			// Log error but continue indexing other notes
			continue
		}
	}

	// Phase 2: Swap index atomically with brief write lock
	si.mu.Lock()
	si.index = newIndex
	si.titleIndex = newTitleIndex
	si.titleMap = newTitleMap
	si.mu.Unlock()

	return nil
}

// indexNoteTo indexes a single note into the provided index map
func (si *SearchIndex) indexNoteTo(notePath string, index map[string]*list.List, titleIndex map[string]*list.List, titleMap map[string]string) error {
	fullPath := filepath.Join(si.notesDir, notePath)
	content, err := readFileContent(fullPath)
	if err != nil {
		return err
	}

	// Extract title from frontmatter or first line
	title := extractTitle(content, notePath)
	titleMap[notePath] = title

	// Tokenize content for full-text index
	terms := tokenize(content)

	// Add each term to full-text index
	for pos, term := range terms {
		if _, ok := index[term]; !ok {
			index[term] = list.New()
		}
		index[term].PushBack(IndexEntry{
			NotePath: notePath,
			Position: pos,
		})
	}

	// Tokenize title for title index
	titleTerms := tokenize(title)
	for _, term := range titleTerms {
		if _, ok := titleIndex[term]; !ok {
			titleIndex[term] = list.New()
		}
		titleIndex[term].PushBack(TitleEntry{
			NotePath: notePath,
			Title:    title,
			Score:    0, // score calculated at query time
		})
	}

	// 同时索引文件名（笔记名称）用于标题搜索
	fileName := extractFileName(notePath)
	fileNameTerms := tokenize(fileName)
	for _, term := range fileNameTerms {
		if _, ok := titleIndex[term]; !ok {
			titleIndex[term] = list.New()
		}
		titleIndex[term].PushBack(TitleEntry{
			NotePath: notePath,
			Title:    title,
			Score:    0,
		})
	}

	return nil
}

// indexNote indexes a single note into the main index (must hold lock)
// This version removes old entries first, then re-indexes.
func (si *SearchIndex) indexNote(notePath string) error {
	fullPath := filepath.Join(si.notesDir, notePath)
	content, err := readFileContent(fullPath)
	if err != nil {
		return err
	}

	// Extract and store title
	title := extractTitle(content, notePath)
	si.titleMap[notePath] = title

	// Remove old title entries from titleIndex
	for term, entries := range si.titleIndex {
		for e := entries.Front(); e != nil; {
			next := e.Next()
			entry := e.Value.(TitleEntry)
			if entry.NotePath == notePath {
				entries.Remove(e)
			}
			e = next
		}
		if entries.Len() == 0 {
			delete(si.titleIndex, term)
		}
	}

	// Tokenize content for full-text index
	terms := tokenize(content)

	// Remove old content entries from index
	for term, entries := range si.index {
		for e := entries.Front(); e != nil; {
			next := e.Next()
			entry := e.Value.(IndexEntry)
			if entry.NotePath == notePath {
				entries.Remove(e)
			}
			e = next
		}
		if entries.Len() == 0 {
			delete(si.index, term)
		}
	}

	// Add each term to full-text index
	for pos, term := range terms {
		if _, ok := si.index[term]; !ok {
			si.index[term] = list.New()
		}
		si.index[term].PushBack(IndexEntry{
			NotePath: notePath,
			Position: pos,
		})
	}

	// Tokenize title for title index
	titleTerms := tokenize(title)
	for _, term := range titleTerms {
		if _, ok := si.titleIndex[term]; !ok {
			si.titleIndex[term] = list.New()
		}
		si.titleIndex[term].PushBack(TitleEntry{
			NotePath: notePath,
			Title:    title,
			Score:    0,
		})
	}

	// 同时索引文件名（笔记名称）
	fileName := extractFileName(notePath)
	fileNameTerms := tokenize(fileName)
	for _, term := range fileNameTerms {
		if _, ok := si.titleIndex[term]; !ok {
			si.titleIndex[term] = list.New()
		}
		si.titleIndex[term].PushBack(TitleEntry{
			NotePath: notePath,
			Title:    title,
			Score:    0,
		})
	}

	return nil
}

// indexNoteFresh indexes a single note WITHOUT removing old entries first.
// Use this when the caller has already removed old entries (e.g. UpdateIndex).
// Must hold the write lock when calling.
func (si *SearchIndex) indexNoteFresh(notePath string) error {
	fullPath := filepath.Join(si.notesDir, notePath)
	content, err := readFileContent(fullPath)
	if err != nil {
		return err
	}

	title := extractTitle(content, notePath)
	si.titleMap[notePath] = title

	terms := tokenize(content)
	for pos, term := range terms {
		if _, ok := si.index[term]; !ok {
			si.index[term] = list.New()
		}
		si.index[term].PushBack(IndexEntry{
			NotePath: notePath,
			Position: pos,
		})
	}

	titleTerms := tokenize(title)
	for _, term := range titleTerms {
		if _, ok := si.titleIndex[term]; !ok {
			si.titleIndex[term] = list.New()
		}
		si.titleIndex[term].PushBack(TitleEntry{
			NotePath: notePath,
			Title:    title,
			Score:    0,
		})
	}

	fileName := extractFileName(notePath)
	fileNameTerms := tokenize(fileName)
	for _, term := range fileNameTerms {
		if _, ok := si.titleIndex[term]; !ok {
			si.titleIndex[term] = list.New()
		}
		si.titleIndex[term].PushBack(TitleEntry{
			NotePath: notePath,
			Title:    title,
			Score:    0,
		})
	}

	return nil
}

// UpdateIndex updates the index for a single note (incremental)
// Calls removeNoteFromIndex first, then indexNoteFresh (which skips the redundant removal).
func (si *SearchIndex) UpdateIndex(notePath string) error {
	si.mu.Lock()
	defer si.mu.Unlock()

	si.removeNoteFromIndex(notePath)

	return si.indexNoteFresh(notePath)
}

// RemoveFromIndex removes a note from the index
func (si *SearchIndex) RemoveFromIndex(notePath string) {
	si.mu.Lock()
	defer si.mu.Unlock()

	si.removeNoteFromIndex(notePath)
}

// removeNoteFromIndex removes all entries for a note (must hold lock)
func (si *SearchIndex) removeNoteFromIndex(notePath string) {
	for term, entries := range si.index {
		for e := entries.Front(); e != nil; {
			next := e.Next()
			entry := e.Value.(IndexEntry)
			if entry.NotePath == notePath {
				entries.Remove(e)
			}
			e = next
		}

		if entries.Len() == 0 {
			delete(si.index, term)
		}
	}

	for term, entries := range si.titleIndex {
		for e := entries.Front(); e != nil; {
			next := e.Next()
			entry := e.Value.(TitleEntry)
			if entry.NotePath == notePath {
				entries.Remove(e)
			}
			e = next
		}

		if entries.Len() == 0 {
			delete(si.titleIndex, term)
		}
	}

	delete(si.titleMap, notePath)
}

// Search performs a search using the inverted index
// Uses read lock for concurrent search access
// Supports prefix matching for partial searches (e.g., "gol" matches "golang")
// CJK and non-CJK queries use the same unified path: tokenize → prefix match → verify
func (si *SearchIndex) Search(query string) ([]models.SearchResult, error) {
	si.mu.RLock()
	defer si.mu.RUnlock()

	if query == "" {
		return []models.SearchResult{}, nil
	}

	terms := tokenize(query)
	if len(terms) == 0 {
		return si.searchFromDisk(query)
	}

	var results []models.SearchResult
	matchedNotes := make(map[string]bool)

	firstTerm := terms[0]
	candidateNotes := si.findNotesWithPrefix(firstTerm)

	if len(candidateNotes) == 0 {
		return si.searchFromDisk(query)
	}

	for notePath := range candidateNotes {
		if matchedNotes[notePath] {
			continue
		}

		if si.noteContainsTermsWithPrefix(notePath, terms) {
			content, err := si.noteService.GetNoteContent(notePath)
			if err != nil {
				continue
			}

			if si.contentContainsAllKeywords(content, terms) {
				matchedNotes[notePath] = true
				result := si.buildSearchResult(notePath, content, query)
				results = append(results, result)
			}
		}
	}

	if len(results) == 0 {
		return si.searchFromDisk(query)
	}

	return results, nil
}

// searchFromDisk searches for non-CJK queries by scanning all notes from disk
// Used as fallback when index search returns no results
func (si *SearchIndex) searchFromDisk(query string) ([]models.SearchResult, error) {
	// Use NoteService to scan all notes
	notes, _, err := si.noteService.ScanNotes(false)
	if err != nil {
		return nil, err
	}

	// Escape the query for regex
	escapedQuery := regexp.QuoteMeta(query)

	// Case-insensitive pattern
	pattern, err := regexp.Compile("(?i)" + escapedQuery)
	if err != nil {
		return nil, err
	}

	return si.searchNotesWithPattern(notes, pattern, query)
}

// searchNotesWithPattern searches notes using a regex pattern and builds results
func (si *SearchIndex) searchNotesWithPattern(notes []models.Note, pattern *regexp.Regexp, query string) ([]models.SearchResult, error) {
	results := []models.SearchResult{}
	matchedNotes := make(map[string]bool)

	for _, note := range notes {
		if matchedNotes[note.Path] {
			continue
		}

		content, err := si.noteService.GetNoteContent(note.Path)
		if err != nil {
			continue
		}

		if pattern.MatchString(content) {
			matchedNotes[note.Path] = true
			result := si.buildSearchResult(note.Path, content, query)
			results = append(results, result)
		}
	}

	return results, nil
}

// findNotesWithPrefix finds all notes that contain terms starting with the given prefix
func (si *SearchIndex) findNotesWithPrefix(prefix string) map[string]bool {
	notes := make(map[string]bool)
	
	for term, entries := range si.index {
		if strings.HasPrefix(term, prefix) {
			for e := entries.Front(); e != nil; e = e.Next() {
				entry := e.Value.(IndexEntry)
				notes[entry.NotePath] = true
			}
		}
	}
	
	return notes
}

// noteContainsTermsWithPrefix checks if a note contains all terms (with prefix matching)
func (si *SearchIndex) noteContainsTermsWithPrefix(notePath string, terms []string) bool {
	for _, term := range terms {
		found := false
		// Check if any indexed term starts with this query term
		for indexedTerm, entries := range si.index {
			if strings.HasPrefix(indexedTerm, term) {
				for e := entries.Front(); e != nil; e = e.Next() {
					entry := e.Value.(IndexEntry)
					if entry.NotePath == notePath {
						found = true
						break
					}
				}
			}
			if found {
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// SearchByTitle searches only note titles with prefix and fuzzy matching
func (si *SearchIndex) SearchByTitle(query string) ([]models.SearchResult, error) {
	si.mu.RLock()
	defer si.mu.RUnlock()

	if query == "" {
		return []models.SearchResult{}, nil
	}

	queryLower := strings.ToLower(query)
	queryTerms := tokenize(query)

	if len(queryTerms) == 0 {
		// Single character or short query - try prefix matching on titleIndex
		return si.searchTitleByPrefix(queryLower)
	}

	// Multi-term query: find notes whose titles contain all terms (with prefix matching)
	type titleScore struct {
		notePath string
		title    string
		score    float64
	}
	var matches []titleScore
	matchedNotes := make(map[string]bool)

	// Find candidate notes using first term prefix
	candidates := si.findTitlesWithPrefix(queryTerms[0])

	for notePath, title := range candidates {
		if matchedNotes[notePath] {
			continue
		}

		// Check if title or filename contains all query terms (prefix matching)
		if si.titleContainsTerms(notePath, title, queryTerms) {
			matchedNotes[notePath] = true
			// 计算分数：如果标题包含所有查询词，使用标题计分；否则使用文件名计分
			titleLower := strings.ToLower(title)
			fileName := extractFileName(notePath)
			allInTitle := true
			for _, term := range queryTerms {
				if !strings.Contains(titleLower, term) {
					allInTitle = false
					break
				}
			}
			var score float64
			if allInTitle {
				score = si.calculateTitleScore(title, queryLower)
			} else {
				score = si.calculateTitleScore(fileName, queryLower)
			}
			matches = append(matches, titleScore{
				notePath: notePath,
				title:    title,
				score:    score,
			})
		}
	}

	// Sort by score descending
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].score > matches[j].score
	})

	// Build search results
	var results []models.SearchResult
	for _, m := range matches {
		result := si.buildTitleResult(m.notePath, m.title, query)
		result.Score = m.score
		results = append(results, result)
	}

	// If no results from index, fall back to scanning all notes from disk
	// This handles cases where the index might not be fully up-to-date
	if len(results) == 0 {
		return si.searchTitleFromDisk(query)
	}

	return results, nil
}

// findTitlesWithPrefix finds all notes whose titles contain terms starting with the prefix
// Also includes notes where indexed title terms contain the prefix (Contains match for CJK)
func (si *SearchIndex) findTitlesWithPrefix(prefix string) map[string]string {
	result := make(map[string]string)

	for term, entries := range si.titleIndex {
		if strings.HasPrefix(term, prefix) || strings.Contains(term, prefix) {
			for e := entries.Front(); e != nil; e = e.Next() {
				entry := e.Value.(TitleEntry)
				if title, ok := si.titleMap[entry.NotePath]; ok {
					result[entry.NotePath] = title
				}
			}
		}
	}

	return result
}

// titleContainsTerms checks if a title or filename contains all query terms as substrings
func (si *SearchIndex) titleContainsTerms(notePath string, title string, terms []string) bool {
	titleLower := strings.ToLower(title)
	fileName := extractFileName(notePath)
	fileNameLower := strings.ToLower(fileName)
	for _, term := range terms {
		if !strings.Contains(titleLower, term) && !strings.Contains(fileNameLower, term) {
			return false
		}
	}
	return true
}

// calculateTitleScore calculates relevance score for title matching
func (si *SearchIndex) calculateTitleScore(title string, query string) float64 {
	titleLower := strings.ToLower(title)
	score := 0.0

	// Exact match gets highest score
	if titleLower == query {
		score = 100.0
		return score
	}

	// Starts with query gets high score
	if strings.HasPrefix(titleLower, query) {
		score = 80.0
		return score
	}

	// Contains query gets medium score
	if strings.Contains(titleLower, query) {
		score = 60.0
		// Bonus for word boundary match
		if strings.HasPrefix(titleLower, query) || strings.Contains(titleLower, " "+query) {
			score += 10.0
		}
		return score
	}

	// Fuzzy: count matched terms from tokenized query
	queryTerms := tokenize(query)
	if len(queryTerms) == 0 {
		// For single-char queries, try direct containment
		if strings.Contains(titleLower, query) {
			score = 40.0
		}
		return score
	}

	matchedTerms := 0
	for _, term := range queryTerms {
		if strings.Contains(titleLower, term) {
			matchedTerms++
		}
	}
	if len(queryTerms) > 0 {
		score = float64(matchedTerms) / float64(len(queryTerms)) * 40.0
	}

	return score
}

// searchTitleFromDisk searches titles by scanning all notes from disk
// Used as fallback when title index search returns no results
func (si *SearchIndex) searchTitleFromDisk(query string) ([]models.SearchResult, error) {
	// Use NoteService to scan all notes
	notes, _, err := si.noteService.ScanNotes(false)
	if err != nil {
		return nil, err
	}

	queryLower := strings.ToLower(query)
	var results []models.SearchResult

	for _, note := range notes {
		content, err := si.noteService.GetNoteContent(note.Path)
		if err != nil {
			continue
		}

		// Extract title from content
		title := extractTitle(content, note.Path)
		titleLower := strings.ToLower(title)

		// Check if title contains the query (case-insensitive)
		if strings.Contains(titleLower, queryLower) {
			result := si.buildTitleResult(note.Path, title, query)
			result.Score = si.calculateTitleScore(title, queryLower)
			results = append(results, result)
		} else {
			// 同时检查文件名匹配
			fileName := extractFileName(note.Path)
			if strings.Contains(fileName, queryLower) {
				result := si.buildTitleResult(note.Path, title, query)
				result.Score = si.calculateTitleScore(fileName, queryLower)
				results = append(results, result)
			}
		}
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}

// searchTitleByPrefix handles single-term or short prefix searches
func (si *SearchIndex) searchTitleByPrefix(prefix string) ([]models.SearchResult, error) {
	type titleScore struct {
		notePath string
		title    string
		score    float64
	}
	var matches []titleScore
	seen := make(map[string]bool)

	prefixLower := strings.ToLower(prefix)

	for term, entries := range si.titleIndex {
		if strings.HasPrefix(term, prefixLower) {
			for e := entries.Front(); e != nil; e = e.Next() {
				entry := e.Value.(TitleEntry)
				if seen[entry.NotePath] {
					continue
				}
				seen[entry.NotePath] = true

				title := entry.Title
				score := si.calculateTitleScore(title, prefixLower)
				if score == 0 {
					score = si.calculateTitleScore(extractFileName(entry.NotePath), prefixLower)
				}
				matches = append(matches, titleScore{
					notePath: entry.NotePath,
					title:    title,
					score:    score,
				})
			}
		}
	}

	// Sort by score
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].score > matches[j].score
	})

	var results []models.SearchResult
	for _, m := range matches {
		result := si.buildTitleResult(m.notePath, m.title, prefix)
		result.Score = m.score
		results = append(results, result)
	}

	// If no results from index, fall back to scanning all notes from disk
	if len(results) == 0 {
		return si.searchTitleFromDisk(prefix)
	}

	return results, nil
}

// SearchSmart performs smart search: title matches first, content matches as fallback
func (si *SearchIndex) SearchSmart(query string) ([]models.SearchResult, error) {
	si.mu.RLock()
	defer si.mu.RUnlock()

	if query == "" {
		return []models.SearchResult{}, nil
	}

	// Step 1: Search titles (high priority)
	titleResults, _ := si.searchByTitleInternal(query)

	// Step 2: Search full content (fallback)
	contentResults, _ := si.searchInternal(query)

	// Step 3: Merge results, title matches first with boosted score
	seen := make(map[string]bool)
	var results []models.SearchResult

	// Add title matches first (already scored)
	for _, r := range titleResults {
		seen[r.Path] = true
		// Boost title matches by adding 50 to their score
		r.Score += 50.0
		results = append(results, r)
	}

	// Add content matches that weren't in title matches
	for _, r := range contentResults {
		if !seen[r.Path] {
			results = append(results, r)
		}
	}

	return results, nil
}

// searchByTitleInternal is the internal version without lock (caller holds lock)
func (si *SearchIndex) searchByTitleInternal(query string) ([]models.SearchResult, error) {
	if query == "" {
		return []models.SearchResult{}, nil
	}

	queryLower := strings.ToLower(query)
	queryTerms := tokenize(query)

	if len(queryTerms) == 0 {
		return si.searchTitleByPrefixInternal(queryLower)
	}

	type titleScore struct {
		notePath string
		title    string
		score    float64
	}
	var matches []titleScore
	matchedNotes := make(map[string]bool)

	candidates := si.findTitlesWithPrefix(queryTerms[0])

	for notePath, title := range candidates {
		if matchedNotes[notePath] {
			continue
		}

		if si.titleContainsTerms(notePath, title, queryTerms) {
			matchedNotes[notePath] = true
			// 计算分数：如果标题包含所有查询词，使用标题计分；否则使用文件名计分
			titleLower := strings.ToLower(title)
			fileName := extractFileName(notePath)
			allInTitle := true
			for _, term := range queryTerms {
				if !strings.Contains(titleLower, term) {
					allInTitle = false
					break
				}
			}
			var score float64
			if allInTitle {
				score = si.calculateTitleScore(title, queryLower)
			} else {
				score = si.calculateTitleScore(fileName, queryLower)
			}
			matches = append(matches, titleScore{
				notePath: notePath,
				title:    title,
				score:    score,
			})
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].score > matches[j].score
	})

	var results []models.SearchResult
	for _, m := range matches {
		result := si.buildTitleResult(m.notePath, m.title, query)
		result.Score = m.score
		results = append(results, result)
	}

	// If no results from index, fall back to scanning all notes from disk
	if len(results) == 0 {
		return si.searchTitleFromDisk(query)
	}

	return results, nil
}

// searchTitleByPrefixInternal is the internal version without lock
func (si *SearchIndex) searchTitleByPrefixInternal(prefix string) ([]models.SearchResult, error) {
	type titleScore struct {
		notePath string
		title    string
		score    float64
	}
	var matches []titleScore
	seen := make(map[string]bool)

	prefixLower := strings.ToLower(prefix)

	for term, entries := range si.titleIndex {
		if strings.HasPrefix(term, prefixLower) {
			for e := entries.Front(); e != nil; e = e.Next() {
				entry := e.Value.(TitleEntry)
				if seen[entry.NotePath] {
					continue
				}
				seen[entry.NotePath] = true

				score := si.calculateTitleScore(entry.Title, prefixLower)
				if score == 0 {
					score = si.calculateTitleScore(extractFileName(entry.NotePath), prefixLower)
				}
				matches = append(matches, titleScore{
					notePath: entry.NotePath,
					title:    entry.Title,
					score:    score,
				})
			}
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].score > matches[j].score
	})

	var results []models.SearchResult
	for _, m := range matches {
		result := si.buildTitleResult(m.notePath, m.title, prefix)
		result.Score = m.score
		results = append(results, result)
	}

	// If no results from index, fall back to scanning all notes from disk
	if len(results) == 0 {
		return si.searchTitleFromDisk(prefix)
	}

	return results, nil
}

// searchInternal is the internal version of Search without lock (caller holds lock)
func (si *SearchIndex) searchInternal(query string) ([]models.SearchResult, error) {
	if query == "" {
		return []models.SearchResult{}, nil
	}

	terms := tokenize(query)
	if len(terms) == 0 {
		return si.searchFromDisk(query)
	}

	var results []models.SearchResult
	matchedNotes := make(map[string]bool)

	firstTerm := terms[0]
	candidateNotes := si.findNotesWithPrefix(firstTerm)

	if len(candidateNotes) == 0 {
		return si.searchFromDisk(query)
	}

	for notePath := range candidateNotes {
		if matchedNotes[notePath] {
			continue
		}

		if si.noteContainsTermsWithPrefix(notePath, terms) {
			content, err := si.noteService.GetNoteContent(notePath)
			if err != nil {
				continue
			}

			if si.contentContainsAllKeywords(content, terms) {
				matchedNotes[notePath] = true
				result := si.buildSearchResult(notePath, content, query)
				results = append(results, result)
			}
		}
	}

	if len(results) == 0 {
		return si.searchFromDisk(query)
	}

	return results, nil
}

// buildTitleResult builds a search result for title-only matches
func (si *SearchIndex) buildTitleResult(notePath string, title string, query string) models.SearchResult {
	folder := filepath.Dir(notePath)
	if folder == "." {
		folder = ""
	}

	fileType := getFileType(notePath)

	// Create a match context showing the title is matched
	context := title
	if query != "" {
		// Highlight the query in the title
		escapedQuery := regexp.QuoteMeta(query)
		pattern := regexp.MustCompile("(?i)" + escapedQuery)
		context = pattern.ReplaceAllString(title, "<mark class=\"search-highlight\">$0</mark>")
	}

	return models.SearchResult{
		Name:   title,
		Path:   notePath,
		Folder: folder,
		Type:   fileType,
		Matches: []models.MatchContext{
			{
				LineNumber: 1,
				Context:    context,
			},
		},
	}
}

// contentContainsAllKeywords verifies that note content contains all query keywords as substrings
func (si *SearchIndex) contentContainsAllKeywords(content string, terms []string) bool {
	contentLower := strings.ToLower(content)
	for _, term := range terms {
		if !strings.Contains(contentLower, term) {
			return false
		}
	}
	return true
}

// buildSearchResult builds a search result with context
func (si *SearchIndex) buildSearchResult(notePath string, content string, query string) models.SearchResult {
	// Escape query for regex
	escapedQuery := regexp.QuoteMeta(query)
	pattern := regexp.MustCompile("(?i)" + escapedQuery)

	matches := pattern.FindAllStringIndex(content, -1)

	var matchedLines []models.MatchContext
	for i, match := range matches {
		if i >= 3 { // Limit to 3 matches per file
			break
		}

		startIndex := match[0]
		endIndex := match[1]

		// Create context window: ±50 characters
		contextStart := startIndex - 50
		if contextStart < 0 {
			contextStart = 0
		}
		contextEnd := endIndex + 50
		if contextEnd > len(content) {
			contextEnd = len(content)
		}

		// Extract context
		context := content[contextStart:contextEnd]
		context = strings.ReplaceAll(context, "\n", " ")

		// Calculate line number
		lineNumber := strings.Count(content[:startIndex], "\n") + 1

		matchedLines = append(matchedLines, models.MatchContext{
			LineNumber: lineNumber,
			Context:    context,
		})
	}

	// Extract the actual title from content
	title := extractTitle(content, notePath)
	folder := filepath.Dir(notePath)
	if folder == "." {
		folder = ""
	}

	// Determine file type based on extension
	fileType := getFileType(notePath)

	return models.SearchResult{
		Name:    title,
		Path:    notePath,
		Folder:  folder,
		Type:    fileType,
		Matches: matchedLines,
	}
}

// getFileType determines the file type based on extension
func getFileType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".md":
		return "note"
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".bmp":
		return "image"
	case ".mp3", ".wav", ".ogg", ".m4a", ".flac":
		return "audio"
	case ".mp4", ".webm", ".mov", ".avi", ".mkv":
		return "video"
	case ".pdf":
		return "document"
	default:
		return "note" // Default to note
	}
}

// GetIndexedTerms returns all indexed terms (for debugging)
func (si *SearchIndex) GetIndexedTerms() []string {
	si.mu.RLock()
	defer si.mu.RUnlock()

	terms := make([]string, 0, len(si.index))
	for term := range si.index {
		terms = append(terms, term)
	}
	return terms
}

// GetIndexSize returns the number of unique terms in the index
func (si *SearchIndex) GetIndexSize() int {
	si.mu.RLock()
	defer si.mu.RUnlock()

	return len(si.index)
}

// extractTitle extracts the title from note content or derives it from the filename
func extractTitle(content string, notePath string) string {
	// Try to extract title from frontmatter
	lines := strings.SplitN(content, "\n", 30) // Check first 30 lines for frontmatter
	inFrontmatter := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			}
			break
		}
		if inFrontmatter && strings.HasPrefix(trimmed, "title:") {
			title := strings.TrimPrefix(trimmed, "title:")
			title = strings.TrimSpace(title)
			// Remove quotes if present
			title = strings.Trim(title, "\"'")
			if title != "" {
				return title
			}
		}
	}

	// Fallback: use first non-empty, non-frontmatter line
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || trimmed == "---" || strings.HasPrefix(trimmed, "#") && strings.HasPrefix(trimmed, "##") {
			continue
		}
		// If it's a level-1 heading, use it as title
		if strings.HasPrefix(trimmed, "# ") {
			return strings.TrimPrefix(trimmed, "# ")
		}
		if trimmed != "---" {
			// Return first meaningful line (truncated)
			if len(trimmed) > 100 {
				return trimmed[:100]
			}
			return trimmed
		}
	}

	// Last fallback: derive from filename
	name := filepath.Base(notePath)
	name = strings.TrimSuffix(name, filepath.Ext(name))
	// Replace hyphens/underscores with spaces
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.ReplaceAll(name, "_", " ")
	return name
}

// extractFileName extracts searchable name from file path (without extension, hyphens/underscores as spaces)
func extractFileName(notePath string) string {
	name := filepath.Base(notePath)
	name = strings.TrimSuffix(name, filepath.Ext(name))
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.ReplaceAll(name, "_", " ")
	return name
}

// tokenize splits text into terms, supporting both ASCII and CJK characters
// CJK strategy: extract full sequences + bigram sliding window (no single-char tokens)
// Example: "上海市旅游攻略" → ["上海市旅游攻略", "上海市", "海市旅", "市旅游", "旅游攻", "游攻略"]
func tokenize(text string) []string {
	text = strings.ToLower(text)

	termMap := make(map[string]bool)
	var terms []string

	asciiWords := asciiWordRE.FindAllString(text, -1)
	for _, word := range asciiWords {
		if len(word) >= 2 {
			if !termMap[word] {
				termMap[word] = true
				terms = append(terms, word)
			}
		}
	}

	cjkWords := cjkWordRE.FindAllString(text, -1)
	for _, word := range cjkWords {
		runes := []rune(word)
		runeLen := len(runes)

		if runeLen >= 2 && !termMap[word] {
			termMap[word] = true
			terms = append(terms, word)
		}

		if runeLen >= 3 {
			for i := 0; i+2 <= runeLen; i++ {
				bigram := string(runes[i : i+2])
				if !termMap[bigram] {
					termMap[bigram] = true
					terms = append(terms, bigram)
				}
			}
		}
	}

	return terms
}
