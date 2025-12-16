package phases

import (
	"strings"

	"github.com/alkime/memos/internal/tui/style"
	"github.com/charmbracelet/bubbles/key"
)

func renderKeyHelp(keyBinding key.Binding, suffix ...string) string {
	s := style.Help.Render("[") + style.Key.Render(keyBinding.Help().Key) +
		style.Help.Render("] ") +
		style.Help.Render(keyBinding.Help().Desc)

	s += strings.Join(suffix, "")

	return s
}

func renderGlobalKeyHelp() string {
	km := DefaultKeyMap()
	s := renderKeyHelp(km.Quit, " ")
	s += renderKeyHelp(km.ForceQuit, "\n")
	return s
}
