package cli

import (
	"fmt"

	"github.com/pinglei-he/blackbelt/internal/blackbelt"
	"github.com/pinglei-he/blackbelt/internal/config"
	"github.com/spf13/cobra"
)

func newStackCommand(value config.Config) *cobra.Command {
	command := &cobra.Command{
		Use:     "stack",
		Aliases: []string{"s"},
		Short:   "Inspect and maintain the current PR stack",
		Args:    cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			switch value.Stack.DefaultCommand {
			case "log":
				return blackbelt.Run(command.Context(), blackbelt.Options{DryRun: true})
			case "draw":
				return blackbelt.Run(command.Context(), blackbelt.Options{})
			default:
				return fmt.Errorf("unknown stack.default-command %q", value.Stack.DefaultCommand)
			}
		},
	}
	command.AddCommand(newStackLogCommand(), newStackDrawCommand())
	return command
}

func newStackLogCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "log",
		Short: "Show the PR tree rooted below trunk around the current change",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			return blackbelt.Run(command.Context(), blackbelt.Options{DryRun: true})
		},
	}
}

func newStackDrawCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "draw",
		Aliases: []string{"diagram", "d"},
		Short:   "Create or update the stack diagram on every PR",
		Args:    cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			return blackbelt.Run(command.Context(), blackbelt.Options{})
		},
	}
}
