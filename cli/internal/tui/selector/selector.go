package selector

import (
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
	"golang.org/x/term"
)

var (
	// Colors (extracted from webapp/tailwind.config.js)
	yapiBg        = lipgloss.Color("#1a1b26")
	yapiBgElevated= lipgloss.Color("#2a2d3b")
	yapiFg        = lipgloss.Color("#a9b1d6")
	yapiFgMuted   = lipgloss.Color("#565f89")
	yapiAccent    = lipgloss.Color("#ff9e64")
	yapiBorder    = lipgloss.Color("#414868")

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
)

type Model struct {
	files         []string
	filteredFiles []string
	cursor        int
	selectedSet   map[string]struct{} // multi-select
	viewport      viewport.Model
	textInput     textinput.Model
	multi         bool
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
		files:         files,
		filteredFiles: files,
		selectedSet:   make(map[string]struct{}),
		viewport:      vp,
		textInput:     ti,
		multi:         multi,
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
		// This message is sent when the terminal is resized.
		// We can use it to recalculate layouts.
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.loadFileContent()
			}
			return m, nil

		case "down", "j":
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
	// --- Get Terminal Size ---
	width, height, _ := term.GetSize(int(os.Stdout.Fd()))

	// --- Constants ---
	const minWidthForHorizontalLayout = 100
	const leftPanelWidth = 50
	const leftPanelPadding = 2

	// --- File List ---
	fileList := ""
	for i, file := range m.filteredFiles {
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
	fileList = lipgloss.NewStyle().
		Padding(1, 0, 0, 0).
		Render(fileList)

	// --- Viewport ---
	viewportTitle := titleStyle.Render("Preview")
	viewportContentStyle := lipgloss.NewStyle().
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(yapiBorder)

	// --- Left Panel (input + file list) ---
	leftPanel := lipgloss.JoinVertical(
		lipgloss.Left,
		m.textInput.View(),
		fileList,
	)

	// --- Decide Layout ---
	var mainContent string
	if width < minWidthForHorizontalLayout {
		// --- VERTICAL LAYOUT (for small screens) ---
		m.viewport.Width = width - appStyle.GetHorizontalFrameSize() - viewportContentStyle.GetHorizontalFrameSize()
		m.viewport.Height = height - 15 // Adjust height dynamically

		viewportContent := viewportContentStyle.Render(m.viewport.View())
		viewportFull := lipgloss.JoinVertical(lipgloss.Left, viewportTitle, viewportContent)

		mainContent = lipgloss.JoinVertical(
			lipgloss.Left,
			leftPanel,
			lipgloss.NewStyle().MarginTop(1).Render(viewportFull),
		)
	} else {
		// --- HORIZONTAL LAYOUT (for large screens) ---
		m.viewport.Width = width - appStyle.GetHorizontalFrameSize() - leftPanelWidth - leftPanelPadding - viewportContentStyle.GetHorizontalFrameSize()
		m.viewport.Height = height - 12 // Adjust height dynamically

		viewportContent := viewportContentStyle.Render(m.viewport.View())
		viewportFull := lipgloss.JoinVertical(lipgloss.Left, viewportTitle, viewportContent)

		mainContent = lipgloss.JoinHorizontal(
			lipgloss.Top,
			lipgloss.NewStyle().Width(leftPanelWidth).PaddingRight(leftPanelPadding).Render(leftPanel),
			lipgloss.NewStyle().Render(viewportFull),
		)
	}

	// --- Final Layout ---
	return appStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			titleStyle.Render("🐑 yapi"),
			lipgloss.NewStyle().MarginTop(1).Render(mainContent),
			footerStyle.Render("↑/↓ move • type to filter • space select • enter accept • q quit"),
		),
	)
}

func (m Model) SelectedList() []string {
	out := make([]string, 0, len(m.selectedSet))
	for f := range m.selectedSet {
		out = append(out, f)
	}
	return out
}

