package services

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gonote/internal/models"
)

// TemplateService handles template operations
type TemplateService struct {
	notesDir string
}

// NewTemplateService creates a new TemplateService
func NewTemplateService(notesDir string) *TemplateService {
	return &TemplateService{notesDir: notesDir}
}

// GetTemplates returns all templates
func (s *TemplateService) GetTemplates() ([]models.Template, error) {
	templates := []models.Template{}
	templatesPath := filepath.Join(s.notesDir, "_templates")

	// Check if templates directory exists
	if _, err := os.Stat(templatesPath); os.IsNotExist(err) {
		return templates, nil
	}

	// Security check
	if !ValidatePathSecurity(s.notesDir, "_templates") {
		return templates, nil
	}

	// Read template files
	entries, err := os.ReadDir(templatesPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		fullPath := filepath.Join(templatesPath, entry.Name())

		// Security check
		if !ValidatePathSecurityAbs(s.notesDir, fullPath) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		templates = append(templates, models.Template{
			Name:     strings.TrimSuffix(entry.Name(), ".md"),
			Path:     "_templates/" + entry.Name(),
			Modified: info.ModTime().UTC().Format(time.RFC3339Nano),
		})
	}

	// Sort by name
	sort.Slice(templates, func(i, j int) bool {
		return templates[i].Name < templates[j].Name
	})

	return templates, nil
}

// GetTemplateContent returns the content of a template
func (s *TemplateService) GetTemplateContent(templateName string) (string, error) {
	templatePath := filepath.Join(s.notesDir, "_templates", templateName+".md")

	// Check if exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return "", fmt.Errorf("template not found")
	}

	// Security check
	if !ValidatePathSecurity(s.notesDir, "_templates/"+templateName+".md") {
		return "", fmt.Errorf("invalid path")
	}

	content, err := os.ReadFile(templatePath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// ApplyTemplatePlaceholders replaces template placeholders with actual values
func ApplyTemplatePlaceholders(content, notePath string) string {
	now := time.Now()

	// Get note name and folder
	noteName := strings.TrimSuffix(filepath.Base(notePath), ".md")
	folder := filepath.Dir(notePath)
	if folder == "." {
		folder = "Root"
	} else {
		folder = filepath.Base(folder)
	}

	// Define replacements
	replacements := map[string]string{
		"{{date}}":      now.Format("2006-01-02"),
		"{{time}}":      now.Format("15:04:05"),
		"{{datetime}}":  now.Format("2006-01-02 15:04:05"),
		"{{timestamp}}": fmt.Sprintf("%d", now.Unix()),
		"{{year}}":      now.Format("2006"),
		"{{month}}":     now.Format("01"),
		"{{day}}":       now.Format("02"),
		"{{title}}":     noteName,
		"{{folder}}":    folder,
	}

	result := content
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// CreateNoteFromTemplate creates a new note from a template
func (s *TemplateService) CreateNoteFromTemplate(templateName, notePath string) (string, error) {
	// Get template content
	templateContent, err := s.GetTemplateContent(templateName)
	if err != nil {
		return "", err
	}

	// Apply placeholders
	content := ApplyTemplatePlaceholders(templateContent, notePath)

	// Ensure .md extension
	if !strings.HasSuffix(notePath, ".md") {
		notePath += ".md"
	}

	// Save the note
	ns := NewNoteService(s.notesDir)
	if err := ns.SaveNote(notePath, content); err != nil {
		return "", err
	}

	return notePath, nil
}
