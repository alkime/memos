package recording

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type CapturedPacketMsg struct {
	Samples []byte
	Bytes   int64
	Elapsed int64
}

type Model struct {
	samples []int16 // 16-bit coming from the audio.Device

	totalBytes int64 // sent from Recorder in CapturedPacketMsg
	maxBytes   int64 // configured limit

	elapsed     time.Duration // sent from Recorder in CapturedPacketMsg
	maxDuration time.Duration // configured limit
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case CapturedPacketMsg:
		m.elapsed = time.Duration(msg.Elapsed)
		m.totalBytes = msg.Bytes
		return m, nil
	default:
		return m, nil
	}
}

func (m Model) View() string {
	// todo: make this more bester.
	return fmt.Sprintf("[RECORDING %s +|+ %s]\n",
		formatDuration(m.elapsed, m.maxDuration, 90),
		formatBytes(m.totalBytes, m.maxBytes, 90),
	)
}

// format functions copy n pasted from recorder.go.
// todo: clean these up after this lands

// formatWithBold wraps text in ANSI bold codes if shouldBold is true.
// this is leftover, we should definitely move to using bubbletea style system (called lipgloss ofc)
func formatWithBold(text string, shouldBold bool) string {
	if shouldBold {
		return fmt.Sprintf("\033[1m%s\033[0m", text)
	}

	return text
}

// formatDuration formats elapsed and maxDuration duration with optional bold.
func formatDuration(elapsed, maxDuration time.Duration, thresholdPerc int) string {
	// Format as HH:MM:SS
	formatTime := func(d time.Duration) string {
		h := int(d.Hours())
		m := int(d.Minutes()) % 60
		s := int(d.Seconds()) % 60
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}

	elapsedStr := formatTime(elapsed)
	maxStr := formatTime(maxDuration)
	percent := int(float64(elapsed) / float64(maxDuration) * 100)

	text := fmt.Sprintf("%s / %s (%d%%)", elapsedStr, maxStr, percent)

	return formatWithBold(text, percent >= thresholdPerc)
}

// formatBytes formats bytes in MB with optional bold.
func formatBytes(current, maxBytes int64, thresholdPerc int) string {
	currentMB := float64(current) / (1024 * 1024)
	maxMB := float64(maxBytes) / (1024 * 1024)
	percent := int(float64(current) / float64(maxBytes) * 100)

	text := fmt.Sprintf("%.1f MB / %.1f MB (%d%%)", currentMB, maxMB, percent)

	return formatWithBold(text, percent >= thresholdPerc)
}
