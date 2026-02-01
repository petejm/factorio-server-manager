package api

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrPathTraversal = errors.New("path traversal attempt detected")
	ErrInvalidPath   = errors.New("invalid path")
)

// ValidatePathComponent validates a filename/directory name has no traversal sequences.
// It rejects empty strings, paths containing "..", and paths with path separators.
func ValidatePathComponent(name string) (string, error) {
	if name == "" {
		return "", ErrInvalidPath
	}
	if strings.Contains(name, "..") || strings.ContainsAny(name, "/\\") {
		return "", ErrPathTraversal
	}
	cleaned := filepath.Clean(name)
	if cleaned != name || cleaned == "." || cleaned == ".." {
		return "", ErrPathTraversal
	}
	return cleaned, nil
}

// SafeJoinPath safely joins a base directory with user input, preventing path traversal.
// It validates the user input and ensures the resulting path stays within the base directory.
func SafeJoinPath(baseDir, userInput string) (string, error) {
	cleanedInput, err := ValidatePathComponent(userInput)
	if err != nil {
		return "", err
	}
	joined := filepath.Join(baseDir, cleanedInput)
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return "", err
	}
	absJoined, err := filepath.Abs(joined)
	if err != nil {
		return "", err
	}
	// Ensure the joined path is within the base directory
	if !strings.HasPrefix(absJoined, absBase+string(os.PathSeparator)) && absJoined != absBase {
		return "", ErrPathTraversal
	}
	return absJoined, nil
}

// ValidateFileExtension checks if the filename has the expected extension (case-insensitive).
// The expectedExt should include the dot, e.g., ".zip"
func ValidateFileExtension(filename, expectedExt string) bool {
	return strings.EqualFold(filepath.Ext(filename), expectedExt)
}
