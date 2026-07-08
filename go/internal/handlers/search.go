package handlers

import (
	"github.com/gofiber/fiber/v2"

	"gonote/internal/models/config"
	"gonote/internal/models"
	"gonote/internal/services"
)

// SearchHandler handles search-related requests
type SearchHandler struct {
	service     *services.SearchService
	searchIndex *services.SearchIndex
	config      *config.Config
	useIndex    bool
}

// NewSearchHandler creates a new SearchHandler
func NewSearchHandler(service *services.SearchService, cfg *config.Config) *SearchHandler {
	return &SearchHandler{service: service, config: cfg, useIndex: false}
}

// NewSearchHandlerWithIndex creates a new SearchHandler with search index support
func NewSearchHandlerWithIndex(service *services.SearchService, searchIndex *services.SearchIndex, cfg *config.Config) *SearchHandler {
	return &SearchHandler{
		service:     service,
		searchIndex: searchIndex,
		config:      cfg,
		useIndex:    searchIndex != nil,
	}
}

// Search performs a full-text search
func (h *SearchHandler) Search(c *fiber.Ctx) error {
	query := c.Query("q", "")
	if query == "" {
		return c.JSON(fiber.Map{"results": []models.SearchResult{}})
	}

	// Get search mode: full (default), title, smart
	mode := c.Query("mode", "full")

	// Get pagination parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 50)

	var results []models.SearchResult
	var err error

	// Use indexed search if available, otherwise fall back to full scan
	if h.useIndex && h.searchIndex != nil {
		switch mode {
		case "title":
			results, err = h.searchIndex.SearchByTitle(query)
		case "smart":
			results, err = h.searchIndex.SearchSmart(query)
		default: // "full"
			results, err = h.searchIndex.Search(query)
		}
	} else {
		// Fallback: full scan search (mode not supported in legacy search)
		results, err = h.service.Search(query)
	}

	if err != nil {
		return fiber.NewError(500, err.Error())
	}

	// Always use pagination
	if limit <= 0 || limit > 200 {
		limit = 200
	}
	paginatedResult := services.PaginateSearchResults(results, page, limit)
	return c.JSON(models.SearchResultsResponse(paginatedResult))
}

func (h *SearchHandler) RegisterRoutes(api fiber.Router) {
	api.Get("/search", h.Search)
}
