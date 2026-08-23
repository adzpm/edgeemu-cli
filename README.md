# edgeemu-cli

CLI for searching ROMs on [edgeemu.net](https://edgeemu.net).

## Install

```sh
go install github.com/adzpm/edgeemu-cli@latest
```

Or build from source:

```sh
go build -o edgeemu .
```

## Usage

### Search

```sh
edgeemu search sonic                          # search all systems
edgeemu search sonic -s sega-genesis          # search a specific system
edgeemu search sonic -l 10                    # limit results
edgeemu search sonic -c name,size,dls,url     # pick columns
edgeemu search sonic --json                   # JSON output (all fields)
```

```
┌───┬──────────────────────────────────┬───────────────────────────┬─────────────────────────────────────────────┐
│ # │               Name               │          System           │                     URL                     │
├───┼──────────────────────────────────┼───────────────────────────┼─────────────────────────────────────────────┤
│ 1 │ Sonic The Hedgehog (USA, Europe) │ Sega Mega Drive / Genesis │ https://edgeemu.net/download/sega-genesis/… │
│ 2 │ Sonic The Hedgehog 2 (World)     │ Sega Mega Drive / Genesis │ https://edgeemu.net/download/sega-genesis/… │
└───┴──────────────────────────────────┴───────────────────────────┴─────────────────────────────────────────────┘
```

Available columns: `name`, `system`, `size`, `unpacked`, `dls`, `hash`, `url`.

By default the table shows `name` and `url`; the `system` column is added only when searching across all systems (no
`-s` flag). Use `-c` to pick any set.

> Note: the site returns at most 100 results per query. If you hit exactly 100,
> narrow the query or search within a specific system via `-s`.

### Systems

```sh
edgeemu systems       # list all system IDs for the -s flag
edgeemu systems -r    # refresh the cached list
```

The list is cached for 24 hours in the user cache directory (`~/Library/Caches/edgeemu` on macOS, `~/.cache/edgeemu` on
Linux).

### Shell completion

```sh
edgeemu install-completion           # detects your shell from $SHELL
edgeemu install-completion zsh       # or name it explicitly (zsh, bash, fish)
```

Adds a completion hook to your shell rc file. Completes commands, flags, system IDs after `-s` (instantly, from the
cache) and column IDs after `-c`. To print the raw completion script instead, use `edgeemu completion <shell>`.

## Commands

| Command                      | Aliases | Description              |
|------------------------------|---------|--------------------------|
| `search <query>`             | `s`     | Search for ROMs          |
| `systems`                    |         | List available systems   |
| `install-completion [shell]` |         | Install shell completion |

## Flags

| Flag            | Commands | Description                        |
|-----------------|----------|------------------------------------|
| `-s, --system`  | search   | System to search in (default: all) |
| `-l, --limit`   | search   | Max results to show                |
| `-c, --columns` | search   | Comma-separated columns to show    |
| `--json`        | search   | Output results as JSON             |
| `-r, --refresh` | systems  | Bypass the cache and refetch       |
