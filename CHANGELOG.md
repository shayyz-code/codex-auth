# Changelog

All notable changes to this project should be documented in this file.

This project follows the spirit of Keep a Changelog and uses semantic versioning once releases are automated.

## Unreleased

## 0.1.3 - 2026-05-19

### Added

- Added repository tests that assert CI, release, docs, and packaging contracts.
- Added per-platform npm package metadata, a root npm launcher, binary staging, and tag-gated npm publishing.
- Added Sigstore keyless signing for tagged release binaries.
- Added Homebrew formula generation and tag-gated updates for `shayyz-code/homebrew-tap`.
- Added README badges, installation instructions, and release secret setup notes.

### Changed

- Updated release automation to validate staged npm packages before publishing.
- Updated migration and distribution milestones to reflect completed Go, npm, signing, and Homebrew work.

### Changed

- Renamed the CLI package and command from `codex-auth` to `codex-su`.
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
- Added npm `bin` metadata and a prepack build so package dry runs expose the `codex-su` executable.
