package services

import (
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gonote/internal/models/logger"
)


// MediaExtensions defines supported media file extensions
var MediaExtensions = map[string][]string{
	"image":    {".jpg", ".jpeg", ".png", ".gif", ".webp"},
	"audio":    {".mp3", ".wav", ".ogg", ".m4a"},
	"video":    {".mp4", ".webm", ".mov", ".avi"},
	"document": {".pdf"},
}

// AllMediaExtensions is a flat set for quick lookup
var AllMediaExtensions = make(map[string]bool)

func init() {
	for _, exts := range MediaExtensions {
		for _, ext := range exts {
			AllMediaExtensions[ext] = true
		}
	}
}



// Regex patterns for extracting media references from markdown
var (
	// markdownImageRegex matches markdown image syntax: ![alt](path)
	markdownImageRegex = regexp.MustCompile(`!\[.*?\]\((.*?)\)`)

	// wikilinkMediaRegex matches wikilinks with media extensions: [[path.ext]]
	wikilinkMediaRegex = regexp.MustCompile(`\[\[(.*?\.(?:jpg|jpeg|png|gif|webp|mp3|wav|ogg|m4a|mp4|webm|mov|avi|pdf))\]\]`)

	// mediaWikilinkRegex matches media embed wikilinks: ![[file.png]] or ![[file.png|alt text]]
	// This is the primary format used by the frontend for embedding media
	mediaWikilinkRegex = regexp.MustCompile(`!\[\[([^\]|]+\.(?:jpg|jpeg|png|gif|webp|mp3|wav|ogg|m4a|mp4|webm|mov|avi|pdf))(?:\|[^\]]+)?\]\]`)

	// htmlImgSrcRegex matches HTML img tags: <img src="path">
	htmlImgSrcRegex = regexp.MustCompile(`<img[^>]+src=["']([^"']+)["']`)

	// htmlVideoSrcRegex matches HTML video tags: <video src="path">
	htmlVideoSrcRegex = regexp.MustCompile(`<video[^>]+src=["']([^"']+)["']`)

	// htmlAudioSrcRegex matches HTML audio tags: <audio src="path">
	htmlAudioSrcRegex = regexp.MustCompile(`<audio[^>]+src=["']([^"']+)["']`)

	// htmlSourceSrcRegex matches HTML source tags: <source src="path">
	htmlSourceSrcRegex = regexp.MustCompile(`<source[^>]+src=["']([^"']+)["']`)
)

// GetMediaType determines the media type based on file extension
func GetMediaType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	for mediaType, exts := range MediaExtensions {
		for _, e := range exts {
			if e == ext {
				return mediaType
			}
		}
	}

	return ""
}

// IsMediaFile checks if a file is a supported media type
func IsMediaFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return AllMediaExtensions[ext]
}

// IsImageFile checks if a file is an image
func IsImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	for _, e := range MediaExtensions["image"] {
		if e == ext {
			return true
		}
	}
	return false
}

// MediaService handles media operations
type MediaService struct {
	notesDir string
}

// NewMediaService creates a new MediaService
func NewMediaService(notesDir string) *MediaService {
	return &MediaService{notesDir: notesDir}
}

// GetMedia returns the content of a media file
func (s *MediaService) GetMedia(mediaPath string) ([]byte, string, error) {
	fullPath := filepath.Join(s.notesDir, mediaPath)

	// Security check
	if !ValidatePathSecurity(s.notesDir, mediaPath) {
		return nil, "", ErrInvalidPath
	}

	// Check if exists and read
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, "", err
	}

	// Determine content type
	contentType := GetFileContentType(mediaPath)

	return data, contentType, nil
}

// UploadMedia saves an uploaded media file
func (s *MediaService) UploadMedia(notePath, filename string, data []byte) (string, error) {
	ns := NewNoteService(s.notesDir)
	return ns.SaveUploadedImage(notePath, filename, data)
}

// MoveMedia moves a media file to a new location
func (s *MediaService) MoveMedia(oldPath, newPath string) error {
	oldFull := filepath.Join(s.notesDir, oldPath)
	newFull := filepath.Join(s.notesDir, newPath)

	// Security checks
	if !ValidatePathSecurity(s.notesDir, oldPath) || !ValidatePathSecurity(s.notesDir, newPath) {
		return ErrInvalidPath
	}

	// Check source exists
	if _, err := os.Stat(oldFull); err != nil {
		return err
	}

	// Create parent directory if needed
	if err := os.MkdirAll(filepath.Dir(newFull), 0755); err != nil {
		return err
	}

	return os.Rename(oldFull, newFull)
}

