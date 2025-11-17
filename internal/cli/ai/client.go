package ai

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Client handles Anthropic API requests for content generation.
type Client struct {
	apiKey string
	model  string
}

// NewClient creates a new AI client.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		model:  "claude-sonnet-4-5-20250929",
	}
}

// GenerateFirstDraft creates a lightly edited first draft from raw transcript.
func (c *Client) GenerateFirstDraft(transcript string) (string, error) {
	if c.apiKey == "" {
		return "", errors.New("API key required: set ANTHROPIC_API_KEY or use --api-key")
	}

	client := anthropic.NewClient(option.WithAPIKey(c.apiKey))

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(c.model),
		MaxTokens: 4096,
		System: []anthropic.TextBlockParam{
			{Text: FirstDraftSystemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(transcript)),
		},
	}

	ctx := context.Background()
	resp, err := client.Messages.New(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to generate first draft via Anthropic API: %w", err)
	}

	// Extract text from response
	if len(resp.Content) == 0 {
		return "", errors.New("empty response from Anthropic API")
	}

	textBlock, ok := resp.Content[0].AsAny().(anthropic.TextBlock)
	if !ok {
		return "", errors.New("unexpected response type from Anthropic API")
	}

	return textBlock.Text, nil
}

// GenerateCopyEdit performs final copy editing and returns markdown with frontmatter and extracted title.
func (c *Client) GenerateCopyEdit(firstDraft string, currentDate string) (markdown string, title string, error error) {
	if c.apiKey == "" {
		return "", "", errors.New("API key required: set ANTHROPIC_API_KEY or use --api-key")
	}

	client := anthropic.NewClient(option.WithAPIKey(c.apiKey))

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(c.model),
		MaxTokens: 4096,
		System: []anthropic.TextBlockParam{
			{Text: CopyEditSystemPrompt(currentDate)},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(firstDraft)),
		},
	}

	ctx := context.Background()
	resp, err := client.Messages.New(ctx, params)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate copy edit via Anthropic API: %w", err)
	}

	// Extract text from response
	if len(resp.Content) == 0 {
		return "", "", errors.New("empty response from Anthropic API")
	}

	textBlock, ok := resp.Content[0].AsAny().(anthropic.TextBlock)
	if !ok {
		return "", "", errors.New("unexpected response type from Anthropic API")
	}

	markdown = textBlock.Text

	// Extract title from frontmatter
	// Simple parsing: look for 'title: "..."' pattern
	title, err = extractTitleFromFrontmatter(markdown)
	if err != nil {
		return "", "", fmt.Errorf("failed to extract title from frontmatter: %w", err)
	}

	return markdown, title, nil
}

// extractTitleFromFrontmatter parses the title from Hugo frontmatter.
func extractTitleFromFrontmatter(markdown string) (string, error) {
	// Match: title: "Some Title" or title: 'Some Title' or title: Some Title
	titleRegex := regexp.MustCompile(`(?m)^title:\s*["']?([^"'\n]+)["']?`)
	matches := titleRegex.FindStringSubmatch(markdown)
	if len(matches) < 2 {
		return "", errors.New("title not found in frontmatter")
	}

	title := strings.TrimSpace(matches[1])
	if title == "" {
		return "", errors.New("title is empty in frontmatter")
	}

	return title, nil
}
