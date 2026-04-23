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

	return os.Rename(oldFull, newFull)
}

// RenameFolder renames a folder (alias for MoveFolder)
func RenameFolder(notesDir, oldPath, newPath string) error {
	return MoveFolder(notesDir, oldPath, newPath)
}

// ToPosixPath converts a path to forward slashes for consistency
func ToPosixPath(path string) string {
	return filepath.ToSlash(path)
}
