// Package blackbelt discovers jj PR stacks and maintains their GitHub notes.
package blackbelt

const marker = "<!-- jj-stack-note -->"
const legacyMarker = "<sub>Stack maintained by agent.</sub>"
const footer = `<small>⚡ This stack has a 🥋 black belt in <a href="https://github.com/jj-vcs/jj">Jujutsu (jj)</a></small>`

type Options struct {
	DryRun bool
	All    bool
	JSON   bool
	Revset string
}
type Bookmark struct{ Name, CommitID string }
type History struct {
	Trunk   string       `json:"trunk"`
	Parents map[int]*int `json:"parents"`
}
type PullRequest struct {
	Number       int      `json:"number"`
	Title        string   `json:"title"`
	State        string   `json:"state"`
	Head         string   `json:"head"`
	Base         string   `json:"base"`
	URL          string   `json:"url"`
	Draft        bool     `json:"draft"`
	Review       string   `json:"review,omitempty"`
	CommitID     string   `json:"commit_id,omitempty"`
	Parent       *int     `json:"parent,omitempty"`
	ExpectedBase string   `json:"expected_base,omitempty"`
	BaseWarning  string   `json:"base_warning,omitempty"`
	Historical   bool     `json:"historical,omitempty"`
	Commits      []string `json:"commits"`
	CommitCount  int      `json:"commit_count"`
}
