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
	for _, want := range []string{marker, "jj-stack-data:v1", `<a href="https://example.test/pull/8">#8</a>`, `<a href="https://github.com/EviHex/jj-blackbelt">black belt</a>`, "🔵", "🟣", "◆  prod"} {
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

func TestValidateBasesSkipsMergedParents(t *testing.T) {
	merged := 1
	prs := []PullRequest{
		{Number: 1, State: "MERGED", Head: "parent", Base: "prod"},
		{Number: 2, State: "OPEN", Head: "child", Base: "prod", Parent: &merged},
	}
	validate(prs, "prod")
	if prs[1].BaseWarning != "" || prs[1].ExpectedBase != "prod" {
		t.Fatalf("merged parent validation = %#v", prs[1])
	}
}

func TestValidateBasesReportsLiveParentMismatch(t *testing.T) {
	parent := 1
	prs := []PullRequest{
		{Number: 1, State: "OPEN", Head: "parent", Base: "prod"},
		{Number: 2, State: "OPEN", Head: "child", Base: "prod", Parent: &parent},
	}
	validate(prs, "prod")
	if prs[1].ExpectedBase != "parent" || prs[1].BaseWarning == "" {
		t.Fatalf("live parent validation = %#v", prs[1])
	}
}

func TestStackComponentsSeparatesTrunkChildren(t *testing.T) {
	root := 1
	prs := []PullRequest{{Number: 1}, {Number: 2, Parent: &root}, {Number: 9}}
	components := stackComponents(prs)
	if len(components) != 2 || len(components[0]) != 2 || len(components[1]) != 1 {
		t.Fatalf("stackComponents() = %#v", components)
	}
}
