package content

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Generator handles Hugo markdown generation.
type Generator struct {
	contentDir string
}

// NewGenerator creates a new content generator.
func NewGenerator(contentDir string) *Generator {
	return &Generator{
		contentDir: contentDir,
	}
}

// GeneratePost creates a Hugo markdown post from a transcript.
func (g *Generator) GeneratePost(transcriptPath string, outputPath string) error {
	// Read transcript
	transcript, err := os.ReadFile(transcriptPath)
	if err != nil {
		return fmt.Errorf("failed to read transcript file %s: %w", transcriptPath, err)
	}

	// Generate timestamp for title
	now := time.Now()
	title := fmt.Sprintf("Voice Memo %s", now.Format("2006-01-02 15:04"))

	// Create frontmatter
	frontmatter := fmt.Sprintf(`---
title: "%s"
date: %s
draft: true
---

`, title, now.Format(time.RFC3339))

	// Combine frontmatter and transcript
	content := frontmatter + string(transcript)

	// Ensure output directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", dir, err)
	}

	// Write markdown file
	//nolint:gosec // Markdown files need to be readable
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write markdown file %s: %w", outputPath, err)
	}

	// Archive source files on success
	return g.archiveFiles(transcriptPath)
}

// archiveFiles moves transcript and corresponding audio to archive.
func (g *Generator) archiveFiles(transcriptPath string) error {
	// Determine archive directory
	// Look for .memos parent directory in the transcript path
	absPath, err := filepath.Abs(transcriptPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for %s: %w", transcriptPath, err)
	}

	var memosRoot string
	dir := filepath.Dir(absPath)

	// Walk up the directory tree to find .memos
	for {
		if filepath.Base(dir) == ".memos" {
			memosRoot = dir
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding .memos, use ~/.memos
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get user home directory: %w", err)
			}
			memosRoot = filepath.Join(homeDir, ".memos")
			break
		}
		dir = parent
	}

	archiveDir := filepath.Join(memosRoot, "archive")

	// Create archive directory
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return fmt.Errorf("failed to create archive directory %s: %w", archiveDir, err)
	}

	// Get base name (without extension)
	base := transcriptPath[:len(transcriptPath)-len(filepath.Ext(transcriptPath))]
	wavPath := base + ".wav"

	// Archive transcript
	transcriptDest := filepath.Join(archiveDir, filepath.Base(transcriptPath))
	if err := os.Rename(transcriptPath, transcriptDest); err != nil {
		return fmt.Errorf("failed to archive transcript %s to %s: %w", transcriptPath, transcriptDest, err)
	}

	// Archive audio if it exists
	if _, err := os.Stat(wavPath); err == nil {
		wavDest := filepath.Join(archiveDir, filepath.Base(wavPath))
		if err := os.Rename(wavPath, wavDest); err != nil {
			// Attempt to restore transcript on failure
			_ = os.Rename(transcriptDest, transcriptPath)
			return fmt.Errorf("failed to archive audio %s to %s: %w", wavPath, wavDest, err)
		}
	}

	return nil
}
