# Go Migration Plan

## Goal

Move `codex-su` from a Node.js oclif CLI to a small Go binary while preserving current user workflows and making production distribution simpler.

## Current App

- Runtime: Node.js 18+.
- CLI framework: oclif.
- Commands: `save`, `use`, `list`, `current`.
- State paths: `~/.codex/auth.json`, `~/.codex/accounts/*.json`, `~/.codex/current`.
- Platform behavior: symlink on macOS/Linux, copy on Windows.

## Migration Phases

1. Stabilize TypeScript behavior

- Add tests for path handling, validation, save/use/list/current behavior, and Windows copy behavior.
- Add `CODEX_HOME` or `--codex-home` override so tests never mutate a real `~/.codex`.
- Define machine-readable output with `--json`.

2. Introduce Go implementation

- Create `go.mod` with packages for commands, account storage, and filesystem operations.
- Use the standard library where possible; keep dependencies minimal.
- Add golden tests that compare expected command output to the TypeScript CLI behavior.
- Keep the TypeScript CLI as the reference until Go reaches parity.

3. Switch production binary

- Change npm package to ship the Go binary through platform-specific optional packages or a postinstall downloader.
- Keep `codex-su` as the only production command.
- Publish checksums and signatures for every binary artifact.

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
