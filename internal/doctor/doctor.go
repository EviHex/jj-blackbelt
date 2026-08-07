// Package doctor checks whether blackbelt's external tools are ready.
package doctor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const MinimumJJ = "0.40.0"

type Result struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

// Run performs all preflight checks and prints a human or JSON report.
func Run(ctx context.Context, jsonOutput bool) error {
	results := []Result{checkJJ(ctx), checkGH(ctx), checkGitHubAuth(ctx), checkRepository(ctx)}
	failed := false
	for _, result := range results {
		failed = failed || result.Status == "error"
	}
	if jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(results); err != nil {
			return err
		}
	} else {
		for _, result := range results {
			symbol := "✓"
			if result.Status == "warning" {
				symbol = "!"
			} else if result.Status == "error" {
				symbol = "✗"
			}
			fmt.Printf("%s %-12s %s\n", symbol, result.Name, result.Detail)
		}
	}
	if failed {
		return fmt.Errorf("one or more required checks failed")
	}
	return nil
}

func checkJJ(ctx context.Context) Result {
	output, err := exec.CommandContext(ctx, "jj", "--version").CombinedOutput()
	if err != nil {
		return Result{"jj", "error", "not found"}
	}
	found := extractVersion(string(output))
	if compareVersions(found, MinimumJJ) < 0 {
		return Result{"jj", "error", fmt.Sprintf("%s is older than required %s", found, MinimumJJ)}
	}
	return Result{"jj", "ok", found}
}

func checkGH(ctx context.Context) Result {
	output, err := exec.CommandContext(ctx, "gh", "--version").CombinedOutput()
	if err != nil {
		return Result{"gh", "error", "not found"}
	}
	return Result{"gh", "ok", extractVersion(string(output))}
}

func checkGitHubAuth(ctx context.Context) Result {
	output, err := exec.CommandContext(ctx, "gh", "auth", "status").CombinedOutput()
	if err != nil {
		detail := strings.TrimSpace(string(output))
		if detail == "" {
			detail = "not authenticated; run gh auth login"
		}
		return Result{"GitHub auth", "error", firstLine(detail)}
	}
	return Result{"GitHub auth", "ok", "authenticated"}
}

func checkRepository(ctx context.Context) Result {
	output, err := exec.CommandContext(ctx, "jj", "--ignore-working-copy", "root").CombinedOutput()
	if err != nil {
		return Result{"repository", "warning", "not inside a jj repository"}
	}
	return Result{"repository", "ok", strings.TrimSpace(string(output))}
}

var versionPattern = regexp.MustCompile(`\d+\.\d+(?:\.\d+)?`)

func extractVersion(value string) string {
	if match := versionPattern.FindString(value); match != "" {
		return match
	}
	return "unknown"
}

func compareVersions(left, right string) int {
	parse := func(value string) [3]int {
		var result [3]int
		for index, part := range strings.Split(value, ".") {
			if index >= len(result) {
				break
			}
			result[index], _ = strconv.Atoi(part)
		}
		return result
	}
	a, b := parse(left), parse(right)
	for index := range a {
		if a[index] < b[index] {
			return -1
		}
		if a[index] > b[index] {
			return 1
		}
	}
	return 0
}

func firstLine(value string) string {
	line, _, _ := strings.Cut(value, "\n")
	return line
}
