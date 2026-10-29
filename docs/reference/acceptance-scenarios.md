# Acceptance scenarios

Status: accepted as the starting shape. It is expected to change as the first scenarios are written. Anything marked as open is open on purpose.

An acceptance scenario describes one run of the built binary, from outside. It is data. It names the files a run starts with, what the run does, and what must be true afterward. A scenario sees the binary only through its arguments, standard input and output, exit code, and files. It never imports the code, so it keeps working if the implementation language changes. See [0025](../decisions/0025-contracts-and-tests.md) and [0010](../decisions/0010-plan-editor.md).

Scenarios are the lasting statement of what the tool does. A person approves them, and they are few and high level. The contracts under `contracts/` judge one component. A scenario judges the whole binary.

## Where they live

- A scenario is one file, `acceptance/scenarios/<name>.yaml`.
- Files a scenario starts with, and golden outputs it compares against, live under `acceptance/testdata/`. A scenario may also point at a file elsewhere in the repository, such as the ACME example under `docs/reference/examples/acme/`.
- The runner reads the scenarios and drives the binary. See [testing](../stack/testing.md).

## A scenario

```yaml
name: assemble-acme
description: The ACME plan assembles into the same document each time.
source: docs/reference/examples/acme/README.md
steps:
  - run: [onboard-kit, assemble, --plan, acme.obk.yaml]
    files:
      acme.obk.yaml: {from: docs/reference/examples/acme/acme.obk.yaml}
    expect:
      exit: 0
      files:
        acme.md: {same-as: docs/reference/examples/acme/acme.md}
```

| Field | Meaning |
| -- | -- |
| `name` | The scenario's name. It matches the file name |
| `description` | One sentence, in plain words |
| `source` | Where the expectation comes from: a doc, an ADR, or an issue. A person can check the scenario against it |
| `steps` | One or more steps, run in order in one temporary folder |

A step is one of three kinds.

### A command step

| Field | Meaning |
| -- | -- |
| `run` | The command as an argument list. The first item is `onboard-kit`, which the runner replaces with the path of the built binary |
| `files` | Files to place in the temporary folder before the step. Each is `path: text`, or `path: {from: <file>}` to copy one |
| `stdin` | Optional text for standard input |
| `expect` | What must be true after the run. See below |

### A stdio step

`onboard-kit stdio` holds a plan and answers line-delimited JSON requests. A stdio step sends requests and checks each response.

```yaml
steps:
  - run: [onboard-kit, stdio]
    talk:
      - send: {command: questions}
        expect: {ok: true}
      - send: {command: submit, question: plan.team_state, value: established}
        expect: {ok: true, problems: []}
    expect:
      exit: 0
```

The runner writes one request per line, reads one response per line, and closes standard input when the list ends. `expect` after `talk` describes the process, which should exit 0 once standard input closes.

### A terminal step

A terminal step runs the binary in a pseudo-terminal behind a terminal emulator, as decided in [0570](../decisions/0570-test-the-terminal-screens-in-a-real-terminal.md). Use it for what a command step cannot see: key decoding, the drawn screen, the terminal modes, resizing, signals, and the exit code of a full-screen program.

```yaml
steps:
  - run: [onboard-kit, walk, --plan, acme.obk.yaml]
    files:
      acme.obk.yaml: {from: docs/reference/examples/acme/acme.obk.yaml}
    terminal:
      size: {columns: 110, rows: 32}
      env: {COLORTERM: truecolor}
      actions:
        - wait-for: "Onboarding plan"
        - expect-modes: {alternate_screen: true, cursor_visible: false}
        - expect-colors: terminal-only
        - press: q
        - wait-exit: true
        - expect-modes: {alternate_screen: false, cursor_visible: true, bracketed_paste: false}
    expect:
      exit: 0
```

| Field | Meaning |
| -- | -- |
| `terminal.size` | `columns` and `rows`. The default is 80 by 24 |
| `terminal.env` | Environment variables for the program. The runner starts from `PATH`, `HOME` set to the scenario folder, and `TERM` set to `xterm-256color`. Set `TERM`, `COLORTERM`, and `NO_COLOR` here, because the color profile follows them |
| `terminal.actions` | A list run in order. Each action has exactly one field |
| `expect` | `exit` and `files` and `unchanged` work as in a command step. A terminal step has no `stdout` or `stderr` |

An action is one of these.

