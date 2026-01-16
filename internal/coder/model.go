package coder

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Pane represents which pane is currently focused
type Pane int

const (
	PaneChat Pane = iota
	PaneContext
	PaneInput
)

// Message represents a chat message
type Message struct {
	Role    string // "user" or "assistant"
	Content string
}

// Model is the main Bubble Tea model for the coder TUI
type Model struct {
	// Window dimensions
	width  int
	height int

	// Panes
	chatViewport    viewport.Model
	contextViewport viewport.Model
	inputArea       textarea.Model
	help            help.Model
	spinner         spinner.Model

	// Modals
	confirmModal *ConfirmationModal
	filePicker   *FilePicker

	// State
	focusedPane      Pane
	messages         []Message
	isLoading        bool
	modelName        string
	workingDir       string
	streamingContent string // Current streaming response
	statusMessage    string
	contextFiles     []string // Files added to context

	// LiteLLM
	llmClient   *LiteLLMClient
	ctx         context.Context
	cancelFunc  context.CancelFunc
	textChan    <-chan string
	errChan     <-chan error

	// Styling
	styles Styles
	keys   KeyMap

	// Flags
	showHelp bool
	quitting bool
}


// New creates a new Model with default settings
func New(workingDir string, modelName string) Model {
	// Create textarea for input
	ta := textarea.New()
	ta.Placeholder = "Ask general questions..."
	ta.Focus()
	ta.Prompt = "│ "
	ta.CharLimit = 4000
	ta.SetWidth(60)
	ta.SetHeight(1)
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false) // Enter sends, Shift+Enter for newline

	// Create viewports (will be resized on WindowSizeMsg)
	chatVp := viewport.New(80, 20)
	chatVp.SetContent(Banner(80) + "\nType a message to get started.")

	contextVp := viewport.New(40, 20)
	contextVp.SetContent("Your Grimoire is empty.\n\nPress Ctrl+O to add scrolls (files)\nfor the Spirit to read.")

	// Create spinner for loading state
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(DefaultTheme.Primary)

	// Create LiteLLM client
	llmConfig := DefaultLiteLLMConfig(modelName)
	llmClient := NewLiteLLMClient(llmConfig)

	ctx, cancel := context.WithCancel(context.Background())

	return Model{
		chatViewport:    chatVp,
		contextViewport: contextVp,
		inputArea:       ta,
		help:            help.New(),
		spinner:         s,
		confirmModal:    NewConfirmationModal(),
		filePicker:      NewFilePicker(workingDir),
		focusedPane:     PaneInput,
		messages:        []Message{},
		modelName:       modelName,
		workingDir:      workingDir,
		styles:          DefaultStyles(),
		keys:            DefaultKeyMap(),
		llmClient:       llmClient,
		ctx:             ctx,
		cancelFunc:      cancel,
		statusMessage:   "Ready",
		contextFiles:    []string{},
	}
}


// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		m.spinner.Tick,
		tea.EnableMouseCellMotion,
		m.startLiteLLMCmd(),
	)
}

type sidecarStartedMsg struct{}

