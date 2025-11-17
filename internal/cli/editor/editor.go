package editor

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
)

// Open opens the specified file in the user's preferred editor.
// It uses the $EDITOR environment variable, defaulting to "vi" if not set.
// Returns an error if the editor command fails.
func Open(filePath string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	slog.Info("Opening file in editor", "editor", editor, "path", filePath)

	cmd := exec.Command(editor, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		slog.Error("Failed to open editor", "error", err)
		slog.Info("You can manually edit the file", "path", filePath)
		return fmt.Errorf("failed to open editor: %w", err)
	}

	return nil
}
