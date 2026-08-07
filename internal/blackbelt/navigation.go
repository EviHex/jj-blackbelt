package blackbelt

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// NavigateOptions controls movement between PR bookmark nodes.
type NavigateOptions struct {
	Direction string
	Target    string
	Steps     int
	DryRun    bool
	JSON      bool
	Revset    string
}

// Navigate moves the jj working copy to another PR bookmark in the stack.
func Navigate(ctx context.Context, options NavigateOptions) error {
	return navigate(ctx, shell{}, options, os.Stdin, os.Stdout)
}

func navigate(ctx context.Context, r runner, options NavigateOptions, in, out *os.File) error {
	_, _, _, prs, _, _, err := loadStack(ctx, r, Options{Revset: options.Revset})
	if err != nil {
		return err
	}
	current := currentNumber(ctx, r, prs)
	if current == 0 && options.Direction != "goto" {
		return errors.New("current change is not inside a live PR in this stack")
	}
	choose := func(candidates []PullRequest) (PullRequest, error) {
		if len(candidates) == 1 {
			return candidates[0], nil
		}
		if !terminal(in) || !terminal(out) {
			return PullRequest{}, fmt.Errorf("multiple PRs are available (%s); use bb stack goto", candidateNames(candidates))
		}
		fmt.Fprintln(out, "Choose a PR:")
		for index, candidate := range candidates {
			fmt.Fprintf(out, "  %d. #%d  %s (%s)\n", index+1, candidate.Number, candidate.Title, candidate.Head)
		}
		fmt.Fprint(out, "> ")
		line, err := bufio.NewReader(in).ReadString('\n')
		if err != nil {
			return PullRequest{}, err
		}
		index, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil || index < 1 || index > len(candidates) {
			return PullRequest{}, errors.New("invalid selection")
		}
		return candidates[index-1], nil
	}
	target, err := navigationTarget(prs, current, options, choose)
	if err != nil {
		return err
	}
	if options.JSON {
		return json.NewEncoder(out).Encode(struct {
			Number   int    `json:"number"`
			Bookmark string `json:"bookmark"`
			CommitID string `json:"commit_id"`
		}{target.Number, target.Head, target.CommitID})
	}
	if options.DryRun {
		fmt.Fprintf(out, "#%d  %s (%s)\n", target.Number, target.Title, target.Head)
		return nil
	}
	if target.CommitID == "" {
		return fmt.Errorf("PR #%d no longer has a local bookmark commit", target.Number)
	}
	if _, err := must(ctx, r, "jj", "edit", target.CommitID); err != nil {
		return err
	}
	fmt.Fprintf(out, "Now at #%d  %s (%s)\n", target.Number, target.Title, target.Head)
	return nil
}

func navigationTarget(prs []PullRequest, current int, options NavigateOptions, choose func([]PullRequest) (PullRequest, error)) (PullRequest, error) {
	byNumber := map[int]PullRequest{}
	children := map[int][]PullRequest{}
	for _, pr := range prs {
		byNumber[pr.Number] = pr
		if pr.Parent != nil {
			children[*pr.Parent] = append(children[*pr.Parent], pr)
		}
	}
	for parent := range children {
		sort.Slice(children[parent], func(i, j int) bool { return children[parent][i].Number < children[parent][j].Number })
	}
	if options.Direction == "goto" {
		needle := strings.TrimPrefix(options.Target, "#")
		for _, pr := range prs {
			if fmt.Sprint(pr.Number) == needle || pr.Head == options.Target {
				return pr, nil
			}
		}
		return PullRequest{}, fmt.Errorf("no PR or bookmark %q in the current stack", options.Target)
	}
	position, ok := byNumber[current]
	if !ok {
		return PullRequest{}, errors.New("cannot resolve the current PR")
	}
	steps := options.Steps
	if steps < 1 {
		steps = 1
	}
	switch options.Direction {
	case "up":
		for range steps {
			candidates := children[position.Number]
			if len(candidates) == 0 {
				return PullRequest{}, fmt.Errorf("PR #%d is already at the top of this path", position.Number)
			}
			var err error
			position, err = choose(candidates)
			if err != nil {
				return PullRequest{}, err
			}
		}
		return position, nil
	case "down":
		for range steps {
			if position.Parent == nil || byNumber[*position.Parent].Number == 0 {
				return PullRequest{}, fmt.Errorf("PR #%d is already at the bottom of this stack", position.Number)
			}
			position = byNumber[*position.Parent]
		}
		return position, nil
	case "top":
		for {
			candidates := children[position.Number]
			if len(candidates) == 0 {
				return position, nil
			}
			var err error
			position, err = choose(candidates)
			if err != nil {
				return PullRequest{}, err
			}
		}
	case "bottom":
		for position.Parent != nil && byNumber[*position.Parent].Number != 0 {
			position = byNumber[*position.Parent]
		}
		return position, nil
	default:
		return PullRequest{}, fmt.Errorf("unknown navigation direction %q", options.Direction)
	}
}

func candidateNames(prs []PullRequest) string {
	values := make([]string, len(prs))
	for i, pr := range prs {
		values[i] = fmt.Sprintf("#%d %s", pr.Number, pr.Head)
	}
	return strings.Join(values, ", ")
}
