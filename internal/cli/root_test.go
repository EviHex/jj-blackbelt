package cli

import (
	"strings"
	"testing"
)

func TestRootCommandHelp(t *testing.T) {
	command := newRootCommand()
	var output strings.Builder
	command.SetOut(&output)
	command.SetArgs([]string{"--help"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	for _, expected := range []string{"Make jj PR stacks obvious on GitHub", "--dry-run"} {
		if !strings.Contains(output.String(), expected) {
			t.Errorf("help output missing %q:\n%s", expected, output.String())
		}
	}
}

func TestRootCommandRejectsArguments(t *testing.T) {
	command := newRootCommand()
	command.SetArgs([]string{"unexpected"})

	if err := command.Execute(); err == nil {
		t.Fatal("Execute() unexpectedly succeeded")
	}
}
