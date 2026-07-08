package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	"gonote/internal/models"
	"gonote/internal/models/config"
)

func TestNewSystemHandler(t *testing.T) {
	cfg := &config.Config{}
	handler := NewSystemHandler(cfg, "")

	assert.NotNil(t, handler)
	assert.Equal(t, cfg, handler.config)
}

func TestSystemHandler_HealthCheck(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Name:    "GoNote",
			Version: "0.25",
		},
	}
	handler := NewSystemHandler(cfg, "")

	app := fiber.New()
	app.Get("/health", handler.HealthCheck)

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var health models.HealthResponse
	err = json.NewDecoder(resp.Body).Decode(&health)
	assert.NoError(t, err)
	assert.Equal(t, "ok", health.Status)
	assert.Equal(t, "GoNote", health.App)
	assert.Equal(t, "0.25", health.Version)
}

func TestSystemHandler_GetConfig(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Name:    "GoNote",
			Version: "0.25",
		},
		Search: config.SearchConfig{
			Enabled: true,
		},
		Authentication: config.AuthConfig{
			Enabled: true,
		},
	}
	handler := NewSystemHandler(cfg, "")

	app := fiber.New()
	app.Get("/api/config", handler.GetConfig)

	req := httptest.NewRequest("GET", "/api/config", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var configResp models.ConfigResponse
	err = json.NewDecoder(resp.Body).Decode(&configResp)
	assert.NoError(t, err)
	assert.Equal(t, "GoNote", configResp.Name)
	assert.Equal(t, "0.25", configResp.Version)
	assert.True(t, configResp.SearchEnabled)
	assert.True(t, configResp.Authentication.Enabled)
}

func TestSystemHandler_GetConfig_AuthDisabled(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Name:    "GoNote",
			Version: "0.25",
		},
		Authentication: config.AuthConfig{
			Enabled: false,
		},
	}
	handler := NewSystemHandler(cfg, "")

	app := fiber.New()
	app.Get("/api/config", handler.GetConfig)

	req := httptest.NewRequest("GET", "/api/config", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var configResp models.ConfigResponse
	err = json.NewDecoder(resp.Body).Decode(&configResp)
	assert.NoError(t, err)
	assert.False(t, configResp.Authentication.Enabled)
}

func TestSystemHandler_ServiceWorker_FileExists(t *testing.T) {
	tmpDir := t.TempDir()
	swPath := filepath.Join(tmpDir, "frontend")
	err := os.MkdirAll(swPath, 0755)
	assert.NoError(t, err)

	swContent := "// Test service worker"
	err = os.WriteFile(filepath.Join(swPath, "sw.js"), []byte(swContent), 0644)
	assert.NoError(t, err)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	cfg := &config.Config{}
	handler := NewSystemHandler(cfg, filepath.Join(tmpDir, "frontend"))

	app := fiber.New()
	app.Get("/sw.js", handler.ServiceWorker)

	req := httptest.NewRequest("GET", "/sw.js", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "application/javascript")
}

func TestSystemHandler_ServiceWorker_FileNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	cfg := &config.Config{}
	handler := NewSystemHandler(cfg, tmpDir)

	app := fiber.New()
	app.Get("/sw.js", handler.ServiceWorker)

	req := httptest.NewRequest("GET", "/sw.js", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "application/javascript")
}