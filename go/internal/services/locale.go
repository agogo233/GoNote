package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gonote/internal/models"
)

// LocaleService handles locale operations
type LocaleService struct {
	localesDir string
}

// NewLocaleService creates a new LocaleService
func NewLocaleService(localesDir string) *LocaleService {
	return &LocaleService{localesDir: localesDir}
}

// Locale flags mapping
var localeFlags = map[string]string{
	"en":    "🇺🇸",
	"zh":    "🇨🇳",
	"zh-CN": "🇨🇳",
	"zh-TW": "🇹🇼",
	"ja":    "🇯🇵",
	"ko":    "🇰🇷",
	"de":    "🇩🇪",
	"fr":    "🇫🇷",
	"es":    "🇪🇸",
	"pt":    "🇵🇹",
	"ru":    "🇷🇺",
	"it":    "🇮🇹",
}

// Locale names mapping
var localeNames = map[string]string{
	"en":    "English",
	"zh":    "中文",
	"zh-CN": "简体中文",
	"zh-TW": "繁體中文",
	"ja":    "日本語",
	"ko":    "한국어",
	"de":    "Deutsch",
	"fr":    "Français",
	"es":    "Español",
	"pt":    "Português",
	"ru":    "Русский",
	"it":    "Italiano",
}

// GetLocales returns all available locales
func (s *LocaleService) GetLocales() ([]models.Locale, error) {
	locales := []models.Locale{}

	// Check if locales directory exists
	if _, err := os.Stat(s.localesDir); os.IsNotExist(err) {
		return locales, nil
	}

	// Read locale files
	entries, err := os.ReadDir(s.localesDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		code := strings.TrimSuffix(entry.Name(), ".json")

		flag := "🌐"
		if f, ok := localeFlags[code]; ok {
			flag = f
		}

		name := code
		if n, ok := localeNames[code]; ok {
			name = n
		}

		locales = append(locales, models.Locale{
			Code: code,
			Name: name,
			Flag: flag,
		})
	}

	// Sort by code
	sort.Slice(locales, func(i, j int) bool {
		return locales[i].Code < locales[j].Code
	})

	return locales, nil
}

// GetLocaleContent returns the content of a locale file
func (s *LocaleService) GetLocaleContent(code string) (map[string]interface{}, error) {
	// 🔒 安全检查：防止路径遍历攻击
	if !ValidatePathSecurity(s.localesDir, code+".json") {
		return nil, fmt.Errorf("invalid locale path")
	}

	localePath := filepath.Join(s.localesDir, code+".json")

	// Check if exists
	if _, err := os.Stat(localePath); os.IsNotExist(err) {
		return nil, nil
	}

	content, err := os.ReadFile(localePath)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, err
	}

	return data, nil
}
