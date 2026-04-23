package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gonote/internal/models/config"
	"gonote/internal/services"
)

func setupFolderApp(t *testing.T) (*fiber.App, *config.Config, string) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Storage: config.StorageConfig{
			NotesDir: tmpDir,
		},
	}
	handler := NewFolderHandler(cfg)

	app := fiber.New()
	app.Post("/api/folders", handler.Create)
	app.Delete("/api/folders/*", handler.Delete)
	app.Put("/api/folders/move", handler.Move)
	app.Put("/api/folders/rename", handler.Rename)

	return app, cfg, tmpDir
}

func TestFolderCreate(t *testing.T) {
	app, _, _ := setupFolderApp(t)

	body := `{"path": "new-folder"}`
	req := httptest.NewRequest("POST", "/api/folders", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &result)

	assert.True(t, result["success"].(bool))
	assert.Equal(t, "new-folder", result["path"])
}

func TestFolderCreateNested(t *testing.T) {
	app, _, tmpDir := setupFolderApp(t)

	body := `{"path": "parent/child/grandchild"}`
	req := httptest.NewRequest("POST", "/api/folders", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Verify all directories were created
	_, err = os.Stat(filepath.Join(tmpDir, "parent", "child", "grandchild"))
	assert.NoError(t, err)
}

func TestFolderCreateInvalidPath(t *testing.T) {
	app, _, _ := setupFolderApp(t)

	body := `{"path": "../../../etc/passwd"}`
	req := httptest.NewRequest("POST", "/api/folders", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestFolderCreateInvalidBody(t *testing.T) {
	app, _, _ := setupFolderApp(t)

	body := `invalid json`
	req := httptest.NewRequest("POST", "/api/folders", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestFolderDelete(t *testing.T) {
	app, _, tmpDir := setupFolderApp(t)

	// Create folder first
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "to-delete"), 0755))

	req := httptest.NewRequest("DELETE", "/api/folders/to-delete", nil)

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Verify folder is deleted
	_, err = os.Stat(filepath.Join(tmpDir, "to-delete"))
	assert.True(t, os.IsNotExist(err))
}

func TestFolderDeleteWithContent(t *testing.T) {
	app, _, tmpDir := setupFolderApp(t)

	// Create folder with content
	folderPath := filepath.Join(tmpDir, "folder-with-content")
	require.NoError(t, os.MkdirAll(folderPath, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(folderPath, "note.md"), []byte("content"), 0644))

	req := httptest.NewRequest("DELETE", "/api/folders/folder-with-content", nil)

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Verify folder and content are deleted
	_, err = os.Stat(folderPath)
	assert.True(t, os.IsNotExist(err))
}

func TestFolderDeleteNonExistent(t *testing.T) {
	app, _, _ := setupFolderApp(t)

	req := httptest.NewRequest("DELETE", "/api/folders/nonexistent", nil)

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)
}

func TestFolderDeleteInvalidPath(t *testing.T) {
	app, _, _ := setupFolderApp(t)

	req := httptest.NewRequest("DELETE", "/api/folders/../../../etc", nil)

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestFolderDeleteWithSpecialChars(t *testing.T) {
	app, _, tmpDir := setupFolderApp(t)

	// Create folder with Chinese characters
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "测试文件夹"), 0755))

	req := httptest.NewRequest("DELETE", "/api/folders/%E6%B5%8B%E8%AF%95%E6%96%87%E4%BB%B6%E5%A4%B9", nil)

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestFolderMove(t *testing.T) {
	app, _, tmpDir := setupFolderApp(t)

	// Create source folder
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "old-folder"), 0755))

	body := `{"oldPath": "old-folder", "newPath": "new-folder"}`
	req := httptest.NewRequest("PUT", "/api/folders/move", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Verify old folder is gone
	_, err = os.Stat(filepath.Join(tmpDir, "old-folder"))
	assert.True(t, os.IsNotExist(err))

	// Verify new folder exists
	_, err = os.Stat(filepath.Join(tmpDir, "new-folder"))
	assert.NoError(t, err)
}

func TestFolderMoveToSubdirectory(t *testing.T) {
	app, _, tmpDir := setupFolderApp(t)

	// Create source folder with content
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "folder"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "folder", "note.md"), []byte("content"), 0644))

	body := `{"oldPath": "folder", "newPath": "parent/folder"}`
	req := httptest.NewRequest("PUT", "/api/folders/move", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Verify moved with content
	data, err := os.ReadFile(filepath.Join(tmpDir, "parent", "folder", "note.md"))
	require.NoError(t, err)
	assert.Equal(t, "content", string(data))
}

