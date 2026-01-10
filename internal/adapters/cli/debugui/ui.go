package debugui

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fsnotify/fsnotify"
)

// Styles
var (
	highlightColor = lipgloss.Color("212")
	dimColor       = lipgloss.Color("240")
	redColor       = lipgloss.Color("196")
	greenColor     = lipgloss.Color("46")
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

	passStyle = lipgloss.NewStyle().Foreground(greenColor).Bold(true)
	failStyle = lipgloss.NewStyle().Foreground(redColor).Bold(true)
)

// Log Types
type LogType int

const (
	LogTypeRaw LogType = iota
	LogTypeHTTP
	LogTypeDB
	LogTypeTest
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

	// Test fields
	TestName   string `json:"test_name,omitempty"`
	TestStatus string `json:"test_status,omitempty"` // PASS, FAIL, SKIP
	TestOutput string `json:"test_output,omitempty"`
}

// Messages
type logMsg string
type testMsg LogEntry
type processFinishedMsg error
type watchErrorMsg error

// Model
type Model struct {
	subProcess *exec.Cmd
	persist    bool
	testWatch  bool

	// State
	activeTab int // 0: Console, 1: HTTP, 2: Database, 3: Tests
	logs      []LogEntry
	httpLogs  []LogEntry
	dbLogs    []LogEntry
	testLogs  []LogEntry

	// UI Components
	ready        bool
	viewport     viewport.Model // For raw console
	httpTable    table.Model
	dbTable      table.Model
	testTable    table.Model
	testStatus   string // "Running...", "PASS", "FAIL"

	// Channels for I/O
	stdout io.ReadCloser
	stderr io.ReadCloser
	logCh  chan string
	testCh chan LogEntry

	// Window size
	width, height int

	// Watcher
	watcher *fsnotify.Watcher
}

