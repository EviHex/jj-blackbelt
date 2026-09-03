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

Create or update the stack diagram comment on every PR.

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
