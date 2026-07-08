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
	notePath, ok := resolvePathParamTrimmed(c, h.config.Storage.NotesDir)
	if !ok {
		return nil
	}

	// Calculate statistics
	stats, err := h.service.CalculateStatistics(notePath)
	if err != nil {
		return c.Status(500).JSON(models.APIResponse{Success: false, Message: err.Error()})
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
