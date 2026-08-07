# blackbelt

`blackbelt` keeps a stacked pull-request story legible on GitHub for people using [Jujutsu](https://github.com/jj-vcs/jj).

It discovers the tracked bookmarks connected to your current jj stack, resolves their GitHub pull requests, and creates or updates the same navigation comment on every PR. The comment is a compact, clickable graph with PR state, up to three commit subjects, merged ancestors, and base-branch warnings.

```console
$ jj blackbelt --dry-run
PR stack — 3 PRs

○  #103  Add the API                                      🔵 Reviewed  👈
│   • feat: add the API
│
◆  #102  Introduce the model                              🟣 Merged
│   • feat: introduce the model
│
◆  prod
```

## What it does

- Finds the connected tree of tracked `origin` bookmarks around `@`; local-only bookmarks stay invisible.
- Handles both linear stacks and branching trees.
- Creates or updates one marked stack comment per PR; reruns are idempotent.
- Preserves merged PRs from encoded comment history, even after their bookmarks no longer exist.
- Shows draft, open, reviewed, merged, and closed state; merged PRs use an immutable `◆` node.
- Validates each open PR's GitHub base against its nearest unmerged jj parent and reports mismatches.
- Uses GitHub's comparison commits, rather than assuming one jj commit per bookmark.
- Provides a colored, OSC-8-link terminal preview with `--dry-run`.

## What it is not

blackbelt is deliberately lightweight. It does not create, rebase, submit, or merge pull requests; jj and `gh` remain your workflow tools. It also does not implement a replacement jj engine or a new stacking model.

[jj-spice](https://github.com/alejoborbo/jj-spice) operates at a lower level, using jj's Rust library to provide first-class stack manipulation. blackbelt instead brings its own minimal discovery layer to an existing jj + GitHub workflow and makes the resulting stack obvious to reviewers on GitHub.

## Requirements

- Go 1.24+
- [`jj`](https://github.com/jj-vcs/jj)
- [`gh`](https://cli.github.com/) authenticated for the target repository

## Run

During development:

```bash
go run ./cmd/blackbelt --dry-run
go run ./cmd/blackbelt
```

Build a personal binary:

```bash
go build -o ~/.local/bin/blackbelt ./cmd/blackbelt
```

Then wire it into jj:

```toml
[aliases]
blackbelt = ["util", "exec", "--", "blackbelt"]
bb = ["blackbelt"]
```

`jj blackbelt` updates comments; `jj blackbelt --dry-run` prints one clickable terminal diagram without changing GitHub.

## Development

```bash
go test ./...
go vet ./...
```

The git-ignored `FUTURE.md` is intentionally a local product backlog, not a promise or an implementation plan.
