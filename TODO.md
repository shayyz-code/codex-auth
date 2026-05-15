# TODO

## Current Go CLI

- [x] Add unit tests for account name validation, snapshot creation, account switching, and current-account detection.
- [x] Replace direct writes with atomic temp-file writes where state files can be partially written.
- [x] Replace handwritten argument parsing with Cobra.
- [x] Add CLI workflow tests for command parsing and failures.
- [ ] Add `--json` output for scripts and automation.
- [x] Add `--codex-home <path>` flag support alongside the existing `CODEX_HOME` override.
- [ ] Add a migration note for existing users moving from `codex-auth` to `codex-su`.
- [ ] Replace single-binary npm packaging with per-platform packages before publishing npm broadly.

## Go Migration

- [x] Create a Go module after the TypeScript CLI behavior is covered by tests.
- [x] Port command parsing and account operations behind small packages.
- [x] Preserve command compatibility for `save`, `use`, `list`, and `current`.
- [x] Produce cross-platform release binaries with checksums.
- [ ] Sign release binaries.
- [ ] Add Homebrew formula release automation after Go binaries exist.

## Distribution

- [ ] Publish npm package from GitHub Actions after platform binary packaging is added.
- [ ] Publish GitHub Release artifacts on version tags.
- [ ] Add Homebrew tap once the Go binary release is stable.
- [ ] Evaluate Scoop, Winget, Arch AUR, and Docker distribution after the binary interface stabilizes.