func (m Model) startLiteLLMCmd() tea.Cmd {
	return func() tea.Msg {
		// Attempt to start sidecar
		// We use a separate context for startup to avoid race with model context
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		
		if err := m.llmClient.StartSidecar(ctx); err != nil {
			// If sidecar fails, we might just be in a mode where we expect it running?
			// Or we return error.
			return StreamErrorMsg{Error: fmt.Errorf("AI Engine failed to start: %v", err)}
		}
		return sidecarStartedMsg{}
	}
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle modals first (they capture input)
	if m.confirmModal.IsVisible() {
		approved, handled, cmd := m.confirmModal.Update(msg)
		if handled {
			if approved {
				m.statusMessage = "Tool approved"
			} else {
				m.statusMessage = "Tool denied"
			}
			return m, cmd
		}
	}

	if m.filePicker.IsVisible() {
		selected, done, cmd := m.filePicker.Update(msg)
		if done {
			if selected != "" {
				m.addFileToContext(selected)
			}
			return m, cmd
		}
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case msg.String() == "ctrl+c":
			m.quitting = true
			if m.llmClient != nil {
				m.llmClient.StopSidecar()
			}
			return m, tea.Quit

		case msg.String() == "ctrl+o":
			// Open file picker
			m.filePicker.SetSize(m.width, m.height)
			return m, m.filePicker.Show()

		case msg.String() == "tab":
			m.cycleFocus()
			return m, nil

		case msg.String() == "?":
			if m.focusedPane != PaneInput {
				m.showHelp = !m.showHelp
			}

		case msg.String() == "ctrl+e":
			// Open external editor
			return m, openEditor(m.inputArea.Value())

		case msg.String() == "ctrl+l":
			m.messages = []Message{}
			m.chatViewport.SetContent(Banner(m.chatViewport.Width))
			return m, nil

		case msg.String() == "enter":
			if m.focusedPane == PaneInput && strings.TrimSpace(m.inputArea.Value()) != "" && !m.isLoading {
				content := m.inputArea.Value()
				m.messages = append(m.messages, Message{Role: "user", Content: content})
				m.updateChatContent()
				m.inputArea.Reset()
				m.isLoading = true
				m.streamingContent = ""
				m.statusMessage = "Thinking..."
				
				// Start streaming from LiteLLM
				return m, m.sendToLLM()
			}

		case msg.String() == "esc":
			if m.showHelp {
				m.showHelp = false
				return m, nil
			}
		}

	case tea.MouseMsg:
		if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
			return m, nil
		}
		
		// Simple click detection based on layout
		// Context pane: left 30%
		// Input pane: bottom 3 lines + border
		inputY := m.height - 3
		contextWidth := m.width * 30 / 100

		if msg.Y >= inputY {
			if m.focusedPane != PaneInput {
				m.focusedPane = PaneInput
				m.inputArea.Focus()
			}
		} else if msg.X < contextWidth {
			if m.focusedPane != PaneContext {
				m.focusedPane = PaneContext
				m.inputArea.Blur()
			}
		} else {
			if m.focusedPane != PaneChat {
				m.focusedPane = PaneChat
				m.inputArea.Blur()
			}
		}
		return m, nil


	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()
		m.confirmModal.SetSize(m.width, m.height)
		m.filePicker.SetSize(m.width, m.height)
		return m, nil


	case spinner.TickMsg:
		if m.isLoading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case StreamChunkMsg:
		m.streamingContent += msg.Content
		m.updateChatContentStreaming()
		m.chatViewport.GotoBottom()
		// Continue receiving chunks
		return m, m.receiveStreamChunk()

	case StreamDoneMsg:
		m.isLoading = false
		m.statusMessage = "Ready"
		// Finalize the assistant message
		if m.streamingContent != "" {
			m.messages = append(m.messages, Message{
				Role:    "assistant",
				Content: m.streamingContent,
			})
			m.streamingContent = ""
		}
		m.updateChatContent()
		m.chatViewport.GotoBottom()
		return m, nil

	case sidecarStartedMsg:
		m.statusMessage = "AI Online"
		return m, nil

	case editorFinishedMsg:
		if msg.err == nil {
			content, err := os.ReadFile(msg.path)
			if err == nil {
				m.inputArea.SetValue(string(content))
			}
		}
		os.Remove(msg.path)
		return m, nil

	case StreamErrorMsg:
		m.isLoading = false
		m.statusMessage = "Error"
		// Add error message to chat
		m.messages = append(m.messages, Message{
			Role:    "assistant",
			Content: fmt.Sprintf("❌ Error: %v", msg.Error),
		})
		m.updateChatContent()
		return m, nil

	case streamStartedMsg:
		m.textChan = msg.textChan
		m.errChan = msg.errChan
		m.statusMessage = "Streaming..."
		return m, m.receiveStreamChunk()
	}


	// Update focused component
	var cmd tea.Cmd
	switch m.focusedPane {
	case PaneChat:
		m.chatViewport, cmd = m.chatViewport.Update(msg)
		cmds = append(cmds, cmd)
	case PaneContext:
		m.contextViewport, cmd = m.contextViewport.Update(msg)
		cmds = append(cmds, cmd)
	case PaneInput:
		m.inputArea, cmd = m.inputArea.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View implements tea.Model
func (m Model) View() string {
	if m.quitting {
		return "Goodbye! 👋\n"
	}

	if m.width == 0 {
		return "Loading..."
	}

	// Render modals on top of everything
	if m.confirmModal.IsVisible() {
		return m.confirmModal.View()
	}

	if m.filePicker.IsVisible() {
		return m.filePicker.View()
	}

	if m.showHelp {
		return m.renderHelp()
	}

	return m.renderLayout()
}


// cycleFocus moves focus to the next pane
func (m *Model) cycleFocus() {
	switch m.focusedPane {
	case PaneInput:
		m.focusedPane = PaneChat
		m.inputArea.Blur()
	case PaneChat:
		m.focusedPane = PaneContext
	case PaneContext:
		m.focusedPane = PaneInput
		m.inputArea.Focus()
	}
}

// updateLayout recalculates pane dimensions based on window size
func (m *Model) updateLayout() {
	// Reserve space for: status bar (1) + pane titles (1) + input area (3) + help bar (1) + borders (2)
	contentHeight := m.height - 8
	if contentHeight < 5 {
		contentHeight = 5
	}
	inputHeight := 1

	// Split width: 30% context, 70% chat
	contextWidth := m.width * 30 / 100
	chatWidth := m.width - contextWidth - 2 // Account for gap between panes

	m.contextViewport.Width = contextWidth - 4
	m.contextViewport.Height = contentHeight

	m.chatViewport.Width = chatWidth - 4
	m.chatViewport.Height = contentHeight

	m.inputArea.SetWidth(m.width - 4)
	m.inputArea.SetHeight(inputHeight)

	// Restore welcome message if chat is empty
	if len(m.messages) == 0 && m.streamingContent == "" {
		welcome := Banner(m.chatViewport.Width) + 
			"\nType a message and press Enter to chat with AI.\n\nTry:\n" +
			m.styles.Muted.Render("• \"What files are in this project?\"\n") +
			m.styles.Muted.Render("• \"Explain the coder package\"\n") +
			m.styles.Muted.Render("• \"Help me add a new module\"")
			
		m.chatViewport.SetContent(welcome)
	}
}

// addFileToContext adds a file to the context
func (m *Model) addFileToContext(filePath string) {
	// Check if already added
	for _, f := range m.contextFiles {
		if f == filePath {
			return
		}
	}
	m.contextFiles = append(m.contextFiles, filePath)
	m.updateContextPane()
	m.statusMessage = "Added: " + filePath
	
	// Dynamic prompt
	lastFile := filePath
	if idx := strings.LastIndex(filePath, "/"); idx != -1 {
		lastFile = filePath[idx+1:]
	}
	m.inputArea.Placeholder = fmt.Sprintf("Ask about %s...", lastFile)
}

// updateContextPane refreshes the context pane with current files
func (m *Model) updateContextPane() {
	var sb strings.Builder
	
	if len(m.contextFiles) == 0 {
		sb.WriteString("Your Grimoire is empty.\n")
		sb.WriteString("Press Ctrl+O to add scrolls.")
	} else {
		sb.WriteString(fmt.Sprintf("%d scroll(s) in Grimoire:\n\n", len(m.contextFiles)))
		for _, f := range m.contextFiles {
			sb.WriteString(fmt.Sprintf("📜 %s\n", f))
		}
		sb.WriteString("\nCtrl+O to add more")
	}
	
	m.contextViewport.SetContent(sb.String())
}


// updateChatContent refreshes the chat viewport with all messages
func (m *Model) updateChatContent() {
	var sb strings.Builder

	for _, msg := range m.messages {
		if msg.Role == "user" {
			sb.WriteString(m.styles.Title.Render("You") + "\n")
			sb.WriteString(m.styles.UserMessage.Render(msg.Content) + "\n\n")
		} else {
			sb.WriteString(m.styles.Subtitle.Render("🐙 Kthulu") + "\n")
			sb.WriteString(m.styles.AIMessage.Render(msg.Content) + "\n\n")
		}
	}

	m.chatViewport.SetContent(sb.String())
}

// renderLayout renders the main split-pane layout
func (m Model) renderLayout() string {
	// Status bar
	statusBar := m.renderStatusBar()

	// Context pane (left)
	contextTitleStyle := m.styles.PaneTitle
	contextStyle := m.styles.ContextPane
	if m.focusedPane == PaneContext {
		contextTitleStyle = m.styles.PaneTitleActive
		contextStyle = contextStyle.BorderForeground(DefaultTheme.BorderFocus)
	} else {
		contextStyle = contextStyle.BorderForeground(DefaultTheme.Border)
	}
	contextTitle := contextTitleStyle.Render("📖 Grimoire")
	
	contextPane := lipgloss.JoinVertical(
		lipgloss.Left,
		contextTitle,
		contextStyle.
			Width(m.width*30/100-4).
			Height(m.height-8). // Adjusted for height calculation
			Render(m.contextViewport.View()),
	)

	// Chat pane (right)
	chatTitleStyle := m.styles.PaneTitle
	chatStyle := m.styles.ChatPane
	if m.focusedPane == PaneChat {
		chatTitleStyle = m.styles.PaneTitleActive
		chatStyle = chatStyle.BorderForeground(DefaultTheme.BorderFocus)
	} else {
		chatStyle = chatStyle.BorderForeground(DefaultTheme.Border)
	}
	chatTitle := chatTitleStyle.Render("💬 Chat")
	
	chatPane := lipgloss.JoinVertical(
		lipgloss.Left,
		chatTitle,
		chatStyle.
			Width(m.width-(m.width*30/100)-2). // Fill remaining width
			Height(m.height-8).
			Render(m.chatViewport.View()),
	)

	// Main content (horizontal split)
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, contextPane, chatPane)

	// Input area
	inputStyle := m.styles.InputPane
	if m.focusedPane == PaneInput {
		inputStyle = inputStyle.BorderForeground(DefaultTheme.BorderFocus)
	} else {
		inputStyle = inputStyle.BorderForeground(DefaultTheme.Border)
	}
	inputPane := inputStyle.Width(m.width - 2).Render(m.inputArea.View())

	// Help Bar
	helpBar := m.styles.Muted.
		Width(m.width).
		Align(lipgloss.Center).
		Render("[Tab] Next Pane • [Ctrl+O] Grimoire • [Ctrl+E] Editor • [Esc] Quit")

	// Combine all
	return lipgloss.JoinVertical(
		lipgloss.Left,
		statusBar,
		mainContent,
		inputPane,
		helpBar,
	)
}


