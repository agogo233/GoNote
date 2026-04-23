package services

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidatePathSecurity_ValidPath(t *testing.T) {
	tests := []struct {
		name       string
		notesDir   string
		targetPath string
		expected   bool
	}{
		{
			name:       "simple path",
			notesDir:   "/app/data",
			targetPath: "notes/test.md",
			expected:   true,
		},
		{
			name:       "nested path",
			notesDir:   "/app/data",
			targetPath: "folder/sub/note.md",
			expected:   true,
		},
		{
			name:       "chinese characters",
			notesDir:   "/app/data",
			targetPath: "简单的笔记.md",
			expected:   true,
		},
		{
			name:       "chinese folder",
			notesDir:   "/app/data",
			targetPath: "测试文件夹/笔记.md",
			expected:   true,
		},
		{
			name:       "empty path",
			notesDir:   "/app/data",
			targetPath: "",
			expected:   true,
		},
		{
			name:       "dot path",
			notesDir:   "/app/data",
			targetPath: ".",
			expected:   true,
		},
		{
			name:       "single file",
			notesDir:   "/app/data",
			targetPath: "readme.md",
			expected:   true,
		},
		{
			name:       "underscore prefix",
			notesDir:   "/app/data",
			targetPath: "_attachments/image.png",
			expected:   true,
		},
		{
			name:       "special characters in name",
			notesDir:   "/app/data",
			targetPath: "notes/test-note_v2.md",
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidatePathSecurity(tt.notesDir, tt.targetPath)
			if result != tt.expected {
				t.Errorf("ValidatePathSecurity(%q, %q) = %v, expected %v",
					tt.notesDir, tt.targetPath, result, tt.expected)
			}
		})
	}
}

func TestValidatePathSecurity_PathTraversal(t *testing.T) {
	tests := []struct {
		name       string
		notesDir   string
		targetPath string
		expected   bool
	}{
		{
			name:       "basic traversal",
			notesDir:   "/app/data",
			targetPath: "../etc/passwd",
			expected:   false,
		},
		{
			name:       "nested traversal",
			notesDir:   "/app/data",
			targetPath: "notes/../../etc/passwd",
			expected:   false,
		},
		{
			name:       "deep traversal",
			notesDir:   "/app/data",
			targetPath: "../../../etc/passwd",
			expected:   false,
		},
		{
			name:       "absolute path becomes relative",
			notesDir:   "/app/data",
			targetPath: "/etc/passwd",
			// Note: ValidatePathSecurity joins notesDir + targetPath,
			// so /app/data + /etc/passwd = /app/data/etc/passwd which is valid
			expected: true,
		},
		{
			name:       "mixed traversal",
			notesDir:   "/app/data",
			targetPath: "notes/../../../app/config.yaml",
			expected:   false,
		},
		{
			name:       "traversal in middle",
			notesDir:   "/app/data",
			targetPath: "folder/../..//etc/passwd",
			expected:   false,
		},
		{
			name:       "dot dot slash",
			notesDir:   "/app/data",
			targetPath: "..",
			expected:   false,
		},
		{
			name:       "traversal to parent of data",
			notesDir:   "/app/data",
			targetPath: "../config.yaml",
			expected:   false,
		},
		{
			name:       "complex traversal",
			notesDir:   "/app/data",
			targetPath: "a/b/c/../../../..//etc/passwd",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidatePathSecurity(tt.notesDir, tt.targetPath)
			if result != tt.expected {
				t.Errorf("ValidatePathSecurity(%q, %q) = %v, expected %v",
					tt.notesDir, tt.targetPath, result, tt.expected)
			}
		})
	}
}

func TestValidatePathSecurity_EdgeCases(t *testing.T) {
	// Test with real temporary directory
	tmpDir := t.TempDir()

	t.Run("valid path in temp dir", func(t *testing.T) {
		result := ValidatePathSecurity(tmpDir, "test.md")
		if !result {
			t.Error("Expected valid path to be accepted")
		}
	})

	t.Run("traversal from temp dir", func(t *testing.T) {
		result := ValidatePathSecurity(tmpDir, "../outside.md")
		if result {
			t.Error("Expected traversal attack to be blocked")
		}
	})

	t.Run("subdirectory creation", func(t *testing.T) {
		subDir := filepath.Join(tmpDir, "subdir")
		if err := os.Mkdir(subDir, 0755); err != nil {
			t.Fatalf("Failed to create subdir: %v", err)
		}
		result := ValidatePathSecurity(tmpDir, "subdir/note.md")
		if !result {
			t.Error("Expected subdirectory path to be valid")
		}
	})
}

