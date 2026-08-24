# edgeemu-cli

[![release](https://img.shields.io/github/v/release/adzpm/edgeemu-cli?sort=semver)](https://github.com/adzpm/edgeemu-cli/releases/latest)
[![ci](https://github.com/adzpm/edgeemu-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/adzpm/edgeemu-cli/actions/workflows/ci.yml)
[![coverage](https://raw.githubusercontent.com/adzpm/edgeemu-cli/badges/coverage.svg)](https://github.com/adzpm/edgeemu-cli/actions/workflows/ci.yml)
[![go report](https://goreportcard.com/badge/github.com/adzpm/edgeemu-cli)](https://goreportcard.com/report/github.com/adzpm/edgeemu-cli)
[![go version](https://img.shields.io/github/go-mod/go-version/adzpm/edgeemu-cli)](go.mod)
[![license](https://img.shields.io/github/license/adzpm/edgeemu-cli)](LICENSE)

CLI for searching ROMs on [edgeemu.net](https://edgeemu.net).

## Install

Homebrew (macOS / Linux):

```sh
brew install adzpm/tap/edgeemu
```

Debian / Ubuntu — grab the `.deb` from the [latest release](https://github.com/adzpm/edgeemu-cli/releases/latest):

```sh
sudo dpkg -i edgeemu_*_linux_amd64.deb
```

Go:

```sh
go install github.com/adzpm/edgeemu-cli/cmd/edgeemu@latest
```

Or build from source:

```sh
go build -o edgeemu ./cmd/edgeemu
```

Prebuilt binaries for macOS, Linux and Windows (amd64/arm64) are attached to
every [release](https://github.com/adzpm/edgeemu-cli/releases).

## Usage

### Search

```sh
edgeemu search sonic                          # search all systems
edgeemu search sonic -s sega-genesis          # search a specific system
edgeemu search sonic -l 10                    # limit results
edgeemu search sonic -c name,size,dls,url     # pick fields to show
edgeemu search sonic -f json                  # machine-readable output (json, yaml, xml or csv)
```

The default `list` format never truncates anything, and download URLs sit alone on their own line so the terminal always
keeps them clickable:

```
1. Sonic The Hedgehog (USA, Europe)
   system: Sega Mega Drive / Genesis · size: 377.87k · unpacked: 512.00k · dls: 588 · hash: F9394E97
   https://edgeemu.net/download/sega-genesis/Sonic%20The%20Hedgehog%20%28USA%2C%20Europe%29.zip

2. Sonic The Hedgehog 2 (World)
   system: Sega Mega Drive / Genesis · size: 732.08k · unpacked: 1.00m · dls: 432 · hash: 24AB4C3A
   https://edgeemu.net/download/sega-genesis/Sonic%20The%20Hedgehog%202%20%28World%29.zip
```

Every field is shown by default. Use `-c` to narrow the output to specific fields: `name`, `system`, `size`, `unpacked`,
`dls`, `hash`, `url`. The selection applies to every output format — in `json`, `yaml`, `xml` and `csv` the field IDs
are used as keys (or the header row), and only the selected fields are encoded.

> Note: the site returns at most 100 results per query. If you hit exactly 100,
> narrow the query or search within a specific system via `-s`.

### Systems

```sh
edgeemu systems            # list all system IDs for the -s flag
edgeemu systems -f json    # machine-readable output (json, yaml, xml or csv)
edgeemu systems -r         # refresh the cached list
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

## Development

Common tasks are defined in the [Taskfile](https://taskfile.dev) (`brew install go-task golangci-lint`):

```sh
task build        # build with the version stamped from git
task check        # full gate: fmt, vet, lint, tests (also runs in CI)
task test         # tests only
task snapshot     # local goreleaser dry run
task --list       # everything else
```

## Commands

| Command                      | Aliases | Description              |
|------------------------------|---------|--------------------------|
| `search <query>`             | `s`     | Search for ROMs          |
| `systems`                    |         | List available systems   |
| `install-completion [shell]` |         | Install shell completion |

## Flags

| Flag            | Commands        | Description                            |
|-----------------|-----------------|----------------------------------------|
| `-s, --system`  | search          | System to search in (default: all)     |
| `-l, --limit`   | search          | Max results to show                    |
| `-c, --columns` | search          | Comma-separated fields to show         |
| `-f, --format`  | search, systems | Output format: list, json, yaml, xml or csv |
| `-r, --refresh` | systems         | Bypass the cache and refetch           |
