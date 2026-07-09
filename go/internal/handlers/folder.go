package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"gonote/internal/models/config"
	"gonote/internal/models"
	"gonote/internal/services"
	"gonote/internal/middleware"
)

// CacheInvalidator interface for cache invalidation
type CacheInvalidator interface {
	InvalidateCache()
}

// FolderHandler handles folder-related requests
type FolderHandler struct {
	config       *config.Config
	cacheInv     CacheInvalidator
	shareService *services.ShareService
	searchIndex  *services.SearchIndex
}

// NewFolderHandler creates a new FolderHandler
func NewFolderHandler(cfg *config.Config) *FolderHandler {
	return &FolderHandler{config: cfg}
}

// NewFolderHandlerWithCache creates a new FolderHandler with cache invalidation
func NewFolderHandlerWithCache(cfg *config.Config, cacheInv CacheInvalidator) *FolderHandler {
	return &FolderHandler{config: cfg, cacheInv: cacheInv}
}

// NewFolderHandlerFull creates a new FolderHandler with all dependencies
func NewFolderHandlerFull(cfg *config.Config, cacheInv CacheInvalidator, shareService *services.ShareService, searchIndex *services.SearchIndex) *FolderHandler {
	return &FolderHandler{config: cfg, cacheInv: cacheInv, shareService: shareService, searchIndex: searchIndex}
}

// Create creates a new folder
func (h *FolderHandler) Create(c *fiber.Ctx) error {
	var req struct {
		Path string `json:"path"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"detail": "Invalid request body"})
	}

	// Security check
	if !validatePath(c, h.config.Storage.NotesDir, req.Path) {
		return nil
	}

	if err := services.CreateFolder(h.config.Storage.NotesDir, req.Path); err != nil {
		return fiber.NewError(500, err.Error())
	}

	// Invalidate cache after folder creation
	if h.cacheInv != nil {
		h.cacheInv.InvalidateCache()
	}

	return c.JSON(models.FolderResponse{
		Success: true,
		Path:    req.Path,
		Message: "Folder created successfully",
	})
}

// Delete deletes a folder and cleans up share tokens and search index
func (h *FolderHandler) Delete(c *fiber.Ctx) error {
	folderPath, ok := resolvePathParam(c, h.config.Storage.NotesDir)
	if !ok {
		return nil
	}

	if err := services.DeleteFolder(h.config.Storage.NotesDir, folderPath); err != nil {
		return fiber.NewError(500, err.Error())
	}

	// Clean up share tokens and search index for all notes under this folder
	notes, err := services.ListNotesUnderPath(h.config.Storage.NotesDir, folderPath)
	if err == nil {
		for _, notePath := range notes {
			if h.shareService != nil {
				h.shareService.DeleteTokenForNote(notePath)
			}
			if h.searchIndex != nil {
				h.searchIndex.RemoveFromIndex(notePath)
			}
		}
	}

	if h.cacheInv != nil {
		h.cacheInv.InvalidateCache()
	}

	return c.JSON(models.FolderResponse{
		Success: true,
		Path:    folderPath,
		Message: "Folder deleted successfully",
	})
}

// Move moves a folder, updating share tokens and search index for all notes
func (h *FolderHandler) Move(c *fiber.Ctx) error {
	var req struct {
		OldPath string `json:"oldPath"`
		NewPath string `json:"newPath"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"detail": "Invalid request body"})
	}

	// Security checks
	if !validatePath(c, h.config.Storage.NotesDir, req.OldPath) ||
		!validatePath(c, h.config.Storage.NotesDir, req.NewPath) {
		return nil
	}

	// Update share token paths before moving the folder
	if h.shareService != nil {
		h.shareService.MoveFolderTokens(req.OldPath, req.NewPath)
	}

	// Collect old note paths for search index cleanup
	var oldNotePaths []string
	if h.searchIndex != nil {
		oldNotePaths, _ = services.ListNotesUnderPath(h.config.Storage.NotesDir, req.OldPath)
	}

	// Update all wikilink references before moving
	backlinkService := services.NewBacklinkService(h.config.Storage.NotesDir)
	backlinkService.UpdateFolderBacklinks(req.OldPath, req.NewPath)

	if err := services.MoveFolder(h.config.Storage.NotesDir, req.OldPath, req.NewPath); err != nil {
		return fiber.NewError(500, err.Error())
	}

	// Update search index: remove old paths, add new paths
	if h.searchIndex != nil {
		for _, oldPath := range oldNotePaths {
			h.searchIndex.RemoveFromIndex(oldPath)
			newPath := strings.Replace(oldPath, req.OldPath, req.NewPath, 1)
			h.searchIndex.UpdateIndex(newPath)
		}
	}

	if h.cacheInv != nil {
		h.cacheInv.InvalidateCache()
	}

	return c.JSON(models.FolderResponse{
		Success: true,
		Path:    req.NewPath,
		Message: "Folder moved successfully",
	})
}

// Rename renames a folder, updating share tokens and search index for all notes
func (h *FolderHandler) Rename(c *fiber.Ctx) error {
	var req struct {
		OldPath string `json:"oldPath"`
		NewPath string `json:"newPath"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"detail": "Invalid request body"})
	}

	// Security checks
	if !validatePath(c, h.config.Storage.NotesDir, req.OldPath) ||
		!validatePath(c, h.config.Storage.NotesDir, req.NewPath) {
		return nil
	}

	// Update share token paths before renaming the folder
	if h.shareService != nil {
		h.shareService.MoveFolderTokens(req.OldPath, req.NewPath)
	}

	// Collect old note paths for search index cleanup
	var oldNotePaths []string
	if h.searchIndex != nil {
		oldNotePaths, _ = services.ListNotesUnderPath(h.config.Storage.NotesDir, req.OldPath)
	}

	// Update all wikilink references before renaming
	backlinkService := services.NewBacklinkService(h.config.Storage.NotesDir)
	backlinkService.UpdateFolderBacklinks(req.OldPath, req.NewPath)

	if err := services.RenameFolder(h.config.Storage.NotesDir, req.OldPath, req.NewPath); err != nil {
		return fiber.NewError(500, err.Error())
	}

	// Update search index: remove old paths, add new paths
	if h.searchIndex != nil {
		for _, oldPath := range oldNotePaths {
			h.searchIndex.RemoveFromIndex(oldPath)
			newPath := strings.Replace(oldPath, req.OldPath, req.NewPath, 1)
			h.searchIndex.UpdateIndex(newPath)
		}
	}

	if h.cacheInv != nil {
		h.cacheInv.InvalidateCache()
	}

	return c.JSON(models.FolderResponse{
		Success: true,
		Path:    req.NewPath,
		Message: "Folder renamed successfully",
	})
}

func (h *FolderHandler) RegisterRoutes(api fiber.Router) {
	api.Post("/folders", middleware.EndpointLimiterSimple(30), h.Create)
	api.Post("/folders/move", middleware.EndpointLimiterSimple(20), h.Move)
	api.Post("/folders/rename", middleware.EndpointLimiterSimple(30), h.Rename)
	api.Delete("/folders/*", middleware.EndpointLimiterSimple(20), h.Delete)
}