// FindOrphanedMedia finds media files that are not referenced in any notes
// Returns a list of orphaned media file paths (relative to notes dir)
func (s *MediaService) FindOrphanedMedia() ([]string, error) {
	// Step 1: Collect all media files in _attachments directories
	mediaFiles, err := s.collectAllMediaFiles()
	if err != nil {
		return nil, err
	}

	// Step 2: Collect all referenced media from notes
	referencedFiles, err := s.collectReferencedMedia()
	if err != nil {
		return nil, err
	}

	// Step 3: Find orphaned files (exist but not referenced)
	var orphanedFiles []string
	for _, mediaFile := range mediaFiles {
		if !referencedFiles[mediaFile] {
			orphanedFiles = append(orphanedFiles, mediaFile)
		}
	}

	return orphanedFiles, nil
}

// collectAllMediaFiles scans all _attachments directories for media files
func (s *MediaService) collectAllMediaFiles() ([]string, error) {
	var mediaFiles []string
	var walkErrors []string

	err := filepath.WalkDir(s.notesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			walkErrors = append(walkErrors, fmt.Sprintf("WalkDir error at %s: %v", path, err))
			return nil // Skip errors but log them
		}

		// Skip dot directories
		if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
			return fs.SkipDir
		}

		// Skip dot files
		if strings.HasPrefix(d.Name(), ".") {
			return nil
		}

		// Only process files
		if d.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(s.notesDir, path)
		if err != nil {
			return nil
		}

		// Check if path contains _attachments
		if !strings.Contains(relPath, "_attachments") {
			return nil
		}

		// Check if it's a media file
		if !IsMediaFile(d.Name()) {
			return nil
		}

		// Security check
		if !ValidatePathSecurity(s.notesDir, relPath) {
			return nil
		}

		mediaFiles = append(mediaFiles, ToPosixPath(relPath))
		return nil
	})

	// Log walk errors if any
	if len(walkErrors) > 0 {
		for _, errMsg := range walkErrors[:min(len(walkErrors), 10)] {
			logger.Printf("Warning: %s", errMsg)
		}
		if len(walkErrors) > 10 {
			logger.Printf("Warning: ... and %d more walk errors", len(walkErrors)-10)
		}
	}

	return mediaFiles, err
}

