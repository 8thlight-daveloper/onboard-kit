# 0095. Reports

Status: proposed

Builds on [0085](0085-walking-sections-and-recording-outcomes.md) and [0090](0090-plan-and-onboarding-record.md). It records what a report is and the questions that are still open. Nothing here is built.

## Context

An onboarder wants to know where newcomers stand: "we need to roll AWS access out by 11/26, who hasn't done it yet?" Newcomers' onboarding records, kept at the plan's results path, hold the answers to that. A report reads them.

## Decision

### What a report is

A report is a **derived, read-only view over one or more onboarding records**. It is computed on demand and stored nowhere.

- Its inputs are the onboarding records, the plan for the section list and any due dates, and today's date from the driver.
- It is a pure function, so the same inputs give the same output. It reads files and changes none.
- It is not the assembled document, which is a plan written out for a person to follow. It is not a screen either, though a screen could show one.
- A newcomer's own overview is the same function applied to one record. The onboarder's report is the same function applied to many.
- It reads records and the plan and nothing else. It does not read git. A record's `plan.hash` says which plan the person was on.

### Who can run it

Whoever can read the folder of records can run it. The repository controls readership, and the tool enforces nothing more. A report is only as private as the repository, which the shared repository notice of 0085 already says to newcomers.

## Not decided

Each question has a leaning, which is a suggestion and not a decision.

1. **Axis.** By section answers "who has not verified AWS access yet", and by person answers "how is Jane doing". Both come from the same data. Which is the first view, and is the other a filter or a separate command? *Leaning:* section first, because the example that started this was a rollout. The person view is a filter on the same output.
2. **Deadlines.** "By 11/26" needs a due date on a section, written by the onboarder in the plan. It is content, so it changes the section's hash. Is it an absolute date, or relative to each newcomer's start date, or both? Timing already has relative windows, and an org-wide rollout is absolute. *Leaning:* both, as two fields, since they answer different questions.
3. **What needs attention.** Do `failed`, long-`deferred`, overdue, and changed-since-verified outcomes surface at the top? Does a newcomer who started and went quiet count, which the outcomes' `on` times would show? *Leaning:* yes to all, in one "needs attention" group above the rest.
4. **Notes.** Do `note` and `differently` appear? `differently` is the most useful text, because it says where the plan is wrong, so sections could be ranked by how often it is recorded. *Leaning:* show them, with a mode that leaves notes out for a report that will be shared more widely.
5. **Form.** A terminal table, markdown, CSV, or a TUI screen. A command that prints a table is the least work and is testable. A stable plain form also lets people script it, for example to chase overdue people. *Leaning:* a command with a table and a machine-readable form first, and a screen only if onboarders ask to drill in.
6. **Files with problems.** A folder holds files that fail to decode, come from an old plan version, belong to another plan, or point at a plan hash the plan no longer has. *Leaning:* skip them, and list what was skipped and why.
7. **History.** A record holds only the latest outcome. History comes from git, which the tool does not read. *Leaning:* none for now. People can use `git log`.
8. **Offboarding.** A report of who still holds access to a system is the other half of "who has not done it". See the Later item in 0090.
