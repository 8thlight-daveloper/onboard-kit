---
category: ux
status: proposed
affects: [command/create, command/frontdoor, command/places, command/registry, command/root, tui]
builds-on: ['0530', '0610', '0630', '0640']
---

# 0650. The front door, first build

Builds on [0530](0530-new-plan-from-core-set.md),
[0610](0610-people-work-in-screens-and-flags-serve-scripts.md),
[0630](0630-the-preparation-tab-saves-what-was-reviewed.md), and
[0640](0640-the-invitation-tab-writes-an-invitation.md). It records the choices
made while building the entry that 0610 decided, and the parts that are interim.
An agent made each choice while the person who owns the decisions was away. They
are proposals until that person accepts them.

## Decision

**How the tool knows what a file is.** An archive signature at the start makes an
invitation. Otherwise the plan decoder reads the file. A file it reads as a
record or a preparation is refused with a sentence that names what it is. A file
with a top-level title or summary is a section template, and is refused as
neither a plan nor an invitation. A plan with no sections is still a plan. Without
the title and summary rule, the lenient plan decoder accepted a section template
as a plan, and the editor would then have rewritten it.

**The plan list is a plain file.** `plans.txt` in the state home lists known plans
with the time each was opened. It is documented in
[the plan list](../reference/plan-list.md). A damaged list is reported on the first
screen and on About, and is never rewritten.

**A plan is recorded when a person opens or creates it at a terminal.** `new --out`
with no terminal records nothing. A script that records in the real state home
would make the check write to a person's own folder.

**Interim: a new plan is the fixed core set, written after a path confirmation.**
`new` stages the 0530 core set, shows the path, and writes only after `y`. It never
overwrites, and opens the plan in the editor. The reviewed creation draft of 0530
is not built. The core ids live in `internal/command/create/core.go` until the
catalog manifest owns them. 28 of the 31 sections in 0530 exist in the catalog.
`standup-attendance`, `codebase-tutor`, and `ai-tool-policy` are skipped, and a test
pins that exact set. The two pairing sessions are `ticket-pairing-navigator` and
`ticket-pairing-driver`. `new --templates DIR` still works while existing scenarios
use it.

**Interim: an invitation runs the typed start flow.** `onboard-kit invitation.zip`,
or "Begin my onboarding", ends the screen and runs `start --invitation FILE --root .` with its typed lines. The screen says where the records will go before it
hands over. The newcomer's own screens are not built.

**About is reached with `a`** from the first screen and from any help page. It lists
every place in use as data in `internal/command/places`, and lists the Downloads
folder as not used yet.

**Interim: no role header yet.** "Building: Acme plan" and the Answer tab of 0610
are not built. The screens keep their current tab names.

## Consequences

- `just build` also writes `output/obk`. Release archives, the Homebrew formula, and
  the Nix package still need the alias. Homebrew was not installed where this was
  built, so the name `obk` is not yet checked against it.
- The three missing catalog sections need a catalog change, which needs contract
  changes that a person authorizes.
- Everything marked interim is replaced when the newcomer's screens and the
  reviewed creation draft land.
