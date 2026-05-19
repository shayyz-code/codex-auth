<h1 align="center">codex-auth</h1>

<p align="center">
  <a href="https://github.com/shayyz-code/codex-auth/actions/workflows/ci.yml"><img alt="CI" src="https://img.shields.io/github/actions/workflow/status/shayyz-code/codex-auth/ci.yml?style=flat-square"></a>
  <a href="https://github.com/shayyz-code/codex-auth/releases"><img alt="GitHub release" src="https://img.shields.io/github/v/release/shayyz-code/codex-auth?sort=semver&style=flat-square"></a>
  <a href="https://www.npmjs.com/package/@shayyz-code/codex-auth"><img alt="npm" src="https://img.shields.io/npm/v/%40shayyz-code%2Fcodex-auth?style=flat-square"></a>
  <a href="https://github.com/shayyz-code/codex-auth/blob/main/LICENSE"><img alt="License" src="https://img.shields.io/github/license/shayyz-code/codex-auth?style=flat-square"></a>
</p>

<p align="center">A command-line tool that lets you manage and switch between multiple Codex accounts.</p>

> [!WARNING]
> Not affiliated with OpenAI or Codex. Not an official tool.

## How it Works

Codex stores your authentication session in a single `auth.json` file. This tool works by creating named snapshots of that file for each of your accounts. When you want to switch, `codex-auth` swaps the active `~/.codex/auth.json` with the snapshot you select, instantly changing your logged-in account.

## Requirements

- Go 1.24 or newer to build from source.
- Homebrew or npm if you install through those package managers.

## Installation

### Homebrew

```sh
brew tap shayyz-code/tap
brew install codex-auth
```

### npm

```sh
npm install -g @shayyz-code/codex-auth
```

### GitHub Releases

Download the binary for your platform from the [latest release](https://github.com/shayyz-code/codex-auth/releases/latest), then put it somewhere on your `PATH`.

### Go install

```sh
go install github.com/shayyz-code/codex-auth/cmd/codex-auth@latest
```

### Build locally

```sh
go build -o bin/codex-auth ./cmd/codex-auth
```

## Usage

```sh
# save the current logged-in token as a named account
codex-auth save <name>

# switch active account (symlinks on macOS/Linux; copies on Windows)
codex-auth use <name>

# or pick interactively
codex-auth use

# list accounts
codex-auth list

# show current account name
codex-auth current
```

### Command reference

- `codex-auth save <name>` - validates `<name>`, ensures `auth.json` exists, then snapshots it to `~/.codex/accounts/<name>.json`. The requested name is always honored, so `save new-account` writes that account even if the active auth matches another saved snapshot.
- `codex-auth use [name]` - accepts a name or launches an interactive selector with the current account pre-selected. On startup, the live Codex `auth.json` is matched against saved snapshots and `current` is refreshed before commands run. If `<name>` is mistyped, the closest saved account is suggested. Before switching away from an unsaved live Codex login, interactive mode asks whether to save it. Copies on Windows, creates a symlink elsewhere, and records the active name. Interactive terminal output uses color when supported; piped output and `--json` remain stable for automation.
- `codex-auth list` - lists all saved snapshots alphabetically and marks the active one with `*`.
- `codex-auth current` - prints the active account name, or a friendly message if none is active.
- `--color auto|always|never` - controls terminal color. `auto` respects TTY detection and `NO_COLOR`; use `always` to force the enhanced interactive styling.

Notes:

- Works on macOS/Linux (symlink) and Windows (copy).
- Release binaries are built for macOS, Linux, and Windows from tagged releases.
- Set `CODEX_HOME` or pass `--codex-home <path>` to use a nonstandard Codex config directory for tests, automation, or isolated environments.

## Release Setup

Tagged releases publish GitHub binaries, npm packages, and the Homebrew tap formula. Configure these GitHub Actions secrets before creating a release tag:

- `NPM_TOKEN` - npm automation token with publish access to `@shayyz-code/codex-auth` and the platform binary packages.
- `HOMEBREW_TAP_TOKEN` - GitHub token with write access to `shayyz-code/homebrew-tap`.

Release tags must use the `v*.*.*` format, for example `v0.1.4`.

## State Compatibility

`codex-auth` stores account snapshots in `~/.codex/accounts` and records the active account in `~/.codex/current`. Existing snapshots in that layout are recognized automatically.
