package services

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNoteService(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	assert.NotNil(t, svc)
	assert.Equal(t, tmpDir, svc.notesDir)
	assert.NotNil(t, svc.cache)
	assert.NotNil(t, svc.tagCache)
}

func TestNewNoteServiceWithCache(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteServiceWithCache(tmpDir, 5*time.Minute, 1000)

	assert.NotNil(t, svc)
	assert.Equal(t, tmpDir, svc.notesDir)
}

func TestScanNotesEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	notes, folders, err := svc.ScanNotes(false)

	assert.NoError(t, err)
	assert.Empty(t, notes)
	assert.Empty(t, folders)
}

func TestScanNotesWithMarkdown(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	// Create test notes
	note1 := filepath.Join(tmpDir, "note1.md")
	note2 := filepath.Join(tmpDir, "subdir", "note2.md")

	require.NoError(t, os.WriteFile(note1, []byte("# Note 1"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Dir(note2), 0755))
	require.NoError(t, os.WriteFile(note2, []byte("# Note 2"), 0644))

	notes, folders, err := svc.ScanNotes(false)

	assert.NoError(t, err)
	assert.Len(t, notes, 2)
	assert.Contains(t, folders, "subdir")

	// Verify note properties
	var foundNote1, foundNote2 bool
	for _, note := range notes {
		if note.Name == "note1" {
			foundNote1 = true
			assert.Equal(t, "", note.Folder)
			assert.Equal(t, "note", note.Type)
		}
		if note.Name == "note2" {
			foundNote2 = true
			assert.Equal(t, "subdir", note.Folder)
		}
	}
	assert.True(t, foundNote1)
	assert.True(t, foundNote2)
}

func TestScanNotesWithMedia(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	// Create markdown and media files
	note := filepath.Join(tmpDir, "note.md")
	image := filepath.Join(tmpDir, "image.png")
	video := filepath.Join(tmpDir, "video.mp4")

	require.NoError(t, os.WriteFile(note, []byte("# Note"), 0644))
	require.NoError(t, os.WriteFile(image, []byte("fake image"), 0644))
	require.NoError(t, os.WriteFile(video, []byte("fake video"), 0644))

	// Scan without media
	notesNoMedia, _, err := svc.ScanNotes(false)
	assert.NoError(t, err)
	assert.Len(t, notesNoMedia, 1)

	// Scan with media
	notesWithMedia, _, err := svc.ScanNotes(true)
	assert.NoError(t, err)
	assert.Len(t, notesWithMedia, 3)

	// Verify media types
	mediaTypes := make(map[string]bool)
	for _, n := range notesWithMedia {
		mediaTypes[n.Type] = true
	}
	assert.True(t, mediaTypes["note"])
	assert.True(t, mediaTypes["image"])
	assert.True(t, mediaTypes["video"])
}

func TestScanNotesSkipsDotFiles(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	// Create normal and dot files
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "visible.md"), []byte("# Visible"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".hidden.md"), []byte("# Hidden"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".hiddenDir"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".hiddenDir", "note.md"), []byte("# Hidden Note"), 0644))

	notes, _, err := svc.ScanNotes(false)

	assert.NoError(t, err)
	assert.Len(t, notes, 1)
	assert.Equal(t, "visible", notes[0].Name)
}

func TestGetNoteContent(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	content := "# Test Note\n\nThis is test content."
	notePath := "test.md"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, notePath), []byte(content), 0644))

	result, err := svc.GetNoteContent(notePath)

	assert.NoError(t, err)
	assert.Equal(t, content, result)
}

func TestGetNoteContentInvalidPath(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	_, err := svc.GetNoteContent("../../../etc/passwd")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid path")
}

func TestGetNoteContentNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	_, err := svc.GetNoteContent("nonexistent.md")

	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestSaveNote(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	content := "# New Note\n\nContent here."
	err := svc.SaveNote("new-note.md", content)

	assert.NoError(t, err)

	// Verify file was created
	data, err := os.ReadFile(filepath.Join(tmpDir, "new-note.md"))
	require.NoError(t, err)
	assert.Equal(t, content, string(data))
}

