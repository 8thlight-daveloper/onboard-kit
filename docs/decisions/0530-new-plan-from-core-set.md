---
category: ux
status: accepted
approval: person
note: The core list is the first version. A later change edits the catalog manifest and its fixtures.
affects: [command/create, catalog/creator]
builds-on: ['0150', '0320', '0410', '0430']
---

# 0530. A new plan starts from a core set of the catalog

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

The purpose of the kit is to let a team stop keeping developer onboarding in a
shared document that goes out of date: getting the application running, running
the tests, getting access to systems, and learning how the team works. It
should also make sure the team does its part. Running commands and sending
messages for the newcomer are goals, and this decision does not rely on them.

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
- **The core set.** It is a developer onboarding path. It holds what a
  developer's team owns: team orientation and culture, access and setup,
  learning the code and contributing, practice, and feedback with
  accountability. It leaves out what HR and IT already run, so a newcomer does
  not get two sets of instructions that can disagree. The first proposal, with
  `ticket-pairing` staged twice, once as navigator and once as driver:
  - **Team and culture:** `onboarding-welcome`, `onboarding-help`,
    `onboarding-owner-meeting`, `onboarding-buddy-meeting`, `people-map`,
    `working-agreements`, `regular-meetings`, `team-chat`, and
    `standup-attendance` (new in [0550](0550-developer-onboarding-catalog-additions.md)).
  - **Access and setup:** `source-host-account`, `repository-access`,
    `ticket-system-access`, `ci-access`, `environment-access`,
    `documentation-access`, `project-dependencies`, `application-local`,
    `project-tests`.
  - **Learn the code and contribute:** `codebase-tutor` (new), `contribution-practices`,
    `ticket-pairing` twice, `pr-post`, and `ai-tool-policy` (new).
  - **Practice:** `deploy-shadow`, `deploy-lead`, `retrospective-participation`.
  - **Feedback and accountability:** `week-one-check-in`, `day-thirty-review`,
    `day-ninety-review`, `onboarding-correction`.
- **Groups to add.** A group is a list of catalog IDs that selects nothing on its
  own. A person adds one by the same add action.
  - **Company setup, usually handled by HR or IT:** `company-email`,
    `company-identity`, `company-calendar`, `video-meetings`, `working-device`,
    `working-hours`, `security-orientation`, `office-access`, `internal-network`.
    Organizations with no HR or IT process for these add the group.
  - **Systems access:** `hosting-access`, `observability-access`,
    `secrets-access`, `artifact-access`.
  - **AI-assisted development:** `ai-workflow-walkthrough`,
    `ai-work-walkthrough-before-pr`, `spec-pairing`, `agent-conventions-tour`.
  - **Team practices:** `ticket-writing-session`, `prioritization-session`,
    `teammate-introduction`, `product-orientation`.
  - **Design work:** the existing design sections.
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
2. **Changes to the core list.** The list above is accepted as the first
   version. The new entries it names need
   [0550](0550-developer-onboarding-catalog-additions.md) first, so a first
   build can stage only the entries that exist.
3. **The plan identity.** Decided in 0540: the first run suggests it from the
   organization name, and the person confirms it.
4. **Group names.** The groups above follow the catalog categories. Whether they
   should follow the categories exactly is open.
