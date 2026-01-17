package coder

import tea "github.com/charmbracelet/bubbletea"

// Message represents a chat message
type Message struct {
	Role       string
	Content    string
	ToolCalls  []ToolCall
	ToolCallID string
}

// StreamStartMsg indicates streaming has begun
type StreamStartMsg struct{}

// StreamChunkMsg contains a text chunk from the AI
type StreamChunkMsg struct {
	Content string
}

// StreamDoneMsg indicates streaming is complete
type StreamDoneMsg struct {
	FullResponse string
}

// StreamErrorMsg indicates an error during streaming
type StreamErrorMsg struct {
	Error error
}

// StreamStartedMsg is sent when the stream actually starts (from LLM)
type StreamStartedMsg struct {
	EventChan <-chan ChatEvent
}

// ToolUseMsg indicates the AI wants to use a tool
type ToolUseMsg struct {
	ToolID   string
	ToolName string
	Args     map[string]any
}

// ToolResultMsg contains the result of a tool execution
type ToolResultMsg struct {
	ToolID  string
	Success bool
	Output  string
	Error   error
}

// ToolApprovalMsg is sent when user approves/denies a tool
type ToolApprovalMsg struct {
	ToolID   string
	Approved bool
}

// StatusUpdateMsg updates the status bar
type StatusUpdateMsg struct {
	Status string
}

// SidecarReadyMsg indicates LiteLLM is ready
type SidecarReadyMsg struct{}

// SidecarErrorMsg indicates LiteLLM failed to start
type SidecarErrorMsg struct {
	Error error
}

// startStreaming initiates a chat completion stream
func (m *Model) startStreaming() tea.Cmd {
	return func() tea.Msg {
		return StreamStartedMsg{}
	}
}

// processStream handles the streaming response
func (m *Model) processStream(textChan <-chan string, errChan <-chan error) tea.Cmd {
	return func() tea.Msg {
		select {
		case text, ok := <-textChan:
			if !ok {
				return StreamDoneMsg{}
			}
			return StreamChunkMsg{Content: text}
		case err := <-errChan:
			if err != nil {
				return StreamErrorMsg{Error: err}
			}
			return StreamDoneMsg{}
		}
	}
}
