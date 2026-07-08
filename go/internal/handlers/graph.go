package handlers

import (
	"github.com/gofiber/fiber/v2"

	"gonote/internal/models/config"
	"gonote/internal/services"
)

// GraphHandler handles graph-related requests
type GraphHandler struct {
	service *services.GraphService
	config  *config.Config
}

// NewGraphHandler creates a new GraphHandler
func NewGraphHandler(service *services.GraphService, cfg *config.Config) *GraphHandler {
	return &GraphHandler{service: service, config: cfg}
}

// Get returns the knowledge graph
func (h *GraphHandler) Get(c *fiber.Ctx) error {
	graph, err := h.service.GetGraph()
	if err != nil {
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(graph)
}

func (h *GraphHandler) RegisterRoutes(api fiber.Router) {
	api.Get("/graph", h.Get)
}