func TestSaveNoteAddsMdExtension(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	err := svc.SaveNote("note", "content")

	assert.NoError(t, err)

	// Verify .md was added
	_, err = os.Stat(filepath.Join(tmpDir, "note.md"))
	assert.NoError(t, err)
}

func TestSaveNoteInSubdirectory(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	err := svc.SaveNote("subdir/note.md", "content")

	assert.NoError(t, err)

	// Verify directory and file were created
	_, err = os.Stat(filepath.Join(tmpDir, "subdir", "note.md"))
	assert.NoError(t, err)
}

func TestSaveNoteInvalidPath(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	err := svc.SaveNote("../../../tmp/malicious.md", "content")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid path")
}

func TestDeleteNote(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	notePath := "to-delete.md"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, notePath), []byte("content"), 0644))

	err := svc.DeleteNote(notePath)

	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(tmpDir, notePath))
	assert.True(t, os.IsNotExist(err))
}

func TestDeleteNoteNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	err := svc.DeleteNote("nonexistent.md")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeleteNoteInvalidPath(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	err := svc.DeleteNote("../../../etc/passwd")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid path")
}

func TestMoveNote(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	oldPath := "old.md"
	newPath := "new.md"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, oldPath), []byte("content"), 0644))

	err := svc.MoveNote(oldPath, newPath)

	assert.NoError(t, err)

	// Verify old file is gone
	_, err = os.Stat(filepath.Join(tmpDir, oldPath))
	assert.True(t, os.IsNotExist(err))

	// Verify new file exists
	data, err := os.ReadFile(filepath.Join(tmpDir, newPath))
	require.NoError(t, err)
	assert.Equal(t, "content", string(data))
}

func TestMoveNoteToSubdirectory(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	oldPath := "old.md"
	newPath := "subdir/new.md"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, oldPath), []byte("content"), 0644))

	err := svc.MoveNote(oldPath, newPath)

	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(tmpDir, newPath))
	assert.NoError(t, err)
}

func TestMoveNoteSourceNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	err := svc.MoveNote("nonexistent.md", "new.md")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestMoveNoteDestinationExists(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "old.md"), []byte("old"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "new.md"), []byte("new"), 0644))

	err := svc.MoveNote("old.md", "new.md")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestNoteExists(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "exists.md"), []byte("content"), 0644))

	assert.True(t, svc.NoteExists("exists.md"))
	assert.False(t, svc.NoteExists("nonexistent.md"))
}

func TestGetNoteMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	content := "line1\nline2\nline3"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "meta.md"), []byte(content), 0644))

	meta, err := svc.GetNoteMetadata("meta.md")

	assert.NoError(t, err)
	assert.NotNil(t, meta)
	assert.Equal(t, 3, meta.Lines)
	assert.Greater(t, meta.Size, int64(0))
	assert.NotEmpty(t, meta.Modified)
}

func TestGetNoteMetadataNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	_, err := svc.GetNoteMetadata("nonexistent.md")

	assert.Error(t, err)
}

func TestGetNoteMetadataInvalidPath(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	_, err := svc.GetNoteMetadata("../../../etc/passwd")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid path")
}

func TestGetAttachmentDir(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	// Root level note
	dir := svc.GetAttachmentDir("")
	assert.Equal(t, filepath.Join(tmpDir, "_attachments"), dir)

	// Note in subdirectory
	dir = svc.GetAttachmentDir("subdir/note.md")
	assert.Equal(t, filepath.Join(tmpDir, "subdir", "_attachments"), dir)
}

func TestGetAttachmentsEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	// Create a note without any media references
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "test.md"), []byte("# Test\n\nNo media here"), 0644))

	attachments, err := svc.GetAttachments("test.md")

	assert.NoError(t, err)
	assert.Empty(t, attachments)
}

