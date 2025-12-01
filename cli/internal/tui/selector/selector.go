package selector

import (
	"io/ioutil"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	files    []string
	cursor   int
	selected string
	viewport viewport.Model
}

func New(files []string) Model {
	vp := viewport.New(80, 20)
	m := Model{files: files, viewport: vp}
	m.loadFileContent()
	return m
}

func (m *Model) loadFileContent() {
	if m.cursor >= 0 && m.cursor < len(m.files) {
		content, err := ioutil.ReadFile(m.files[m.cursor])
		if err != nil {
			m.viewport.SetContent("Error reading file")
		} else {
			m.viewport.SetContent(string(content))
		}
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.loadFileContent()
			}
		case "down", "j":
			if m.cursor < len(m.files)-1 {
				m.cursor++
				m.loadFileContent()
			}
		case "enter":
			m.selected = m.files[m.cursor]
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	fileList := ""
	for i, file := range m.files {
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
		mainContent,
		footerStyle.Render("Use arrow keys to navigate, enter to select, q to quit"),
	)
}

func (m Model) Selected() string {
	return m.selected
}

