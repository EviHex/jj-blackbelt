package blackbelt

import (
	"context"
	"fmt"
	"testing"
)

type queuedRunner struct {
	outputs []string
	calls   int
}

func (r *queuedRunner) run(_ context.Context, _ string, _ ...string) command {
	if r.calls >= len(r.outputs) {
		return command{err: fmt.Errorf("unexpected command %d", r.calls+1)}
	}
	output := r.outputs[r.calls]
	r.calls++
	return command{output: output}
}

func TestConnectedExpandsFromCurrentPathIntoSiblingBranches(t *testing.T) {
	runner := &queuedRunner{outputs: []string{
		"T\n",
		"L\ta\tA\nL\tb\tB\nL\tc\tC\nL\td\tD\nL\tunrelated\tX\n",
		"L\tc\tC\n",
		"A\nB\nC\n",
		"A\nB\nC\nD\n",
		"A\nB\nC\nD\n",
	}}

	got, err := connected(context.Background(), runner, "prod", "", false)
	if err != nil {
		t.Fatalf("connected() error = %v", err)
	}
	if len(got) != 4 {
		t.Fatalf("connected() = %#v, want four bookmarks in the current tree", got)
	}
	for _, bookmark := range got {
		if bookmark.Name == "unrelated" {
			t.Fatalf("connected() included unrelated trunk child: %#v", got)
		}
	}
}

func TestConnectedAllReturnsEveryCandidate(t *testing.T) {
	runner := &queuedRunner{outputs: []string{
		"T\n",
		"L\ta\tA\nL\tunrelated\tX\n",
	}}

	got, err := connected(context.Background(), runner, "prod", "", true)
	if err != nil {
		t.Fatalf("connected() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("connected(all) = %#v, want all candidates", got)
	}
}
