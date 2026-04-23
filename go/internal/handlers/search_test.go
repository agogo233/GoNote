package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	"gonote/internal/models"
	"gonote/internal/models/config"
	"gonote/internal/services"
)

func TestNewSearchHandler(t *testing.T) {
	cfg := &config.Config{}
	searchService := services.NewSearchService("../data")
	handler := NewSearchHandler(searchService, cfg)

	assert.NotNil(t, handler)
	assert.Equal(t, searchService, handler.service)
	assert.Equal(t, cfg, handler.config)
	assert.False(t, handler.useIndex)
}

func TestNewSearchHandlerWithIndex(t *testing.T) {
	cfg := &config.Config{}
	searchService := services.NewSearchService("../data")
	noteService := services.NewNoteService("../data")
	searchIndex := services.NewSearchIndex("../data", noteService)
	handler := NewSearchHandlerWithIndex(searchService, searchIndex, cfg)

	assert.NotNil(t, handler)
	assert.Equal(t, searchService, handler.service)
	assert.Equal(t, searchIndex, handler.searchIndex)
	assert.True(t, handler.useIndex)
}

func TestSearchHandler_Search_EmptyQuery(t *testing.T) {
	cfg := &config.Config{}
	searchService := services.NewSearchService("../data")
	handler := NewSearchHandler(searchService, cfg)

	app := fiber.New()
	app.Get("/api/search", handler.Search)

	req := httptest.NewRequest("GET", "/api/search", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	assert.Empty(t, result["results"])
}

func TestSearchHandler_Search_WithQuery(t *testing.T) {
	cfg := &config.Config{}
	searchService := services.NewSearchService("../testdata")
	handler := NewSearchHandler(searchService, cfg)

	app := fiber.New()
	app.Get("/api/search", handler.Search)

	req := httptest.NewRequest("GET", "/api/search?q=test", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	assert.NotNil(t, result["results"])
}

func TestSearchHandler_Search_WithPagination(t *testing.T) {
	cfg := &config.Config{}
	searchService := services.NewSearchService("../testdata")
	handler := NewSearchHandler(searchService, cfg)

	app := fiber.New()
	app.Get("/api/search", handler.Search)

	req := httptest.NewRequest("GET", "/api/search?q=test&page=1&limit=10", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result models.SearchResultsResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	assert.NotNil(t, result.Results)
	assert.NotNil(t, result.Pagination)
}

func TestSearchHandler_Search_WithIndex(t *testing.T) {
	cfg := &config.Config{}
	searchService := services.NewSearchService("../testdata")
	noteService := services.NewNoteService("../testdata")
	searchIndex := services.NewSearchIndex("../testdata", noteService)

	// Build index
	_ = searchIndex.BuildIndex()

	handler := NewSearchHandlerWithIndex(searchService, searchIndex, cfg)

	app := fiber.New()
	app.Get("/api/search", handler.Search)

	req := httptest.NewRequest("GET", "/api/search?q=test", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	assert.NotNil(t, result["results"])
}
