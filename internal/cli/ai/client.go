package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Client handles Anthropic API requests for content generation.
type Client struct {
	apiKey string
	model  anthropic.Model
}

// NewClient creates a new AI client.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		model:  anthropic.ModelClaudeSonnet4_5_20250929,
	}
}

// Mode represents the content generation mode for blog posts.
type Mode string

const (
	// ModeMemos is the default mode with full frontmatter (tags, voiceBased, etc.).
	ModeMemos Mode = "memos"
	// ModeJournal is a minimal mode for personal journal entries.
	ModeJournal Mode = "journal"
)

// CopyEditToolInput defines the tool input schema for copy-edit.
type CopyEditToolInput struct {
	Title    string   `json:"title"`
	Markdown string   `json:"markdown"`
	Changes  []string `json:"changes"`
}

// CopyEditResult wraps the output from GenerateCopyEdit.
type CopyEditResult struct {
	Title    string
	Markdown string
	Changes  []string
}

// getCopyEditTool returns the tool definition for copy-edit structured output.
func getCopyEditTool() anthropic.ToolParam {
	return anthropic.ToolParam{
		Name: "save_copy_edit",
		Description: anthropic.String(
			"Save the copy-edited blog post with title, markdown content, and list of changes",
		),
		InputSchema: anthropic.ToolInputSchemaParam{
			Type: "object",
			Properties: map[string]interface{}{
				"title": map[string]interface{}{
					"type":        "string",
					"description": "The blog post title (extracted from or to be used in frontmatter)",
				},
				"markdown": map[string]interface{}{
					"type":        "string",
					"description": "The complete markdown file including frontmatter and content",
				},
				"changes": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "string",
					},
					"description": "Bullet-point list of changes made during copy-edit",
				},
			},
			Required: []string{"title", "markdown", "changes"},
		},
	}
}

// GenerateFirstDraft creates a lightly edited first draft from raw transcript.
func (c *Client) GenerateFirstDraft(transcript string) (string, error) {
	if c.apiKey == "" {
		return "", errors.New("API key required: set ANTHROPIC_API_KEY or use --api-key")
	}

	client := anthropic.NewClient(option.WithAPIKey(c.apiKey))

	params := anthropic.MessageNewParams{
		Model:     c.model,
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

// parseCopyEditToolUse extracts CopyEditToolInput from response content blocks.
func parseCopyEditToolUse(content []anthropic.ContentBlockUnion) (*CopyEditToolInput, error) {
	for _, block := range content {
		if toolUse, ok := block.AsAny().(anthropic.ToolUseBlock); ok {
			var toolInput CopyEditToolInput
			inputBytes, err := json.Marshal(toolUse.Input)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal tool input: %w", err)
			}
			if err := json.Unmarshal(inputBytes, &toolInput); err != nil {
				return nil, fmt.Errorf("failed to parse tool input: %w", err)
			}

			return &toolInput, nil
		}
	}

	return nil, errors.New("no tool use found in Anthropic API response")
}

// GenerateCopyEdit performs final copy editing and returns the result.
func (c *Client) GenerateCopyEdit(
	firstDraft string,
	currentDate string,
	mode Mode,
) (*CopyEditResult, error) {
	if c.apiKey == "" {
		return nil, errors.New("API key required: set ANTHROPIC_API_KEY or use --api-key")
	}

	client := anthropic.NewClient(option.WithAPIKey(c.apiKey))
	toolDef := getCopyEditTool()

	// Create tool union param using the SDK constructor
	tool := anthropic.ToolUnionParamOfTool(toolDef.InputSchema, toolDef.Name)
	tool.OfTool.Description = toolDef.Description

	// Select appropriate prompt based on mode
	var systemPrompt string
	if mode == ModeJournal {
		systemPrompt = CopyEditSystemPromptJournal(currentDate)
	} else {
		systemPrompt = CopyEditSystemPromptMemos(currentDate)
	}

	params := anthropic.MessageNewParams{
		Model:     c.model,
		MaxTokens: 4096,
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(firstDraft)),
		},
		Tools:      []anthropic.ToolUnionParam{tool},
		ToolChoice: anthropic.ToolChoiceParamOfTool("save_copy_edit"),
	}

	ctx := context.Background()
	resp, err := client.Messages.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to generate copy edit via Anthropic API: %w", err)
	}

	// Parse tool use from response
	if len(resp.Content) == 0 {
		return nil, errors.New("empty response from Anthropic API")
	}

	toolInput, err := parseCopyEditToolUse(resp.Content)
	if err != nil {
		return nil, err
	}

	return &CopyEditResult{
		Title:    toolInput.Title,
		Markdown: toolInput.Markdown,
		Changes:  toolInput.Changes,
	}, nil
}
