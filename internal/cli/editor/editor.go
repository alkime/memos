package editor

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
)

// Open opens the specified file in the user's preferred editor.
// It uses the $MEMOS_EDITOR environment variable, defaulting "open" if not set.
// Returns an error if the editor command fails.
func Open(ctx context.Context, filePath string) error {
	editor := os.Getenv("MEMOS_EDITOR")
	if editor == "" {
		// todo: only works on macOS ... update for other platforms.
		editor = "open"
	}

	slog.Info("Opening file in editor", "editor", editor, "path", filePath)

	cmd := exec.CommandContext(ctx, editor, filePath)
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
