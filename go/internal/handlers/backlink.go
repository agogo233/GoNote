package handlers

import (
	"github.com/gofiber/fiber/v2"

	"gonote/internal/models/config"
	"gonote/internal/services"
	"gonote/internal/middleware"
)

// BacklinkHandler handles backlink-related requests
type BacklinkHandler struct {
	config         *config.Config
	backlinkService *services.BacklinkService
}

// NewBacklinkHandler creates a new BacklinkHandler
func NewBacklinkHandler(cfg *config.Config, linkIndex *services.LinkIndex) *BacklinkHandler {
	return &BacklinkHandler{
		config:          cfg,
		backlinkService: services.NewBacklinkService(cfg.Storage.NotesDir, linkIndex),
	}
}

// GetBacklinks returns backlinks for a note
func (h *BacklinkHandler) GetBacklinks(c *fiber.Ctx) error {
	notePath, ok := resolvePathParamTrimmed(c, h.config.Storage.NotesDir)
	if !ok {
		return nil
	}

	backlinks, err := h.backlinkService.FindBacklinks(notePath)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"detail": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"backlinks": backlinks,
		"count":     len(backlinks),
	})
}

func (h *BacklinkHandler) RegisterRoutes(api fiber.Router) {
	api.Get("/backlinks/*", middleware.EndpointLimiterSimple(60), h.GetBacklinks)
}