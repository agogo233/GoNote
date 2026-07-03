package services

import (
	"encoding/base64"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// 导出用预编译正则
var (
	// 导出时匹配带可选 alt 的媒体嵌入 wikilink: ![[file.png|alt text]]
	exportWikilinkMediaRegex = regexp.MustCompile(`!\[\[([^\]|]+)(?:\|([^\]]+))?\]\]`)
	// 匹配标准 Markdown 图片：![alt](path)
	exportMarkdownImgRegex = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)
)

// ExportService handles HTML export operations
type ExportService struct {
	notesDir   string
	themesDir  string
	libsDir    string // Path to frontend libs directory
}

// NewExportService creates a new ExportService
func NewExportService(notesDir, themesDir string) *ExportService {
	// Try to find the libs directory in common locations
	libsDir := findLibsDirectory(notesDir)

	return &ExportService{
		notesDir:  notesDir,
		themesDir: themesDir,
		libsDir:   libsDir,
	}
}

// findLibsDirectory searches for the frontend libs directory
func findLibsDirectory(notesDir string) string {
	// Try common locations relative to notesDir
	searchPaths := []string{
		filepath.Join(notesDir, "../shared/frontend/libs"),
		filepath.Join(notesDir, "../../shared/frontend/libs"),
		filepath.Join(notesDir, "../../../shared/frontend/libs"),
		filepath.Join(notesDir, "./frontend/libs"),
		"./shared/frontend/libs",
		"../shared/frontend/libs",
	}

	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// readLibFile reads a library file from the libs directory
func (s *ExportService) readLibFile(relPath string) string {
	if s.libsDir == "" {
		return ""
	}
	// Security: prevent path traversal
	if strings.Contains(relPath, "..") || strings.HasPrefix(relPath, "/") || strings.HasPrefix(relPath, "\\") {
		return ""
	}
	fullPath := filepath.Join(s.libsDir, relPath)
	// Additional validation: ensure resolved path is within libsDir
	absFullPath, err := filepath.Abs(fullPath)
	if err != nil {
		return ""
	}
	absLibsDir, err := filepath.Abs(s.libsDir)
	if err != nil {
		return ""
	}
	if !IsPathInside(absFullPath, absLibsDir) {
		return ""
	}
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return ""
	}
	return string(data)
}

