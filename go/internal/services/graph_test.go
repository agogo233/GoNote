package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGraphService(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewGraphService(tmpDir)

	assert.NotNil(t, svc)
	assert.Equal(t, tmpDir, svc.notesDir)
}

func TestGetGraphEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewGraphService(tmpDir)

	graph, err := svc.GetGraph()

	assert.NoError(t, err)
	assert.NotNil(t, graph)
	assert.Empty(t, graph.Nodes)
	assert.Empty(t, graph.Edges)
}

func TestGetGraphSingleNode(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewGraphService(tmpDir)

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note.md"), []byte("# Single Note"), 0644))

	graph, err := svc.GetGraph()

	assert.NoError(t, err)
	assert.Len(t, graph.Nodes, 1)
	assert.Equal(t, "note.md", graph.Nodes[0].ID)
	assert.Equal(t, "note", graph.Nodes[0].Label)
	assert.Empty(t, graph.Edges)
}

func TestGetGraphMultipleNodes(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewGraphService(tmpDir)

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note1.md"), []byte("Content 1"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note2.md"), []byte("Content 2"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note3.md"), []byte("Content 3"), 0644))

	graph, err := svc.GetGraph()

	assert.NoError(t, err)
	assert.Len(t, graph.Nodes, 3)

	ids := make(map[string]bool)
	for _, node := range graph.Nodes {
		ids[node.ID] = true
	}
	assert.True(t, ids["note1.md"])
	assert.True(t, ids["note2.md"])
	assert.True(t, ids["note3.md"])
}

func TestGetGraphWikilink(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewGraphService(tmpDir)

	// note1 links to note2
	content1 := "This links to [[note2]]"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note1.md"), []byte(content1), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note2.md"), []byte("Content"), 0644))

	graph, err := svc.GetGraph()

	assert.NoError(t, err)
	assert.Len(t, graph.Nodes, 2)
	assert.Len(t, graph.Edges, 1)
	assert.Equal(t, "note1.md", graph.Edges[0].Source)
	assert.Equal(t, "note2.md", graph.Edges[0].Target)
	assert.Equal(t, "link", graph.Edges[0].Type)
}

func TestGetGraphWikilinkWithDisplayText(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewGraphService(tmpDir)

	// Wikilink with display text: [[note2|Display Text]]
	content1 := "This links to [[note2|Display Text]]"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note1.md"), []byte(content1), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note2.md"), []byte("Content"), 0644))

	graph, err := svc.GetGraph()

	assert.NoError(t, err)
	assert.Len(t, graph.Edges, 1)
	assert.Equal(t, "note2.md", graph.Edges[0].Target)
}

func TestGetGraphMarkdownLink(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewGraphService(tmpDir)

	// Markdown link: [text](note2.md)
	content1 := "This links to [another note](note2.md)"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note1.md"), []byte(content1), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note2.md"), []byte("Content"), 0644))

	graph, err := svc.GetGraph()

	assert.NoError(t, err)
	assert.Len(t, graph.Edges, 1)
	assert.Equal(t, "note2.md", graph.Edges[0].Target)
}

func TestGetGraphMultipleLinks(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewGraphService(tmpDir)

	content := "Links to [[note2]] and [[note3]] and also [note4](note4.md)"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note1.md"), []byte(content), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note2.md"), []byte("Content"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note3.md"), []byte("Content"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note4.md"), []byte("Content"), 0644))

	graph, err := svc.GetGraph()

	assert.NoError(t, err)
	assert.Len(t, graph.Edges, 3)
}

func TestGetGraphDedupLinks(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewGraphService(tmpDir)

	// Multiple links to the same note should only create one edge
	content := "Links to [[note2]] and again [[note2]]"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note1.md"), []byte(content), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note2.md"), []byte("Content"), 0644))

	graph, err := svc.GetGraph()

	assert.NoError(t, err)
	assert.Len(t, graph.Edges, 1) // Only one edge despite two links
}

