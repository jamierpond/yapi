package tui

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"yapi.run/cli/internal/executor"
	"yapi.run/cli/internal/runner"
	"yapi.run/cli/internal/tui/theme"
	"yapi.run/cli/internal/validation"
)

type watchModel struct {
	filepath    string
	viewport    viewport.Model
	content     string
	lastMod     time.Time
	lastRun     time.Time
	duration    time.Duration
	err         error
	width       int
	height      int
	ready       bool
	status      string
	statusStyle lipgloss.Style
}

type tickMsg time.Time
type fileChangedMsg struct{}
type runResultMsg struct {
	content  string
	err      error
	duration time.Duration
}

func tickCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func checkFileCmd(path string, lastMod time.Time) tea.Cmd {
	return func() tea.Msg {
		info, err := os.Stat(path)
		if err != nil {
			return nil
		}
		if info.ModTime().After(lastMod) {
			return fileChangedMsg{}
		}
		return nil
	}
}

func runYapiCmd(path string) tea.Cmd {
	return func() tea.Msg {
		analysis, err := validation.AnalyzeConfigFile(path)
		if err != nil {
			return runResultMsg{err: fmt.Errorf("analysis error: %w", err)}
		}

		var outputText string

		// Warnings
		for _, w := range analysis.Warnings {
			outputText += theme.Warn.Render("[WARN] "+w) + "\n"
		}

		// Diagnostics
		for _, d := range analysis.Diagnostics {
			prefix := "[INFO]"
			style := theme.Info
			if d.Severity == validation.SeverityWarning {
				prefix = "[WARN]"
				style = theme.Warn
			}
			if d.Severity == validation.SeverityError {
				prefix = "[ERROR]"
				style = theme.Error
			}
			outputText += style.Render(prefix+" "+d.Message) + "\n"
		}

		if analysis.HasErrors() {
			return runResultMsg{content: outputText, err: nil}
		}

		req := analysis.Request
		if req == nil {
			return runResultMsg{err: fmt.Errorf("no request parsed")}
		}

		if outputText != "" {
			outputText += "\n"
		}

		// Create executor
		httpClient := &http.Client{Timeout: 30 * time.Second}
		factory := executor.NewFactory(httpClient)
		exec, err := factory.Create(req.Metadata["transport"])
		if err != nil {
			return runResultMsg{err: err}
		}

		opts := runner.Options{NoColor: false}
		output, result, err := runner.RunAndFormat(context.Background(), exec, req, analysis.Warnings, opts)
		if err != nil {
			return runResultMsg{err: err}
		}

		return runResultMsg{content: outputText + output, duration: result.Duration}
	}
}

func NewWatchModel(path string) watchModel {
	return watchModel{
		filepath:    path,
		content:     "Loading...",
		status:      "starting",
		statusStyle: theme.Info,
	}
}

func (m watchModel) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		runYapiCmd(m.filepath),
	)
}

func (m watchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "r":
			m.status = "running..."
			m.statusStyle = theme.Info
			return m, runYapiCmd(m.filepath)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := 3
		footerHeight := 2
		verticalMargin := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width-4, msg.Height-verticalMargin)
			m.viewport.SetContent(m.content)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width - 4
			m.viewport.Height = msg.Height - verticalMargin
		}

	case tickMsg:
		cmds = append(cmds, tickCmd())
		cmds = append(cmds, checkFileCmd(m.filepath, m.lastMod))

	case fileChangedMsg:
		info, _ := os.Stat(m.filepath)
		if info != nil {
			m.lastMod = info.ModTime()
		}
		m.status = "running..."
		m.statusStyle = theme.Info
		cmds = append(cmds, runYapiCmd(m.filepath))

	case runResultMsg:
		m.lastRun = time.Now()
		m.duration = msg.duration
		if msg.err != nil {
			m.err = msg.err
			m.content = theme.Error.Render(msg.err.Error())
			m.status = "error"
			m.statusStyle = theme.Error
		} else {
			m.err = nil
			m.content = msg.content
			m.status = "ok"
			m.statusStyle = theme.Success
		}
		if m.ready {
			m.viewport.SetContent(m.content)
			m.viewport.GotoTop()
		}
	}

	if m.ready {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m watchModel) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// Header
	filename := filepath.Base(m.filepath)
	title := theme.Title.Render(" yapi watch ")
	fileInfo := theme.Info.Render(filename)
	statusText := m.statusStyle.Render(fmt.Sprintf("[%s]", m.status))
	timeText := theme.Info.Render(m.lastRun.Format("15:04:05"))
	durationText := theme.Info.Render(fmt.Sprintf("(%s)", m.duration.Round(time.Millisecond)))

	header := lipgloss.JoinHorizontal(
		lipgloss.Center,
		title,
		"  ",
		fileInfo,
		"  ",
		statusText,
		"  ",
		timeText,
		"  ",
		durationText,
	)

	// Footer
	help := theme.Help.Render("q: quit • r: refresh • ↑/↓: scroll")

	// Content
	content := theme.BorderedBox.Width(m.width - 2).Render(m.viewport.View())

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		content,
		help,
	)
}

func RunWatch(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// Get initial mod time
	info, err := os.Stat(absPath)
	if err != nil {
		return err
	}

	model := NewWatchModel(absPath)
	model.lastMod = info.ModTime()

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	_, err = p.Run()
	return err
}
