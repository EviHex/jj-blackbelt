<section class="hero" markdown>
  <div class="hero-copy" markdown>

# Your jj stack earned a black belt.

blackbelt makes stacked pull requests built with [Jujutsu](https://github.com/jj-vcs/jj) obvious on GitHub—without asking you to replace the workflow you already like.

[Get started in three commands](quickstart.md){ .md-button .md-button--primary }
[See the commands](commands.md){ .md-button }

  </div>
  <div class="hero-art">
    <img src="assets/blackbelt.png" alt="A blue jay in a gi wearing a black belt">
  </div>
</section>

## Bring your own stack

You keep using `jj` to create, edit, split, squash, and rebase changes. One bookmark per pull request is enough. blackbelt reads that graph, resolves the matching GitHub PRs, and adds the context that terminals and reviewers need.

```console
$ jj stack
PR stack — 3 PRs

○  #103  Add the dashboard                         🟡 Draft
│
○  #102  Expose the metrics                        🔵 Reviewed
│
○  #101  Introduce the data model                  🟢 Open  👈
│
◆  main
```

<div class="grid cards" markdown>

-   :material-family-tree:{ .lg .middle } **See the whole PR tree**

    ---

    Discover the stack around your current jj change, including real splits and merged history.

-   :material-draw:{ .lg .middle } **Make GitHub reviewer-friendly**

    ---

    Put one clickable, idempotent diagram on every PR with status, commits, and a “you are here” marker.

-   :material-compass-outline:{ .lg .middle } **Navigate at PR granularity**

    ---

    Move up, down, top, bottom, or directly to a PR—even when a PR contains several jj commits.

-   :material-stethoscope:{ .lg .middle } **Catch a flattened stack**

    ---

    Diagnose incorrect GitHub bases, inspect the proposed correction, then opt into fixing it.

</div>

## One small addition to jj

```console
$ bb util alias
$ jj stack draw
✓ Fetching PR details from GitHub
✓ PR #101: up to date
✓ PR #102: updated
✓ PR #103: updated
```

No import. No repository initialization. No second stack database. The alias simply delegates `jj stack ...` to `bb stack ...`.

!!! tip "blackbelt complements jj"
    jj remains the source of truth and owns your commit graph. blackbelt is the thin PR-awareness layer: discovery, diagrams, diagnostics, and navigation.

[Take the two-minute quickstart →](quickstart.md){ .md-button .md-button--primary }
