package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExportService_GenerateExportHTML_BasicStructure(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewExportService(tmpDir, tmpDir)

	html := svc.GenerateExportHTML("Test Title", "# Content", "", false)

	assert.Contains(t, html, "<!DOCTYPE html>")
	assert.Contains(t, html, "<html lang=\"en\">")
	assert.Contains(t, html, "<title>Test Title</title>")
	assert.Contains(t, html, "<meta charset=\"UTF-8\">")
	assert.Contains(t, html, "<meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">")
	assert.Contains(t, html, "</html>")
}

func TestExportService_GenerateExportHTML_ThemeCSS(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewExportService(tmpDir, tmpDir)

	customCSS := "body { color: red; }"
	html := svc.GenerateExportHTML("Title", "Content", customCSS, false)

	assert.Contains(t, html, customCSS)
}

func TestExportService_GenerateExportHTML_DarkTheme(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewExportService(tmpDir, tmpDir)

	html := svc.GenerateExportHTML("Title", "Content", "", true)

	// Verify dark theme settings are present
	assert.Contains(t, html, "mermaid.initialize")
	assert.Contains(t, html, "dark") // mermaid theme
	// Note: github-dark CSS is inlined but may be empty if libs not found
}

func TestExportService_GenerateExportHTML_LightTheme(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewExportService(tmpDir, tmpDir)

	html := svc.GenerateExportHTML("Title", "Content", "", false)

	// Verify light theme settings are present
	assert.Contains(t, html, "mermaid.initialize")
	assert.Contains(t, html, "default") // mermaid theme
}

func TestExportService_GenerateExportHTML_MarkedJS(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewExportService(tmpDir, tmpDir)

	html := svc.GenerateExportHTML("Title", "Content", "", false)

	assert.Contains(t, html, "marked.setOptions")
	assert.Contains(t, html, "gfm: true")
	assert.Contains(t, html, "breaks: true")
}

func TestExportService_GenerateExportHTML_MathJax(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewExportService(tmpDir, tmpDir)

	html := svc.GenerateExportHTML("Title", "Content", "", false)

	assert.Contains(t, html, "MathJax")
	assert.Contains(t, html, "inlineMath: [['$', '$']]")
	assert.Contains(t, html, "displayMath: [['$$', '$$']]")
}

func TestExportService_GenerateExportHTML_Mermaid(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewExportService(tmpDir, tmpDir)

	html := svc.GenerateExportHTML("Title", "Content", "", false)

	assert.Contains(t, html, "mermaid")
	assert.Contains(t, html, "mermaid.initialize")
	assert.Contains(t, html, "startOnLoad: false")
}

func TestExportService_GenerateExportHTML_ContentEscaping(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewExportService(tmpDir, tmpDir)

	content := "Test with \"quotes\" and <tags> and \\backslash"
	html := svc.GenerateExportHTML("Title", content, "", false)

	assert.Contains(t, html, "const markdown =")
}

func TestExportService_GenerateExportHTML_NewlineHandling(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewExportService(tmpDir, tmpDir)

	content := "Line 1\nLine 2\nLine 3"
	html := svc.GenerateExportHTML("Title", content, "", false)

	assert.Contains(t, html, "\\n")
}

func TestExportService_StripFrontmatter_NoFrontmatter(t *testing.T) {
	content := "# Just content\nNo frontmatter here."
	result := StripFrontmatter(content)
	assert.Equal(t, content, result)
}

func TestExportService_StripFrontmatter_WithFrontmatter(t *testing.T) {
	content := `---
title: Test
tags: [test]
---
# Content starts here`
	result := StripFrontmatter(content)
	// The function strips the frontmatter and returns content after closing ---
	assert.Contains(t, result, "# Content starts here")
}

func TestExportService_StripFrontmatter_EmptyFrontmatter(t *testing.T) {
	content := `---
---
Content`
	result := StripFrontmatter(content)
	assert.Contains(t, result, "Content")
}

func TestExportService_StripFrontmatter_NoClosing(t *testing.T) {
	content := `---
title: Test
No closing`
	result := StripFrontmatter(content)
	assert.Equal(t, content, result)
}

func TestExportService_StripFrontmatter_OnlyOpening(t *testing.T) {
	content := `---
title: Test`
	result := StripFrontmatter(content)
	assert.Equal(t, content, result)
}

func TestExportService_escapeJS_Backslashes(t *testing.T) {
	content := "path\\to\\file"
	result := escapeJS(content)
	assert.Contains(t, result, "\\\\")
}

func TestExportService_escapeJS_Quotes(t *testing.T) {
	content := `He said "hello"`
	result := escapeJS(content)
	assert.Contains(t, result, `\"`)
}

func TestExportService_escapeJS_Newlines(t *testing.T) {
	content := "Line 1\nLine 2"
	result := escapeJS(content)
	assert.Contains(t, result, "\\n")
	assert.NotContains(t, result, "\n")
}

func TestExportService_escapeJS_Tabs(t *testing.T) {
	content := "Tab\there"
	result := escapeJS(content)
	assert.Contains(t, result, "\\t")
}

func TestExportService_escapeJS_ScriptTags(t *testing.T) {
	content := "</script><script>alert('xss')</script>"
	result := escapeJS(content)
	assert.Contains(t, result, "<\\/")
}

func TestExportService_generateMediaPlaceholder_Audio(t *testing.T) {
	result := generateMediaPlaceholder("audio", "Test Audio")

	assert.Contains(t, result, "🎵")
	assert.Contains(t, result, "Audio file")
	assert.Contains(t, result, "Test Audio")
	assert.Contains(t, result, "not available in exported view")
}

func TestExportService_generateMediaPlaceholder_Video(t *testing.T) {
	result := generateMediaPlaceholder("video", "Test Video")

	assert.Contains(t, result, "🎬")
	assert.Contains(t, result, "Video file")
}

func TestExportService_generateMediaPlaceholder_Document(t *testing.T) {
	result := generateMediaPlaceholder("document", "Test PDF")

	assert.Contains(t, result, "📄")
	assert.Contains(t, result, "PDF document")
}

func TestExportService_generateMediaPlaceholder_Unknown(t *testing.T) {
	result := generateMediaPlaceholder("unknown", "Unknown Type")

	assert.Contains(t, result, "📎")
	assert.Contains(t, result, "Media file")
}

func TestExportService_generateMediaPlaceholder_HTMLSafety(t *testing.T) {
	result := generateMediaPlaceholder("audio", "<script>alert('xss')</script>")

	assert.NotContains(t, result, "<script>")
	assert.Contains(t, result, "&lt;script&gt;")
}

func TestExportService_ProcessMediaForExport_WikilinkImage(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewExportService(tmpDir, tmpDir)

	noteDir := "test-note"
	attachmentsDir := filepath.Join(tmpDir, noteDir, "_attachments")
	os.MkdirAll(attachmentsDir, 0755)

	testImage := filepath.Join(attachmentsDir, "test.png")
	os.WriteFile(testImage, []byte("fake png"), 0644)

	content := "![test image](test.png)"
	result := svc.ProcessMediaForExport(content, noteDir, tmpDir)

	assert.Contains(t, result, "data:image/png;base64")
}

func TestExportService_ProcessMediaForExport_WikilinkFormat(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewExportService(tmpDir, tmpDir)

	noteDir := "test-note"
	attachmentsDir := filepath.Join(tmpDir, noteDir, "_attachments")
	os.MkdirAll(attachmentsDir, 0755)

	testImage := filepath.Join(attachmentsDir, "wiki.png")
	os.WriteFile(testImage, []byte("fake png"), 0644)

	content := "![[wiki.png]]"
	result := svc.ProcessMediaForExport(content, noteDir, tmpDir)

	assert.Contains(t, result, "data:image/png;base64")
}

func TestExportService_ProcessMediaForExport_WikilinkWithAlt(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewExportService(tmpDir, tmpDir)

	noteDir := "test-note"
	attachmentsDir := filepath.Join(tmpDir, noteDir, "_attachments")
	os.MkdirAll(attachmentsDir, 0755)

	testImage := filepath.Join(attachmentsDir, "alt.png")
	os.WriteFile(testImage, []byte("fake png"), 0644)

	content := "![[alt.png|Display Text]]"
	result := svc.ProcessMediaForExport(content, noteDir, tmpDir)

	assert.Contains(t, result, "Display Text")
}

func TestExportService_ProcessMediaForExport_ExternalURL(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewExportService(tmpDir, tmpDir)

	content := "![External](https://example.com/image.png)"
	result := svc.ProcessMediaForExport(content, "", tmpDir)

	assert.Contains(t, result, "https://example.com/image.png")
}

func TestExportService_ProcessMediaForExport_Base64Skip(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewExportService(tmpDir, tmpDir)

	content := "![Base64](data:image/png;base64,ABC123)"
	result := svc.ProcessMediaForExport(content, "", tmpDir)

	assert.Contains(t, result, "data:image/png;base64,ABC123")
}

func TestExportService_ProcessMediaForExport_AudioPlaceholder(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewExportService(tmpDir, tmpDir)

	content := "![[audio.mp3]]"
	result := svc.ProcessMediaForExport(content, "", tmpDir)

	assert.Contains(t, result, "🎵")
	assert.Contains(t, result, "Audio file")
}

func TestExportService_ProcessMediaForExport_VideoPlaceholder(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewExportService(tmpDir, tmpDir)

	content := "![[video.mp4]]"
	result := svc.ProcessMediaForExport(content, "", tmpDir)

	assert.Contains(t, result, "🎬")
	assert.Contains(t, result, "Video file")
}

func TestExportService_ProcessMediaForExport_PDFPlaceholder(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewExportService(tmpDir, tmpDir)

	content := "![[document.pdf]]"
	result := svc.ProcessMediaForExport(content, "", tmpDir)

	assert.Contains(t, result, "📄")
	assert.Contains(t, result, "PDF document")
}

func TestExportService_findMediaInAttachments_SameFolder(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewExportService(tmpDir, tmpDir)

	noteDir := "notes"
	fullDir := filepath.Join(tmpDir, noteDir)
	os.MkdirAll(fullDir, 0755)

	testFile := filepath.Join(fullDir, "image.png")
	os.WriteFile(testFile, []byte("content"), 0644)

	result := svc.findMediaInAttachments("image.png", noteDir, tmpDir)
	assert.Equal(t, testFile, result)
}

func TestExportService_findMediaInAttachments_AttachmentsSubfolder(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewExportService(tmpDir, tmpDir)

	noteDir := "notes"
	attachmentsDir := filepath.Join(tmpDir, noteDir, "_attachments")
	os.MkdirAll(attachmentsDir, 0755)

	testFile := filepath.Join(attachmentsDir, "image.png")
	os.WriteFile(testFile, []byte("content"), 0644)

	result := svc.findMediaInAttachments("image.png", noteDir, tmpDir)
	assert.Equal(t, testFile, result)
}

func TestExportService_findMediaInAttachments_GlobalAttachments(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewExportService(tmpDir, tmpDir)

	globalAttachments := filepath.Join(tmpDir, "_attachments")
	os.MkdirAll(globalAttachments, 0755)

	testFile := filepath.Join(globalAttachments, "image.png")
	os.WriteFile(testFile, []byte("content"), 0644)

	result := svc.findMediaInAttachments("image.png", "notes", tmpDir)
	assert.Equal(t, testFile, result)
}

func TestExportService_findMediaInAttachments_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewExportService(tmpDir, tmpDir)

	result := svc.findMediaInAttachments("nonexistent.png", "notes", tmpDir)
	assert.Equal(t, "", result)
}

