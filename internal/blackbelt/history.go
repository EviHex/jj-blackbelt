package blackbelt

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

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
