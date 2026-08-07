# Quickstart

If you already have `jj` and an authenticated `gh`, blackbelt takes three commands to install and wire into your workflow.

## 1. Install

```console
go install github.com/pinglei-he/blackbelt/cmd/bb@latest
```

Make sure Go's binary directory is on your `PATH` (usually `~/go/bin`). blackbelt currently requires Go 1.24+, jj 0.40+, and the GitHub CLI.

## 2. Check the environment

```console
bb doctor
```

`doctor` checks the jj version, the GitHub CLI, GitHub authentication, and whether the current directory is a jj repository. It does not change anything.

## 3. Make it feel native

```console
bb util alias
```

That installs one user-level jj alias. From now on, both spellings work:

```console
bb stack
jj stack
```

## Use the stack you already have

There is nothing to initialize. Start anywhere inside an existing stack and ask blackbelt what it sees:

```console
jj stack
```

blackbelt finds the unique PR root between trunk and your current change, then includes that root's descendants. A split in the jj graph becomes a split in the PR tree; unrelated stacks branching from trunk stay out.

Useful next moves:

```console
jj stack up             # edit the child PR
jj stack down           # edit the parent PR
jj stack goto 1234      # edit PR #1234
jj stack order          # check GitHub PR bases
jj stack draw           # update the diagram on every PR
```

Navigation acts on PR bookmarks, not individual commits. Add `--dry-run` to a navigation command when you only want to see its destination.

## What `draw` posts

Each PR gets the same connected stack diagram with clickable PR numbers, live status, commit summaries, and base warnings. Only the `👈` marker changes to identify the PR being viewed. Re-running the command updates the existing marked comment instead of adding another one.

Merged ancestors stay visible, even after their bookmark is forgotten, as long as a surviving stack comment contains blackbelt's hidden topology metadata.

## That's it

Keep arranging commits with jj and opening PRs with your existing `gh` workflow. Reach for blackbelt when you want to see, navigate, validate, or explain the PR stack.

[Learn the stack model](concepts.md){ .md-button }
[Browse every command](commands.md){ .md-button .md-button--primary }
