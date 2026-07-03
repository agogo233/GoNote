package services

import (
	"sort"
	"strings"

	"gonote/internal/models"
)

// TagService handles tag-related operations
// Tag caching is delegated to NoteService to avoid duplicate caches (I-3).
type TagService struct {
	notesDir   string
	noteService *NoteService  // Shared NoteService instance
}

// NewTagService creates a new TagService
func NewTagService(noteService *NoteService, notesDir string) *TagService {
	return &TagService{
		notesDir:    notesDir,
		noteService: noteService,
	}
}

// ParseTags extracts tags from YAML frontmatter
func ParseTags(content string) []string {
	tags := []string{}

	// Check if content starts with frontmatter
	if !strings.HasPrefix(strings.TrimSpace(content), "---") {
		return tags
	}

	lines := strings.Split(content, "\n")
	if len(lines) < 2 || strings.TrimSpace(lines[0]) != "---" {
		return tags
	}

	// Find closing ---
	endIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			endIdx = i
			break
		}
	}

	if endIdx == -1 {
		return tags
	}

	// Parse tags field
	inTagsList := false
	for i := 1; i < endIdx; i++ {
		line := strings.TrimSpace(lines[i])

		if strings.HasPrefix(line, "tags:") {
			rest := strings.TrimSpace(line[5:])

			// Inline array format: tags: [tag1, tag2, tag3]
			if strings.HasPrefix(rest, "[") && strings.HasSuffix(rest, "]") {
				tagsStr := rest[1 : len(rest)-1]
				for _, t := range strings.Split(tagsStr, ",") {
					t = strings.TrimSpace(t)
					if t != "" {
						tags = append(tags, strings.ToLower(t))
					}
				}
				break
			} else if rest != "" {
				// Single tag
				tags = append(tags, strings.ToLower(rest))
				break
			} else {
				// Multi-line list format
				inTagsList = true
			}
		} else if inTagsList {
			if strings.HasPrefix(line, "-") {
				// List item
				tag := strings.TrimSpace(line[1:])
				if tag != "" {
					tags = append(tags, strings.ToLower(tag))
				}
			} else if line != "" && !strings.HasPrefix(line, "#") {
				// End of tags list
				break
			}
		}
	}

	return sortUniqueTags(tags)
}

// sortUniqueTags removes duplicates and sorts tags
func sortUniqueTags(tags []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, tag := range tags {
		if !seen[tag] {
			seen[tag] = true
			result = append(result, tag)
		}
	}

	sort.Strings(result)
	return result
}

// GetAllTags returns all tags with their counts
func (s *TagService) GetAllTags() (map[string]int, error) {
	tagCounts := make(map[string]int)

	notes, _, err := s.scanNotes()
	if err != nil {
		return nil, err
	}

	for _, note := range notes {
		// Use tags already parsed by NoteService.ScanNotes
		for _, tag := range note.Tags {
			tagCounts[tag]++
		}
	}

	return tagCounts, nil
}

// GetNotesByTag returns all notes with a specific tag
func (s *TagService) GetNotesByTag(tag string) ([]models.Note, error) {
	tagLower := strings.ToLower(tag)
	var matchingNotes []models.Note

	notes, _, err := s.scanNotes()
	if err != nil {
		return nil, err
	}

	for _, note := range notes {
		// Use tags already parsed by NoteService.ScanNotes
		for _, t := range note.Tags {
			if t == tagLower {
				matchingNotes = append(matchingNotes, note)
				break
			}
		}
	}

	return matchingNotes, nil
}

// scanNotes is a helper to scan notes
func (s *TagService) scanNotes() ([]models.Note, []string, error) {
	// Use shared NoteService for scanning to ensure cache consistency
	return s.noteService.ScanNotes(false)
}

// GetTagsCached returns tags for a file with caching
// Delegates to NoteService's unified tag cache (I-3).
func (s *TagService) GetTagsCached(filePath string) []string {
	return s.noteService.GetTagsCached(filePath)
}

// ClearCache clears the tag cache
// Delegates to NoteService's unified invalidation (I-3).
func (s *TagService) ClearCache() {
	s.noteService.InvalidateCache()
}

// FilterNotesByTags filters notes that contain ALL specified tags (AND logic)
// Returns notes sorted by modified date (newest first)
func FilterNotesByTags(notes []models.Note, tags []string) []models.Note {
	if len(tags) == 0 {
		return notes
	}

	// Convert tags to lowercase for case-insensitive matching
	tagSet := make(map[string]bool)
	for _, t := range tags {
		tagSet[strings.ToLower(strings.TrimSpace(t))] = true
	}

	var result []models.Note
	for _, note := range notes {
		// Check if note has ALL specified tags
		if noteHasAllTags(note, tagSet) {
			result = append(result, note)
		}
	}

	return result
}

// noteHasAllTags checks if a note contains all tags in the set (AND logic)
func noteHasAllTags(note models.Note, requiredTags map[string]bool) bool {
	// If no tags required, match all
	if len(requiredTags) == 0 {
		return true
	}

	// If note has no tags, can't match
	if len(note.Tags) == 0 {
		return false
	}

	// Check each required tag
	for requiredTag := range requiredTags {
		found := false
		for _, noteTag := range note.Tags {
			if noteTag == requiredTag {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}
