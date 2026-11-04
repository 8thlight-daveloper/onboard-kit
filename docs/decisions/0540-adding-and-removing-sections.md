---
category: ux
status: accepted
approval: person
affects: [catalog/editor, catalog/materializer, tui]
builds-on: ['0530', '0430', '0440', '0050', '0070']
---

# 0540. Add and remove sections in the Sections screen

Builds on [0530](0530-new-plan-from-core-set.md),
[0430](0430-reviewed-catalog-creation.md),
[0440](0440-stable-catalog-instance-editing.md),
[0050](0050-section-switch.md), and
[0070](0070-sub-sections.md).

## Context

The Sections screen only switches existing sections on and off. A person cannot
add a catalog section, remove one, or write one of their own from any screen.
Catalog creation and editing exist only as line prompts ([0430](0430-reviewed-catalog-creation.md),
[0450](0450-reviewed-existing-catalog-edit.md)). They have no select all and no
sub-sections. Nothing in the tool writes a custom section. The only way is to
write a section file by hand, or to edit the plan file by hand.

The catalog materializer and the catalog editor refuse any section that is not
a catalog section, and refuse a catalog section that has a parent. So a custom
section and a sub-section cannot be added by the code that exists.

The terminal screens also save each change as it is made. There is no draft that
holds changes back, so "nothing is written before the save" of 0430 has no
screen to live in yet.

## Decision

Adding and removing happens inside the Sections screen, with the keys the
Contacts screen already uses. There is no separate library tab.

| Key | Action |
| -- | -- |
| `a` | Open the add picker for catalog sections |
| `n` | Write a new section of your own |
| `A` | Add a part under the chosen section (a sub-section) |
| `x` | Remove the chosen section, after a preview and `y` |
| `space` | Switch on or off, as today |

The picker lists catalog sections by category, with search. Each row says what
is in the plan already, such as "in plan" or "in plan x2, add another". The
picker always lets a person add a section again. A catalog section and
a custom section appear in the same list with the same row and the same keys.
The row of a custom section says "your own".

### The first plan is a draft

Before the first save the plan is a draft. The screen shows two tabs, Sections
and Review. Questions, Walk, Overview and Contacts appear after the save. A
draft writes nothing. The core set of [0530](0530-new-plan-from-core-set.md)
is staged in it.

### Removing

- **Before the first save,** `x` removes the section. The preview names the
  parts under it, which go too, and warns that a custom section's text is lost.
- **After the save,** `x` reads "Switch off". The section's id stays held, as
  [0440](0440-stable-catalog-instance-editing.md) requires. A person can switch
  it on again. Nothing is deleted after the save, because answers and outcomes
  may refer to it.
- **A parent switched off** shows its parts as "off because the parent is off".
  The stored switches of the parts do not change, so switching the parent on
  again restores them.

### Ids

- The first add of a catalog section suggests its catalog id, which the person
  can edit.
- A second add of the same catalog section asks "What is this one for?" and
  suggests an id built from the answer, such as
  `repository-access-billing-api`. The person can edit it.
- Every suggested id is the person's input once they confirm it.
- A custom section's id is made from its title and shown as it is typed.
- A note beside the field says the id cannot change after the save.

### Sub-sections

- Only custom sections can be parts of another section. A catalog section stays
  flat, as the materializer requires today.
- Any section, catalog or custom, can be the parent.
- A part shows "Show this part only when" followed by a question of its parent,
  which is the rule of [0070](0070-sub-sections.md).

### Select all and deselect all

They work on the list the picker is showing, ask for a count and `y`, and act
only on the draft. Select all stages each catalog section once, with no
exceptions, since a person can add a section again afterward. Deselect all
spares custom sections.

### Custom sections

A custom section is a section in the plan with no `source`. Its text can be
edited until the first save. The form asks for the title, what the newcomer
does, and the steps. A step can carry a command, as the
[step schema](../reference/section-template-schema/steps.md) allows: a command
as an argument list with no shell, the environment it runs with, and the check
that it passed. The secrets check runs on every field the form accepts, as the
schema already requires for environment values. This decision does not say how
the form lays out commands, links and questions. That is the next decision, and
until it is made `n` should not ship.

## Why

- One list that says what is in the plan beats a library and a plan side by
  side, which would show the same fact twice. A sixth tab would also crowd a
  tab strip that is already about 70 columns wide.
- Embedded and custom sections work the same way, so a person learns one set of
  keys.
- Ids are permanent and chosen when the person knows least, so the screen shows
  the id as it forms and says it is permanent before the save.
- Removing a section after the save would lose answers. Switching off keeps them.
- No flag marks a section as repeatable. The picker allows any section again, so
  there is no list to keep up to date.

## Decided by a person

- **The core set** of [0530](0530-new-plan-from-core-set.md) is accepted.
- **Contract edits.** Select all, custom-section creation, a draft the screen can
  drive, and the core flag and groups of 0530 are authorized as
  [0550](0550-developer-onboarding-catalog-additions.md) records.
- **Cascade.** Turning a parent off turns off its parts in effect, as
  [0070](0070-sub-sections.md) says. The cascade is computed, not stored.
  [0440](0440-stable-catalog-instance-editing.md) means that the stored switch
  of each part stays as the person set it. Switching the parent back on then
  restores each part exactly, including a part the person had switched off
  alone. A stored cascade could not tell those apart.
- **A suggested id.** An id the screen proposes, shows, and lets the person edit
  and confirm is the person's own input. It is not an inferred id under
  [0430](0430-reviewed-catalog-creation.md).
- **Commands and secrets.** The custom form offers commands, and the secrets
  check runs on every field.
- **Parents and parts.** A custom section can be a part of any section. A
  catalog section is never a part, because its parent would change the content
  hash that holds its template baseline. A person may revisit this if people ask
  for it.
- **Review after the save.** Adding a section to a saved plan shows a short
  confirmation that names the section, its id and what changes, with no file
  bytes. The organization's own review of the plan file, in its pull or merge
  request, is the real review.
- **Plan id and file name.** The first run asks for the organization name and
  suggests a lowercase dashed form of it, such as `acme-corp`, for the person to
  edit and confirm. The file is `acme-corp.obk.yaml`. The tool refuses an
  existing file and asks for another name.
- **Reports.** A report shows a switched-off section that has recorded outcomes
  in its own group at the end, labeled "switched off", with its outcomes. An
  option can hide the group. The report code has not been read, so this is the
  intended behavior and not a description of what exists.

## Risks

- **A half-built custom form is worse than none.** No component validates a
  custom section yet, and no `lint` command is routed. Do not ship `n` before
  the check exists.
- **A wrong id after the save** can only be switched off and added again,
  which leaves the first one's history behind.
- **Select all can produce a plan of 59 sections that nobody finishes.** Skipping
  repeatable sections can also hide a plan with no repository access in it.

## Touches

- ADRs: [0530](0530-new-plan-from-core-set.md),
  [0430](0430-reviewed-catalog-creation.md),
  [0410](0410-explicit-catalog-instances.md),
  [0440](0440-stable-catalog-instance-editing.md),
  [0450](0450-reviewed-existing-catalog-edit.md),
  [0050](0050-section-switch.md), [0070](0070-sub-sections.md),
  [0030](0030-self-contained-plan-file.md), and
  [0320](0320-embedded-catalog-availability.md).
- Contracts: catalog reader, creator, materializer, editor and edit guide.
- Docs: `docs/reference/tui.md`, the onboarding catalog reference, and the
  section template schema pages.
- Code: `internal/tui`, and new acceptance scenarios.
- A new decision for custom-section authoring, and a design note for the draft.
