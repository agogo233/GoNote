package services

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"gonote/internal/models"
)

var ErrConflict = errors.New("note conflict: modified by another source")

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
	searchIndex    *SearchIndex   // 注入：performScan 增量同步搜索索引
	fileMtimes     sync.Map       // path → time.Time（已知 mtime 的 snapshot），用于增量同步
	pathMu         sync.Map       // path → *sync.Mutex（per-path 写入串行锁）
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

// SetSearchIndex 注入 SearchIndex 引用。用于在 performScan 中增量同步外部写入的笔记到搜索索引。
// 采用 setter 而非构造参数，避免与 SearchIndex 构造形成循环依赖（SearchIndex 构造也接收 NoteService）。
func (s *NoteService) SetSearchIndex(si *SearchIndex) {
	s.searchIndex = si
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

// getMu returns (and lazily creates) a per-path mutex for serializing writes.
func (s *NoteService) getMu(path string) *sync.Mutex {
	mu, _ := s.pathMu.LoadOrStore(path, &sync.Mutex{})
	return mu.(*sync.Mutex)
}

// SaveNote saves content to a note (backward-compatible: no optimistic lock).
func (s *NoteService) SaveNote(notePath, content string) error {
	return s.SaveNoteWithCheck(notePath, content, "")
}

// SaveNoteWithCheck saves content to a note with optional mtime-based optimistic lock.
// If knownMtime is non-empty and does not match the current file mtime, returns ErrConflict.
// The caller should pass an empty knownMtime to skip the check (backward-compatible).
func (s *NoteService) SaveNoteWithCheck(notePath, content, knownMtime string) error {
	if !strings.HasSuffix(notePath, ".md") {
		notePath += ".md"
	}

	fullPath := filepath.Join(s.notesDir, notePath)

	if !ValidatePathSecurity(s.notesDir, notePath) {
		return fmt.Errorf("invalid path")
	}

	mu := s.getMu(notePath)
	mu.Lock()
	defer mu.Unlock()

	// Optimistic lock: if knownMtime provided, check current file mtime.
	if knownMtime != "" {
		info, err := os.Stat(fullPath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("stat note: %w", err)
		}
		if err == nil {
			currentMtime := info.ModTime().UTC().Format(time.RFC3339Nano)
			if currentMtime != knownMtime {
				return ErrConflict
			}
		}
		// File does not exist yet (new note): skip check, proceed to create.
	}

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return err
	}

	s.invalidateNoteCache(notePath)

	return nil
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
// 顺序：先 rename（可逆性差但失败无副作用），成功后再改 backlinks；
// backlinks 改写失败时回滚 rename，避免出现"链接已改但文件没搬"的不可恢复状态。
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

	// 1) 先 rename：失败直接返回，无任何副作用
	if err := os.Rename(oldFull, newFull); err != nil {
		return err
	}

	// 2) rename 成功后改 backlinks；失败则回滚 rename
	backlinkService := NewBacklinkService(s.notesDir)
	if _, err := backlinkService.UpdateAllBacklinks(oldPath, newPath); err != nil {
		// 补偿：把文件搬回去，尽量恢复原状
		if rbErr := os.Rename(newFull, oldFull); rbErr != nil {
			return fmt.Errorf("move succeeded but backlink update failed (%v), and rollback also failed (%v)", err, rbErr)
		}
		return fmt.Errorf("move succeeded but backlink update failed, rolled back rename: %w", err)
	}

	// Fine-grained cache invalidation for both paths
	s.invalidateNoteCache(oldPath)
	s.invalidateNoteCache(newPath)

	return nil
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
		// 初始扫描：无论 panic 与否都关闭 ready，避免 WaitForReady 永久阻塞
		func() {
			defer close(s.ready) // 必须先于 recover 的 defer 注册，确保 panic 时仍执行
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Background scanner panic recovered: %v\n", r)
				}
			}()
			s.performScan()
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
	s.scanAndUpdate(true) // Include media
	notes := s.scanAndUpdate(false) // Exclude media（仅含 markdown）

	// 增量同步搜索索引：将扫描得到的 markdown 笔记集合与上次 snapshot 对比，
	// 对新增/变更调 UpdateIndex，对消失调 RemoveFromIndex。
	// 复用 scanAndUpdate 已读的 notes 列表，避免二次 stat。
	s.syncSearchIndex(notes)

	if s.onScanComplete != nil {
		s.onScanComplete()
	}
}

// syncSearchIndex 增量同步搜索索引。currentNotes 为本次扫描得到的 markdown 笔记列表
// （调用方应传入 includeMedia=false 的扫描结果）。基于 mtime 做差分，避免每次重建。
func (s *NoteService) syncSearchIndex(currentNotes []models.Note) {
	if s.searchIndex == nil {
		return
	}

	// 当前路径 → mtime，以及所有路径集合
	seen := make(map[string]time.Time, len(currentNotes))
	for _, n := range currentNotes {
		// doScan 只对 markdown 计算 tags；媒体文件 Type 为 mediaType。这里只需 markdown。
		if n.Type != "note" {
			continue
		}
		mtime, err := time.Parse(time.RFC3339, n.Modified)
		if err != nil {
			continue
		}
		seen[n.Path] = mtime
	}

	// 1) 新增/变更：相对 fileMtimes snapshot，mtime 不同或新路径
	for path, mtime := range seen {
		var prev time.Time
		if v, ok := s.fileMtimes.Load(path); ok {
			prev, _ = v.(time.Time)
		}
		if !prev.Equal(mtime) {
			if err := s.searchIndex.UpdateIndex(path); err == nil {
				s.fileMtimes.Store(path, mtime)
			}
		}
	}

	// 2) 消失：之前在 snapshot 现在 seen 中没有 → 移除
	s.fileMtimes.Range(func(key, _ any) bool {
		path, _ := key.(string)
		if _, ok := seen[path]; !ok {
			s.searchIndex.RemoveFromIndex(path)
			s.fileMtimes.Delete(path)
		}
		return true
	})
}

// scanAndUpdate performs the scan and updates the cache
func (s *NoteService) scanAndUpdate(includeMedia bool) []models.Note {
	notes, folders := s.doScan(includeMedia)

	cacheKey := fmt.Sprintf("%s:%v", cachePrefixNotesList, includeMedia)
	s.cache.Set(cacheKey, &scanCacheEntry{Notes: notes, Folders: folders})
	return notes
}

// WaitForReady blocks until the first scan completes.
// This ensures API requests have data available.
// If background scanner is not enabled (ready is nil), returns immediately.
// 带 30s 超时 fallback：即便 ready 未关闭也避免请求永久挂死
func (s *NoteService) WaitForReady() {
	if s.ready == nil {
		return
	}
	select {
	case <-s.ready:
	case <-time.After(30 * time.Second):
		fmt.Printf("Warn: WaitForReady timed out after 30s, serving possibly stale data\n")
	}
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

// ClearCache clears the note cache.
// Uses Clear() instead of replacing the pointer to avoid leaking the old cache's cleanup goroutine.
func (s *NoteService) ClearCache() {
	s.cache.Clear()
	s.tagCache.Clear()
}

