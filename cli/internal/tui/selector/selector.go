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
	selectedSet   map[string]struct{} // multi-select
	viewport      viewport.Model
	textInput     textinput.Model
	multi         bool
}

func New(files []string, multi bool) Model {
	ti := textinput.New()
	ti.Placeholder = "Type to filter..."
	ti.Focus()

	vp := viewport.New(80, 20)
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
	fileList := ""
	for i, file := range m.filteredFiles {
		prefix := "  "
		if _, ok := m.selectedSet[file]; ok {
			prefix = "* " // like fzf's multi select
		}
		if m.cursor == i {
			fileList += selectedItemStyle.Render("> " + prefix + file)
		} else {
			fileList += itemStyle.Render("  " + prefix + file)
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
		footerStyle.Render("↑/↓ move • type to filter • space select • enter accept • q quit"),
	)
}

func (m Model) SelectedList() []string {
	out := make([]string, 0, len(m.selectedSet))
	for f := range m.selectedSet {
		out = append(out, f)
	}
	return out
}

