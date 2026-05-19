# Agent Operating Guide

## Working Rules

- Prefer small, reviewable changes.
- Commit completed changes yourself after checks pass, unless the user explicitly asks not to commit.
- Use the commit message format below and mention the commit hash in the final response.
- Preserve user edits. Do not revert unrelated changes in the worktree.
- Keep production behavior stable unless a task explicitly requests a breaking change.

## Standard Checks

- Run `go test ./...` before handing off Go changes.
- Run `go build -o bin/codex-auth ./cmd/codex-auth` for CLI behavior changes.
- Prefer Cobra commands and command tests for CLI behavior changes.
- For documentation-only changes, run a quick spell/readability pass and report that no code checks were needed.
- If checks cannot run, state the exact command and failure.

## Commit Message Format

Use this shape when summarizing completed work:

```text
type(scope): concise summary

- Detail the behavior or documentation changed.
- Mention tests or checks run.
- Call out follow-up work when relevant.
```

Common types: `feat`, `fix`, `docs`, `test`, `build`, `ci`, `refactor`, `chore`.

## Migration Principles

- Match the existing CLI behavior first, then improve internals.
- Keep Go tests around account storage behavior and add CLI tests when output changes.
- Keep command names, output formats, and storage paths intentional and documented.
- Treat release automation as production code: tagged releases should be repeatable, auditable, and reversible.
