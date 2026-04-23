package services

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gonote/internal/models"
)

// Cache key prefixes for fine-grained invalidation
const (
	cachePrefixNotesList  = "notes:list"   // Note list cache
	cachePrefixContent    = "content:"     // Single note content
	cachePrefixTags       = "tags:"        // Single note tags
	cachePrefixScan       = "scan:"        // Scan results (legacy, kept for compatibility)
)

type NoteService struct {
	notesDir       string
	cache          *Cache
	tagCache       *Cache
	scanInterval   time.Duration
	stopScanner    chan struct{}
	scanTrigger    chan struct{}  // Trigger immediate scan
	ready          chan struct{}  // Closed when first scan completes
	onScanComplete func()         // Callback after scan completes (for WebSocket broadcast)
}

type scanCacheEntry struct {
	Notes   []models.Note
	Folders []string
}

type tagCacheEntry struct {
	ModifiedTime time.Time
	Tags         []string
}

func NewNoteService(notesDir string) *NoteService {
	return &NoteService{
		notesDir:     notesDir,
		cache:        NewCache(DefaultCapacity, DefaultTTL),
		tagCache:     NewCache(DefaultCapacity, DefaultTTL),
		scanInterval: 30 * time.Second,
		stopScanner:  nil, // No background scanner
		scanTrigger:  nil,
		ready:        nil, // nil means no background scanner, Skip WaitForReady
	}
}

func NewNoteServiceWithCache(notesDir string, cacheTTL time.Duration, cacheCapacity int) *NoteService {
	return &NoteService{
		notesDir:     notesDir,
		cache:        NewCache(cacheCapacity, cacheTTL),
		tagCache:     NewCache(cacheCapacity, cacheTTL),
		scanInterval: 30 * time.Second,
		stopScanner:  nil, // No background scanner
		scanTrigger:  nil,
		ready:        nil, // nil means no background scanner, Skip WaitForReady
	}
}

// NewNoteServiceWithScanner creates a NoteService with background scanner enabled
func NewNoteServiceWithScanner(notesDir string, cacheTTL time.Duration, cacheCapacity int, scanInterval time.Duration) *NoteService {
	return &NoteService{
		notesDir:     notesDir,
		cache:        NewCache(cacheCapacity, cacheTTL),
		tagCache:     NewCache(cacheCapacity, cacheTTL),
		scanInterval: scanInterval,
		stopScanner:  make(chan struct{}),
		scanTrigger:  make(chan struct{}, 1),
		ready:        make(chan struct{}),
	}
}

// SetOnScanComplete sets a callback function to be called after each scan completes
func (s *NoteService) SetOnScanComplete(callback func()) {
	s.onScanComplete = callback
}

// ScanNotes scans all notes in the notes directory.
// If background scanner is enabled, it waits for the first scan to complete
// and then reads from cache. Falls back to direct scan if scanner is not running.
func (s *NoteService) ScanNotes(includeMedia bool) ([]models.Note, []string, error) {
	// If background scanner is enabled, wait for first scan and read from cache
	if s.ready != nil {
		s.WaitForReady()
	}

	// Use simplified cache key with prefix
	cacheKey := fmt.Sprintf("%s:%v", cachePrefixNotesList, includeMedia)

	if val, ok := s.cache.Get(cacheKey); ok {
		if entry, ok := val.(*scanCacheEntry); ok {
			return entry.Notes, entry.Folders, nil
		}
	}

	// Try to derive from media cache if not including media
	if !includeMedia {
		mediaCacheKey := fmt.Sprintf("%s:%v", cachePrefixNotesList, true)
		if val, ok := s.cache.Get(mediaCacheKey); ok {
			if entry, ok := val.(*scanCacheEntry); ok {
				var notes []models.Note
				for _, note := range entry.Notes {
					if note.Type == "note" {
						notes = append(notes, note)
					}
				}
				result := &scanCacheEntry{Notes: notes, Folders: entry.Folders}
				s.cache.Set(cacheKey, result)
				return notes, entry.Folders, nil
			}
		}
	}

	// Fallback: direct scan if cache is empty (should not happen with background scanner)
	return s.directScan(includeMedia)
}

