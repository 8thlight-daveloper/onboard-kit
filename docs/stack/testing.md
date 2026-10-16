# Testing

Part of the [stack](README.md). The test strategy is in [0025](../decisions/0025-contracts-and-tests.md). This file holds the tools that carry it out in the current language.

## Current

| Choice | Current | Why | Since | Replace when |
| -- | -- | -- | -- | -- |
| Test runner | `go test -race`, parallel by default | See [0025](../decisions/0025-contracts-and-tests.md). Tests call `t.Parallel()` and use `t.TempDir()` for files. | 2026-10-15 |  |
| Acceptance runner | A Go program that reads the scenarios and drives the binary | It must stay replaceable. Scenarios are data, and the runner is the only language-specific part. | 2026-10-15 | The implementation language changes, or a neutral runner is worth having. |
| Mutation tool | To decide | Needed for the contract-strength check in 0025. |  |  |

## Settings

- Golden regeneration is a separate explicit command and never runs as a side effect of a test. The command is not defined yet.
- Time comes in as a parameter, so no test depends on the clock.
- `just pre-commit` runs the dependency check, the architecture tests, and the full test suite.

## Retired

| Choice | Was | Since | Until | Replaced by | Why it changed |
| -- | -- | -- | -- | -- | -- |

## Considered and not chosen

| Choice | Option | Status | Why | Revisit when |
| -- | -- | -- | -- | -- |