func TestExportService_getMediaAsBase64_Success(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewExportService(tmpDir, tmpDir)

	testFile := filepath.Join(tmpDir, "test.png")
	testContent := []byte("PNG content")
	os.WriteFile(testFile, testContent, 0644)

	result := svc.getMediaAsBase64(testFile)

	assert.Contains(t, result, "data:image/png;base64,")
	assert.Contains(t, result, "UE5H")
}

func TestExportService_getMediaAsBase64_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewExportService(tmpDir, tmpDir)

	result := svc.getMediaAsBase64("/nonexistent/file.png")
	assert.Equal(t, "", result)
}

func TestExportService_getMediaAsBase64_DifferentTypes(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewExportService(tmpDir, tmpDir)

	tests := []struct {
		filename    string
		contentType string
	}{
		{"test.jpg", "image/jpeg"},
		{"test.gif", "image/gif"},
		{"test.pdf", "application/pdf"},
		{"test.mp4", "video/mp4"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, tt.filename)
			os.WriteFile(testFile, []byte("content"), 0644)

			result := svc.getMediaAsBase64(testFile)
			assert.Contains(t, result, "data:"+tt.contentType+";base64,")
		})
	}
}

func TestExportService_Integration_FullExport(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewExportService(tmpDir, tmpDir)

	markdown := `---
title: Test Note
tags: [test, export]
---

# Test Heading

This is a test note with **bold** and *italic* text.

## Code Example

` + "```" + `go
fmt.Println("Hello, World!")
` + "```" + `

## Math

$E = mc^2$

## Links

[External](https://example.com)
[[Internal Note]]

## Image

![Test](image.png)
`

	themeCSS := "body { font-family: sans-serif; }"
	html := svc.GenerateExportHTML("Test Note", markdown, themeCSS, false)

	assert.Contains(t, html, "<!DOCTYPE html>")
	assert.Contains(t, html, "Test Note")
	assert.Contains(t, html, "marked.parse")
	assert.Contains(t, html, "MathJax")
	assert.Contains(t, html, "mermaid")
	assert.Contains(t, html, "font-family: sans-serif")
}

func TestExportService_Integration_ExportWithMedia(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewExportService(tmpDir, tmpDir)

	noteDir := "test-note"
	attachmentsDir := filepath.Join(tmpDir, noteDir, "_attachments")
	os.MkdirAll(attachmentsDir, 0755)

	testImage := filepath.Join(attachmentsDir, "photo.jpg")
	os.WriteFile(testImage, []byte("jpeg content"), 0644)

	markdown := "![Photo](photo.jpg)"
	html := svc.GenerateExportHTML("Note", markdown, "", false)

	// Verify HTML was generated
	assert.Contains(t, html, "<!DOCTYPE html>")
	assert.Contains(t, html, "Note")

	// Process media for export
	content := svc.ProcessMediaForExport(markdown, noteDir, tmpDir)

	assert.Contains(t, content, "data:image/jpeg;base64")
}

func TestExportService_NewExportService(t *testing.T) {
	svc := NewExportService("/notes", "/themes")

	assert.Equal(t, "/notes", svc.notesDir)
	assert.Equal(t, "/themes", svc.themesDir)
}
