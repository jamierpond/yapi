package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/charmbracelet/bubbletea"
	"cli/internal/tui/selector"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
)

func findFiles() ([]string, error) {
	var configFiles []string
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	dir := currentDir
	for i := 0; i < 5; i++ { // Limit search to current + 4 parent directories
		files, err := filepath.Glob(filepath.Join(dir, "*.yapi.yml"))
		if err != nil {
			return nil, fmt.Errorf("failed to glob for config files in %s: %w", dir, err)
		}
		for _, file := range files {
			// Store relative path to CWD
			relPath, err := filepath.Rel(currentDir, file)
			if err != nil {
				relPath = file // Fallback to absolute path if relative fails
			}
			configFiles = append(configFiles, relPath)
		}

		// Also search for config files in the adjacent 'examples' directory
		if dir == currentDir { // Only search 'examples' relative to the initial CWD
			exampleFiles, err := filepath.Glob(filepath.Join(filepath.Dir(currentDir), "examples", "*.yapi.yml"))
			if err != nil {
				return nil, fmt.Errorf("failed to glob for example config files: %w", err)
			}
			for _, file := range exampleFiles {
				relPath, err := filepath.Rel(currentDir, file)
				if err != nil {
					relPath = file
				}
				configFiles = append(configFiles, relPath)
			}
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir { // Reached root or unable to go up
			break
		}
		dir = parentDir
	}

	if len(configFiles) == 0 {
		return nil, fmt.Errorf("no .yapi.yml files found in current or parent directories")
	}

	sort.Strings(configFiles)
	return configFiles, nil
}


func FindConfigFileSingle() (string, error) {
	files, err := findFiles()
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		return "", fmt.Errorf("no .yapi.yml files found")
	}

	var in, out *os.File
	// Prefer /dev/tty for interactive TUI so it still works when stdout is piped.
	// Example: yapi pick | jq
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err == nil {
		in = tty
		out = tty
		defer tty.Close()
	} else if isatty.IsTerminal(os.Stdout.Fd()) {
		in = os.Stdin
		out = os.Stdout
	} else {
		// No TTY at all (CI, cron, etc) -> non-interactive fallback
		return files[0], nil
	}


	os.Setenv("CLICOLOR_FORCE", "1")
	// Render TUI to the chosen terminal, not to stdout.
	renderer := lipgloss.NewRenderer(out)
	lipgloss.SetDefaultRenderer(renderer)

	p := tea.NewProgram(
		selector.New(files, false),
		tea.WithInput(in),
		tea.WithOutput(out),
		tea.WithAltScreen(),
	)

	m, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("failed to run selector: %w", err)
	}

	model := m.(selector.Model)
	selected := model.SelectedList()
	if len(selected) == 0 {
		return "", fmt.Errorf("no config file selected")
	}

	// The caller still prints the final path(s) to stdout,
	// which can safely be piped to jq, xargs, etc.
	return selected[0], nil
}

func FindConfigFileMulti(multi bool) ([]string, error) {
	files, err := findFiles()
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no .yapi.yml files found")
	}

	var in, out *os.File
	// Same TTY detection strategy as FindConfigFileSingle.
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err == nil {
		in = tty
		out = tty
		defer tty.Close()
	} else if isatty.IsTerminal(os.Stdout.Fd()) {
		in = os.Stdin
		out = os.Stdout
	} else {
		// No TTY -> just return the list for non-interactive use
		return files, nil
	}

	os.Setenv("CLICOLOR_FORCE", "1")
	lipgloss.SetDefaultRenderer(lipgloss.NewRenderer(out))

	p := tea.NewProgram(
		selector.New(files, multi),
		tea.WithInput(in),
		tea.WithOutput(out),
		tea.WithAltScreen(),
	)

	m, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run selector: %w", err)
	}

	model := m.(selector.Model)
	selected := model.SelectedList()
	if len(selected) == 0 {
		return nil, fmt.Errorf("no config file selected")
	}

	return selected, nil
}

