package cli

import (
	"fmt"
	"strconv"

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
				return blackbelt.Run(command.Context(), blackbelt.Options{DryRun: true, All: value.Stack.Log.All, Revset: value.Stack.Log.Revset})
			case "draw":
				return blackbelt.Run(command.Context(), blackbelt.Options{Revset: value.Stack.Log.Revset})
			default:
				return fmt.Errorf("unknown stack.default-command %q", value.Stack.DefaultCommand)
			}
		},
	}
	command.AddCommand(
		newStackLogCommand(value), newStackDrawCommand(value), newStackOrderCommand(value),
		newStackNavigationCommand(value, "up"), newStackNavigationCommand(value, "down"),
		newStackNavigationCommand(value, "top"), newStackNavigationCommand(value, "bottom"),
		newStackGotoCommand(value),
	)
	return command
}

func newStackNavigationCommand(value config.Config, direction string) *cobra.Command {
	var dryRun, jsonOutput bool
	use := direction
	if direction == "up" || direction == "down" {
		use += " [n]"
	}
	argsValidator := cobra.NoArgs
	if direction == "up" || direction == "down" {
		argsValidator = cobra.MaximumNArgs(1)
	}
	command := &cobra.Command{
		Use:   use,
		Short: navigationDescription(direction),
		Args:  argsValidator,
		RunE: func(command *cobra.Command, args []string) error {
			steps := 1
			if len(args) == 1 {
				value, err := strconv.Atoi(args[0])
				if err != nil || value < 1 {
					return fmt.Errorf("n must be a positive integer")
				}
				steps = value
			}
			return blackbelt.Navigate(command.Context(), blackbelt.NavigateOptions{Direction: direction, Steps: steps, DryRun: dryRun, JSON: jsonOutput, Revset: value.Stack.Log.Revset})
		},
	}
	command.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "print the target without moving the jj working copy")
	command.Flags().BoolVar(&jsonOutput, "json", false, "write the target as JSON")
	return command
}

func newStackGotoCommand(value config.Config) *cobra.Command {
	var dryRun, jsonOutput bool
	command := &cobra.Command{
		Use:   "goto <PR-number|bookmark>",
		Short: "Move to a PR by number or bookmark",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			return blackbelt.Navigate(command.Context(), blackbelt.NavigateOptions{Direction: "goto", Target: args[0], DryRun: dryRun, JSON: jsonOutput, Revset: value.Stack.Log.Revset})
		},
	}
	command.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "print the target without moving the jj working copy")
	command.Flags().BoolVar(&jsonOutput, "json", false, "write the target as JSON")
	return command
}

func navigationDescription(direction string) string {
	switch direction {
	case "up":
		return "Move to a child PR above the current PR"
	case "down":
		return "Move to the parent PR below the current PR"
	case "top":
		return "Move to a topmost PR in the current stack"
	default:
		return "Move to the bottom PR in the current stack"
	}
}

func newStackLogCommand(value config.Config) *cobra.Command {
	all := value.Stack.Log.All
	jsonOutput := false
	revset := value.Stack.Log.Revset
	command := &cobra.Command{
		Use:   "log",
		Short: "Show the PR tree rooted below trunk around the current change",
		Long: "Show the complete tracked PR tree rooted below trunk on the path to the current change. " +
			"Parallel descendants are included. Use --all to show every tracked PR stack in the repository.",
		Args: cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			return blackbelt.Run(command.Context(), blackbelt.Options{DryRun: true, All: all, JSON: jsonOutput, Revset: revset})
		},
	}
	command.Flags().BoolVarP(&all, "all", "a", all, "show all tracked PR stacks")
	command.Flags().BoolVar(&jsonOutput, "json", false, "write the resolved stack as JSON")
	command.Flags().StringVarP(&revset, "revisions", "r", revset, "use revisions matching the jj revset as stack seeds")
	return command
}

func newStackDrawCommand(value config.Config) *cobra.Command {
	return &cobra.Command{
		Use:     "draw",
		Aliases: []string{"diagram", "d"},
		Short:   "Create or update the stack diagram on every PR",
		Args:    cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			return blackbelt.Run(command.Context(), blackbelt.Options{Revset: value.Stack.Log.Revset})
		},
	}
}

func newStackOrderCommand(value config.Config) *cobra.Command {
	all := false
	fix := false
	jsonOutput := false
	command := &cobra.Command{
		Use:   "order",
		Short: "Check whether GitHub PR bases match the jj stack",
		Long: "Check every open PR's GitHub base against its nearest unmerged parent in the jj stack. " +
			"The command is read-only unless --fix is supplied.",
		Args: cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			return blackbelt.Order(command.Context(), blackbelt.OrderOptions{
				All: all, Fix: fix, JSON: jsonOutput, Revset: value.Stack.Log.Revset,
			})
		},
	}
	command.Flags().BoolVarP(&all, "all", "a", all, "check all tracked PR stacks")
	command.Flags().BoolVar(&fix, "fix", fix, "retarget incorrect PR bases on GitHub")
	command.Flags().BoolVar(&jsonOutput, "json", jsonOutput, "write diagnostics as JSON")
	return command
}
