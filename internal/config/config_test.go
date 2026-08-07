package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaults(t *testing.T) {
	value := Defaults()
	if value.Stack.DefaultCommand != "log" {
		t.Fatalf("default stack command = %q", value.Stack.DefaultCommand)
	}
}

func TestLoadFilesAppliesLayersInOrder(t *testing.T) {
	directory := t.TempDir()
	user := filepath.Join(directory, "user.toml")
	repository := filepath.Join(directory, "repo.toml")
	if err := os.WriteFile(user, []byte("[stack]\ndefault-command = 'draw'\n[stack.log]\nall = true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(repository, []byte("[stack]\ndefault-command = 'log'\n[stack.log]\nrevset = 'mine()'\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	value, err := loadFiles([]string{user, repository})
	if err != nil {
		t.Fatalf("loadFiles() error = %v", err)
	}
	if value.Stack.DefaultCommand != "log" || value.Stack.Log.Revset != "mine()" || !value.Stack.Log.All {
		t.Fatalf("loadFiles() = %#v", value)
	}
}
