package coder

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all keyboard shortcuts
type KeyMap struct {
	Quit        key.Binding
	Help        key.Binding
	Submit      key.Binding
	Cancel      key.Binding
	CycleFocus  key.Binding
	ClearChat   key.Binding
	OpenFile    key.Binding
	ScrollUp    key.Binding
	ScrollDown  key.Binding
	PageUp      key.Binding
	PageDown    key.Binding
	GoToTop     key.Binding
	GoToBottom  key.Binding
}

// DefaultKeyMap returns the default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "send"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		CycleFocus: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "cycle focus"),
		),
		ClearChat: key.NewBinding(
			key.WithKeys("ctrl+l"),
			key.WithHelp("ctrl+l", "clear chat"),
		),
		OpenFile: key.NewBinding(
			key.WithKeys("ctrl+o"),
			key.WithHelp("ctrl+o", "open file"),
		),
		ScrollUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "scroll up"),
		),
		ScrollDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "scroll down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "ctrl+u"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "ctrl+d"),
			key.WithHelp("pgdn", "page down"),
		),
		GoToTop: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("home/g", "go to top"),
		),
		GoToBottom: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("end/G", "go to bottom"),
		),
	}
}

// ShortHelp returns a short help string
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Submit, k.CycleFocus, k.Help, k.Quit}
}

// FullHelp returns all keybindings for the help view
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Submit, k.Cancel, k.CycleFocus},
		{k.ScrollUp, k.ScrollDown, k.PageUp, k.PageDown},
		{k.ClearChat, k.OpenFile},
		{k.Help, k.Quit},
	}
}
