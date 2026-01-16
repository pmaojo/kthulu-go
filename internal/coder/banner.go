package coder

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Banner returns the KTHULU ASCII art with gradient
func Banner(width int) string {
	// KTHULU ASCII Art (Compact style)
	art := []string{
		` __  __  _____  _   _  __ __  _    __ __ `,
		`|  |/  /|_   _|| |_| ||  |  || |__|  |  |`,
		`|__|\__\  |_|  |_| |_| \___/ |____|\___/ `,
	}

	// Gradient colors (OrangeRed to Arcane Violet)
	colors := []string{
		"#FF4500", "#FF5633", "#FF6666", "#FF7699", "#FF86CC", 
		"#C364E0", "#A844F0", "#9D00FF", "#8A00E0", "#7600C2",
	}

	var sb strings.Builder
	
	// Left align with small margin
	padding := 2
	leftPad := strings.Repeat(" ", padding)

	for i, line := range art {
		// Calculate color for this line
		colorIdx := (i * len(colors)) / len(art)
		if colorIdx >= len(colors) {
			colorIdx = len(colors) - 1
		}
		
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(colors[colorIdx]))
		sb.WriteString(leftPad + style.Render(line) + "\n")
	}

	// Add subtitle
	sub := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6C7086")).
		Italic(true).
		Render("The Software Foundry")
	
	sb.WriteString("\n" + leftPad + sub + "\n\n")

	return sb.String()
}
