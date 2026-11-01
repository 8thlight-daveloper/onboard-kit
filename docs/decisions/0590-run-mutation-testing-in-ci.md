# 0590. Run mutation testing in CI, not on a laptop

Status: draft, not yet accepted by a person.

Builds on [0025](0025-contracts-and-tests.md). Changes the run points in the
mutation cadence of the [testing stack](../stack/testing.md). Keeps the
acceptance rules of [0130](0130-survivor-risk-acceptance.md) and `AGENTS.md`.

## Context

A full mutation run of one component takes tens of minutes to hours. The rebuilt
catalog edit guide has 2710 mutants and ran for most of an afternoon on a
laptop, with the fans on. An earlier unthrottled run panicked the machine twice.
Throttling fixed the crashes and left the cost: the laptop is slow for hours and
a person waits on a result that arrives long after the change.

The result is still worth having. Survivors have found gaps that broken variants
missed. What fails is where and when the run happens, not whether to run it.

Going public removes the cost of a hosted runner. Hosted runners are free for a
public repository.

## Decision

Run mutation testing nightly on GitHub, for the components that changed. Stop
starting runs on a laptop by default.

**The job.** A workflow, `.github/workflows/mutate.yml`, runs a mutation test per
component. It runs nightly once the repository is public. Until then a person
starts it by hand, for chosen components or for all of them. A manual run never
moves the starting point of the nightly run.

**Only what changed.** A first job finds the head commit of the last successful
run and lists the components whose `internal/<component>/` or
`contracts/<component>/` changed since then. Only those run. A change to the
domain model, `go.mod`, `go.sum`, the mutation tool, the Nix files, the
`justfile` or the workflow runs every component. With no earlier successful run,
or no common history with it, every component runs. With nothing changed, the
run ends in seconds. A run that fails does not move the starting point, so the
next night covers the same changes again.

**What a run reports.** Each component job runs the full suite, with no `--short`
flag, so no survivor needs a second look against the suite. It writes the score
and the survivors to the job summary and keeps the whole log as an artifact for
90 days. A run never fails on a score and never edits a file in the repository.

**Processing results.** Run history is the worklist. A person reads the summary
when a component changes, and sorts each survivor into one of three outcomes:

- A gap in the contract, which becomes a contract change. A person authorizes
  it, and an agent can draft the test.
- A change that alters no behavior, which a person adds to
  `tools/mutation/accepted/<component>.txt` after checking the reason.
- Dead code, which becomes a component change.

A survivor nobody fixes shows up again the next time its component runs. That
repetition is the accepted cost of keeping no state outside the history.

**Agents.** An agent does not start `just mutate` unless a person asks. It reads
the nightly result when a task needs it. The rules that stay: an agent never adds
a line to the accepted lists, and a claim that a survivor is harmless needs the
scratch-copy check in "Trust but verify". The behavior judge may sort survivors
into the three outcomes only as a trial under [0100](0100-behavior-judgment.md)
and [0130](0130-survivor-risk-acceptance.md). It cannot accept one.

**Local runs.** `just mutate` stays in the `justfile` because the workflow calls
it. A person may still run it on one component, with `--short --cached`, when the
nightly result is too slow for the task.

## Consequences

- The laptop is free during a contract or build task. A result arrives the next
  morning, not within the session.
- The build rule in `AGENTS.md` no longer ends with a mutation run. A builder's
  report covers the contract tests and `just check`. Survivors come from the
  nightly run.
- A component that depends on another component's internals is not rerun when
  only that dependency changes. A manual run, or a change to a shared path above,
  covers it.
- The nightly trigger waits for the repository to go public, because a first run
  with no earlier success mutates every component and a private repository pays
  for those minutes.
- The workflow has not run yet. It needs the repository on GitHub, and a hosted
  runner may show timing or load that the laptop did not. The first run of every
  component is the check.
- Nothing stops a merge on a survivor.

## Not decided

1. **A gate.** Whether a new survivor should block a merge, or a falling score
   should fail the run. Start with reports and decide after real volume.
2. **A committed baseline.** A list of known survivors would show only new ones
   in the summary. It is state a person must keep up, and it overlaps the
   accepted lists. Add it if the repeated survivors become noise.
3. **Splitting a long component.** The job limit is six hours. If one component
   exceeds it, split the run by source file.
4. **Using the behavior judge on the results.** It needs a plain report first.
