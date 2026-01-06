package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"yapi.run/cli/internal/cli/color"
	"yapi.run/cli/internal/observability"
)

type historyEntry struct {
	Timestamp string   `json:"timestamp"`
	Event     string   `json:"event,omitempty"`  // legacy single event
	Events    []string `json:"events,omitempty"` // new merged events
	Command   string   `json:"command,omitempty"`
	FromTUI   bool     `json:"from_tui,omitempty"`
	// Fields from request tracking
	OS      string         `json:"os,omitempty"`
	Arch    string         `json:"arch,omitempty"`
	Version string         `json:"version,omitempty"`
	Commit  string         `json:"commit,omitempty"`
	Props   map[string]any `json:"-"` // For parsing additional fields
}

func historyE(cmd *cobra.Command, args []string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")

	count := 10
	if len(args) == 1 {
		n, err := fmt.Sscanf(args[0], "%d", &count)
		if err != nil || n != 1 || count < 1 {
			return fmt.Errorf("invalid count: %s", args[0])
		}
	}

	data, err := os.ReadFile(observability.HistoryFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No history yet")
			return nil
		}
		return fmt.Errorf("failed to read history: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		fmt.Println("No history yet")
		return nil
	}

	start := len(lines) - count
	if start < 0 {
		start = 0
	}

	entries := lines[start:]

	if jsonOutput {
		fmt.Println("[")
		for i, line := range entries {
			fmt.Print("  " + line)
			if i < len(entries)-1 {
				fmt.Println(",")
			} else {
				fmt.Println()
			}
		}
		fmt.Println("]")
		return nil
	}

	// Pretty print for humans
	for _, line := range entries {
		var entry historyEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		t, _ := time.Parse(time.RFC3339, entry.Timestamp)
		timeStr := color.Dim(t.Format("2006-01-02 15:04:05"))

		// New merged format has Command field directly
		if entry.Command != "" {
			fmt.Printf("%s  %s\n", timeStr, entry.Command)
			continue
		}

		// Legacy: request_executed entries had method/url
		if entry.Event == "request_executed" {
			var raw map[string]any
			if err := json.Unmarshal([]byte(line), &raw); err == nil {
				method, _ := raw["method"].(string)
				url, _ := raw["url"].(string)
				status, _ := raw["status_code"].(float64)
				if method != "" && url != "" {
					fmt.Printf("%s  %s %s %s\n", timeStr, color.Cyan(method), url, color.Dim(fmt.Sprintf("[%d]", int(status))))
					continue
				}
			}
		}
	}
	return nil
}

// logHistoryCmd writes a command to history as JSON
func logHistoryCmd(cmdStr string) {
	logHistoryEntry(cmdStr, false)
}

// logHistoryFromTUI writes a TUI-selected command to history
func logHistoryFromTUI(cmdStr string) {
	logHistoryEntry(cmdStr, true)
}

func logHistoryEntry(cmdStr string, fromTUI bool) {
	props := map[string]any{
		"command": cmdStr,
	}
	if fromTUI {
		props["from_tui"] = true
	}
	observability.Track("command", props)
}

// reconstructCommand builds the full command string from cobra command and args
func reconstructCommand(cmd *cobra.Command, args []string) string {
	parts := []string{"yapi", cmd.Name()}

	// Add flags that were set
	cmd.Flags().Visit(func(f *pflag.Flag) {
		if f.Value.Type() == "bool" {
			parts = append(parts, "--"+f.Name)
		} else {
			parts = append(parts, fmt.Sprintf("--%s=%q", f.Name, f.Value.String()))
		}
	})

	// Add args (quote paths)
	for _, arg := range args {
		absPath, err := filepath.Abs(arg)
		if err == nil && fileExists(absPath) {
			parts = append(parts, fmt.Sprintf("%q", absPath))
		} else {
			parts = append(parts, arg)
		}
	}

	return strings.Join(parts, " ")
}
