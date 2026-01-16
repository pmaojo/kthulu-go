package coder

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pmaojo/kthulu-go/internal/coder/tools"
	"github.com/pmaojo/kthulu-go/internal/coder/skills"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/glamour"
)

// Pane represents which pane is currently focused
type Pane int

const (
	PaneChat Pane = iota
	PaneContext
	PaneInput
)

type Model struct {
	// State
	ctx              context.Context
	cancel           context.CancelFunc
	workingDir       string
	modelName        string
	
	// UI State
	width            int
	height           int
	focusedPane      Pane
	inputArea        textarea.Model
	chatViewport     viewport.Model
	contextViewport  viewport.Model
	logViewport      viewport.Model
	toolsViewport    viewport.Model
	spinner          spinner.Model
	help             help.Model
	keys             KeyMap
	confirmModal     ConfirmationModal
	filePicker       FilePicker
	showHelp         bool
	quitting         bool
	
	// Chat State
	messages         []Message
	isLoading        bool
	streamingContent string
	statusMessage    string
	contextFiles     []string
	rulesContent     string
	logs             []string
	activeTools      []string
	pulseInfo        string

	// LLM
	llmClient        *AIClient
	eventChan        <-chan ChatEvent
	toolBuffer       map[int]*ToolCall
	
	// Tools & MCP
	toolRegistry     *tools.Registry
	mcpManager       *MCPManager
	skillManager     *skills.Manager
	
	// Styling
	styles           Styles
}

// New creates a new model
func New(workingDir, modelName string) Model {
	ctx, cancel := context.WithCancel(context.Background())

	s := DefaultStyles()

	ta := textarea.New()
	ta.Placeholder = "Ask Kthulu..."
	ta.Focus()
	ta.Prompt = "┃ "
	ta.CharLimit = 0
	ta.SetHeight(1)
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	chatVp := viewport.New(0, 0)
	contextVp := viewport.New(0, 0)
	logVp := viewport.New(0, 0)
	logVp.Style = s.Muted
	toolsVp := viewport.New(0, 0)
	toolsVp.Style = s.Muted

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	// Create tool registry, MCP manager and skills manager
	registry := tools.NewRegistry()
	mcpManager := NewMCPManager()
	skillManager := skills.NewManager()

	// Initialize LLM Client
	client := NewAIClient(DefaultAIConfig(modelName))

	m := Model{
		ctx:              ctx,
		cancel:           cancel,
		workingDir:       workingDir,
		modelName:        modelName,
		styles:           s,
		keys:             DefaultKeyMap(),
		help:             help.New(),
		
		chatViewport:     chatVp,
		contextViewport:  contextVp,
		logViewport:      logVp,
		toolsViewport:    toolsVp,
		inputArea:        ta,
		spinner:          sp,
		
		messages:         []Message{},
		focusedPane:      PaneInput,
		toolBuffer:       make(map[int]*ToolCall),

		toolRegistry:     registry,
		mcpManager:       mcpManager,
		skillManager:     skillManager,
		llmClient:        client,
	}
	m.updatePulseInfo()
	m.updateToolsPane()
	return m
}


// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		m.spinner.Tick,
		tea.EnableMouseCellMotion,
		m.initMCPServersCmd(),
		m.loadSkillsCmd(),
		m.loadRulesCmd(),
	)
}



type mcpServersStartedMsg struct {
	connected []string
}

func (m Model) initMCPServersCmd() tea.Cmd {
	return func() tea.Msg {
		connected := InitializeMCPServers(context.Background(), m.mcpManager)
		return mcpServersStartedMsg{connected: connected}
	}
}

type skillsLoadedMsg struct {
	count int
}

func (m Model) loadSkillsCmd() tea.Cmd {
	return func() tea.Msg {
		// Look in typical locations
		// 1. .kthulu/skills relative to working dir
		// 2. ~/.kthulu/skills (global) - Optional, maybe later
		skillPath := filepath.Join(m.workingDir, ".kthulu", "skills")
		err := m.skillManager.LoadSkills(skillPath)
		if err != nil {
			// Just return 0 count, ignore error for UI noise? or log it?
			return skillsLoadedMsg{count: 0}
		}
		return skillsLoadedMsg{count: len(m.skillManager.All())}
	}
}

