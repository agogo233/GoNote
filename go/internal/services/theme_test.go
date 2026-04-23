package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewThemeService(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewThemeService(tmpDir)

	assert.NotNil(t, service)
	assert.Equal(t, tmpDir, service.themesDir)
}

func TestThemeService_GetThemes(t *testing.T) {
	// Create temp directory with test themes
	tmpDir := t.TempDir()

	// Create test theme files
	themes := map[string]string{
		"light.css": `/* @theme-type: light */ body { background: #fff; }`,
		"dark.css":  `/* @theme-type: dark */ body { background: #000; }`,
		"custom.css": `body { background: #123; }`,
	}

	for name, content := range themes {
		err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644)
		assert.NoError(t, err)
	}

	service := NewThemeService(tmpDir)
	themeList, err := service.GetThemes()

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(themeList), 3)

	// Check sorting
	for i := 0; i < len(themeList)-1; i++ {
		assert.LessOrEqual(t, themeList[i].Name, themeList[i+1].Name)
	}
}

func TestThemeService_GetThemes_NoThemesDir(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewThemeService(tmpDir)

	themes, err := service.GetThemes()
	assert.NoError(t, err)
	assert.Empty(t, themes)
}

func TestThemeService_GetThemes_EmptyThemesDir(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.MkdirAll(tmpDir, 0755)
	assert.NoError(t, err)

	service := NewThemeService(tmpDir)
	themes, err := service.GetThemes()

	assert.NoError(t, err)
	assert.Empty(t, themes)
}

func TestThemeService_GetThemes_WithIcons(t *testing.T) {
	tmpDir := t.TempDir()

	// Create themes with known icons
	themes := map[string]string{
		"dracula.css": `body { background: #282a36; }`,
		"nord.css":    `body { background: #2e3440; }`,
		"monokai.css": `body { background: #272822; }`,
	}

	for name, content := range themes {
		err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644)
		assert.NoError(t, err)
	}

	service := NewThemeService(tmpDir)
	themeList, err := service.GetThemes()

	assert.NoError(t, err)
	assert.Equal(t, 3, len(themeList))

	// Check icons
	for _, theme := range themeList {
		switch theme.ID {
		case "dracula":
			assert.Contains(t, theme.Name, "🧛")
		case "nord":
			assert.Contains(t, theme.Name, "❄️")
		case "monokai":
			assert.Contains(t, theme.Name, "🎞️")
		}
	}
}

func TestThemeService_ParseThemeType_Light(t *testing.T) {
	tmpDir := t.TempDir()
	themeContent := `/* @theme-type: light */
body { background: #fff; }`

	themePath := filepath.Join(tmpDir, "test.css")
	err := os.WriteFile(themePath, []byte(themeContent), 0644)
	assert.NoError(t, err)

	service := NewThemeService(tmpDir)
	themeType := service.parseThemeType(themePath)

	assert.Equal(t, "light", themeType)
}

func TestThemeService_ParseThemeType_Dark(t *testing.T) {
	tmpDir := t.TempDir()
	themeContent := `/* @theme-type: dark */
body { background: #000; }`

	themePath := filepath.Join(tmpDir, "test.css")
	err := os.WriteFile(themePath, []byte(themeContent), 0644)
	assert.NoError(t, err)

	service := NewThemeService(tmpDir)
	themeType := service.parseThemeType(themePath)

	assert.Equal(t, "dark", themeType)
}

func TestThemeService_ParseThemeType_Default(t *testing.T) {
	tmpDir := t.TempDir()
	themeContent := `body { background: #000; }`

	themePath := filepath.Join(tmpDir, "test.css")
	err := os.WriteFile(themePath, []byte(themeContent), 0644)
	assert.NoError(t, err)

	service := NewThemeService(tmpDir)
	themeType := service.parseThemeType(themePath)

	assert.Equal(t, "dark", themeType) // Default
}

func TestThemeService_ParseThemeType_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewThemeService(tmpDir)

	themeType := service.parseThemeType(filepath.Join(tmpDir, "nonexistent.css"))
	assert.Equal(t, "dark", themeType) // Default
}

func TestThemeService_GetThemeCSS(t *testing.T) {
	tmpDir := t.TempDir()
	themeContent := `body { background: #test; }`

	err := os.WriteFile(filepath.Join(tmpDir, "test.css"), []byte(themeContent), 0644)
	assert.NoError(t, err)

	service := NewThemeService(tmpDir)
	css, err := service.GetThemeCSS("test")

	assert.NoError(t, err)
	assert.Contains(t, css, "background")
}

func TestThemeService_GetThemeCSS_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewThemeService(tmpDir)

	css, err := service.GetThemeCSS("nonexistent")
	assert.NoError(t, err)
	assert.Empty(t, css)
}

func TestThemeService_GetThemeCSS_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "empty.css"), []byte(""), 0644)
	assert.NoError(t, err)

	service := NewThemeService(tmpDir)
	css, err := service.GetThemeCSS("empty")

	assert.NoError(t, err)
	assert.Empty(t, css)
}
