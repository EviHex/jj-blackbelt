package blackbelt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

func Run(ctx context.Context, options Options) error {
	return run(ctx, shell{}, options, os.Stdout, os.Stderr)
}
func run(ctx context.Context, r runner, options Options, out, errOut *os.File) error {
	root, repo, trunk, local, comments, historical, err := loadStack(ctx, r, options)
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
		if options.JSON {
			return writeStackJSON(out, repo, trunk, current, local)
		}
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

func loadStack(ctx context.Context, r runner, options Options) (string, string, string, []PullRequest, map[int][]map[string]any, map[int]PullRequest, error) {
	root := strings.TrimSpace(runCmd(ctx, r, "jj", "--ignore-working-copy", "root"))
	if root == "" {
		return "", "", "", nil, nil, nil, errors.New("run this command inside a jj repository")
	}
	repo, trunk, err := metadata(ctx, r)
	if err != nil {
		return "", "", "", nil, nil, nil, err
	}
	bookmarks, err := connected(ctx, r, trunk, options.Revset, options.All)
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

type jsonStack struct {
	Repository string        `json:"repository"`
	Trunk      string        `json:"trunk"`
	Current    int           `json:"current,omitempty"`
	PRs        []PullRequest `json:"prs"`
}

func writeStackJSON(out *os.File, repo, trunk string, current int, prs []PullRequest) error {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(jsonStack{Repository: repo, Trunk: trunk, Current: current, PRs: prs})
}
