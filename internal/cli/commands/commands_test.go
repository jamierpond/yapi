package commands

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestTestCmdFromManifest(t *testing.T) {
	// Find the test command spec in the manifest
	var testSpec *CommandSpec
	for i, spec := range cmdManifest {
		if spec.Use == "test [directory]" {
			testSpec = &cmdManifest[i]
			break
		}
	}

	if testSpec == nil {
		t.Fatal("test command not found in manifest")
	}

	// Test command properties from manifest
	if testSpec.Short != "Run all *.test.yapi.yml files in the current directory or specified directory" {
		t.Errorf("test command Short = %v, want expected description", testSpec.Short)
	}

	// Test building the command with handlers
	handlers := &Handlers{
		Test: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	testSpec.Handler = getHandler(handlers, testSpec.Use)
	cmd := BuildCommand(*testSpec)

	if cmd.Use != "test [directory]" {
		t.Errorf("built command Use = %v, want 'test [directory]'", cmd.Use)
	}

	// Check verbose flag exists
	verboseFlag := cmd.Flags().Lookup("verbose")
	if verboseFlag == nil {
		t.Error("test command missing verbose flag")
		return
	}

	// Check verbose flag shorthand
	if verboseFlag.Shorthand != "v" {
		t.Errorf("verbose flag shorthand = %v, want 'v'", verboseFlag.Shorthand)
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
