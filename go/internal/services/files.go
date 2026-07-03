package services

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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
	// Ensure the target is within notesDir by checking prefix with separator
	// This prevents /app/data-backup from matching /app/data
	if absTarget == absNotesDir {
		return true
	}
	return strings.HasPrefix(absTarget, absNotesDir+string(filepath.Separator))
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
	re := regexp.MustCompile(`[\\/:*?"<>|\x00-\x1f]`)
	name = re.ReplaceAllString(name, "_")

	// Collapse multiple underscores
	reMultiple := regexp.MustCompile(`_+`)
	name = reMultiple.ReplaceAllString(name, "_")

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