| Action | Meaning |
| -- | -- |
| `wait-for: TEXT` | Wait until the text is on the screen. It fails after 5 seconds, or when the step runs out of time, and shows the screen |
| `wait-gone: TEXT` | Wait until the text is no longer on the screen. It passes at once when the text is not there, so it also checks that the text is absent |
| `press: KEY` | Press a key or a list of keys, such as `q`, `Enter`, `Tab`, `Ctrl+S`, `Escape`, or `Up`. The runner pauses 40 milliseconds after each press, because Bubble Tea drops keys that arrive in one read |
| `type: TEXT` | Type the text, with the same pause afterward |
| `resize: {columns, rows}` | Change the window size |
| `signal: NAME` | Send a real signal to the program with the operating system: `INT`, `TERM`, `HUP`, or `QUIT`. Signals sent through the emulator do not match real ones, so the runner does not use them |
| `wait-exit: true` | Wait for the program to exit |
| `expect-modes: {...}` | Check the terminal modes now: `alternate_screen`, `cursor_visible`, and `bracketed_paste`. Name only the modes to check |
| `expect-colors: terminal-only` | Check that every cell on the screen now is the default color or one of the terminal's sixteen indexed colors, as [0560](../decisions/0560-use-the-terminal-colors.md) requires |
| `expect-colors: plain` | Check that every cell is the default color with no bold, faint, italic, reverse, underline, strike, or blink, as 0560 requires with `NO_COLOR` or a terminal with no color |

The exit code of a program that a signal ended is 128 plus the signal number, as a shell reports it. When the scenario gives `expect.exit`, the runner waits for the program to exit. When it does not, the runner ends the program with `SIGKILL` after the last action. The runner reads cell colors with garbage collection paused, because the emulator binding can crash a process that reads cells while the collector moves its stack.

## What a scenario can expect

| Field | Meaning |
| -- | -- |
| `exit` | The exit code, exactly |
| `stdout`, `stderr` | `{json: <value>}` compares as JSON by subset. `{contains: <text>}` checks for the text. `{equals: <text>}` compares exactly. Left out means not checked |
| `files` | For each path, `{same-as: <file>}` compares the bytes exactly, and `{json: <value>}` compares by subset. `{absent: true}` says the file must not exist |
| `unchanged` | Paths whose bytes must be the same as before the step |

Files also support the exclusive bounded structural [YAML scalar predicate](acceptance-yaml-scalar-assertions.md), for example to require a stored nonempty UTC timestamp without guessing its value. Its strict YAML and timestamp profile applies only to that new assertion.

**Subset comparison of JSON.** An expected object matches when every key it names is present and matches. Extra keys in the actual object are fine, and key order does not matter. An expected list matches when it has the same items in the same order. An expected value of `null` matches `null`. This lets a scenario say only what it cares about, so an added field does not break it.

## How the runner treats a run

- Each scenario gets its own temporary folder, removed afterward. Resolve that newly owned folder to its physical absolute path before file setup or child execution. Every step uses that same physical working directory, even when the selected temporary parent is reached through a symlink.
- A run that does not finish within a time limit is killed, and the scenario fails with what was captured.
- A scenario fails when any expectation is not met. The failure names the step and the expectation.
- Scenarios do not depend on each other, and may run in parallel.

Physical sandbox resolution is test setup, not normalization of product paths
or command arguments. Preserve the selected parent, its unrelated files, ambient
working directory, TMPDIR and PWD. Do not strip ACLs, change groups or relax the
product's path rules. Remove only the scenario's new folder after success or an
ordinary failure.

## Choices the first runner made

These came from building the runner, and any of them can change.

- An `exit` left out is not checked, as with `stdout` and `stderr`.
- Inside a list, items compare by subset, so an object in a list may have extra keys. The list length and order must match.
- For a stdio step, `stdout` means all response lines joined.
- Numbers compare as 64-bit floats. A bare YAML date stays text.
- Each step has a time limit, 30 seconds by default.
- A scenario's `name` must match its file name. An unknown field in a scenario is an error.
- A path in `files`, `unchanged`, `from`, or `same-as` stays inside its root: no `..` and no absolute path.
- `talk` and `stdin` do not mix in one step.
- Environment variables pass through from the test process, except in a terminal step, which builds its own.

## Open

These are not decided, and the first scenarios will show what is needed.

- Whether scenarios need environment variables, or a way to fix the date or other inputs.
- Whether a scenario can expect a failure message in words, and how loosely.
- Reusing one set of files across several scenarios.
- A way to mark a scenario as known to fail, for work in progress.
- Whether a scenario file can be plain JSON as well as YAML.
