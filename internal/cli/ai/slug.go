package ai

import (
	"regexp"
	"strings"
)

// GenerateSlug converts a title to a URL-friendly slug.
// Example: "Voice CLI Improvements" -> "voice-cli-improvements"
func GenerateSlug(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove special characters (keep alphanumeric and hyphens)
	reg := regexp.MustCompile(`[^a-z0-9-]+`)
	slug = reg.ReplaceAllString(slug, "")

	// Collapse multiple hyphens to single hyphen
	reg = regexp.MustCompile(`-+`)
	slug = reg.ReplaceAllString(slug, "-")

	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")

	return slug
}
