package blackbelt

import (
	"strings"
	"testing"
)

func TestHTMLPreservesTheEstablishedCommentContract(t *testing.T) {
	parent := 7
	prs := []PullRequest{
		{Number: 8, Title: "child", State: "OPEN", Review: "APPROVED", Head: "child", URL: "https://example.test/pull/8", Parent: &parent, Commits: []string{"feat: child"}, CommitCount: 1},
		{Number: 7, Title: "parent", State: "MERGED", Head: "parent", URL: "https://example.test/pull/7", Commits: []string{"feat: parent"}, CommitCount: 1},
	}
	body := renderHTML(prs, "prod", 8)
	for _, want := range []string{marker, "jj-stack-data:v1", `<a href="https://example.test/pull/8">#8</a>`, "🔵", "🟣", "◆  prod"} {
		if !strings.Contains(body, want) {
			t.Errorf("comment missing %q:\n%s", want, body)
		}
	}
}

func TestStatusAndCommitLimit(t *testing.T) {
	p := PullRequest{State: "OPEN", Review: "APPROVED", Commits: []string{"one", "two", "three", "four"}, CommitCount: 4}
	_, label := status(p)
	if label != "Reviewed" {
		t.Fatalf("status = %q", label)
	}
	got := displayCommits(p)
	if len(got) != 4 || got[3] != "… plus 1 more commit" {
		t.Fatalf("commits = %#v", got)
	}
}
