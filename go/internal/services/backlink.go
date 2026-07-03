package services

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Pre-compiled regex patterns for performance
var (
	backlinkWikilinkRegex = regexp.MustCompile(`\[\[([^\]|]+)(?:\|[^\]]+)?\]\]`)
)

// BacklinkService handles backlink discovery and wikilink updates
type BacklinkService struct {
	notesDir string
}

// NewBacklinkService creates a new BacklinkService
func NewBacklinkService(notesDir string) *BacklinkService {
	return &BacklinkService{notesDir: notesDir}
}

// BacklinkLink contains information about a single backlink occurrence
type BacklinkLink struct {
	Line    int    `json:"line"`     // Line number (1-indexed)
	Context string `json:"context"`  // Context snippet around the link
	Type    string `json:"type"`     // Link type: "wikilink" or "markdown"
}

// BacklinkInfo contains information about a backlink
type BacklinkInfo struct {
	SourcePath string         `json:"source_path"` // Path of the note containing the link
	LinkTexts  []string       `json:"link_texts"`  // The link texts found (e.g., "note-name", "folder/note")
	Links      []BacklinkLink `json:"links"`       // Individual link occurrences with line numbers and context
}

// FindBacklinks finds all notes that link to the specified note
func (s *BacklinkService) FindBacklinks(notePath string) ([]BacklinkInfo, error) {
	var backlinks []BacklinkInfo

	// Get the note name and path variations to search for
	noteName := strings.TrimSuffix(filepath.Base(notePath), ".md")
	notePathWithoutExt := strings.TrimSuffix(notePath, ".md")

	// Walk through all notes
	err := filepath.WalkDir(s.notesDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		// Skip directories and non-markdown files
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Skip the note itself
		relPath, err := filepath.Rel(s.notesDir, path)
		if err != nil {
			return nil
		}
		if ToPosixPath(relPath) == ToPosixPath(notePath) {
			return nil
		}

		// Read note content
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		// Find wikilinks to this note with line numbers and context
		linkTexts, links := s.extractWikilinksToNoteWithLines(string(content), noteName, notePathWithoutExt)
		if len(links) > 0 {
			backlinks = append(backlinks, BacklinkInfo{
				SourcePath: ToPosixPath(relPath),
				LinkTexts:  linkTexts,
				Links:      links,
			})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return backlinks, nil
}

// extractWikilinksToNoteWithLines extracts wikilinks with line numbers and context
func (s *BacklinkService) extractWikilinksToNoteWithLines(content, noteName, notePathWithoutExt string) ([]string, []BacklinkLink) {
	var linkTexts []string
	var links []BacklinkLink
	seenTexts := make(map[string]bool)

	lines := strings.Split(content, "\n")

	for lineNum, line := range lines {
		// Match wikilinks: [[link]] or [[link|display text]]
		matches := backlinkWikilinkRegex.FindAllStringSubmatchIndex(line, -1)

		for _, match := range matches {
			linkText := strings.TrimSpace(line[match[2]:match[3]])

			// Remove anchor if present
			if idx := strings.Index(linkText, "#"); idx != -1 {
				linkText = linkText[:idx]
			}

			// Check if this link references our note
			if s.linkMatchesNote(linkText, noteName, notePathWithoutExt) {
				// Add to link texts (unique)
				if !seenTexts[linkText] {
					linkTexts = append(linkTexts, linkText)
					seenTexts[linkText] = true
				}

				// Create context snippet (30 chars before and after)
				start := match[0] - 30
				end := match[1] + 30
				if start < 0 {
					start = 0
				}
				if end > len(line) {
					end = len(line)
				}
				context := line[start:end]
				if start > 0 {
					context = "..." + context
				}
				if end < len(line) {
					context = context + "..."
				}

				links = append(links, BacklinkLink{
					Line:    lineNum + 1, // 1-indexed
					Context: context,
					Type:    "wikilink",
				})
			}
		}
	}

	return linkTexts, links
}

// linkMatchesNote checks if a wikilink text references a specific note
func (s *BacklinkService) linkMatchesNote(linkText, noteName, notePathWithoutExt string) bool {
	// Exact match with note name
	if linkText == noteName {
		return true
	}

	// Exact match with path (without extension)
	if linkText == notePathWithoutExt {
		return true
	}

	// Check if link ends with /noteName
	if strings.HasSuffix(linkText, "/"+noteName) {
		return true
	}

	// Check if the note name matches (case-insensitive)
	if strings.EqualFold(linkText, noteName) {
		return true
	}

	return false
}

// UpdateWikilinks updates wikilinks in a note from old reference to new reference
func (s *BacklinkService) UpdateWikilinks(sourcePath, oldLinkText, newLinkText string) error {
	fullPath := filepath.Join(s.notesDir, sourcePath)

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return err
	}

	// Replace wikilinks: [[old]] -> [[new]] and [[old|text]] -> [[new|text]]
	updated := s.replaceWikilink(string(content), oldLinkText, newLinkText)

	if updated == string(content) {
		return nil // No changes
	}

	return os.WriteFile(fullPath, []byte(updated), 0644)
}

// replaceWikilink replaces wikilink references in content
func (s *BacklinkService) replaceWikilink(content, oldLinkText, newLinkText string) string {
	// Pattern to match wikilinks with optional display text
	// [[old]] or [[old|display text]]
	pattern := regexp.MustCompile(`\[\[` + regexp.QuoteMeta(oldLinkText) + `(\|[^\]]+)?\]\]`)
	
	// Replace with new link text
	result := pattern.ReplaceAllStringFunc(content, func(match string) string {
		// Check if there's a display text
		if strings.Contains(match, "|") {
			// Extract display text
			idx := strings.Index(match, "|")
			displayText := match[idx+1 : len(match)-2] // Remove "]]"
			return "[[" + newLinkText + "|" + displayText + "]]"
		}
		return "[[" + newLinkText + "]]"
	})

	// Also replace if the old link appears as a path suffix
	// e.g., [[folder/oldname]] when moving to [[folder/newname]]
	if !strings.Contains(oldLinkText, "/") && !strings.Contains(newLinkText, "/") {
		// Both are simple names, also check for path-based references
		pathPattern := regexp.MustCompile(`\[\[([^\]|]*/)?` + regexp.QuoteMeta(oldLinkText) + `(\|[^\]]+)?\]\]`)
		result = pathPattern.ReplaceAllStringFunc(result, func(match string) string {
			// Extract prefix path if any
			start := strings.Index(match, "[[") + 2
			end := strings.Index(match, "]]")
			inner := match[start:end]
			
			var prefix, displayText string
			if idx := strings.Index(inner, "|"); idx != -1 {
				displayText = inner[idx:]
				inner = inner[:idx]
			}
			if slashIdx := strings.LastIndex(inner, "/"); slashIdx != -1 {
				prefix = inner[:slashIdx+1]
			}
			
			// Replace old name with new name, keeping prefix
			return "[[" + prefix + newLinkText + displayText + "]]"
		})
	}

	return result
}

// UpdateAllBacklinks updates all notes that reference the moved/renamed note
func (s *BacklinkService) UpdateAllBacklinks(oldPath, newPath string) (int, error) {
	// Find all backlinks
	backlinks, err := s.FindBacklinks(oldPath)
	if err != nil {
		return 0, err
	}

	// Get old and new link texts
	oldNoteName := strings.TrimSuffix(filepath.Base(oldPath), ".md")
	newNoteName := strings.TrimSuffix(filepath.Base(newPath), ".md")
	oldPathWithoutExt := strings.TrimSuffix(oldPath, ".md")
	newPathWithoutExt := strings.TrimSuffix(newPath, ".md")

	updatedCount := 0

	for _, backlink := range backlinks {
		needsUpdate := false
		
		// Check each link text
		for _, linkText := range backlink.LinkTexts {
			// Determine the replacement text
			var newLinkText string
			
			if linkText == oldNoteName || linkText == strings.ToLower(oldNoteName) {
				// Simple name reference -> update to new simple name
				newLinkText = newNoteName
				needsUpdate = true
			} else if linkText == oldPathWithoutExt {
				// Full path reference -> update to new path
				newLinkText = newPathWithoutExt
				needsUpdate = true
			} else if strings.HasSuffix(linkText, "/"+oldNoteName) {
				// Path with note name -> update the note name part
				prefix := linkText[:strings.LastIndex(linkText, "/")+1]
				newLinkText = prefix + newNoteName
				needsUpdate = true
			}

			if needsUpdate {
				if err := s.UpdateWikilinks(backlink.SourcePath, linkText, newLinkText); err != nil {
					continue // Skip on error
				}
				updatedCount++
				break // Only count once per file
			}
		}
	}

	return updatedCount, nil
}

// CountBacklinks returns the number of notes that link to the specified note
func (s *BacklinkService) CountBacklinks(notePath string) (int, error) {
	backlinks, err := s.FindBacklinks(notePath)
	if err != nil {
		return 0, err
	}
	return len(backlinks), nil
}

// UpdateFolderBacklinks updates all wikilinks when a folder is moved/renamed
// This scans all notes and updates any wikilinks pointing to notes inside the moved folder
func (s *BacklinkService) UpdateFolderBacklinks(oldFolderPath, newFolderPath string) (int, error) {
	// Normalize paths
	oldFolderPath = strings.Trim(oldFolderPath, "/")
	newFolderPath = strings.Trim(newFolderPath, "/")

	updatedCount := 0

	// Walk through all notes
	err := filepath.WalkDir(s.notesDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		// Skip directories and non-markdown files
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Read note content
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		contentStr := string(content)
		updated := s.replaceFolderWikilinks(contentStr, oldFolderPath, newFolderPath)

		// If content changed, write back
		if updated != contentStr {
			if err := os.WriteFile(path, []byte(updated), 0644); err == nil {
				updatedCount++
			}
		}

		return nil
	})

	return updatedCount, err
}

