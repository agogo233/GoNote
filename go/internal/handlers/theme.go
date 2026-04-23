package handlers

import (
	"github.com/gofiber/fiber/v2"

	"gonote/internal/models/config"
	"gonote/internal/models"
	"gonote/internal/services"
)

// ThemeHandler handles theme-related requests
type ThemeHandler struct {
	service *services.ThemeService
	config  *config.Config
}

// NewThemeHandler creates a new ThemeHandler
func NewThemeHandler(service *services.ThemeService, cfg *config.Config) *ThemeHandler {
	return &ThemeHandler{service: service, config: cfg}
}

// List returns all themes
func (h *ThemeHandler) List(c *fiber.Ctx) error {
	themes, err := h.service.GetThemes()
	if err != nil {
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(models.ThemesResponse{Themes: themes})
}

// Get returns a theme's CSS
func (h *ThemeHandler) Get(c *fiber.Ctx) error {
	themeID := c.Params("id")

	css, err := h.service.GetThemeCSS(themeID)
	if err != nil {
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(models.ThemeResponse{
		CSS:     css,
		ThemeID: themeID,
	})
}
