# TODO

## Current Go CLI

- [x] Add unit tests for account name validation, snapshot creation, account switching, and current-account detection.
- [x] Replace direct writes with atomic temp-file writes where state files can be partially written.
- [x] Replace handwritten argument parsing with Cobra.
- [x] Add CLI workflow tests for command parsing and failures.
- [x] Add `--json` output for scripts and automation.
- [x] Add `--codex-home <path>` flag support alongside the existing `CODEX_HOME` override.
- [x] Add a state compatibility note for existing account snapshots.
- [x] Add repository tests that assert CI and release workflows keep running the required checks.
- [x] Add a tested npm launcher that resolves per-platform binary packages.
- [x] Add tested npm package manifests for each supported platform binary.
- [x] Add release workflow validation that stages binaries into npm packages and dry-runs package creation.
- [x] Replace single-binary npm packaging with per-platform packages before publishing npm broadly.

## Go Migration

- [x] Create a Go module after the TypeScript CLI behavior is covered by tests.
- [x] Port command parsing and account operations behind small packages.
- [x] Preserve command compatibility for `save`, `use`, `list`, and `current`.
- [x] Produce cross-platform release binaries with checksums.
- [x] Sign release binaries.
- [x] Add Homebrew formula release automation after Go binaries exist.

## Distribution

- [x] Publish npm package from GitHub Actions after platform binary packaging is added.
- [x] Publish GitHub Release artifacts on version tags.
- [x] Add Homebrew tap once the Go binary release is stable.
- [x] Document required `NPM_TOKEN` and `HOMEBREW_TAP_TOKEN` GitHub Actions secrets.
- [ ] Configure the documented release secrets in GitHub Actions before tagging.
- [ ] Evaluate Scoop, Winget, Arch AUR, and Docker distribution after the binary interface stabilizes.
