package coder

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmationModal handles tool approval/denial
type ConfirmationModal struct {
	visible    bool
	toolName   string
	toolArgs   map[string]any
	toolID     string
	width      int
	height     int
	selected   int // 0=Approve, 1=Deny, 2=Always
	styles     ConfirmationStyles
}

// ConfirmationStyles defines the modal appearance
type ConfirmationStyles struct {
	Overlay     lipgloss.Style
	Modal       lipgloss.Style
	Title       lipgloss.Style
	Content     lipgloss.Style
	Button      lipgloss.Style
	ButtonFocus lipgloss.Style
	Warning     lipgloss.Style
}

// DefaultConfirmationStyles returns default modal styles
func DefaultConfirmationStyles() ConfirmationStyles {
	return ConfirmationStyles{
		Overlay: lipgloss.NewStyle().
			Background(DefaultTheme.Background),
		Modal: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(DefaultTheme.Primary).
			Padding(1, 2).
			Background(lipgloss.Color("#100505")), // Slightly lighter than void for depth
		Title: lipgloss.NewStyle().
			Foreground(DefaultTheme.Primary).
			Bold(true).
			MarginBottom(1),
		Content: lipgloss.NewStyle().
			Foreground(DefaultTheme.Foreground),
		Button: lipgloss.NewStyle().
			Foreground(DefaultTheme.Muted).
			Padding(0, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(DefaultTheme.Border),
		ButtonFocus: lipgloss.NewStyle().
			Foreground(DefaultTheme.Background).
			Background(DefaultTheme.Primary).
			Bold(true).
			Padding(0, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(DefaultTheme.Primary),
		Warning: lipgloss.NewStyle().
			Foreground(DefaultTheme.Warning).
			Bold(true),
	}
}


// NewConfirmationModal creates a new confirmation modal
func NewConfirmationModal() *ConfirmationModal {
	return &ConfirmationModal{
		styles: DefaultConfirmationStyles(),
	}
}

// Show displays the modal for a tool
func (m *ConfirmationModal) Show(toolID, toolName string, args map[string]any) {
	m.visible = true
	m.toolID = toolID
	m.toolName = toolName
	m.toolArgs = args
	m.selected = 0
}

// Hide hides the modal
func (m *ConfirmationModal) Hide() {
	m.visible = false
}

// IsVisible returns whether the modal is showing
func (m *ConfirmationModal) IsVisible() bool {
	return m.visible
}

// SetSize updates the modal dimensions
func (m *ConfirmationModal) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Update handles key events for the modal
func (m *ConfirmationModal) Update(msg tea.Msg) (approved bool, handled bool, cmd tea.Cmd) {
	if !m.visible {
		return false, false, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("left", "h"))):
			if m.selected > 0 {
				m.selected--
			}
			return false, true, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("right", "l"))):
			if m.selected < 2 {
				m.selected++
			}
			return false, true, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("a"))):
			m.selected = 0
			m.Hide()
			return true, true, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("d"))):
			m.selected = 1
			m.Hide()
			return false, true, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			approved := m.selected == 0 || m.selected == 2
			m.Hide()
			return approved, true, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			m.Hide()
			return false, true, nil
		}
	}

	return false, false, nil
}

// View renders the modal
func (m *ConfirmationModal) View() string {
	if !m.visible {
		return ""
	}

	// Build content
	var content strings.Builder

	// Title
	title := m.styles.Title.Render(fmt.Sprintf("⚠️  Tool Confirmation: %s", m.toolName))
	content.WriteString(title + "\n\n")

	// Tool details
	content.WriteString(m.styles.Content.Render("This tool wants to execute:\n"))
	
	// Format args nicely
	for k, v := range m.toolArgs {
		argStr := fmt.Sprintf("  • %s: %v", k, truncateArg(v, 60))
		content.WriteString(m.styles.Content.Render(argStr) + "\n")
	}
	content.WriteString("\n")

	// Warning for dangerous tools
	if m.toolName == "bash" || m.toolName == "write_file" {
		warning := m.styles.Warning.Render("⚠️  This action may modify your system!")
		content.WriteString(warning + "\n\n")
	}

	// Buttons
	buttons := m.renderButtons()
	content.WriteString(buttons + "\n")

	// Help text
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6C7086")).
		Render("←/→ navigate • Enter confirm • A approve • D deny • Esc cancel")
	content.WriteString("\n" + help)

	// Calculate modal size
	modalWidth := m.width * 60 / 100
	if modalWidth < 50 {
		modalWidth = 50
	}
	if modalWidth > 80 {
		modalWidth = 80
	}

	// Render modal box
	modal := m.styles.Modal.
		Width(modalWidth).
		Render(content.String())

	// Center in viewport
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		modal,
	)
}

func (m *ConfirmationModal) renderButtons() string {
	buttons := []string{"✓ Approve", "✗ Deny", "✓ Always"}
	rendered := make([]string, len(buttons))

	for i, btn := range buttons {
		if i == m.selected {
			rendered[i] = m.styles.ButtonFocus.Render(btn)
		} else {
			rendered[i] = m.styles.Button.Render(btn)
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, rendered...)
}

func truncateArg(v any, maxLen int) string {
	s := fmt.Sprintf("%v", v)
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}