// doScan performs the actual scan logic and returns notes and folders
func (s *NoteService) doScan(includeMedia bool) ([]models.Note, []string) {
	var notes []models.Note
	foldersSet := make(map[string]bool)
	var walkErrors []string

	filepath.WalkDir(s.notesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Log walk errors instead of silently ignoring
			walkErrors = append(walkErrors, fmt.Sprintf("WalkDir error at %s: %v", path, err))
			return nil // Continue walking despite errors
		}

		// Skip dot directories
		if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
			return fs.SkipDir
		}

		// Skip dot files
		if strings.HasPrefix(d.Name(), ".") {
			return nil
		}

		relPath, err := filepath.Rel(s.notesDir, path)
		if err != nil {
			return nil
		}

		if d.IsDir() {
			// Add folder
			if relPath != "." {
				foldersSet[ToPosixPath(relPath)] = true
			}
			return nil
		}

		// Process files
		ext := strings.ToLower(filepath.Ext(path))
		isMarkdown := ext == ".md"
		mediaType := GetMediaType(d.Name())
		shouldInclude := isMarkdown || (includeMedia && mediaType != "")

		if !shouldInclude {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		folder := ToPosixPath(filepath.Dir(relPath))
		if folder == "." {
			folder = ""
		}

		// Get tags for markdown files
		var tags []string
		if isMarkdown {
			tags = s.GetTagsCached(path)
		}

		noteType := "note"
		if mediaType != "" {
			noteType = mediaType
		}

		notes = append(notes, models.Note{
			Name:     strings.TrimSuffix(d.Name(), filepath.Ext(d.Name())),
			Path:     ToPosixPath(relPath),
			Folder:   folder,
			Modified: info.ModTime().UTC().Format(time.RFC3339),
			Size:     info.Size(),
			Type:     noteType,
			Tags:     tags,
		})

		return nil
	})

	// Log walk errors if any
	if len(walkErrors) > 0 {
		for _, errMsg := range walkErrors[:min(len(walkErrors), 10)] {
			fmt.Printf("Warning: %s\n", errMsg)
		}
		if len(walkErrors) > 10 {
			fmt.Printf("Warning: ... and %d more walk errors\n", len(walkErrors)-10)
		}
	}

	// Sort notes by modified date (newest first)
	sort.Slice(notes, func(i, j int) bool {
		return notes[i].Modified > notes[j].Modified
	})

	// Convert folders set to sorted slice
	folders := make([]string, 0, len(foldersSet))
	for f := range foldersSet {
		folders = append(folders, f)
	}
	sort.Strings(folders)

	return notes, folders
}

// directScan performs a direct scan without using cache (fallback)
func (s *NoteService) directScan(includeMedia bool) ([]models.Note, []string, error) {
	notes, folders := s.doScan(includeMedia)

	// Cache the result
	cacheKey := fmt.Sprintf("%s:%v", cachePrefixNotesList, includeMedia)
	s.cache.Set(cacheKey, &scanCacheEntry{Notes: notes, Folders: folders})

	return notes, folders, nil
}

// GetNoteContent returns the content of a note
func (s *NoteService) GetNoteContent(notePath string) (string, error) {
	fullPath := filepath.Join(s.notesDir, notePath)

	// Security check
	if !ValidatePathSecurity(s.notesDir, notePath) {
		return "", fmt.Errorf("invalid path")
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// SaveNote saves content to a note
func (s *NoteService) SaveNote(notePath, content string) error {
	if !strings.HasSuffix(notePath, ".md") {
		notePath += ".md"
	}

	fullPath := filepath.Join(s.notesDir, notePath)

	if !ValidatePathSecurity(s.notesDir, notePath) {
		return fmt.Errorf("invalid path")
	}

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}

	// Fine-grained cache invalidation: only invalidate related entries
	s.invalidateNoteCache(notePath)

	return os.WriteFile(fullPath, []byte(content), 0644)
}

// DeleteNote deletes a note
func (s *NoteService) DeleteNote(notePath string) error {
	fullPath := filepath.Join(s.notesDir, notePath)

	if !ValidatePathSecurity(s.notesDir, notePath) {
		return fmt.Errorf("invalid path")
	}

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("note not found")
	}

	// Fine-grained cache invalidation
	s.invalidateNoteCache(notePath)

	return os.Remove(fullPath)
}