type rulesLoadedMsg struct {
	content string
}

func (m Model) loadRulesCmd() tea.Cmd {
	return func() tea.Msg {
		// Priority 1: KTHULU.md in root
		path := filepath.Join(m.workingDir, "KTHULU.md")
		if _, err := os.Stat(path); err != nil {
			// Priority 2: .kthulu/rules.md
			path = filepath.Join(m.workingDir, ".kthulu", "rules.md")
			if _, err := os.Stat(path); err != nil {
				return rulesLoadedMsg{content: ""}
			}
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return rulesLoadedMsg{content: ""}
		}
		return rulesLoadedMsg{content: string(content)}
	}
}

func (m *Model) handleSlashCommand(input string) tea.Cmd {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}
	cmd := parts[0]
	args := parts[1:]

	switch cmd {
	case "/help":
		m.messages = append(m.messages, Message{
			Role: "system",
			Content: "Available commands:\n" +
				"• /help - Show this help\n" +
				"• /clear - Clear chat history\n" +
				"• /add <file> - Add file to context\n" +
				"• /mode <name> - Switch persona (expert, learner, etc.)",
		})
		m.updateChatContent()
		
	case "/clear":
		m.messages = []Message{}
		m.chatViewport.SetContent(Banner(m.chatViewport.Width))
		m.statusMessage = "Cleared history"

	case "/add":
		if len(args) == 0 {
			m.statusMessage = "Usage: /add <file>"
		} else {
			f := args[0]
			// Handle absolute or relative paths
			path := f
			if !filepath.IsAbs(f) {
				path = filepath.Join(m.workingDir, f)
			}
			m.addFileToContext(path)
		}

	case "/mode":
		if len(args) == 0 {
			m.statusMessage = "Mode: Default"
			// Todo show current mode
		} else {
			m.statusMessage = "Mode switched to " + args[0]
			// Logic to switch mode...
		}

	default:
		m.statusMessage = "Unknown command: " + cmd
	}
	return nil
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle modals first
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
		return m.handleKeyMsg(msg)

	case tea.MouseMsg:
		return m.handleMouseMsg(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()
		m.confirmModal.SetSize(m.width, m.height)
		m.filePicker.SetSize(m.width, m.height)
		return m, nil

	case logMsg:
		m.log(string(msg))
		return m, nil

	case spinner.TickMsg:
		if m.isLoading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case StreamStartedMsg, StreamChunkMsg, StreamErrorMsg, StreamToolCallMsg, StreamFinishMsg, StreamDoneMsg:
		return m.handleStreamMsg(msg)

	case ToolExecutionResultMsg, ToolStartedMsg, ToolFinishedMsg, toolsRegisteredMsg:
		return m.handleToolMsg(msg)
	
	case mcpServersStartedMsg:
		if len(msg.connected) > 0 {
			m.statusMessage = fmt.Sprintf("MCP: %d connected", len(msg.connected))
		}
		return m, func() tea.Msg { return toolsRegisteredMsg{count: 0} }

	case skillsLoadedMsg:
		if msg.count > 0 {
			m.statusMessage = fmt.Sprintf("Skills: %d loaded", msg.count)
		}
		return m, nil

	case rulesLoadedMsg:
		if msg.content != "" {
			m.rulesContent = msg.content
			m.statusMessage = "Project Rules Loaded"
		}
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

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.String() == "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case msg.String() == "ctrl+k":
		if m.focusedPane != PaneInput {
			m.focusedPane = PaneInput
			m.inputArea.Focus()
		}
		m.inputArea.SetValue("/")
		return m, nil

	case msg.String() == "ctrl+o":
		m.filePicker.SetSize(m.width, m.height)
		return m, m.filePicker.Show()

	case msg.String() == "tab":
		m.cycleFocus()
		return m, nil

	case msg.String() == "?":
		if m.focusedPane != PaneInput {
			m.showHelp = !m.showHelp
		}
		return m, nil

	case msg.String() == "ctrl+e":
		return m, openEditor(m.inputArea.Value())

	case msg.String() == "ctrl+l":
		m.messages = []Message{}
		m.chatViewport.SetContent(Banner(m.chatViewport.Width))
		return m, nil

	case msg.String() == "enter":
		if m.focusedPane == PaneInput && strings.TrimSpace(m.inputArea.Value()) != "" && !m.isLoading {
			content := m.inputArea.Value()
			m.inputArea.Reset()

			if strings.HasPrefix(strings.TrimSpace(content), "/") {
				return m, m.handleSlashCommand(strings.TrimSpace(content))
			}

			m.messages = append(m.messages, Message{Role: "user", Content: content})
			m.updateChatContent()
			m.isLoading = true
			m.streamingContent = ""
			m.statusMessage = "Thinking..."
			return m, m.sendToLLM()
		}

	case msg.String() == "esc":
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}
	}
	// Propagate key msg to components if not handled
	return m.updateComponents(msg)
}

