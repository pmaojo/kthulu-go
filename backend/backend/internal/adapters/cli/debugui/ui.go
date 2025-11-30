package debugui

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	highlightColor = lipgloss.Color("212")
	dimColor       = lipgloss.Color("240")
	activeTabStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(highlightColor).
			Padding(0, 1).
			Bold(true)
	inactiveTabStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(dimColor).
			Padding(0, 1)
	windowStyle = lipgloss.NewStyle().Padding(1)
)

// Log Types
type LogType int

const (
	LogTypeRaw LogType = iota
	LogTypeHTTP
	LogTypeDB
)

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Type      LogType   `json:"type"`
	Raw       string    `json:"raw"`

	// HTTP fields
	Method    string `json:"method,omitempty"`
	Path      string `json:"path,omitempty"`
	Status    string `json:"status,omitempty"`
	Duration  string `json:"duration,omitempty"`

	// DB fields
	SQL       string `json:"sql,omitempty"`
	Rows      string `json:"rows,omitempty"`
}

// Messages
type logMsg string
type processFinishedMsg error

// Model
type Model struct {
	subProcess *exec.Cmd
	persist    bool

	// State
	activeTab int // 0: Console, 1: HTTP, 2: Database
	logs      []LogEntry
	httpLogs  []LogEntry
	dbLogs    []LogEntry

	// UI Components
	ready        bool
	viewport     viewport.Model // For raw console
	httpTable    table.Model
	dbTable      table.Model

	// Channels for I/O
	stdout io.ReadCloser
	stderr io.ReadCloser
	logCh  chan string

	// Window size
	width, height int
}

