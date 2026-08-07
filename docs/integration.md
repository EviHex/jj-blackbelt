# Shell & jj integration

## Install `jj stack`

```console
bb util alias
```

This adds a user-level jj alias equivalent to:

```toml
[aliases]
stack = ["util", "exec", "--", "bb", "stack"]
```

It delegates arguments unchanged, so every stack command has both forms:

```console
bb s order
jj stack order
```

## Direct shell completion

=== "Zsh"

    ```zsh
    source <(bb completion zsh)
    ```

=== "Bash"

    ```bash
    source <(bb completion bash)
    ```

=== "Fish"

    ```fish
    bb completion fish | source
    ```

=== "PowerShell"

    ```powershell
    bb completion powershell | Out-String | Invoke-Expression
    ```

## Completion through `jj stack`

jj cannot introspect the nested command model of an external alias. blackbelt therefore generates a small completion bridge for Bash, Zsh, and Fish. Load it after the normal jj and bb completions.

=== "Zsh"

    ```zsh
    source <(COMPLETE=zsh jj)
    source <(bb completion zsh)
    source <(bb completion jj zsh)
    ```

=== "Bash"

    ```bash
    source <(COMPLETE=bash jj)
    source <(bb completion bash)
    source <(bb completion jj bash)
    ```

=== "Fish"

    ```fish
    COMPLETE=fish jj | source
    bb completion fish | source
    bb completion jj fish | source
    ```

The bridge intercepts only `jj stack ...`; normal jj completion remains intact. PowerShell supports direct `bb` completion, but blackbelt does not install an unsafe competing completer for an existing native jj registration.
