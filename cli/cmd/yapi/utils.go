package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"yapi.run/cli/internal/config"
	"yapi.run/cli/internal/tui"
	"yapi.run/cli/internal/validation"
)

// selectConfigFile returns the config file path, handling interactive TUI selection when no args provided.
// Returns (selectedPath, fromTUI, error).
func selectConfigFile(args []string, cmdName string) (string, bool, error) {
	return selectConfigFileWithOptions(args, cmdName, false)
}

// selectConfigFileIncludingProject returns the config file path, including project config files in TUI.
// Returns (selectedPath, fromTUI, error).
func selectConfigFileIncludingProject(args []string, cmdName string) (string, bool, error) {
	return selectConfigFileWithOptions(args, cmdName, true)
}

func selectConfigFileWithOptions(args []string, cmdName string, includeProjectConfig bool) (string, bool, error) {
	if len(args) > 0 {
		return args[0], false, nil
	}

	var selectedPath string
	var err error
	if includeProjectConfig {
		selectedPath, err = tui.FindConfigFileSingleIncludingProject()
	} else {
		selectedPath, err = tui.FindConfigFileSingle()
	}
	if err != nil {
		return "", false, fmt.Errorf("failed to select config file: %w", err)
	}

	// Log to history with from_tui flag
	absPath, _ := filepath.Abs(selectedPath)
	logHistoryFromTUI(fmt.Sprintf("yapi %s %q", cmdName, absPath))

	return selectedPath, true, nil
}

// projectEnvResult holds the result of loading a project and optional environment
type projectEnvResult struct {
	project     *config.ProjectConfigV1
	projectRoot string
	envVars     map[string]string
	envName     string
}

// loadProjectAndEnv loads project config and optional environment variables.
// Returns nil result with no error if no project found (not an error condition).
// Returns error only for actual load failures.
func loadProjectAndEnv(configPath string, requestedEnv string, checkRequirement bool) (*projectEnvResult, error) {
	// Resolve absolute path
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}
	configDir := filepath.Dir(absPath)

	// Try to find project root
	projectRoot, err := config.FindProjectRoot(configDir)
	if err != nil {
		// No project found - not an error if no env was requested
		if requestedEnv != "" {
			return nil, fmt.Errorf("--env flag specified but no yapi.config.yml found in directory tree")
		}
		return nil, nil
	}

	// Load project config
	project, err := config.LoadProject(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load project config: %w", err)
	}

	result := &projectEnvResult{
		project:     project,
		projectRoot: projectRoot,
	}

	// Determine which environment to use
	envName := requestedEnv
	if envName == "" && project.DefaultEnvironment != "" {
		envName = project.DefaultEnvironment
	}

	// Check if config requires an environment (only if requested and no explicit env)
	if envName == "" && checkRequirement {
		configData, readErr := os.ReadFile(configPath) // #nosec G304 -- configPath is validated user-provided config file path
		if readErr != nil {
			return nil, fmt.Errorf("failed to read config: %w", readErr)
		}
		requestEnvFileVars := validation.ExtractRequestEnvFileVars(string(configData))
		req := validation.CheckEnvironmentRequirement(string(configData), project, projectRoot, requestEnvFileVars)
		if req.Required {
			return nil, fmt.Errorf("%s", req.Message)
		}
		// Config doesn't need environment - return project info only
		return result, nil
	}

	// Load environment if specified
	if envName != "" {
		// Validate environment exists
		if _, ok := project.Environments[envName]; !ok {
			availableEnvs := project.ListEnvironments()
			sort.Strings(availableEnvs)
			return nil, fmt.Errorf("environment '%s' not found in yapi.config.yml\nAvailable environments: %s",
				envName, strings.Join(availableEnvs, ", "))
		}

		// Load environment variables
		envVars, err := project.ResolveEnvFiles(projectRoot, envName)
		if err != nil {
			return nil, fmt.Errorf("failed to load environment '%s': %w", envName, err)
		}

		result.envVars = envVars
		result.envName = envName
	}

	return result, nil
}

