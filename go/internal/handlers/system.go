package handlers

import (
	"os"

	"github.com/gofiber/fiber/v2"

	"gonote/internal/models/config"
	"gonote/internal/models"
)

// SystemHandler handles system-related requests
type SystemHandler struct {
	config *config.Config
}

// NewSystemHandler creates a new SystemHandler
func NewSystemHandler(cfg *config.Config) *SystemHandler {
	return &SystemHandler{config: cfg}
}

// HealthCheck returns health status
func (h *SystemHandler) HealthCheck(c *fiber.Ctx) error {
	return c.JSON(models.HealthResponse{
		Status:  "ok",
		App:     h.config.App.Name,
		Version: h.config.App.Version,
	})
}

// GetConfig returns public configuration
func (h *SystemHandler) GetConfig(c *fiber.Ctx) error {
	return c.JSON(models.ConfigResponse{
		Name:           h.config.App.Name,
		Version:        h.config.App.Version,
		SearchEnabled:  h.config.Search.Enabled,
		DemoMode:       config.DemoMode,
		AlreadyDonated: config.AlreadyDonated,
		Authentication: struct {
			Enabled bool `json:"enabled"`
		}{
			Enabled: h.config.Authentication.Enabled,
		},
	})
}

// ServiceWorker returns the service worker script
func (h *SystemHandler) ServiceWorker(c *fiber.Ctx) error {
	// Read service worker file
	swPath := "frontend/sw.js"
	content, err := os.ReadFile(swPath)
	if err != nil {
		// Return a minimal service worker
		c.Set("Content-Type", "application/javascript")
		return c.SendString(`
const CACHE_NAME = 'gonote-v1';
self.addEventListener('install', (e) => {
  self.skipWaiting();
});
self.addEventListener('activate', (e) => {
  e.waitUntil(clients.claim());
});
self.addEventListener('fetch', (e) => {
  e.respondWith(fetch(e.request));
});
`)
	}

	c.Set("Content-Type", "application/javascript")
	return c.Send(content)
}
