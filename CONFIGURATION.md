# Configuration

blackbelt reads layered TOML configuration. Later files override values from earlier files:

1. Built-in defaults.
2. User configuration at the platform configuration directory (`~/.config/blackbelt/config.toml` on macOS and Linux when `XDG_CONFIG_HOME` is unset).
3. Repository-local configuration at `<jj-root>/.jj/blackbelt.toml`. This file is private to the local jj repository and is not committed.

```toml
[stack]
# Command run by `bb stack` and `bb s` when no subcommand is supplied.
default-command = "log"

[stack.log]
# Empty means the full tracked PR tree rooted below trunk around @.
revset = ""

# Include every detected stack instead of only the tree around @.
all = false
```

Command-line flags override TOML values. Repository configuration overrides user configuration.
