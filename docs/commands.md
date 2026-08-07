# Command reference

`stack` has the shorthand `s`. Running `bb stack` or `bb s` without a subcommand runs the configured default command, initially `log`.

## `bb doctor`

Check the installed jj and GitHub CLI versions, GitHub authentication, and repository context.

```console
bb doctor [--json]
```

## `bb stack log`

Render the PR tree around the current change without modifying GitHub.

```console
bb stack log [--all] [--json] [--revisions REVSET]
```

- `--all`, `-a`: render all tracked PR stacks as separate trees.
- `--json`: emit the repository, trunk, current PR, topology, status, bases, and commits as JSON.
- `--revisions`, `-r`: select stack seeds with a jj revset before connected-tree expansion.

## `bb stack draw`

Create or update the stack-navigation comment on every PR.

```console
bb stack draw
bb stack diagram
bb stack d
```

The update is idempotent. An unchanged comment is reported as up to date. GitHub calls show interactive progress in a terminal.

## `bb stack order`

Compare each open PR's GitHub base with its nearest unmerged jj parent, or trunk when all structural parents have merged.

```console
bb stack order [--all] [--fix] [--json]
```

The default is read-only and exits unsuccessfully when mismatches exist. `--fix` explicitly retargets only mismatched PRs.

## Navigation

Navigation edits the jj commit at a PR bookmark. It skips commits inside a multi-commit PR.

```console
bb stack up [n] [--dry-run] [--json]
bb stack down [n] [--dry-run] [--json]
bb stack top [--dry-run] [--json]
bb stack bottom [--dry-run] [--json]
bb stack goto <PR-number|bookmark> [--dry-run] [--json]
```

- `up`: move toward a child PR; prompts when the tree splits.
- `down`: move toward the parent PR.
- `top`: follow the current branch to a topmost PR; prompts at splits.
- `bottom`: move to the lowest still-live PR in the stack.
- `goto`: accept `123`, `#123`, or a bookmark name.
- `--dry-run`, `-n`: print the resolved destination without calling `jj edit`.

In non-interactive use, an ambiguous split reports the available children rather than guessing.

## `bb util alias`

Install the user-level `jj stack` delegation alias.

```console
bb util alias [--dry-run]
```

Use `--dry-run` to inspect the `jj config set` command without changing configuration.

## Completion

```console
bb completion bash
bb completion zsh
bb completion fish
bb completion powershell
bb completion jj bash
bb completion jj zsh
bb completion jj fish
```

See [Shell & jj integration](integration.md) for setup examples and the `jj stack` completion bridge.