// MoveNote moves a note to a new location and updates all wikilink references
func (s *NoteService) MoveNote(oldPath, newPath string) error {
	oldFull := filepath.Join(s.notesDir, oldPath)
	newFull := filepath.Join(s.notesDir, newPath)

	if !ValidatePathSecurity(s.notesDir, oldPath) || !ValidatePathSecurity(s.notesDir, newPath) {
		return fmt.Errorf("invalid path")
	}

	if _, err := os.Stat(oldFull); os.IsNotExist(err) {
		return fmt.Errorf("source note does not exist")
	}

	if _, err := os.Stat(newFull); err == nil {
		return fmt.Errorf("destination already exists")
	}

	if err := os.MkdirAll(filepath.Dir(newFull), 0755); err != nil {
		return err
	}

	// Update all wikilink references before moving
	backlinkService := NewBacklinkService(s.notesDir)
	backlinkService.UpdateAllBacklinks(oldPath, newPath)

	// Fine-grained cache invalidation for both paths
	s.invalidateNoteCache(oldPath)
	s.invalidateNoteCache(newPath)

	return os.Rename(oldFull, newFull)
}

// NoteExists checks if a note exists
func (s *NoteService) NoteExists(notePath string) bool {
	fullPath := filepath.Join(s.notesDir, notePath)
	_, err := os.Stat(fullPath)
	return err == nil
}

// GetNoteMetadata returns metadata for a note
func (s *NoteService) GetNoteMetadata(notePath string) (*models.NoteMetadata, error) {
	fullPath := filepath.Join(s.notesDir, notePath)

	// Security check
	if !ValidatePathSecurity(s.notesDir, notePath) {
		return nil, fmt.Errorf("invalid path")
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}

	// Count lines
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}
	// Handle empty file case: 0 bytes = 0 lines
	var lineCount int
	if len(content) == 0 {
		lineCount = 0
	} else {
		lineCount = bytes.Count(content, []byte("\n")) + 1
	}

	return &models.NoteMetadata{
		Created:  info.ModTime().UTC().Format(time.RFC3339), // Using ModTime as creation time
		Modified: info.ModTime().UTC().Format(time.RFC3339),
		Size:     info.Size(),
		Lines:    lineCount,
	}, nil
}

// GetAttachmentDir returns the attachments directory for a note
func (s *NoteService) GetAttachmentDir(notePath string) string {
	if notePath == "" {
		return filepath.Join(s.notesDir, "_attachments")
	}

	folder := filepath.Dir(notePath)
	if folder == "." {
		return filepath.Join(s.notesDir, "_attachments")
	}

	return filepath.Join(s.notesDir, folder, "_attachments")
}

