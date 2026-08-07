package blackbelt

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type command struct {
	output string
	err    error
}
type runner interface {
	run(context.Context, string, ...string) command
}
type shell struct{}

func (shell) run(ctx context.Context, name string, args ...string) command {
	c := exec.CommandContext(ctx, name, args...)
	b, e := c.Output()
	if e != nil {
		if errors.Is(e, exec.ErrNotFound) {
			return command{"", executableNotFound(name)}
		}
		if x := new(exec.ExitError); errors.As(e, &x) {
			return command{"", commandFailure(name, args, x.Stderr)}
		}
		return command{"", e}
	}
	return command{string(b), nil}
}

func commandFailure(name string, args []string, stderr []byte) error {
	message := strings.TrimSpace(string(stderr))
	prefix := name + ":"
	for strings.HasPrefix(message, prefix) {
		message = strings.TrimSpace(strings.TrimPrefix(message, prefix))
	}
	operation := name + " command"
	if name == "gh" {
		operation = "GitHub CLI command"
		if len(args) > 0 && args[0] == "api" {
			operation = "GitHub API request"
		}
	} else if name == "jj" {
		operation = "jj command"
	}
	if message == "" {
		return fmt.Errorf("%s failed", operation)
	}
	return fmt.Errorf("%s failed: %s", operation, message)
}

func executableNotFound(name string) error {
	label := fmt.Sprintf("required executable %q", name)
	if name == "gh" {
		label = `GitHub CLI executable "gh"`
	} else if name == "jj" {
		label = `jj executable "jj"`
	}
	return fmt.Errorf("%s was not found; run bb doctor", label)
}
