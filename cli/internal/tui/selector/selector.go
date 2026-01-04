// Package selector provides a TUI file picker component.
package selector

import (
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
	"yapi.run/cli/internal/tui/theme"
)

// Model is the bubbletea model for the file selector.
type Model struct {
	files           []string
	filteredFiles   []string
	cursor          int
	selectedSet     map[string]struct{} // multi-select
	viewport        viewport.Model
	textInput       textinput.Model
	multi           bool
	isVertical      bool
	maxVisibleFiles int
}

// New creates a new file selector model.
func New(files []string, multi bool) Model {
	ti := textinput.New()
	ti.Placeholder = "Type to filter..."
	ti.Focus()
	ti.PromptStyle = lipgloss.NewStyle().Foreground(theme.Accent)
	ti.TextStyle = lipgloss.NewStyle().Foreground(theme.Fg)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(theme.FgMuted)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(theme.Accent)

	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(theme.Fg).
		Background(theme.BgElevated)

	m := Model{
		files:           files,
		filteredFiles:   files,
		selectedSet:     make(map[string]struct{}),
		viewport:        vp,
		textInput:       ti,
		multi:           multi,
		maxVisibleFiles: 10,
	}
	m.loadFileContent()
	return m
}

func (m *Model) loadFileContent() {
	if m.cursor >= 0 && m.cursor < len(m.filteredFiles) {
		content, err := os.ReadFile(m.filteredFiles[m.cursor])
		if err != nil {
			m.viewport.SetContent("Error reading file")
			return
		}
		m.viewport.SetContent(string(content))
		return
	}
	m.viewport.SetContent("")
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		const minWidthForHorizontalLayout = 100
		const minHeightForHorizontalLayout = 19
		const leftPanelWidth = 50
		const leftPanelPadding = 2

		// Chrome heights: theme.App border(2) + padding(2) + header(1) + margin(1) + footer(2) + viewportBorder(2) + viewportPadding(2)
		const chromeHeight = 12

		if msg.Width < minWidthForHorizontalLayout || msg.Height < minHeightForHorizontalLayout {
			m.isVertical = true
			availableWidth := msg.Width - theme.App.GetHorizontalFrameSize()
			m.textInput.Width = availableWidth
			m.viewport.Width = availableWidth - theme.ViewportContent.GetHorizontalFrameSize()
			// In vertical mode, split remaining height between file list and preview
			availableForContent := msg.Height - chromeHeight
			// Give file list ~1/3, preview ~2/3, with minimums
			m.maxVisibleFiles = max(3, availableForContent/3)
			m.viewport.Height = max(5, availableForContent-m.maxVisibleFiles-2) // -2 for preview title + margin
		} else {
			m.isVertical = false
			m.maxVisibleFiles = 10
			m.textInput.Width = leftPanelWidth
			m.viewport.Width = msg.Width - theme.App.GetHorizontalFrameSize() - leftPanelWidth - leftPanelPadding - theme.ViewportContent.GetHorizontalFrameSize()
			m.viewport.Height = msg.Height - chromeHeight
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "up", "ctrl+k":
			if m.cursor > 0 {
				m.cursor--
				m.loadFileContent()
			}
			return m, nil

		case "down", "ctrl+j":
			if m.cursor < len(m.filteredFiles)-1 {
				m.cursor++
				m.loadFileContent()
			}
			return m, nil

		case "pgup":
			m.viewport.LineUp(5)
			return m, nil

		case "pgdown":
			m.viewport.LineDown(5)
			return m, nil

		case " ":
			// toggle selection
			if m.multi && len(m.filteredFiles) > 0 {
				p := m.filteredFiles[m.cursor]
				if _, ok := m.selectedSet[p]; ok {
					delete(m.selectedSet, p)
				} else {
					m.selectedSet[p] = struct{}{}
				}
			}
			return m, nil

		case "enter":
			// In single-select mode, ensure current cursor is selected
			if !m.multi && len(m.filteredFiles) > 0 && m.cursor < len(m.filteredFiles) {
				m.selectedSet = map[string]struct{}{
					m.filteredFiles[m.cursor]: {},
				}
			}
			return m, tea.Quit
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	m.filterFiles()
	m.viewport, _ = m.viewport.Update(msg)
	return m, cmd
}

func (m *Model) filterFiles() {
	query := m.textInput.Value()
	if query == "" {
		m.filteredFiles = m.files
	} else {
		matches := fuzzy.Find(query, m.files)
		m.filteredFiles = make([]string, len(matches))
		for i, match := range matches {
			m.filteredFiles[i] = match.Str
		}
	}

	if m.cursor >= len(m.filteredFiles) {
		if len(m.filteredFiles) > 0 {
			m.cursor = len(m.filteredFiles) - 1
		} else {
			m.cursor = 0
		}
	}
	m.loadFileContent()
}

// visibleWindow calculates the start and end indices for a scrolling window.
func visibleWindow(total, cursor, max int) (start, end int) {
	if max <= 0 || total <= max {
		return 0, total
	}
	start = cursor - max/2
	if start < 0 {
		start = 0
	}
	end = start + max
	if end > total {
		end = total
		start = end - max
		if start < 0 {
			start = 0
		}
	}
	return
}

// View implements tea.Model.
func (m Model) View() string {
	fileList := ""
	maxVisible := m.maxVisibleFiles
	if maxVisible < 1 {
		maxVisible = 10
	}

	start, end := visibleWindow(len(m.filteredFiles), m.cursor, maxVisible)

	for i := start; i < end; i++ {
		file := m.filteredFiles[i]
		prefix := "  "
		if _, ok := m.selectedSet[file]; ok {
			prefix = lipgloss.NewStyle().Foreground(theme.Accent).Render("* ")
		}

		style := theme.Item
		if m.cursor == i {
			style = theme.SelectedItem
		}

		renderedLine := style.Render("> " + prefix + file)
		if m.cursor != i {
			renderedLine = style.Render("  " + prefix + file)
		}
		fileList += renderedLine + "\n"
	}
	// --- Viewport ---
	viewportContent := theme.ViewportContent.Render(m.viewport.View())

	// --- Left Panel (input + file list) ---
	leftPanel := lipgloss.JoinVertical(
		lipgloss.Left,
		m.textInput.View(),
		fileList,
	)

	// --- Assemble Layout ---
	var mainContent string
	if m.isVertical {
		// In vertical mode, skip Preview title to save space
		mainContent = lipgloss.JoinVertical(
			lipgloss.Left,
			leftPanel,
			viewportContent,
		)
	} else {
		const leftPanelWidth = 50
		const leftPanelPadding = 2
		viewportTitle := theme.TitleAccent.Render("Preview")
		viewportFull := lipgloss.JoinVertical(lipgloss.Left, viewportTitle, viewportContent)
		mainContent = lipgloss.JoinHorizontal(
			lipgloss.Top,
			lipgloss.NewStyle().Width(leftPanelWidth).PaddingRight(leftPanelPadding).Render(leftPanel),
			lipgloss.NewStyle().Render(viewportFull),
		)
	}

	// --- Header ---
	header := theme.TitleAccent.Render("yapi")

	// --- Final Layout ---
	var content string
	if m.isVertical {
		// Compact layout: small margin after header, no footer
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			"",
			mainContent,
		)
	} else {
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			lipgloss.NewStyle().MarginTop(1).Render(mainContent),
			theme.Footer.Render("↑/↓ move | type to filter | space select | enter accept | esc quit"),
		)
	}
	return theme.App.Render(content)
}

// SelectedList returns the list of selected file paths.
func (m Model) SelectedList() []string {
	out := make([]string, 0, len(m.selectedSet))
	for f := range m.selectedSet {
		out = append(out, f)
	}
	return out
}
