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

A step is one of two kinds.

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

## What a scenario can expect

| Field | Meaning |
| -- | -- |
| `exit` | The exit code, exactly |
| `stdout`, `stderr` | `{json: <value>}` compares as JSON by subset. `{contains: <text>}` checks for the text. `{equals: <text>}` compares exactly. Left out means not checked |
| `files` | For each path, `{same-as: <file>}` compares the bytes exactly, and `{json: <value>}` compares by subset. `{absent: true}` says the file must not exist |
| `unchanged` | Paths whose bytes must be the same as before the step |

**Subset comparison of JSON.** An expected object matches when every key it names is present and matches. Extra keys in the actual object are fine, and key order does not matter. An expected list matches when it has the same items in the same order. An expected value of `null` matches `null`. This lets a scenario say only what it cares about, so an added field does not break it.

## How the runner treats a run

- Each scenario gets its own temporary folder, removed afterward.
- A run that does not finish within a time limit is killed, and the scenario fails with what was captured.
- A scenario fails when any expectation is not met. The failure names the step and the expectation.
- Scenarios do not depend on each other, and may run in parallel.

## Open

These are not decided, and the first scenarios will show what is needed.

- Whether scenarios need environment variables, or a way to fix the date or other inputs.
- Whether a scenario can expect a failure message in words, and how loosely.
- Reusing one set of files across several scenarios.
- A way to mark a scenario as known to fail, for work in progress.
- Whether a scenario file can be plain JSON as well as YAML.
