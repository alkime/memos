package ai

import "testing"

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		expected string
	}{
		{
			name:     "simple title",
			title:    "Voice CLI Improvements",
			expected: "voice-cli-improvements",
		},
		{
			name:     "title with special characters",
			title:    "AI-Powered Content: First Draft!",
			expected: "ai-powered-content-first-draft",
		},
		{
			name:     "title with multiple spaces",
			title:    "Multiple    Spaces   Here",
			expected: "multiple-spaces-here",
		},
		{
			name:     "title with leading/trailing spaces",
			title:    "  Trimmed Title  ",
			expected: "trimmed-title",
		},
		{
			name:     "title already lowercase",
			title:    "already-lowercase",
			expected: "already-lowercase",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateSlug(tt.title)
			if got != tt.expected {
				t.Errorf("GenerateSlug(%q) = %q, want %q", tt.title, got, tt.expected)
			}
		})
	}
}
