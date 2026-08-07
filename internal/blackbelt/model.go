// Package blackbelt discovers jj PR stacks and maintains their GitHub notes.
package blackbelt

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
