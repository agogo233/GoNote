package handlers

import (
	"github.com/gofiber/fiber/v2"

	"gonote/internal/models/config"
	"gonote/internal/models"
	"gonote/internal/services"
)

// TagHandler handles tag-related requests
type TagHandler struct {
	service *services.TagService
	config  *config.Config
}

// NewTagHandler creates a new TagHandler
func NewTagHandler(service *services.TagService, cfg *config.Config) *TagHandler {
	return &TagHandler{service: service, config: cfg}
}

// List returns all tags with counts
func (h *TagHandler) List(c *fiber.Ctx) error {
	tags, err := h.service.GetAllTags()
	if err != nil {
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(models.TagsResponse{Tags: tags})
}

// GetNotesByTag returns all notes with a specific tag
func (h *TagHandler) GetNotesByTag(c *fiber.Ctx) error {
	tag := c.Params("*")

	notes, err := h.service.GetNotesByTag(tag)
	if err != nil {
		return fiber.NewError(500, err.Error())
	}

	// Ensure notes is never nil to avoid JSON null in response
	if notes == nil {
		notes = []models.Note{}
	}

	return c.JSON(models.TagNotesResponse{
		Tag:   tag,
		Count: len(notes),
		Notes: notes,
	})
}

func (h *TagHandler) RegisterRoutes(api fiber.Router) {
	api.Get("/tags", h.List)
	api.Get("/tags/*", h.GetNotesByTag)
}
