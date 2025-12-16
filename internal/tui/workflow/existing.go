package workflow

import (
	"os"
	"strings"

	"github.com/alkime/memos/internal/tui/style"
	"github.com/charmbracelet/bubbles/key"
)

// existingOutputKeyMap defines keys for handling existing output files.
type existingOutputKeyMap struct {
	UseExisting key.Binding
	Redo        key.Binding
}

func defaultExistingOutputKeyMap() existingOutputKeyMap {
	return existingOutputKeyMap{
		UseExisting: key.NewBinding(
			key.WithKeys("enter", "y"),
			key.WithHelp("enter/y", "use existing"),
		),
		Redo: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "redo"),
		),
	}
}

// existingOutputState tracks whether a phase's output file already exists.
type existingOutputState struct {
	found bool
	path  string
	keys  existingOutputKeyMap
}

// newExistingOutputState creates a new existing output state by checking if the file exists.
func newExistingOutputState(outputPath string) existingOutputState {
	state := existingOutputState{
		path: outputPath,
		keys: defaultExistingOutputKeyMap(),
	}

	if _, err := os.Stat(outputPath); err == nil {
		state.found = true
	}

	return state
}

// renderExistingOutputView renders the UI when an output file already exists.
func renderExistingOutputView(state existingOutputState, fileDescription string) string {
	var sb strings.Builder

	sb.WriteString(style.Success.Render("✓ " + fileDescription + " already exists!"))
	sb.WriteString("\n\n")

	sb.WriteString(style.Label.Render("File: "))
	sb.WriteString(style.Muted.Render(state.path))
	sb.WriteString("\n\n")

	sb.WriteString(renderKeyHelp(state.keys.UseExisting, " "))
	sb.WriteString(renderKeyHelp(state.keys.Redo, "\n"))
	sb.WriteString(renderGlobalKeyHelp())

	return sb.String()
}
