# Testing

Part of the [stack](README.md). The test strategy is in [0025](../decisions/0025-contracts-and-tests.md). This file holds the tools that carry it out in the current language.

## Current

| Choice | Current | Why | Since | Replace when |
| -- | -- | -- | -- | -- |
| Test runner | `go test -race`, parallel by default | See [0025](../decisions/0025-contracts-and-tests.md). Tests call `t.Parallel()` and use `t.TempDir()` for files. | 2026-10-15 |  |
| Acceptance runner | A Go program that reads the scenarios and drives the binary | It must stay replaceable. Scenarios are data, and the runner is the only language-specific part. | 2026-10-15 | The implementation language changes, or a neutral runner is worth having. |
| Mutation tool | `go-mutesting`, the avito-tech fork, at pseudo-version `v0.0.0-20261002081621-9ce5f731e321`, run by `just mutate <component>` | Needed for the contract-strength check in 0025. It runs with a script, `tools/mutation/exec.sh`, that builds a mutant and then runs the component's contract tests, so it works when the tests live in another package. It is installed with `go install` outside the module, so `go.mod` is unchanged. Installing it downloads a newer Go toolchain for the tool. | 2026-10-16 | A maintained tool can run a contract suite in another package without a script. |

## Settings

- Golden regeneration is a separate explicit command and never runs as a side effect of a test. The command is not defined yet.
- Time comes in as a parameter, so no test depends on the clock.
- `just check` runs the dependency check, the architecture tests, and the full test suite, along with the format and lint checks. CI runs the same recipe.

### Accepted survivors

A mutant that survives is either a gap in the contract or a change that alters no behavior. A person decides which, after checking it. The ones that alter no behavior go in `tools/mutation/accepted/<component>.txt`, one line each: the mutant's checksum, the file, the change, and why it alters nothing. `just mutate` passes the checksums to `go-mutesting --blacklist`, so only unreviewed survivors show. A checksum depends on the mutated file, so an edit to a file brings its survivors back for a fresh look. A rebuild deletes the component's folder, and old entries then match nothing. A builder reports survivors and never edits the list. See "Trust but verify" in `AGENTS.md`.

The `contract-hardening` skill, at [`.claude/skills/contract-hardening`](../../.claude/skills/contract-hardening/SKILL.md), is the procedure for finding the ambiguity in a contract: broken variants, the mutation tool, a second independent build, and a differential run. It carries out the contract checks in [0025](../decisions/0025-contracts-and-tests.md) and is worth its cost for a component with a large input format.

Mutation tools tried:

| Tool | Result |
| -- | -- |
| `gremlins` v0.6.0 | Not chosen. It runs only the mutated package's own tests, and our contract tests live in a different package. It stops at the coverage step with "no packages to test" |
| `go-mutesting`, avito-tech fork | Chosen. It runs a command per mutant, and the command can run any tests |

## Retired

| Choice | Was | Since | Until | Replaced by | Why it changed |
| -- | -- | -- | -- | -- | -- |

## Considered and not chosen

| Choice | Option | Status | Why | Revisit when |
| -- | -- | -- | -- | -- |
