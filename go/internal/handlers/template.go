package handlers

import (
	"github.com/gofiber/fiber/v2"

	"gonote/internal/models/config"
	"gonote/internal/models"
	"gonote/internal/services"
)

// TemplateHandler handles template-related requests
type TemplateHandler struct {
	service *services.TemplateService
	config  *config.Config
}

// NewTemplateHandler creates a new TemplateHandler
func NewTemplateHandler(service *services.TemplateService, cfg *config.Config) *TemplateHandler {
	return &TemplateHandler{service: service, config: cfg}
}

// List returns all templates
func (h *TemplateHandler) List(c *fiber.Ctx) error {
	templates, err := h.service.GetTemplates()
	if err != nil {
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(models.TemplatesResponse{Templates: templates})
}

// Get returns a template's content
func (h *TemplateHandler) Get(c *fiber.Ctx) error {
	templateName := c.Params("*")

	content, err := h.service.GetTemplateContent(templateName)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"detail": "Template not found"})
	}

	return c.JSON(fiber.Map{
		"name":    templateName,
		"content": content,
	})
}

// CreateFromTemplate creates a new note from a template
func (h *TemplateHandler) CreateFromTemplate(c *fiber.Ctx) error {
	var req struct {
		TemplateName string `json:"templateName"`
		NotePath     string `json:"notePath"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"detail": "Invalid request body"})
	}

	notePath, err := h.service.CreateNoteFromTemplate(req.TemplateName, req.NotePath)
	if err != nil {
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(fiber.Map{
		"success":  true,
		"notePath": notePath,
		"message":  "Note created from template",
	})
}
