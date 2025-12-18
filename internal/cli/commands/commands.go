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
	rootCmd.PersistentFlags().BoolVar(&cfg.BinaryOutput, "binary-output", false, "Display binary content to stdout (by default binary content is hidden when outputting to a terminal)")

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
	},
	{
		Use:   "watch [file]",
		Short: "Watch a yapi config file and re-run on changes",
		Args:  cobra.MaximumNArgs(1),
		Flags: []FlagSpec{
			{Name: "pretty", Shorthand: "p", Type: "bool", Default: false, Usage: "Enable pretty TUI mode"},
			{Name: "no-pretty", Type: "bool", Default: false, Usage: "Disable pretty TUI mode"},
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
			{Name: "verbose", Shorthand: "v", Type: "bool", Default: false, Usage: "Show verbose output for each test"},
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
	default:
		return nil
	}
}
