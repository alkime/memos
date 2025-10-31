package content

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Generator handles Hugo markdown generation
type Generator struct {
	contentDir string
}

// NewGenerator creates a new content generator
func NewGenerator(contentDir string) *Generator {
	return &Generator{
		contentDir: contentDir,
	}
}

// GeneratePost creates a Hugo markdown post from a transcript
func (g *Generator) GeneratePost(transcriptPath string, outputPath string) error {
	// Read transcript
	transcript, err := os.ReadFile(transcriptPath)
	if err != nil {
		return err
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
		return err
	}

	// Write markdown file
	return os.WriteFile(outputPath, []byte(content), 0644)
}