func TestValidatePathSecurityAbs(t *testing.T) {
	tests := []struct {
		name      string
		notesDir  string
		absTarget string
		expected  bool
	}{
		{
			name:      "valid absolute path",
			notesDir:  "/app/data",
			absTarget: "/app/data/notes/test.md",
			expected:  true,
		},
		{
			name:      "path outside notes dir",
			notesDir:  "/app/data",
			absTarget: "/etc/passwd",
			expected:  false,
		},
		{
			name:      "exact match to notes dir",
			notesDir:  "/app/data",
			absTarget: "/app/data",
			expected:  true,
		},
		{
			name:      "similar but different dir",
			notesDir:  "/app/data",
			absTarget: "/app/data-backup/test.md",
			expected:  false,
		},
		{
			name:      "parent directory",
			notesDir:  "/app/data",
			absTarget: "/app",
			expected:  false,
		},
		{
			name:      "data-prefix without separator",
			notesDir:  "/app/data",
			absTarget: "/app/datafile.md",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidatePathSecurityAbs(tt.notesDir, tt.absTarget)
			if result != tt.expected {
				t.Errorf("ValidatePathSecurityAbs(%q, %q) = %v, expected %v",
					tt.notesDir, tt.absTarget, result, tt.expected)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal filename",
			input:    "normal.md",
			expected: "normal.md",
		},
		{
			name:     "filename with less than",
			input:    "file<name>.md",
			// After replacement: file_name_.md -> collapse underscores -> file_name.md
			expected: "file_name.md",
		},
		{
			name:     "filename with greater than",
			input:    "file>name.md",
			expected: "file_name.md",
		},
		{
			name:     "filename with colon",
			input:    "file:name.md",
			expected: "file_name.md",
		},
		{
			name:     "filename with question mark",
			input:    "file?.md",
			// After replacement: file_.md -> trim underscores -> file.md
			expected: "file.md",
		},
		{
			name:     "filename with asterisk",
			input:    "file*.md",
			// After replacement: file_.md -> trim underscores -> file.md
			expected: "file.md",
		},
		{
			name:     "filename with pipe",
			input:    "file|name.md",
			expected: "file_name.md",
		},
		{
			name:     "filename with double quote",
			input:    "file\"name.md",
			expected: "file_name.md",
		},
		{
			name:     "filename with backslash",
			input:    "file\\name.md",
			expected: "file_name.md",
		},
		{
			name:     "filename with forward slash",
			input:    "file/name.md",
			expected: "file_name.md",
		},
		{
			name:     "empty filename",
			input:    "",
			expected: "unnamed",
		},
		{
			name:     "only spaces",
			input:    "   .md",
			expected: "unnamed.md",
		},
		{
			name:     "only underscores",
			input:    "___.md",
			expected: "unnamed.md",
		},
		{
			name:     "multiple dangerous chars",
			input:    "a<b>c:d?e*f|g\"h.md",
			expected: "a_b_c_d_e_f_g_h.md",
		},
		{
			name:     "multiple underscores collapse",
			input:    "file___name.md",
			expected: "file_name.md",
		},
		{
			name:     "leading trailing underscores",
			input:    "_filename_.md",
			expected: "filename.md",
		},
		{
			name:     "double extension",
			input:    "file.tar.gz",
			expected: "file.tar.gz",
		},
		{
			name:     "no extension",
			input:    "filename",
			expected: "filename",
		},
		{
			name:     "chinese filename",
			input:    "测试笔记.md",
			expected: "测试笔记.md",
		},
		{
			name:     "unicode chars preserved",
			input:    "笔记-2024.md",
			expected: "笔记-2024.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeFilename(%q) = %q, expected %q",
					tt.input, result, tt.expected)
			}
		})
	}
}

func TestEnsureDirectories(t *testing.T) {
	t.Run("create single directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		targetDir := filepath.Join(tmpDir, "newdir")

		err := EnsureDirectories(targetDir)
		if err != nil {
			t.Fatalf("EnsureDirectories failed: %v", err)
		}

		info, err := os.Stat(targetDir)
		if err != nil {
			t.Fatalf("Directory not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("Expected a directory")
		}
	})

	t.Run("create nested directories", func(t *testing.T) {
		tmpDir := t.TempDir()
		targetDir := filepath.Join(tmpDir, "a", "b", "c")

		err := EnsureDirectories(targetDir)
		if err != nil {
			t.Fatalf("EnsureDirectories failed: %v", err)
		}

		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			t.Error("Nested directories not created")
		}
	})

	t.Run("create multiple directories", func(t *testing.T) {
		tmpDir := t.TempDir()
		dir1 := filepath.Join(tmpDir, "dir1")
		dir2 := filepath.Join(tmpDir, "dir2")
		dir3 := filepath.Join(tmpDir, "dir3")

		err := EnsureDirectories(dir1, dir2, dir3)
		if err != nil {
			t.Fatalf("EnsureDirectories failed: %v", err)
		}

		for _, dir := range []string{dir1, dir2, dir3} {
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				t.Errorf("Directory %s not created", dir)
			}
		}
	})

	t.Run("existing directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Should not error on existing directory
		err := EnsureDirectories(tmpDir)
		if err != nil {
			t.Fatalf("EnsureDirectories failed on existing dir: %v", err)
		}
	})
}

