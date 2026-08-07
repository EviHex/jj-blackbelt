package blackbelt

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestShellRunCommandFailureKeepsStderrWithoutDuplicatePrefix(t *testing.T) {
	directory := t.TempDir()
	program := filepath.Join(directory, "gh")
	if err := os.WriteFile(program, []byte("#!/bin/sh\nprintf '\\n  gh: API rate limit exceeded\\nTry again later.\\n' >&2\nexit 1\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", directory)

	got := (shell{}).run(context.Background(), "gh")
	if got.err == nil {
		t.Fatal("run() error = nil, want command failure")
	}
	if want := "GitHub CLI command failed: API rate limit exceeded\nTry again later."; got.err.Error() != want {
		t.Errorf("run() error = %q, want %q", got.err, want)
	}
}

func TestShellRunReportsMissingExecutable(t *testing.T) {
	const name = "blackbelt-command-test-missing"

	got := (shell{}).run(context.Background(), name)
	if got.err == nil {
		t.Fatal("run() error = nil, want executable-not-found error")
	}
	if want := `required executable "` + name + `" was not found; run bb doctor`; got.err.Error() != want {
		t.Errorf("run() error = %q, want %q", got.err, want)
	}
}

func TestCommandFailureNamesGitHubAPIRequests(t *testing.T) {
	err := commandFailure("gh", []string{"api", "graphql"}, []byte("gh: Not Found (HTTP 404)"))
	if want := "GitHub API request failed: Not Found (HTTP 404)"; err.Error() != want {
		t.Errorf("commandFailure() = %q, want %q", err, want)
	}
}
