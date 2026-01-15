package workflow

import (
	"io"

	"github.com/alkime/memos/internal/content"
	tea "github.com/charmbracelet/bubbletea"
)

// Transcriber transcribes audio to text.
type Transcriber interface {
	TranscribeFile(audioFile io.Reader) (string, error)
}

// Writer generates AI content from transcripts and drafts.
type Writer interface {
	GenerateFirstDraft(transcript string, mode content.Mode) (string, error)
	GenerateCopyEdit(firstDraft, currentDate string, mode content.Mode) (*content.CopyEditResult, error)
}

// EditorLauncher opens files in an external editor.
type EditorLauncher interface {
	Launch(filePath string) tea.Cmd
}
