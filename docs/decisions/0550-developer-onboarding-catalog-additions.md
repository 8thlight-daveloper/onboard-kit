---
category: behavior
status: accepted
approval: person
affects: [catalog/reader]
builds-on: ['0320', '0530']
---

# 0550. Catalog additions for developer onboarding

Builds on [0320](0320-embedded-catalog-availability.md) and
[0530](0530-new-plan-from-core-set.md). It changes the fixed count of eighty
catalog identities in 0320.

## Context

Onboard-kit exists so that a team can stop keeping developer onboarding in a
shared document that goes out of date. The document usually covers getting the
application running, running the tests, getting access to systems, and
learning how the team works. A newcomer follows it while stuck, and feedback on
a wrong step is slow and awkward.

The kit should hold those steps in a plan the team keeps current. It should
also show the newcomer how the team works, and give the team a recorded way to
see that this happens.

The catalog covers much of this already. It does not cover four things:

- Attending a daily standup.
- Working with AI-assisted tools. Pairing, review and specification all change
  when a tool writes part of the code.
- Learning the codebase with a guided tutor.
- Joining the team's ticket refinement and prioritization.

The kit also should not duplicate the standard processes that HR and IT run:
the company account, calendar, device, working hours, and compliance training.
[0530](0530-new-plan-from-core-set.md) moves those into an opt-in group.

## Decision

Add the entries below to the catalog. Each is a section template in the existing
loader format, with generic content and onboarder questions for every
organization-specific fact. No entry names a vendor, except `codebase-tutor`.

| Entry | Result for the newcomer | Requires |
| -- | -- | -- |
| `standup-attendance` | Attend a standup and give an update | none |
| `ai-tool-policy` | Know which tools are approved, what code and data may go into them, how secrets are handled, and whether a change discloses AI use. Where AI use is not allowed, the section states that. | none |
| `ai-workflow-walkthrough` | Watch another developer's tools and flow, including the context they give, how they check the result, and one time it was wrong | `ai-tool-policy` |
| `ai-work-walkthrough-before-pr` | Run and read generated work and show what was checked, explain it to a coworker, and have the ticket owner confirm it matches the ticket, before the change goes to review | `ai-tool-policy` |
| `spec-pairing` | Pair on turning a ticket into acceptance criteria and limits that an agent can act on, then compare the result with the intent | `ai-tool-policy` |
| `agent-conventions-tour` | Read the team's instructions for coding agents, and propose one correction | `ai-tool-policy` |
| `ticket-writing-session` | Attend a refinement session as an observer, then contribute | none |
| `prioritization-session` | Attend a prioritization session and learn why work is chosen | none |
| `codebase-tutor` | Learn the codebase with a guided tutor that cites real files | `ai-tool-policy` |

Pairing as navigator and as driver use `ticket-pairing` twice with distinct ids,
as [0540](0540-adding-and-removing-sections.md) allows. They need no new entry.

### The codebase tutor

`codebase-tutor` names `repo-tutor`, a separate program that reads a repository
and teaches it to a new developer. This is the one entry that names a tool. It
sits beside the tool-neutral `codebase-orientation`, which stays for
organizations that use something else.

- **Separate program.** The step tells the newcomer to install and run the
  tutor. The kit never embeds it. The tutor is licensed under GPL version 3 or
  later and the kit under MIT, so embedding or linking it would put the kit
  binary under the GPL, and would break
  [0005](0005-self-contained-install.md). A separate program does neither.
- **Model access.** The tutor sends excerpts of the repository to a model
  through the person's installed agent tool. That is why the entry requires
  `ai-tool-policy`.
- **Delivery.** The tutor is about to have a first release, available as a
  download, through Nix, and through Homebrew. The entry's install step depends
  on that release. Do not publish the entry before the release exists.
- **Questions for the onboarder.** The repository to study, the starting goal,
  and which agent tool to use.
- **No data bridge.** The kit does not read the tutor's progress. A later
  decision could let a newcomer export it as an outcome under
  [0290](0290-explicit-outcome-recording.md). It would be the newcomer's explicit
  act, never an automatic report.

### Order of work

Add the entries in batches. The AI group depends on `ai-tool-policy`, so it
comes first.

1. `ai-tool-policy`.
2. `standup-attendance`.
3. `codebase-tutor`, after the tutor's release.
4. The rest of the AI group.
5. The two session entries.

Each batch needs authored content that the loader accepts, a catalog manifest
entry, contract fixtures, and availability scenarios.

## Changes to 0320

0320 fixes the catalog at exactly eighty identities. This decision replaces that
count with the identities in the manifest. The rule that identities are stable
and never reused stays. The section-template count grows from 59 by the number
of entries added. The preparation guides and tool-help topics do not change.

## What this decision does not claim

The kit's purpose includes running commands and sending messages for the
newcomer. [0065](0065-messages-and-evidence.md) designs steps that carry a message,
and [0020](0020-binary-structure-and-regenerative-components.md) places running
commands in a later package. This decision does not check how much of either
works. Until they do, these sections are checklists with checks and a record
the team can read. They do not run anything.

## Decided by a person

- **Contract edits.** A person authorized the contract changes this decision and
  [0530](0530-new-plan-from-core-set.md) and
  [0540](0540-adding-and-removing-sections.md) describe: the catalog manifest and
  reader, the creator, and the availability scenarios, for each batch above. Any
  other contract change needs its own authorization.
- **`ai-tool-policy` is in the default core.**
- **The two session entries are opt-in.** `ticket-writing-session` and
  `prioritization-session` are in an opt-in group, not the default core.
- **No separate `verify-generated-change` entry.** Running and reading a
  generated change, and showing what was checked, is part of
  `ai-work-walkthrough-before-pr`.
