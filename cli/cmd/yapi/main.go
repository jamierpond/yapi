package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
	"yapi.run/cli/internal/briefing"
	"yapi.run/cli/internal/cli/color"
	"yapi.run/cli/internal/cli/commands"
	"yapi.run/cli/internal/cli/middleware"
	"yapi.run/cli/internal/core"
	"yapi.run/cli/internal/langserver"
	"yapi.run/cli/internal/observability"
)

// Set via ldflags at build time
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func init() {
	if version != "dev" {
		return
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		version = info.Main.Version
	}
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			if len(s.Value) >= 7 {
				commit = s.Value[:7]
			}
		case "vcs.time":
			date = s.Value
		}
	}
}

type rootCommand struct {
	urlOverride  string
	noColor      bool
	binaryOutput bool
	insecure     bool
	httpClient   *http.Client
	engine       *core.Engine
}

// io returns the appropriate writer and color flag based on strict mode
func (app *rootCommand) io(strict bool) (io.Writer, bool) {
	if strict {
		return os.Stderr, app.noColor
	}
	return os.Stdout, app.noColor
}

func main() {
	observability.Init(version, commit)
	defer observability.Close()

	// Wire observability hook - main.go is the composition root
	requestHook := func(stats map[string]any) {
		observability.Track("request_executed", stats)
	}

	httpClient := &http.Client{}
	app := &rootCommand{
		httpClient: httpClient,
		engine:     core.NewEngine(httpClient, core.WithRequestHook(requestHook)),
	}

	cfg := &commands.Config{}
	handlers := &commands.Handlers{
		RunInteractive: app.runInteractiveE,
		Run:            app.runE,
		Watch:          app.watchE,
		History:        historyE,
		LSP:            lspE,
		Version:        versionE,
		Validate:       validateE,
		Share:          shareE,
		Test:           app.testE,
		List:           listE,
		Stress:         app.stressE,
		About:          aboutE,
		Import:         importE,
		Send:           app.sendE,
		Docs:           docsE,
	}

	rootCmd := commands.BuildRoot(cfg, handlers)

	// Wire up the config to app after flags are parsed
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		app.urlOverride = cfg.URLOverride
		app.noColor = cfg.NoColor
		app.binaryOutput = cfg.BinaryOutput
		app.insecure = cfg.Insecure
		color.SetNoColor(app.noColor)
	}
	rootCmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		// Log command to history (skip meta commands)
		switch cmd.Name() {
		case "history", "version", "lsp", "help", "yapi", "about", "docs":
			return
		}
		logHistoryCmd(reconstructCommand(cmd, args))
	}

	// Wrap all commands with observability middleware
	middleware.WrapWithObservability(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, color.Red(err.Error()))
		observability.Close() // Ensure cleanup before exit
		os.Exit(1)            //nolint:gocritic // exitAfterDefer: Close() called explicitly above
	}
}

func lspE(cmd *cobra.Command, args []string) error {
	langserver.Run()
	return nil
}

func versionE(cmd *cobra.Command, args []string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")

	if jsonOutput {
		info := map[string]any{
			"version": version,
			"commit":  commit,
			"date":    date,
		}
		return json.NewEncoder(os.Stdout).Encode(info)
	}

	fmt.Printf("yapi %s\n", version)
	fmt.Printf("  commit: %s\n", commit)
	fmt.Printf("  built:  %s\n", date)
	return nil
}

func aboutE(cmd *cobra.Command, args []string) error {
	fmt.Print(briefing.Content)
	return nil
}
