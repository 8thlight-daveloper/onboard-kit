# Invitation and preparation flow

Status: preparation editing and explicit export implemented; guided authoring,
confirmed import and record-backed startup remain planned.

Decisions [0140](../decisions/0140-person-specific-preparation.md) and
[0150](../decisions/0150-onboarding-library-and-preparation-flow.md) establish
separate ownership, a full embedded library, and guided invitation preparation.
This view orders the work needed to deliver them. It is not approval to change
protected contracts or the domain model.

## Inputs, ownership, and writes

| Flow | Reads | Writes | Must leave alone |
| -- | -- | -- | -- |
| Configure the organization | Embedded library and organization plan | Organization plan | Existing preparation and records |
| Prepare one invitation | Organization plan and supplied recipient details | That person's preparation, initial record seed, and invitation archive | Other people's files and existing newcomer answers or outcomes |
| Open an invitation | Validated archive and any existing local files | Local plan copy, preparation copy, and the person's record through the approved import rules | Existing work until identity is confirmed; unrelated files |
| Update preparation | Organization plan and that person's current preparation; record only when needed for a read-only progress view | That person's preparation | Shared plan and all newcomer records |
| Walk onboarding | Organization plan copy, preparation, and record | Onboardee record | Plan and preparation |
| Review support needs | Plan, preparation, records, and an injected date | Nothing | Every input file |

The archive is a delivery container, not a fourth source of truth. The local
plan and preparation copies must support offline use. File sharing publishes
later changes; the tool does not silently run Git, send messages, or provision
accounts.

## Ordered delivery

### 1. Preparation format and ownership

The preparation model, validator, decoder, encoder, typed editor, and locator are
implemented. The preparation command supplies editing and a read-only combined
view of preparation and record outcomes. See the
[preparation format](../reference/preparation-format.md).
Define the person-specific artifact before adding personalized questions to
the existing plan editor. Specify identity, references to the plan and account,
the location beside the record, and which values are assignments rather than
newcomer-owned answers. Decide how start date and existing invitation seeds
relate to preparation without silently moving existing fields. Under
[0180](../decisions/0180-first-onboarding-use.md), preparation uses plannedStartDate;
first confirmed onboarding use will record firstOpenedAt without an actual-date
question. Existing record dates remain untouched in this increment.

Specify serialization, validation, and the view that combines the three inputs.
Declare the read and write set for each new or changed component. Specify
precedence only for supported overrides; reject ambiguous declarations.

The first test increment must distinguish two newcomers with different partners,
preserve record answers over updates, and reject a preparation file for the
wrong person or plan. Do not introduce an unrestricted generic overlay map.

### 2. Preparation editing and updates

The pure editor, [command flow](../reference/preparation-command.md), and combined
preparation/outcome view are implemented. Guided invitation preparation remains
later work.
Provide an onboarder view scoped to one person. Show missing assignments and
requests separately from verified newcomer outcomes. Saving preparation must
not rewrite the record or shared plan. Loading later preparation changes must
not guess that the newcomer completed an activity.

Test a partner assignment both before and after the newcomer's first saved
record. Test that a changed assignment is visible while personal answers,
notes, and outcomes are preserved. Conflicting same-owner Git edits remain
the editors' responsibility, not a new automatic merge feature.

### 3. Invitation packaging and startup

The opaque archive encoder and decoder of
[0240](../decisions/0240-invitation-archive-envelope.md) are implemented.
They do not validate payload readiness or authorize importing files.
Bootstrap markers, declared prerequisite validation, applicability, global action
scheduling, and flat contextual fragments are implemented under
[0260](../decisions/0260-prerequisites-and-action-order.md). Ordinary reference
documents also show explicit prerequisites without becoming action sequences.
Export packages the first-steps guide. Outcome-verification gates in command
sessions and record-backed startup remain required.

The pure first steps assembler now wraps the real bootstrap fragment with
standard help under [0280](../decisions/0280-first-steps-help-and-sources.md).
Read-only section walker views separate applicability, new-attempt waiting,
held outcomes and actual-definition comparisons. The pure outcome recorder now
rechecks direct verification prerequisites and updates only the returned record
under [0290](../decisions/0290-explicit-outcome-recording.md). These components do
not publish outcomes or start a record-backed command session.

[0300](../decisions/0300-explicit-invitation-export.md) specifies a bounded export
command with explicit pre-access help designation, identity-only seed and
exclusive archive publication. The command and independent binary scenarios
are implemented; see [invitation export](../reference/invitation-export.md).
It does not settle confirmed import or preparation-answer bindings.

Build the guided flow using the same preparation format and editor, then add
the archive writer and reader. The allowed entry layout must include preparation
and satisfy 0090's bounded-entry, size, and untrusted-path rules.

Confirm the account and destination before writes. Test wrong identity, invalid
archives, missing company access, and existing records. Define stale or conflicting
preparation import behavior before implementing it; do not silently replace a
newer preparation copy with an older invitation.

Update standard first-steps wording to explain all three artifacts, initial help,
and what works before company access. The person must not need access to the
organization's repository to learn how to obtain that access.

### 4. Full library and selection

Author every entry in the [catalog](../reference/onboarding-catalog.md), including
onboarder preparation and standard tool instructions. Embed the full set, retaining
custom template loading. Check availability against the catalog, not against
whatever happens to be present in the implementation directory.

The full normative content and independent availability contracts are authored
under [0320](../decisions/0320-embedded-catalog-availability.md): 59 actual Section
templates, 19 preparation guides and two static tool-help topics. The real loader
accepts every Section and current combined-plan validation reports no problems.
Compiled availability and its read-only catalog command remain pending. This
does not imply selected activities or generated personal preparation.

Add organization and team selection through an approved format. Including a
newcomer activity surfaces its preparation; unrelated preparation stays out.
Do not ask newcomers to supply organizational facts merely to fill a template.

### 5. Repeated activities

Define stable identities for configured repositories, documents, introductions,
and other independent instances. Test reorder, add, and remove operations with
existing outcomes. A list index or mutable display title must not reassign
someone's progress to a different resource.

### 6. Goals and support views

Use target dates for planning. Show missing arrangements and their owners in
the onboarder view; show neutral next actions in the newcomer view. Verify
that time passing does not manufacture failure or override real prerequisites.

### 7. Verify the release artifact

Run the approved binary acceptance scenarios against the exact release file
using [local release verification](../stack/release.md#local-verification).
Include first-run coverage outside a source checkout, without a template folder
or development tools. Keep file-based operation offline-capable; do not claim
network isolation from an empty PATH alone.

## Approval boundary

The remaining increments need explicit contracts, not a workaround in shared
plan answers. Under [0025](../decisions/0025-contracts-and-tests.md),
contracts and acceptance scenarios need human approval and implementation needs
independent authorship. This document supplies the scope and expected behaviors;
it does not report those tests as written or verified.