func (m Model) handleMouseMsg(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
		return m, nil
	}
	
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
}

func (m Model) handleStreamMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case StreamStartedMsg:
		m.statusMessage = "Thinking..."
		m.eventChan = msg.EventChan
		m.messages = append(m.messages, Message{
			Role: "assistant",
		})
		m.toolBuffer = make(map[int]*ToolCall)
		return m, m.waitForStream()

	case StreamChunkMsg:
		if len(m.messages) > 0 {
			m.messages[len(m.messages)-1].Content += msg.Content
			m.updateChatContentStreaming()
		}
		return m, m.waitForStream()

	case StreamErrorMsg:
		m.isLoading = false
		m.statusMessage = "Error"
		m.messages = append(m.messages, Message{
			Role:    "assistant",
			Content: fmt.Sprintf("❌ Error: %v", msg.Error),
		})
		m.updateChatContent()
		return m, nil

	case StreamToolCallMsg:
		return m.handleStreamToolCall(msg)

	case StreamFinishMsg:
		return m.handleStreamFinish(msg)

	case StreamDoneMsg:
		m.statusMessage = "Ready"
		return m, nil
	}
	return m, nil
}

func (m Model) handleStreamToolCall(msg StreamToolCallMsg) (tea.Model, tea.Cmd) {
	m.statusMessage = "Receiving tool calls..."
	if m.toolBuffer == nil {
		m.toolBuffer = make(map[int]*ToolCall)
	}

	for _, tc := range msg.ToolCalls {
		if tc.Index == nil {
			continue
		}
		idx := *tc.Index
		if existing, ok := m.toolBuffer[idx]; ok {
			existing.Function.Arguments += tc.Function.Arguments
		} else {
			newTc := tc
			m.toolBuffer[idx] = &newTc
		}
	}
	return m, m.waitForStream()
}

func (m Model) handleStreamFinish(msg StreamFinishMsg) (tea.Model, tea.Cmd) {
	if len(m.toolBuffer) > 0 {
		m.statusMessage = "Executing tools..."
		var calls []ToolCall
		for i := 0; i < len(m.toolBuffer); i++ {
			if tc, ok := m.toolBuffer[i]; ok {
				calls = append(calls, *tc)
			}
		}
		if len(m.messages) > 0 {
			m.messages[len(m.messages)-1].ToolCalls = calls
		}
		return m, m.executeToolsCmd(calls)
	}
	return m, m.waitForStream()
}

func (m Model) handleToolMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ToolExecutionResultMsg:
		m.isLoading = false
		m.statusMessage = "Ready"

		for _, res := range msg.Results {
			m.messages = append(m.messages, Message{
				Role:       "tool",
				Content:    res.Output,
				ToolCallID: res.ToolCallID,
			})
			if res.Error != nil {
				m.log(fmt.Sprintf("❌ Tool %s error: %v", res.Name, res.Error))
			} else {
				m.log(fmt.Sprintf("✅ Tool %s finished.", res.Name))
			}
		}
		m.updateChatContent()
		return m, m.sendToLLM()

	case toolsRegisteredMsg:
		m.statusMessage = fmt.Sprintf("Tools: %d ready", msg.count)
		return m, nil

	case ToolStartedMsg:
		m.activeTools = append(m.activeTools, string(msg))
		m.updateToolsPane()
		return m, nil

	case ToolFinishedMsg:
		for i, t := range m.activeTools {
			if t == string(msg) {
				m.activeTools = append(m.activeTools[:i], m.activeTools[i+1:]...)
				break
			}
		}
		m.updateToolsPane()
		return m, nil
	}
	return m, nil
}

