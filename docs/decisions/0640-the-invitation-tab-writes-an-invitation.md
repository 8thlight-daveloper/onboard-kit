---
category: ux
status: proposed
affects: [command/invite, command/walk, tui]
builds-on: ['0300', '0610', '0630']
---

# 0640. The Invitation tab writes an invitation

Builds on [0300](0300-explicit-invitation-export.md),
[0610](0610-people-work-in-screens-and-flags-serve-scripts.md), and
[0630](0630-the-preparation-tab-saves-what-was-reviewed.md). It records the choices
made while building the Invitation tab that those records leave open.

## Context

0610 puts invitation export in a tab, so an onboarder no longer runs `prepare` and
`invite export` by hand. The export itself is 0300, and its rules do not change.
Building the tab met five cases that no record settles. An agent made each choice
while the person who owns the decisions was away. They are proposals until that
person accepts them.

## Decision

**The command package offers two functions.** `invite.Export` does what the command
does after it reads its arguments, and the command calls it. `invite.Check` reads
the same files with the same bounded readers and runs the same checks, and writes
nothing. The tab needs readiness before anything is written, so it needs a check
that is not an export. The command's behavior and contract do not change.

**Readiness shows every blocker at once.** The export stops at the first failing
phase. The check collects preparation errors, onboarder blockers, and a missing
owner together, so the list says "not ready, 3 to fix". A person who is not ready
opens a page that lists what blocks, and not the form.

**The form takes pre-access help from named contacts only.** Only a contact with a
name and an email or Slack handle is offered, as the export requires. The field
starts as "required, not chosen". The pre-access contact goes first in the help
order. The guide beside the form is built by the preparation reviewer with the
picked help.

**A second step comes before the write.** The file name is typed, so `y` cannot be
the write key on the form. Enter or ctrl+s opens a summary beside the guide that
names the whole path, and `y` there writes the file. If the file exists, the tab
offers the next free name, `-2` and then `-3`. It never overwrites.

**The result leads with the line for the newcomer.** The line reads "Give this file
to NAME. They run: obk PATH", from one constant. A second line says the tool sends
nothing. The path is the one the onboarder chose. The newcomer saves the file
elsewhere, so that path may mislead them. Whether to show only the file name is
open.

**A seventh tab can window the header.** At 60 to 80 columns the strip is still too
wide when a late tab is current. The header shows the tabs around the current one,
with an ellipsis on each cut side. This goes one step past the clipping in 0630.

## Consequences

- Readiness runs again each time the tab opens, and a half-filled form is dropped
  then. A plan path that contains `..` lists everyone as "cannot be checked", and
  the page says why.
- The scenario proves a file exists because a second export refuses to replace it.
  Checking the zip's contents needs test data files that no person has approved.
- Scrolling the guide to the picked help after a pick is open in the terminal page.

## Not decided

1. **The path in the newcomer's line.** The onboarder's path, or only the file
   name, since the newcomer saves it in their own folder.