func (s *NoteService) GetAttachments(notePath string) ([]models.Attachment, error) {
	// Read note content to extract referenced media
	content, err := s.GetNoteContent(notePath)
	if err != nil {
		// If note doesn't exist or can't be read, return empty
		return []models.Attachment{}, nil
	}

	// Extract media references from content: ![[filename]] or ![[filename|alt text]]
	// This regex matches the wikilink format for embedded media
	mediaRefRegex := regexp.MustCompile(`!\[\[([^\]|]+)(?:\|[^\]]+)?\]\]`)
	matches := mediaRefRegex.FindAllStringSubmatch(content, -1)

	// Build a set of referenced media filenames
	referencedMedia := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			filename := strings.TrimSpace(match[1])
			referencedMedia[filename] = true
		}
	}

	// If no media references, return empty
	if len(referencedMedia) == 0 {
		return []models.Attachment{}, nil
	}

	// Get attachment directory
	attachmentsDir := s.GetAttachmentDir(notePath)

	if _, err := os.Stat(attachmentsDir); os.IsNotExist(err) {
		return []models.Attachment{}, nil
	}

	// Walk the attachments directory and find referenced files
	var attachments []models.Attachment

	err = filepath.Walk(attachmentsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		filename := info.Name()

		// Check if this file is referenced in the note
		if referencedMedia[filename] {
			relPath, err := filepath.Rel(s.notesDir, path)
			if err != nil {
				return err
			}

			attachments = append(attachments, models.Attachment{
				Name: filename,
				Path: filepath.ToSlash(relPath),
				Size: info.Size(),
				Type: GetMediaType(filename),
			})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return attachments, nil
}

// SaveUploadedImage saves an uploaded image to the attachments directory
func (s *NoteService) SaveUploadedImage(notePath, filename string, data []byte) (string, error) {
	// Sanitize filename
	sanitizedName := SanitizeFilename(filename)

	// Get extension and name
	ext := filepath.Ext(sanitizedName)
	name := strings.TrimSuffix(sanitizedName, ext)

	// Add timestamp to prevent collisions
	timestamp := time.Now().Format("20060102150405")
	finalFilename := fmt.Sprintf("%s-%s%s", name, timestamp, ext)

	// Get attachments directory
	attachmentsDir := s.GetAttachmentDir(notePath)

	// Create directory if needed
	if err := os.MkdirAll(attachmentsDir, 0755); err != nil {
		return "", err
	}

	fullPath := filepath.Join(attachmentsDir, finalFilename)

	// Security check
	absPath, _ := filepath.Abs(fullPath)
	absNotesDir, _ := filepath.Abs(s.notesDir)
	if !strings.HasPrefix(absPath, absNotesDir) {
		return "", fmt.Errorf("invalid path")
	}

	// Write file
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return "", err
	}

	// Return relative path
	relPath, _ := filepath.Rel(s.notesDir, fullPath)
	return ToPosixPath(relPath), nil
}

// invalidateNoteCache invalidates cache entries related to a specific note
// This is more efficient than clearing the entire cache
func (s *NoteService) invalidateNoteCache(notePath string) {
	// Normalize path
	notePath = ToPosixPath(notePath)
	
	// Invalidate note list cache (modification time changes)
	s.cache.DeleteByPrefix(cachePrefixNotesList)
	
	// Invalidate content cache for this note
	s.cache.Delete(cachePrefixContent + notePath)
	
	// Invalidate tag cache for this note
	s.tagCache.Delete(cachePrefixTags + notePath)
	
	// Also invalidate legacy scan cache keys for compatibility
	s.cache.DeleteByPrefix(cachePrefixScan)
	
	// Trigger immediate background scan to update cache
	s.TriggerScan()
}

func (s *NoteService) InvalidateCache() {
	s.cache.Clear()
	s.tagCache.Clear()
}

func (s *NoteService) GetTagsCached(filePath string) []string {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil
	}

	cacheKey := fmt.Sprintf("tags:%s", filePath)
	if val, ok := s.tagCache.Get(cacheKey); ok {
		if entry, ok := val.(*tagCacheEntry); ok {
			if entry.ModifiedTime.Equal(info.ModTime()) {
				return entry.Tags
			}
		}
	}

	// Cache miss - parse tags from file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}

	tags := ParseTags(string(content))

	// Cache the result
	s.tagCache.Set(cacheKey, &tagCacheEntry{
		ModifiedTime: info.ModTime(),
		Tags:         tags,
	})

	return tags
}

func (s *NoteService) SetTagsCached(filePath string, mtime time.Time, tags []string) {
	cacheKey := fmt.Sprintf("tags:%s", filePath)
	s.tagCache.Set(cacheKey, &tagCacheEntry{
		ModifiedTime: mtime,
		Tags:         tags,
	})
}

