package handlers

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"

	"gonote/internal/models/config"
	"gonote/internal/models"
	"gonote/internal/services"
)

// ShareHandler handles share-related requests
type ShareHandler struct {
	shareService  *services.ShareService
	exportService *services.ExportService
	config        *config.Config
}

// NewShareHandler creates a new ShareHandler
func NewShareHandler(shareService *services.ShareService, exportService *services.ExportService, cfg *config.Config) *ShareHandler {
	return &ShareHandler{
		shareService:  shareService,
		exportService: exportService,
		config:        cfg,
	}
}

// Create creates a share token for a note
func (h *ShareHandler) Create(c *fiber.Ctx) error {
	notePath := c.Params("*")
	// Decode URL-encoded path (for Chinese and other special characters)
	decodedPath, err := url.PathUnescape(notePath)
	if err == nil {
		notePath = decodedPath
	}

	// Security check - prevent directory traversal
	if !services.ValidatePathSecurity(h.config.Storage.NotesDir, notePath) {
		return c.Status(400).JSON(fiber.Map{"success": false, "detail": "Invalid note path"})
	}

	theme := c.Query("theme", "light")

	token, err := h.shareService.CreateShareToken(notePath, theme)
	if err != nil {
		return fiber.NewError(500, err.Error())
	}

	// Build full URL with protocol and host
	protocol := "http"
	if c.Protocol() == "https" {
		protocol = "https"
	}
	host := c.Hostname()
	fullURL := fmt.Sprintf("%s://%s/share/%s", protocol, host, token)

	return c.JSON(models.ShareCreateResponse{
		Success: true,
		Token:   token,
		URL:     fullURL,
		Path:    notePath,
		Theme:   theme,
	})
}

// GetStatus returns share status for a note
func (h *ShareHandler) GetStatus(c *fiber.Ctx) error {
	notePath := c.Params("*")
	// Decode URL-encoded path (for Chinese and other special characters)
	if decodedPath, err := url.PathUnescape(notePath); err == nil {
		notePath = decodedPath
	}

	// Security check - prevent directory traversal
	if !services.ValidatePathSecurity(h.config.Storage.NotesDir, notePath) {
		return c.Status(400).JSON(fiber.Map{"success": false, "detail": "Invalid note path"})
	}

	info, err := h.shareService.GetShareInfo(notePath)
	if err != nil {
		return fiber.NewError(500, err.Error())
	}

	if info.Shared {
		// Build full URL with protocol and host
		protocol := "http"
		if c.Protocol() == "https" {
			protocol = "https"
		}
		host := c.Hostname()
		info.URL = fmt.Sprintf("%s://%s/share/%s", protocol, host, info.Token)
	}

	return c.JSON(info)
}

// Revoke revokes a share token
func (h *ShareHandler) Revoke(c *fiber.Ctx) error {
	notePath := c.Params("*")
	// Decode URL-encoded path (for Chinese and other special characters)
	if decodedPath, err := url.PathUnescape(notePath); err == nil {
		notePath = decodedPath
	}

	// Security check - prevent directory traversal
	if !services.ValidatePathSecurity(h.config.Storage.NotesDir, notePath) {
		return c.Status(400).JSON(fiber.Map{"success": false, "detail": "Invalid note path"})
	}

	if err := h.shareService.RevokeShareToken(notePath); err != nil {
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(models.APIResponse{
		Success: true,
		Message: "Share token revoked",
	})
}

// ListSharedNotes returns all shared note paths
func (h *ShareHandler) ListSharedNotes(c *fiber.Ctx) error {
	paths, err := h.shareService.GetAllSharedPaths()
	if err != nil {
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(models.SharedNotesResponse{Paths: paths})
}

// ViewSharedNote renders a shared note
func (h *ShareHandler) ViewSharedNote(c *fiber.Ctx) error {
	token := c.Params("token")

	info, exists := h.shareService.GetNoteByToken(token)
	if !exists {
		return c.Status(404).SendString("Shared note not found")
	}

	// Get note content
	ns := services.NewNoteService(h.config.Storage.NotesDir)
	notePath := info.Path
	// Add .md extension if not present
	if filepath.Ext(notePath) == "" {
		notePath += ".md"
	}

	// Security check - validate path to prevent directory traversal
	if !services.ValidatePathSecurity(h.config.Storage.NotesDir, notePath) {
		return c.Status(400).SendString("Invalid note path")
	}

	content, err := ns.GetNoteContent(notePath)
	if err != nil {
		return c.Status(500).SendString("Failed to load note")
	}

	// Strip frontmatter
	content = services.StripFrontmatter(content)

	// Process media for export
	noteFolder := ""
	if idx := len(info.Path) - len(info.Path[strings.LastIndex(info.Path, "/")+1:]); idx > 0 {
		noteFolder = info.Path[:idx-1]
	}
	content = h.exportService.ProcessMediaForExport(content, noteFolder, h.config.Storage.NotesDir)

	// Get theme CSS
	ts := services.NewThemeService("themes")
	themeCSS, _ := ts.GetThemeCSS(info.Theme)

	// Determine if dark theme
	isDark := info.Theme != "light" && strings.Contains(strings.ToLower(info.Theme), "dark") ||
		info.Theme == "dracula" || info.Theme == "nord" || info.Theme == "monokai" ||
		info.Theme == "gruvbox-dark" || info.Theme == "cobalt2"

	// Generate HTML
	title := info.Path
	if idx := strings.LastIndex(info.Path, "/"); idx != -1 {
		title = info.Path[idx+1:]
	}
	title = strings.TrimSuffix(title, ".md")

	html := h.exportService.GenerateExportHTML(title, content, themeCSS, isDark)

	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(html)
}


