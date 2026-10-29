# 0530. A new plan starts from a core set of the catalog

Status: proposed. A person has not yet approved it.

Builds on [0150](0150-onboarding-library-and-preparation-flow.md),
[0320](0320-embedded-catalog-availability.md),
[0410](0410-explicit-catalog-instances.md), and
[0430](0430-reviewed-catalog-creation.md).

## Context

The binary carries 59 section templates, yet a person with only the binary
cannot start a plan from them, since `onboard-kit new` currently needs
`--templates FOLDER` specified. `onboard-kit catalog create` can use the
embedded templates, but it starts with nothing selected. The person must
know each exact ID and add it by hand.

A manual test therefore borrows a folder of example files, and the plan it
builds has five sections. That is not what a real onboarder starts from.

Starting from nothing asks a person to choose from a long list before they know
what the plan looks like. Trimming a plan that is already useful is easier than
building one from a blank page. A starting plan also shows how sections fit
together before the person changes any of them.

[0320](0320-embedded-catalog-availability.md) says availability must not
silently select activities. [0120](0120-defaults-require-acceptance.md) says a
default never becomes an answer without acceptance. This decision keeps both.

## Decision

A new plan starts with a core set of catalog sections already staged in the
draft. The person reviews the exact plan, removes what they do not want, adds
what they do, and saves. Nothing is written before the save at the review.

- **The start.** With no `--templates` option, `new` opens the reviewed
  creation draft of 0430 with the core set staged. Each staged section uses its
  catalog ID as its activity ID. The review shows the complete plan bytes and
  the preparation guidance it brings, as 0430 requires. Save publishes one new
  file and never overwrites.
- **Trimming.** Before the save, a person removes a staged section freely,
  because nothing is saved yet. After the save, the Sections screen turns a
  section off, as 0410 and 0440 describe.
- **An empty start stays available.** `catalog create` keeps its empty draft for
  a person who wants to choose every section. `new` with `--templates` is
  unchanged.
- **The core set.** It holds the sections that suit any newcomer in any
  discipline, so no one starts with software or design work they do not do. The
  first proposal:
  - `onboarding-welcome`
  - `onboarding-help`
  - `onboarding-owner-meeting`
  - `onboarding-buddy-meeting`
  - `people-map`
  - `working-agreements`
  - `company-email`
  - `company-identity`
  - `team-chat`
  - `working-hours`
  - `security-orientation`
  - `week-one-check-in`
- **Groups to add.** A person adds software work, tools and repositories, or
  design work as a group, by the same add action. A group is a list of catalog
  IDs and selects nothing on its own.
- **Select all and deselect all.** The draft offers both, for the whole list
  and for each group. Select all stages every section template in scope that is
  not already staged, once each, with no exceptions. A person can add a
  section again from the picker, as [0540](0540-adding-and-removing-sections.md)
  describes. Deselect all removes every staged section in
  scope. Both are explicit actions, and a person can undo them by the other
  one or by removing single sections. Neither touches preparation guides or
  tool help, which are not sections. With these, the choice of default matters
  less, since a person who wants everything or nothing is one action away.
- **No silent selection.** The core set is an offer. The review shows every
  staged section, and nothing becomes part of a plan until the person saves.
  Select all stages sections only in the draft, and the review lists them like
  any others.

## Where the core set lives

The catalog manifest is the owner of catalog identity and classification
([0320](0320-embedded-catalog-availability.md)), so it should carry the core
flag and the group names. Putting them in the creator would hide a product
choice in code. This changes the catalog reader contract and the creator
contract, so it needs a person to authorize those contract edits before any
build. The contract tests must then pin the exact core set, so a regenerated
component cannot choose a different one.

The line-oriented draft of 0430 has no select all or deselect all, and its
grammar is pinned by the creator contract. Adding them is a contract change too,
for a person to authorize.

## Consequences

- A new person gets a useful plan from the binary alone, which restores what
  [0005](0005-self-contained-install.md) promises.
- Manual and acceptance testing can start from a plan the system generated, so
  the Acme example folder and the demo seed are no longer needed to start.
- The first-run screen of the bare command (a later decision) calls this
  creation path, so it must exist first.
- The core set is a product choice that someone will want to change. Each change
  edits the manifest and its contract fixtures, and no code.

## Not decided

1. **Sub-sections.** [0540](0540-adding-and-removing-sections.md) decides how a
   person adds and removes them. How a person writes a custom section is a
   later decision.
2. **The final core list.** The twelve above are a proposal. The onboarder
   should confirm them, and a person should decide whether a calendar, video
   meetings, and device entry belong.
3. **The plan identity.** Decided in 0540: the first run suggests it from the
   organization name, and the person confirms it.
4. **Group names.** The groups above follow the catalog categories. Whether they
   should follow the categories exactly is open.
