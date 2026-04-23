package handlers

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"gonote/internal/models/config"
	"gonote/internal/models"
	"gonote/internal/services"
)

// MediaHandler handles media-related requests
type MediaHandler struct {
	service    *services.MediaService
	noteService *services.NoteService
	config     *config.Config
}

// NewMediaHandler creates a new MediaHandler
func NewMediaHandler(service *services.MediaService, noteService *services.NoteService, cfg *config.Config) *MediaHandler {
	return &MediaHandler{service: service, noteService: noteService, config: cfg}
}

// validateUpload checks file size and type against configuration limits
func (h *MediaHandler) validateUpload(file *multipart.FileHeader) error {
	// Check file size
	maxSize := int64(h.config.Upload.MaxFileSizeMB) * 1024 * 1024
	if file.Size > maxSize {
		return fmt.Errorf("file too large: %d bytes (max: %dMB)", file.Size, h.config.Upload.MaxFileSizeMB)
	}

	// Check file type if restrictions are set
	if len(h.config.Upload.AllowedTypes) > 0 {
		contentType := file.Header.Get("Content-Type")
		if !h.isAllowedType(contentType) {
			return fmt.Errorf("file type not allowed: %s", contentType)
		}
	}

	return nil
}

// isAllowedType checks if a MIME type is in the allowed list
func (h *MediaHandler) isAllowedType(contentType string) bool {
	if len(h.config.Upload.AllowedTypes) == 0 {
		return true // No restrictions
	}

	// Handle charset suffix (e.g., "text/plain; charset=utf-8")
	contentType = strings.Split(contentType, ";")[0]
	contentType = strings.TrimSpace(contentType)

	for _, allowed := range h.config.Upload.AllowedTypes {
		if strings.EqualFold(contentType, allowed) {
			return true
		}
		// Support wildcard patterns like "image/*"
		if strings.HasSuffix(allowed, "/*") {
			prefix := strings.TrimSuffix(allowed, "*")
			if strings.HasPrefix(strings.ToLower(contentType), strings.ToLower(prefix)) {
				return true
			}
		}
	}

	return false
}

// Get returns a media file using streaming for large files
func (h *MediaHandler) Get(c *fiber.Ctx) error {
	mediaPath := c.Params("*")

	// Security check
	if !services.ValidatePathSecurity(h.config.Storage.NotesDir, mediaPath) {
		return c.Status(400).JSON(fiber.Map{"detail": "Invalid path"})
	}

	fullPath := filepath.Join(h.config.Storage.NotesDir, mediaPath)

	// Check if file exists
	info, err := os.Stat(fullPath)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"detail": "Media not found"})
	}

	// Get content type
	contentType := services.GetFileContentType(info.Name())
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Set content type
	c.Set("Content-Type", contentType)

	// For large files (>10MB), use SendFile for streaming
	// This avoids loading the entire file into memory
	const largeFileThreshold = 10 * 1024 * 1024 // 10MB
	if info.Size() > largeFileThreshold {
		c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", info.Name()))
	}

	// SendFile streams the file directly to response
	return c.SendFile(fullPath)
}

// Upload uploads a media file with size and type validation
func (h *MediaHandler) Upload(c *fiber.Ctx) error {
	// Parse multipart form
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"detail": "No file uploaded"})
	}

	// Validate file size and type
	if err := h.validateUpload(file); err != nil {
		return c.Status(413).JSON(fiber.Map{"detail": err.Error()})
	}

	notePath := c.FormValue("note_path", "")

	// Sanitize filename and generate unique name
	sanitizedName := services.SanitizeFilename(file.Filename)
	ext := filepath.Ext(sanitizedName)
	name := strings.TrimSuffix(sanitizedName, ext)
	timestamp := time.Now().Format("20060102150405")
	finalFilename := fmt.Sprintf("%s-%s%s", name, timestamp, ext)

	// Get attachments directory
	attachmentsDir := h.noteService.GetAttachmentDir(notePath)
	if err := os.MkdirAll(attachmentsDir, 0755); err != nil {
		return fiber.NewError(500, err.Error())
	}

	fullPath := filepath.Join(attachmentsDir, finalFilename)

	// Security check - ensure path is within notes directory
	absPath, _ := filepath.Abs(fullPath)
	absNotesDir, _ := filepath.Abs(h.config.Storage.NotesDir)
	if !strings.HasPrefix(absPath, absNotesDir) {
		return c.Status(400).JSON(fiber.Map{"detail": "Invalid path"})
	}

	// Stream file directly to disk (no memory loading)
	if err := c.SaveFile(file, fullPath); err != nil {
		return fiber.NewError(500, err.Error())
	}

	// Get relative path for response
	relPath, _ := filepath.Rel(h.config.Storage.NotesDir, fullPath)
	path := services.ToPosixPath(relPath)

	// Determine media type
	mediaType := services.GetMediaType(file.Filename)

	// Invalidate cache to ensure fresh data is returned on subsequent calls
	h.noteService.InvalidateCache()

	return c.JSON(models.MediaUploadResponse{
		Success:  true,
		Path:     path,
		Filename: file.Filename,
		Type:     mediaType,
		Message:  "File uploaded successfully",
	})
}