// collectReferencedMedia parses all notes and extracts media references
func (s *MediaService) collectReferencedMedia() (map[string]bool, error) {
	referenced := make(map[string]bool)
	var walkErrors []string

	err := filepath.WalkDir(s.notesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			walkErrors = append(walkErrors, fmt.Sprintf("WalkDir error at %s: %v", path, err))
			return nil // Skip errors but log them
		}

		// Skip dot directories
		if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
			return fs.SkipDir
		}

		// Skip dot files
		if strings.HasPrefix(d.Name(), ".") {
			return nil
		}

		// Only process markdown files
		if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}

		relPath, err := filepath.Rel(s.notesDir, path)
		if err != nil {
			return nil
		}

		// Security check
		if !ValidatePathSecurity(s.notesDir, relPath) {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		contentStr := string(content)

		// Extract markdown image references: ![alt](path)
		mdMatches := markdownImageRegex.FindAllStringSubmatch(contentStr, -1)
		for _, match := range mdMatches {
			if len(match) > 1 {
				refPath := match[1]
				// Clean up the path (remove URL query params and fragments)
				refPath = strings.Split(refPath, "?")[0]
				refPath = strings.Split(refPath, "#")[0]

				// Normalize path
				refPath = s.normalizeMediaReference(refPath, filepath.Dir(relPath))
				if refPath != "" {
					referenced[refPath] = true
				}
			}
		}

		// Extract wikilink references: [[filename.ext]]
		wikiMatches := wikilinkMediaRegex.FindAllStringSubmatch(contentStr, -1)
		for _, match := range wikiMatches {
			if len(match) > 1 {
				refPath := match[1]
				refPath = s.normalizeMediaReference(refPath, filepath.Dir(relPath))
				if refPath != "" {
					referenced[refPath] = true
				}
			}
		}

		// Extract media embed wikilinks: ![[file.png]] or ![[file.png|alt text]]
		// This is the primary format used by the frontend for embedding media
		mediaWikiMatches := mediaWikilinkRegex.FindAllStringSubmatch(contentStr, -1)
		for _, match := range mediaWikiMatches {
			if len(match) > 1 {
				refPath := match[1]
				refPath = s.normalizeMediaReference(refPath, filepath.Dir(relPath))
				if refPath != "" {
					referenced[refPath] = true
				}
			}
		}

		// Extract HTML img src references
		htmlImgMatches := htmlImgSrcRegex.FindAllStringSubmatch(contentStr, -1)
		for _, match := range htmlImgMatches {
			if len(match) > 1 {
				refPath := match[1]
				refPath = s.normalizeMediaReference(refPath, filepath.Dir(relPath))
				if refPath != "" {
					referenced[refPath] = true
				}
			}
		}

		// Extract HTML video src references
		htmlVideoMatches := htmlVideoSrcRegex.FindAllStringSubmatch(contentStr, -1)
		for _, match := range htmlVideoMatches {
			if len(match) > 1 {
				refPath := match[1]
				refPath = s.normalizeMediaReference(refPath, filepath.Dir(relPath))
				if refPath != "" {
					referenced[refPath] = true
				}
			}
		}

		// Extract HTML audio src references
		htmlAudioMatches := htmlAudioSrcRegex.FindAllStringSubmatch(contentStr, -1)
		for _, match := range htmlAudioMatches {
			if len(match) > 1 {
				refPath := match[1]
				refPath = s.normalizeMediaReference(refPath, filepath.Dir(relPath))
				if refPath != "" {
					referenced[refPath] = true
				}
			}
		}

		// Extract HTML source src references (used in video/audio with multiple sources)
		htmlSourceMatches := htmlSourceSrcRegex.FindAllStringSubmatch(contentStr, -1)
		for _, match := range htmlSourceMatches {
			if len(match) > 1 {
				refPath := match[1]
				refPath = s.normalizeMediaReference(refPath, filepath.Dir(relPath))
				if refPath != "" {
					referenced[refPath] = true
				}
			}
		}

		return nil
	})

	// Log walk errors if any
	if len(walkErrors) > 0 {
		for _, errMsg := range walkErrors[:min(len(walkErrors), 10)] {
			logger.Printf("Warning: %s", errMsg)
		}
		if len(walkErrors) > 10 {
			logger.Printf("Warning: ... and %d more walk errors", len(walkErrors)-10)
		}
	}

	return referenced, err
}

// normalizeMediaReference converts a media reference to a normalized relative path
func (s *MediaService) normalizeMediaReference(refPath, noteDir string) string {
	// URL decode if needed
	if decoded, err := url.QueryUnescape(refPath); err == nil {
		refPath = decoded
	}

	// Handle different path formats
	refPath = filepath.ToSlash(refPath)

	// If path starts with /, it's relative to notes dir
	if strings.HasPrefix(refPath, "/") {
		refPath = strings.TrimPrefix(refPath, "/")
		if ValidatePathSecurity(s.notesDir, refPath) {
			return ToPosixPath(refPath)
		}
		return ""
	}

	// If path contains _attachments, use it as-is (relative to notes dir)
	if strings.Contains(refPath, "_attachments") {
		if ValidatePathSecurity(s.notesDir, refPath) {
			return ToPosixPath(refPath)
		}
		return ""
	}

	// If path is just a filename (no path separators), try to find it in _attachments directories
	// This handles the common case: ![[image.png]] referencing _attachments/image.png
	if !strings.Contains(refPath, "/") && !strings.Contains(refPath, "\\") {
		// Try global _attachments directory first
		globalAttachPath := filepath.Join("_attachments", refPath)
		globalAttachPath = ToPosixPath(globalAttachPath)
		fullGlobalPath := filepath.Join(s.notesDir, globalAttachPath)
		if _, err := os.Stat(fullGlobalPath); err == nil {
			if ValidatePathSecurity(s.notesDir, globalAttachPath) {
				return globalAttachPath
			}
		}

		// Try note's local _attachments directory
		if noteDir != "" && noteDir != "." {
			localAttachPath := filepath.Join(noteDir, "_attachments", refPath)
			localAttachPath = ToPosixPath(localAttachPath)
			fullLocalPath := filepath.Join(s.notesDir, localAttachPath)
			if _, err := os.Stat(fullLocalPath); err == nil {
				if ValidatePathSecurity(s.notesDir, localAttachPath) {
					return localAttachPath
				}
			}
		}
	}

	// Otherwise, resolve relative to note's directory
	if noteDir == "." {
		noteDir = ""
	}

	var fullPath string
	if noteDir == "" {
		fullPath = refPath
	} else {
		fullPath = filepath.Join(noteDir, refPath)
	}

	fullPath = ToPosixPath(fullPath)

	// Security check
	if !ValidatePathSecurity(s.notesDir, fullPath) {
		return ""
	}

	return fullPath
}

