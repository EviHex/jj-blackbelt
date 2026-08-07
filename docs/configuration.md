# Configuration

blackbelt uses TOML and follows a small, predictable precedence chain:

1. Built-in defaults.
2. User configuration.
3. Repository-local configuration.
4. Command-line flags.

## Locations

The user file is `$XDG_CONFIG_HOME/blackbelt/config.toml`. When `XDG_CONFIG_HOME` is unset, macOS and Linux use `~/.config/blackbelt/config.toml`; Windows uses the standard user configuration directory.

Repository settings live at `<jj-repository-root>/.jj/blackbelt.toml` and override user settings.

## Reference

```toml
[stack]
# Command used by bare `bb stack` and `bb s`.
# Supported values: "log", "draw".
default-command = "log"

[stack.log]
# Optional jj revset used as stack-discovery seeds.
# Empty means discover the tree around the current change.
revset = ""

# Show every discovered PR stack by default.
all = false
```

Equivalent flags override these values for one invocation:

```console
bb stack log --revisions 'bookmarks("pinglei.he/.*")' --all
```

!!! note
    A configured revset selects seed revisions. blackbelt still expands those seeds into connected PR trees so the result retains useful stack context.
