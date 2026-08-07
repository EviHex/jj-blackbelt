// Package config loads blackbelt's layered TOML configuration.
package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config is blackbelt's complete configuration.
type Config struct {
	Stack Stack `toml:"stack"`
}

// Stack configures stack commands.
type Stack struct {
	DefaultCommand string `toml:"default-command"`
	Log            Log    `toml:"log"`
}

// Log configures stack discovery and display.
type Log struct {
	Revset string `toml:"revset"`
	All    bool   `toml:"all"`
}

// Defaults returns the built-in configuration.
func Defaults() Config {
	return Config{Stack: Stack{DefaultCommand: "log"}}
}

// Load overlays user and repository configuration on the built-in defaults.
func Load(ctx context.Context) (Config, error) {
	paths, err := Paths(ctx)
	if err != nil {
		return Config{}, err
	}
	return loadFiles(paths)
}

func loadFiles(paths []string) (Config, error) {
	value := Defaults()
	for _, path := range paths {
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return Config{}, fmt.Errorf("inspect config %s: %w", path, err)
		}
		if _, err := toml.DecodeFile(path, &value); err != nil {
			return Config{}, fmt.Errorf("load config %s: %w", path, err)
		}
	}
	if value.Stack.DefaultCommand == "" {
		value.Stack.DefaultCommand = "log"
	}
	return value, nil
}

// Paths returns configuration files from lowest to highest precedence.
func Paths(ctx context.Context) ([]string, error) {
	directory, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("resolve user config directory: %w", err)
	}
	paths := []string{filepath.Join(directory, "blackbelt", "config.toml")}
	command := exec.CommandContext(ctx, "jj", "--ignore-working-copy", "root")
	output, err := command.Output()
	if err == nil {
		root := strings.TrimSpace(string(output))
		if root != "" {
			paths = append(paths, filepath.Join(root, ".jj", "blackbelt.toml"))
		}
	}
	return paths, nil
}