func (m Model) updateComponents(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
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

type toolsRegisteredMsg struct {
	count int
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
	headerHeight := 7
	logHeight := 5
	inputAreaHeight := 4 // border + 1 line
	
	// Reserve space for: header (7) + status bar (1) + pane titles (1) + log title (1) + log viewport (5) + input area (4) + help bar (1)
	reservedHeight := headerHeight + 1 + 1 + 1 + logHeight + inputAreaHeight + 1
	
	contentHeight := m.height - reservedHeight
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

	// Split bottom row: Logs (60%) and Tools (40%)
	logWidth := (m.width * 60 / 100) - 2
	toolsWidth := m.width - logWidth - 2

	m.logViewport.Width = logWidth - 2
	m.logViewport.Height = logHeight

	m.toolsViewport.Width = toolsWidth - 2
	m.toolsViewport.Height = logHeight

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
	
	// Calculate available width for messages (viewport width minus padding/borders)
	maxWidth := m.chatViewport.Width - 6 // Extra padding for borders
	if maxWidth < 20 {
		maxWidth = 20
	}

	// Re-initialize glamour with correct width
	// We ignore error for simplicity in TUI
	m.styles.Markdown, _ = glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(maxWidth),
	)

	for _, msg := range m.messages {
		if msg.Role == "user" {
			sb.WriteString(m.styles.Title.Render("You") + "\n")
			// User messages are usually short, so keep simple rendering or use glamour too?
			// Let's use glamour for consistency if user pastes code
			rendered, err := m.styles.Markdown.Render(msg.Content)
			if err != nil {
				rendered = msg.Content
			}
			sb.WriteString(m.styles.UserMessage.Render(rendered) + "\n\n")
		} else {
			sb.WriteString(m.styles.Subtitle.Render("🐙 Kthulu") + "\n")
			rendered, err := m.styles.Markdown.Render(msg.Content)
			if err != nil {
				rendered = msg.Content
			}
			sb.WriteString(m.styles.AIMessage.Render(rendered) + "\n\n")
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

	// Bottom panes (Logs and Tools)
	logTitle := m.styles.PaneTitle.Render("📜 Activity Logs")
	toolsTitle := m.styles.PaneTitle.Render("🛠️ Active Tools")

	logPane := m.styles.Pane.Width(m.logViewport.Width + 2).Render(m.logViewport.View())
	toolsPane := m.styles.Pane.Width(m.toolsViewport.Width + 2).Render(m.toolsViewport.View())
	
	bottomPanes := lipgloss.JoinHorizontal(lipgloss.Top, 
		lipgloss.JoinVertical(lipgloss.Left, logTitle, logPane),
		lipgloss.JoinVertical(lipgloss.Left, toolsTitle, toolsPane),
	)

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
		m.renderHeader(),
		statusBar,
		mainContent,
		bottomPanes,
		inputPane,
		helpBar,
	)
}

func (m Model) renderHeader() string {
	logo := `  _  _________ _    _ _    _ _      _    _ 
 | |/ /__   __| |  | | |  | | |    | |  | |
 | ' /   | |  | |__| | |  | | |    | |  | |
 |  <    | |  |  __  | |  | | |    | |  | |
 | . \   | |  | |  | | |__| | |____| |__| |
 |_|\_\  |_|  |_|  |_|\____/|______\____/ `

	return m.styles.Title.
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Width(m.width).
		Align(lipgloss.Center).
		PaddingTop(1).
		Render(logo)
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
	
	// Calculate available width for messages
	maxWidth := m.chatViewport.Width - 6
	if maxWidth < 20 {
		maxWidth = 20
	}

	// Refresh renderer for width
	m.styles.Markdown, _ = glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(maxWidth),
	)

	for _, msg := range m.messages {
		if msg.Role == "user" {
			sb.WriteString(m.styles.Title.Render("You") + "\n")
			rendered, err := m.styles.Markdown.Render(msg.Content)
			if err != nil {
				rendered = msg.Content
			}
			sb.WriteString(m.styles.UserMessage.Render(rendered) + "\n\n")
		} else {
			sb.WriteString(m.styles.Subtitle.Render("🐙 Kthulu") + "\n")
			rendered, err := m.styles.Markdown.Render(msg.Content)
			if err != nil {
				rendered = msg.Content
			}
			sb.WriteString(m.styles.AIMessage.Render(rendered) + "\n\n")
		}
	}

	// Add streaming content if present
	if m.streamingContent != "" {
		sb.WriteString(m.styles.Subtitle.Render("🐙 Kthulu") + "\n")
		// For streaming, we might be rendering incomplete markdown (e.g. half a code block).
		// Glamour handles this surprisingly well usually, but might flash.
		rendered, err := m.styles.Markdown.Render(m.streamingContent)
		if err != nil {
			rendered = m.streamingContent
		}
		sb.WriteString(m.styles.AIMessage.Render(rendered))
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
			// Strip Index from ToolCalls for history compatibility (Fix 422)
			var historyToolCalls []ToolCall
			if len(msg.ToolCalls) > 0 {
				historyToolCalls = make([]ToolCall, len(msg.ToolCalls))
				for j, tc := range msg.ToolCalls {
					historyToolCalls[j] = tc
					historyToolCalls[j].Index = nil // Omit from JSON
				}
			}

			chatMessages[i] = ChatMessage{
				Role:       msg.Role,
				Content:    msg.Content,
				ToolCalls:  historyToolCalls,
				ToolCallID: msg.ToolCallID,
			}
		}

		// Add skills info
		skillsInfo := ""
		if m.skillManager != nil {
			loadedSkills := m.skillManager.All()
			if len(loadedSkills) > 0 {
				skillsInfo = "\n\n## Available Skills\n"
				for _, s := range loadedSkills {
					skillsInfo += fmt.Sprintf("- %s: %s (Path: %s)\n", s.Name, s.Description, s.Path)
				}
				skillsInfo += "\nTo use a skill, read its definition file using view_file or read_file to understand its capabilities and instructions."
			}
		}

		// Add rules info
		rulesInfo := ""
		if m.rulesContent != "" {
			rulesInfo = "\n\n## Project Rules & Guidelines\n" + m.rulesContent
		}

		// Add system prompt
		systemPrompt := ChatMessage{
			Role: "system",
			Content: `You are Kthulu Coder, an AI coding assistant with access to powerful tools.

## CRITICAL: You MUST use tools to accomplish tasks. Do NOT just explain or provide tutorials.

## Available Tools
You have the following tools available. USE THEM to help the user:
- **bash**: Execute shell commands (go build, npm run, git, ls, cat, etc.)
- **read_file**: Read file contents from the filesystem
- **write_file**: Write or create files on the filesystem
- **grep**: Search for patterns in files
- **think**: Reason through complex problems step-by-step
- **kthulu**: Run kthulu CLI commands

## When to Use Tools
- User asks to "build" something → Use tools to create files, run builds
- User asks about files → Use read_file to examine them
- User asks to run commands → Use bash
- User asks to modify code → Use write_file
- User asks about project structure → Use kthulu status or kthulu analyze
- User asks to add a module/component → Use kthulu add module/component
- User asks for code review/optimization → Use kthulu ai review/optimize

## Kthulu Framework Context
The Kthulu framework uses:
- Hexagonal Architecture (Ports & Adapters)
- Modular Monolith with Vertical Slices
- Uber fx for dependency injection
- CLI-first approach

Be action-oriented. When asked to do something, DO IT using your tools.` + rulesInfo + skillsInfo,
		}
		allMessages := append([]ChatMessage{systemPrompt}, chatMessages...)

		// Prepare tools
		var toolDefs []Tool
		if m.toolRegistry != nil {
			for _, t := range m.toolRegistry.ToOpenAIFormat() {
				// Manual conversion from map to Tool struct
				if fn, ok := t["function"].(map[string]interface{}); ok {
					toolDefs = append(toolDefs, Tool{
						Type: "function",
						Function: FunctionDef{
							Name:        fn["name"].(string),
							Description: fn["description"].(string),
							Parameters:  fn["parameters"],
						},
					})
				}
			}
		}

		// Start streaming
		eventChan := m.llmClient.StreamChat(m.ctx, allMessages, toolDefs)

		return StreamStartedMsg{EventChan: eventChan}
	}
}

