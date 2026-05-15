# TODO

## Current TypeScript CLI

- [x] Add unit tests for account name validation, snapshot creation, account switching, and current-account detection.
- [ ] Replace direct writes with atomic temp-file writes where state files can be partially written.
- [ ] Add `--json` output for scripts and automation.
- [ ] Add `--codex-home <path>` flag support alongside the existing `CODEX_HOME` override.
- [ ] Add a migration note for existing users moving from `codex-auth` to `codex-su`.

## Go Migration

- [ ] Create a Go module after the TypeScript CLI behavior is covered by tests.
- [ ] Port command parsing and account operations behind small packages.
- [ ] Preserve command compatibility for `save`, `use`, `list`, and `current`.
- [ ] Produce signed cross-platform release binaries with checksums.
- [ ] Add Homebrew formula release automation after Go binaries exist.

## Distribution

- [ ] Publish npm package from GitHub Actions using npm provenance.
- [ ] Publish GitHub Release artifacts on version tags.
- [ ] Add Homebrew tap once the Go binary release is stable.
- [ ] Evaluate Scoop, Winget, Arch AUR, and Docker distribution after the binary interface stabilizes.
