package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"yapi.run/cli/internal/tui/selector"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-git/go-git/v5"
	"github.com/mattn/go-isatty"
)

// yapiFilePattern matches *.yapi.yaml or *.yapi.yml in subdirectories only
var yapiFilePattern = regexp.MustCompile(`^.+/.+\.yapi\.ya?ml$`)

func findFiles() ([]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Open the git repository (searches up for .git)
	repo, err := git.PlainOpenWithOptions(cwd, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return nil, fmt.Errorf("not in a git repository: %w", err)
	}

	// Get worktree to find repo root
	wt, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}
	repoRoot := wt.Filesystem.Root()

	// Read the git index (staged files = tracked files)
	idx, err := repo.Storer.Index()
	if err != nil {
		return nil, fmt.Errorf("failed to read git index: %w", err)
	}

	// Calculate relative path from repo root to cwd
	relCwd, err := filepath.Rel(repoRoot, cwd)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate relative path: %w", err)
	}
	if relCwd == "." {
		relCwd = ""
	}

	var configFiles []string
	for _, entry := range idx.Entries {
		path := entry.Name

		// Skip if not under current directory
		if relCwd != "" && !strings.HasPrefix(path, relCwd+"/") {
			continue
		}

		// Get path relative to cwd
		var relPath string
		if relCwd != "" {
			relPath = strings.TrimPrefix(path, relCwd+"/")
		} else {
			relPath = path
		}

		// Must be in a subdirectory and match .yapi.y[a]ml
		if yapiFilePattern.MatchString(relPath) {
			configFiles = append(configFiles, relPath)
		}
	}

	if len(configFiles) == 0 {
		return nil, fmt.Errorf("no .yapi.yaml/.yapi.yml files found in subdirectories")
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
