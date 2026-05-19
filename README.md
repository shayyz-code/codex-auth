<h1 align="center">CODEX-AUTH</h1>

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
make build
```

## Usage

```sh
# save the current logged-in token as a named account
codex-auth save <name>

# switch active account
codex-auth use <name>

# or pick interactively with arrow keys
codex-auth use

# rename a saved account
codex-auth rename <old-name> <new-name>

# or pick the account to rename interactively
codex-auth rename

# list accounts
codex-auth list

# show current account name
codex-auth current
```

### Command reference

- `codex-auth save <name>` - validates `<name>`, ensures `auth.json` exists, then snapshots it to `~/.codex/accounts/<name>.json`. The requested name is always honored, so `save new-account` writes that account even if the active auth matches another saved snapshot.
- `codex-auth use [name]` - accepts a name or launches an interactive selector with saved names and detected account emails. In a terminal, use up/down arrows and Enter to choose; piped input can still provide a saved name or email. On startup, the live Codex `auth.json` is matched against saved snapshots and `current` is refreshed before commands run. If `<name>` is mistyped, the closest saved account is suggested. Before switching away from an unsaved live Codex login, interactive mode asks whether to save it. Copies the saved snapshot into place and records the active name. Interactive terminal output uses color when supported; piped output and `--json` remain stable for automation.
- `codex-auth rename [old-name] [new-name]` - renames a saved snapshot and shows both the saved name and detected account email when available. Without arguments, it opens the same email-aware picker used by `use`.
- `codex-auth list` - renders a table of saved snapshots alphabetically, shows detected emails when available, and marks the active one with `*`.
- `codex-auth current` - prints the active account name and detected email when available, or a friendly message if none is active.
- `--color auto|always|never` - controls terminal color. `auto` respects TTY detection and `NO_COLOR`; use `always` to force the enhanced interactive styling.

Notes:

- Uses regular file copies on all platforms so external Codex logins cannot overwrite saved account snapshots through `auth.json`. Older symlink-based activations are detached automatically if Codex appears to have written through the symlink.
- Release binaries are built for macOS, Linux, and Windows from tagged releases.
- Set `CODEX_HOME` or pass `--codex-home <path>` to use a nonstandard Codex config directory for tests, automation, or isolated environments.

## Release Setup

Tagged releases publish GitHub binaries, npm packages, and the Homebrew tap formula. Configure these GitHub Actions secrets before creating a release tag:

- `NPM_TOKEN` - npm automation token with publish access to `@shayyz-code/codex-auth` and the platform binary packages.
- `HOMEBREW_TAP_TOKEN` - GitHub token with write access to `shayyz-code/homebrew-tap`.

Release tags must use the `v*.*.*` format, for example `v0.2.0`.

### Release checklist

1. Update package metadata with `make version VERSION=0.2.0`.
2. Update `CHANGELOG.md` with the release date and user-facing changes.
3. Run `make check`.
4. Commit the release preparation changes.
5. Create and push the tag:

```sh
git tag v0.2.0
git push origin main v0.2.0
```

The release workflow builds binaries, attaches checksums and Sigstore bundles to the GitHub Release, publishes npm packages, and updates the Homebrew tap.

## State Compatibility

`codex-auth` stores account snapshots in `~/.codex/accounts` and records the active account in `~/.codex/current`. Existing snapshots in that layout are recognized automatically.
