# Contributing

Thanks for helping improve `codex-auth`. Keep changes small, tested, and easy to review.

## Development Setup

Requirements:

- Go 1.24 or newer.
- Node.js 22 or newer for npm packaging tests.
- `make`.

Build and test locally:

```sh
make build
make test
make check
```

Use `--codex-home <path>` or `CODEX_HOME` when testing account behavior so local Codex credentials are not affected.

## Workflow

1. Open an issue or describe the change before large behavior changes.
2. Keep pull requests focused on one feature, fix, or documentation update.
3. Preserve existing CLI command names, output formats, and storage paths unless the change intentionally updates them.
4. Add or update tests when account storage, command behavior, release automation, or npm packaging changes.
5. Update `README.md` and `CHANGELOG.md` for user-facing changes.

## Checks

- Run `make test` for Go or npm launcher changes.
- Run `make build` for CLI behavior changes.
- Run `make check` when changes touch shared behavior, release automation, or both Go and npm packaging.
- Documentation-only changes should at least pass `git diff --check`.

## Version Updates

Use the Makefile target so all npm package versions stay aligned:

```sh
make version VERSION=0.2.1
```

This updates the root npm package, platform package versions, and root optional dependency pins. Release tags still use the `v*.*.*` format, such as `v0.2.1`.

## Release Preparation

Before creating a release tag:

1. Run `make version VERSION=<version>`.
2. Move user-facing changes from `Unreleased` into a dated `CHANGELOG.md` section.
3. Run `make check`.
4. Commit the release preparation changes.
5. Tag and push with `git tag v<version>` and `git push origin main v<version>`.

Tagged releases publish GitHub binaries, npm packages, and the Homebrew tap formula through GitHub Actions.
