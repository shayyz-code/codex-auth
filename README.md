<h1 align="center">codex-su</h1>

<p align="center">A command-line tool that lets you manage and switch between multiple Codex accounts.</p>

<p align="center">
  <a href="https://github.com/shayyz-code/codex-su/actions/workflows/ci.yml"><img alt="CI" src="https://github.com/shayyz-code/codex-su/actions/workflows/ci.yml/badge.svg"></a>
  <a href="https://github.com/shayyz-code/codex-su/releases"><img alt="GitHub release" src="https://img.shields.io/github/v/release/shayyz-code/codex-su?sort=semver"></a>
  <a href="https://www.npmjs.com/package/codex-su"><img alt="npm" src="https://img.shields.io/npm/v/codex-su"></a>
  <a href="https://github.com/shayyz-code/codex-su/blob/main/LICENSE"><img alt="License" src="https://img.shields.io/github/license/shayyz-code/codex-su"></a>
</p>

> [!WARNING]
> Not affiliated with OpenAI or Codex. Not an official tool.

## How it Works

Codex stores your authentication session in a single `auth.json` file. This tool works by creating named snapshots of that file for each of your accounts. When you want to switch, `codex-su` swaps the active `~/.codex/auth.json` with the snapshot you select, instantly changing your logged-in account.

## Requirements

- Go 1.24 or newer to build from source.
- Homebrew or npm if you install through those package managers.

## Installation

### Homebrew

```sh
brew tap shayyz-code/tap
brew install codex-su
```

### npm

```sh
npm install -g codex-su
```

### GitHub Releases

Download the binary for your platform from the [latest release](https://github.com/shayyz-code/codex-su/releases/latest), then put it somewhere on your `PATH`.

### Go install

```sh
go install github.com/shayyz-code/codex-su/cmd/codex-su@latest
```

### Build locally

```sh
go build -o bin/codex-su ./cmd/codex-su
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
- Release binaries are built for macOS, Linux, and Windows from tagged releases.
- Set `CODEX_HOME` or pass `--codex-home <path>` to use a nonstandard Codex config directory for tests, automation, or isolated environments.

## Release Setup

Tagged releases publish GitHub binaries, npm packages, and the Homebrew tap formula. Configure these GitHub Actions secrets before creating a release tag:

- `NPM_TOKEN` - npm automation token with publish access to `codex-su` and the platform binary packages.
- `HOMEBREW_TAP_TOKEN` - GitHub token with write access to `shayyz-code/homebrew-tap`.

Release tags must use the `v*.*.*` format, for example `v0.1.3`.

## Migrating from `codex-auth`

If you are moving from the legacy `codex-auth` tool:

1.  **Compatible State**: `codex-su` uses the same directory structure (`~/.codex/accounts`) and file format as `codex-auth`. Your existing snapshots will be recognized automatically.
2.  **Binary Name**: The command has been renamed to `codex-su` to follow the "switch user" convention.
3.  **Removal**: You can safely uninstall the old tool after installing `codex-su`.
