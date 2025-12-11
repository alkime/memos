// Package msg defines shared message types for TUI phase transitions.
package msg

// RecordingCompleteMsg signals that audio recording has finished.
type RecordingCompleteMsg struct {
	AudioPath string
}

// TranscriptionCompleteMsg signals successful transcription.
type TranscriptionCompleteMsg struct {
	Transcript string
	OutputPath string
}

// TranscriptionErrorMsg signals a transcription failure.
type TranscriptionErrorMsg struct {
	Err error
}

// FirstDraftCompleteMsg signals successful first draft generation.
type FirstDraftCompleteMsg struct {
	DraftPath string
	Content   string
}

// FirstDraftErrorMsg signals a first draft generation failure.
type FirstDraftErrorMsg struct {
	Err error
}

// ProceedToFirstDraftMsg signals user wants to generate first draft.
type ProceedToFirstDraftMsg struct{}

// SkipFirstDraftMsg signals user wants to skip first draft generation.
type SkipFirstDraftMsg struct{}

// AudioFinalizingCompleteMsg signals that MP3 conversion is complete.
type AudioFinalizingCompleteMsg struct {
	AudioPath string
}

// AudioFinalizingErrorMsg signals that MP3 conversion failed.
type AudioFinalizingErrorMsg struct {
	Err error
}

// RetryMsg signals user wants to retry the failed operation.
type RetryMsg struct{}

// OpenEditorCompleteMsg signals the editor process completed.
type OpenEditorCompleteMsg struct {
	FilePath string
	Err      error
}