// replaceFolderWikilinks replaces wikilinks pointing to notes in old folder to new folder
func (s *BacklinkService) replaceFolderWikilinks(content, oldFolder, newFolder string) string {
	// Pattern to match wikilinks that might reference the old folder
	wikilinkPattern := backlinkWikilinkRegex

	result := wikilinkPattern.ReplaceAllStringFunc(content, func(match string) string {
		// Extract the link target
		start := strings.Index(match, "[[") + 2
		end := strings.Index(match, "]]")
		inner := match[start:end]

		var linkTarget, displayText string
		if idx := strings.Index(inner, "|"); idx != -1 {
			linkTarget = inner[:idx]
			displayText = inner[idx:]
		} else {
			linkTarget = inner
		}

		linkTarget = strings.TrimSpace(linkTarget)

		// Remove anchor for checking
		anchor := ""
		if idx := strings.Index(linkTarget, "#"); idx != -1 {
			anchor = linkTarget[idx:]
			linkTarget = linkTarget[:idx]
		}

		// Check if this link starts with the old folder path
		if oldFolder != "" && strings.HasPrefix(linkTarget, oldFolder+"/") {
			// Replace old folder with new folder
			newTarget := newFolder + strings.TrimPrefix(linkTarget, oldFolder)
			return "[[" + newTarget + anchor + displayText + "]]"
		}

		// Also handle case where oldFolder is empty (moving to root)
		if oldFolder == "" && newFolder != "" {
			// This would be moving from root to a folder, less common
			// For now, we don't handle this case
		}

		return match
	})

	return result
}
