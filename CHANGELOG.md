# Changelog

All notable changes to this project should be documented in this file.

This project follows the spirit of Keep a Changelog and uses semantic versioning once releases are automated.

## Unreleased

## 0.2.0 - 2026-05-19

### Added

- Added closest-name suggestions when `use <name>` does not match a saved account.
- Added prompts to save an unsaved live Codex login before switching accounts.
- Added startup syncing that matches the live Codex auth file to saved snapshots and refreshes the current account.
- Added account renaming with saved-name and detected-email context.
- Added an arrow-key account picker for interactive terminal use.
- Added bordered account table output for `list`.
- Added `make version VERSION=<semver>` to update root and platform npm package versions together.
- Added contributor documentation.
- Added README screenshots and notices for release readiness.
- Added a README features section.

### Changed

- `save <name>` now always writes to the requested account name, even when the live auth matches another saved account.
- `use` no longer accepts numeric prompt input; terminal selection uses the interactive picker and piped input accepts saved names or emails.
- Cancelling the interactive picker exits cleanly instead of falling back to a secondary prompt.
- Deferred additional distribution channel evaluation until after the `v0.2.0` release.
- Clarified that README requirements apply only when building from source or running repository checks.

## 0.1.4 - 2026-05-19

### Changed

- Standardized the project, command, release artifact, Homebrew formula, and npm package naming on `codex-auth`.
- Bumped the root and platform npm packages to `0.1.4` for the next publish attempt.

## 0.1.3 - 2026-05-19

### Added

- Added repository tests that assert CI, release, docs, and packaging contracts.
- Added per-platform npm package metadata, a root npm launcher, binary staging, and tag-gated npm publishing.
- Added Sigstore keyless signing for tagged release binaries.
- Added Homebrew formula generation and tag-gated updates for `shayyz-code/homebrew-tap`.
- Added README badges, installation instructions, and release secret setup notes.

### Changed

- Renamed npm publishing from the unavailable unscoped package to scoped `@shayyz-code/codex-auth` packages.
- Updated release automation to validate staged npm packages before publishing.
- Updated migration and distribution milestones to reflect completed Go, npm, signing, and Homebrew work.

### Changed

- Standardized README usage examples around the production command name.
- Migrated the CLI implementation from TypeScript/oClif to a Go binary.
- Replaced Node CI with Go test and build checks.
- Replaced handwritten Go argument parsing with Cobra.

### Added

- Added CI and release workflow scaffolding.
- Added migration, release, todo, and agent operation documentation.
- Added `CODEX_HOME` support for tests, automation, and isolated environments.
- Added account-service tests for snapshot, switch, current-account, and validation behavior.
- Added cross-platform release binary build scaffolding with SHA-256 checksums.
- Added CLI workflow tests for command parsing, failures, version output, and `--codex-home`.
- Added npm `bin` metadata and a prepack build so package dry runs expose the `codex-auth` executable.
