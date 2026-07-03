package handlers

import (
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"gonote/internal/models/config"
	"gonote/internal/models"
	"gonote/internal/services"
)

type NoteHandler struct {
	service     *services.NoteService
	config      *config.Config
	searchIndex *services.SearchIndex
	tagService  *services.TagService
}

func NewNoteHandler(service *services.NoteService, cfg *config.Config) *NoteHandler {
	return &NoteHandler{service: service, config: cfg}
}

func NewNoteHandlerWithIndex(service *services.NoteService, cfg *config.Config, searchIndex *services.SearchIndex) *NoteHandler {
	return &NoteHandler{service: service, config: cfg, searchIndex: searchIndex}
}

func NewNoteHandlerWithTagService(service *services.NoteService, tagService *services.TagService, cfg *config.Config, searchIndex *services.SearchIndex) *NoteHandler {
	return &NoteHandler{service: service, tagService: tagService, config: cfg, searchIndex: searchIndex}
}

func (h *NoteHandler) List(c *fiber.Ctx) error {
	includeMedia := c.Query("include_media", "false") == "true"
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 50)
	tagsParam := c.Query("tags", "") // Comma-separated tags for filtering

	notes, folders, err := h.service.ScanNotes(includeMedia)
	if err != nil {
		return fiber.NewError(500, err.Error())
	}

	// Apply tag filtering if tags parameter is provided
	if tagsParam != "" {
		tagList := strings.Split(tagsParam, ",")
		notes = services.FilterNotesByTags(notes, tagList)
	}

	paginatedResult := services.Paginate(notes, page, limit)

	return c.JSON(models.NotesListResponse{
		Notes:      paginatedResult.Notes,
		Folders:    folders,
		Pagination: paginatedResult.Pagination,
	})
}

// Get returns a single note or creates a new one
func (h *NoteHandler) Get(c *fiber.Ctx) error {
	notePath := c.Params("*")
	notePath = strings.TrimPrefix(notePath, "/")

	// URL decode the path to handle special characters (Chinese, emoji, etc.)
	decodedPath, err := url.PathUnescape(notePath)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"detail": "Invalid path encoding"})
	}
	notePath = decodedPath

	// Security check
	if !services.ValidatePathSecurity(h.config.Storage.NotesDir, notePath) {
		return c.Status(400).JSON(fiber.Map{"detail": "Invalid path"})
	}

	// Check if note exists
	if !h.service.NoteExists(notePath) {
		// Return empty note for new notes
		return c.JSON(models.NoteContent{
			Path:    notePath,
			Content: "",
			Metadata: models.NoteMetadata{
				Created:  "",
				Modified: "",
				Size:     0,
				Lines:    0,
			},
		})
	}

	content, err := h.service.GetNoteContent(notePath)
	if err != nil {
		return fiber.NewError(500, err.Error())
	}

	metadata, err := h.service.GetNoteMetadata(notePath)
	if err != nil {
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(models.NoteContent{
		Path:     notePath,
		Content:  content,
		Metadata: *metadata,
	})
}

