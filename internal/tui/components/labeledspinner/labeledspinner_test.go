package labeledspinner_test

import (
	"testing"

	"github.com/alkime/memos/internal/tui/components/labeledspinner"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/assert"
)

//nolint:gochecknoinits // recommend for CI by bubbletea folks
func init() {
	lipgloss.SetColorProfile(termenv.Ascii)
}

func TestLabeledSpinner(t *testing.T) {
	m := labeledspinner.New(spinner.Dot, "Title", "Subtitle", "Help")
	t.Run("initial state", func(t *testing.T) {
		assert.Equal(t, "Title", m.Title)
		assert.Equal(t, "Subtitle", m.Subtitle)
		assert.Equal(t, "Help", m.Help)
		assert.Equal(t, spinner.Dot, m.Spinner.Spinner)
	})

	v0 := m.View()
	t.Run("view output", func(t *testing.T) {
		assert.Contains(t, v0, "Title")
		assert.Contains(t, v0, "Subtitle")
		assert.Contains(t, v0, "Help")
		assert.Contains(t, v0, spinner.Dot.Frames[0])
	})

	t.Run("check updates", func(t *testing.T) {
		assert.Contains(t, v0, spinner.Dot.Frames[0])
		m, _ = m.Update(spinner.TickMsg{})
		v1 := m.View()
		assert.Contains(t, v1, spinner.Dot.Frames[1])
		m, _ = m.Update(spinner.TickMsg{})
		v2 := m.View()
		assert.Contains(t, v2, spinner.Dot.Frames[2])
	})
}
