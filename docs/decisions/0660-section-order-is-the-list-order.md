---
category: behavior
status: proposed
affects: [catalog/editor, catalog/materializer, catalog/reader, document/assembler, plan/decoder, plan/editor, plan/encoder, plan/instantiator, plan/refresher, plan/structurevalidator, section/hasher, section/loader, section/scheduler, section/selector, tui]
builds-on: ['0260', '0530', '0540', '0610']
---

# 0660. Section order is the list order

Builds on [0260](0260-prerequisites-and-action-order.md),
[0530](0530-new-plan-from-core-set.md),
[0540](0540-adding-and-removing-sections.md), and
[0610](0610-people-work-in-screens-and-flags-serve-scripts.md). It replaces the
`id` and `after` ordering inside a timing window, and keeps the windows.

## Context

A new plan from the embedded catalog reads in an order no one chose.

- **The order inside a window is the alphabet.** The reading order sorts by timing,
  then `after`, then `id`. The scheduler's tiebreak ends with the exact `id`. The
  section schema says there is no numeric position to maintain, so a plan's own
  list order is not used.
- **The catalog sets no other order.** None of its 59 section templates declares
  `requires` or a parent. 29 have the timing `anytime`, which sorts after every
  dated window. So `repository-access` reads after `application-local`, and
  `ci-access` reads before `repository-access`.
- **An onboarder cannot change it.** The only ways are to invent `requires` edges
  or to change a timing, and neither says what the person means.

The catalog has the information for a better default: a category and a target on
every entry. A person should still decide the final order for their team.

## Decision

**Inside a timing window, the plan's list order is the order.** The windows stay as
the selector contract has them: `before-day-one`, `week-one`, `days-8-to-30`,
`days-31-to-90`, then `anytime`. Within one window, sections appear in the order of
the plan's `sections` list. The `id` tiebreak goes away, because a list position is
unique. Sub-sections still follow their parent, in the order of the list.

**The action order uses the same position.** `requires` still comes first, then
bootstrap, then the window. The rank in the reading order, now the list position,
replaces the `id` as the last tiebreak.

**The `after` field is removed.** Nothing has been released, so there is no deprecation.
The format, the loader, the encoder, the hasher, the validators, the selector, and the
scheduler stop knowing it. A file that still has an `after` key carries it as an ordinary
unknown field, which the plan keeps and never reads. List position does what `after` did.

**`requires` is for real prerequisites only.** An edge gates the dependent: it waits
until the prerequisite has a verified outcome, and switching off a prerequisite
switches off its dependents. A section that merely reads better after another does not
need an edge. Position says that.

**A section always has a window, and a new one lands at the end of it.** A section with
no `timing` is `anytime`, which is a window. A section added from the catalog lands where
the default sequence puts it among the sections already in its window, or at the end
when it has no place there. A section written by hand asks for its window when it is
added. Sections of different windows never interleave, so an empty window has no
position to settle. The Sections screen shows each row's window, and warns on a row
that reads before something it requires.

**A new plan starts in a logical order.** The catalog owns a default sequence, and
the materializer writes the sections of a new plan in it. The proposal for that
sequence, and for the timing and `requires` changes it needs, is
[the default order of the catalog sections](../reference/catalog-default-order.md).
A section added later lands where the default sequence puts it among the sections
already in its window.

**The onboarder reorders on the Sections screen.** A move stays inside one window.
To move a section to another window, change its timing. Order and timing stay
separate and visible. The newcomer cannot reorder, because the order shapes every
newcomer's document and first steps. The screen shows a warning on a row that
reads before something it requires. It does not refuse the move, because the reading
order is not an execution order, as 0260 says.

**A template update keeps the person's order.** The plan refresher places a section
that is new in an update by the default sequence among its neighbors, and never moves
a section the person already has.

## Consequences

- The selector, scheduler, and assembler contracts change where they say `id` and
  `after` decide order. The materializer, catalog editor, plan editor, and plan
  refresher change to write, insert, move, and preserve position. These are
  contract changes that a person authorizes.
- The catalog's timings and `requires` change as the proposal lists. A plan made
  before the change sees a changed template baseline, so this is best done before
  the first release.
- The Sections screen gains a move action. It uses a grab key and arrows, with no
  chord, in the style of the focus zones on the Answer screen.
- The schema page that says there is no position to maintain needs a new sentence.
- Removing `after` changes the domain model's section, which is slow-changing and
  needs this record. It also changes the example plan and the fixtures that carry an
  `after` key. A person authorizes those contract and fixture edits.
