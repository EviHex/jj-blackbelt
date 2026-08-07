package blackbelt

import (
	"os"
	"strings"
	"testing"
)

func TestActivityShowsInteractiveSuccess(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	stream, err := os.CreateTemp(t.TempDir(), "progress")
	if err != nil {
		t.Fatal(err)
	}
	progress := newActivity("Fetching PR details from GitHub", stream, true)
	progress.finish(true)
	if _, err := stream.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(stream.Name())
	if err != nil {
		t.Fatal(err)
	}
	if want := "✓ Fetching PR details from GitHub"; !strings.Contains(string(contents), want) {
		t.Fatalf("progress output = %q, want %q", contents, want)
	}
}

func TestActivityShowsInteractiveFailure(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	stream, err := os.CreateTemp(t.TempDir(), "progress")
	if err != nil {
		t.Fatal(err)
	}
	progress := newActivity("Updating stack note on PR #8700", stream, true)
	progress.finish(false)
	if _, err := stream.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(stream.Name())
	if err != nil {
		t.Fatal(err)
	}
	if want := "✗ Updating stack note on PR #8700"; !strings.Contains(string(contents), want) {
		t.Fatalf("progress output = %q, want %q", contents, want)
	}
}

func TestActivityStaysQuietWhenNotInteractive(t *testing.T) {
	stream, err := os.CreateTemp(t.TempDir(), "progress")
	if err != nil {
		t.Fatal(err)
	}
	progress := newActivity("Fetching PR details from GitHub", stream, false)
	progress.finish(true)
	contents, err := os.ReadFile(stream.Name())
	if err != nil {
		t.Fatal(err)
	}
	if len(contents) != 0 {
		t.Fatalf("non-interactive progress output = %q, want empty", contents)
	}
}
