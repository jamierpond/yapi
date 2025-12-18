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

	rootCmd.AddCommand(newRunCmd(handlers))
	rootCmd.AddCommand(newWatchCmd(handlers))
	rootCmd.AddCommand(newHistoryCmd(handlers))
	rootCmd.AddCommand(newLSPCmd(handlers))
	rootCmd.AddCommand(newVersionCmd(handlers))
	rootCmd.AddCommand(newValidateCmd(handlers))
	rootCmd.AddCommand(newShareCmd(handlers))
	rootCmd.AddCommand(newTestCmd(handlers))

	return rootCmd
}

func newRunCmd(h *Handlers) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [file]",
		Short: "Run a request defined in a yapi config file (reads from stdin if no file specified)",
		Args:  cobra.MaximumNArgs(1),
		Run:   func(cmd *cobra.Command, args []string) {}, // no-op for doc generation
	}
	if h != nil && h.Run != nil {
		cmd.RunE = h.Run
	}
	return cmd
}

func newWatchCmd(h *Handlers) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "watch [file]",
		Short: "Watch a yapi config file and re-run on changes",
		Args:  cobra.MaximumNArgs(1),
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	if h != nil && h.Watch != nil {
		cmd.RunE = h.Watch
	}

	cmd.Flags().BoolP("pretty", "p", false, "Enable pretty TUI mode")
	cmd.Flags().Bool("no-pretty", false, "Disable pretty TUI mode")

	return cmd
}

func newHistoryCmd(h *Handlers) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history [count]",
		Short: "Show yapi command history (default: last 10)",
		Args:  cobra.MaximumNArgs(1),
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	if h != nil && h.History != nil {
		cmd.RunE = h.History
	}

	cmd.Flags().Bool("json", false, "Output as JSON")

	return cmd
}

func newLSPCmd(h *Handlers) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lsp",
		Short: "Run the yapi language server over stdio",
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	if h != nil && h.LSP != nil {
		cmd.RunE = h.LSP
	}
	return cmd
}

func newVersionCmd(h *Handlers) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	if h != nil && h.Version != nil {
		cmd.RunE = h.Version
	}

	cmd.Flags().Bool("json", false, "Output version info as JSON")

	return cmd
}

func newValidateCmd(h *Handlers) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate [file]",
		Short: "Validate a yapi config file",
		Long:  "Validate a yapi config file and report diagnostics. Use - to read from stdin.",
		Args:  cobra.MaximumNArgs(1),
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	if h != nil && h.Validate != nil {
		cmd.RunE = h.Validate
	}

	cmd.Flags().Bool("json", false, "Output diagnostics as JSON")

	return cmd
}

func newShareCmd(h *Handlers) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "share <file>",
		Short: "Generate a shareable yapi.run link for a config file",
		Args:  cobra.ExactArgs(1),
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	if h != nil && h.Share != nil {
		cmd.RunE = h.Share
	}

	return cmd
}

func newTestCmd(h *Handlers) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test [directory]",
		Short: "Run all *.test.yapi.yml files in the current directory or specified directory",
		Args:  cobra.MaximumNArgs(1),
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	if h != nil && h.Test != nil {
		cmd.RunE = h.Test
	}

	cmd.Flags().BoolP("verbose", "v", false, "Show verbose output for each test")

	return cmd
}
