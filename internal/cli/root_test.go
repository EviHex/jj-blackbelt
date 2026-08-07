package cli

import (
	"strings"
	"testing"

	"github.com/EviHex/jj-blackbelt/internal/config"
)

func TestRootCommandHelp(t *testing.T) {
	command := newRootCommand(config.Defaults())
	var output strings.Builder
	command.SetOut(&output)
	command.SetArgs([]string{"--help"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	for _, expected := range []string{
		"Usage: bb <command> [flags]",
		"stack (s) draw (diagram, d)",
		"Run the configured default (currently log)",
		"bb s d     => bb stack draw",
		"completion jj <shell>",
	} {
		if !strings.Contains(output.String(), expected) {
			t.Errorf("help output missing %q:\n%s", expected, output.String())
		}
	}
	if strings.Contains(output.String(), "Available Commands:") {
		t.Errorf("root help used Cobra's generic command list:\n%s", output.String())
	}
}

func TestRootHelpShowsConfiguredDefaultCommand(t *testing.T) {
	value := config.Defaults()
	value.Stack.DefaultCommand = "draw"
	command := newRootCommand(value)
	var output strings.Builder
	command.SetOut(&output)
	command.SetArgs([]string{"--help"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if want := "Run the configured default (currently draw)"; !strings.Contains(output.String(), want) {
		t.Errorf("help output missing %q:\n%s", want, output.String())
	}
}

func TestSubcommandHelpKeepsCommandSpecificFlags(t *testing.T) {
	command := newRootCommand(config.Defaults())
	var output strings.Builder
	command.SetOut(&output)
	command.SetArgs([]string{"stack", "log", "--help"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	for _, want := range []string{"Usage:", "bb stack log", "--revisions"} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("subcommand help missing %q:\n%s", want, output.String())
		}
	}
}

func TestRootCommandRejectsArguments(t *testing.T) {
	command := newRootCommand(config.Defaults())
	command.SetArgs([]string{"unexpected"})

	if err := command.Execute(); err == nil {
		t.Fatal("Execute() unexpectedly succeeded")
	}
}

func TestStackCommandAliases(t *testing.T) {
	root := newRootCommand(config.Defaults())
	stack, _, err := root.Find([]string{"s"})
	if err != nil || stack.Name() != "stack" {
		t.Fatalf("find s = %v, %v", stack, err)
	}
	draw, _, err := root.Find([]string{"s", "d"})
	if err != nil || draw.Name() != "draw" {
		t.Fatalf("find s d = %v, %v", draw, err)
	}
}
