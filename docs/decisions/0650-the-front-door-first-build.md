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

**A new plan starts with three plan questions.** Before any section question the plan
asks `organization` and `team_name`, both required onboarder questions, and
`plan_source`, an optional free-text question for the repository address and the path of
the plan file. The first two block an invitation until they are answered, and the first
gives the header its name. This is the step of 0670 that does not need a contract. The
`extends` link, the structured source, and scope on sections are not built.

**Interim: an invitation runs the typed start flow.** `onboard-kit invitation.zip`,
or "Begin my onboarding", ends the screen and runs `start --invitation FILE --root .` with its typed lines. The screen says where the records will go before it
hands over. The newcomer's own screens are not built.

**Interim: records go in the tool's data folder.** The typed start flow needs a root
before it opens the invitation, and the plan id is inside it. So the front door passes
`<data folder>/onboarding` as the root, and records land in
`onboarding/records/<account>/` there. The data folder is `OBK_DATA_HOME`, then
`OBK_HOME`, then the platform data folder: `XDG_DATA_HOME/obk` or `~/.local/share/obk`
on Linux, and `~/Library/Application Support/obk` on macOS. The front door creates the
folder when it hands over. 0670 puts records at `records/<plan id>/<account>/`. That
needs the invitation read first, so it waits for the newcomer's own screens. Two
invitations for the same account name in different organizations share this root until
then.

**About is reached with `a`** from the first screen and from any help page. It lists
every place in use as data in `internal/command/places`, and lists the Downloads
folder as not used yet.

**The header names the role.** The onboarder's title is "Building: ACME Corp plan",
or "Building an onboarding plan" until the organization is answered. The newcomer's
is "Your onboarding at ACME Corp", or "Your onboarding". The onboarder's tabs are
Answer, Overview, Review, Sections, Contacts, Preparation, and Invitation. The
newcomer's are Your questions, Overview, and Review. The title updates as soon as
the organization is answered. A blank or deferred organization counts as not
answered. A name longer than 24 columns is cut with an ellipsis, because without a
cap the title disappeared at 120 columns on the later tabs. Control characters in
the name show as escapes.

**The newcomer's third tab stays Review, and the Answer key stays `w`.** 0610 says
Document. That screen holds the newcomer's blockers, deferred questions, and
problems beside the document preview, so Document alone would mislead. No free
letter appears in both "Answer" and "Your questions", so both people open the
screen with `w`, and `a` stays with Contacts and About. `u` for the newcomer is
open.

**Longer titles drop the key letters sooner.** With seven tabs the letters show at
120 columns and no longer at 110. At 80 and 100 columns they were already dropped.

**Code and files say Answer.** The answering screen's identifiers, files, and
wireframes were renamed from walk. The `walk` command and its package stay until
they retire. Three scenarios that waited for the old header text changed, with a
person's authorization: the narrow-window scenario, the newcomer scenario, and the
core plan scenario, whose wait for the progress count became a wait for the Answer
screen's position line, because the longer title pushes the count out at 110
columns.

## Consequences

- `just build` also writes `output/obk`. Release archives, the Homebrew formula, and
  the Nix package still need the alias. Homebrew was not installed where this was
  built, so the name `obk` is not yet checked against it.
- The three missing catalog sections need a catalog change, which needs contract
  changes that a person authorizes.
- Everything marked interim is replaced when the newcomer's screens and the
  reviewed creation draft land.
- The name of the newcomer's third tab is open in the terminal page, together with
  `u` as its key.
