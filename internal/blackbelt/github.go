package blackbelt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

func pullRequests(ctx context.Context, r runner, repo string, bs []Bookmark, history map[int]*int) ([]map[string]any, map[int][]map[string]any, map[int]PullRequest, error) {
	owner, name, _ := strings.Cut(repo, "/")
	if owner == "" || name == "" {
		return nil, nil, nil, fmt.Errorf("invalid GitHub repository %q", repo)
	}
	args := []string{"api", "graphql"}
	vars := []string{"$owner:String!", "$name:String!"}
	fields := []string{}
	heads := map[string]bool{}
	for _, b := range bs {
		heads[b.Name] = true
	}
	hs := keys(heads)
	sort.Strings(hs)
	for i, h := range hs {
		v := fmt.Sprintf("head%d", i)
		vars = append(vars, "$"+v+":String!")
		fields = append(fields, fmt.Sprintf(`pr%d:pullRequests(first:1,states:[OPEN,MERGED,CLOSED],headRefName:$%s,orderBy:{field:CREATED_AT,direction:DESC}){nodes{number title state headRefName baseRefName url isDraft reviewDecision commits(first:100){totalCount nodes{commit{messageHeadline messageBody}}} comments(last:100){totalCount nodes{databaseId body}}}}`, i, v))
		args = append(args, "-F", v+"="+h)
	}
	ns := make([]int, 0, len(history))
	for n := range history {
		ns = append(ns, n)
	}
	sort.Ints(ns)
	for i, n := range ns {
		v := fmt.Sprintf("number%d", i)
		vars = append(vars, "$"+v+":Int!")
		fields = append(fields, fmt.Sprintf(`historical%d:pullRequest(number:$%s){number title state headRefName baseRefName url isDraft reviewDecision commits(first:100){totalCount nodes{commit{messageHeadline messageBody}}} comments(last:100){totalCount nodes{databaseId body}}}`, i, v))
		args = append(args, "-F", fmt.Sprintf("%s=%d", v, n))
	}
	query := fmt.Sprintf("query(%s){repository(owner:$owner,name:$name){%s}}", strings.Join(vars, ","), strings.Join(fields, ""))
	args = append(args, "-f", "query="+query, "-F", "owner="+owner, "-F", "name="+name)
	out, e := mustWithActivity(ctx, r, "Fetching PR details from GitHub", "gh", args...)
	if e != nil {
		return nil, nil, nil, fmt.Errorf("fetch PR details from %s: %w", repo, e)
	}
	var payload map[string]any
	if e = json.Unmarshal([]byte(out), &payload); e != nil {
		return nil, nil, nil, fmt.Errorf("parse PR details from %s: %w", repo, e)
	}
	repoValue := object(object(payload["data"])["repository"])
	if repoValue == nil {
		return nil, nil, nil, fmt.Errorf("fetch PR details from %s: GitHub repository was not found", repo)
	}
	values := []map[string]any{}
	comments := map[int][]map[string]any{}
	historical := map[int]PullRequest{}
	for i := range hs {
		nodes := array(object(repoValue[fmt.Sprintf("pr%d", i)])["nodes"])
		if len(nodes) == 0 {
			continue
		}
		pr := object(nodes[0])
		if pr == nil {
			continue
		}
		values = append(values, pr)
		n := intValue(pr["number"])
		comments[n] = objects(array(object(pr["comments"])["nodes"]))
	}
	for i, n := range ns {
		if p := object(repoValue[fmt.Sprintf("historical%d", i)]); p != nil {
			historical[n] = fromGH(p, "")
			comments[n] = objects(array(object(p["comments"])["nodes"]))
		}
	}
	return values, comments, historical, nil
}
func fromGH(v map[string]any, commit string) PullRequest {
	p := PullRequest{Number: intValue(v["number"]), Title: stringValue(v["title"]), State: stringValue(v["state"]), Head: stringValue(v["headRefName"]), Base: stringValue(v["baseRefName"]), URL: stringValue(v["url"]), Draft: boolValue(v["isDraft"]), Review: stringValue(v["reviewDecision"]), CommitID: commit}
	c := object(v["commits"])
	p.CommitCount = intValue(c["totalCount"])
	for _, n := range array(c["nodes"]) {
		co := object(object(n)["commit"])
		s := singleLine(stringValue(co["messageHeadline"]))
		if s != "(no description)" && strings.TrimSpace(stringValue(co["messageBody"])) != "" && !strings.HasSuffix(s, "…") && !strings.HasSuffix(s, "...") {
			s += " …"
		}
		if s != "(no description)" {
			p.Commits = append(p.Commits, s)
		}
	}
	if p.CommitCount == 0 {
		p.CommitCount = len(p.Commits)
	}
	return p
}
func match(bs []Bookmark, values []map[string]any) ([]PullRequest, error) {
	byHead := map[string]map[string]any{}
	for _, v := range values {
		h := stringValue(v["headRefName"])
		if old := byHead[h]; old == nil || intValue(v["number"]) > intValue(old["number"]) {
			byHead[h] = v
		}
	}
	seen := map[int]bool{}
	commits := map[string]bool{}
	var out []PullRequest
	for _, b := range bs {
		v := byHead[b.Name]
		if v == nil {
			continue
		}
		p := fromGH(v, b.CommitID)
		if seen[p.Number] {
			continue
		}
		if commits[p.CommitID] {
			return nil, errors.New("multiple PR bookmarks point at the same commit; tree parentage is ambiguous")
		}
		commits[p.CommitID] = true
		seen[p.Number] = true
		out = append(out, p)
	}
	if len(out) == 0 {
		return nil, errors.New("no PRs match tracked origin bookmarks in the tree around @")
	}
	return out, nil
}
