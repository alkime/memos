package workflow

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/alkime/memos/internal/content"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/muesli/termenv"
)

func init() {
	lipgloss.SetColorProfile(termenv.Ascii)
}

// outputChecker provides helpers for testing teatest output.
type outputChecker struct {
	intervl, timeout time.Duration
}

func defaultChecker() outputChecker {
	return outputChecker{
		intervl: 100 * time.Millisecond,
		timeout: 3 * time.Second,
	}
}

func (o outputChecker) check(t *testing.T, tm *teatest.TestModel, checkFunc func(buf []byte) bool) {
	t.Helper()
	teatest.WaitFor(t, tm.Output(), checkFunc,
		teatest.WithCheckInterval(o.intervl),
		teatest.WithDuration(o.timeout))
}

func (o outputChecker) checkString(t *testing.T, tm *teatest.TestModel, substr string) {
	t.Helper()
	o.check(t, tm, func(buf []byte) bool {
		return bytes.Contains(buf, []byte(substr))
	})
}

// mockTranscriber implements Transcriber for testing.
type mockTranscriber struct {
	result string
	err    error
	called bool
}

func (m *mockTranscriber) TranscribeFile(_ io.Reader) (string, error) {
	m.called = true
	return m.result, m.err
}

// mockWriter implements Writer for testing.
type mockWriter struct {
	firstDraftResult string
	copyEditResult   *content.CopyEditResult
	err              error
	firstDraftCalled bool
	copyEditCalled   bool
}

func (m *mockWriter) GenerateFirstDraft(_ string, _ content.Mode) (string, error) {
	m.firstDraftCalled = true
	return m.firstDraftResult, m.err
}

func (m *mockWriter) GenerateCopyEdit(_, _ string, _ content.Mode) (*content.CopyEditResult, error) {
	m.copyEditCalled = true
	return m.copyEditResult, m.err
}

// mockEditorLauncher implements EditorLauncher for testing.
type mockEditorLauncher struct {
	launched bool
	filePath string
}

func (m *mockEditorLauncher) Launch(filePath string) tea.Cmd {
	m.launched = true
	m.filePath = filePath
	return func() tea.Msg {
		return editorCompleteMsg{err: nil}
	}
}

// mockKnob implements remotectl.Knob for testing.
type mockKnob struct {
	state bool
}

func (m *mockKnob) Read() bool   { return m.state }
func (m *mockKnob) On()          { m.state = true }
func (m *mockKnob) Off()         { m.state = false }
func (m *mockKnob) Toggle()      { m.state = !m.state }

// mockCappedDial implements remotectl.CappedDial[int64] for testing.
type mockCappedDial struct {
	current, max int64
}

func (m *mockCappedDial) Read() int64             { return m.current }
func (m *mockCappedDial) Cap() (int64, int64)     { return m.current, m.max }

// mockLevels implements remotectl.Levels[int16] for testing.
type mockLevels struct {
	samples []int16
}

func (m *mockLevels) Read() []int16 { return m.samples }
