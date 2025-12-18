package commands

import (
	"github.com/spf13/cobra"
)

// CommandSpec defines the specification for a command.
type CommandSpec struct {
	Use     string
	Aliases []string
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
		Use:     spec.Use,
		Aliases: spec.Aliases,
		Short:   spec.Short,
		Args:    spec.Args,
		Run:     func(cmd *cobra.Command, args []string) {}, // no-op for doc generation
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
		case "string":
			defaultVal := ""
			if flag.Default != nil {
				defaultVal = flag.Default.(string)
			}
			if flag.Shorthand != "" {
				cmd.Flags().StringP(flag.Name, flag.Shorthand, defaultVal, flag.Usage)
			} else {
				cmd.Flags().String(flag.Name, defaultVal, flag.Usage)
			}
		case "int":
			defaultVal := 0
			if flag.Default != nil {
				defaultVal = flag.Default.(int)
			}
			if flag.Shorthand != "" {
				cmd.Flags().IntP(flag.Name, flag.Shorthand, defaultVal, flag.Usage)
			} else {
				cmd.Flags().Int(flag.Name, defaultVal, flag.Usage)
			}
		}
	}

	return cmd
}
