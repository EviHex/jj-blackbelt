// Package blackbelt discovers jj PR stacks and maintains their GitHub notes.
package blackbelt

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const marker = "<!-- jj-stack-note -->"
const legacyMarker = "<sub>Stack maintained by agent.</sub>"
const footer = `<small>⚡ This stack has a 🥋 black belt in <a href="https://github.com/jj-vcs/jj">Jujutsu (jj)</a></small>`

type Options struct{ DryRun bool }
type Bookmark struct{ Name, CommitID string }
type History struct {
	Trunk   string       `json:"trunk"`
	Parents map[int]*int `json:"parents"`
}
type PullRequest struct {
	Number                        int
	Title, State, Head, Base, URL string
	Draft                         bool
	Review                        string
	CommitID                      string
	Parent                        *int
	ExpectedBase, BaseWarning     string
	Historical                    bool
	Commits                       []string
	CommitCount                   int
}
type command struct {
	output string
	err    error
}
type runner interface {
	run(context.Context, string, ...string) command
}
type shell struct{}

func (shell) run(ctx context.Context, name string, args ...string) command {
	c := exec.CommandContext(ctx, name, args...)
	b, e := c.Output()
	if e != nil {
		if x := new(exec.ExitError); errors.As(e, &x) {
			return command{"", fmt.Errorf("%s: %s", name, strings.TrimSpace(string(x.Stderr)))}
		}
		return command{"", e}
	}
	return command{string(b), nil}
}

func Run(ctx context.Context, options Options) error {
	return run(ctx, shell{}, options, os.Stdout, os.Stderr)
}
func run(ctx context.Context, r runner, options Options, out, errOut *os.File) error {
	root, repo, trunk, local, comments, historical, err := loadStack(ctx, r)
	if err != nil {
		return err
	}
	history := findHistory(local, comments)
	if err := writeHistory(root, history); err != nil {
		return err
	}
	for number := range history.Parents {
		if hasPR(local, number) {
			continue
		}
		if pr, ok := historical[number]; ok && pr.State == "MERGED" {
			pr.Historical = true
			local = append(local, pr)
		}
	}
	mergeParents(local, history)
	if options.DryRun {
		current := currentNumber(ctx, r, local)
		interactive := terminal(out)
		_, e := fmt.Fprint(out, renderTerminal(local, trunk, current, interactive))
		return e
	}
	for _, pr := range local {
		if countStackComments(comments[pr.Number]) > 1 {
			return fmt.Errorf("PR #%d has multiple stack comments; refusing to choose one", pr.Number)
		}
		body := renderHTML(local, trunk, pr.Number)
		existing := oneStackComment(comments[pr.Number])
		if existing != nil && strings.TrimSpace(stringValue(existing["body"])) == strings.TrimSpace(body) {
			fmt.Fprintf(out, "\033[32m✓ PR #%d: up to date\033[0m\n", pr.Number)
			continue
		}
		if err := update(ctx, r, repo, pr.Number, existing, body); err != nil {
			return err
		}
	}
	for _, pr := range local {
		if pr.BaseWarning != "" {
			fmt.Fprintf(errOut, "warning: PR #%d targets '%s', expected '%s'\n", pr.Number, pr.Base, pr.ExpectedBase)
		}
	}
	return nil
}

