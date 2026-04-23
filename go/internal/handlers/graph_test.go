package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	"gonote/internal/models/config"
	"gonote/internal/services"
)

func TestNewGraphHandler(t *testing.T) {
	cfg := &config.Config{}
	graphService := services.NewGraphService("../data")

	handler := NewGraphHandler(graphService, cfg)

	assert.NotNil(t, handler)
	assert.Equal(t, graphService, handler.service)
	assert.Equal(t, cfg, handler.config)
}

func TestGraphHandler_Get(t *testing.T) {
	// Setup with empty data directory
	tmpDir := t.TempDir()
	cfg := &config.Config{}
	graphService := services.NewGraphService(tmpDir)
	handler := NewGraphHandler(graphService, cfg)

	app := fiber.New()
	app.Get("/api/graph", handler.Get)

	// Test
	req := httptest.NewRequest("GET", "/api/graph", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}
