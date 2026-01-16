package coder

import (
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// Theme defines the color scheme for the TUI
type Theme struct {
	Primary     lipgloss.Color
	Secondary   lipgloss.Color
	Accent      lipgloss.Color
	Background  lipgloss.Color
	Foreground  lipgloss.Color
	Muted       lipgloss.Color
	Success     lipgloss.Color
	Warning     lipgloss.Color
	Error       lipgloss.Color
	Border      lipgloss.Color
	BorderFocus lipgloss.Color
}

// InfernoTheme is the new default "Inferno" theme (Red/Orange/Gold)
var InfernoTheme = Theme{
	Primary:     lipgloss.Color("#FF4500"), // OrangeRed
	Secondary:   lipgloss.Color("#FFD700"), // Gold
	Accent:      lipgloss.Color("#DC143C"), // Crimson
	Background:  lipgloss.Color("#100505"), // Very dark red/black
	Foreground:  lipgloss.Color("#FFDAB9"), // PeachPuff
	Muted:       lipgloss.Color("#602020"), // Dark red/brown
	Success:     lipgloss.Color("#32CD32"), // LimeGreen
	Warning:     lipgloss.Color("#FFA500"), // Orange
	Error:       lipgloss.Color("#FF0000"), // Red
	Border:      lipgloss.Color("#502020"), // Dark red/brown border
	BorderFocus: lipgloss.Color("#FF4500"), // OrangeRed focus
}

// DefaultTheme aliases InfernoTheme for backward compatibility
var DefaultTheme = InfernoTheme

// Styles contains all lipgloss styles for the TUI
type Styles struct {
	// Layout
	App             lipgloss.Style
	ChatPane        lipgloss.Style
	ContextPane     lipgloss.Style
	InputPane       lipgloss.Style
	StatusBar       lipgloss.Style
	PaneTitle       lipgloss.Style
	PaneTitleActive lipgloss.Style
	Pane            lipgloss.Style

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

	// Renderer
	Markdown *glamour.TermRenderer
}

// NewStyles creates styles from a theme
func NewStyles(theme Theme) Styles {
	border := lipgloss.RoundedBorder()
	thickBorder := lipgloss.ThickBorder()

	// Initialize Glamour renderer
	// We use "dark" style as base which is standard
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(80), // Dynamic breadth usually better, but default is safe
	)

	return Styles{
		// Layout styles
		App: lipgloss.NewStyle().
			Background(theme.Background),

		ChatPane: lipgloss.NewStyle().
			Border(border).
			BorderForeground(theme.Border).
			Padding(0, 1),

		ContextPane: lipgloss.NewStyle().
			Border(border).
			BorderForeground(theme.Border).
			Padding(0, 1),

		InputPane: lipgloss.NewStyle().
			Border(thickBorder).
			BorderForeground(theme.Muted).
			Padding(0, 1),

		StatusBar: lipgloss.NewStyle().
			Background(lipgloss.Color("#200505")). // Darker red background for status
			Foreground(theme.Muted).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(theme.Border).
			Padding(0, 1).
			Bold(true),

		PaneTitle: lipgloss.NewStyle().
			Foreground(theme.Muted).
			Background(theme.Background).
			Padding(0, 1).
			Bold(true),

		PaneTitleActive: lipgloss.NewStyle().
			Foreground(theme.Background).
			Background(theme.Primary). // Inverted for active "Powerline" feel
			Padding(0, 1).
			Bold(true),

		Pane: lipgloss.NewStyle().
			Border(border).
			BorderForeground(theme.Border).
			Padding(0, 1),

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
			PaddingLeft(1).
			PaddingRight(1).
			BorderStyle(lipgloss.NormalBorder()).
			BorderLeft(true).
			BorderForeground(theme.Accent),

		AIMessage: lipgloss.NewStyle().
			Foreground(theme.Foreground).
			PaddingLeft(1).
			// No border for AI, rely on Markdown
			// But maybe a subtle left line?
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
			Background(theme.Secondary).
			Foreground(theme.Foreground).
			Bold(true),

		Help: lipgloss.NewStyle().
			Foreground(theme.Muted).
			Italic(true),

		Markdown: renderer,
	}
}

// DefaultStyles returns styles with the default theme (now Inferno)
func DefaultStyles() Styles {
	return NewStyles(InfernoTheme)
}
