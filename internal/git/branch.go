package git

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// GetCurrentBranch returns the name of the current git branch.
// Returns an error if not in a git repository or if git command fails.
func GetCurrentBranch() (string, error) {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current git branch: %w", err)
	}

	branch := strings.TrimSpace(string(output))
	if branch == "" {
		return "", fmt.Errorf("git branch name is empty")
	}

	return branch, nil
}

// SanitizeBranchName sanitizes a branch name to make it safe for use as a directory name.
// Replaces characters that are invalid in file paths with hyphens.
func SanitizeBranchName(name string) string {
	// Replace invalid filesystem characters with hyphens
	replacer := strings.NewReplacer(
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "-",
		"?", "-",
		"\"", "-",
		"<", "-",
		">", "-",
		"|", "-",
	)

	sanitized := replacer.Replace(name)

	// Trim leading/trailing spaces and hyphens
	sanitized = strings.Trim(sanitized, " -")

	return sanitized
}
