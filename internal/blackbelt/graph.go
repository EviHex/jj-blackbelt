package blackbelt

import (
	"sort"
	"strings"
)

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
