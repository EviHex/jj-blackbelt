package blackbelt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

func runCmd(ctx context.Context, r runner, name string, args ...string) string {
	v := r.run(ctx, name, args...)
	if v.err != nil {
		return ""
	}
	return v.output
}
func must(ctx context.Context, r runner, name string, args ...string) (string, error) {
	v := r.run(ctx, name, args...)
	return v.output, v.err
}
func ids(ctx context.Context, r runner, rev string) ([]string, error) {
	s, e := must(ctx, r, "jj", "--ignore-working-copy", "log", "--no-graph", "-r", rev, "-T", `commit_id ++ "\n"`)
	return nonempty(s), e
}
func bookmarks(ctx context.Context, r runner, rev string) ([]Bookmark, error) {
	s, e := must(ctx, r, "jj", "--ignore-working-copy", "bookmark", "list", "--tracked", "--remote", "origin", "--revision", rev, "--template", `if(remote, "R\t" ++ name ++ "\t" ++ if(self.normal_target(), self.normal_target().commit_id()) ++ "\n", "L\t" ++ name ++ "\t" ++ if(self.normal_target(), self.normal_target().commit_id()) ++ "\n")`)
	if e != nil {
		return nil, e
	}
	var result []Bookmark
	for _, line := range nonempty(s) {
		p := strings.Split(line, "\t")
		if len(p) == 3 && p[0] == "L" && p[1] != "" && p[2] != "" {
			result = append(result, Bookmark{p[1], p[2]})
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}
func union(ids []string) string { sort.Strings(ids); return "(" + strings.Join(ids, " | ") + ")" }
func connected(ctx context.Context, r runner, trunk string) ([]Bookmark, error) {
	ts, e := ids(ctx, r, `bookmarks(exact:"`+trunk+`")`)
	if e != nil || len(ts) != 1 {
		return nil, fmt.Errorf("GitHub default branch bookmark %q must resolve to one commit", trunk)
	}
	candidates, e := bookmarks(ctx, r, fmt.Sprintf("bookmarks() ~ ::%s", ts[0]))
	if e != nil {
		return nil, e
	}
	seeds, e := bookmarks(ctx, r, fmt.Sprintf("%s..@ | @::", ts[0]))
	if e != nil {
		return nil, e
	}
	if len(seeds) == 0 {
		return nil, errors.New("no tracked origin bookmarks found around @")
	}
	byID := map[string]bool{}
	for _, b := range candidates {
		byID[b.CommitID] = true
	}
	selected := map[string]bool{}
	for _, b := range seeds {
		if byID[b.CommitID] {
			selected[b.CommitID] = true
		}
	}
	if len(selected) == 0 {
		return nil, errors.New("no unmerged tracked origin bookmarks found around @")
	}
	for {
		chosen := keys(selected)
		comparable, e := ids(ctx, r, fmt.Sprintf("%s & ((::%s) | (%s::))", union(keys(byID)), union(chosen), union(chosen)))
		if e != nil {
			return nil, e
		}
		changed := false
		for _, id := range comparable {
			if byID[id] && !selected[id] {
				selected[id] = true
				changed = true
			}
		}
		if !changed {
			break
		}
	}
	var result []Bookmark
	for _, b := range candidates {
		if selected[b.CommitID] {
			result = append(result, b)
		}
	}
	return result, nil
}
func metadata(ctx context.Context, r runner) (string, string, error) {
	s, e := must(ctx, r, "jj", "--ignore-working-copy", "git", "remote", "list")
	if e == nil {
		for _, line := range nonempty(s) {
			p := strings.Fields(line)
			if len(p) >= 2 && p[0] == "origin" {
				repo := repoFromURL(p[1])
				names, _ := must(ctx, r, "jj", "--ignore-working-copy", "bookmark", "list", "--revision", "trunk()", "--template", `if(!remote, name ++ "\n", "")`)
				ns := nonempty(names)
				if repo != "" && len(ns) == 1 {
					return repo, ns[0], nil
				}
			}
		}
	}
	out, e := must(ctx, r, "gh", "repo", "view", "--json", "nameWithOwner,defaultBranchRef")
	if e != nil {
		return "", "", e
	}
	var v struct {
		Name    string `json:"nameWithOwner"`
		Default struct {
			Name string `json:"name"`
		} `json:"defaultBranchRef"`
	}
	if json.Unmarshal([]byte(out), &v) != nil || v.Name == "" {
		return "", "", errors.New("cannot resolve repository metadata")
	}
	return v.Name, v.Default.Name, nil
}
func repoFromURL(v string) string {
	v = strings.TrimSuffix(strings.TrimSuffix(v, "/"), ".git")
	if i := strings.Index(v, ":"); i >= 0 && !strings.Contains(v, "://") {
		v = v[i+1:]
	}
	p := strings.Split(strings.Trim(v, "/"), "/")
	if len(p) < 2 {
		return ""
	}
	return p[len(p)-2] + "/" + p[len(p)-1]
}
