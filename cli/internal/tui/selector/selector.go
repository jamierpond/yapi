package selector

import (
	"io/ioutil"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#6151E2")).
			Padding(0, 1)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("#6151E2"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#3C3C3C")).
			Padding(0, 1)
)

type Model struct {
	files         []string
	filteredFiles []string
	cursor        int
	selected      string
	viewport      viewport.Model
	textInput     textinput.Model
}

func New(files []string) Model {
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.Focus()

	vp := viewport.New(80, 20)
	m := Model{
		files:         files,
		filteredFiles: files,
		viewport:      vp,
		textInput:     ti,
	}
	m.loadFileContent()
	return m
}

func (m *Model) loadFileContent() {
	if m.cursor >= 0 && m.cursor < len(m.filteredFiles) {
		content, err := ioutil.ReadFile(m.filteredFiles[m.cursor])
		if err != nil {
			m.viewport.SetContent("Error reading file")
		} else {
			m.viewport.SetContent(string(content))
		}
	} else {
		m.viewport.SetContent("")
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
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
		case "enter":
			if len(m.filteredFiles) > 0 && m.cursor < len(m.filteredFiles) {
				m.selected = m.filteredFiles[m.cursor]
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
	fileList := ""
	for i, file := range m.filteredFiles {
		if m.cursor == i {
			fileList += selectedItemStyle.Render("> " + file)
		} else {
			fileList += itemStyle.Render("  " + file)
		}
		fileList += "\n"
	}

	mainContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(40).Render(fileList),
		lipgloss.NewStyle().Width(80).Render(m.viewport.View()),
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render("Select a config file"),
		m.textInput.View(),
		mainContent,
		footerStyle.Render("Use arrow keys to navigate, enter to select, q to quit"),
	)
}

func (m Model) Selected() string {
	return m.selected
}

