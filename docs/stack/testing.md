# Testing

Part of the [stack](README.md). The test strategy is in [0025](../decisions/0025-contracts-and-tests.md). This file holds the tools that carry it out in the current language.

## Current

| Choice | Current | Why | Since | Replace when |
| -- | -- | -- | -- | -- |
| Test runner | `go test -race`, parallel by default | See [0025](../decisions/0025-contracts-and-tests.md). Tests call `t.Parallel()` and use `t.TempDir()` for files. | 2026-10-15 |  |
| Acceptance runner | A Go program that reads the scenarios and drives the binary | It must stay replaceable. Scenarios are data, and the runner is the only language-specific part. | 2026-10-15 | The implementation language changes, or a neutral runner is worth having. |
| Mutation tool | `gomu` v0.3.0, run by `just mutate <component>` through the small program in `tools/mutation/gomurun` | Needed for the contract-strength check in 0025. gomu builds each mutant over an overlay, so it never changes a source file and can run beside `just check`. It forwards test flags, so the program points it at the component's contract folder, which is another package. It checks that each mutant compiles and leaves the ones that do not out of the score. It has the widest set of mutators, including removal of assignments, which the earlier tool skipped. It is installed with `go install` into `.direnv/bin`, so `go.mod` is unchanged. The program runs it one mutant at a time with Go limited to two parallel jobs, because an unthrottled run overloaded a laptop. It also gives `go test` its own `-timeout`: a mutant that makes a loop run forever hangs the test, gomu stops only the `go` command, and the compiled test binary would otherwise keep spinning for ten minutes. A hundred such mutants overloaded the machine and caused two kernel panics. | 2026-10-21 | gomu stops being maintained, or it can name a survivor stably itself. |

## Settings

- Golden regeneration is a separate explicit command and never runs as a side effect of a test. The command is not defined yet.
- Time comes in as a parameter, so no test depends on the clock.
- `just check` runs the dependency check, the architecture tests, and the full test suite, along with the format and lint checks. CI runs the same recipe.

### Accepted survivors

A mutant that survives is either a gap in the contract or a change that alters no behavior. A person decides which, after checking it. The ones that alter no behavior go in `tools/mutation/accepted/<component>.txt`, one line each: the survivor's key, then the file, the change, and `::` and why it alters nothing. `just mutate` prints every survivor as exactly the first part of that line, so a person adds a line by copying it and writing the reason. The key is a hash of the file name, the mutator, the original and mutated text, and the text of the line, so it survives a line moving and changes when the line is edited. A survivor whose line was edited comes back for a fresh look. A rebuild deletes the component's folder, and old entries then match nothing. A builder reports survivors and never edits the list. See "Trust but verify" in `AGENTS.md`.

The accepted lists that existed before 2026-10-21 hold the earlier tool's checksums, which no longer match anything. Each one needs a person to review it again against a gomu run, and to replace its lines with the new keys.

Identical mutations are numbered in source order across the full report, including killed mutants. Killing an earlier identical mutant must not change a later survivor's key or give it an exception meant for the earlier one. Keys from a run before this rule need a fresh match when identical mutations occur in the same file.

The `contract-hardening` skill, at [`.claude/skills/contract-hardening`](../../.claude/skills/contract-hardening/SKILL.md), is the procedure for finding the ambiguity in a contract: broken variants, the mutation tool, a second independent build, and a differential run. It carries out the contract checks in [0025](../decisions/0025-contracts-and-tests.md) and is worth its cost for a component with a large input format.

Mutation tools tried:

| Tool | Result |
| -- | -- |
| `gremlins` v0.6.0 | Not chosen. It runs only the mutated package's own tests, and our contract tests live in a different package. It stops at the coverage step with "no packages to test" |
| `go-mutesting`, avito-tech fork | Used until 2026-10-21, still available as `just mutate-old`. It overwrites source files in place, so it cannot run beside `just check`, and it silently skips every mutant that assigns to a struct, slice, or map field, which hid the deep-copy gaps. Upstream pull requests from outside Avito mostly sit unmerged |
| `ooze` | Not chosen. It has no statement or assignment removal, it counts a mutant that does not compile as killed, and it walks the whole repository for each run |
| `go-mutesting` with two local patches | Not chosen. It fixed the skips, but it would need a patched build to maintain, and it still overwrites files |
| `gomu` | Chosen. See the table above |

## Retired

| Choice | Was | Since | Until | Replaced by | Why it changed |
| -- | -- | -- | -- | -- | -- |

## Considered and not chosen

| Choice | Option | Status | Why | Revisit when |
| -- | -- | -- | -- | -- |