func TestFolderMoveInvalidOldPath(t *testing.T) {
	app, _, _ := setupFolderApp(t)

	body := `{"oldPath": "../../../etc", "newPath": "new-folder"}`
	req := httptest.NewRequest("PUT", "/api/folders/move", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestFolderMoveInvalidNewPath(t *testing.T) {
	app, _, tmpDir := setupFolderApp(t)

	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "folder"), 0755))

	body := `{"oldPath": "folder", "newPath": "../../../etc/malicious"}`
	req := httptest.NewRequest("PUT", "/api/folders/move", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestFolderMoveNonExistentSource(t *testing.T) {
	app, _, _ := setupFolderApp(t)

	body := `{"oldPath": "nonexistent", "newPath": "new-folder"}`
	req := httptest.NewRequest("PUT", "/api/folders/move", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)
}

func TestFolderMoveDestinationExists(t *testing.T) {
	app, _, tmpDir := setupFolderApp(t)

	// Create both source and destination
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "source"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "dest"), 0755))

	body := `{"oldPath": "source", "newPath": "dest"}`
	req := httptest.NewRequest("PUT", "/api/folders/move", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)
}

func TestFolderRename(t *testing.T) {
	app, _, tmpDir := setupFolderApp(t)

	// Create folder
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "old-name"), 0755))

	body := `{"oldPath": "old-name", "newPath": "new-name"}`
	req := httptest.NewRequest("PUT", "/api/folders/rename", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Verify rename
	_, err = os.Stat(filepath.Join(tmpDir, "new-name"))
	assert.NoError(t, err)
}

func TestFolderRenameInvalidBody(t *testing.T) {
	app, _, _ := setupFolderApp(t)

	body := `invalid json`
	req := httptest.NewRequest("PUT", "/api/folders/rename", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestFolderHandlerWithCacheInvalidator(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Storage: config.StorageConfig{
			NotesDir: tmpDir,
		},
	}

	// Mock cache invalidator
	invalidated := false
	mockCache := &mockCacheInvalidator{
		invalidateFunc: func() {
			invalidated = true
		},
	}

	handler := NewFolderHandlerWithCache(cfg, mockCache)

	app := fiber.New()
	app.Post("/api/folders", handler.Create)

	body := `{"path": "test-folder"}`
	req := httptest.NewRequest("POST", "/api/folders", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.True(t, invalidated, "Cache should be invalidated after folder creation")
}

// Mock cache invalidator for testing
type mockCacheInvalidator struct {
	invalidateFunc func()
}

func (m *mockCacheInvalidator) InvalidateCache() {
	if m.invalidateFunc != nil {
		m.invalidateFunc()
	}
}

// Test folder service functions directly
func TestCreateFolder(t *testing.T) {
	tmpDir := t.TempDir()

	err := services.CreateFolder(tmpDir, "new-folder")

	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(tmpDir, "new-folder"))
	assert.NoError(t, err)
}

func TestCreateFolderNested(t *testing.T) {
	tmpDir := t.TempDir()

	err := services.CreateFolder(tmpDir, "parent/child")

	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(tmpDir, "parent", "child"))
	assert.NoError(t, err)
}

func TestCreateFolderInvalidPath(t *testing.T) {
	tmpDir := t.TempDir()

	err := services.CreateFolder(tmpDir, "../../../etc")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid path")
}

func TestDeleteFolder(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "to-delete"), 0755))

	err := services.DeleteFolder(tmpDir, "to-delete")

	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(tmpDir, "to-delete"))
	assert.True(t, os.IsNotExist(err))
}

func TestDeleteFolderNonExistent(t *testing.T) {
	tmpDir := t.TempDir()

	err := services.DeleteFolder(tmpDir, "nonexistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestDeleteFolderIsFile(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("content"), 0644))

	err := services.DeleteFolder(tmpDir, "file.txt")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a directory")
}

func TestMoveFolder(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "source"), 0755))

	err := services.MoveFolder(tmpDir, "source", "destination")

	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(tmpDir, "destination"))
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(tmpDir, "source"))
	assert.True(t, os.IsNotExist(err))
}

func TestMoveFolderWithContent(t *testing.T) {
	tmpDir := t.TempDir()
	sourcePath := filepath.Join(tmpDir, "source")
	require.NoError(t, os.MkdirAll(sourcePath, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(sourcePath, "note.md"), []byte("content"), 0644))

	err := services.MoveFolder(tmpDir, "source", "moved")

	assert.NoError(t, err)
	data, err := os.ReadFile(filepath.Join(tmpDir, "moved", "note.md"))
	require.NoError(t, err)
	assert.Equal(t, "content", string(data))
}