// renderStatusBar renders the status bar at the top
func (m Model) renderStatusBar() string {
	left := fmt.Sprintf(" 🐙 Kthulu Coder │ %s ", m.modelName)
	
	// Show status in the middle
	status := m.statusMessage
	if m.isLoading {
		status = m.spinner.View() + " " + m.statusMessage
	}
	
	right := fmt.Sprintf(" %s │ %s ", status, m.workingDir)

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}

	return m.styles.StatusBar.
		Width(m.width).
		Render(left + strings.Repeat(" ", gap) + right)
}


// renderHelp renders the help screen
func (m Model) renderHelp() string {
	title := m.styles.Title.Render("⌨️  Keyboard Shortcuts")
	helpContent := m.help.View(m.keys)
	footer := m.styles.Muted.Render("\nPress ? or Esc to close")

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center, title, "", helpContent, footer),
	)
}

// updateChatContentStreaming updates chat with current streaming content
func (m *Model) updateChatContentStreaming() {
	var sb strings.Builder

	for _, msg := range m.messages {
		if msg.Role == "user" {
			sb.WriteString(m.styles.Title.Render("You") + "\n")
			sb.WriteString(m.styles.UserMessage.Render(msg.Content) + "\n\n")
		} else {
			sb.WriteString(m.styles.Subtitle.Render("🐙 Kthulu") + "\n")
			sb.WriteString(m.styles.AIMessage.Render(msg.Content) + "\n\n")
		}
	}

	// Add streaming content if present
	if m.streamingContent != "" {
		sb.WriteString(m.styles.Subtitle.Render("🐙 Kthulu") + "\n")
		sb.WriteString(m.styles.AIMessage.Render(m.streamingContent))
		if m.isLoading {
			sb.WriteString(m.spinner.View())
		}
		sb.WriteString("\n\n")
	}

	m.chatViewport.SetContent(sb.String())
}

