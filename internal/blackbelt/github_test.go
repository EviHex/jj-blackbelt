package blackbelt

import (
	"context"
	"errors"
	"testing"
)

func TestPullRequestsLoadsCommentsForHistoricalPRs(t *testing.T) {
	runner := &queuedRunner{outputs: []string{`{
  "data": {
    "repository": {
      "historical0": {
        "number": 7806,
        "state": "MERGED",
        "comments": {
          "nodes": [{"databaseId": 5173173319, "body": "<!-- jj-stack-note -->"}]
        }
      }
    }
  }
}`}}
	parent := 0

	_, comments, historical, err := pullRequests(context.Background(), runner, "ddoghq/dd-go", nil, map[int]*int{7806: &parent})
	if err != nil {
		t.Fatal(err)
	}
	if historical[7806].State != "MERGED" {
		t.Fatalf("historical PR = %#v", historical[7806])
	}
	if len(comments[7806]) != 1 || int64Value(comments[7806][0]["databaseId"]) != 5173173319 {
		t.Fatalf("historical comments = %#v", comments[7806])
	}
}

func TestPullRequestsAddsFetchContext(t *testing.T) {
	cause := errors.New("GitHub API request failed: Not Found (HTTP 404)")
	_, _, _, err := pullRequests(context.Background(), failingRunner{err: cause}, "ddoghq/dd-go", nil, nil)
	if want := "fetch PR details from ddoghq/dd-go: " + cause.Error(); err == nil || err.Error() != want {
		t.Fatalf("pullRequests() error = %v, want %q", err, want)
	}
}

type failingRunner struct{ err error }

func (r failingRunner) run(_ context.Context, _ string, _ ...string) command {
	return command{err: r.err}
}