func formatBytes(b int) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// isTerminal checks if the given file is a terminal (TTY)
func isTerminal(f *os.File) bool {
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// maxOutputSize is the threshold above which output is auto-saved to a file instead of printed.
// 1MB is a reasonable limit - terminals struggle with larger outputs.
const maxOutputSize = 1024 * 1024

// OutputResult holds the result of rendering output.
type OutputResult struct {
	Printed   bool   // Output was printed to terminal
	SavedPath string // If not empty, output was saved to this file
	SaveErr   error  // If not nil, save failed with this error
}

// renderOutput prints body to stdout, or saves to a file if too large.
// configPath is used to generate a meaningful filename for auto-saved files.
// When output is saved to file, no message is printed - caller handles that via OutputSavedError.
func renderOutput(body string, configPath string) OutputResult {
	if len(body) <= maxOutputSize {
		fmt.Println(strings.TrimRight(body, "\n\r"))
		return OutputResult{Printed: true}
	}

	// Output too large - save to file
	outputPath := generateOutputPath(configPath)
	if err := os.WriteFile(outputPath, []byte(body), 0600); err != nil {
		fmt.Fprintf(os.Stderr, "failed to save large output: %v\n", err)
		return OutputResult{SaveErr: err}
	}

	return OutputResult{SavedPath: outputPath}
}

// Logger provides leveled logging for verbose output.
type Logger struct {
	verbose bool
}

// NewLogger creates a logger. If verbose is false, Debug calls are no-ops.
func NewLogger(verbose bool) *Logger {
	return &Logger{verbose: verbose}
}

// Verbose prints a message only if verbose mode is enabled.
func (l *Logger) Verbose(format string, args ...any) {
	if !l.verbose {
		return
	}
	fmt.Fprintf(os.Stderr, "[VERBOSE] "+format+"\n", args...)
}

// Section prints a section header only if verbose mode is enabled.
func (l *Logger) Section(name string) {
	if !l.verbose {
		return
	}
	fmt.Fprintf(os.Stderr, "\n=== %s ===\n", name)
}

// Request logs request details in verbose mode.
func (l *Logger) Request(method, url string, headers map[string]string, body string) {
	if !l.verbose {
		return
	}
	l.Section("REQUEST")
	fmt.Fprintf(os.Stderr, "%s %s\n", method, url)

	if len(headers) > 0 {
		fmt.Fprintln(os.Stderr, "\nHeaders:")
		for k, v := range headers {
			fmt.Fprintf(os.Stderr, "  %s: %s\n", k, v)
		}
	}

	if body != "" {
		fmt.Fprintln(os.Stderr, "\nBody:")
		if len(body) > 1000 {
			fmt.Fprintf(os.Stderr, "  %s... (%d bytes total)\n", body[:1000], len(body))
		} else {
			fmt.Fprintf(os.Stderr, "  %s\n", body)
		}
	}
	fmt.Fprintln(os.Stderr, "")
}

// Response logs response details in verbose mode.
func (l *Logger) Response(statusCode int, headers map[string]string, duration time.Duration, bodySize int) {
	if !l.verbose {
		return
	}
	l.Section("RESPONSE")
	fmt.Fprintf(os.Stderr, "Status: %d\n", statusCode)
	fmt.Fprintf(os.Stderr, "Time: %s\n", duration)
	fmt.Fprintf(os.Stderr, "Size: %s\n", formatBytes(bodySize))

	if len(headers) > 0 {
		fmt.Fprintln(os.Stderr, "\nHeaders:")
		for k, v := range headers {
			fmt.Fprintf(os.Stderr, "  %s: %s\n", k, v)
		}
	}
	fmt.Fprintln(os.Stderr, "")
}

// generateOutputPath creates a filename like: config-name-output-20060102-150405.json
func generateOutputPath(configPath string) string {
	base := filepath.Base(configPath)
	name := strings.TrimSuffix(base, ".yapi.yml")
	if name == base {
		name = strings.TrimSuffix(base, filepath.Ext(base))
	}

	timestamp := time.Now().Format("20060102-150405")
	return fmt.Sprintf("%s-output-%s.json", name, timestamp)
}