// GenerateExportHTML generates a standalone HTML document for a note
func (s *ExportService) GenerateExportHTML(title, content, themeCSS string, isDark bool) string {
	// Escape content for JavaScript string
	escapedContent := escapeJS(content)

	highlightTheme := "github"
	mermaidTheme := "default"
	if isDark {
		highlightTheme = "github-dark"
		mermaidTheme = "dark"
	}

	// Read library files and inline them for fully self-contained export
	highlightJs := s.readLibFile("highlight.js/11.11.1/highlight.min.js")
	highlightThemeCss := s.readLibFile("highlight.js/11.11.1/styles/" + highlightTheme + ".min.css")
	mathJaxJs := s.readLibFile("mathjax/3.2.2/es5/tex-mml-chtml.js")
	mermaidJs := s.readLibFile("mermaid/11.12.2/dist/mermaid.min.js")
	dompurifyJs := s.readLibFile("dompurify/3.2.4/purify.min.js")

	// Build HTML template using string concatenation to avoid backtick issues
	// Inline all library code for fully self-contained export (works offline, on any device)
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>` + html.EscapeString(title) + `</title>

    <!-- Highlight.js for code syntax highlighting (inline) -->
    <style>` + highlightThemeCss + `</style>
    <script>` + highlightJs + `<` + `/script>

    <!-- Marked.js for markdown parsing (inline) -->
    <script>` + s.readLibFile("marked/12.0.2/marked.min.js") + `<` + `/script>

    <!-- DOMPurify for XSS sanitization (inline) -->
    <script>` + dompurifyJs + `<` + `/script>

    <!-- MathJax for LaTeX math rendering (inline) -->
    <script>
        MathJax = {
            tex: {
                inlineMath: [['$', '$']],
                displayMath: [['$$', '$$']],
                processEscapes: true,
                processEnvironments: true
            },
            options: {
                skipHtmlTags: ['script', 'noscript', 'style', 'textarea', 'pre', 'code']
            },
            startup: {
                pageReady: () => {
                    return MathJax.startup.defaultPageReady().then(() => {
                        document.querySelectorAll('pre code:not(.language-mermaid)').forEach((block) => {
                            hljs.highlightElement(block);
                        });
                    });
                }
            }
        };
    </script>
    <script>` + mathJaxJs + `<` + `/script>

    <!-- Mermaid.js for diagrams (inline) -->
    <script>` + mermaidJs + `<` + `/script>
    <script>
        mermaid.initialize({
            startOnLoad: false,
            theme: '` + mermaidTheme + `',
            securityLevel: 'strict',
            fontFamily: 'inherit'
        });

        document.addEventListener('DOMContentLoaded', async () => {
            const mermaidBlocks = document.querySelectorAll('pre code.language-mermaid');
            for (let i = 0; i < mermaidBlocks.length; i++) {
                const block = mermaidBlocks[i];
                const pre = block.parentElement;
                try {
                    const code = block.textContent;
                    const id = 'mermaid-diagram-' + i;
                    const { svg } = await mermaid.render(id, code);
                    const container = document.createElement('div');
                    container.className = 'mermaid-rendered';
                    container.style.cssText = 'background-color: transparent; padding: 20px; text-align: center; overflow-x: auto;';
                    container.innerHTML = svg;
                    pre.parentElement.replaceChild(container, pre);
                } catch (error) {
                    console.error('Mermaid rendering error:', error);
                }
            }
        });
    </script>
    
    <style>
        /* Theme CSS */
        ` + themeCSS + `
        
        /* Base styles */
        * { box-sizing: border-box; }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
            margin: 0;
            padding: 2rem;
            max-width: 900px;
            margin-left: auto;
            margin-right: auto;
            background-color: var(--bg-primary, #ffffff);
            color: var(--text-primary, #333333);
        }
        
        .markdown-preview { line-height: 1.6; }
        
        .markdown-preview h1, .markdown-preview h2, .markdown-preview h3,
        .markdown-preview h4, .markdown-preview h5, .markdown-preview h6 {
            margin-top: 1.5em;
            margin-bottom: 0.5em;
            font-weight: 600;
            line-height: 1.25;
        }
        
        .markdown-preview h1 { font-size: 2em; border-bottom: 1px solid var(--border-color, #e1e4e8); padding-bottom: 0.3em; }
        .markdown-preview h2 { font-size: 1.5em; border-bottom: 1px solid var(--border-color, #e1e4e8); padding-bottom: 0.3em; }
        .markdown-preview h3 { font-size: 1.25em; }
        .markdown-preview h4 { font-size: 1em; }
        
        .markdown-preview p { margin: 1em 0; }
        .markdown-preview a { color: var(--accent-primary, #0366d6); text-decoration: none; }
        .markdown-preview a:hover { text-decoration: underline; }
        .markdown-preview img { max-width: 100%; height: auto; border-radius: 4px; }
        
        .markdown-preview code:not(pre code) { 
            background-color: var(--bg-tertiary, #f6f8fa);
            color: var(--accent-primary, #0366d6);
            padding: 0.2rem 0.4rem;
            border-radius: 0.25rem;
            font-size: 0.875rem;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            font-weight: 500;
        }
        
        .markdown-preview pre { 
            background-color: var(--bg-tertiary, #f6f8fa);
            margin-bottom: 1.5rem;
            border-radius: 0.5rem;
            overflow-x: auto;
            border: 1px solid var(--border-primary, #e1e4e8);
        }
        
        .markdown-preview pre code {
            background: transparent;
            padding: 1rem;
            display: block;
            font-size: 0.875rem;
            line-height: 1.6;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            color: inherit;
        }
        
        .markdown-preview blockquote {
            margin: 1em 0;
            padding: 0 1em;
            border-left: 4px solid var(--accent-primary, #0366d6);
            color: var(--text-secondary, #6a737d);
        }
        
        .markdown-preview ul, .markdown-preview ol { padding-left: 2em; margin: 1em 0; }
        .markdown-preview li { margin: 0.25em 0; }
        
        .markdown-preview table { border-collapse: collapse; width: 100%; margin: 1em 0; }
        .markdown-preview th, .markdown-preview td { border: 1px solid var(--border-color, #e1e4e8); padding: 0.5em 1em; text-align: left; }
        .markdown-preview th { background-color: var(--bg-secondary, #f6f8fa); font-weight: 600; }
        .markdown-preview hr { border: none; border-top: 1px solid var(--border-color, #e1e4e8); margin: 2em 0; }
        .markdown-preview input[type="checkbox"] { margin-right: 0.5em; }
        
        @media (max-width: 768px) { body { padding: 1rem; } }
        @media print { body { padding: 0.5in; max-width: none; } }
    </style>
</head>
<body>
    <div class="markdown-preview" id="content"></div>
    
    <script>
        marked.setOptions({ gfm: true, breaks: true, headerIds: true, mangle: false });
        const markdown = "` + escapedContent + `";
        document.getElementById('content').innerHTML = DOMPurify.sanitize(marked.parse(markdown));
    </script>
</body>
</html>`

	return html
}

// ProcessMediaForExport processes media references for standalone HTML export
func (s *ExportService) ProcessMediaForExport(content string, noteFolder, notesDir string) string {
	// Handle wikilink media: ![[file.png]] or ![[file.mp3|alt text]]
	wikilinkPattern := exportWikilinkMediaRegex
	content = wikilinkPattern.ReplaceAllStringFunc(content, func(match string) string {
		submatches := wikilinkPattern.FindStringSubmatch(match)
		mediaName := strings.TrimSpace(submatches[1])
		altText := mediaName
		if len(submatches) > 2 && submatches[2] != "" {
			altText = strings.TrimSpace(submatches[2])
		}

		mediaType := GetMediaType(mediaName)
		if mediaType == "audio" || mediaType == "video" || mediaType == "document" {
			return generateMediaPlaceholder(mediaType, altText)
		}

		// For images, embed as base64
		resolvedPath := s.findMediaInAttachments(mediaName, noteFolder, notesDir)
		if resolvedPath != "" {
			base64URL := s.getMediaAsBase64(resolvedPath)
			if base64URL != "" {
				return fmt.Sprintf("![%s](%s)", altText, base64URL)
			}
		}

		return fmt.Sprintf(`<span style="color:#999;">🖼️ %s</span>`, altText)
	})

	// Handle standard markdown images: ![alt](path)
	imgPattern := exportMarkdownImgRegex
	content = imgPattern.ReplaceAllStringFunc(content, func(match string) string {
		submatches := imgPattern.FindStringSubmatch(match)
		altText := submatches[1]
		mediaPath := submatches[2]

		// Skip external URLs
		if strings.HasPrefix(mediaPath, "http://") || strings.HasPrefix(mediaPath, "https://") {
			return match
		}

		// Skip base64
		if strings.HasPrefix(mediaPath, "data:") {
			return match
		}

		mediaType := GetMediaType(mediaPath)
		if mediaType == "audio" || mediaType == "video" || mediaType == "document" {
			return generateMediaPlaceholder(mediaType, altText)
		}

		// For images, embed as base64
		resolvedPath := s.findMediaInAttachments(filepath.Base(mediaPath), noteFolder, notesDir)
		if resolvedPath != "" {
			base64URL := s.getMediaAsBase64(resolvedPath)
			if base64URL != "" {
				return fmt.Sprintf("![%s](%s)", altText, base64URL)
			}
		}

		return match
	})

	return content
}

// findMediaInAttachments searches for a media file in attachment locations
func (s *ExportService) findMediaInAttachments(mediaName, noteFolder, notesDir string) string {
	searchPaths := []string{
		filepath.Join(notesDir, noteFolder, mediaName),
		filepath.Join(notesDir, noteFolder, "_attachments", mediaName),
		filepath.Join(notesDir, "_attachments", mediaName),
	}

	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// getMediaAsBase64 returns a media file as base64 data URL
func (s *ExportService) getMediaAsBase64(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	contentType := getContentType(path)
	base64Data := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", contentType, base64Data)
}

// StripFrontmatter removes YAML frontmatter from markdown content
func StripFrontmatter(content string) string {
	if !strings.HasPrefix(strings.TrimSpace(content), "---") {
		return content
	}

	lines := strings.Split(content, "\n")
	if len(lines) < 2 || strings.TrimSpace(lines[0]) != "---" {
		return content
	}

	// Find closing ---
	endIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			endIdx = i
			break
		}
	}

	if endIdx == -1 {
		return content
	}

	return strings.Join(lines[endIdx+1:], "\n")
}

// generateMediaPlaceholder generates a placeholder for non-embeddable media
func generateMediaPlaceholder(mediaType, altText string) string {
	safeAlt := html.EscapeString(altText)

	icons := map[string]string{"audio": "🎵", "video": "🎬", "document": "📄"}
	labels := map[string]string{"audio": "Audio file", "video": "Video file", "document": "PDF document"}

	icon := "📎"
	if i, ok := icons[mediaType]; ok {
		icon = i
	}
	label := "Media file"
	if l, ok := labels[mediaType]; ok {
		label = l
	}

	return fmt.Sprintf(`<div style="margin:1.5rem 0;padding:1.5rem;background:#f8f9fa;border:1px solid #dee2e6;border-radius:0.5rem;display:flex;align-items:center;gap:1rem;">
<span style="font-size:2rem;">%s</span>
<div>
<div style="font-weight:600;color:#212529;">%s</div>
<div style="font-size:0.875rem;color:#6c757d;">%s — not available in exported view</div>
</div>
</div>`, icon, safeAlt, label)
}

// escapeJS escapes content for JavaScript string
func escapeJS(content string) string {
	content = strings.ReplaceAll(content, "\\", "\\\\")
	content = strings.ReplaceAll(content, `"`, `\"`)
	content = strings.ReplaceAll(content, "\n", "\\n")
	content = strings.ReplaceAll(content, "\r", "\\r")
	content = strings.ReplaceAll(content, "\t", "\\t")
	content = strings.ReplaceAll(content, "</", "<\\/")
	return content
}

// getContentType returns MIME type
func getContentType(filename string) string {
	return GetFileContentType(filename)
}
