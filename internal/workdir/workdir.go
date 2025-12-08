// Package workdir provides utilities for managing the voice CLI working directory.
package workdir

import (
	"fmt"
	"os"
	"path/filepath"
)

// Root returns the base directory for all voice CLI working files.
// The path is expanded at runtime to resolve to:
//
//	$HOME/Documents/Alkime/Memos
func Root() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(home, "Documents", "Alkime", "Memos"), nil
}

// WorkPath returns the full path for a working directory with the given name.
// The name typically corresponds to a git branch or user-specified identifier.
func WorkPath(workingName string) (string, error) {
	root, err := Root()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "work", workingName), nil
}

// FilePath returns the full path for a file in a working directory.
func FilePath(workingName, filename string) (string, error) {
	workPath, err := WorkPath(workingName)
	if err != nil {
		return "", err
	}
	return filepath.Join(workPath, filename), nil
}

// Prep ensures that the working directory for the given name exists.
func Prep(workingName string) error {
	workPath, err := WorkPath(workingName)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(workPath, 0755); err != nil {
		return fmt.Errorf("failed to create working directory %s: %w", workPath, err)
	}

	return nil
}