// DeleteOrphanedMedia finds and deletes orphaned media files
// Returns a list of successfully deleted file paths
func (s *MediaService) DeleteOrphanedMedia() ([]string, error) {
	// Get list of orphaned media files
	orphanedFiles, err := s.FindOrphanedMedia()
	if err != nil {
		return nil, fmt.Errorf("failed to find orphaned media: %w", err)
	}

	var deletedFiles []string
	var deletedDirs []string // Track directories that might become empty
	deletedDirSet := make(map[string]bool)

	for _, orphanPath := range orphanedFiles {
		// Validate path security before deletion
		if !ValidatePathSecurity(s.notesDir, orphanPath) {
			logger.Printf("[OrphanedMedia] Skipped invalid path: %s", orphanPath)
			continue
		}

		fullPath := filepath.Join(s.notesDir, orphanPath)

		// Check if it's a file (not a directory)
		info, err := os.Stat(fullPath)
		if err != nil {
			logger.Printf("[OrphanedMedia] Failed to stat file %s: %v", orphanPath, err)
			continue
		}
		if info.IsDir() {
			logger.Printf("[OrphanedMedia] Skipped directory: %s", orphanPath)
			continue
		}

		// Delete the file
		if err := os.Remove(fullPath); err != nil {
			logger.Printf("[OrphanedMedia] Failed to delete %s: %v", orphanPath, err)
			continue
		}

		deletedFiles = append(deletedFiles, orphanPath)
		logger.Printf("[OrphanedMedia] Deleted: %s", orphanPath)

		// Track parent directory for potential cleanup
		parentDir := filepath.Dir(orphanPath)
		if strings.Contains(parentDir, "_attachments") && !deletedDirSet[parentDir] {
			deletedDirSet[parentDir] = true
			deletedDirs = append(deletedDirs, parentDir)
		}
	}

	// Clean up empty directories (from deepest to shallowest)
	// Sort by depth (deepest first)
	for i := len(deletedDirs) - 1; i >= 0; i-- {
		dirPath := deletedDirs[i]
		fullDirPath := filepath.Join(s.notesDir, dirPath)

		// Check if directory is empty
		entries, err := os.ReadDir(fullDirPath)
		if err != nil {
			continue // Directory doesn't exist or can't be read
		}

		if len(entries) == 0 {
			// Directory is empty, remove it
			if err := os.Remove(fullDirPath); err != nil {
				logger.Printf("[OrphanedMedia] Failed to remove empty directory %s: %v", dirPath, err)
			} else {
				logger.Printf("[OrphanedMedia] Removed empty directory: %s", dirPath)
			}
		}
	}

	return deletedFiles, nil
}

// ErrInvalidPath is returned when path validation fails
var ErrInvalidPath = &PathError{Msg: "invalid path"}

// PathError represents a path-related error
type PathError struct {
	Msg string
}

func (e *PathError) Error() string {
	return e.Msg
}

// GetFileContentType returns the MIME type for a file
func GetFileContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	contentTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".ogg":  "audio/ogg",
		".m4a":  "audio/mp4",
		".mp4":  "video/mp4",
		".webm": "video/webm",
		".mov":  "video/quicktime",
		".avi":  "video/x-msvideo",
		".pdf":  "application/pdf",
	}

	if ct, ok := contentTypes[ext]; ok {
		return ct
	}

	return "application/octet-stream"
}
