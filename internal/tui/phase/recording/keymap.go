package recording

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines the key bindings for the recording phase.
type KeyMap struct {
	Toggle key.Binding
	Finish key.Binding
}

// DefaultKeyMap returns the default key bindings for the recording phase.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Toggle: key.NewBinding(
			key.WithKeys("space"),
			key.WithHelp("space", "start/stop recording"),
		),
		Finish: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "finish recording"),
		),
	}
}

// ShortHelp returns the short help bindings for the recording phase.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Toggle, k.Finish}
}

// FullHelp returns the full help bindings for the recording phase.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Toggle, k.Finish},
	}
}
