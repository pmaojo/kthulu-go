package coder

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)


// FilePicker provides fuzzy file selection
type FilePicker struct {
	visible     bool
	input       textinput.Model
	files       []string
	filtered    []string
	selected    int
	maxVisible  int
	offset      int
	rootDir     string
	width       int
	height      int
	styles      FilePickerStyles
}

// FilePickerStyles defines the picker appearance
type FilePickerStyles struct {
	Modal       lipgloss.Style
	Title       lipgloss.Style
	Input       lipgloss.Style
	Item        lipgloss.Style
	ItemFocus   lipgloss.Style
	Dir         lipgloss.Style
	Count       lipgloss.Style
}

// DefaultFilePickerStyles returns default picker styles
func DefaultFilePickerStyles() FilePickerStyles {
	return FilePickerStyles{
		Modal: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(DefaultTheme.Primary).
			Padding(1, 2).
			Background(lipgloss.Color("#100505")),
		Title: lipgloss.NewStyle().
			Foreground(DefaultTheme.Primary).
			Bold(true).
			MarginBottom(1),
		Input: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(DefaultTheme.Muted).
			Padding(0, 1),
		Item: lipgloss.NewStyle().
			Foreground(DefaultTheme.Foreground).
			PaddingLeft(2),
		ItemFocus: lipgloss.NewStyle().
			Foreground(DefaultTheme.Accent).
			Bold(true).
			PaddingLeft(1).
			SetString("▶ "),
		Dir: lipgloss.NewStyle().
			Foreground(DefaultTheme.Secondary),
		Count: lipgloss.NewStyle().
			Foreground(DefaultTheme.Muted),
	}
}


// NewFilePicker creates a new file picker
func NewFilePicker(rootDir string) *FilePicker {
	ti := textinput.New()
	ti.Placeholder = "Type to search files..."
	ti.Focus()
	ti.CharLimit = 100

	return &FilePicker{
		input:      ti,
		rootDir:    rootDir,
		maxVisible: 15,
		styles:     DefaultFilePickerStyles(),
	}
}

// Show displays the file picker
func (fp *FilePicker) Show() tea.Cmd {
	fp.visible = true
	fp.input.Reset()
	fp.selected = 0
	fp.offset = 0
	
	// Scan files in background
	return fp.scanFiles
}

// Hide hides the file picker
func (fp *FilePicker) Hide() {
	fp.visible = false
}

// IsVisible returns whether the picker is showing
func (fp *FilePicker) IsVisible() bool {
	return fp.visible
}

// SetSize updates dimensions
func (fp *FilePicker) SetSize(width, height int) {
	fp.width = width
	fp.height = height
	fp.maxVisible = (height * 60 / 100) - 6
	if fp.maxVisible < 5 {
		fp.maxVisible = 5
	}
}

// Selected returns the currently selected file
func (fp *FilePicker) Selected() string {
	if fp.selected >= 0 && fp.selected < len(fp.filtered) {
		return fp.filtered[fp.selected]
	}
	return ""
}

// Update handles input for the picker
func (fp *FilePicker) Update(msg tea.Msg) (selected string, done bool, cmd tea.Cmd) {
	if !fp.visible {
		return "", false, nil
	}

	switch msg := msg.(type) {
	case filesScannedMsg:
		fp.files = msg.files
		fp.filter()
		return "", false, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			fp.Hide()
			return "", true, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			selected := fp.Selected()
			fp.Hide()
			return selected, true, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "ctrl+p"))):
			if fp.selected > 0 {
				fp.selected--
				if fp.selected < fp.offset {
					fp.offset = fp.selected
				}
			}
			return "", false, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "ctrl+n"))):
			if fp.selected < len(fp.filtered)-1 {
				fp.selected++
				if fp.selected >= fp.offset+fp.maxVisible {
					fp.offset = fp.selected - fp.maxVisible + 1
				}
			}
			return "", false, nil
		}
	}

	// Update text input
	var cmd2 tea.Cmd
	fp.input, cmd2 = fp.input.Update(msg)
	
	// Re-filter on input change
	oldLen := len(fp.filtered)
	fp.filter()
	if len(fp.filtered) != oldLen {
		fp.selected = 0
		fp.offset = 0
	}

	return "", false, cmd2
}

// View renders the file picker
func (fp *FilePicker) View() string {
	if !fp.visible {
		return ""
	}

	var content strings.Builder

	// Title
	title := fp.styles.Title.Render("📂 Select File")
	content.WriteString(title + "\n\n")

	// Search input
	inputView := fp.styles.Input.Render(fp.input.View())
	content.WriteString(inputView + "\n\n")

	// File count
	count := fp.styles.Count.Render(
		strings.Repeat(" ", 2) + 
		formatCount(len(fp.filtered), len(fp.files)))
	content.WriteString(count + "\n")

	// File list
	end := fp.offset + fp.maxVisible
	if end > len(fp.filtered) {
		end = len(fp.filtered)
	}

	for i := fp.offset; i < end; i++ {
		file := fp.filtered[i]
		var line string
		
		if i == fp.selected {
			line = fp.styles.ItemFocus.Render("") + fp.styles.ItemFocus.Render(file)
		} else {
			// Color directories differently
			if isDir(filepath.Join(fp.rootDir, file)) {
				line = fp.styles.Dir.Render("  📁 " + file)
			} else {
				line = fp.styles.Item.Render("  📄 " + file)
			}
		}
		content.WriteString(line + "\n")
	}

	// Help
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6C7086")).
		Render("\n↑/↓ navigate • Enter select • Esc cancel")
	content.WriteString(help)

	// Modal size
	modalWidth := fp.width * 60 / 100
	if modalWidth < 50 {
		modalWidth = 50
	}

	modal := fp.styles.Modal.
		Width(modalWidth).
		Render(content.String())

	return lipgloss.Place(
		fp.width, fp.height,
		lipgloss.Center, lipgloss.Center,
		modal,
	)
}

// filter applies fuzzy filtering based on input
func (fp *FilePicker) filter() {
	query := strings.ToLower(fp.input.Value())
	if query == "" {
		fp.filtered = fp.files
		return
	}

	fp.filtered = make([]string, 0)
	for _, file := range fp.files {
		if fuzzyMatch(strings.ToLower(file), query) {
			fp.filtered = append(fp.filtered, file)
		}
	}
}

// fuzzyMatch performs simple fuzzy matching
func fuzzyMatch(s, pattern string) bool {
	pi := 0
	for si := 0; si < len(s) && pi < len(pattern); si++ {
		if s[si] == pattern[pi] {
			pi++
		}
	}
	return pi == len(pattern)
}

// scanFiles scans the directory for files
func (fp *FilePicker) scanFiles() tea.Msg {
	var files []string
	
	filepath.Walk(fp.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip hidden files and common ignore patterns
		name := info.Name()
		if strings.HasPrefix(name, ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip common directories
		if info.IsDir() {
			skip := []string{"node_modules", "vendor", "dist", "build", "__pycache__"}
			for _, s := range skip {
				if name == s {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Get relative path
		rel, err := filepath.Rel(fp.rootDir, path)
		if err != nil {
			return nil
		}

		files = append(files, rel)
		return nil
	})

	// Sort by name
	sort.Strings(files)

	return filesScannedMsg{files: files}
}

type filesScannedMsg struct {
	files []string
}

func formatCount(filtered, total int) string {
	if filtered == total {
		return fmt.Sprintf("%d files", total)
	}
	return fmt.Sprintf("%d / %d files", filtered, total)
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
