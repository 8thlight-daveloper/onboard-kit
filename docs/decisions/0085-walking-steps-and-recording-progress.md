# 0085. Walking steps and recording progress

Status: proposed

Builds on [0030](0030-self-contained-plan-file.md), [0055](0055-variable-audience.md), [0060](0060-personal-plan.md), [0065](0065-messages-and-evidence.md), and [0075](0075-refreshing-a-personal-plan.md). It answers the progress record that 0065 and 0075 leave open, and it is the first stage of the onboardee runtime, [#28](https://github.com/8thlight-daveloper/onboard-kit/issues/28).

## Context

The onboardee's walk asks questions and stops there. The plan already holds the steps the newcomer has to do, each with an owner, notes, a command or a link, and checks. Nothing records whether a step is done, so the newcomer cannot see their progress, and the organization cannot see that a step was skipped or went stale.

This ADR is proposed and not accepted. It fixes the shape of the progress record, which becomes part of the plan file, so it deserves a person's review before the components are built.

## Decision

### Progress is a list in the personal plan

A personal plan may hold a top-level `progress` list. The org plan never does. Each entry is about one step.

```yaml
progress:
  - step: access.source-control
    status: done
    on: "2026-10-21T14:05:00Z"
    sectionHash: 5d3c...
    note: Asked in the IT channel
  - step: clone-and-run-tests.run-tests
    status: skipped
    on: "2026-10-21T15:30:00Z"
    sectionHash: 91ab...
    note: Waiting on a laptop
```

- `step` is `<section id>.<step id>`. Ids are permanent, see 0075.
- `status` is `done` or `skipped`. A step with no entry is not done, so there is no `todo` value to keep in step.
- `on` is a UTC time in RFC 3339 form. A driver passes it in, because the core reads no clock.
- `sectionHash` is the section's content hash when the entry was made, the `source.hash` the section carries.
- `note` is optional text. A skip needs one. `evidence` is optional text, such as a link, and 0065 says which checks require it.
- Entries are sorted by `step`, and each step has at most one.

Progress is kept apart from answers, as 0060 anticipated. It has nothing to do with the onboarder's answers.

### A refresh keeps progress and flags what changed

A refresh takes the section content from the org plan and leaves `progress` as it is. A step is **changed since it was finished** when its entry's `sectionHash` differs from the section's current hash. The comparison is made when the progress is read, so the refresher needs no new rule beyond keeping the list. A progress entry for a step that no longer exists stays, and a prune may remove it. The validator reports a duplicate entry as an error.

### Two new components

- **A step finder** returns the steps of the active sections in the order the section selector gives, each with its resolved text and one state: not done, done, skipped, or changed since it was finished. It reads the plan and changes nothing.
- **A progress recorder** returns a new plan with one step marked `done` or `skipped`, with the time, the note, and the evidence passed in, or with the entry removed to undo it. It refuses a step the plan does not have, and a skip with no note.

### The walk of steps

The onboardee gets a Steps screen, a tab like the others. It lists the steps by section, in order, with their state. A person opens a step to read it: what to do, who owns it, who to ask if stuck, and the command or link. They then mark it done or skip it with a reason, or undo a mark.

### What a check does in this stage

The tool runs nothing in this stage. The action runner, [#29](https://github.com/8thlight-daveloper/onboard-kit/issues/29), and its trust rules come later.

- An `ask` check is put to the person, and the step can be marked done when every answer is yes.
- A `run` check is shown with its command. The person runs it themselves and says whether it passed.
- An `env` check is shown by name. The person says whether it is set. The tool never reads the value.

So a check is advice that the person confirms, and marking a step done is the person's own statement.

### Messages and evidence

A step's `message` and the clipboard help of 0065 follow in the next stage and use the same record. The `evidence` field is already there for them.

## Why

- One list in the one plan file keeps the plan self-contained, as 0030 requires, and it travels with the person.
- A section hash makes "this step changed since you finished it" a comparison and not a stored flag, so it can never be stale.
- Running nothing first keeps this stage free of the trust questions, and it is already useful: a newcomer sees what is done, what is left, and who to ask.

## Consequences

- **The domain model changes.** A Plan gains `progress`, a list of entries with the fields above. A computed Step state type is added. This needs `slow-change`.
- **Contracts**, each its own small increment: the decoder and encoder read and write `progress`, the step finder, the progress recorder, the refresher (progress survives), and the validator (a duplicate entry, a step that does not exist is a warning).
- **The TUI** gains the Steps screen for the onboardee, and a way to undo a mark.
- **The plan hash.** The record is not part of any section's hash, and the encoder writes it after the sections.

## Not decided

- Whether an onboarder can see a newcomer's progress, and how it would be shared. This ADR keeps progress on the newcomer's machine.
- Whether `skipped` needs a reason in the org's view or only the newcomer's.
- Whether a done step is undone when its section changes, or only flagged. This ADR only flags.
- The evidence patterns for Teams and email, which 0065 leaves open.
