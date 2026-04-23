package services

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"gonote/internal/models"
)

// ThemeService handles theme operations
type ThemeService struct {
	themesDir string
}

// NewThemeService creates a new ThemeService
func NewThemeService(themesDir string) *ThemeService {
	return &ThemeService{themesDir: themesDir}
}

// Theme icons mapping
var themeIcons = map[string]string{
	"light":             "🌞",
	"dark":              "🌙",
	"dracula":           "🧛",
	"nord":              "❄️",
	"monokai":           "🎞️",
	"vue-high-contrast": "💚",
	"cobalt2":           "🌊",
	"vs-blue":           "🔷",
	"gruvbox-dark":      "🟫",
	"matcha-light":      "🍵",
}

// GetThemes returns all available themes
func (s *ThemeService) GetThemes() ([]models.Theme, error) {
	themes := []models.Theme{}

	// Check if themes directory exists
	if _, err := os.Stat(s.themesDir); os.IsNotExist(err) {
		return themes, nil
	}

	// Read theme files
	entries, err := os.ReadDir(s.themesDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".css") {
			continue
		}

		themeID := strings.TrimSuffix(entry.Name(), ".css")
		themeName := strings.ReplaceAll(strings.ReplaceAll(themeID, "-", " "), "_", " ")
		// Use cases.Title instead of deprecated strings.Title
		caser := cases.Title(language.English)
		themeName = caser.String(themeName)

		icon := "🎨"
		if i, ok := themeIcons[themeID]; ok {
			icon = i
		}

		// Parse theme metadata
		themeType := s.parseThemeType(filepath.Join(s.themesDir, entry.Name()))

		themes = append(themes, models.Theme{
			ID:      themeID,
			Name:    icon + " " + themeName,
			Type:    themeType,
			Builtin: false,
		})
	}

	// Sort by name
	sort.Slice(themes, func(i, j int) bool {
		return themes[i].Name < themes[j].Name
	})

	return themes, nil
}

// parseThemeType parses theme type from CSS file comments
func (s *ThemeService) parseThemeType(themePath string) string {
	// Default to dark for backward compatibility
	themeType := "dark"

	file, err := os.Open(themePath)
	if err != nil {
		return themeType
	}
	defer file.Close()

	// Read first few lines using standard bufio.Scanner
	scanner := bufio.NewScanner(file)
	for i := 0; i < 10 && scanner.Scan(); i++ {
		line := scanner.Text()
		if strings.Contains(line, "@theme-type:") {
			if strings.Contains(line, "light") {
				return "light"
			}
			return "dark"
		}
	}

	return themeType
}

// GetThemeCSS returns the CSS content for a theme
func (s *ThemeService) GetThemeCSS(themeID string) (string, error) {
	// 🔒 安全检查：防止路径遍历攻击
	if !ValidatePathSecurity(s.themesDir, themeID+".css") {
		return "", fmt.Errorf("invalid theme path")
	}

	themePath := filepath.Join(s.themesDir, themeID+".css")

	// Check if exists
	if _, err := os.Stat(themePath); os.IsNotExist(err) {
		return "", nil
	}

	content, err := os.ReadFile(themePath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
