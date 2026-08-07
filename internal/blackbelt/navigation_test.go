package blackbelt

import "testing"

func TestNavigationTraversesPRNodes(t *testing.T) {
	one, two := 1, 2
	prs := []PullRequest{
		{Number: 1, Head: "root", CommitID: "A"},
		{Number: 2, Head: "left", CommitID: "B", Parent: &one},
		{Number: 3, Head: "tip", CommitID: "C", Parent: &two},
	}
	choose := func(values []PullRequest) (PullRequest, error) { return values[0], nil }

	up, err := navigationTarget(prs, 1, NavigateOptions{Direction: "up", Steps: 2}, choose)
	if err != nil || up.Number != 3 {
		t.Fatalf("up two = %#v, %v", up, err)
	}
	bottom, err := navigationTarget(prs, 3, NavigateOptions{Direction: "bottom"}, choose)
	if err != nil || bottom.Number != 1 {
		t.Fatalf("bottom = %#v, %v", bottom, err)
	}
}

func TestNavigationPromptsAtTreeSplit(t *testing.T) {
	root := 1
	prs := []PullRequest{
		{Number: 1, Head: "root"},
		{Number: 2, Head: "left", Parent: &root},
		{Number: 3, Head: "right", Parent: &root},
	}
	called := false
	choose := func(values []PullRequest) (PullRequest, error) {
		called = true
		return values[1], nil
	}

	got, err := navigationTarget(prs, 1, NavigateOptions{Direction: "up"}, choose)
	if err != nil || got.Number != 3 || !called {
		t.Fatalf("split navigation = %#v, %v, called=%v", got, err, called)
	}
}

func TestNavigationGotoAcceptsNumberAndBookmark(t *testing.T) {
	prs := []PullRequest{{Number: 42, Head: "topic"}}
	choose := func(values []PullRequest) (PullRequest, error) { return values[0], nil }
	for _, target := range []string{"42", "#42", "topic"} {
		got, err := navigationTarget(prs, 0, NavigateOptions{Direction: "goto", Target: target}, choose)
		if err != nil || got.Number != 42 {
			t.Fatalf("goto %q = %#v, %v", target, got, err)
		}
	}
}