// StreamFinishMsg is sent when the stream finishes with a reason
type StreamFinishMsg struct {
	Reason string
}

// StreamToolCallMsg is sent when tool calls are received
type StreamToolCallMsg struct {
	ToolCalls []ToolCall
}

// waitForStream receives the next event from the stream
func (m *Model) waitForStream() tea.Cmd {
	return func() tea.Msg {
		if m.eventChan == nil {
			return StreamDoneMsg{}
		}

		select {
		case event, ok := <-m.eventChan:
			if !ok {
				return StreamDoneMsg{}
			}

			switch event.Type {
			case "content":
				return StreamChunkMsg{Content: event.Content}
			case "tool_calls":
				return StreamToolCallMsg{ToolCalls: event.ToolCalls}
			case "finish":
				return StreamFinishMsg{Reason: event.FinishReason}
			case "error":
				return StreamErrorMsg{Error: event.Error}
			case "done":
				return StreamDoneMsg{}
			default:
				// Ignore unknown types
				return m.waitForStream()()
			}

		case <-m.ctx.Done():
			return StreamDoneMsg{}
		}
	}
}

// GetMessages returns the chat history
func (m Model) GetMessages() []Message {
	return m.messages
}

// ToolExecutionResultMsg contains the results of tool executions
type ToolExecutionResultMsg struct {
	Results []ToolResult
}

