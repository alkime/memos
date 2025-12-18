package waveform_test

import (
	"strings"
	"testing"

	"github.com/alkime/memos/internal/tui/components/waveform"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	lipgloss.SetColorProfile(termenv.Ascii)
}

// mockLevels implements uictl.Levels[int16] for testing.
type mockLevels struct {
	samples []int16
}

func (m *mockLevels) Read() []int16 {
	return m.samples
}

func TestWaveform_EmptyView(t *testing.T) {
	t.Parallel()

	mock := &mockLevels{samples: nil}
	m := waveform.New(mock, 5, 1)

	view := m.View()
	assert.Contains(t, view, "▁▁▁▁▁")
}

func TestWaveform_NilLevels(t *testing.T) {
	t.Parallel()

	m := waveform.New(nil, 5, 1)

	view := m.View()
	assert.Contains(t, view, "▁▁▁▁▁")
}

func TestWaveform_SilentAudio(t *testing.T) {
	t.Parallel()

	mock := &mockLevels{samples: []int16{0, 0, 0, 0, 0}}
	m := waveform.New(mock, 5, 1)

	view := m.View()
	// Silent audio shows spaces (level 0)
	assert.Contains(t, view, "     ")
}

func TestWaveform_MaxAmplitude(t *testing.T) {
	t.Parallel()

	mock := &mockLevels{samples: []int16{32767, 32767, 32767, 32767, 32767}}
	m := waveform.New(mock, 5, 1)

	view := m.View()
	assert.Contains(t, view, "█████")
}

func TestWaveform_VaryingAmplitude(t *testing.T) {
	t.Parallel()

	// Create samples that should produce varying bar heights
	mock := &mockLevels{samples: []int16{0, 8000, 32767, 8000, 0}}
	m := waveform.New(mock, 5, 1)

	view := m.View()
	// Just verify it's not all the same - middle should be different from edges
	runes := []rune(view)
	require.GreaterOrEqual(t, len(runes), 5)
	assert.NotEqual(t, runes[0], runes[2], "middle should be different from edges")
}

func TestWaveform_NegativeAmplitude(t *testing.T) {
	t.Parallel()

	// Negative samples should show as positive amplitude (absolute value)
	mock := &mockLevels{samples: []int16{-32768, -32768, -32768}}
	m := waveform.New(mock, 3, 1)

	view := m.View()
	// Should show max amplitude for max negative values
	assert.Contains(t, view, "███")
}

func TestWaveform_AggregatesSamples(t *testing.T) {
	t.Parallel()

	// More samples than width - should aggregate
	samples := make([]int16, 100)
	for i := range samples {
		samples[i] = 20000 // High amplitude
	}

	mock := &mockLevels{samples: samples}
	m := waveform.New(mock, 10, 1)

	view := m.View()
	runes := []rune(view)
	require.GreaterOrEqual(t, len(runes), 10)
}

func TestWaveform_FewerSamplesThanWidth(t *testing.T) {
	t.Parallel()

	// Fewer samples than width - should handle gracefully
	mock := &mockLevels{samples: []int16{32767, 32767, 32767}}
	m := waveform.New(mock, 10, 1)

	view := m.View()
	runes := []rune(view)
	// Should produce output of at least 10 characters
	require.GreaterOrEqual(t, len(runes), 10)
}

func TestWaveform_Update(t *testing.T) {
	t.Parallel()

	mock := &mockLevels{samples: []int16{1000, 2000, 3000}}
	m := waveform.New(mock, 5, 1)

	// Update with tick should return a command
	newM, cmd := m.Update(waveform.TickMsg{})
	assert.NotNil(t, cmd)
	assert.NotNil(t, newM)
}

func TestWaveform_Init(t *testing.T) {
	t.Parallel()

	mock := &mockLevels{samples: []int16{1000}}
	m := waveform.New(mock, 5, 1)

	// Init should return a tick command
	cmd := m.Init()
	assert.NotNil(t, cmd)
}

func TestWaveform_MultiRow(t *testing.T) {
	t.Parallel()

	// Test multi-row rendering
	mock := &mockLevels{samples: []int16{32767, 16000, 8000, 4000, 0}}
	m := waveform.New(mock, 5, 3)

	view := m.View()
	// Should contain newlines for multi-row output
	assert.Contains(t, view, "\n")

	// Count rows
	lines := strings.Split(view, "\n")
	assert.Equal(t, 3, len(lines), "should have 3 rows")
}

func TestWaveform_HeightZeroDefaultsToOne(t *testing.T) {
	t.Parallel()

	mock := &mockLevels{samples: []int16{32767}}
	m := waveform.New(mock, 5, 0)

	view := m.View()
	// Should not crash and should produce output
	assert.NotEmpty(t, view)
	// Height 0 defaults to 1, so no newlines
	assert.NotContains(t, view, "\n")
}