// sendToLLM initiates a chat request to LiteLLM
func (m *Model) sendToLLM() tea.Cmd {
	return func() tea.Msg {
		// Convert messages to ChatMessage format
		chatMessages := make([]ChatMessage, len(m.messages))
		for i, msg := range m.messages {
			chatMessages[i] = ChatMessage{
				Role:    msg.Role,
				Content: msg.Content,
			}
		}

		// Add system prompt
		systemPrompt := ChatMessage{
			Role: "system",
			Content: `You are Kthulu Coder, an AI coding assistant specialized in the Kthulu framework.

Kthulu uses:
- Hexagonal Architecture (Ports & Adapters)
- Modular Monolith with Vertical Slices
- Uber fx for dependency injection
- CLI-first approach

Be concise, helpful, and follow best practices for Go development.`,
		}
		allMessages := append([]ChatMessage{systemPrompt}, chatMessages...)

		// Start streaming
		textChan, errChan := m.llmClient.StreamChat(m.ctx, allMessages)
		
		// Store channels for receiving
		return streamStartedMsg{textChan: textChan, errChan: errChan}
	}
}

// streamStartedMsg is sent when streaming begins
type streamStartedMsg struct {
	textChan <-chan string
	errChan  <-chan error
}

// receiveStreamChunk receives the next chunk from the stream
func (m *Model) receiveStreamChunk() tea.Cmd {
	return func() tea.Msg {
		if m.textChan == nil {
			return StreamDoneMsg{}
		}

		select {
		case text, ok := <-m.textChan:
			if !ok {
				return StreamDoneMsg{}
			}
			return StreamChunkMsg{Content: text}
		case err := <-m.errChan:
			if err != nil {
				return StreamErrorMsg{Error: err}
			}
			return StreamDoneMsg{}
		case <-m.ctx.Done():
			return StreamDoneMsg{}
		}
	}
}

// GetMessages returns the chat history
func (m Model) GetMessages() []Message {
	return m.messages
}
