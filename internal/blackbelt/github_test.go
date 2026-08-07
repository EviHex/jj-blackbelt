package blackbelt

import (
	"context"
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