// Move moves a media file
func (h *MediaHandler) Move(c *fiber.Ctx) error {
	var req struct {
		OldPath string `json:"oldPath"`
		NewPath string `json:"newPath"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"detail": "Invalid request body"})
	}

	// Security checks
	if !services.ValidatePathSecurity(h.config.Storage.NotesDir, req.OldPath) ||
		!services.ValidatePathSecurity(h.config.Storage.NotesDir, req.NewPath) {
		return c.Status(400).JSON(fiber.Map{"detail": "Invalid path"})
	}

	if err := h.service.MoveMedia(req.OldPath, req.NewPath); err != nil {
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(models.APIResponse{
		Success: true,
		Message: "Media moved successfully",
	})
}

// GetNotesDir returns the notes directory
func (h *MediaHandler) GetNotesDir() string {
	return h.config.Storage.NotesDir
}

// GetAttachmentDir returns the attachment directory for a note
func (h *MediaHandler) GetAttachmentDir(notePath string) string {
	ns := services.NewNoteService(h.config.Storage.NotesDir)
	return ns.GetAttachmentDir(notePath)
}

// ListOrphanedMedia returns a list of orphaned media files
func (h *MediaHandler) ListOrphanedMedia(c *fiber.Ctx) error {
	orphanedPaths, err := h.service.FindOrphanedMedia()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"detail": "Failed to find orphaned media"})
	}

	var files = []models.OrphanedMediaFile{}
	var totalSize int64

	for _, path := range orphanedPaths {
		fullPath := h.config.Storage.NotesDir + "/" + path
		info, err := os.Stat(fullPath)
		if err != nil {
			continue // Skip files that can't be accessed
		}

		filename := filepath.Base(path)
		mediaType := services.GetMediaType(filename)

		files = append(files, models.OrphanedMediaFile{
			Path:      path,
			Filename:  filename,
			Size:      info.Size(),
			MediaType: mediaType,
			Type:      mediaType,
		})

		totalSize += info.Size()
	}

	return c.JSON(models.OrphanedMediaResponse{
		Success:   true,
		Count:     len(files),
		Files:     files,
		TotalSize: totalSize,
	})
}

// CleanupOrphanedMedia finds and deletes orphaned media files
func (h *MediaHandler) CleanupOrphanedMedia(c *fiber.Ctx) error {
	// First, find orphaned files to calculate sizes before deletion
	orphanedFiles, err := h.service.FindOrphanedMedia()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"detail": "Failed to find orphaned media"})
	}

	// Calculate total size before deletion
	var freedSpace int64
	for _, orphanPath := range orphanedFiles {
		fullPath := filepath.Join(h.config.Storage.NotesDir, orphanPath)
		if info, err := os.Stat(fullPath); err == nil {
			freedSpace += info.Size()
		}
	}

	// Delete orphaned files
	deletedFiles, err := h.service.DeleteOrphanedMedia()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"detail": "Failed to delete orphaned media"})
	}

	// Build response
	message := "No orphaned media files found"
	if len(deletedFiles) > 0 {
		message = "Successfully deleted orphaned media files"
	}

	return c.JSON(models.CleanupMediaResponse{
		Success:      true,
		DeletedCount: len(deletedFiles),
		DeletedFiles: deletedFiles,
		FreedSpace:   freedSpace,
		Message:      message,
	})
}


