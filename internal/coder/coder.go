package coder

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the Kthulu Coder TUI
func Run(workingDir string, modelName string) ([]Message, error) {
	if workingDir == "" {
		var err error
		workingDir, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	if modelName == "" {
		modelName = "gemini-2.5-flash"
	}

	m := New(workingDir, modelName)

	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),       // Use alternate screen buffer
		tea.WithMouseCellMotion(), // Enable mouse support
	)

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	if finalModel, ok := finalModel.(Model); ok {
		return finalModel.GetMessages(), nil
	}

	return nil, fmt.Errorf("could not retrieve final model state")
}