func loadStack(ctx context.Context, r runner) (string, string, string, []PullRequest, map[int][]map[string]any, map[int]PullRequest, error) {
	root := strings.TrimSpace(runCmd(ctx, r, "jj", "--ignore-working-copy", "root"))
	if root == "" {
		return "", "", "", nil, nil, nil, errors.New("run this command inside a jj repository")
	}
	repo, trunk, err := metadata(ctx, r)
	if err != nil {
		return "", "", "", nil, nil, nil, err
	}
	bookmarks, err := connected(ctx, r, trunk)
	if err != nil {
		return "", "", "", nil, nil, nil, err
	}
	values, comments, historical, err := pullRequests(ctx, r, repo, bookmarks, readHistory(root).Parents)
	if err != nil {
		return "", "", "", nil, nil, nil, err
	}
	prs, err := match(bookmarks, values)
	if err != nil {
		return "", "", "", nil, nil, nil, err
	}
	if err = assignParents(ctx, r, prs); err != nil {
		return "", "", "", nil, nil, nil, err
	}
	validate(prs, trunk)
	return root, repo, trunk, prs, comments, historical, nil
}
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
		fields = append(fields, fmt.Sprintf(`historical%d:pullRequest(number:$%s){number title state headRefName baseRefName url isDraft reviewDecision commits(first:100){totalCount nodes{commit{messageHeadline messageBody}}}}`, i, v))
		args = append(args, "-F", fmt.Sprintf("%s=%d", v, n))
	}
	query := fmt.Sprintf("query(%s){repository(owner:$owner,name:$name){%s}}", strings.Join(vars, ","), strings.Join(fields, ""))
	args = append(args, "-f", "query="+query, "-F", "owner="+owner, "-F", "name="+name)
	out, e := must(ctx, r, "gh", args...)
	if e != nil {
		return nil, nil, nil, e
	}
	var payload map[string]any
	if e = json.Unmarshal([]byte(out), &payload); e != nil {
		return nil, nil, nil, e
	}
	repoValue := object(object(payload["data"])["repository"])
	if repoValue == nil {
		return nil, nil, nil, errors.New("GitHub repository was not found")
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

type event struct {
	kind, graph string
	pr          *PullRequest
}

func events(prs []PullRequest, trunk string, current int) []event {
	by := map[int]*PullRequest{}
	children := map[int][]*PullRequest{}
	var roots []*PullRequest
	for i := range prs {
		p := &prs[i]
		by[p.Number] = p
	}
	for i := range prs {
		p := &prs[i]
		if p.Parent == nil || by[*p.Parent] == nil {
			roots = append(roots, p)
		} else {
			children[*p.Parent] = append(children[*p.Parent], p)
		}
	}
	sort.Slice(roots, func(i, j int) bool { return roots[i].Number < roots[j].Number })
	for _, c := range children {
		sort.Slice(c, func(i, j int) bool { return c[i].Number < c[j].Number })
	}
	var ordered []*PullRequest
	var visit func(*PullRequest)
	visit = func(p *PullRequest) {
		for _, c := range children[p.Number] {
			visit(c)
		}
		ordered = append(ordered, p)
	}
	for _, p := range roots {
		visit(p)
	}
	lanes := []*int{}
	out := []event{}
	rail := func() string { return strings.TrimRight(strings.Repeat("│ ", len(lanes)), " ") }
	for _, p := range ordered {
		idx := -1
		for i, x := range lanes {
			if x != nil && *x == p.Number {
				idx = i
			}
		}
		if idx < 0 {
			n := p.Number
			lanes = append(lanes, &n)
			idx = len(lanes) - 1
		}
		symbol := "○"
		if p.State == "MERGED" {
			symbol = "◆"
		}
		parts := make([]string, len(lanes))
		for i := range lanes {
			parts[i] = "│ "
			if i == idx {
				parts[i] = symbol + "  "
			}
		}
		_, state := status(*p)
		dot, _ := status(*p)
		finger := ""
		if p.Number == current {
			finger = "  👈"
		}
		out = append(out, event{"node", strings.Join(parts, "") + "{node} " + dot + " " + state + finger, p})
		out = append(out, event{"commits", strings.Repeat("│ ", len(lanes)), p})
		if p.BaseWarning != "" {
			out = append(out, event{"warning", strings.Repeat("│ ", len(lanes)), p})
		}
		parent := p.Parent
		if parent != nil {
			found := -1
			for i, x := range lanes {
				if x != nil && *x == *parent {
					found = i
				}
			}
			if found >= 0 && found != idx {
				width := len(lanes)*2 - 1
				chars := make([]rune, width)
				for i := range lanes {
					chars[i*2] = '│'
				}
				lo, hi := found, idx
				if lo > hi {
					lo, hi = hi, lo
				}
				chars[lo*2] = '├'
				chars[hi*2] = '╯'
				for i := lo*2 + 1; i < hi*2; i++ {
					if i%2 == 1 {
						chars[i] = '─'
					} else {
						chars[i] = '┼'
					}
				}
				out = append(out, event{"graph", strings.TrimRight(string(chars), " "), nil})
				lanes = append(lanes[:idx], lanes[idx+1:]...)
				continue
			}
		}
		lanes[idx] = parent
		out = append(out, event{"graph", rail(), nil})
	}
	out = append(out, event{"trunk", trunk, nil})
	return out
}
func renderTerminal(prs []PullRequest, trunk string, current int, links bool) string {
	lines := []string{fmt.Sprintf("PR stack — %d PR%s", len(prs), plural(len(prs))), ""}
	for _, e := range events(prs, trunk, current) {
		switch e.kind {
		case "node":
			n := fmt.Sprintf("#%d", e.pr.Number)
			if links {
				n = "\033]8;;" + e.pr.URL + "\033\\" + n + "\033]8;;\033\\"
			} else {
				n += "  " + e.pr.URL
			}
			if links {
				n = "\033[1;36m" + n + "  " + singleLine(e.pr.Title) + "\033[0m"
			} else {
				n += "  " + singleLine(e.pr.Title)
			}
			lines = append(lines, e.graph[:strings.Index(e.graph, "{node}")]+n+e.graph[strings.Index(e.graph, "}")+1:])
		case "commits":
			for _, m := range displayCommits(*e.pr) {
				if links {
					m = "\033[2m" + m + "\033[0m"
				}
				lines = append(lines, e.graph+" • "+m)
			}
		case "warning":
			warning := "⚠ " + e.pr.BaseWarning
			if links {
				warning = "\033[31m" + warning + "\033[0m"
			}
			lines = append(lines, e.graph+" "+warning)
		case "trunk":
			lines = append(lines, "◆  "+e.graph)
		default:
			lines = append(lines, e.graph)
		}
	}
	return strings.Join(append(lines, ""), "\n")
}
func renderHTML(prs []PullRequest, trunk string, current int) string {
	es := events(prs, trunk, current)
	width := 0
	for _, e := range es {
		if e.kind == "node" {
			w := len(e.graph[:strings.Index(e.graph, "{node}")]) + len(fmt.Sprintf("#%d  %s", e.pr.Number, singleLine(e.pr.Title)))
			if w > width {
				width = w
			}
		}
	}
	width += 2
	lines := []string{"This change belongs to the following stack:", "<pre>"}
	for _, e := range es {
		switch e.kind {
		case "node":
			prefix := e.graph[:strings.Index(e.graph, "{node}")]
			meta := fmt.Sprintf("#%d  %s", e.pr.Number, singleLine(e.pr.Title))
			padding := strings.Repeat(" ", max(1, width-len(prefix)-len(meta)))
			suffix := e.graph[strings.Index(e.graph, "}")+1:]
			lines = append(lines, fmt.Sprintf(`%s<strong><a href="%s">#%d</a>  %s</strong>%s%s`, prefix, html.EscapeString(e.pr.URL), e.pr.Number, html.EscapeString(singleLine(e.pr.Title)), padding, strings.TrimLeft(suffix, " ")))
		case "commits":
			for _, m := range displayCommits(*e.pr) {
				lines = append(lines, e.graph+"  • <small>"+html.EscapeString(m)+"</small>")
			}
		case "warning":
			lines = append(lines, e.graph+" <strong>⚠ "+html.EscapeString(e.pr.BaseWarning)+"</strong>")
		case "trunk":
			lines = append(lines, "◆  "+html.EscapeString(e.graph))
		default:
			lines = append(lines, e.graph)
		}
	}
	lines = append(lines, "</pre>", footer, encodeHistory(prs, trunk), marker, "")
	return strings.Join(lines, "\n")
}
func displayCommits(p PullRequest) []string {
	m := p.Commits
	if len(m) == 0 {
		m = []string{singleLine(p.Title)}
	}
	out := append([]string{}, m[:min(3, len(m))]...)
	if n := p.CommitCount - len(out); n > 0 {
		out = append(out, fmt.Sprintf("… plus %d more commit%s", n, plural(n)))
	}
	return out
}
func encodeHistory(prs []PullRequest, trunk string) string {
	type node struct {
		Number int  `json:"number"`
		Parent *int `json:"parent"`
	}
	nodes := make([]node, len(prs))
	for i, p := range prs {
		nodes[i] = node{p.Number, p.Parent}
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Number < nodes[j].Number })
	b, _ := json.Marshal(struct {
		Version int    `json:"version"`
		Trunk   string `json:"trunk"`
		Nodes   []node `json:"nodes"`
	}{1, trunk, nodes})
	return "<!-- jj-stack-data:v1:" + base64.RawURLEncoding.EncodeToString(b) + " -->"
}

