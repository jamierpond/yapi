// Package commands defines the CLI command structure for the yapi application.
package commands

import (
	"github.com/spf13/cobra"
)

// Config holds configuration for command execution
type Config struct {
	URLOverride  string
	NoColor      bool
	BinaryOutput bool
	Environment  string // Target environment from project config
}

// Handlers contains the callback functions for command execution
type Handlers struct {
	RunInteractive func(cmd *cobra.Command, args []string) error
	Run            func(cmd *cobra.Command, args []string) error
	Watch          func(cmd *cobra.Command, args []string) error
	History        func(cmd *cobra.Command, args []string) error
	LSP            func(cmd *cobra.Command, args []string) error
	Version        func(cmd *cobra.Command, args []string) error
	Validate       func(cmd *cobra.Command, args []string) error
	Share          func(cmd *cobra.Command, args []string) error
	Test           func(cmd *cobra.Command, args []string) error
	List           func(cmd *cobra.Command, args []string) error
	Stress         func(cmd *cobra.Command, args []string) error
}

// BuildRoot builds the root command tree with optional handlers.
// If handlers is nil, commands are built without RunE functions (for doc generation).
func BuildRoot(cfg *Config, handlers *Handlers) *cobra.Command {
	if cfg == nil {
		cfg = &Config{}
	}

	rootCmd := &cobra.Command{
		Use:           "yapi",
		Short:         "yapi is a unified API client for HTTP, gRPC, and TCP",
		SilenceUsage:  true,
		SilenceErrors: true,
		Run:           func(cmd *cobra.Command, args []string) {},
	}

	if handlers != nil && handlers.RunInteractive != nil {
		rootCmd.RunE = handlers.RunInteractive
	}

	rootCmd.PersistentFlags().StringVarP(&cfg.URLOverride, "url", "u", "", "Override the URL specified in the config file")
	rootCmd.PersistentFlags().BoolVar(&cfg.NoColor, "no-color", false, "Disable color output")
	rootCmd.PersistentFlags().BoolVar(&cfg.BinaryOutput, "binary-output", false, "Display binary content to stdout (by default binary content is hidden)")

	// Build commands from manifest
	for _, spec := range cmdManifest {
		spec.Handler = getHandler(handlers, spec.Use)
		rootCmd.AddCommand(BuildCommand(spec))
	}

	return rootCmd
}

// cmdManifest defines all CLI commands as declarative data
var cmdManifest = []CommandSpec{
	{
		Use:   "run [file]",
		Short: "Run a request defined in a yapi config file (reads from stdin if no file specified)",
		Args:  cobra.MaximumNArgs(1),
		Flags: []FlagSpec{
			{Name: "env", Shorthand: "e", Type: "string", Default: "", Usage: "Target environment from yapi.config.yml"},
		},
	},
	{
		Use:   "watch [file]",
		Short: "Watch a yapi config file and re-run on changes",
		Args:  cobra.MaximumNArgs(1),
		Flags: []FlagSpec{
			{Name: "pretty", Shorthand: "p", Type: "bool", Default: false, Usage: "Enable pretty TUI mode"},
			{Name: "no-pretty", Type: "bool", Default: false, Usage: "Disable pretty TUI mode"},
			{Name: "env", Shorthand: "e", Type: "string", Default: "", Usage: "Target environment from yapi.config.yml"},
		},
	},
	{
		Use:   "history [count]",
		Short: "Show yapi command history (default: last 10)",
		Args:  cobra.MaximumNArgs(1),
		Flags: []FlagSpec{
			{Name: "json", Type: "bool", Default: false, Usage: "Output as JSON"},
		},
	},
	{
		Use:   "lsp",
		Short: "Run the yapi language server over stdio",
	},
	{
		Use:   "version",
		Short: "Print version information",
		Flags: []FlagSpec{
			{Name: "json", Type: "bool", Default: false, Usage: "Output version info as JSON"},
		},
	},
	{
		Use:   "validate [file]",
		Short: "Validate a yapi config file",
		Long:  "Validate a yapi config file and report diagnostics. Use - to read from stdin.",
		Args:  cobra.MaximumNArgs(1),
		Flags: []FlagSpec{
			{Name: "json", Type: "bool", Default: false, Usage: "Output diagnostics as JSON"},
		},
	},
	{
		Use:   "share [file]",
		Short: "Generate a shareable yapi.run link for a config file",
		Args:  cobra.MaximumNArgs(1),
	},
	{
		Use:   "test [directory]",
		Short: "Run all *.test.yapi.yml files in the current directory or specified directory",
		Args:  cobra.MaximumNArgs(1),
		Flags: []FlagSpec{
			{Name: "all", Shorthand: "a", Type: "bool", Default: false, Usage: "Run all *.yapi.yml files (not just *.test.yapi.yml)"},
			{Name: "verbose", Shorthand: "v", Type: "bool", Default: false, Usage: "Show verbose output for each test"},
			{Name: "env", Shorthand: "e", Type: "string", Default: "", Usage: "Target environment from yapi.config.yml"},
			{Name: "parallel", Shorthand: "p", Type: "int", Default: 1, Usage: "Number of parallel threads to run tests on"},
		},
	},
	{
		Use:     "list [directory]",
		Aliases: []string{"ls"},
		Short:   "List all yapi config files in the current directory or project",
		Args:    cobra.MaximumNArgs(1),
		Flags: []FlagSpec{
			{Name: "json", Type: "bool", Default: false, Usage: "Output as JSON"},
		},
	},
	{
		Use:     "stress [file]",
		Aliases: []string{"pwn"},
		Short:   "Load test a yapi config file with concurrent requests",
		Args:    cobra.MaximumNArgs(1),
		Flags: []FlagSpec{
			{Name: "parallel", Shorthand: "p", Type: "int", Default: 1, Usage: "Number of concurrent requests"},
			{Name: "num-requests", Shorthand: "n", Type: "int", Default: 100, Usage: "Total number of requests to make"},
			{Name: "duration", Shorthand: "d", Type: "string", Default: "", Usage: "Duration to run test (e.g., 10s, 1m) - overrides num-requests"},
			{Name: "env", Shorthand: "e", Type: "string", Default: "", Usage: "Target environment from yapi.config.yml"},
			{Name: "yes", Shorthand: "y", Type: "bool", Default: false, Usage: "Skip confirmation prompt"},
		},
	},
}

// getHandler maps command names to handlers
func getHandler(h *Handlers, use string) func(*cobra.Command, []string) error {
	if h == nil {
		return nil
	}
	// Extract command name from "use" string (e.g., "run [file]" -> "run")
	cmdName := use
	if idx := len(use); idx > 0 {
		for i, r := range use {
			if r == ' ' || r == '[' {
				cmdName = use[:i]
				break
			}
		}
	}

	switch cmdName {
	case "run":
		return h.Run
	case "watch":
		return h.Watch
	case "history":
		return h.History
	case "lsp":
		return h.LSP
	case "version":
		return h.Version
	case "validate":
		return h.Validate
	case "share":
		return h.Share
	case "test":
		return h.Test
	case "list":
		return h.List
	case "stress":
		return h.Stress
	default:
		return nil
	}
}