// StartCacheCleanup starts the background cleanup goroutines for both caches.
// This should be called once during application startup to prevent memory leaks
// from expired cache entries.
func (s *NoteService) StartCacheCleanup() {
	s.cache.StartCleanup()
	s.tagCache.StartCleanup()
}

// StopCacheCleanup stops the background cleanup goroutines.
// This should be called during application shutdown.
func (s *NoteService) StopCacheCleanup() {
	s.cache.StopCleanup()
	s.tagCache.StopCleanup()
}

// StartBackgroundScanner starts the background note scanner.
// It performs an initial scan immediately, then scans periodically.
// API requests will block until the first scan completes.
// If a panic occurs, the scanner will automatically restart after a delay.
func (s *NoteService) StartBackgroundScanner() {
	go func() {
		// Initial scan with panic recovery
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Background scanner panic recovered: %v\n", r)
				}
			}()
			s.performScan()
			close(s.ready) // Signal that first scan is complete
		}()

		ticker := time.NewTicker(s.scanInterval)
		defer ticker.Stop()

		for {
			select {
			case <-s.stopScanner:
				return
			case <-s.scanTrigger:
				// Immediate scan with panic recovery
				func() {
					defer func() {
						if r := recover(); r != nil {
							fmt.Printf("Background scanner panic recovered during scheduled scan: %v\n", r)
							// Continue running - don't let one panic stop the scanner
						}
					}()
					s.performScan()
				}()
			case <-ticker.C:
				// Periodic scan with panic recovery
				func() {
					defer func() {
						if r := recover(); r != nil {
							fmt.Printf("Background scanner panic recovered during scheduled scan: %v\n", r)
							// Continue running - don't let one panic stop the scanner
						}
					}()
					s.performScan()
				}()
			}
		}
	}()
}

// StopBackgroundScanner stops the background scanner.
// This should be called during application shutdown.
// If background scanner is not enabled, this is a no-op.
func (s *NoteService) StopBackgroundScanner() {
	if s.stopScanner == nil {
		return
	}
	close(s.stopScanner)
}

// TriggerScan triggers an immediate background scan.
// This is called after user operations (create/delete/move) to update cache.
// If background scanner is not enabled, this is a no-op.
func (s *NoteService) TriggerScan() {
	if s.scanTrigger == nil {
		return
	}
	select {
	case s.scanTrigger <- struct{}{}:
		// Scan triggered
	default:
		// Scan already pending, skip
	}
}

// performScan executes the actual scan and updates cache
func (s *NoteService) performScan() {
	// Scan both with and without media to populate both cache entries
	s.scanAndUpdate(true)  // Include media
	s.scanAndUpdate(false) // Exclude media

	// Notify listeners (WebSocket broadcast)
	if s.onScanComplete != nil {
		s.onScanComplete()
	}
}

// scanAndUpdate performs the scan and updates the cache
func (s *NoteService) scanAndUpdate(includeMedia bool) {
	notes, folders := s.doScan(includeMedia)

	cacheKey := fmt.Sprintf("%s:%v", cachePrefixNotesList, includeMedia)
	s.cache.Set(cacheKey, &scanCacheEntry{Notes: notes, Folders: folders})
}

// WaitForReady blocks until the first scan completes.
// This ensures API requests have data available.
// If background scanner is not enabled (ready is nil), returns immediately.
func (s *NoteService) WaitForReady() {
	if s.ready == nil {
		return
	}
	<-s.ready
}

// IsReady returns true if the first scan has completed.
// If background scanner is not enabled (ready is nil), returns true immediately.
func (s *NoteService) IsReady() bool {
	if s.ready == nil {
		return true // No background scanner, always ready
	}
	select {
	case <-s.ready:
		return true
	default:
		return false
	}
}

// ClearCache clears the note cache
func (s *NoteService) ClearCache() {
	s.cache = NewCache(DefaultCapacity, DefaultTTL)
}

