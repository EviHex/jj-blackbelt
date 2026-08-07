package cli

import (
	"github.com/pinglei-he/blackbelt/internal/jjalias"
	"github.com/spf13/cobra"
)

func newUtilCommand() *cobra.Command {
	command := &cobra.Command{Use: "util", Short: "Installation and integration utilities"}
	command.AddCommand(newAliasCommand())
	return command
}

func newAliasCommand() *cobra.Command {
	var dryRun bool
	command := &cobra.Command{
		Use:   "alias",
		Short: "Install the jj stack alias for bb stack",
		Long:  "Install an idempotent user-level jj alias so that `jj stack ...` delegates to `bb stack ...`.",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			return jjalias.Install(command.Context(), dryRun)
		},
	}
	command.Flags().BoolVar(&dryRun, "dry-run", false, "show the alias without changing jj configuration")
	return command
}
