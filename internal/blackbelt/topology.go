package blackbelt

import (
	"context"
	"fmt"
	"sort"
)

func assignParents(ctx context.Context, r runner, prs []PullRequest) error {
	by := map[string]int{}
	ids := []string{}
	for i, p := range prs {
		by[p.CommitID] = i
		ids = append(ids, p.CommitID)
	}
	for i := range prs {
		parents, e := idsFor(ctx, r, fmt.Sprintf("heads(((::%s) ~ %s) & %s)", prs[i].CommitID, prs[i].CommitID, union(ids)))
		if e != nil {
			return e
		}
		if len(parents) > 1 {
			return fmt.Errorf("PR #%d has multiple PR ancestors; merge-shaped DAGs are not supported", prs[i].Number)
		}
		if len(parents) == 1 {
			x := prs[by[parents[0]]].Number
			prs[i].Parent = &x
		}
	}
	return nil
}

func stackComponents(prs []PullRequest) [][]PullRequest {
	byNumber := make(map[int]PullRequest, len(prs))
	for _, pr := range prs {
		byNumber[pr.Number] = pr
	}
	groups := map[int][]PullRequest{}
	for _, pr := range prs {
		root := pr.Number
		seen := map[int]bool{}
		position := pr
		for position.Parent != nil && byNumber[*position.Parent].Number != 0 && !seen[*position.Parent] {
			seen[*position.Parent] = true
			position = byNumber[*position.Parent]
			root = position.Number
		}
		groups[root] = append(groups[root], pr)
	}
	roots := make([]int, 0, len(groups))
	for root := range groups {
		roots = append(roots, root)
	}
	sort.Ints(roots)
	result := make([][]PullRequest, 0, len(roots))
	for _, root := range roots {
		result = append(result, groups[root])
	}
	return result
}
func idsFor(ctx context.Context, r runner, s string) ([]string, error) { return ids(ctx, r, s) }
func validate(prs []PullRequest, trunk string) {
	by := map[int]*PullRequest{}
	for i := range prs {
		by[prs[i].Number] = &prs[i]
	}
	for i := range prs {
		expected := trunk
		parent := prs[i].Parent
		seen := map[int]bool{prs[i].Number: true}
		for parent != nil && by[*parent] != nil {
			if seen[*parent] {
				break
			}
			seen[*parent] = true
			p := by[*parent]
			if p.State != "MERGED" {
				expected = p.Head
				break
			}
			parent = p.Parent
		}
		prs[i].ExpectedBase = expected
		if prs[i].State == "OPEN" && prs[i].Base != expected {
			prs[i].BaseWarning = fmt.Sprintf("base is %s; expected %s", prs[i].Base, expected)
		}
	}
}
func mergeParents(prs []PullRequest, h History) {
	present := map[int]bool{}
	for _, p := range prs {
		present[p.Number] = true
	}
	for i := range prs {
		if prs[i].Historical {
			if parent := h.Parents[prs[i].Number]; parent != nil && present[*parent] {
				prs[i].Parent = parent
			}
		} else if prs[i].Parent == nil {
			if parent := h.Parents[prs[i].Number]; parent != nil && present[*parent] {
				prs[i].Parent = parent
			}
		}
	}
}
func status(p PullRequest) (string, string) {
	if p.State == "OPEN" && p.Draft {
		return "🟡", "Draft"
	}
	if p.State == "OPEN" && p.Review == "APPROVED" {
		return "🔵", "Reviewed"
	}
	if p.State == "OPEN" {
		return "🟢", "Open"
	}
	if p.State == "MERGED" {
		return "🟣", "Merged"
	}
	return "🔴", "Closed"
}
