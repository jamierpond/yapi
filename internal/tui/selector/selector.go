package selector

import (
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

var (
	// Colors (extracted from webapp/tailwind.config.js)
	yapiBg         = lipgloss.Color("#1a1b26")
	yapiBgElevated = lipgloss.Color("#2a2d3b")
	yapiFg         = lipgloss.Color("#a9b1d6")
	yapiFgMuted    = lipgloss.Color("#565f89")
	yapiAccent     = lipgloss.Color("#ff9e64")
	yapiBorder     = lipgloss.Color("#414868")

	// Styles
	appStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(yapiBorder)

	titleStyle = lipgloss.NewStyle().
			Foreground(yapiBg).
			Background(yapiAccent).
			Padding(0, 1).
			Bold(true)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(yapiAccent).
				Bold(true)

	footerStyle = lipgloss.NewStyle().
			Foreground(yapiFgMuted).
			Padding(0, 1).
			MarginTop(1)

	viewportContentStyle = lipgloss.NewStyle().
				Padding(1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(yapiBorder)
)

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

func New(files []string, multi bool) Model {
	ti := textinput.New()
	ti.Placeholder = "Type to filter..."
	ti.Focus()
	ti.PromptStyle = lipgloss.NewStyle().Foreground(yapiAccent)
	ti.TextStyle = lipgloss.NewStyle().Foreground(yapiFg)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(yapiFgMuted)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(yapiAccent)

	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(yapiFg).
		Background(yapiBgElevated)

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

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		const minWidthForHorizontalLayout = 100
		const minHeightForHorizontalLayout = 19
		const leftPanelWidth = 50
		const leftPanelPadding = 2

		// Chrome heights: appStyle border(2) + padding(2) + header(1) + margin(1) + footer(2) + viewportBorder(2) + viewportPadding(2)
		const chromeHeight = 12

		if msg.Width < minWidthForHorizontalLayout || msg.Height < minHeightForHorizontalLayout {
			m.isVertical = true
			availableWidth := msg.Width - appStyle.GetHorizontalFrameSize()
			m.textInput.Width = availableWidth
			m.viewport.Width = availableWidth - viewportContentStyle.GetHorizontalFrameSize()
			// In vertical mode, split remaining height between file list and preview
			availableForContent := msg.Height - chromeHeight
			// Give file list ~1/3, preview ~2/3, with minimums
			m.maxVisibleFiles = max(3, availableForContent/3)
			m.viewport.Height = max(5, availableForContent-m.maxVisibleFiles-2) // -2 for preview title + margin
		} else {
			m.isVertical = false
			m.maxVisibleFiles = 10
			m.textInput.Width = leftPanelWidth
			m.viewport.Width = msg.Width - appStyle.GetHorizontalFrameSize() - leftPanelWidth - leftPanelPadding - viewportContentStyle.GetHorizontalFrameSize()
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

		case "pgup", "b":
			m.viewport.LineUp(5)
			return m, nil

		case "pgdown", "f":
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

func (m Model) View() string {
	// --- File List (with virtual scrolling) ---
	fileList := ""
	maxVisible := m.maxVisibleFiles
	if maxVisible < 1 {
		maxVisible = 10
	}
	var visibleFileStartIndex int

	// Determine the slice of files to show
	if len(m.filteredFiles) > maxVisible {
		visibleFileStartIndex = m.cursor - (maxVisible / 2)
		if visibleFileStartIndex < 0 {
			visibleFileStartIndex = 0
		}
		endIndex := visibleFileStartIndex + maxVisible
		if endIndex > len(m.filteredFiles) {
			endIndex = len(m.filteredFiles)
			visibleFileStartIndex = endIndex - maxVisible
			if visibleFileStartIndex < 0 {
				visibleFileStartIndex = 0
			}
		}
	}

	endIndex := visibleFileStartIndex + maxVisible
	if endIndex > len(m.filteredFiles) {
		endIndex = len(m.filteredFiles)
	}

	// Render only the visible files
	for i := visibleFileStartIndex; i < endIndex; i++ {
		file := m.filteredFiles[i]
		prefix := "  "
		if _, ok := m.selectedSet[file]; ok {
			prefix = lipgloss.NewStyle().Foreground(yapiAccent).Render("* ")
		}

		style := itemStyle
		if m.cursor == i {
			style = selectedItemStyle
		}

		renderedLine := style.Render("> " + prefix + file)
		if m.cursor != i {
			renderedLine = style.Render("  " + prefix + file)
		}
		fileList += renderedLine + "\n"
	}
	// --- Viewport ---
	viewportContent := viewportContentStyle.Render(m.viewport.View())

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
		viewportTitle := titleStyle.Render("Preview")
		viewportFull := lipgloss.JoinVertical(lipgloss.Left, viewportTitle, viewportContent)
		mainContent = lipgloss.JoinHorizontal(
			lipgloss.Top,
			lipgloss.NewStyle().Width(leftPanelWidth).PaddingRight(leftPanelPadding).Render(leftPanel),
			lipgloss.NewStyle().Render(viewportFull),
		)
	}

	// --- Header ---
	header := titleStyle.Render("🐑 yapi")

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
			footerStyle.Render("↑/↓ move | type to filter | space select | enter accept | esc quit"),
		)
	}
	return appStyle.Render(content)
}

func (m Model) SelectedList() []string {
	out := make([]string, 0, len(m.selectedSet))
	for f := range m.selectedSet {
		out = append(out, f)
	}
	return out
}
