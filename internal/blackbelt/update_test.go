package blackbelt

import (
	"context"
	"errors"
	"slices"
	"testing"
)

type recordingRunner struct {
	name string
	args []string
	err  error
}

func (r *recordingRunner) run(_ context.Context, name string, args ...string) command {
	r.name = name
	r.args = args
	return command{err: r.err}
}

func TestUpdateAddsCommentOperationContext(t *testing.T) {
	cause := errors.New("GitHub API request failed: Not Found (HTTP 404)")
	for _, test := range []struct {
		name    string
		comment map[string]any
		want    string
	}{
		{name: "create", want: "create stack note on PR #8700: " + cause.Error()},
		{name: "update", comment: map[string]any{"databaseId": float64(1)}, want: "update stack note on PR #8700: " + cause.Error()},
	} {
		t.Run(test.name, func(t *testing.T) {
			runner := &recordingRunner{err: cause}
			if err := update(context.Background(), runner, "ddoghq/dd-go", 8700, test.comment, "body"); err == nil || err.Error() != test.want {
				t.Fatalf("update() error = %v, want %q", err, test.want)
			}
		})
	}
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
