package commands

import (
	"github.com/spf13/cobra"
)

// CommandSpec defines the specification for a command.
type CommandSpec struct {
	Use     string
	Short   string
	Long    string
	Args    cobra.PositionalArgs
	Handler func(*cobra.Command, []string) error
	Flags   []FlagSpec
}

// FlagSpec defines a command flag.
type FlagSpec struct {
	Name      string
	Shorthand string
	Type      string // "bool", "string", etc.
	Default   interface{}
	Usage     string
}

// BuildCommand creates a cobra command from a spec.
func BuildCommand(spec CommandSpec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   spec.Use,
		Short: spec.Short,
		Args:  spec.Args,
		Run:   func(cmd *cobra.Command, args []string) {}, // no-op for doc generation
	}

	if spec.Long != "" {
		cmd.Long = spec.Long
	}

	if spec.Handler != nil {
		cmd.RunE = spec.Handler
	}

	for _, flag := range spec.Flags {
		switch flag.Type {
		case "bool":
			defaultVal := false
			if flag.Default != nil {
				defaultVal = flag.Default.(bool)
			}
			if flag.Shorthand != "" {
				cmd.Flags().BoolP(flag.Name, flag.Shorthand, defaultVal, flag.Usage)
			} else {
				cmd.Flags().Bool(flag.Name, defaultVal, flag.Usage)
			}
		}
	}

	return cmd
}
