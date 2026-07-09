package services

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// 文件名清理用预编译正则
var (
	reFilenameDangerous = regexp.MustCompile(`[\\/:*?"<>|\x00-\x1f]`)
	reFilenameUnderscore = regexp.MustCompile(`_+`)
)

// ValidatePathSecurity ensures a path is within the notes directory
func ValidatePathSecurity(notesDir, targetPath string) bool {
	absNotesDir, err := filepath.Abs(notesDir)
	if err != nil {
		return false
	}

	absTarget, err := filepath.Abs(filepath.Join(notesDir, targetPath))
	if err != nil {
		return false
	}

	// Ensure the target is within notesDir by checking prefix with separator
	// This prevents /app/data-backup from matching /app/data
	if absTarget == absNotesDir {
		return true
	}
	return strings.HasPrefix(absTarget, absNotesDir+string(filepath.Separator))
}

// ValidatePathSecurityAbs validates an already-absolute path
func ValidatePathSecurityAbs(notesDir, absTarget string) bool {
	absNotesDir, err := filepath.Abs(notesDir)
	if err != nil {
		return false
	}
	return IsPathInside(absTarget, absNotesDir)
}

// IsPathInside reports whether absTarget is equal to or nested under absParent.
// Both must be cleaned absolute paths. Prevents /app/data-backup from matching
// /app/data by requiring a path separator boundary.
func IsPathInside(absTarget, absParent string) bool {
	if absTarget == absParent {
		return true
	}
	return strings.HasPrefix(absTarget, absParent+string(filepath.Separator))
}

// AtomicWrite writes data to path atomically: it first writes to a sibling
// temporary file in the same directory (guaranteeing same filesystem so
// rename is atomic) and then renames it over the target. On failure the
// target file is left untouched. The temporary file is removed on error.
func AtomicWrite(path string, data []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	cleaned := false
	defer func() {
		if !cleaned {
			tmp.Close()
			os.Remove(tmpName)
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	cleaned = true
	return nil
}

// SanitizeFilename removes dangerous characters from filenames
func SanitizeFilename(filename string) string {
	if filename == "" {
		return "unnamed"
	}

	// Get the extension first
	parts := strings.Split(filename, ".")
	name := parts[0]
	ext := ""
	if len(parts) > 1 {
		ext = "." + strings.Join(parts[1:], ".")
	}

	// Remove dangerous characters: \ / : * ? " < > | and control chars
	name = reFilenameDangerous.ReplaceAllString(name, "_")

	// Collapse multiple underscores
	name = reFilenameUnderscore.ReplaceAllString(name, "_")

	// Strip leading/trailing underscores and spaces
	name = strings.Trim(name, "_ ")

	if name == "" {
		name = "unnamed"
	}

	return name + ext
}

// EnsureDirectories creates required directories if they don't exist
func EnsureDirectories(dirs ...string) error {
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

// CreateFolder creates a new folder in the notes directory
func CreateFolder(notesDir, folderPath string) error {
	fullPath := filepath.Join(notesDir, folderPath)

	// Security check
	if !ValidatePathSecurity(notesDir, folderPath) {
		return fmt.Errorf("invalid path")
	}

	return os.MkdirAll(fullPath, 0755)
}

// DeleteFolder deletes a folder and all its contents
func DeleteFolder(notesDir, folderPath string) error {
	fullPath := filepath.Join(notesDir, folderPath)

	// Security check
	if !ValidatePathSecurity(notesDir, folderPath) {
		return fmt.Errorf("invalid path")
	}

	// Check if folder exists
	info, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("folder does not exist")
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory")
	}

	return os.RemoveAll(fullPath)
}

// MoveFolder moves a folder to a new location
// 顺序：先 rename 成功后再更新 folder 内所有 wikilinks 反向引用，
// backlinks 改写失败时回滚 rename，避免出现"链接已改但目录没搬"的不可恢复状态。
func MoveFolder(notesDir, oldPath, newPath string) error {
	oldFull := filepath.Join(notesDir, oldPath)
	newFull := filepath.Join(notesDir, newPath)

	// Security checks
	if !ValidatePathSecurity(notesDir, oldPath) || !ValidatePathSecurity(notesDir, newPath) {
		return fmt.Errorf("invalid path")
	}

	// Check source exists
	if _, err := os.Stat(oldFull); os.IsNotExist(err) {
		return fmt.Errorf("source folder does not exist")
	}

	// Check target doesn't exist
	if _, err := os.Stat(newFull); err == nil {
		return fmt.Errorf("destination already exists")
	}

	// Create parent directory if needed
	if err := os.MkdirAll(filepath.Dir(newFull), 0755); err != nil {
		return err
	}

	// 1) 先 rename：失败直接返回，无任何副作用
	if err := os.Rename(oldFull, newFull); err != nil {
		return err
	}

	// 2) rename 成功后更新所有指向该文件夹内笔记的 wikilinks；失败则回滚 rename
	backlinkService := NewBacklinkService(notesDir)
	if _, err := backlinkService.UpdateFolderBacklinks(oldPath, newPath); err != nil {
		if rbErr := os.Rename(newFull, oldFull); rbErr != nil {
			return fmt.Errorf("folder moved but backlink update failed (%v), and rollback also failed (%v)", err, rbErr)
		}
		return fmt.Errorf("folder moved but backlink update failed, rolled back rename: %w", err)
	}

	return nil
}

// RenameFolder renames a folder (alias for MoveFolder)
func RenameFolder(notesDir, oldPath, newPath string) error {
	return MoveFolder(notesDir, oldPath, newPath)
}

// ToPosixPath converts a path to forward slashes for consistency
func ToPosixPath(path string) string {
	return filepath.ToSlash(path)
}

// ListNotesUnderPath lists all .md file paths (relative to notesDir) under the given prefix
func ListNotesUnderPath(notesDir, prefix string) ([]string, error) {
	absDir := filepath.Join(notesDir, prefix)
	var notes []string

	err := filepath.WalkDir(absDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		rel, err := filepath.Rel(notesDir, path)
		if err != nil {
			return nil
		}
		notes = append(notes, ToPosixPath(rel))
		return nil
	})
	if os.IsNotExist(err) {
		return nil, nil
	}
	return notes, err
}
