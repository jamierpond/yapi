package tui

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"yapi.run/cli/internal/core"
	"yapi.run/cli/internal/output"
	"yapi.run/cli/internal/runner"
	"yapi.run/cli/internal/tui/theme"
	"yapi.run/cli/internal/validation"
)

var httpClient = &http.Client{Timeout: 30 * time.Second}
var engine = core.NewEngine(httpClient)

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
		runRes := engine.RunConfig(
			context.Background(),
			path,
			runner.Options{NoColor: false},
		)

		if runRes.Error != nil && runRes.Analysis == nil {
			return runResultMsg{err: runRes.Error}
		}
		if runRes.Analysis == nil {
			return runResultMsg{err: fmt.Errorf("no analysis produced")}
		}

		var b strings.Builder

		// Render warnings/diagnostics in TUI style.
		for _, w := range runRes.Analysis.Warnings {
			fmt.Fprintln(&b, theme.Warn.Render("[WARN] "+w))
		}
		for _, d := range runRes.Analysis.Diagnostics {
			prefix, style := "[INFO]", theme.Info
			if d.Severity == validation.SeverityWarning {
				prefix, style = "[WARN]", theme.Warn
			}
			if d.Severity == validation.SeverityError {
				prefix, style = "[ERROR]", theme.Error
			}
			fmt.Fprintln(&b, style.Render(prefix+" "+d.Message))
		}

		if runRes.Analysis.HasErrors() || runRes.Result == nil {
			return runResultMsg{content: b.String()}
		}

		if b.Len() > 0 {
			b.WriteString("\n")
		}

		out := output.Highlight(runRes.Result.Body, runRes.Result.ContentType, false)
		b.WriteString(out)

		// Add expectation result if present
		if runRes.ExpectRes != nil && (runRes.ExpectRes.AssertionsTotal > 0 || runRes.ExpectRes.StatusChecked) {
			b.WriteString("\n")
			if runRes.ExpectRes.AllPassed() {
				b.WriteString(theme.Success.Render(fmt.Sprintf("assertions: %d/%d passed", runRes.ExpectRes.AssertionsPassed, runRes.ExpectRes.AssertionsTotal)))
			} else {
				b.WriteString(theme.Error.Render(fmt.Sprintf("assertions: %d/%d passed", runRes.ExpectRes.AssertionsPassed, runRes.ExpectRes.AssertionsTotal)))
			}
		}

		// Handle expectation error
		if runRes.Error != nil {
			return runResultMsg{
				content:  b.String(),
				duration: runRes.Result.Duration,
				err:      runRes.Error,
			}
		}

		return runResultMsg{
			content:  b.String(),
			duration: runRes.Result.Duration,
		}
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
	title := theme.Title.Render(" 🐑 yapi watch ")
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
