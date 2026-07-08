package services

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"gonote/internal/models"
)

var (
	linkIdxWikilinkRegex = regexp.MustCompile(`\[\[([^\]|]+)(?:\|[^\]]+)?\]\]`)
	linkIdxMdLinkRegex   = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+\.md)\)`)
)

type linkIdxLinkInfo struct {
	target string
	linkType string
}

type LinkIndex struct {
	mu         sync.RWMutex
	backlinks  map[string][]string
	graphEdges []models.GraphEdge
	graphNodes map[string]bool
	notesDir   string
}

func NewLinkIndex(notesDir string) *LinkIndex {
	return &LinkIndex{
		backlinks:  make(map[string][]string),
		graphEdges: make([]models.GraphEdge, 0),
		graphNodes: make(map[string]bool),
		notesDir:   notesDir,
	}
}

func (li *LinkIndex) RebuildFull() {
	li.mu.Lock()
	defer li.mu.Unlock()

	li.backlinks = make(map[string][]string)
	li.graphEdges = make([]models.GraphEdge, 0)
	li.graphNodes = make(map[string]bool)

	filepath.WalkDir(li.notesDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(li.notesDir, path)
		relPath = ToPosixPath(relPath)

		links := li.extractLinksFromContent(string(content))
		for _, link := range links {
			resolved := li.resolveLinkNoScan(link.target)
			if resolved == "" {
				continue
			}
			li.backlinks[resolved] = append(li.backlinks[resolved], relPath)
			li.graphEdges = append(li.graphEdges, models.GraphEdge{
				Source: relPath,
				Target: resolved,
				Type:   link.linkType,
			})
		}

		li.graphNodes[relPath] = true
		return nil
	})
}

func (li *LinkIndex) GetBacklinkSources(notePath string) []string {
	li.mu.RLock()
	defer li.mu.RUnlock()

	result := li.backlinks[notePath]
	if result == nil {
		return []string{}
	}
	out := make([]string, len(result))
	copy(out, result)
	return out
}

func (li *LinkIndex) GetGraph() *models.GraphData {
	li.mu.RLock()
	defer li.mu.RUnlock()

	nodes := make([]models.GraphNode, 0, len(li.graphNodes))
	for path := range li.graphNodes {
		name := strings.TrimSuffix(filepath.Base(path), ".md")
		nodes = append(nodes, models.GraphNode{ID: path, Label: name})
	}

	edges := make([]models.GraphEdge, len(li.graphEdges))
	copy(edges, li.graphEdges)

	return &models.GraphData{Nodes: nodes, Edges: edges}
}

func (li *LinkIndex) GetNodeCount() int {
	li.mu.RLock()
	defer li.mu.RUnlock()
	return len(li.graphNodes)
}

func (li *LinkIndex) extractLinksFromContent(content string) []linkIdxLinkInfo {
	links := []linkIdxLinkInfo{}
	seen := make(map[string]bool)

	wikilinkMatches := linkIdxWikilinkRegex.FindAllStringSubmatch(content, -1)
	for _, match := range wikilinkMatches {
		target := strings.TrimSpace(match[1])
		if idx := strings.Index(target, "#"); idx != -1 {
			target = target[:idx]
		}
		if target != "" && !seen[target] {
			links = append(links, linkIdxLinkInfo{target: target, linkType: "wikilink"})
			seen[target] = true
		}
	}

	mdLinkMatches := linkIdxMdLinkRegex.FindAllStringSubmatch(content, -1)
	for _, match := range mdLinkMatches {
		target := strings.TrimSpace(match[2])
		if idx := strings.Index(target, "#"); idx != -1 {
			target = target[:idx]
		}
		target = strings.TrimSuffix(target, ".md")
		if target != "" && !seen[target] {
			links = append(links, linkIdxLinkInfo{target: target, linkType: "mdlink"})
			seen[target] = true
		}
	}

	return links
}

func (li *LinkIndex) resolveLinkNoScan(target string) string {
	if strings.Contains(target, "/") {
		variants := []string{target, target + ".md"}
		for _, v := range variants {
			fullPath := filepath.Join(li.notesDir, v)
			if _, err := os.Stat(fullPath); err == nil {
				return ToPosixPath(v)
			}
		}
		return ""
	}

	baseDir := filepath.Join(li.notesDir, target)
	if _, err := os.Stat(baseDir + ".md"); err == nil {
		return target + ".md"
	}

	entries, err := os.ReadDir(li.notesDir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if e.IsDir() {
			subPath := filepath.Join(li.notesDir, e.Name(), target+".md")
			if _, err := os.Stat(subPath); err == nil {
				return ToPosixPath(filepath.Join(e.Name(), target+".md"))
			}
		}
	}

	return ""
}