package viewtranscript

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines the key bindings for the view transcript phase.
type KeyMap struct {
	Proceed key.Binding
	Skip    key.Binding
}

// DefaultKeyMap returns the default key bindings for the view transcript phase.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Proceed: key.NewBinding(
			key.WithKeys("y", "enter"),
			key.WithHelp("y/enter", "generate first draft"),
		),
		Skip: key.NewBinding(
			key.WithKeys("n", "s"),
			key.WithHelp("n/s", "skip"),
		),
	}
}

// ShortHelp returns the short help bindings for the view transcript phase.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Proceed, k.Skip}
}

// FullHelp returns the full help bindings for the view transcript phase.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Proceed, k.Skip},
	}
}
