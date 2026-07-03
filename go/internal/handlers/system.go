package handlers

import (
	"os"

	"github.com/gofiber/fiber/v2"

	"gonote/internal/models/config"
	"gonote/internal/models"
	"gonote/internal/services"
)

// SystemHandler handles system-related requests
type SystemHandler struct {
	config      *config.Config
	noteService *services.NoteService
	searchIndex *services.SearchIndex
}

// NewSystemHandler creates a new SystemHandler
func NewSystemHandler(cfg *config.Config) *SystemHandler {
	return &SystemHandler{config: cfg}
}

// SetReadinessDeps injects service dependencies needed for /readyz checks.
func (h *SystemHandler) SetReadinessDeps(noteService *services.NoteService, searchIndex *services.SearchIndex) {
	h.noteService = noteService
	h.searchIndex = searchIndex
}

// HealthCheck returns health status (liveness probe)
func (h *SystemHandler) HealthCheck(c *fiber.Ctx) error {
	return c.JSON(models.HealthResponse{
		Status:  "ok",
		App:     h.config.App.Name,
		Version: h.config.App.Version,
	})
}

// ReadinessCheck returns detailed readiness status (readiness probe).
// Checks: notes_dir writable, scanner initial scan completed, search index built.
func (h *SystemHandler) ReadinessCheck(c *fiber.Ctx) error {
	checks := make(map[string]string)
	overall := "ok"

	// 1. Notes dir writable
	if h.config != nil && h.config.Storage.NotesDir != "" {
		dir := h.config.Storage.NotesDir
		f, err := os.CreateTemp(dir, ".ready-*")
		if err != nil {
			checks["notes_dir"] = "unwritable: " + err.Error()
			overall = "degraded"
		} else {
			tmpPath := f.Name()
			f.Close()
			os.Remove(tmpPath)
			checks["notes_dir"] = "ok"
		}
	} else {
		checks["notes_dir"] = "not_configured"
		overall = "degraded"
	}

	// 2. Scanner initial scan completed
	if h.noteService != nil && h.noteService.IsScannerReady() {
		checks["scanner"] = "ok"
	} else if h.noteService != nil {
		checks["scanner"] = "initializing"
		overall = "degraded"
	} else {
		checks["scanner"] = "not_configured"
	}

	// 3. Search index built
	if h.searchIndex != nil && h.searchIndex.IsReady() {
		checks["search_index"] = "ok"
	} else if h.searchIndex != nil {
		checks["search_index"] = "building"
		overall = "degraded"
	} else {
		checks["search_index"] = "not_configured"
	}

	statusCode := 200
	if overall != "ok" {
		statusCode = 503
	}

	return c.Status(statusCode).JSON(models.ReadinessResponse{
		Status:  overall,
		App:     h.config.App.Name,
		Version: h.config.App.Version,
		Checks:  checks,
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
