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

### Due dates come from timing and a start date

A record holds a `startDate`. A section's due date is the start date plus the last day of its timing window: day 0 for `before-day-one`, day 7 for `week-one`, day 30 for `days-8-to-30`, and day 90 for `days-31-to-90`. A section with `anytime` has no due date. Nothing is added to a section.

- The onboarder gives the start date when they make the invitation. `invite` takes `--start`, defaults it to today, and prints the date it used and the due dates that follow from it, so the onboarder is told and can correct it.
- A section is **overdue** when its due date has passed and its outcome is not `verified`. It is **coming due** when its due date is near. The driver passes today's date in, as always.
- An absolute deadline that is the same for everyone, such as a rollout by a fixed date, is not covered. See Not decided.

### What needs attention

The report puts one group above the rest: sections that are `failed`, overdue, coming due, long `deferred`, changed since they were verified, or whose newcomer has gone quiet. The outcomes' times show a newcomer who has gone quiet.

### Notes

The report shows `note` and `differently`, and can leave them out for a report that will be shared more widely. `differently` is the most useful text, because it says where the plan is wrong, so sections can be ranked by how often it is recorded.

### Files with problems, and history

A folder holds files that fail to decode, come from an old plan version, belong to another plan, or point at a plan hash the plan no longer has. The report skips them and lists what it skipped and why. A record holds only the latest outcome, and history comes from git, which the tool does not read, so the report has none.

### A view that depends on the data

A tool may show different things for different data, if it stays predictable. The view is a pure function of the records it can read. It says what it saw, such as "3 unfinished onboardings in `records/`". It adds a notice and a screen to open and does not replace the screen a person expected, as the prune notice does on the Review. The same view can also be asked for on purpose.

So when an onboarder opens a plan and unfinished records exist at the results path, the tool shows a quiet notice and an Onboardings screen. It lists accounts with their progress, unfinished ones first, then by account.

### Two views of the same data

- **The person view** answers "who is still onboarding and might need help?" The action is to help one person. It is the first view, and the first iteration.
- **The section view** answers "how is the plan doing?" The action is to fix the plan. It counts, for each section, how many newcomers verified it, failed it, deferred it, and recorded something different. A section where everyone recorded something different needs its instructions fixed. It also answers "who has not verified AWS access yet", by listing the names behind a count.
- The section view shows counts and not percentages. With a handful of newcomers, "3 of 3" is honest and "100%" invites over-reading.
- A one-line notice on the plan, such as "3 unfinished onboardings, 1 needs attention", points to the screen. The numbers need a screen and not a status line. The section view is a second tab on the same screen, and a second iteration, once there is real data to judge it by.

## Not decided

- **Form.** A terminal table, markdown, CSV, or a screen. The first iteration will be a command that prints a table, with a machine-readable form, and the Onboardings screen above. Everything else waits to be seen.
- **Coming due.** How near counts as coming due.
- **Absolute deadlines.** A date that is the same for everyone, such as a rollout by a fixed date, is not expressible, because every due date is relative to a person's start date. A later field on a section could carry one.
- **Offboarding.** A report of who was given access is possible now, because verified outcomes say it. Checking that access is still in place, or revoking it, needs the plan to say how. A section could later carry instructions to revoke and to verify the revocation. See the Later item in 0090.
