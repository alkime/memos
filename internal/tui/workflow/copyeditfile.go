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
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type copyEditFileState int

const (
	copyEditFileProcessing copyEditFileState = iota
	copyEditFileReview
	copyEditFileCompleted
)

type copyEditFilePhase struct {
	spinner  labeledspinner.Model
	filePath string
	mode     content.Mode
	client   *content.Writer
	state    copyEditFileState

	// Result from Claude
	result *content.CopyEditResult

	// Final state
	applied bool
}

// NewCopyEditFilePhase creates a new copy-edit file phase.
func NewCopyEditFilePhase(filePath, apiKey string, mode content.Mode) tea.Model {
	filename := filepath.Base(filePath)
	return &copyEditFilePhase{
		spinner: labeledspinner.New(
			spinner.Pulse,
			fmt.Sprintf("Copy editing %s...", filename),
			"Claude is polishing your post",
			"This may take a moment",
		),
		filePath: filePath,
		mode:     mode,
		client:   content.NewWriter(apiKey),
		state:    copyEditFileProcessing,
	}
}

type copyEditFileCompleteMsg struct {
	result *content.CopyEditResult
}

type copyEditFileErrorMsg struct {
	err error
}

func (cef *copyEditFilePhase) Init() tea.Cmd {
	return tea.Sequence(
		cef.spinner.Init(),
		cef.copyEditCmd(),
	)
}

func (cef *copyEditFilePhase) Update(teaMsg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := teaMsg.(type) {
	case copyEditFileCompleteMsg:
		cef.state = copyEditFileReview
		cef.result = msg.result

		return cef, nil

	case copyEditFileErrorMsg:
		slog.Error("Copy edit failed", "error", msg.err)
		return cef, tea.Quit

	case tea.KeyMsg:
		return cef.handleKeyMsg(msg)
	}

	// Update spinner if still processing
	if cef.state == copyEditFileProcessing {
		var cmd tea.Cmd
		cef.spinner, cmd = cef.spinner.Update(teaMsg)

		return cef, cmd
	}

	return cef, nil
}

func (cef *copyEditFilePhase) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	km := DefaultKeyMap()

	switch {
	case key.Matches(msg, km.Quit), key.Matches(msg, km.ForceQuit):
		// In review state, q/esc discards changes
		if cef.state == copyEditFileReview {
			cef.state = copyEditFileCompleted
			cef.applied = false

			return cef, tea.Quit
		}
		// In completed state, quit normally
		return cef, tea.Quit

	case msg.Type == tea.KeyEnter:
		// In review state, Enter applies changes
		if cef.state == copyEditFileReview {
			return cef, cef.applyChangesCmd()
		}
	}

	return cef, nil
}

func (cef *copyEditFilePhase) View() string {
	switch cef.state {
	case copyEditFileProcessing:
		return cef.spinner.View()
	case copyEditFileReview:
		return cef.reviewView()
	case copyEditFileCompleted:
		return cef.completedView()
	}

	return ""
}

func (cef *copyEditFilePhase) reviewView() string {
	var sb strings.Builder

	// Header
	sb.WriteString(style.Title.Render("=== Copy Edit Complete ==="))
	sb.WriteString("\n\n")

	// Title
	sb.WriteString(style.Label.Render("Title: "))
	sb.WriteString(cef.result.Title)
	sb.WriteString("\n\n")

	// Changes list
	sb.WriteString(style.Label.Render("Changes:"))
	sb.WriteString("\n")
	for _, change := range cef.result.Changes {
		sb.WriteString("  ")
		sb.WriteString(style.Bullet.Render("\u2022 "))
		sb.WriteString(change)
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	// Action help
	sb.WriteString(style.Help.Render("["))
	sb.WriteString(style.Key.Render("Enter"))
	sb.WriteString(style.Help.Render("] Apply changes  "))
	sb.WriteString(style.Help.Render("["))
	sb.WriteString(style.Key.Render("q"))
	sb.WriteString(style.Help.Render("/"))
	sb.WriteString(style.Key.Render("Esc"))
	sb.WriteString(style.Help.Render("] Discard\n"))

	return sb.String()
}

func (cef *copyEditFilePhase) completedView() string {
	var sb strings.Builder

	if cef.applied {
		sb.WriteString(style.Success.Render("Changes applied!"))
		sb.WriteString("\n")
		sb.WriteString(style.Label.Render("Saved: "))
		sb.WriteString(style.Muted.Render(cef.filePath))
	} else {
		sb.WriteString(style.Muted.Render("Changes discarded."))
	}

	sb.WriteString("\n")

	return sb.String()
}

func (cef *copyEditFilePhase) copyEditCmd() tea.Cmd {
	return func() tea.Msg {
		// Read file content
		fileContent, err := os.ReadFile(cef.filePath)
		if err != nil {
			return copyEditFileErrorMsg{err: fmt.Errorf("failed to read file %s: %w", cef.filePath, err)}
		}

		// Generate copy edit via Claude API
		currentDate := time.Now().Format("2006-01-02")
		result, err := cef.client.GenerateCopyEdit(string(fileContent), currentDate, cef.mode)
		if err != nil {
			return copyEditFileErrorMsg{err: fmt.Errorf("copy edit generation failed: %w", err)}
		}

		slog.Info("Copy edit complete", "title", result.Title, "changes", len(result.Changes))

		return copyEditFileCompleteMsg{result: result}
	}
}

func (cef *copyEditFilePhase) applyChangesCmd() tea.Cmd {
	return func() tea.Msg {
		// Write the copy-edited content back to the file
		//nolint:gosec // Blog posts need to be readable
		if err := os.WriteFile(cef.filePath, []byte(cef.result.Markdown), 0o644); err != nil {
			slog.Error("Failed to write file", "error", err, "path", cef.filePath)
			return copyEditFileErrorMsg{err: fmt.Errorf("failed to write file %s: %w", cef.filePath, err)}
		}

		slog.Info("Changes applied", "path", cef.filePath)

		// Transition to completed state and quit
		cef.state = copyEditFileCompleted
		cef.applied = true

		return tea.Quit()
	}
}
