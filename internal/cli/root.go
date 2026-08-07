// Package cli defines blackbelt's command-line interface.
package cli

import (
	"context"

	"github.com/pinglei-he/blackbelt/internal/blackbelt"
	"github.com/spf13/cobra"
)

// Execute runs blackbelt's root command.
func Execute(ctx context.Context) error {
	return newRootCommand().ExecuteContext(ctx)
}

func newRootCommand() *cobra.Command {
	var dryRun bool

	command := &cobra.Command{
		Use:           "blackbelt",
		Short:         "Make jj PR stacks obvious on GitHub",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			return blackbelt.Run(command.Context(), blackbelt.Options{DryRun: dryRun})
		},
	}
	command.Flags().BoolVar(
		&dryRun,
		"dry-run",
		false,
		"print one terminal diagram without changing GitHub comments",
	)
	return command
}
