package blackbelt

import (
	"context"
	"slices"
	"testing"
)

type recordingRunner struct {
	name string
	args []string
}

func (r *recordingRunner) run(_ context.Context, name string, args ...string) command {
	r.name = name
	r.args = args
	return command{}
}

func TestUpdateFormatsLargeCommentIDAsInteger(t *testing.T) {
	runner := &recordingRunner{}
	comment := map[string]any{"databaseId": float64(5214848964)}

	if err := update(context.Background(), runner, "ddoghq/dd-go", 8700, comment, "body"); err != nil {
		t.Fatal(err)
	}
	if runner.name != "gh" {
		t.Fatalf("command = %q, want gh", runner.name)
	}
	want := "repos/ddoghq/dd-go/issues/comments/5214848964"
	if !slices.Contains(runner.args, want) {
		t.Fatalf("arguments = %#v, want endpoint %q", runner.args, want)
	}
}
