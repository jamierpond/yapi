// Package tui provides terminal UI components for yapi.
package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"yapi.run/cli/internal/tui/selector"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-git/go-git/v5"
	"github.com/mattn/go-isatty"
)

// getTTY returns input and output file handles for interactive TUI.
// On Unix, it tries /dev/tty first to work when stdout is piped.
// On Windows, it uses stdin/stdout directly.
// Returns nil, nil if no TTY is available.
func getTTY() (in, out *os.File, cleanup func()) {
	cleanup = func() {} // no-op by default

	// On Unix, try /dev/tty for piped scenarios
	if runtime.GOOS != "windows" {
		tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
		if err == nil {
			return tty, tty, func() { _ = tty.Close() }
		}
	}

	// Fall back to stdin/stdout if they're terminals
	if isatty.IsTerminal(os.Stdin.Fd()) && isatty.IsTerminal(os.Stdout.Fd()) {
		return os.Stdin, os.Stdout, cleanup
	}

	// Also check for Cygwin/MSYS terminals on Windows
	if isatty.IsCygwinTerminal(os.Stdin.Fd()) && isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		return os.Stdin, os.Stdout, cleanup
	}

	return nil, nil, cleanup
}

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

// FindConfigFileSingle prompts the user to select a single config file.
func FindConfigFileSingle() (string, error) {
	files, err := findFiles()
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		return "", fmt.Errorf("no .yapi.yml files found")
	}

	in, out, cleanup := getTTY()
	defer cleanup()

	if in == nil || out == nil {
		// No TTY at all (CI, cron, etc) -> non-interactive fallback
		return files[0], nil
	}

	// Render TUI to the chosen terminal, not to stdout.
	lipgloss.SetDefaultRenderer(lipgloss.NewRenderer(out))

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
