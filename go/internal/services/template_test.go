package services

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTemplateService(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewTemplateService(tmpDir)

	assert.NotNil(t, service)
	assert.Equal(t, tmpDir, service.notesDir)
}

func TestTemplateService_GetTemplates(t *testing.T) {
	// Create temp directory with templates
	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "_templates")
	err := os.MkdirAll(templateDir, 0755)
	assert.NoError(t, err)

	templateContent := `---
name: Test Template
---
# {{title}}
`
	err = os.WriteFile(filepath.Join(templateDir, "test.md"), []byte(templateContent), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(templateDir, "another.md"), []byte(templateContent), 0644)
	assert.NoError(t, err)

	service := NewTemplateService(tmpDir)
	templates, err := service.GetTemplates()

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(templates), 2)
}

func TestTemplateService_GetTemplates_NoTemplatesDir(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewTemplateService(tmpDir)

	templates, err := service.GetTemplates()
	assert.NoError(t, err)
	assert.Empty(t, templates)
}

func TestTemplateService_GetTemplateContent(t *testing.T) {
	// Create temp directory with template
	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "_templates")
	err := os.MkdirAll(templateDir, 0755)
	assert.NoError(t, err)

	templateContent := `---
name: Test Template
---
# {{title}}
Content here
`
	err = os.WriteFile(filepath.Join(templateDir, "test.md"), []byte(templateContent), 0644)
	assert.NoError(t, err)

	service := NewTemplateService(tmpDir)
	content, err := service.GetTemplateContent("test")

	assert.NoError(t, err)
	assert.Contains(t, content, "Content here")
}

func TestTemplateService_GetTemplateContent_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewTemplateService(tmpDir)

	content, err := service.GetTemplateContent("nonexistent")
	assert.Error(t, err)
	assert.Empty(t, content)
}

func TestApplyTemplatePlaceholders(t *testing.T) {
	now := time.Now()
	notePath := "folder/test-note.md"

	content := `# {{title}}
Created: {{date}}
Time: {{time}}
Folder: {{folder}}
`

	result := ApplyTemplatePlaceholders(content, notePath)

	assert.Contains(t, result, "# test-note")
	assert.Contains(t, result, now.Format("2006-01-02"))
	assert.Contains(t, result, now.Format("15:04"))
	assert.Contains(t, result, "Folder: folder")
}

func TestApplyTemplatePlaceholders_RootFolder(t *testing.T) {
	notePath := "note.md"
	content := `Folder: {{folder}}`

	result := ApplyTemplatePlaceholders(content, notePath)

	assert.Contains(t, result, "Folder: Root")
}

func TestApplyTemplatePlaceholders_AllPlaceholders(t *testing.T) {
	notePath := "test.md"
	content := `Date: {{date}}
Time: {{time}}
DateTime: {{datetime}}
Timestamp: {{timestamp}}
Year: {{year}}
Month: {{month}}
Day: {{day}}
Title: {{title}}
Folder: {{folder}}
`

	result := ApplyTemplatePlaceholders(content, notePath)
	now := time.Now()

	assert.Contains(t, result, "Date: "+now.Format("2006-01-02"))
	assert.Contains(t, result, "Time: "+now.Format("15:04"))
	assert.Contains(t, result, "Year: "+now.Format("2006"))
	assert.Contains(t, result, "Month: "+now.Format("01"))
	assert.Contains(t, result, "Day: "+now.Format("02"))
	assert.Contains(t, result, "Title: test")
	assert.Contains(t, result, "Folder: Root")
}

func TestTemplateService_CreateNoteFromTemplate(t *testing.T) {
	// Create temp directory with templates and notes
	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "_templates")
	err := os.MkdirAll(templateDir, 0755)
	assert.NoError(t, err)

	// Template without frontmatter for simpler testing
	templateContent := `# {{title}}
Created on {{date}}
`
	err = os.WriteFile(filepath.Join(templateDir, "test.md"), []byte(templateContent), 0644)
	assert.NoError(t, err)

	service := NewTemplateService(tmpDir)
	notePath, err := service.CreateNoteFromTemplate("test", "new-note.md")

	assert.NoError(t, err)
	assert.Equal(t, "new-note.md", notePath)

	// Verify note was created
	content, err := os.ReadFile(filepath.Join(tmpDir, "new-note.md"))
	assert.NoError(t, err)
	assert.True(t, strings.Contains(string(content), "# new-note"))
}

func TestTemplateService_CreateNoteFromTemplate_TemplateNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewTemplateService(tmpDir)

	notePath, err := service.CreateNoteFromTemplate("nonexistent", "note.md")
	assert.Error(t, err)
	assert.Empty(t, notePath)
}
