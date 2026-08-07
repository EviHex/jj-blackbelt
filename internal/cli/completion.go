package cli

import (
	"github.com/pinglei-he/blackbelt/internal/completion"
	"github.com/spf13/cobra"
)

func addCompletionCommands(root *cobra.Command) {
	root.InitDefaultCompletionCmd()
	command, _, err := root.Find([]string{"completion"})
	if err != nil {
		return
	}
	command.AddCommand(&cobra.Command{
		Use:       "jj <bash|zsh|fish>",
		Short:     "Generate a completion bridge for the jj stack alias",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"bash", "zsh", "fish"},
		RunE: func(command *cobra.Command, args []string) error {
			return completion.WriteJJBridge(command.OutOrStdout(), args[0])
		},
	})
}