var dataRE = regexp.MustCompile(`<!-- jj-stack-data:v1:([A-Za-z0-9_-]+) -->`)
var pullRE = regexp.MustCompile(`/pull/(\d+)`)

func parseHistory(body string) History {
	m := dataRE.FindStringSubmatch(body)
	if len(m) == 2 {
		var v struct {
			Trunk string `json:"trunk"`
			Nodes []struct {
				Number int  `json:"number"`
				Parent *int `json:"parent"`
			} `json:"nodes"`
		}
		if b, e := base64.RawURLEncoding.DecodeString(m[1]); e == nil && json.Unmarshal(b, &v) == nil {
			h := History{v.Trunk, map[int]*int{}}
			for _, n := range v.Nodes {
				h.Parents[n.Number] = n.Parent
			}
			return h
		}
	}
	nums := pullRE.FindAllStringSubmatch(body, -1)
	h := History{Parents: map[int]*int{}}
	for i, m := range nums {
		var n int
		fmt.Sscanf(m[1], "%d", &n)
		if _, seen := h.Parents[n]; seen {
			continue
		}
		if i+1 < len(nums) {
			var p int
			fmt.Sscanf(nums[i+1][1], "%d", &p)
			h.Parents[n] = &p
		} else {
			h.Parents[n] = nil
		}
	}
	return h
}
func readHistory(root string) History {
	b, e := os.ReadFile(filepath.Join(root, ".jj", "jj-note-history.json"))
	if e != nil {
		return History{Parents: map[int]*int{}}
	}
	var h History
	if json.Unmarshal(b, &h) != nil || h.Parents == nil {
		h.Parents = map[int]*int{}
	}
	return h
}
func writeHistory(root string, h History) error {
	p := filepath.Join(root, ".jj", "jj-note-history.json")
	b, e := json.Marshal(h)
	if e != nil {
		return e
	}
	return os.WriteFile(p, b, 0600)
}
func findHistory(prs []PullRequest, comments map[int][]map[string]any) History {
	for i := len(prs) - 1; i >= 0; i-- {
		if c := oneStackComment(comments[prs[i].Number]); c != nil {
			return parseHistory(stringValue(c["body"]))
		}
	}
	return History{Parents: map[int]*int{}}
}
func oneStackComment(cs []map[string]any) map[string]any {
	var found map[string]any
	for _, c := range cs {
		b := stringValue(c["body"])
		if strings.Contains(b, marker) || strings.Contains(b, legacyMarker) {
			if found != nil {
				return nil
			}
			found = c
		}
	}
	return found
}
func countStackComments(cs []map[string]any) int {
	count := 0
	for _, c := range cs {
		b := stringValue(c["body"])
		if strings.Contains(b, marker) || strings.Contains(b, legacyMarker) {
			count++
		}
	}
	return count
}
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
		_, e = must(ctx, r, "gh", "pr", "comment", fmt.Sprint(number), "--repo", repo, "--body-file", p.Name())
	} else {
		_, e = must(ctx, r, "gh", "api", "-X", "PATCH", fmt.Sprintf("repos/%s/issues/comments/%v", repo, comment["databaseId"]), "-F", "body=@"+p.Name())
	}
	return e
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
func object(v any) map[string]any { m, _ := v.(map[string]any); return m }
func array(v any) []any           { a, _ := v.([]any); return a }
func objects(v []any) []map[string]any {
	out := []map[string]any{}
	for _, x := range v {
		if o := object(x); o != nil {
			out = append(out, o)
		}
	}
	return out
}
func stringValue(v any) string { s, _ := v.(string); return s }
func intValue(v any) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case json.Number:
		n, _ := x.Int64()
		return int(n)
	}
	return 0
}
func boolValue(v any) bool { b, _ := v.(bool); return b }
func nonempty(s string) []string {
	var out []string
	for _, v := range strings.Split(s, "\n") {
		if v = strings.TrimSpace(v); v != "" {
			out = append(out, v)
		}
	}
	return out
}
func keys[T comparable](m map[T]bool) []T {
	out := make([]T, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
func singleLine(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "(no description)"
	}
	a, _, found := strings.Cut(s, "\n")
	if found {
		return strings.TrimRight(a, " ") + " …"
	}
	return strings.TrimRight(a, " ")
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
