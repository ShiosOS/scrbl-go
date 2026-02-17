# scrbl

`scrbl` is a daily notes tool with a terminal UI and optional sync.
It stores notes as markdown day files, uses embedded Neovim for editing, and can
export your latest summary section as Slack-ready markdown to your clipboard.

scrbl serves as my own unique solution for taking
daily notes as a software engineer. I could have made a nvim plugin instead
or a more traditional note taker. However, this flow works more for me as it is
an intentional note taker outside of my dev environment and
makes me want to take notes.

## Features

- Local markdown notes, one file per day (`YYYY-MM-DD.md`)
- Embedded Neovim compose/edit flow inside the TUI
- Optional sync to a small HTTP server backed by SQLite
- `summary` command that copies `## Summary` to clipboard for Slack

## Requirements

- Go `1.24+`
- Neovim (`nvim`) in your `PATH` for TUI editing
- Optional: running `scrbl-server` if you want sync

## Install

From the repository root:

```bash
go install .
```

This installs `scrbl` to your Go bin directory.

## Quick Start

1. Start the server (optional, only needed for sync)
2. Initialize client config
3. Open the TUI

```bash
# 1) start server (from ./server)
go run . -port 8080 -db ./scrbl.db -api-key dev-key

# 2) set client config (from repo root)
scrbl init --server http://localhost:8080 --api-key dev-key

# 3) open TUI
scrbl
```

## CLI Commands

Run `scrbl help` for full usage.

- `scrbl` or `scrbl tui`
  - Open the terminal UI
- `scrbl init --server <url> --api-key <key> [--notes-dir <dir>]`
  - Create or update local config
- `scrbl config show`
  - Print the effective config JSON
- `scrbl migrate [--dry-run] [--sync]`
  - Normalize old note format
  - `--sync` pushes all notes after migration
- `scrbl summary [--date YYYY-MM-DD] [--stdout]`
  - Copy `## Summary` from a day note as Slack markdown
- `scrbl sync push [--date YYYY-MM-DD | --all]`
  - Push local note(s) to the server
- `scrbl sync pull [--date YYYY-MM-DD | --all]`
  - Pull remote note(s) to local files

## TUI Keys

### Stream mode

- `j` / `k` (or arrows): move the stream guide line
- `e` or `Enter`: edit the day under the guide line
- `i`: create a new note entry for today
- `r`: reload notes
- `[` / `]`: jump to previous or next day
- `q`: quit

### Compose mode (embedded Neovim)

Most keys are passed directly to Neovim.

- `:w` or `Ctrl+S`: save
- `:q`: leave compose and return to stream
- `:wq` or `:x`: save and return to stream
- `Ctrl+G`: return to stream
- `Ctrl+C`: quit app

## Note Format

- File name: `YYYY-MM-DD.md`
- First heading for a new day: `# YYYY.MM.DD`
- New entries append as plain markdown blocks (no timestamp headers)

Example:

```md
# 2026.02.17

Worked on stream highlighting and guide behavior.

## Summary

Shipped guide-driven day selection and cleaner status text.
```

## Summary Export for Slack

`scrbl summary` reads the most recent day, extracts the `## Summary` section,
converts common markdown to Slack-friendly markdown, and copies it to clipboard.

Supported summary heading variants include:

- `## Summary`
- `## Daily Summary`

## Config

Default path:

- `~/.scrbl/config.json`

Override path:

- `SCRBL_CONFIG=/path/to/config.json`

Config keys:

```json
{
  "notes_dir": "C:/Users/you/.scrbl/notes",
  "server_url": "http://localhost:8080",
  "api_key": "dev-key"
}
```

## Server

The server lives in the `server/` submodule and exposes:

- `GET /health`
- `GET /api/notes`
- `GET /api/notes/:date`
- `PUT /api/notes/:date`
- `GET /api/search?q=query`

Auth uses `Authorization: Bearer <api_key>` when `API_KEY` is set.

Run server locally:

```bash
cd server
go run . -port 8080 -db ./scrbl.db -api-key dev-key
```

You can also use environment variables:

- `PORT`
- `DB_PATH`
- `API_KEY`

## Development

From project root:

```bash
go fmt ./...
go test ./...
```

For the server module:

```bash
cd server
go fmt ./...
go test ./...
```
