# codex-su

A command-line tool that lets you manage and switch between multiple Codex accounts.

> [!WARNING]
> Not affiliated with OpenAI or Codex. Not an official tool.

## How it Works

Codex stores your authentication session in a single `auth.json` file. This tool works by creating named snapshots of that file for each of your accounts. When you want to switch, `codex-su` swaps the active `~/.codex/auth.json` with the snapshot you select, instantly changing your logged-in account.

## Requirements

- Node.js 18 or newer

## Install (npm)

```sh
npm i -g codex-su
```

## Usage

```sh
# save the current logged-in token as a named account
codex-su save <name>

# switch active account (symlinks on macOS/Linux; copies on Windows)
codex-su use <name>

# or pick interactively
codex-su use

# list accounts
codex-su list

# show current account name
codex-su current
```

### Command reference

- `codex-su save <name>` - validates `<name>`, ensures `auth.json` exists, then snapshots it to `~/.codex/accounts/<name>.json`.
- `codex-su use [name]` - accepts a name or launches an interactive selector with the current account pre-selected. Copies on Windows, creates a symlink elsewhere, and records the active name.
- `codex-su list` - lists all saved snapshots alphabetically and marks the active one with `*`.
- `codex-su current` - prints the active account name, or a friendly message if none is active.

Notes:

- Works on macOS/Linux (symlink) and Windows (copy).
- Requires Node 18+.
- Set `CODEX_HOME` to use a nonstandard Codex config directory for tests, automation, or isolated environments.