func NewModel(cmd *exec.Cmd, persist bool, testWatch bool) Model {
	// Initialize HTTP Table
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

	// Initialize Test Table
	testColumns := []table.Column{
		{Title: "Status", Width: 8},
		{Title: "Test/Scenario", Width: 50},
		{Title: "Time", Width: 10},
		{Title: "Details", Width: 30},
	}
	testTable := table.New(
		table.WithColumns(testColumns),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	testTable.SetStyles(s)

	return Model{
		subProcess: cmd,
		persist:    persist,
		testWatch:  testWatch,
		logs:       make([]LogEntry, 0),
		httpLogs:   make([]LogEntry, 0),
		dbLogs:     make([]LogEntry, 0),
		testLogs:   make([]LogEntry, 0),
		httpTable:  httpTable,
		dbTable:    dbTable,
		testTable:  testTable,
		logCh:      make(chan string, 1000), // Buffered channel
		testCh:     make(chan LogEntry, 100),
		testStatus: "Waiting...",
	}
}

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		tea.EnterAltScreen,
		m.waitForLog(),
	}

	if m.subProcess != nil {
		cmds = append(cmds, m.startSubprocess())
	}

	if m.testWatch {
		cmds = append(cmds, m.startTestWatcher())
		cmds = append(cmds, m.waitForTest())
	}

	return tea.Batch(cmds...)
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
			if m.watcher != nil {
				m.watcher.Close()
			}
			return m, tea.Quit
		case "tab":
			numTabs := 3
			if m.testWatch {
				numTabs = 4
			}
			m.activeTab = (m.activeTab + 1) % numTabs
		case "shift+tab":
			numTabs := 3
			if m.testWatch {
				numTabs = 4
			}
			m.activeTab--
			if m.activeTab < 0 {
				m.activeTab = numTabs - 1
			}
		case "r":
			// Manual re-run tests
			if m.testWatch {
				m.testLogs = nil // Clear previous logs
				m.updateTestTable()
				m.testStatus = "Running..."
				go m.runTests()
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
		m.testTable.SetWidth(msg.Width - 4)
		m.testTable.SetHeight(msg.Height - 10)

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

	case testMsg:
		entry := LogEntry(msg)

		// Logic to update overall status
		if entry.TestStatus == "FAIL" {
			m.testStatus = "FAIL 🔴"
		} else if m.testStatus != "FAIL 🔴" && entry.TestStatus == "PASS" {
			// Only set to PASS if we haven't failed already in this run
			// Ideally we reset status at start of run
			m.testStatus = "PASS 🟢"
		}

		m.testLogs = append(m.testLogs, entry)
		m.updateTestTable()
		return m, m.waitForTest()

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
	case 3:
		m.testTable, cmd = m.testTable.Update(msg)
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
	if m.testWatch {
		tabs = append(tabs, fmt.Sprintf("🧪 Tests [%s]", m.testStatus))
	}

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
	case 3:
		content = m.testTable.View()
	}

	// Add footer or status bar if needed
	if m.testWatch && m.activeTab == 3 {
		content += "\n" + lipgloss.NewStyle().Foreground(dimColor).Render("Watching for changes... (Press 'r' to force run)")
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

func (m *Model) updateTestTable() {
	rows := make([]table.Row, len(m.testLogs))
	for i := range m.testLogs {
		idx := len(m.testLogs) - 1 - i
		log := m.testLogs[idx]
		rows[i] = table.Row{log.TestStatus, log.TestName, log.Timestamp.Format("15:04:05"), log.TestOutput}
	}
	m.testTable.SetRows(rows)
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
		if m.subProcess == nil {
			return nil
		}

		// If command is just a placeholder echo, don't try to pipe heavily
		if len(m.subProcess.Args) > 0 && strings.Contains(m.subProcess.Args[len(m.subProcess.Args)-1], "No app running") {
			if err := m.subProcess.Start(); err != nil {
				return processFinishedMsg(err)
			}
			m.subProcess.Wait()
			return processFinishedMsg(nil)
		}

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

// Test Watcher Logic
func (m *Model) startTestWatcher() tea.Cmd {
	return func() tea.Msg {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return watchErrorMsg(err)
		}
		// m.watcher = watcher // Note: Can't easily assign back to model in Cmd, but we can manage it.
		// Actually, we need to store it to close it.
		// For now, we'll let it leak until quit or manage it differently.
		// A better pattern is to handle the creation in Init or update the model.
		// But since this is a Cmd, we can't modify m. We'll just run the loop.

		// Add recursive watch paths
		// Naive implementation: watch root and features
		_ = watcher.Add(".")
		_ = watcher.Add("./features")
		_ = watcher.Add("./backend/features")

		// We also want to watch subdirectories.
		filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if info != nil && info.IsDir() && !strings.HasPrefix(path, ".") && !strings.Contains(path, "node_modules") {
				_ = watcher.Add(path)
			}
			return nil
		})

		// Run once immediately
		m.runTests()

		// Debounce timer
		var timer *time.Timer

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return nil
				}
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					// Debounce
					if timer != nil {
						timer.Stop()
					}
					timer = time.AfterFunc(500*time.Millisecond, func() {
						m.runTests()
					})
				}
			case _, ok := <-watcher.Errors:
				if !ok {
					return nil
				}
			}
		}
	}
}

func (m *Model) runTests() {
	// Reset status for new run
	// Note: We can't modify m directly here as it's a method on a copy (value receiver)
	// or called from a goroutine.
	// But we can send a message to clear logs?
	// For simplicity in this structure, we just run and the new logs will append.
	// Ideally we'd send a "TestsStarting" message.

	// Determine test path
	testPath := "./..."
	if _, err := os.Stat("backend/features"); err == nil {
		testPath = "./backend/features/..."
	} else if _, err := os.Stat("features"); err == nil {
		testPath = "./features/..."
	}

	cmd := exec.Command("go", "test", "-json", "-v", testPath)
	output, _ := cmd.CombinedOutput()

	// Parse JSON output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var event struct {
			Time    time.Time `json:"Time"`
			Action  string    `json:"Action"`
			Package string    `json:"Package"`
			Test    string    `json:"Test"`
			Output  string    `json:"Output"`
		}

		if err := json.Unmarshal([]byte(line), &event); err == nil {
			// We care about PASS/FAIL events
			if event.Action == "pass" || event.Action == "fail" {
				if event.Test != "" {
					status := "PASS"
					if event.Action == "fail" {
						status = "FAIL"
					}

					entry := LogEntry{
						Timestamp:  event.Time,
						Type:       LogTypeTest,
						TestName:   event.Test,
						TestStatus: status,
						TestOutput: strings.TrimSpace(event.Output),
					}
					m.testCh <- entry
				}
			}
		}
	}
}

func (m *Model) waitForTest() tea.Cmd {
	return func() tea.Msg {
		return testMsg(<-m.testCh)
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