// CreateOrUpdate creates or updates a note
func (h *NoteHandler) CreateOrUpdate(c *fiber.Ctx) error {
	notePath := c.Params("*")
	notePath = strings.TrimPrefix(notePath, "/")

	// URL decode the path to handle special characters (Chinese, emoji, etc.)
	decodedPath, err := url.PathUnescape(notePath)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"detail": "Invalid path encoding"})
	}
	notePath = decodedPath

	// Handle /notes/move specially - redirect to Move handler
	if notePath == "move" {
		return h.Move(c)
	}

	// Security check
	if !services.ValidatePathSecurity(h.config.Storage.NotesDir, notePath) {
		return c.Status(400).JSON(fiber.Map{"detail": "Invalid path"})
	}

	// Parse request body
	var req struct {
		Content  string `json:"content"`
		Modified string `json:"modified,omitempty"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"detail": "Invalid request body"})
	}

	// Save note with optional mtime-based optimistic lock
	err = h.service.SaveNoteWithCheck(notePath, req.Content, req.Modified)
	if err != nil {
		if errors.Is(err, services.ErrConflict) {
			fullPath := filepath.Join(h.config.Storage.NotesDir, notePath)
			info, statErr := os.Stat(fullPath)
			serverMtime := ""
			if statErr == nil {
				serverMtime = info.ModTime().UTC().Format(time.RFC3339Nano)
			}
			return c.Status(409).JSON(fiber.Map{
				"detail":   "Note modified by another source",
				"modified": serverMtime,
			})
		}
		return fiber.NewError(500, err.Error())
	}

	// Invalidate caches
	h.service.InvalidateCache()
	
	// Also invalidate tag cache if tag service is available
	if h.tagService != nil {
		h.tagService.ClearCache()
	}

	if h.searchIndex != nil {
		h.searchIndex.UpdateIndex(notePath)
	}

	// Get authoritative mtime after save
	fullPath := filepath.Join(h.config.Storage.NotesDir, notePath)
	info, statErr := os.Stat(fullPath)
	serverMtime := ""
	if statErr == nil {
		serverMtime = info.ModTime().UTC().Format(time.RFC3339Nano)
	}

	return c.JSON(models.NoteSaveResponse{
		Success:  true,
		Path:     notePath,
		Message:  "Note saved successfully",
		Content:  req.Content,
		Modified: serverMtime,
	})
}

// Delete deletes a note
func (h *NoteHandler) Delete(c *fiber.Ctx) error {
	notePath := c.Params("*")
	notePath = strings.TrimPrefix(notePath, "/")

	// URL decode the path to handle special characters (Chinese, emoji, etc.)
	decodedPath, err := url.PathUnescape(notePath)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"detail": "Invalid path encoding"})
	}
	notePath = decodedPath

	// Security check
	if !services.ValidatePathSecurity(h.config.Storage.NotesDir, notePath) {
		return c.Status(400).JSON(fiber.Map{"detail": "Invalid path"})
	}

	if err := h.service.DeleteNote(notePath); err != nil {
		return fiber.NewError(500, err.Error())
	}

	h.service.InvalidateCache()
	
	// Also invalidate tag cache if tag service is available
	if h.tagService != nil {
		h.tagService.ClearCache()
	}

	if h.searchIndex != nil {
		h.searchIndex.RemoveFromIndex(notePath)
	}

	return c.JSON(models.APIResponse{
		Success: true,
		Message: "Note deleted successfully",
	})
}

// Move moves a note to a new location
func (h *NoteHandler) Move(c *fiber.Ctx) error {
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

	if err := h.service.MoveNote(req.OldPath, req.NewPath); err != nil {
		return fiber.NewError(500, err.Error())
	}

	// Invalidate caches
	h.service.InvalidateCache()
	
	// Also invalidate tag cache if tag service is available
	if h.tagService != nil {
		h.tagService.ClearCache()
	}

	// Update share token if exists
	shareService := services.NewShareService(h.config.Storage.NotesDir)
	shareService.UpdateTokenPath(req.OldPath, req.NewPath)

	// Update search index: remove old path, add new path
	if h.searchIndex != nil {
		h.searchIndex.RemoveFromIndex(req.OldPath)
		h.searchIndex.UpdateIndex(req.NewPath)
	}

	return c.JSON(models.NoteMoveResponse{
		Success: true,
		OldPath: req.OldPath,
		NewPath: req.NewPath,
		Message: "Note moved successfully",
	})
}

// GetAttachmentDir returns the attachment directory for a note
func (h *NoteHandler) GetAttachmentDir(notePath string) string {
	return h.service.GetAttachmentDir(notePath)
}

// SaveUploadedImage saves an uploaded image
func (h *NoteHandler) SaveUploadedImage(notePath, filename string, data []byte) (string, error) {
	return h.service.SaveUploadedImage(notePath, filename, data)
}

// GetNotesDir returns the notes directory
func (h *NoteHandler) GetNotesDir() string {
	return h.config.Storage.NotesDir
}

func (h *NoteHandler) GetAttachments(c *fiber.Ctx) error {
	notePath := c.Params("*")
	notePath = strings.TrimPrefix(notePath, "/")

	// URL decode the path to handle special characters (Chinese, emoji, etc.)
	decodedPath, err := url.PathUnescape(notePath)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"detail": "Invalid path encoding"})
	}
	notePath = decodedPath

	if !services.ValidatePathSecurity(h.config.Storage.NotesDir, notePath) {
		return c.Status(400).JSON(fiber.Map{"detail": "Invalid path"})
	}

	attachments, err := h.service.GetAttachments(notePath)
	if err != nil {
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(models.AttachmentsResponse{
		Success:     true,
		Attachments: attachments,
		Count:       len(attachments),
	})
}
