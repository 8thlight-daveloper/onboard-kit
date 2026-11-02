# Testing

Part of the [stack](README.md). The test strategy is in [0025](../decisions/0025-contracts-and-tests.md). This file holds the tools that carry it out in the current language.

## Current

| Choice | Current | Why | Since | Replace when |
| -- | -- | -- | -- | -- |
| Test runner | `go test -race`, parallel by default | See [0025](../decisions/0025-contracts-and-tests.md). Tests call `t.Parallel()` and use `t.TempDir()` for files. | 2026-10-15 |  |
| Acceptance runner | A Go program that reads the scenarios and drives the binary | It must stay replaceable. Scenarios are data, and the runner is the only language-specific part. | 2026-10-15 | The implementation language changes, or a neutral runner is worth having. |
| Terminal test tool | The tui-test Go binding v0.1.1, used by the acceptance runner's terminal step | Runs the real binary in a pseudo-terminal behind a terminal emulator, to test key decoding, the screen, terminal modes, resizing, and exit. See [0570](../decisions/0570-test-the-terminal-screens-in-a-real-terminal.md). It is a test-only dependency of about 54 MB, and the shipped binary is unchanged. It is a 0.x tool with one emulator. Reading cells repeatedly can crash the test process, so the runner pauses garbage collection around each read. | 2026-10-29 | The tool is dropped, or a release below 1.0 breaks the scenarios and no fix follows. |
| Mutation tool | `gomu` v0.3.0, run by `just mutate <component>` through the small program in `tools/mutation/gomurun` | Needed for the contract-strength check in 0025. gomu builds each mutant over an overlay, so it never changes a source file and can run beside `just check`. It forwards test flags, so the program points it at the component's contract folder, which is another package. It checks that each mutant compiles and leaves the ones that do not out of the score. It has the widest set of mutators, including removal of assignments, which the earlier tool skipped. It is installed with `go install` into `.direnv/bin`, so `go.mod` is unchanged. The program runs it one mutant at a time with Go limited to two parallel jobs, because an unthrottled run overloaded a laptop. It also gives `go test` its own `-timeout`: a mutant that makes a loop run forever hangs the test, gomu stops only the `go` command, and the compiled test binary would otherwise keep spinning for ten minutes. A hundred such mutants overloaded the machine and caused two kernel panics. `just mutate <component> --short` passes `-short` to the tests, so a component's slow tests, such as ones that mount disk images, are skipped. A mutant that only a skipped test would kill then survives a short run, so check the survivors of a short run against the full suite before accepting or reporting them. `--cached` reuses the saved results of each source file whose text, and the text of the component's whole contract folder, are unchanged, so a small fix mutates only the files it touched. A change to one source file can in principle change how another file's mutants are killed, which the cache ignores. Use `--cached` while working and a run without it to close a component. The saved results live in `.direnv/mutation-cache`. | 2026-10-21 | gomu stops being maintained, or it can name a survivor stably itself. |

## Settings

- Golden regeneration is a separate explicit command and never runs as a side effect of a test. The command is not defined yet.
- Time comes in as a parameter, so no test depends on the clock.
- `just check` runs the dependency check, the architecture tests, and the full test suite, along with the format and lint checks. CI runs the same recipe.

### Archive file assertions

Whole-binary scenarios may inspect ordered ZIP payloads and metadata with the
[ZIP file assertions](../reference/acceptance-zip-assertions.md). The checker
reads bounded bytes and never extracts entries or imports product codecs.

### Captured values between steps

[Bounded captures](../reference/acceptance-captures.md) let a later command use
actual earlier output artifacts without product-specific test flags or imports.
Only prior named values and the physical sandbox path are interpolated, once.
Native paths remain local and no-follow; extension-enabled diagnostics omit
captured personal values. See [0480](../decisions/0480-bounded-acceptance-captures.md).

### Accepted survivors

[0130](../decisions/0130-survivor-risk-acceptance.md) authorizes agent-only risk
acceptance for the remaining review batch. Its entries identify the approving
agent role and evidence limits; they do not claim equivalence or individual
human approval. Historical identities remain inactive until matched. This is
a scoped exception to the normal rules below, not authority for future batches.

A mutant that survives is either a gap in the contract or a change that alters no behavior. A person decides which, after checking it. The ones that alter no behavior go in `tools/mutation/accepted/<component>.txt`, one line each: the survivor's key, the file, the change, and why it alters nothing, separated by `|`. `just mutate` prints every survivor as exactly the first part of that line, so a person adds a line by copying it and writing the reason. The key is a hash of the file name, the mutator, the original and mutated text, and the text of the line, so it survives a line moving and changes when the line is edited. A survivor whose line was edited comes back for a fresh look. A rebuild deletes the component's folder, and old entries then match nothing. A builder reports survivors and never edits the list. See "Trust but verify" in `AGENTS.md`.

The accepted lists that existed before 2026-10-21 hold the earlier tool's checksums, which no longer match anything. Each one needs a person to review it again against a gomu run, and to replace its lines with the new keys.

Identical mutations are numbered in source order across the full report, including killed mutants. Killing an earlier identical mutant must not change a later survivor's key or give it an exception meant for the earlier one. Keys from a run before this rule need a fresh match when identical mutations occur in the same file.

### Mutation cadence

A full run of a large component takes tens of minutes, and with slow tests it takes hours. Per [0590](../decisions/0590-run-mutation-testing-in-ci.md), the run happens nightly on GitHub for the components that changed. Run it on a laptop only when a person asks:

| When | What | Cost |
| -- | -- | -- |
| Editing a source file | The component's contract tests only | Seconds |
| Before a commit | `just check` | Minutes |
| Nightly, in CI | `just mutate <component> --workers 4` for each changed component, from `.github/workflows/mutate.yml` | Hours, on a hosted runner |
| Working on survivors, by choice | `just mutate <component> --short --cached` | Minutes after the first run, because unchanged files are reused |

A local `--short` run lists new survivors only, and each one needs a recheck against the full suite. A nightly run uses the full suite. A survivor on unchanged text keeps its earlier verdict through its key, so only new keys need review. Throttling stays on: a run uses `nice`, two Go jobs, and one mutant at a time.

The `contract-hardening` skill, at [`.agents/skills/contract-hardening`](../../.agents/skills/contract-hardening/SKILL.md), is the procedure for finding the ambiguity in a contract: broken variants, the mutation tool, a second independent build, and a differential run. It carries out the contract checks in [0025](../decisions/0025-contracts-and-tests.md) and is worth its cost for a component with a large input format.

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
