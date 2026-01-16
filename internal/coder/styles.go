package coder

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme defines the color scheme for the TUI
type Theme struct {
	Primary    lipgloss.Color
	Secondary  lipgloss.Color
	Accent     lipgloss.Color
	Background lipgloss.Color
	Foreground lipgloss.Color
	Muted      lipgloss.Color
	Success    lipgloss.Color
	Warning    lipgloss.Color
	Error      lipgloss.Color
	Border     lipgloss.Color
	BorderFocus lipgloss.Color
}

// DefaultTheme is the "Inferno Black Magick" theme
var DefaultTheme = Theme{
	Primary:     lipgloss.Color("#FF4500"), // OrangeRed - Magickal Fire
	Secondary:   lipgloss.Color("#9D00FF"), // Electric Violet - Arcane Energy
	Accent:      lipgloss.Color("#FFD700"), // Gold - Ancient Runes
	Background:  lipgloss.Color("#050505"), // Void Black
	Foreground:  lipgloss.Color("#E0E0E0"), // Ashes
	Muted:       lipgloss.Color("#505050"), // Dark Charcoal
	Success:     lipgloss.Color("#00FF00"), // Lime - Vitality
	Warning:     lipgloss.Color("#FF8C00"), // Dark Orange - Ember
	Error:       lipgloss.Color("#FF0000"), // Pure Red - Blood
	Border:      lipgloss.Color("#303030"), // Obsidian
	BorderFocus: lipgloss.Color("#FF4500"), // Burning focus
}

// Styles contains all lipgloss styles for the TUI
type Styles struct {
	// Layout
	App         lipgloss.Style
	ChatPane    lipgloss.Style
	ContextPane lipgloss.Style
	InputPane   lipgloss.Style
	StatusBar   lipgloss.Style
	PaneTitle   lipgloss.Style
	PaneTitleActive lipgloss.Style

	// Text
	Title       lipgloss.Style
	Subtitle    lipgloss.Style
	Label       lipgloss.Style
	Muted       lipgloss.Style
	UserMessage lipgloss.Style
	AIMessage   lipgloss.Style
	CodeBlock   lipgloss.Style

	// Status
	Success lipgloss.Style
	Warning lipgloss.Style
	Error   lipgloss.Style

	// Interactive
	Focused  lipgloss.Style
	Selected lipgloss.Style
	Help     lipgloss.Style
}

// NewStyles creates styles from a theme
func NewStyles(theme Theme) Styles {
	border := lipgloss.RoundedBorder()
	thickBorder := lipgloss.ThickBorder()

	return Styles{
		// Layout styles
		App: lipgloss.NewStyle().
			Background(theme.Background),

		ChatPane: lipgloss.NewStyle().
			Border(border).
			BorderForeground(theme.Border).
			Padding(0, 1), // Reduced padding for cleaner look

		ContextPane: lipgloss.NewStyle().
			Border(border).
			BorderForeground(theme.Border).
			Padding(0, 1),

		InputPane: lipgloss.NewStyle().
			Border(thickBorder). // Thick border for input
			BorderForeground(theme.Muted).
			Padding(0, 1),

		StatusBar: lipgloss.NewStyle().
			Background(theme.Background).
			Foreground(theme.Muted).
			Border(lipgloss.NormalBorder(), false, false, true, false). // Bottom border only
			BorderForeground(theme.Border).
			Padding(0, 1).
			Bold(true),

		PaneTitle: lipgloss.NewStyle().
			Foreground(theme.Muted).
			Background(theme.Background).
			Padding(0, 1).
			Bold(true),

		PaneTitleActive: lipgloss.NewStyle().
			Foreground(theme.Primary).
			Background(theme.Background).
			Padding(0, 1).
			Bold(true).
			Underline(true),

		// Text styles
		Title: lipgloss.NewStyle().
			Foreground(theme.Primary).
			Bold(true),

		Subtitle: lipgloss.NewStyle().
			Foreground(theme.Secondary).
			Bold(true),

		Label: lipgloss.NewStyle().
			Foreground(theme.Muted),

		Muted: lipgloss.NewStyle().
			Foreground(theme.Muted),

		UserMessage: lipgloss.NewStyle().
			Foreground(theme.Accent).
			PaddingLeft(2).
			BorderStyle(lipgloss.NormalBorder()).
			BorderLeft(true).
			BorderForeground(theme.Accent),

		AIMessage: lipgloss.NewStyle().
			Foreground(theme.Foreground).
			PaddingLeft(2).
			BorderStyle(lipgloss.NormalBorder()).
			BorderLeft(true).
			BorderForeground(theme.Secondary),
		
		CodeBlock: lipgloss.NewStyle().
			Background(lipgloss.Color("#111111")).
			Foreground(lipgloss.Color("#A9A9A9")).
			Padding(0, 1),

		// Status styles
		Success: lipgloss.NewStyle().
			Foreground(theme.Success),

		Warning: lipgloss.NewStyle().
			Foreground(theme.Warning),

		Error: lipgloss.NewStyle().
			Foreground(theme.Error),

		// Interactive styles
		Focused: lipgloss.NewStyle().
			BorderForeground(theme.BorderFocus),

		Selected: lipgloss.NewStyle().
			Background(theme.Secondary). // Purple selection
			Foreground(theme.Foreground).
			Bold(true),

		Help: lipgloss.NewStyle().
			Foreground(theme.Muted).
			Italic(true),
	}
}

// DefaultStyles returns styles with the default theme
func DefaultStyles() Styles {
	return NewStyles(DefaultTheme)
}
