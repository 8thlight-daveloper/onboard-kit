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

**The job.** A workflow, `.github/workflows/mutate.yml`, runs one job per
component, so results come back as each component finishes and a slow component
delays no other. It runs nightly once the repository is public. Until then a
person starts it by hand, for chosen components or for all of them. A manual run
never moves the starting point of the nightly run.

**Only what changed.** A first job finds the head commit of the last successful
nightly run and lists the components whose `internal/<component>/` or
`contracts/<component>/` changed since then. Only those run. A change to the
domain model, `go.mod`, `go.sum`, the mutation tool, the Nix files, the
`justfile` or the workflow runs every component. With no earlier successful run,
or no common history with it, every component runs. With nothing changed, the
run ends in seconds. A run that fails does not move the starting point, so the
next night covers the same changes again.

**What a run reports.** Each component job runs the full suite, with no `--short`
flag, so no survivor needs a second look against the suite. Its summary shows the
score, a text bar of the mutation time against the time threshold, and the
survivors. The whole log is kept as an artifact for 90 days. A final report job
adds one page for the run: every component's time and score, sorted by time.

**The gate.** Main is red when a component's score falls below its floor. A floor
is one number per component in `tools/mutation/floor/<component>.txt`. A person
raises it as survivors get fixed or accepted. Lowering it is also a commit, so a
regression shows as a diff with a reason. The nightly run turns red, and a red
main means: fix the score or lower the floor before the next merge. Nothing
blocks a merge by itself. The first floors come from one manual run of every
component, set a little under the measured score so that run-to-run noise does
not flip the result. A person approves the margin.

**Processing results.** Run history is the worklist. A person reads the summary
when a component changes, and sorts each survivor into one of three outcomes:

- A gap in the contract, which becomes a contract change. A person authorizes
  it, and an agent can draft the test.
- A change that alters no behavior, which a person adds to
  `tools/mutation/accepted/<component>.txt` after checking the reason.
- Dead code, which becomes a component change.

A survivor nobody fixes shows up again the next time its component runs. That
repetition is the accepted cost of keeping no state outside the history.

**The behavior judge.** After a component's mutation job, a model runs the
[behavior judge](../../.agents/skills/behavior-judge/SKILL.md) in CI on that
component's unreviewed survivors. It sorts each into the three outcomes with a
short reason and may propose test text. Calls go through the batch interface,
because a nightly result is not urgent and batching costs less. A cap on
survivors per run keeps a component with over a thousand from spending the
budget. The output goes to the job summary and an artifact, labeled a trial
judgment under [0100](0100-behavior-judgment.md) and
[0130](0130-survivor-risk-acceptance.md). It never writes to `accepted/` and
never edits a contract or a test. It needs an API key stored as a repository
secret, and it sends source text to the model provider.

**Agents.** An agent does not start `just mutate` unless a person asks. It reads
the nightly result when a task needs it. The rules that stay: an agent never adds
a line to the accepted lists, and a claim that a survivor is harmless needs the
scratch-copy check in "Trust but verify".

**Local runs.** `just mutate` stays in the `justfile` because the workflow calls
it. A person may still run it on one component, with `--short --cached`, when the
nightly result is too slow for the task.

## If this happens

- **The same survivors bury the new ones.** A person reports that repeated
  survivors hide what changed. Add `tools/mutation/baseline/<component>.txt`,
  listing the known survivors, and show only the survivors missing from it. The
  list overlaps the accepted lists, so decide then how the two relate.
- **A component job runs longer than 2 hours, or hits the 6-hour limit.** Split
  that component's run by source file, with one matrix entry per file. The tool
  already works one file at a time. Count mutants per file first, because a few
  files usually carry most of the cost. A split component gets one floor, from
  an extra job that adds the per-file counts. First try more workers, or keep the
  tool's per-file results in the Actions cache so a cut-off job resumes. Both are
  unproven.
- **A component is slow but under the line.** The time bar in its summary shows
  it moving toward 2 hours, so the split is planned and not a surprise.

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
- The workflow, its report job and the time bar need a real run to confirm. The
  gate, the floors and the judge are built after the first full run, because the
  floors come from it.
- Red main is a convention, not a block. It works only while someone owns the
  morning check.

## Not decided

1. **The floor margin.** How far under the first measured score to set each
   floor. Decide it from the first full run and a week of nightly results.
2. **The judge's survivor cap.** The number per run, and what happens to the
   survivors over it.
3. **Using a trend.** Whether to keep scores over nights to chart a trend.
   Artifacts keep the raw logs, so it can be rebuilt later.
