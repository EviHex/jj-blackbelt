package blackbelt

import (
	"fmt"
	"html"
	"strings"
)

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

func renderTerminalForest(prs []PullRequest, trunk string, current int, links bool) string {
	components := stackComponents(prs)
	parts := make([]string, 0, len(components)+1)
	parts = append(parts, fmt.Sprintf("PR stacks — %d stacks, %d PRs\n", len(components), len(prs)))
	for _, component := range components {
		parts = append(parts, renderTerminal(component, trunk, current, links))
	}
	return strings.Join(parts, "\n")
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
