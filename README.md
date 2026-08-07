# blackbelt

`bb` makes stacked pull requests built with [Jujutsu](https://github.com/jj-vcs/jj) obvious on GitHub. It is a lightweight companion to a bring-your-own jj + `gh` workflow: blackbelt visualizes, validates, and navigates the stack without replacing jj's commit graph operations.

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

`bb stack` defaults to `bb stack log`. The default command is configurable.

### Stack discovery

The default log finds the unique tracked PR/bookmark child below trunk on the ancestry path to `@`, then includes its entire descendant tree. This includes parallel branches of the same stack without absorbing unrelated stacks that also branch directly from trunk. Local-only bookmarks remain intentionally invisible.

`--all` includes all tracked, unmerged PR stacks. `--revisions` uses a jj revset to choose seeds before expanding their connected PR trees.

### Draw

`bb stack draw` creates or updates the marked stack-navigation comment on every PR. Comments contain clickable PR links, statuses, commit summaries, base warnings, merged history, and a per-PR current marker. Updates are idempotent.

### Order

`bb stack order` checks each open PR's GitHub base against its nearest unmerged jj parent. It is read-only and exits unsuccessfully when mismatches exist. `--fix` retargets only those incorrect PRs.

### Navigation

Navigation moves between PR bookmark nodes rather than individual jj commits. This matters when one PR contains multiple commits. `up` prompts at a tree split; non-interactive use reports the choices and suggests `goto`. All navigation commands support `--dry-run`.

## Configuration

See [CONFIGURATION.md](CONFIGURATION.md). blackbelt loads user configuration followed by repository-local `.jj/blackbelt.toml`; command-line flags take precedence.

## jj integration

Install the user-level alias:

```console
bb util alias
jj stack log
jj stack draw
```

The alias delegates `jj stack ...` to `bb stack ...`.

## Completion

Cobra generates completion for Bash, Zsh, Fish, and PowerShell:

```console
bb completion bash
bb completion zsh
bb completion fish
bb completion powershell
```

Because `jj stack` is an external jj alias, jj cannot discover bb's nested command model by itself. Bash, Zsh, and Fish can source a small bridge after loading both normal completion scripts:

```console
bb completion jj bash
bb completion jj zsh
bb completion jj fish
```

PowerShell completion works for direct `bb` usage. A safe composable bridge for an existing native `jj` completer is not currently installed.

## Boundary with jj-spice

[jj-spice](https://github.com/alejoborbo/jj-spice) integrates with jj at a lower level using jj's Rust library and owns stack manipulation. blackbelt intentionally leaves rebase, split, squash, arrange, submit, and merge operations to jj and `gh`. Its focus is reviewer-facing diagrams, diagnostics, and PR-aware navigation.

## Requirements and development

- Go 1.24+
- jj 0.40+
- Authenticated GitHub CLI (`gh auth login`)

```bash
go test ./...
go vet ./...
go build -o ~/.local/bin/bb ./cmd/bb
```
