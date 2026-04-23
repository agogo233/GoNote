package handlers

import (
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"

	"gonote/internal/models/config"
	"gonote/internal/services"
)

// StatisticsHandler handles note statistics requests
type StatisticsHandler struct {
	service *services.StatisticsService
	config  *config.Config
}

// NewStatisticsHandler creates a new StatisticsHandler
func NewStatisticsHandler(service *services.StatisticsService, cfg *config.Config) *StatisticsHandler {
	return &StatisticsHandler{service: service, config: cfg}
}

// StatisticsResponse represents the response for statistics endpoint
type StatisticsResponse struct {
	Success bool                        `json:"success"`
	Data    *services.NoteStatistics    `json:"data"`
}

// GetStatistics returns statistics for a note
func (h *StatisticsHandler) GetStatistics(c *fiber.Ctx) error {
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

	// Calculate statistics
	stats, err := h.service.CalculateStatistics(notePath)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"detail": err.Error()})
	}

	return c.JSON(StatisticsResponse{
		Success: true,
		Data:    stats,
	})
}
