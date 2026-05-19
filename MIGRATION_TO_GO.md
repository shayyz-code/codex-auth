# Go Migration Plan

## Goal

Move `codex-auth` from a Node.js oclif CLI to a small Go binary while preserving current user workflows and making production distribution simpler.

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
- Done: CLI output, command failures, `--json` output, and `--codex-home` behavior are covered by Go tests.

3. Switch production binary

- Done: `codex-auth` is the only production command.
- Done: tagged releases build cross-platform binaries with SHA-256 checksums.
- Done: tagged releases sign every binary artifact with Sigstore keyless signing.
- Done: npm packaging uses a root launcher with platform-specific optional binary packages.
- Done: tagged release automation stages, validates, and publishes platform npm packages before publishing the root npm package.
- Done: tagged release automation updates `shayyz-code/homebrew-tap` with a generated formula using `HOMEBREW_TAP_TOKEN`.

4. Expand distribution

- Use GitHub Actions for GitHub Releases, npm publishing, and Homebrew tap updates.
- Add npm publishing for users who install CLIs through Node tooling.
- Evaluate Scoop, Winget, Arch AUR, Docker, and direct shell installer after the binary interface is stable.

## Production Standardization

- Command name: `codex-auth`.
- Versioning: semantic versioning with a `CHANGELOG.md` entry for every release.
- Releases: created from tags only.
- CI: build/test on pull requests and main, with repository tests covering the expected workflow checks.
- CD: publish artifacts only after CI passes.
- State override: support `CODEX_HOME` for tests, automation, and isolated production environments.

## Release Design

- GitHub Release: source archive, binaries, checksums, and Sigstore signature bundles.
- Homebrew: generated formula pushed to `shayyz-code/homebrew-tap` from tagged releases.
- npm: package command remains `codex-auth`; implementation can point to the Go binary once migrated.
- Other package managers: add only after checksums, install tests, and rollback documentation exist.
