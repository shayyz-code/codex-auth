# Go Migration Plan

## Goal

Move `codex-su` from a Node.js oclif CLI to a small Go binary while preserving current user workflows and making production distribution simpler.

## Previous App

- Runtime: Node.js 18+.
- CLI framework: oclif.
- Commands: `save`, `use`, `list`, `current`.
- State paths: `~/.codex/auth.json`, `~/.codex/accounts/*.json`, `~/.codex/current`.
- Platform behavior: symlink on macOS/Linux, copy on Windows.

## Current App

- Runtime: native Go binary.
- Dependencies: Go standard library only.
- Commands: `save`, `use`, `list`, `current`.
- State paths: `~/.codex/auth.json`, `~/.codex/accounts/*.json`, `~/.codex/current`.
- Platform behavior: symlink on macOS/Linux, copy on Windows.

## Migration Phases

1. Stabilize TypeScript behavior

- Add tests for path handling, validation, save/use/list/current behavior, and Windows copy behavior.
- Add `CODEX_HOME` or `--codex-home` override so tests never mutate a real `~/.codex`.
- Define machine-readable output with `--json`.

2. Introduce Go implementation

- Done: created `go.mod` with packages for commands, account storage, and filesystem operations.
- Done: the current implementation uses the standard library only.
- Done: account-service behavior is covered by Go tests.
- Remaining: add CLI output tests and `--json` output before the next release.

3. Switch production binary

- Done: `codex-su` is the only production command.
- Done: tagged releases build cross-platform binaries with SHA-256 checksums.
- Remaining: publish signatures for every binary artifact.
- Remaining: change npm package to ship the Go binary through platform-specific optional packages or a postinstall downloader.

4. Expand distribution

- Use GoReleaser for GitHub Releases and Homebrew tap updates.
- Add npm publishing for users who install CLIs through Node tooling.
- Evaluate Scoop, Winget, Arch AUR, Docker, and direct shell installer after the binary interface is stable.

## Production Standardization

- Command name: `codex-su`.
- Versioning: semantic versioning with a `CHANGELOG.md` entry for every release.
- Releases: created from tags only.
- CI: typecheck/build/test on pull requests and main.
- CD: publish artifacts only after CI passes.
- State override: support `CODEX_HOME` for tests, automation, and isolated production environments.

## Release Design

- GitHub Release: source archive, binaries, checksums, signatures.
- Homebrew: generated formula in a tap after Go binaries exist.
- npm: package command remains `codex-su`; implementation can point to the Go binary once migrated.
- Other package managers: add only after checksums, install tests, and rollback documentation exist.
