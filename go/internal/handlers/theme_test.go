package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	"gonote/internal/models/config"
	"gonote/internal/services"
)

func TestNewThemeHandler(t *testing.T) {
	cfg := &config.Config{}
	themeService := services.NewThemeService("../shared/themes")
	handler := NewThemeHandler(themeService, cfg)

	assert.NotNil(t, handler)
	assert.Equal(t, themeService, handler.service)
	assert.Equal(t, cfg, handler.config)
}

func TestThemeHandler_List(t *testing.T) {
	// Create temp themes directory with test theme
	tmpDir := t.TempDir()
	themeCSS := "body { background: #000; }"
	err := os.MkdirAll(filepath.Join(tmpDir, "test-theme"), 0755)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "test-theme", "theme.css"), []byte(themeCSS), 0644)
	assert.NoError(t, err)

	themeService := services.NewThemeService(tmpDir)
	cfg := &config.Config{}
	handler := NewThemeHandler(themeService, cfg)

	app := fiber.New()
	app.Get("/api/themes", handler.List)

	req := httptest.NewRequest("GET", "/api/themes", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var themesResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&themesResp)
	assert.NoError(t, err)
	assert.NotNil(t, themesResp["themes"])
}

func TestThemeHandler_List_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	themeService := services.NewThemeService(tmpDir)
	cfg := &config.Config{}
	handler := NewThemeHandler(themeService, cfg)

	app := fiber.New()
	app.Get("/api/themes", handler.List)

	req := httptest.NewRequest("GET", "/api/themes", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var themesResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&themesResp)
	assert.NoError(t, err)
	themes := themesResp["themes"].([]interface{})
	assert.Empty(t, themes)
}

func TestThemeHandler_Get(t *testing.T) {
	// Create temp themes directory with test theme
	tmpDir := t.TempDir()
	themeCSS := "/* @theme-type: light */ body { background: #test; }"
	err := os.MkdirAll(filepath.Join(tmpDir, "custom-theme"), 0755)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "custom-theme", "theme.css"), []byte(themeCSS), 0644)
	assert.NoError(t, err)

	themeService := services.NewThemeService(tmpDir)
	cfg := &config.Config{}
	handler := NewThemeHandler(themeService, cfg)

	app := fiber.New()
	app.Get("/api/themes/:id", handler.Get)

	req := httptest.NewRequest("GET", "/api/themes/custom-theme", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var themeResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&themeResp)
	assert.NoError(t, err)
	assert.Equal(t, "custom-theme", themeResp["theme_id"])
}

func TestThemeHandler_Get_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	themeService := services.NewThemeService(tmpDir)
	cfg := &config.Config{}
	handler := NewThemeHandler(themeService, cfg)

	app := fiber.New()
	app.Get("/api/themes/:id", handler.Get)

	req := httptest.NewRequest("GET", "/api/themes/nonexistent", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err)
	// Returns 200 with empty CSS for not found
	assert.Equal(t, 200, resp.StatusCode)
}