func TestGetGraphSubdirectory(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewGraphService(tmpDir)

	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755))
	// Wikilink without leading slash for proper resolution
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "root.md"), []byte("[[subdir/nested]]"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "subdir", "nested.md"), []byte("Content"), 0644))

	graph, err := svc.GetGraph()

	assert.NoError(t, err)
	assert.Len(t, graph.Nodes, 2)
	assert.Len(t, graph.Edges, 1)
	assert.Equal(t, "subdir/nested.md", graph.Edges[0].Target)
}

func TestExtractLinksWikilink(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewGraphService(tmpDir)

	// Create target notes
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "target.md"), []byte(""), 0644))

	content := "Here is a [[target]] link"
	links := svc.extractLinks(content)

	assert.Len(t, links, 1)
	assert.Equal(t, "target.md", links[0])
}

func TestExtractLinksWikilinkWithAnchor(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewGraphService(tmpDir)

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "target.md"), []byte(""), 0644))

	content := "Here is a [[target#section]] link"
	links := svc.extractLinks(content)

	assert.Len(t, links, 1)
	assert.Equal(t, "target.md", links[0])
}

func TestExtractLinksMarkdownLink(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewGraphService(tmpDir)

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "target.md"), []byte(""), 0644))

	content := "Here is a [link text](target.md) link"
	links := svc.extractLinks(content)

	assert.Len(t, links, 1)
	assert.Equal(t, "target.md", links[0])
}

func TestExtractLinksNoMatches(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewGraphService(tmpDir)

	content := "No links here"
	links := svc.extractLinks(content)

	assert.Empty(t, links)
}

func TestExtractLinksNonExistentTarget(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewGraphService(tmpDir)

	content := "Here is a [[nonexistent]] link"
	links := svc.extractLinks(content)

	// Should return empty because target doesn't exist
	assert.Empty(t, links)
}

func TestResolveLinkByName(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewGraphService(tmpDir)

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "target.md"), []byte(""), 0644))

	resolved := svc.resolveLink("target")

	assert.Equal(t, "target.md", resolved)
}

func TestResolveLinkByPath(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewGraphService(tmpDir)

	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "subdir", "target.md"), []byte(""), 0644))

	resolved := svc.resolveLink("subdir/target")

	assert.Equal(t, "subdir/target.md", resolved)
}

func TestResolveLinkWithAnchor(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewGraphService(tmpDir)

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "target.md"), []byte(""), 0644))

	resolved := svc.resolveLink("target#section")

	assert.Equal(t, "target.md", resolved)
}

func TestResolveLinkNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewGraphService(tmpDir)

	resolved := svc.resolveLink("nonexistent")

	assert.Empty(t, resolved)
}

func TestResolveLinkAutoMdExtension(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewGraphService(tmpDir)

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "target.md"), []byte(""), 0644))

	// Should auto-add .md extension
	resolved := svc.resolveLink("target.md")

	assert.Equal(t, "target.md", resolved)
}

func TestGetGraphBidirectionalLinks(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewGraphService(tmpDir)

	// note1 -> note2, note2 -> note1 (bidirectional)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note1.md"), []byte("[[note2]]"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note2.md"), []byte("[[note1]]"), 0644))

	graph, err := svc.GetGraph()

	assert.NoError(t, err)
	assert.Len(t, graph.Nodes, 2)
	assert.Len(t, graph.Edges, 2)

	// Should have edges in both directions
	sources := make(map[string]bool)
	targets := make(map[string]bool)
	for _, edge := range graph.Edges {
		sources[edge.Source] = true
		targets[edge.Target] = true
	}
	assert.True(t, sources["note1.md"])
	assert.True(t, sources["note2.md"])
	assert.True(t, targets["note1.md"])
	assert.True(t, targets["note2.md"])
}

func TestGetGraphComplexNetwork(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewGraphService(tmpDir)

	// Create a network: note1 -> note2 -> note3, note1 -> note3
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note1.md"), []byte("[[note2]] [[note3]]"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note2.md"), []byte("[[note3]]"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "note3.md"), []byte("End note"), 0644))

	graph, err := svc.GetGraph()

	assert.NoError(t, err)
	assert.Len(t, graph.Nodes, 3)
	assert.Len(t, graph.Edges, 3)
}
