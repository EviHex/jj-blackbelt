// Package cli defines blackbelt's command-line interface.
package cli

import (
	"context"

	"github.com/pinglei-he/blackbelt/internal/config"
	"github.com/pinglei-he/blackbelt/internal/doctor"
	"github.com/pinglei-he/blackbelt/internal/version"
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
		Version:       version.Current,
		RunE: func(command *cobra.Command, _ []string) error {
			return command.Help()
		},
	}
	command.AddCommand(newStackCommand(value), newDoctorCommand(), newUtilCommand())
	return command
}

func newDoctorCommand() *cobra.Command {
	var jsonOutput bool
	command := &cobra.Command{
		Use:   "doctor",
		Short: "Check blackbelt's jj and GitHub prerequisites",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			return doctor.Run(command.Context(), jsonOutput)
		},
	}
	command.Flags().BoolVar(&jsonOutput, "json", false, "write check results as JSON")
	return command
}
