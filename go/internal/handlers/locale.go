package handlers

import (
	"github.com/gofiber/fiber/v2"

	"gonote/internal/models/config"
	"gonote/internal/models"
	"gonote/internal/services"
)

// LocaleHandler handles locale-related requests
type LocaleHandler struct {
	service *services.LocaleService
	config  *config.Config
}

// NewLocaleHandler creates a new LocaleHandler
func NewLocaleHandler(service *services.LocaleService, cfg *config.Config) *LocaleHandler {
	return &LocaleHandler{service: service, config: cfg}
}

// List returns all locales
func (h *LocaleHandler) List(c *fiber.Ctx) error {
	locales, err := h.service.GetLocales()
	if err != nil {
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(models.LocalesResponse{Locales: locales})
}

// Get returns a locale's content
func (h *LocaleHandler) Get(c *fiber.Ctx) error {
	code := c.Params("code")

	content, err := h.service.GetLocaleContent(code)
	if err != nil {
		return fiber.NewError(500, err.Error())
	}

	if content == nil {
		return c.Status(404).JSON(fiber.Map{"detail": "Locale not found"})
	}

	return c.JSON(content)
}

func (h *LocaleHandler) RegisterRoutes(app *fiber.App) {
	app.Get("/api/locales", h.List)
	app.Get("/api/locales/:code", h.Get)
}
