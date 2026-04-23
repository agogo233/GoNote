package handlers

import (
	"bytes"
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

func TestNewTemplateHandler(t *testing.T) {
	cfg := &config.Config{}
	templateService := services.NewTemplateService("../data")
	handler := NewTemplateHandler(templateService, cfg)

	assert.NotNil(t, handler)
	assert.Equal(t, templateService, handler.service)
	assert.Equal(t, cfg, handler.config)
}

func TestTemplateHandler_List(t *testing.T) {
	// Create temp directory with test template
	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "_templates")
	err := os.MkdirAll(templateDir, 0755)
	assert.NoError(t, err)

	templateContent := `---
name: Test Template
---
# {{.title}}
`
	err = os.WriteFile(filepath.Join(templateDir, "test.md"), []byte(templateContent), 0644)
	assert.NoError(t, err)

	templateService := services.NewTemplateService(tmpDir)
	cfg := &config.Config{}
	handler := NewTemplateHandler(templateService, cfg)

	app := fiber.New()
	app.Get("/api/templates", handler.List)

	req := httptest.NewRequest("GET", "/api/templates", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var templatesResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&templatesResp)
	assert.NoError(t, err)
	assert.Contains(t, templatesResp, "templates")
}

func TestTemplateHandler_List_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	templateService := services.NewTemplateService(tmpDir)
	cfg := &config.Config{}
	handler := NewTemplateHandler(templateService, cfg)

	app := fiber.New()
	app.Get("/api/templates", handler.List)

	req := httptest.NewRequest("GET", "/api/templates", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var templatesResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&templatesResp)
	assert.NoError(t, err)
	templates := templatesResp["templates"].([]interface{})
	assert.Empty(t, templates)
}

func TestTemplateHandler_Get(t *testing.T) {
	// Create temp directory with test template
	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "_templates")
	err := os.MkdirAll(templateDir, 0755)
	assert.NoError(t, err)

	templateContent := `---
name: Test Template
---
# {{.title}}
Content here
`
	err = os.WriteFile(filepath.Join(templateDir, "test.md"), []byte(templateContent), 0644)
	assert.NoError(t, err)

	templateService := services.NewTemplateService(tmpDir)
	cfg := &config.Config{}
	handler := NewTemplateHandler(templateService, cfg)

	app := fiber.New()
	app.Get("/api/templates/*", handler.Get)

	req := httptest.NewRequest("GET", "/api/templates/test", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var templateResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&templateResp)
	assert.NoError(t, err)
	assert.Equal(t, "test", templateResp["name"])
	assert.Contains(t, templateResp["content"].(string), "Content here")
}

func TestTemplateHandler_Get_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	templateService := services.NewTemplateService(tmpDir)
	cfg := &config.Config{}
	handler := NewTemplateHandler(templateService, cfg)

	app := fiber.New()
	app.Get("/api/templates/*", handler.Get)

	req := httptest.NewRequest("GET", "/api/templates/nonexistent", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
}

func TestTemplateHandler_CreateFromTemplate(t *testing.T) {
	// Create temp directory with test template and notes dir
	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "_templates")
	err := os.MkdirAll(templateDir, 0755)
	assert.NoError(t, err)

	templateContent := `---
name: Test Template
---
# {{.title}}
`
	err = os.WriteFile(filepath.Join(templateDir, "test.md"), []byte(templateContent), 0644)
	assert.NoError(t, err)

	notesDir := filepath.Join(tmpDir, "notes")
	err = os.MkdirAll(notesDir, 0755)
	assert.NoError(t, err)

	templateService := services.NewTemplateService(tmpDir)
	cfg := &config.Config{
		Storage: config.StorageConfig{
			NotesDir: notesDir,
		},
	}
	handler := NewTemplateHandler(templateService, cfg)

	app := fiber.New()
	app.Post("/api/templates/create", handler.CreateFromTemplate)

	body := map[string]string{
		"templateName": "test",
		"notePath":     "new-note.md",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/templates/create", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var respBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, true, respBody["success"])
}

func TestTemplateHandler_CreateFromTemplate_InvalidBody(t *testing.T) {
	tmpDir := t.TempDir()
	templateService := services.NewTemplateService(tmpDir)
	cfg := &config.Config{}
	handler := NewTemplateHandler(templateService, cfg)

	app := fiber.New()
	app.Post("/api/templates/create", handler.CreateFromTemplate)

	req := httptest.NewRequest("POST", "/api/templates/create", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)

	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}
