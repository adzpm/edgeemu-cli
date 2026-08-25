# edgeemu-cli

[![release](https://img.shields.io/github/v/release/adzpm/edgeemu-cli?sort=semver)](https://github.com/adzpm/edgeemu-cli/releases/latest)
[![ci](https://github.com/adzpm/edgeemu-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/adzpm/edgeemu-cli/actions/workflows/ci.yml)
[![coverage](https://raw.githubusercontent.com/adzpm/edgeemu-cli/badges/coverage.svg)](https://github.com/adzpm/edgeemu-cli/actions/workflows/ci.yml)
[![go report](https://goreportcard.com/badge/github.com/adzpm/edgeemu-cli)](https://goreportcard.com/report/github.com/adzpm/edgeemu-cli)
[![go version](https://img.shields.io/github/go-mod/go-version/adzpm/edgeemu-cli)](go.mod)
[![license](https://img.shields.io/github/license/adzpm/edgeemu-cli)](LICENSE)

CLI for searching ROMs on [edgeemu.net](https://edgeemu.net).

> **Note:** edgeemu-cli only finds and shows information — what you do with the links is entirely up to you.
> See the [disclaimer](#disclaimer).

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
edgeemu search sonic -f json                  # machine-readable output, see Formats below
```

The default `list` format never truncates anything, and download URLs sit alone on their own line so the terminal always
keeps them clickable:

```
1. Sonic The Hedgehog (USA, Europe)
   system: Sega Mega Drive / Genesis · size: 377.87k · unpacked: 512.00k · dls: 588 · hash: F9394E97
   https://edgeemu.net/[...]/sega-genesis/[...].zip

2. Sonic The Hedgehog 2 (World)
   system: Sega Mega Drive / Genesis · size: 732.08k · unpacked: 1.00m · dls: 432 · hash: 24AB4C3A
   https://edgeemu.net/[...]/sega-genesis/[...].zip
```

Every field is shown by default. Use `-c` to narrow the output to specific fields (see [Columns](#columns) below).
The selection applies to every output format — in `json`, `yaml`, `xml` and `csv` the field IDs are used as keys
(or the header row), and only the selected fields are encoded.

### Columns

All fields available for `-c/--columns`, in their display order:

| ID         | Description                        | Example                                    |
|------------|------------------------------------|--------------------------------------------|
| `name`     | ROM title as listed on the site    | `Sonic The Hedgehog (USA, Europe)`         |
| `system`   | Console / platform                 | `Sega Mega Drive / Genesis`                |
| `size`     | Download (archive) size            | `377.87k`                                  |
| `unpacked` | Unpacked ROM size                  | `512.00k`                                  |
| `dls`      | Download counter                   | `588`                                      |
| `hash`     | CRC hash of the ROM                | `F9394E97`                                 |
| `url`      | Direct download link               | `https://edgeemu.net/download/…/….zip`     |

```sh
edgeemu search sonic -c name,hash             # just names and hashes
edgeemu search sonic -c url -f csv            # a plain list of links
```

The same IDs are suggested by shell completion after `-c`.

### Formats

All values available for `-f/--format`, for both `search` and `systems`:

| ID     | Description                                                                             |
|--------|-----------------------------------------------------------------------------------------|
| `list` | Human-readable list (default): numbered names, fields on one line, URL on its own line  |
| `json` | Indented JSON array                                                                     |
| `yaml` | YAML sequence                                                                           |
| `xml`  | XML document rooted at `<roms>` / `<systems>`                                           |
| `csv`  | Comma-separated values with a header row                                                |

```sh
edgeemu search sonic -f json -c name,dls      # [{"name": "...", "dls": 588}, ...]
edgeemu systems -f yaml                       # - id: atari-2600 ...
```

An empty search result encodes as an empty list (`[]`, `<roms></roms>`, or a lone CSV header), never as `null`.
The same IDs are suggested by shell completion after `-f`.

> Note: the site returns at most 100 results per query. If you hit exactly 100,
> narrow the query or search within a specific system via `-s`.

### Dump

Export the full ROM index — either the whole site or one system:

```sh
edgeemu dump -s sega-genesis -f csv > genesis.csv     # one system
edgeemu dump -f json > edgeemu.json                   # everything (~50 systems, takes a while)
edgeemu dump -s sega-genesis -c name,hash -f csv      # -c works here too
```

Unlike `search`, `dump` is not subject to the 100-result cap: it walks the site's public per-letter browse pages
(`/browse/<system>/<letter>`), which list every entry. Progress is reported on stderr, so redirecting stdout captures
only the data. Between requests the tool pauses (`--delay`, default 150ms) to stay polite to the site; a full dump
makes ~27 requests per system.

### Systems

```sh
edgeemu systems            # list all system IDs for the -s flag
edgeemu systems -f json    # machine-readable output, see Formats above
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
cache), field IDs after `-c` and formats after `-f`. To print the raw completion script instead, use
`edgeemu completion <shell>`.

## Development

Common tasks are defined in the [Taskfile](https://taskfile.dev) (`brew install go-task golangci-lint goreleaser`). CI
runs with pinned tool versions — task 3.53.1, golangci-lint 2.13.1, goreleaser 2.18.0 (see the workflow `env`
blocks); keep local tools reasonably close to those.

```sh
task build        # build with the version stamped from git
task check        # full gate: fmt, vet, lint, tests (also runs in CI)
task test         # tests only
task snapshot     # local goreleaser dry run
task --list       # everything else
```

## Commands

| Command                      | Aliases | Description                       |
|------------------------------|---------|-----------------------------------|
| `search <query>`             | `s`     | Search for ROMs                   |
| `dump`                       |         | Dump the full ROM index           |
| `systems`                    |         | List available systems            |
| `install-completion [shell]` |         | Install shell completion          |

## Flags

| Flag            | Commands              | Description                             |
|-----------------|-----------------------|-----------------------------------------|
| `-s, --system`  | search, dump          | System to use (default: all)            |
| `-l, --limit`   | search                | Max results to show                     |
| `-c, --columns` | search, dump          | Fields to show, see [Columns](#columns) |
| `-f, --format`  | search, dump, systems | Output format, see [Formats](#formats)  |
| `-d, --delay`   | dump                  | Pause between requests (default: 150ms) |
| `-r, --refresh` | systems               | Bypass the cache and refetch            |

## Disclaimer

**What this tool is.** edgeemu-cli is an **information tool**: a convenient search interface over the public search of
edgeemu.net. It queries that search and displays the metadata and links that the site itself publishes — nothing more.

**What this tool is not.** It has **no download functionality**: it does not fetch, host, store, or distribute any ROM
files, and it does not bypass any access controls, paywalls, or copyright protection. Every URL it prints is an
ordinary public link to a third-party website, identical to what you would see in a browser.

**What you do with the links is your call.** You are free to open a link — that request goes from *you* directly to
edgeemu.net, without this tool involved. Whether downloading a particular ROM is lawful depends on your jurisdiction,
the site's terms, and whether you own the original media. That decision, and the responsibility for it, is entirely
yours — the same as with any search engine result.

This project is not affiliated with edgeemu.net or any console manufacturer.
