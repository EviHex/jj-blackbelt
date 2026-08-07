package blackbelt

import (
	"context"
	"fmt"
	"os"
)

func update(ctx context.Context, r runner, repo string, number int, comment map[string]any, body string) error {
	p, e := os.CreateTemp("", "blackbelt-*.html")
	if e != nil {
		return e
	}
	defer os.Remove(p.Name())
	if _, e = p.WriteString(body); e != nil {
		return e
	}
	if e = p.Close(); e != nil {
		return e
	}
	if comment == nil {
		_, e = mustWithActivity(ctx, r, fmt.Sprintf("Creating stack note on PR #%d", number), "gh", "pr", "comment", fmt.Sprint(number), "--repo", repo, "--body-file", p.Name())
		if e != nil {
			return fmt.Errorf("create stack note on PR #%d: %w", number, e)
		}
	} else {
		_, e = mustWithActivity(ctx, r, fmt.Sprintf("Updating stack note on PR #%d", number), "gh", "api", "-X", "PATCH", fmt.Sprintf("repos/%s/issues/comments/%d", repo, int64Value(comment["databaseId"])), "-F", "body=@"+p.Name())
		if e != nil {
			return fmt.Errorf("update stack note on PR #%d: %w", number, e)
		}
	}
	return nil
}
func currentNumber(ctx context.Context, r runner, prs []PullRequest) int {
	by := map[string]int{}
	for _, p := range prs {
		if p.CommitID != "" {
			by[p.CommitID] = p.Number
		}
	}
	at, _ := ids(ctx, r, "@")
	if len(at) > 0 && by[at[0]] != 0 {
		return by[at[0]]
	}
	selected := make([]string, 0, len(by))
	for commitID := range by {
		selected = append(selected, commitID)
	}
	if len(selected) == 0 {
		return 0
	}
	if descendants, err := ids(ctx, r, fmt.Sprintf("roots((@::) & %s)", union(selected))); err == nil && len(descendants) == 1 {
		return by[descendants[0]]
	}
	if ancestors, err := ids(ctx, r, fmt.Sprintf("heads((::@) & %s)", union(selected))); err == nil && len(ancestors) == 1 {
		return by[ancestors[0]]
	}
	return 0
}
func hasPR(prs []PullRequest, n int) bool {
	for _, p := range prs {
		if p.Number == n {
			return true
		}
	}
	return false
}
func terminal(f *os.File) bool {
	info, e := f.Stat()
	return e == nil && (info.Mode()&os.ModeCharDevice) != 0 && os.Getenv("TERM") != "dumb"
}