func TestGetAttachments(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	// Create attachments directory and files
	attachDir := filepath.Join(tmpDir, "_attachments")
	require.NoError(t, os.MkdirAll(attachDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(attachDir, "image.png"), []byte("fake image"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(attachDir, "doc.pdf"), []byte("fake pdf"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(attachDir, "unused.png"), []byte("unused image"), 0644))

	// Create a note that references only image.png and doc.pdf
	noteContent := `# Test Note

![[]]  This is an image:
![[image.png]]

And a document:
![[doc.pdf]]
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "test.md"), []byte(noteContent), 0644))

	attachments, err := svc.GetAttachments("test.md")

	assert.NoError(t, err)
	assert.Len(t, attachments, 2)

	// Verify attachment properties - should only have referenced files
	names := make(map[string]bool)
	for _, a := range attachments {
		names[a.Name] = true
		assert.NotEmpty(t, a.Path)
		assert.Greater(t, a.Size, int64(0))
	}
	assert.True(t, names["image.png"])
	assert.True(t, names["doc.pdf"])
	assert.False(t, names["unused.png"]) // Should not be included
}

func TestGetAttachmentsNoNote(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	// Create attachments directory with files but no note
	attachDir := filepath.Join(tmpDir, "_attachments")
	require.NoError(t, os.MkdirAll(attachDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(attachDir, "image.png"), []byte("fake image"), 0644))

	// Should return empty when note doesn't exist
	attachments, err := svc.GetAttachments("nonexistent.md")

	assert.NoError(t, err)
	assert.Empty(t, attachments)
}

func TestSaveUploadedImage(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	imageData := []byte("fake image data")
	relPath, err := svc.SaveUploadedImage("", "test.png", imageData)

	assert.NoError(t, err)
	assert.NotEmpty(t, relPath)
	assert.Contains(t, relPath, ".png")

	// Verify file exists
	_, err = os.Stat(filepath.Join(tmpDir, relPath))
	assert.NoError(t, err)
}

func TestSaveUploadedImageSanitizesFilename(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	imageData := []byte("data")
	relPath, err := svc.SaveUploadedImage("", "test file with spaces & special!.png", imageData)

	assert.NoError(t, err)
	assert.NotEmpty(t, relPath)
	// SanitizeFilename replaces dangerous characters like \ / : * ? " < > |
	// but spaces and & are kept (they are not in the dangerous character list)
	assert.Contains(t, relPath, ".png")
}

func TestInvalidateCache(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	// Add cache entries
	svc.cache.Set("test:key", "value")
	svc.tagCache.Set("test:tag", "tagvalue")

	assert.Equal(t, 1, svc.cache.Len())
	assert.Equal(t, 1, svc.tagCache.Len())

	svc.InvalidateCache()

	assert.Equal(t, 0, svc.cache.Len())
	assert.Equal(t, 0, svc.tagCache.Len())
}

func TestGetTagsCached(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	content := "---\ntags: [tag1, tag2]\n---\n\nContent"
	notePath := filepath.Join(tmpDir, "tagged.md")
	require.NoError(t, os.WriteFile(notePath, []byte(content), 0644))

	tags := svc.GetTagsCached(notePath)

	assert.Len(t, tags, 2)
	assert.Contains(t, tags, "tag1")
	assert.Contains(t, tags, "tag2")
}

func TestGetTagsCachedUsesCache(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	content := "---\ntags: [tag1]\n---\n\nContent"
	notePath := filepath.Join(tmpDir, "tagged.md")
	require.NoError(t, os.WriteFile(notePath, []byte(content), 0644))

	// First call - cache miss
	tags := svc.GetTagsCached(notePath)
	assert.Len(t, tags, 1)

	// Modify file (different mod time should trigger cache miss)
	time.Sleep(10 * time.Millisecond)
	require.NoError(t, os.WriteFile(notePath, []byte("---\ntags: [tag1, tag2]\n---\n\nContent"), 0644))

	// Second call - should re-parse due to changed mod time
	tags = svc.GetTagsCached(notePath)
	assert.Len(t, tags, 2)
}

func TestIsReady(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("without background scanner", func(t *testing.T) {
		svc := NewNoteService(tmpDir)
		assert.True(t, svc.IsReady())
	})

	t.Run("with background scanner", func(t *testing.T) {
		svc := NewNoteServiceWithScanner(tmpDir, time.Minute, 100, time.Hour)
		assert.False(t, svc.IsReady())

		svc.StartBackgroundScanner()
		defer svc.StopBackgroundScanner()

		svc.WaitForReady()
		assert.True(t, svc.IsReady())
	})
}

func TestBackgroundScanner(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteServiceWithScanner(tmpDir, time.Minute, 100, 100*time.Millisecond)
	svc.StartBackgroundScanner()
	defer svc.StopBackgroundScanner()

	// Wait for first scan
	svc.WaitForReady()

	// Create a note
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "test.md"), []byte("content"), 0644))

	// Trigger immediate scan
	svc.TriggerScan()
	time.Sleep(300 * time.Millisecond)

	// Cache should be updated
	notes, _, err := svc.ScanNotes(false)
	assert.NoError(t, err)
	assert.Len(t, notes, 1)
}

func TestSaveNoteWithCheck_OptimisticLockConflict(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	notePath := "test.md"
	content1 := "# Version 1"

	// First save: create note (no knownMtime → skip check)
	err := svc.SaveNoteWithCheck(notePath, content1, "")
	require.NoError(t, err)

	// Guarantee mtime will differ on next write
	time.Sleep(10 * time.Millisecond)

	// Get current mtime from the saved file
	fullPath := filepath.Join(tmpDir, notePath)
	info, err := os.Stat(fullPath)
	require.NoError(t, err)
	goodMtime := info.ModTime().UTC().Format(time.RFC3339Nano)

	// Save with correct mtime → should succeed
	content2 := "# Version 2"
	err = svc.SaveNoteWithCheck(notePath, content2, goodMtime)
	require.NoError(t, err)

	// Stale mtime (from before version 2 write) → should return ErrConflict
	content3 := "# Version 3"
	err = svc.SaveNoteWithCheck(notePath, content3, goodMtime)
	assert.ErrorIs(t, err, ErrConflict)

	// Empty knownMtime → backward compatible skip check → should succeed
	content4 := "# Version 4"
	err = svc.SaveNoteWithCheck(notePath, content4, "")
	require.NoError(t, err)

	// Verify final content is version 4
	result, err := svc.GetNoteContent(notePath)
	require.NoError(t, err)
	assert.Equal(t, content4, result)
}

// TestSaveNoteWithCheck_GetThenSaveNoConflict 锁定回归：GetNoteContentWithMetadata
// 返回的 Modified 串（RFC3339Nano）原样回传给 SaveNoteWithCheck 应通过乐观锁，
// 不得因格式不对称而误报 ErrConflict。
func TestSaveNoteWithCheck_GetThenSaveNoConflict(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	notePath := "regression.md"
	content1 := "# First"

	// 创建笔记
	require.NoError(t, svc.SaveNoteWithCheck(notePath, content1, ""))
	// 确保后续写产生不同亚秒 mtime
	time.Sleep(10 * time.Millisecond)

	// 走真实读取路径，拿到前端会拿到的 Modified 串
	_, meta, err := svc.GetNoteContentWithMetadata(notePath)
	require.NoError(t, err)
	require.NotEmpty(t, meta.Modified)

	// 原样回传保存 → 不应误报冲突
	require.NoError(t, svc.SaveNoteWithCheck(notePath, "# Second", meta.Modified))

	//第二次用同一 stale 串保存 → 现在文件 mtime 已变，应正确报冲突
	assert.ErrorIs(t, svc.SaveNoteWithCheck(notePath, "# Third", meta.Modified), ErrConflict)
}

// TestSaveNoteWithCheck_SubSecondStaleMtime 确认亚秒级 mtime 偏移仍能正确触发冲突检测，
// 防止改用 time.Time.Equal 比较后丢失精度。
func TestSaveNoteWithCheck_SubSecondStaleMtime(t *testing.T) {
	tmpDir := t.TempDir()
	svc := NewNoteService(tmpDir)

	notePath := "stale.md"
	require.NoError(t, svc.SaveNoteWithCheck(notePath, "# v1", ""))

	fullPath := filepath.Join(tmpDir, notePath)
	info, err := os.Stat(fullPath)
	require.NoError(t, err)

	// 当前 mtime 减 1ms → 亚秒级偏移 → 应判冲突
	staleMtime := info.ModTime().UTC().Add(-time.Millisecond).Format(time.RFC3339Nano)
	assert.ErrorIs(t, svc.SaveNoteWithCheck(notePath, "# v2", staleMtime), ErrConflict)

	// 当前 mtime 完全一致 → 应通过
	currentMtime := info.ModTime().UTC().Format(time.RFC3339Nano)
	require.NoError(t, svc.SaveNoteWithCheck(notePath, "# v3", currentMtime))
}
