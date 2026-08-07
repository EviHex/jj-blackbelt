// Package cli defines blackbelt's command-line interface.
package cli

import (
	"context"

	"github.com/pinglei-he/blackbelt/internal/config"
	"github.com/spf13/cobra"
)

// Execute runs blackbelt's root command.
func Execute(ctx context.Context) error {
	value, err := config.Load(ctx)
	if err != nil {
		return err
	}
	return newRootCommand(value).ExecuteContext(ctx)
}

func newRootCommand(value config.Config) *cobra.Command {
	command := &cobra.Command{
		Use:           "bb",
		Short:         "Make jj PR stacks obvious on GitHub",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			return command.Help()
		},
	}
	command.AddCommand(newStackCommand(value))
	return command
}
