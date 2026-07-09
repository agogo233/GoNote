package handlers

import (
	"github.com/gofiber/fiber/v2"

	"gonote/internal/models"
	"gonote/internal/models/config"
	"gonote/internal/services"
	"gonote/internal/middleware"
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

// GetStatistics returns statistics for a note
func (h *StatisticsHandler) GetStatistics(c *fiber.Ctx) error {
	notePath, err := resolvePathParamTrimmed(c, h.config.Storage.NotesDir)
	if err != nil {
		return err
	}

	// Calculate statistics
	stats, err := h.service.CalculateStatistics(notePath)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get statistics")
	}

	return c.JSON(models.APIResponse{
		Success: true,
		Message: "ok",
		Data:    stats,
	})
}

func (h *StatisticsHandler) RegisterRoutes(api fiber.Router) {
	api.Get("/statistics/*", middleware.EndpointLimiterSimple(60), h.GetStatistics)
}
