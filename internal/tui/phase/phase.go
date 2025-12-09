// Package phase defines the TUI workflow phases.
package phase

// Phase represents the current state of the TUI workflow.
type Phase int

const (
	// PhaseRecording is the initial state where audio is being captured.
	PhaseRecording Phase = iota
	// PhaseFinalizingAudio is when waiting for MP3 conversion to complete.
	PhaseFinalizingAudio
	// PhaseTranscribing is when the audio is being sent to Whisper API.
	PhaseTranscribing
	// PhaseViewTranscript displays the transcript for user review.
	PhaseViewTranscript
	// PhaseGeneratingDraft is when the transcript is being processed by Claude.
	PhaseGeneratingDraft
	// PhaseComplete indicates the workflow finished successfully.
	PhaseComplete
	// PhaseError indicates an error occurred during the workflow.
	PhaseError
)

// String returns the human-readable name of the phase.
func (p Phase) String() string {
	switch p {
	case PhaseRecording:
		return "Recording"
	case PhaseFinalizingAudio:
		return "Finalizing Audio"
	case PhaseTranscribing:
		return "Transcribing"
	case PhaseViewTranscript:
		return "View Transcript"
	case PhaseGeneratingDraft:
		return "Generating Draft"
	case PhaseComplete:
		return "Complete"
	case PhaseError:
		return "Error"
	default:
		return "Unknown"
	}
}
