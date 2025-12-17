// Package waveform provides a TUI component for visualizing audio amplitude.
package waveform

import (
	"strings"
	"time"

	"github.com/alkime/memos/internal/tui/style"
	"github.com/alkime/memos/pkg/uictl"
	tea "github.com/charmbracelet/bubbletea"
)

// Block characters for amplitude visualization (8 levels, bottom to top).
// Index 0 = empty (space), 1-8 = increasing fill levels.
const blockChars = " ▁▂▃▄▅▆▇█"

// TickMsg triggers a waveform redraw.
type TickMsg struct{}

// Model displays an oscilloscope-style waveform visualization.
// It reads audio samples from a Levels control and renders them
// as vertical bars showing amplitude over time (left=older, right=newer).
type Model struct {
	levels uictl.Levels[int16] // Data source for samples
	width  int                 // Display width in characters
	height int                 // Display height in rows
}

// New creates a new waveform model.
// The width parameter determines how many columns to render.
// The height parameter determines how many rows tall the waveform is.
// Samples are aggregated to fit the display width.
func New(levels uictl.Levels[int16], width, height int) Model {
	if height < 1 {
		height = 1
	}

	return Model{
		levels: levels,
		width:  width,
		height: height,
	}
}

// Init returns the initial tick command.
func (m Model) Init() tea.Cmd {
	return m.tick()
}

// Update handles tick messages for animation.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if _, ok := msg.(TickMsg); ok {
		return m, m.tick()
	}

	return m, nil
}

// View renders the waveform as ASCII art.
func (m Model) View() string {
	if m.levels == nil {
		return m.renderEmpty()
	}

	samples := m.levels.Read()
	if len(samples) == 0 {
		return m.renderEmpty()
	}

	return m.renderWaveform(samples)
}

// tick schedules the next waveform update at ~20 FPS.
func (m Model) tick() tea.Cmd {
	return tea.Tick(50*time.Millisecond, func(_ time.Time) tea.Msg {
		return TickMsg{}
	})
}

// renderWaveform renders audio samples as vertical bars across multiple rows.
func (m Model) renderWaveform(samples []int16) string {
	// Calculate amplitude level (0 to height*8) for each column
	levels := m.calculateLevels(samples)
	runes := []rune(blockChars)

	var sb strings.Builder

	// Render row by row, from top to bottom
	for row := 0; row < m.height; row++ {
		if row > 0 {
			sb.WriteString("\n")
		}

		var rowSB strings.Builder

		for col := 0; col < m.width; col++ {
			level := levels[col]
			blockIdx := m.blockIndexForRow(level, row)
			rowSB.WriteRune(runes[blockIdx])
		}

		sb.WriteString(style.Progress.Render(rowSB.String()))
	}

	return sb.String()
}

// calculateLevels computes amplitude levels for each column.
// Returns a slice of levels from 0 to height*8.
func (m Model) calculateLevels(samples []int16) []int {
	levels := make([]int, m.width)
	bucketSize := max(1, len(samples)/m.width)
	maxLevel := m.height * 8

	for col := 0; col < m.width; col++ {
		start := col * bucketSize
		if start >= len(samples) {
			levels[col] = 0

			continue
		}

		end := min(start+bucketSize, len(samples))
		maxAmp := maxAbsAmplitude(samples[start:end])

		// Map amplitude to 0..maxLevel using the perceptual curve
		levels[col] = amplitudeToMultiRowLevel(maxAmp, maxLevel)
	}

	return levels
}

// blockIndexForRow returns the block character index (0-8) for a given column level at a row.
// Row 0 is the top, row (height-1) is the bottom.
func (m Model) blockIndexForRow(level, row int) int {
	// Calculate the "base" level for this row (bottom of this row's range)
	// Row 0 (top) covers levels [(height-1)*8, height*8]
	// Row 1 covers [(height-2)*8, (height-1)*8]
	// Row (height-1) (bottom) covers [0, 8]
	rowFromBottom := m.height - 1 - row
	baseLevel := rowFromBottom * 8

	// How much of this row is filled?
	fillAmount := level - baseLevel

	if fillAmount <= 0 {
		return 0 // Empty (space)
	}

	if fillAmount >= 8 {
		return 8 // Full block
	}

	return fillAmount // Partial block (1-7)
}

// renderEmpty renders empty space for when there are no samples.
func (m Model) renderEmpty() string {
	var sb strings.Builder

	for row := 0; row < m.height; row++ {
		if row > 0 {
			sb.WriteString("\n")
		}

		var rowSB strings.Builder

		for i := 0; i < m.width; i++ {
			if row == m.height-1 {
				// Bottom row shows baseline
				rowSB.WriteRune('▁')
			} else {
				rowSB.WriteRune(' ')
			}
		}

		sb.WriteString(style.Muted.Render(rowSB.String()))
	}

	return sb.String()
}

// maxAbsAmplitude returns the maximum absolute amplitude in a slice of samples.
func maxAbsAmplitude(samples []int16) int16 {
	var maxAmp int16

	for _, s := range samples {
		// Handle int16 overflow: -32768 has no positive equivalent
		if s == -32768 {
			return 32767 // Max possible amplitude
		}

		if s < 0 {
			s = -s
		}

		if s > maxAmp {
			maxAmp = s
		}
	}

	return maxAmp
}

// amplitudeToMultiRowLevel maps an amplitude (0-32767) to a display level (0-maxLevel).
// Uses a logarithmic-like scale for better visual perception of volume.
func amplitudeToMultiRowLevel(amp int16, maxLevel int) int {
	if amp == 0 {
		return 0
	}

	// Use square root for perceptual scaling (quieter sounds are more visible)
	// sqrt(32767) ≈ 181, so we scale by maxLevel/181
	// This gives a nice curve where quiet audio is still visible
	const maxAmp = 32767.0

	// Normalize to 0-1, apply sqrt for perception, scale to maxLevel
	normalized := float64(amp) / maxAmp
	scaled := sqrt(normalized) * float64(maxLevel)

	return min(int(scaled), maxLevel)
}

// sqrt computes square root without importing math package.
func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}

	// Newton's method for square root
	z := x / 2

	for range 10 {
		z = (z + x/z) / 2
	}

	return z
}
