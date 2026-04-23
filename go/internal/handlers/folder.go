package handlers

import (
	"net/url"

	"github.com/gofiber/fiber/v2"

	"gonote/internal/models/config"
	"gonote/internal/models"
	"gonote/internal/services"
)

// CacheInvalidator interface for cache invalidation
type CacheInvalidator interface {
	InvalidateCache()
}

// FolderHandler handles folder-related requests
type FolderHandler struct {
	config   *config.Config
	cacheInv CacheInvalidator
}

// NewFolderHandler creates a new FolderHandler
func NewFolderHandler(cfg *config.Config) *FolderHandler {
	return &FolderHandler{config: cfg}
}

// NewFolderHandlerWithCache creates a new FolderHandler with cache invalidation
func NewFolderHandlerWithCache(cfg *config.Config, cacheInv CacheInvalidator) *FolderHandler {
	return &FolderHandler{config: cfg, cacheInv: cacheInv}
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
	if !services.ValidatePathSecurity(h.config.Storage.NotesDir, req.Path) {
		return c.Status(400).JSON(fiber.Map{"detail": "Invalid path"})
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

// Delete deletes a folder
func (h *FolderHandler) Delete(c *fiber.Ctx) error {
	folderPath := c.Params("*")

	// URL decode the path to handle special characters (Chinese, emoji, etc.)
	decodedPath, err := url.PathUnescape(folderPath)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"detail": "Invalid path encoding"})
	}

	// Security check
	if !services.ValidatePathSecurity(h.config.Storage.NotesDir, decodedPath) {
		return c.Status(400).JSON(fiber.Map{"detail": "Invalid path"})
	}

	if err := services.DeleteFolder(h.config.Storage.NotesDir, decodedPath); err != nil {
		return fiber.NewError(500, err.Error())
	}

	if h.cacheInv != nil {
		h.cacheInv.InvalidateCache()
	}

	return c.JSON(models.FolderResponse{
		Success: true,
		Path:    decodedPath,
		Message: "Folder deleted successfully",
	})
}

// Move moves a folder to a new location
func (h *FolderHandler) Move(c *fiber.Ctx) error {
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

	// Update all wikilink references before moving
	backlinkService := services.NewBacklinkService(h.config.Storage.NotesDir)
	backlinkService.UpdateFolderBacklinks(req.OldPath, req.NewPath)

	if err := services.MoveFolder(h.config.Storage.NotesDir, req.OldPath, req.NewPath); err != nil {
		return fiber.NewError(500, err.Error())
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

// Rename renames a folder
func (h *FolderHandler) Rename(c *fiber.Ctx) error {
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

	// Update all wikilink references before renaming
	backlinkService := services.NewBacklinkService(h.config.Storage.NotesDir)
	backlinkService.UpdateFolderBacklinks(req.OldPath, req.NewPath)

	if err := services.RenameFolder(h.config.Storage.NotesDir, req.OldPath, req.NewPath); err != nil {
		return fiber.NewError(500, err.Error())
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
