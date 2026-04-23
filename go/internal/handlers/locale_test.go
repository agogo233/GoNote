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
	"gonote/internal/services"
)

func TestNewLocaleHandler(t *testing.T) {
	cfg := &config.Config{}
	localeService := services.NewLocaleService("../shared/locales")
	handler := NewLocaleHandler(localeService, cfg)

	assert.NotNil(t, handler)
	assert.Equal(t, localeService, handler.service)
	assert.Equal(t, cfg, handler.config)
}

func TestLocaleHandler_List(t *testing.T) {
	// Create temp locales directory with test files
	tmpDir := t.TempDir()
	localeContent := `{"hello": "Hello"}`
	err := os.WriteFile(filepath.Join(tmpDir, "en.json"), []byte(localeContent), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "zh-CN.json"), []byte(localeContent), 0644)
	assert.NoError(t, err)

	localeService := services.NewLocaleService(tmpDir)
	cfg := &config.Config{}
	handler := NewLocaleHandler(localeService, cfg)

	app := fiber.New()
	app.Get("/api/locales", handler.List)

	req := httptest.NewRequest("GET", "/api/locales", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var localesResp models.LocalesResponse
	err = json.NewDecoder(resp.Body).Decode(&localesResp)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(localesResp.Locales), 2)
}

func TestLocaleHandler_List_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	localeService := services.NewLocaleService(tmpDir)
	cfg := &config.Config{}
	handler := NewLocaleHandler(localeService, cfg)

	app := fiber.New()
	app.Get("/api/locales", handler.List)

	req := httptest.NewRequest("GET", "/api/locales", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var localesResp models.LocalesResponse
	err = json.NewDecoder(resp.Body).Decode(&localesResp)
	assert.NoError(t, err)
	assert.Empty(t, localesResp.Locales)
}

func TestLocaleHandler_Get(t *testing.T) {
	// Create temp locales directory with test file
	tmpDir := t.TempDir()
	localeContent := `{"hello": "Hello", "goodbye": "Goodbye"}`
	err := os.WriteFile(filepath.Join(tmpDir, "en.json"), []byte(localeContent), 0644)
	assert.NoError(t, err)

	localeService := services.NewLocaleService(tmpDir)
	cfg := &config.Config{}
	handler := NewLocaleHandler(localeService, cfg)

	app := fiber.New()
	app.Get("/api/locales/:code", handler.Get)

	req := httptest.NewRequest("GET", "/api/locales/en", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var locale map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&locale)
	assert.NoError(t, err)
	assert.Equal(t, "Hello", locale["hello"])
	assert.Equal(t, "Goodbye", locale["goodbye"])
}

func TestLocaleHandler_Get_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	localeService := services.NewLocaleService(tmpDir)
	cfg := &config.Config{}
	handler := NewLocaleHandler(localeService, cfg)

	app := fiber.New()
	app.Get("/api/locales/:code", handler.Get)

	req := httptest.NewRequest("GET", "/api/locales/nonexistent", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
}