type ToolResult struct {
	ToolCallID string
	Name       string
	Output     string
	Error      error
}

// toolStartedMsg and toolFinishedMsg track active tool execution
type toolStartedMsg string
type toolFinishedMsg string

func (m *Model) executeToolsCmd(calls []ToolCall) tea.Cmd {
	var cmds []tea.Cmd
	for _, call := range calls {
		cmds = append(cmds, m.executeTool(call))
	}
	return tea.Batch(cmds...)
}

func (m *Model) executeTool(call ToolCall) tea.Cmd {
	return func() tea.Msg {
		// Emit start
		// Tracking start/end from within a Cmd is hard without a channel.
		// For now we'll just return a result that includes the name.
		
		tool, found := m.toolRegistry.Get(call.Function.Name)
		if !found {
			return ToolResult{
				ToolCallID: call.ID,
				Name:       call.Function.Name,
				Error:      fmt.Errorf("tool not found: %s", call.Function.Name),
			}
		}

		// Parse arguments
		var args map[string]interface{}
		if err := json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
			return ToolResult{
				ToolCallID: call.ID,
				Name:       call.Function.Name,
				Error:      fmt.Errorf("invalid arguments: %v", err),
			}
		}

		// Execute tool
		res, err := tool.Execute(context.Background(), args)
		
		if err != nil {
			return ToolResult{
				ToolCallID: call.ID,
				Name:       call.Function.Name,
				Error:      err,
			}
		}

		return ToolResult{
			ToolCallID: call.ID,
			Name:       call.Function.Name,
			Output:     res.Output,
		}
	}
}

// logMsg is a message to append to the activity log
type logMsg string

func (m *Model) log(msg string) {
	m.logs = append(m.logs, msg)
	if len(m.logs) > 100 {
		m.logs = m.logs[len(m.logs)-100:]
	}
	m.logViewport.SetContent(strings.Join(m.logs, "\n"))
	m.logViewport.GotoBottom()
}

// updateToolsPane refreshes the active tools viewport
func (m *Model) updateToolsPane() {
	if len(m.activeTools) == 0 {
		m.toolsViewport.SetContent(m.styles.Muted.Render("No active tools."))
		return
	}
	
	var sb strings.Builder
	for _, t := range m.activeTools {
		sb.WriteString(fmt.Sprintf("⚡ %s\n", t))
	}
	m.toolsViewport.SetContent(sb.String())
}

// updatePulseInfo gathers project stats for the header/status segment
func (m *Model) updatePulseInfo() {
	// Simple pulse: Current branch + file count
	branch := "main"
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err == nil {
		branch = strings.TrimSpace(string(out))
	}
	
	// Count files (approximate)
	files := 0
	_ = filepath.Walk(m.workingDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && !strings.Contains(path, "/.git/") && !strings.Contains(path, "/node_modules/") {
			files++
		}
		return nil
	})
	
	m.pulseInfo = fmt.Sprintf("🌳 %s  │  📂 %d files", branch, files)
}

// ToolStartedMsg and ToolFinishedMsg are for TUI status updates
type ToolStartedMsg string
type ToolFinishedMsg string