func NewModel(cmd *exec.Cmd, persist bool) Model {
	// Initialize HTTP Table
	httpColumns := []table.Column{
		{Title: "Method", Width: 8},
		{Title: "Path", Width: 40},
		{Title: "Status", Width: 8},
		{Title: "Duration", Width: 12},
		{Title: "Time", Width: 10},
	}
	httpTable := table.New(
		table.WithColumns(httpColumns),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	httpTable.SetStyles(s)

	// Initialize DB Table
	dbColumns := []table.Column{
		{Title: "SQL", Width: 60},
		{Title: "Rows", Width: 8},
		{Title: "Duration", Width: 12},
		{Title: "Time", Width: 10},
	}
	dbTable := table.New(
		table.WithColumns(dbColumns),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	dbTable.SetStyles(s)

	return Model{
		subProcess: cmd,
		persist:    persist,
		logs:       make([]LogEntry, 0),
		httpLogs:   make([]LogEntry, 0),
		dbLogs:     make([]LogEntry, 0),
		httpTable:  httpTable,
		dbTable:    dbTable,
		logCh:      make(chan string, 1000), // Buffered channel
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.startSubprocess(),
		tea.EnterAltScreen,
		m.waitForLog(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.subProcess != nil && m.subProcess.Process != nil {
				_ = m.subProcess.Process.Kill()
			}
			return m, tea.Quit
		case "tab":
			m.activeTab = (m.activeTab + 1) % 3
		case "shift+tab":
			m.activeTab--
			if m.activeTab < 0 {
				m.activeTab = 2
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update Viewport
		if !m.ready {
			m.viewport = viewport.New(msg.Width-4, msg.Height-10) // Leave room for tabs
			m.viewport.SetContent("Starting application...\n")
			m.ready = true
		} else {
			m.viewport.Width = msg.Width - 4
			m.viewport.Height = msg.Height - 10
		}

		// Update Tables
		m.httpTable.SetWidth(msg.Width - 4)
		m.httpTable.SetHeight(msg.Height - 10)
		m.dbTable.SetWidth(msg.Width - 4)
		m.dbTable.SetHeight(msg.Height - 10)

	case logMsg:
		entry := parseLog(string(msg))
		m.logs = append(m.logs, entry)

		// Update Raw Viewport
		m.viewport.SetContent(m.viewport.View() + entry.Raw + "\n")
		m.viewport.GotoBottom()

		// Update Tables if needed
		if entry.Type == LogTypeHTTP {
			m.httpLogs = append(m.httpLogs, entry)
			m.updateHttpTable()
		} else if entry.Type == LogTypeDB {
			m.dbLogs = append(m.dbLogs, entry)
			m.updateDbTable()
		}

		// Persist if enabled
		if m.persist {
			m.appendToFile(entry)
		}

		return m, m.waitForLog() // Continue reading

	case processFinishedMsg:
		m.viewport.SetContent(m.viewport.View() + "\nProcess finished.\n")
		return m, nil
	}

	// Route updates to active component
	switch m.activeTab {
	case 0:
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	case 1:
		m.httpTable, cmd = m.httpTable.Update(msg)
		cmds = append(cmds, cmd)
	case 2:
		m.dbTable, cmd = m.dbTable.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// Tabs
	tabs := []string{"Console", "HTTP Requests", "Database"}
	var renderedTabs []string
	for i, t := range tabs {
		if i == m.activeTab {
			renderedTabs = append(renderedTabs, activeTabStyle.Render(t))
		} else {
			renderedTabs = append(renderedTabs, inactiveTabStyle.Render(t))
		}
	}
	tabRow := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)

	// Content
	var content string
	switch m.activeTab {
	case 0:
		content = m.viewport.View()
	case 1:
		content = m.httpTable.View()
	case 2:
		content = m.dbTable.View()
	}

	return windowStyle.Render(lipgloss.JoinVertical(lipgloss.Left, tabRow, content))
}

// Helper methods

func (m *Model) updateHttpTable() {
	rows := make([]table.Row, len(m.httpLogs))
	// Show latest first
	for i := range m.httpLogs {
		idx := len(m.httpLogs) - 1 - i
		log := m.httpLogs[idx]
		rows[i] = table.Row{log.Method, log.Path, log.Status, log.Duration, log.Timestamp.Format("15:04:05")}
	}
	m.httpTable.SetRows(rows)
}

func (m *Model) updateDbTable() {
	rows := make([]table.Row, len(m.dbLogs))
	for i := range m.dbLogs {
		idx := len(m.dbLogs) - 1 - i
		log := m.dbLogs[idx]
		// Truncate SQL
		sql := log.SQL
		if len(sql) > 55 {
			sql = sql[:55] + "..."
		}
		rows[i] = table.Row{sql, log.Rows, log.Duration, log.Timestamp.Format("15:04:05")}
	}
	m.dbTable.SetRows(rows)
}

func (m *Model) appendToFile(l LogEntry) {
	// In a real implementation, we should keep the file handle open or use a buffer.
	// For this CLI tool, opening/closing is acceptable but json.Marshal is safer.
	f, err := os.OpenFile("debug_events.jsonl", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	data, err := json.Marshal(l)
	if err == nil {
		f.Write(data)
		f.WriteString("\n")
	}
}

// Subprocess logic

func (m *Model) startSubprocess() tea.Cmd {
	return func() tea.Msg {
		stdout, err := m.subProcess.StdoutPipe()
		if err != nil {
			return processFinishedMsg(err)
		}
		stderr, err := m.subProcess.StderrPipe()
		if err != nil {
			return processFinishedMsg(err)
		}

		// Combine stdout and stderr
		m.stdout = stdout
		m.stderr = stderr

		if err := m.subProcess.Start(); err != nil {
			return processFinishedMsg(err)
		}

		// Start reading loops
		go m.readLoop(stdout)
		go m.readLoop(stderr)

		return nil
	}
}

func (m *Model) readLoop(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		// Non-blocking send if buffer is full, to avoid freezing subprocess
		select {
		case m.logCh <- scanner.Text():
		default:
			// If channel is full, we drop the log or handle overflow.
			// For a debug tool, dropping is better than deadlock.
		}
	}
}

func (m *Model) waitForLog() tea.Cmd {
	return func() tea.Msg {
		return logMsg(<-m.logCh)
	}
}

// Regex Parsers

// Gin: [GIN] 2023/10/26 - 12:00:00 | 200 | 1.2ms | ::1 | GET /api/v1/users
var ginRegex = regexp.MustCompile(`\|\s+(\d{3})\s+\|\s+([\d\.]+\w+)\s+\|.*\|\s+([A-Z]+)\s+([/\w-]+)`)

// GORM: 2023/10/26 12:00:00 /path/to/file.go:123 [1.2ms] [rows:1] SELECT * FROM users
// Note: GORM formats vary, this is a best-effort approximation for default logger
var gormRegex = regexp.MustCompile(`\[([\d\.]+\w+)\] \[rows:(\d+)\] (.*)`)

func parseLog(line string) LogEntry {
	entry := LogEntry{
		Timestamp: time.Now(),
		Raw:       line,
		Type:      LogTypeRaw,
	}

	// Check Gin/HTTP
	// We allow standard Gin logs. Customize regex if using Chi/custom middleware.
	if strings.Contains(line, "|") && (strings.Contains(line, "GET") || strings.Contains(line, "POST") || strings.Contains(line, "PUT") || strings.Contains(line, "DELETE")) {
		matches := ginRegex.FindStringSubmatch(line)
		if len(matches) == 5 {
			entry.Type = LogTypeHTTP
			entry.Status = matches[1]
			entry.Duration = matches[2]
			entry.Method = matches[3]
			entry.Path = matches[4]
			return entry
		}
	}

	// Check GORM/DB
	// Gorm logs usually contain [rows:X]
	if strings.Contains(line, "[rows:") {
		matches := gormRegex.FindStringSubmatch(line)
		if len(matches) == 4 {
			entry.Type = LogTypeDB
			entry.Duration = matches[1]
			entry.Rows = matches[2]
			entry.SQL = matches[3]
			return entry
		}
	}

	return entry
}
