package services

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gonote/internal/models"
)

// 图谱用预编译正则
var (
	// 匹配 wikilink 提取目标：[[note]] 或 [[note|display]]
	graphWikilinkRegex = regexp.MustCompile(`\[\[([^\]|]+)(?:\|[^\]]+)?\]\]`)
	// 匹配 Markdown 内部链接：[text](path.md)
	graphMdLinkRegex = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+\.md)\)`)
)

type linkInfo struct {
	Path string
	Type string
}

// GraphService handles knowledge graph operations
type GraphService struct {
	notesDir    string
	noteService *NoteService
	linkIndex   *LinkIndex
}

// NewGraphService creates a new GraphService.
// Pass a shared *NoteService to leverage caching; omit or pass nil to create a new one per request.
func NewGraphService(notesDir string, noteService ...*NoteService) *GraphService {
	s := &GraphService{notesDir: notesDir}
	for _, ns := range noteService {
		s.noteService = ns
	}
	return s
}

// GetGraph returns the knowledge graph data
func (s *GraphService) GetGraph() (*models.GraphData, error) {
	if s.linkIndex != nil && s.linkIndex.GetNodeCount() > 0 {
		return s.linkIndex.GetGraph(), nil
	}

	// Fallback: build graph from disk (for tests or when index is empty)
	ns := s.noteService
	if ns == nil {
		ns = NewNoteService(s.notesDir)
	}
	notes, _, err := ns.ScanNotes(false)
	if err != nil {
		return nil, err
	}

	nodes := []models.GraphNode{}
	edges := []models.GraphEdge{}

	for _, note := range notes {
		nodes = append(nodes, models.GraphNode{
			ID:    note.Path,
			Label: note.Name,
		})

		fullPath := filepath.Join(s.notesDir, note.Path)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		links := s.extractLinks(string(content))
		for _, link := range links {
			edges = append(edges, models.GraphEdge{
				Source: note.Path,
				Target: link.Path,
				Type:   link.Type,
			})
		}
	}

	return &models.GraphData{
		Nodes: nodes,
		Edges: edges,
	}, nil
}

// SetLinkIndex injects a shared LinkIndex for index-based reads.
func (s *GraphService) SetLinkIndex(li *LinkIndex) {
	s.linkIndex = li
}

// extractLinks extracts wikilinks and markdown links from content
func (s *GraphService) extractLinks(content string) []linkInfo {
	links := []linkInfo{}
	seen := make(map[string]bool)

	// Extract wikilinks: [[note]] or [[note|display text]]
	wikilinkMatches := graphWikilinkRegex.FindAllStringSubmatch(content, -1)
	for _, match := range wikilinkMatches {
		target := strings.TrimSpace(match[1])
		// Resolve target to a path
		resolved := s.resolveLink(target)
		if resolved != "" && !seen[resolved] {
			links = append(links, linkInfo{Path: resolved, Type: "wikilink"})
			seen[resolved] = true
		}
	}

	// Extract markdown links: [text](path.md)
	mdLinkMatches := graphMdLinkRegex.FindAllStringSubmatch(content, -1)
	for _, match := range mdLinkMatches {
		target := strings.TrimSpace(match[2])
		// Resolve target to a path
		resolved := s.resolveLink(target)
		if resolved != "" && !seen[resolved] {
			links = append(links, linkInfo{Path: resolved, Type: "mdlink"})
			seen[resolved] = true
		}
	}

	return links
}

// resolveLink resolves a link target to a note path
func (s *GraphService) resolveLink(target string) string {
	// Remove anchor if present
	if idx := strings.Index(target, "#"); idx != -1 {
		target = target[:idx]
	}

	// If it's already a path
	if strings.Contains(target, "/") {
		// Check if it exists
		fullPath := filepath.Join(s.notesDir, target)
		if _, err := os.Stat(fullPath); err == nil {
			return target
		}
		// Try with .md extension
		if !strings.HasSuffix(target, ".md") {
			targetWithExt := target + ".md"
			fullPath := filepath.Join(s.notesDir, targetWithExt)
			if _, err := os.Stat(fullPath); err == nil {
				return targetWithExt
			}
		}
		return ""
	}

	// It's a note name, search for it
	ns := s.noteService
	if ns == nil {
		ns = NewNoteService(s.notesDir)
	}
	notes, _, _ := ns.ScanNotes(false)

	// Exact match
	for _, note := range notes {
		if note.Name == target {
			return note.Path
		}
	}

	// Try with .md extension
	targetWithExt := target
	if !strings.HasSuffix(target, ".md") {
		targetWithExt = target + ".md"
	}
	for _, note := range notes {
		if note.Path == targetWithExt || strings.HasSuffix(note.Path, "/"+targetWithExt) {
			return note.Path
		}
	}

	return ""
}
