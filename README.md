# edgeemu-cli

[![ci](https://github.com/adzpm/edgeemu-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/adzpm/edgeemu-cli/actions/workflows/ci.yml)
[![go version](https://img.shields.io/github/go-mod/go-version/adzpm/edgeemu-cli)](go.mod)
[![release](https://img.shields.io/github/v/release/adzpm/edgeemu-cli?sort=semver)](https://github.com/adzpm/edgeemu-cli/releases/latest)
[![coverage](https://raw.githubusercontent.com/adzpm/edgeemu-cli/badges/coverage.svg)](https://github.com/adzpm/edgeemu-cli/actions/workflows/ci.yml)
[![go report](https://goreportcard.com/badge/github.com/adzpm/edgeemu-cli)](https://goreportcard.com/report/github.com/adzpm/edgeemu-cli)
[![license](https://img.shields.io/github/license/adzpm/edgeemu-cli)](LICENSE)

A fast command-line tool for searching the ROM catalog of [edgeemu.net](https://edgeemu.net):
find games, filter by console, export the index to JSON/YAML/XML/CSV.

> **Note:** edgeemu-cli only finds and shows information — what you do with the links is entirely up to you.
> See the [disclaimer](#disclaimer).

---

## Table of contents

- [Installation](#installation)
- [Quick start](#quick-start)
- [Commands](#commands)
    - [`search` — find ROMs](#search--find-roms)
    - [`systems` — list consoles](#systems--list-consoles)
    - [`dump` — export the full index](#dump--export-the-full-index)
    - [`install-completion` — enable Tab hints](#install-completion--enable-tab-hints)
- [Reference](#reference)
    - [Flags](#flags)
    - [Columns](#columns)
    - [Formats](#formats)
- [FAQ](#faq)
- [Development](#development)
- [Disclaimer](#disclaimer)

---

## Installation

Pick whichever way suits you — all of them install the same `edgeemu` binary.

**Homebrew** (macOS / Linux, recommended):

```sh
brew install adzpm/tap/edgeemu
```

**Debian / Ubuntu** — download the `.deb` from
the [latest release](https://github.com/adzpm/edgeemu-cli/releases/latest), then:

```sh
sudo dpkg -i edgeemu_*_linux_amd64.deb
```

**Go** (any OS with a Go toolchain):

```sh
go install github.com/adzpm/edgeemu-cli/cmd/edgeemu@latest
```

**Prebuilt binaries** for macOS, Linux and Windows (amd64/arm64) are attached to every
[release](https://github.com/adzpm/edgeemu-cli/releases). To build from source: `go build -o edgeemu ./cmd/edgeemu`.

Check that it works:

```sh
edgeemu --version
```

---

## Quick start

Three commands cover 90% of everyday use.

**1. Find a game** (searches every console at once):

```sh
edgeemu search "chrono trigger"
```

```
1. Chrono Trigger (USA)
   system: Super NES (SNES) · size: 2.93m · unpacked: 4.00m · dls: 1018 · hash: 2D206BF7
   https://edgeemu.net/[...]/[...].zip
2. Chrono Trigger (Japan)
   ...
```

**2. See which consoles exist** (their IDs go into the `-s` flag):

```sh
edgeemu systems
```

```
01. Atari 2600                          · id: atari-2600
02. Atari 5200                          · id: atari-5200
...
48. Watara Supervision                  · id: watara-supervision
```

**3. Search one console only:**

```sh
edgeemu search sonic -s sega-genesis
```

That's it. Everything else — output formats, field selection, full exports — is described below.

---

## Commands

### `search` — find ROMs

```
edgeemu search [flags] <query>
```

Searches the site and prints what it finds. The words of the query can go anywhere in the title.

| Flag            | What it does                                             | Default      |
|-----------------|----------------------------------------------------------|--------------|
| `-s, --system`  | Search one console only ([IDs](#systems--list-consoles)) | all consoles |
| `-l, --limit`   | Show at most N results                                   | all          |
| `-c, --columns` | Show only these [fields](#columns)                       | all fields   |
| `-f, --format`  | Output [format](#formats)                                | `list`       |

Examples:

```sh
edgeemu search sonic -s sega-genesis -l 5     # top 5 Sonic games on Genesis
edgeemu search sonic -c name,hash             # names and hashes only
edgeemu search sonic -f json                  # JSON for scripts
```

> The site returns **at most 100 results per query**. If you hit exactly 100, narrow the query,
> add `-s`, or use [`dump`](#dump--export-the-full-index) — it has no such cap.

### `systems` — list consoles

```
edgeemu systems [flags]
```

Prints every console the site knows, with the ID you pass to `-s`:

```
01. Atari 2600                          · id: atari-2600
02. Atari 5200                          · id: atari-5200
...
```

| Flag            | What it does                                    | Default |
|-----------------|-------------------------------------------------|---------|
| `-f, --format`  | Output [format](#formats)                       | `list`  |
| `-r, --refresh` | Re-download the list instead of using the cache | off     |

The list changes rarely, so it is cached for 24 hours (`~/Library/Caches/edgeemu` on macOS, `~/.cache/edgeemu` on
Linux). The cache also powers instant Tab completion.

### `dump` — export the full index

```
edgeemu dump [flags]
```

Exports the **entire** catalog — every game of every console (or just one via `-s`) — in one file:

```sh
edgeemu dump -s sega-genesis -f csv > genesis.csv     # one console (~4 seconds)
edgeemu dump -f json > edgeemu.json                   # the whole site (~50 consoles, takes minutes)
```

| Flag            | What it does                                       | Default      |
|-----------------|----------------------------------------------------|--------------|
| `-s, --system`  | Dump one console only                              | all consoles |
| `-c, --columns` | Export only these [fields](#columns)               | all fields   |
| `-f, --format`  | Output [format](#formats)                          | `list`       |
| `-d, --delay`   | Pause between requests, to stay polite to the site | `150ms`      |

While it runs, a progress bar per console is drawn on stderr — so `> file.json` captures only the data:

```
[09/48] colecovision         [████████████████████████] 100% · ✓ · 197 entries
[10/48] commodore-64         [██████████░░░░░░░░░░░░░░]  40% · h · 1502 entries
```

How it works: unlike `search`, `dump` walks the site's public per-letter browse pages (`/browse/<system>/<letter>`),
which are not capped at 100 entries. A full dump makes ~27 requests per console.

### `install-completion` — enable Tab hints

```
edgeemu install-completion [zsh|bash|fish]
```

One-time setup: adds a completion hook to your shell config (detects the shell from `$SHELL` if not named). Restart the
terminal afterwards. Then the Tab key completes:

- commands and flags,
- console IDs after `-s` (instantly, from the cache),
- field IDs after `-c` and formats after `-f`.

Running it twice is safe — it detects an existing install. To print the raw completion script instead, use
`edgeemu completion <shell>`.

---

## Reference

### Flags

All flags at a glance:

| Flag            | Commands              | Description                             |
|-----------------|-----------------------|-----------------------------------------|
| `-s, --system`  | search, dump          | Console to use (default: all)           |
| `-l, --limit`   | search                | Max results to show                     |
| `-c, --columns` | search, dump          | Fields to show, see [Columns](#columns) |
| `-f, --format`  | search, dump, systems | Output format, see [Formats](#formats)  |
| `-d, --delay`   | dump                  | Pause between requests (default: 150ms) |
| `-r, --refresh` | systems               | Bypass the cache and refetch            |

### Columns

Fields available for `-c/--columns`, in their display order:

| ID         | Description                     | Example                               |
|------------|---------------------------------|---------------------------------------|
| `name`     | ROM title as listed on the site | `Sonic The Hedgehog (USA, Europe)`    |
| `system`   | Console / platform              | `Sega Mega Drive / Genesis`           |
| `size`     | Download (archive) size         | `377.87k`                             |
| `unpacked` | Unpacked ROM size               | `512.00k`                             |
| `dls`      | Download counter                | `588`                                 |
| `hash`     | CRC hash of the ROM             | `F9394E97`                            |
| `url`      | Direct download link            | `https://edgeemu.net/[...]/[...].zip` |

Every field is shown by default; `-c` narrows the set. The selection applies to **every** output format:
in `json`, `yaml`, `xml` and `csv` the field IDs become the keys (or the CSV header), and only the selected fields are
encoded.

```sh
edgeemu search sonic -c name,hash             # just names and hashes
edgeemu search sonic -c url -f csv            # a plain list of links
```

### Formats

Values for `-f/--format`, accepted by `search`, `dump` and `systems`:

| ID     | Description                                                                            |
|--------|----------------------------------------------------------------------------------------|
| `list` | Human-readable list (default): numbered names, fields on one line, URL on its own line |
| `json` | Indented JSON array                                                                    |
| `yaml` | YAML sequence                                                                          |
| `xml`  | XML document rooted at `<roms>` / `<systems>`                                          |
| `csv`  | Comma-separated values with a header row                                               |

```sh
edgeemu search sonic -f json -c name,dls      # [{"name": "...", "dls": 588}, ...]
edgeemu systems -f yaml                       # - id: atari-2600 ...
```

An empty result encodes as an empty list (`[]`, `<roms></roms>`, or a lone CSV header) — never as `null`.

---

## FAQ

**Why do I get exactly 100 results?**
That is the site's own cap on search. Narrow the query, limit it to one console with `-s`, or use `dump` — it reads the
browse pages, which have no cap.

**Why is the first Tab completion after `-s` slow?**
The console list is fetched once and cached for 24 hours. The first completion fills the cache (up to 3 seconds); every
one after that is instant.

**Can this download ROMs?**
No. There is no download command, on purpose — see the [disclaimer](#disclaimer). The `url` field is plain text; opening
it is your own call.

**A full dump is slow. Can I speed it up?**
The pauses are deliberate politeness towards the site (`-d 150ms` by default). You can lower them (`-d 50ms`), but
please don't hammer the site — the data doesn't change that often.

---

## Development

Common tasks live in the [Taskfile](https://taskfile.dev) (`brew install go-task golangci-lint goreleaser`). CI runs
with pinned tool versions — task 3.53.1, golangci-lint 2.13.1, goreleaser 2.18.0 (see the workflow `env` blocks); keep
local tools reasonably close to those.

```sh
task build        # build with the version stamped from git
task check        # full gate: fmt, vet, lint, tests (also runs in CI)
task test         # tests only
task snapshot     # local goreleaser dry run
task --list       # everything else
```

Project layout: `cmd/edgeemu` is the entry point; everything else is in `internal/`
(`client` — HTTP + parsing, `cache` — systems cache, `render` — output formats, `completion` — shell completion, `app` —
CLI wiring).

---

## Disclaimer

**What this tool is.** edgeemu-cli is an **information tool**: a convenient search interface over the public search of
edgeemu.net. It queries that search and displays the metadata and links that the site itself publishes — nothing more.

**What this tool is not.** It has **no download functionality**: it does not fetch, host, store, or distribute any ROM
files, and it does not bypass any access controls, paywalls, or copyright protection. Every URL it prints is an ordinary
public link to a third-party website, identical to what you would see in a browser.

**What you do with the links is your call.** You are free to open a link — that request goes from *you* directly to
edgeemu.net, without this tool involved. Whether downloading a particular ROM is lawful depends on your jurisdiction,
the site's terms, and whether you own the original media. That decision, and the responsibility for it, is entirely
yours — the same as with any search engine result.

This project is not affiliated with edgeemu.net or any console manufacturer.

## License

[MIT](LICENSE)
