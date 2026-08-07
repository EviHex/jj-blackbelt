# Concepts

## The graph is already the stack

blackbelt does not maintain a parallel branch database. It combines two sources you already have:

- jj supplies commits, ancestry, bookmarks, and the current change.
- GitHub supplies pull requests, their base branches, status, reviews, comments, and commit comparisons.

One remotely tracked bookmark maps to one PR. A PR may contain multiple jj commits; navigation still treats it as one node.

## Default discovery

From the current change, blackbelt finds the tracked PR/bookmark path back to the repository's real GitHub default branch. It selects the first PR child below trunk and expands through all of that child's PR descendants.

That gives the useful boundary for a tree-shaped stack:

```text
main
└── shared base
    ├── API
    │   └── API tests   ← current change may be here
    └── UI
```

The UI branch is included because it shares the stack root. Another bookmark attached directly to `main` is a different stack and is excluded. Use `bb stack log --all` to display every discovered stack.

Local-only bookmarks are intentionally invisible: without a remote branch and GitHub PR, they are not part of the reviewer-facing PR stack.

## Merged history

When a parent PR merges, GitHub may delete its branch and `jj tidy` may forget its bookmark. `stack draw` embeds topology metadata in its marked GitHub comment and caches it under `.jj/`. On later runs, blackbelt refreshes those historical PRs from GitHub and retains only nodes GitHub confirms are merged.

This means the diagram can keep a merged parent in its original position while base validation correctly expects the next open PR to target trunk.

## What blackbelt does not own

blackbelt does not replace jj's graph operations or dictate how you submit changes. Use jj for rebasing, splitting, squashing, and arranging commits; use your existing `gh` flow for creating and merging PRs.

[jj-spice](https://github.com/alejoborbo/jj-spice) is intentionally lower-level: it uses jj's Rust library and owns stack manipulation. blackbelt aims to be a lightweight, bring-your-own-stacked-PR companion focused on visibility, diagnostics, and navigation.
