package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cli/internal/config"
	"cli/internal/executor"
	"cli/internal/filter"
	"cli/internal/output"
	"cli/internal/validation"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#69DB7C"))

	warnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFE066"))

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))
)

type watchModel struct {
	filepath    string
	viewport    viewport.Model
	content     string
	lastMod     time.Time
	lastRun     time.Time
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
	content string
	err     error
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
		cfg, err := config.LoadConfig(path)
		if err != nil {
			return runResultMsg{err: fmt.Errorf("config error: %w", err)}
		}

		issues := validation.ValidateConfig(cfg)
		var warnings []string
		for _, issue := range issues {
			if issue.Severity == validation.SeverityError {
				return runResultMsg{err: fmt.Errorf("%s: %s", issue.Field, issue.Message)}
			}
			if issue.Severity == validation.SeverityWarning {
				warnings = append(warnings, fmt.Sprintf("[WARN] %s: %s", issue.Field, issue.Message))
			}
		}

		body, ctype, err := executeConfig(cfg)
		if err != nil {
			return runResultMsg{err: fmt.Errorf("request failed: %w", err)}
		}

		if cfg.JQFilter != "" {
			body, err = filter.ApplyJQ(body, cfg.JQFilter)
			if err != nil {
				return runResultMsg{err: fmt.Errorf("jq filter failed: %w", err)}
			}
			ctype = "application/json"
		}

		result := output.Highlight(body, ctype, false)
		if len(warnings) > 0 {
			result = strings.Join(warnings, "\n") + "\n\n" + result
		}

		return runResultMsg{content: result}
	}
}

func executeConfig(cfg *config.YapiConfig) (string, string, error) {
	switch cfg.Method {
	case "grpc":
		body, err := executor.NewGRPCExecutor().Execute(cfg)
		return body, "application/json", err
	case "tcp":
		body, err := executor.NewTCPExecutor().Execute(cfg)
		return body, "text/plain", err
	case "GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS", "":
		if cfg.Method == "" {
			cfg.Method = "GET"
		}
		resp, err := executor.NewHTTPExecutor().Execute(cfg)
		if err != nil {
			return "", "", err
		}
		return resp.Body, resp.ContentType, nil
	default:
		return "", "", fmt.Errorf("unsupported method: %s", cfg.Method)
	}
}

func NewWatchModel(path string) watchModel {
	return watchModel{
		filepath:    path,
		content:     "Loading...",
		status:      "starting",
		statusStyle: infoStyle,
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
			m.statusStyle = infoStyle
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
		m.statusStyle = infoStyle
		cmds = append(cmds, runYapiCmd(m.filepath))

	case runResultMsg:
		m.lastRun = time.Now()
		if msg.err != nil {
			m.err = msg.err
			m.content = errorStyle.Render(msg.err.Error())
			m.status = "error"
			m.statusStyle = errorStyle
		} else {
			m.err = nil
			m.content = msg.content
			m.status = "ok"
			m.statusStyle = successStyle
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
	title := titleStyle.Render(" 🐑 yapi watch ")
	fileInfo := infoStyle.Render(filename)
	statusText := m.statusStyle.Render(fmt.Sprintf("[%s]", m.status))
	timeText := infoStyle.Render(m.lastRun.Format("15:04:05"))

	header := lipgloss.JoinHorizontal(
		lipgloss.Center,
		title,
		"  ",
		fileInfo,
		"  ",
		statusText,
		"  ",
		timeText,
	)

	// Footer
	help := helpStyle.Render("q: quit • r: refresh • ↑/↓: scroll")

	// Content
	content := borderStyle.Width(m.width - 2).Render(m.viewport.View())

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