func TestCreateFolder(t *testing.T) {
	t.Run("create valid folder", func(t *testing.T) {
		tmpDir := t.TempDir()

		err := CreateFolder(tmpDir, "newfolder")
		if err != nil {
			t.Fatalf("CreateFolder failed: %v", err)
		}

		folderPath := filepath.Join(tmpDir, "newfolder")
		info, err := os.Stat(folderPath)
		if err != nil {
			t.Fatalf("Folder not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("Expected a directory")
		}
	})

	t.Run("create nested folder", func(t *testing.T) {
		tmpDir := t.TempDir()

		err := CreateFolder(tmpDir, "parent/child/grandchild")
		if err != nil {
			t.Fatalf("CreateFolder failed: %v", err)
		}

		folderPath := filepath.Join(tmpDir, "parent/child/grandchild")
		if _, err := os.Stat(folderPath); os.IsNotExist(err) {
			t.Error("Nested folder not created")
		}
	})

	t.Run("block traversal attack", func(t *testing.T) {
		tmpDir := t.TempDir()

		err := CreateFolder(tmpDir, "../outside")
		if err == nil {
			t.Error("Expected error for path traversal")
		}
	})

	t.Run("create existing folder", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create once
		err := CreateFolder(tmpDir, "existing")
		if err != nil {
			t.Fatalf("First CreateFolder failed: %v", err)
		}

		// Create again (should not error)
		err = CreateFolder(tmpDir, "existing")
		if err != nil {
			t.Fatalf("Second CreateFolder should not fail: %v", err)
		}
	})
}

func TestDeleteFolder(t *testing.T) {
	t.Run("delete existing folder", func(t *testing.T) {
		tmpDir := t.TempDir()
		folderPath := filepath.Join(tmpDir, "to-delete")

		// Create folder with content
		os.MkdirAll(folderPath, 0755)
		os.WriteFile(filepath.Join(folderPath, "file.md"), []byte("content"), 0644)

		err := DeleteFolder(tmpDir, "to-delete")
		if err != nil {
			t.Fatalf("DeleteFolder failed: %v", err)
		}

		if _, err := os.Stat(folderPath); !os.IsNotExist(err) {
			t.Error("Folder should be deleted")
		}
	})

	t.Run("delete non-existent folder", func(t *testing.T) {
		tmpDir := t.TempDir()

		err := DeleteFolder(tmpDir, "non-existent")
		if err == nil {
			t.Error("Expected error for non-existent folder")
		}
	})

	t.Run("block traversal on delete", func(t *testing.T) {
		tmpDir := t.TempDir()

		err := DeleteFolder(tmpDir, "../outside")
		if err == nil {
			t.Error("Expected error for path traversal")
		}
	})

	t.Run("delete file not folder", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "file.md")
		os.WriteFile(filePath, []byte("content"), 0644)

		err := DeleteFolder(tmpDir, "file.md")
		if err == nil {
			t.Error("Expected error when deleting file as folder")
		}
	})
}

func TestMoveFolder(t *testing.T) {
	t.Run("move folder successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		oldPath := filepath.Join(tmpDir, "old-folder")
		newPath := filepath.Join(tmpDir, "new-folder")

		// Create source folder with content
		os.MkdirAll(oldPath, 0755)
		os.WriteFile(filepath.Join(oldPath, "note.md"), []byte("content"), 0644)

		err := MoveFolder(tmpDir, "old-folder", "new-folder")
		if err != nil {
			t.Fatalf("MoveFolder failed: %v", err)
		}

		// Check old is gone
		if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
			t.Error("Old folder should be removed")
		}

		// Check new exists
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			t.Error("New folder should exist")
		}
	})

	t.Run("move non-existent folder", func(t *testing.T) {
		tmpDir := t.TempDir()

		err := MoveFolder(tmpDir, "non-existent", "new-location")
		if err == nil {
			t.Error("Expected error for non-existent source")
		}
	})

	t.Run("move to existing destination", func(t *testing.T) {
		tmpDir := t.TempDir()
		oldPath := filepath.Join(tmpDir, "source")
		newPath := filepath.Join(tmpDir, "dest")

		os.MkdirAll(oldPath, 0755)
		os.MkdirAll(newPath, 0755)

		err := MoveFolder(tmpDir, "source", "dest")
		if err == nil {
			t.Error("Expected error when destination exists")
		}
	})

	t.Run("block traversal in move", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.MkdirAll(filepath.Join(tmpDir, "folder"), 0755)

		err := MoveFolder(tmpDir, "folder", "../outside")
		if err == nil {
			t.Error("Expected error for path traversal in destination")
		}

		err = MoveFolder(tmpDir, "../outside", "folder")
		if err == nil {
			t.Error("Expected error for path traversal in source")
		}
	})
}

func TestToPosixPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "already posix",
			input:    "folder/file.md",
			expected: "folder/file.md",
		},
		{
			name:     "simple path",
			input:    "folder/file.md",
			expected: "folder/file.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToPosixPath(tt.input)
			if result != tt.expected {
				t.Errorf("ToPosixPath(%q) = %q, expected %q",
					tt.input, result, tt.expected)
			}
		})
	}
}
