package tools

import (
	"fmt"
	"path/filepath"
	"strings"
)

// resolveInWorkDir resolves a user-supplied path safely inside workDir.
// It rejects any path that would escape workDir (e.g. via ../ traversal).
func resolveInWorkDir(workDir, userPath string) (string, error) {
	// Clean user path first to normalize things like a/../b
	cleanUser := filepath.Clean(userPath)

	// Join with workDir; if cleanUser is absolute, Join will ignore workDir,
	// so we must validate after resolving abs paths.
	candidate := filepath.Join(workDir, cleanUser)

	workAbs, err := filepath.Abs(workDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve workDir abs path: %w", err)
	}

	candAbs, err := filepath.Abs(candidate)
	if err != nil {
		return "", fmt.Errorf("failed to resolve candidate abs path: %w", err)
	}

	rel, err := filepath.Rel(workAbs, candAbs)
	if err != nil {
		return "", fmt.Errorf("failed to compute relative path: %w", err)
	}

	// If rel starts with "..", it's outside. Also handle the ".." exactly case.
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes working directory")
	}

	return candAbs, nil
}
