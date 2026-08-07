package cli

import (
	"strings"
	"testing"

	"github.com/pinglei-he/blackbelt/internal/config"
)

func TestRootCommandHelp(t *testing.T) {
	command := newRootCommand(config.Defaults())
	var output strings.Builder
	command.SetOut(&output)
	command.SetArgs([]string{"--help"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	for _, expected := range []string{"Make jj PR stacks obvious on GitHub", "stack"} {
		if !strings.Contains(output.String(), expected) {
			t.Errorf("help output missing %q:\n%s", expected, output.String())
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
