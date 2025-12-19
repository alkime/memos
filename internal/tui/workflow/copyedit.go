package workflow

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alkime/memos/internal/content"
	"github.com/alkime/memos/internal/tui/components/labeledspinner"
	"github.com/alkime/memos/internal/tui/style"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type copyEditPhase struct {
	spinner   labeledspinner.Model
	inputPath string
	mode      content.Mode
	client    *content.Writer
	outputDir string

	// Completion state
	complete   bool
	outputPath string
	changes    []string
	title      string
}

// NewCopyEditPhase creates a new copy edit phase.
func NewCopyEditPhase(inputPath, apiKey string, mode content.Mode, outputDir string) tea.Model {
	return &copyEditPhase{
		spinner: labeledspinner.New(
			spinner.Pulse,
			"Copy editing draft...",
			"Claude is polishing your post",
			"This may take a moment",
		),
		inputPath: inputPath,
		mode:      mode,
		client:    content.NewWriter(apiKey),
		outputDir: outputDir,
	}
}

type copyEditCompleteMsg struct {
	outputPath string
	changes    []string
	title      string
}

func (cp *copyEditPhase) Init() tea.Cmd {
	return tea.Sequence(
		cp.spinner.Init(),
		cp.copyEditCmd(),
	)
}

func (cp *copyEditPhase) Update(teaMsg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := teaMsg.(type) {
	case copyEditCompleteMsg:
		cp.complete = true
		cp.outputPath = msg.outputPath
		cp.changes = msg.changes
		cp.title = msg.title

		return cp, nil
	}

	// Only update spinner if not complete
	if !cp.complete {
		var cmd tea.Cmd
		cp.spinner, cmd = cp.spinner.Update(teaMsg)

		return cp, cmd
	}

	return cp, nil
}

func (cp *copyEditPhase) View() string {
	if !cp.complete {
		return cp.spinner.View()
	}

	return cp.completeView()
}

func (cp *copyEditPhase) completeView() string {
	var sb strings.Builder

	// Header
	sb.WriteString(style.Title.Render("=== Copy Edit Complete ==="))
	sb.WriteString("\n\n")

	// Title and output path
	sb.WriteString(style.Label.Render("Title: "))
	sb.WriteString(cp.title)
	sb.WriteString("\n")
	sb.WriteString(style.Label.Render("Saved: "))
	sb.WriteString(style.Muted.Render(cp.outputPath))
	sb.WriteString("\n\n")

	// Changes list
	sb.WriteString(style.Label.Render("Changes:"))
	sb.WriteString("\n")
	for _, change := range cp.changes {
		sb.WriteString("  ")
		sb.WriteString(style.Bullet.Render("\u2022 "))
		sb.WriteString(change)
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	// Help text (only global quit keys)
	sb.WriteString(renderGlobalKeyHelp())

	return sb.String()
}

func (cp *copyEditPhase) copyEditCmd() tea.Cmd {
	return func() tea.Msg {
		// Read user-edited first draft
		draftContent, err := os.ReadFile(cp.inputPath)
		if err != nil {
			slog.Error("Failed to read first draft file", "error", err)
			return tea.Quit
		}

		// Generate copy edit via Claude API
		currentDate := time.Now().Format("2006-01-02")
		result, err := cp.client.GenerateCopyEdit(string(draftContent), currentDate, cp.mode)
		if err != nil {
			slog.Error("Copy edit generation failed", "error", err)
			return tea.Quit
		}

		// Generate output path: {outputDir}/{YYYY-MM}-{slug}.md
		slug := content.GenerateSlug(result.Title)
		datePrefix := time.Now().Format("2006-01")
		filename := fmt.Sprintf("%s-%s.md", datePrefix, slug)
		outputPath := filepath.Join(cp.outputDir, filename)

		// Write the final post
		//nolint:gosec // Blog posts need to be readable
		if err := os.WriteFile(outputPath, []byte(result.Markdown), 0o644); err != nil {
			slog.Error("Failed to write copy-edited post", "error", err, "path", outputPath)
			return tea.Quit
		}

		slog.Info("Copy edit complete", "output", outputPath, "title", result.Title)

		return copyEditCompleteMsg{
			outputPath: outputPath,
			changes:    result.Changes,
			title:      result.Title,
		}
	}
}
