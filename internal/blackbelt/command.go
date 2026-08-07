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
		if x := new(exec.ExitError); errors.As(e, &x) {
			return command{"", fmt.Errorf("%s: %s", name, strings.TrimSpace(string(x.Stderr)))}
		}
		return command{"", e}
	}
	return command{string(b), nil}
}
