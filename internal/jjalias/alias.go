// Package jjalias installs blackbelt's jj command alias.
package jjalias

import (
	"context"
	"fmt"
	"os/exec"
)

const stackAlias = `["util", "exec", "--", "bb", "stack"]`

// Install configures `jj stack` to delegate to `bb stack`.
func Install(ctx context.Context, dryRun bool) error {
	if _, err := exec.LookPath("bb"); err != nil && !dryRun {
		return fmt.Errorf("bb is not available on PATH")
	}
	if dryRun {
		fmt.Printf("Would set user jj alias: stack = %s\n", stackAlias)
		return nil
	}
	command := exec.CommandContext(ctx, "jj", "config", "set", "--user", "aliases.stack", stackAlias)
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("install jj alias: %s", string(output))
	}
	fmt.Println("✓ Installed jj stack → bb stack")
	return nil
}
