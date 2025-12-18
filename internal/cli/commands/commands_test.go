package commands

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestNewTestCmd(t *testing.T) {
	tests := []struct {
		name         string
		handlers     *Handlers
		wantUse      string
		wantShort    string
		wantMaxArgs  int
		wantHasFlags bool
	}{
		{
			name:         "creates test command with nil handlers",
			handlers:     nil,
			wantUse:      "test [directory]",
			wantShort:    "Run all *.test.yapi.yml files in the current directory or specified directory",
			wantMaxArgs:  1,
			wantHasFlags: true,
		},
		{
			name: "creates test command with handlers",
			handlers: &Handlers{
				Test: func(cmd *cobra.Command, args []string) error {
					return nil
				},
			},
			wantUse:      "test [directory]",
			wantShort:    "Run all *.test.yapi.yml files in the current directory or specified directory",
			wantMaxArgs:  1,
			wantHasFlags: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newTestCmd(tt.handlers)

			if cmd.Use != tt.wantUse {
				t.Errorf("newTestCmd() Use = %v, want %v", cmd.Use, tt.wantUse)
			}

			if cmd.Short != tt.wantShort {
				t.Errorf("newTestCmd() Short = %v, want %v", cmd.Short, tt.wantShort)
			}

			// Check verbose flag exists
			verboseFlag := cmd.Flags().Lookup("verbose")
			if verboseFlag == nil {
				t.Error("newTestCmd() missing verbose flag")
				return
			}

			// Check verbose flag shorthand
			if verboseFlag.Shorthand != "v" {
				t.Errorf("newTestCmd() verbose flag shorthand = %v, want 'v'", verboseFlag.Shorthand)
			}
		})
	}
}

func TestBuildRoot(t *testing.T) {
	cfg := &Config{
		URLOverride:  "",
		NoColor:      false,
		BinaryOutput: false,
	}

	handlers := &Handlers{
		Run: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Version: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Validate: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Share: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Test: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	rootCmd := BuildRoot(cfg, handlers)

	if rootCmd == nil {
		t.Fatal("BuildRoot() returned nil")
	}

	// Check that test command is added
	testCmd := findCommandByName(rootCmd, "test")
	if testCmd == nil {
		t.Error("BuildRoot() did not add test command")
	}

	// Check that all expected commands exist
	expectedCommands := []string{"run", "version", "validate", "share", "test"}
	for _, cmdName := range expectedCommands {
		if findCommandByName(rootCmd, cmdName) == nil {
			t.Errorf("BuildRoot() missing command: %s", cmdName)
		}
	}
}

// Helper function to find a command by name
func findCommandByName(root *cobra.Command, name string) *cobra.Command {
	for _, cmd := range root.Commands() {
		if cmd.Name() == name {
			return cmd
		}
	}
	return nil
}
