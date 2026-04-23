package services

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindBacklinks(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create test notes
	note1 := filepath.Join(tmpDir, "note1.md")
	note2 := filepath.Join(tmpDir, "note2.md")
	note3 := filepath.Join(tmpDir, "note3.md")

	// note1 links to note2
	os.WriteFile(note1, []byte("This links to [[note2]]"), 0644)
	// note2 has no links
	os.WriteFile(note2, []byte("This is note2"), 0644)
	// note3 links to note2 with display text
	os.WriteFile(note3, []byte("Check [[note2|Note Two]] for more"), 0644)

	service := NewBacklinkService(tmpDir)

	// Find backlinks to note2
	backlinks, err := service.FindBacklinks("note2.md")
	if err != nil {
		t.Fatalf("FindBacklinks failed: %v", err)
	}

	if len(backlinks) != 2 {
		t.Errorf("Expected 2 backlinks, got %d", len(backlinks))
	}

	// Verify backlink sources
	sources := make(map[string]bool)
	for _, bl := range backlinks {
		sources[bl.SourcePath] = true
	}
	if !sources["note1.md"] {
		t.Error("Expected note1.md to be a backlink source")
	}
	if !sources["note3.md"] {
		t.Error("Expected note3.md to be a backlink source")
	}
}

func TestUpdateWikilinks(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a note with wikilinks
	note := filepath.Join(tmpDir, "test.md")
	content := "Link to [[old-note]] and [[old-note|display text]]."
	os.WriteFile(note, []byte(content), 0644)

	service := NewBacklinkService(tmpDir)

	// Update wikilinks
	err := service.UpdateWikilinks("test.md", "old-note", "new-note")
	if err != nil {
		t.Fatalf("UpdateWikilinks failed: %v", err)
	}

	// Read updated content
	updated, _ := os.ReadFile(note)
	expected := "Link to [[new-note]] and [[new-note|display text]]."
	if string(updated) != expected {
		t.Errorf("Expected %q, got %q", expected, string(updated))
	}
}

func TestUpdateAllBacklinks(t *testing.T) {
	tmpDir := t.TempDir()

	// Create notes
	target := filepath.Join(tmpDir, "target.md")
	source1 := filepath.Join(tmpDir, "source1.md")
	source2 := filepath.Join(tmpDir, "source2.md")

	os.WriteFile(target, []byte("Target note"), 0644)
	os.WriteFile(source1, []byte("See [[target]] for details"), 0644)
	os.WriteFile(source2, []byte("Check [[target|the target]]"), 0644)

	service := NewBacklinkService(tmpDir)

	// Update all backlinks (simulating rename from target.md to new-target.md)
	count, err := service.UpdateAllBacklinks("target.md", "new-target.md")
	if err != nil {
		t.Fatalf("UpdateAllBacklinks failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 updates, got %d", count)
	}

	// Verify source1 was updated
	content1, _ := os.ReadFile(source1)
	if string(content1) != "See [[new-target]] for details" {
		t.Errorf("source1 not updated correctly: %s", content1)
	}

	// Verify source2 was updated
	content2, _ := os.ReadFile(source2)
	if string(content2) != "Check [[new-target|the target]]" {
		t.Errorf("source2 not updated correctly: %s", content2)
	}
}

func TestUpdateFolderBacklinks(t *testing.T) {
	tmpDir := t.TempDir()

	// Create folder structure
	os.MkdirAll(filepath.Join(tmpDir, "old-folder"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "other"), 0755)

	// Create notes
	note1 := filepath.Join(tmpDir, "note1.md")
	noteInFolder := filepath.Join(tmpDir, "old-folder/target.md")
	otherNote := filepath.Join(tmpDir, "other/note.md")

	os.WriteFile(note1, []byte("Link to [[old-folder/target]]"), 0644)
	os.WriteFile(noteInFolder, []byte("Target note"), 0644)
	os.WriteFile(otherNote, []byte("Also [[old-folder/target|with text]]"), 0644)

	service := NewBacklinkService(tmpDir)

	// Update folder backlinks (simulating folder rename)
	count, err := service.UpdateFolderBacklinks("old-folder", "new-folder")
	if err != nil {
		t.Fatalf("UpdateFolderBacklinks failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 updates, got %d", count)
	}

	// Verify note1 was updated
	content1, _ := os.ReadFile(note1)
	if string(content1) != "Link to [[new-folder/target]]" {
		t.Errorf("note1 not updated correctly: %s", content1)
	}

	// Verify otherNote was updated
	contentOther, _ := os.ReadFile(otherNote)
	if string(contentOther) != "Also [[new-folder/target|with text]]" {
		t.Errorf("otherNote not updated correctly: %s", contentOther)
	}
}

func TestCountBacklinks(t *testing.T) {
	tmpDir := t.TempDir()

	// Create notes
	note1 := filepath.Join(tmpDir, "note1.md")
	note2 := filepath.Join(tmpDir, "note2.md")
	target := filepath.Join(tmpDir, "target.md")

	os.WriteFile(note1, []byte("[[target]]"), 0644)
	os.WriteFile(note2, []byte("[[target]] [[target|again]]"), 0644)
	os.WriteFile(target, []byte("Target"), 0644)

	service := NewBacklinkService(tmpDir)

	count, err := service.CountBacklinks("target.md")
	if err != nil {
		t.Fatalf("CountBacklinks failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 backlinks, got %d", count)
	}
}

func TestLinkMatchesNote(t *testing.T) {
	service := &BacklinkService{}

	tests := []struct {
		linkText     string
		noteName     string
		notePath     string
		shouldMatch  bool
	}{
		{"my-note", "my-note", "my-note.md", true},           // exact name match
		{"folder/my-note", "my-note", "folder/my-note.md", true}, // path match with suffix
		{"my-note", "my-note", "folder/my-note.md", true},    // simple name matches note name
		{"different-note", "my-note", "my-note.md", false},   // different name
		{"My-Note", "my-note", "my-note.md", true},           // case insensitive
		{"folder/different", "my-note", "folder/my-note.md", false}, // different note in same folder
	}

	for _, tt := range tests {
		result := service.linkMatchesNote(tt.linkText, tt.noteName, tt.notePath)
		if result != tt.shouldMatch {
			t.Errorf("linkMatchesNote(%q, %q, %q) = %v, want %v",
				tt.linkText, tt.noteName, tt.notePath, result, tt.shouldMatch)
		}
	}
}
