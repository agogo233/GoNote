package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLocaleService(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewLocaleService(tmpDir)

	assert.NotNil(t, service)
	assert.Equal(t, tmpDir, service.localesDir)
}

func TestLocaleService_GetLocales(t *testing.T) {
	// Create temp directory with test locale files
	tmpDir := t.TempDir()

	locales := map[string]string{
		"en.json":    `{"hello": "Hello"}`,
		"zh-CN.json": `{"hello": "你好"}`,
		"ja.json":    `{"hello": "こんにちは"}`,
	}

	for name, content := range locales {
		err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644)
		assert.NoError(t, err)
	}

	service := NewLocaleService(tmpDir)
	result, err := service.GetLocales()

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(result), 3)

	// Check sorting (should be alphabetical by code)
	assert.Equal(t, "en", result[0].Code)
	assert.Equal(t, "ja", result[1].Code)
	assert.Equal(t, "zh-CN", result[2].Code)
}

func TestLocaleService_GetLocales_NoDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewLocaleService(tmpDir)

	result, err := service.GetLocales()
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestLocaleService_GetLocales_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.MkdirAll(tmpDir, 0755)
	assert.NoError(t, err)

	service := NewLocaleService(tmpDir)
	result, err := service.GetLocales()

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestLocaleService_GetLocales_WithFlags(t *testing.T) {
	tmpDir := t.TempDir()

	// Create locale files with known flags
	locales := map[string]string{
		"en.json": `{"hello": "Hello"}`,
		"de.json": `{"hello": "Hallo"}`,
		"fr.json": `{"hello": "Bonjour"}`,
	}

	for name, content := range locales {
		err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644)
		assert.NoError(t, err)
	}

	service := NewLocaleService(tmpDir)
	result, err := service.GetLocales()

	assert.NoError(t, err)

	// Check flags
	for _, locale := range result {
		switch locale.Code {
		case "en":
			assert.Equal(t, "🇺🇸", locale.Flag)
		case "de":
			assert.Equal(t, "🇩🇪", locale.Flag)
		case "fr":
			assert.Equal(t, "🇫🇷", locale.Flag)
		}
	}
}

func TestLocaleService_GetLocales_UnknownCode(t *testing.T) {
	tmpDir := t.TempDir()

	// Create locale file with unknown code
	content := `{"hello": "Hello"}`
	err := os.WriteFile(filepath.Join(tmpDir, "xx.json"), []byte(content), 0644)
	assert.NoError(t, err)

	service := NewLocaleService(tmpDir)
	result, err := service.GetLocales()

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(result), 1)

	// Unknown code should have default flag
	for _, locale := range result {
		if locale.Code == "xx" {
			assert.Equal(t, "🌐", locale.Flag)
		}
	}
}

func TestLocaleService_GetLocaleContent(t *testing.T) {
	tmpDir := t.TempDir()

	localeContent := `{"hello": "Hello", "goodbye": "Goodbye", "count": 42}`
	err := os.WriteFile(filepath.Join(tmpDir, "en.json"), []byte(localeContent), 0644)
	assert.NoError(t, err)

	service := NewLocaleService(tmpDir)
	content, err := service.GetLocaleContent("en")

	assert.NoError(t, err)
	assert.NotNil(t, content)
	assert.Equal(t, "Hello", content["hello"])
	assert.Equal(t, "Goodbye", content["goodbye"])
	assert.Equal(t, float64(42), content["count"]) // JSON numbers are float64
}

func TestLocaleService_GetLocaleContent_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewLocaleService(tmpDir)

	content, err := service.GetLocaleContent("nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, content)
}

func TestLocaleService_GetLocaleContent_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()

	// Write invalid JSON
	err := os.WriteFile(filepath.Join(tmpDir, "invalid.json"), []byte("{invalid json}"), 0644)
	assert.NoError(t, err)

	service := NewLocaleService(tmpDir)
	content, err := service.GetLocaleContent("invalid")

	assert.Error(t, err)
	assert.Nil(t, content)
}

func TestLocaleService_GetLocaleContent_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()

	err := os.WriteFile(filepath.Join(tmpDir, "empty.json"), []byte(""), 0644)
	assert.NoError(t, err)

	service := NewLocaleService(tmpDir)
	content, err := service.GetLocaleContent("empty")

	assert.Error(t, err)
	assert.Nil(t, content)
}
