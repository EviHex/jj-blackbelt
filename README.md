<div align="center">
  <img src="docs/assets/blackbelt.png" alt="blackbelt mascot: a blue jay wearing a black belt" width="520">

# blackbelt

**Make your jj PR stack obvious to everyone else.**

[Documentation](https://evihex.github.io/jj-blackbelt/) · [Quickstart](https://evihex.github.io/jj-blackbelt/quickstart/) · [Command reference](https://evihex.github.io/jj-blackbelt/commands/)
</div>

`bb` is a lightweight companion to [Jujutsu](https://github.com/jj-vcs/jj). Bring your own jj workflow, we help your reviewers understand what's going on. Keep creating, editing, and rebasing changes in `jj`; blackbelt discovers the PR tree you already made, helps you navigate it, checks its GitHub bases and status, and gives reviewers a clickable stack diagram.

## Quickstart

If you already use `jj` and `gh`, this is the whole setup:

```console
go install github.com/EviHex/jj-blackbelt/cmd/bb@latest
bb doctor
bb util alias
```

Now blackbelt feels like part of jj:

```console
jj stack          # see the PR tree around your current change
jj stack up       # move to the child PR
jj stack order    # diagnose incorrect GitHub bases
jj stack draw     # add or refresh the diagram on every PR
```

There is no repository initialization, stack metadata, parallel branch model, or replacement workflow. One bookmark per PR is enough; your jj graph remains the source of truth.

## What it does

- Discovers linear and tree-shaped PR stacks from jj bookmarks and GitHub PRs.
- Renders the stack in your terminal or as a reviewer-friendly GitHub comment.
- Preserves merged PRs in diagrams even after their bookmarks are forgotten.
- Navigates by PR rather than by commit, including PRs containing multiple jj commits.
- Diagnoses incorrect PR bases and can repair them explicitly with `--fix`.
- Produces JSON for automation and completion for Bash, Zsh, Fish, and PowerShell.
- Works as either `bb stack ...` / `bb s ...` or the installed `jj stack ...` alias.

## Commands

```text
bb doctor
bb stack (s) log [--all] [--json] [--revisions REVSET]
bb stack (s) draw (diagram, d)
bb stack (s) order [--fix] [--all] [--json]
bb stack (s) up [n] [--dry-run] [--json]
bb stack (s) down [n] [--dry-run] [--json]
bb stack (s) top [--dry-run] [--json]
bb stack (s) bottom [--dry-run] [--json]
bb stack (s) goto <PR-number|bookmark> [--dry-run] [--json]
bb util alias [--dry-run]
```

`bb stack` defaults to `bb stack log`; that default and stack selection are configurable in TOML. See the [configuration guide](https://evihex.github.io/jj-blackbelt/configuration/).

## Alternative to jj-spice or jj-stack

[jj-spice](https://github.com/alejoborbo/jj-spice) integrates with jj at a lower level using jj's Rust library and owns stack manipulation. blackbelt is deliberately bring-your-own-workflow: it leaves rebase, split, squash, arrange, submit, and merge operations to `jj` and GitHub operations. Its focus is reviewer-facing diagrams, diagnostics, and PR-aware navigation.

## Development

Requirements: Go 1.24+, jj 0.40+, and an authenticated GitHub CLI.

```bash
go test ./...
go vet ./...
go build -o ~/.local/bin/bb ./cmd/bb
```

We're still in early stage alpha of the development, breaking changes may occur.

## Contribute

Contributions are welcome:
 - Use conventional commit.
 - If you use LLM-assisted coding, please read your own PRs. The PR owner should be able to justify choices made in the PR.
