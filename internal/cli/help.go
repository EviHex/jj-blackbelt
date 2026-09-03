package cli

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/EviHex/jj-blackbelt/internal/config"
	"github.com/spf13/cobra"
)

func installRootHelp(root *cobra.Command, value config.Config) {
	defaultHelp := root.HelpFunc()
	root.SetHelpFunc(func(command *cobra.Command, args []string) {
		if command != root {
			defaultHelp(command, args)

			return
		}

		writeRootHelp(command.OutOrStdout(), value)
	})
}

func writeRootHelp(output io.Writer, value config.Config) {
	fmt.Fprintln(output, "Usage: bb <command> [flags]")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Make jj PR stacks obvious on GitHub.")
	fmt.Fprintln(output, "Keep jj as your workflow; use blackbelt for PR-aware views, diagrams, and diagnostics.")
	fmt.Fprintln(output)

	writer := tabwriter.NewWriter(output, 0, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "Stack")
	fmt.Fprintf(writer, "  stack (s)\tRun the configured default (currently %s)\n", value.Stack.DefaultCommand)
	fmt.Fprintln(writer, "  stack (s) log\tShow the PR tree around the current change")
	fmt.Fprintln(writer, "  stack (s) draw (diagram, d)\tCreate or update the diagram on every PR")
	fmt.Fprintln(writer, "  stack (s) order\tCheck GitHub PR bases against the jj stack")
	fmt.Fprintln(writer)
	fmt.Fprintln(writer)
	fmt.Fprintln(writer, "Setup")
	fmt.Fprintln(writer, "  doctor\tCheck jj, GitHub CLI, authentication, and repository access")
	fmt.Fprintln(writer, "  util alias\tInstall jj stack as an alias for bb stack")
	fmt.Fprintln(writer, "  completion <shell>\tGenerate Bash, Zsh, Fish, or PowerShell completion")
	fmt.Fprintln(writer, "  completion jj <shell>\tGenerate completion for the jj stack alias")
	_ = writer.Flush()

	fmt.Fprintln(output)
	fmt.Fprintln(output, "Global Flags:")
	fmt.Fprintln(output, "  -h, --help      Show help for the command")
	fmt.Fprintln(output, "  -v, --version   Print version information and quit")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Shorthand examples:")
	fmt.Fprintln(output, "  bb s       => bb stack")
	fmt.Fprintln(output, "  bb s d     => bb stack draw")
	fmt.Fprintln(output, "  jj stack d => bb stack draw  (after bb util alias)")
	fmt.Fprintln(output)
	fmt.Fprintln(output, `Run "bb <command> --help" for command-specific flags and examples.`)
}
