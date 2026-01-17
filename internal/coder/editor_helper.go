package coder

import (
    "os"
    "os/exec"
    
    tea "github.com/charmbracelet/bubbletea"
)

type editorFinishedMsg struct {
	err  error
	path string
}

func openEditor(content string) tea.Cmd {
	file, err := os.CreateTemp("", "kthulu-*.md")
	if err != nil {
		return func() tea.Msg { return nil } // simplified error handling
	}
	
	if _, err := file.WriteString(content); err != nil {
		return func() tea.Msg { return nil }
	}
	file.Close()

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	c := exec.Command(editor, file.Name())
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err: err, path: file.Name()}
	})
}
