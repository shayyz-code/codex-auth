# Changelog

All notable changes to this project should be documented in this file.

This project follows the spirit of Keep a Changelog and uses semantic versioning once releases are automated.

## Unreleased

### Changed

- Renamed the CLI package and command from `codex-auth` to `codex-su`.
- Standardized README usage examples around the production command name.
- Migrated the CLI implementation from TypeScript/oClif to a Go binary.
- Replaced Node CI with Go test and build checks.

### Added

- Added CI and release workflow scaffolding.
- Added migration, release, todo, and agent operation documentation.
- Added `CODEX_HOME` support for tests, automation, and isolated environments.
- Added account-service tests for snapshot, switch, current-account, and validation behavior.
- Added cross-platform release binary build scaffolding with SHA-256 checksums.
