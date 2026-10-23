# Invitation and preparation flow

Status: planned, not the current command implementation.

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

The preparation model and validator are implemented. Serializer contracts are
staged; their implementation remains pending. See the [preparation format](../reference/preparation-format.md).
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

Provide an onboarder view scoped to one person. Show missing assignments and
requests separately from verified newcomer outcomes. Saving preparation must
not rewrite the record or shared plan. Loading later preparation changes must
not guess that the newcomer completed an activity.

Test a partner assignment both before and after the newcomer's first saved
record. Test that a changed assignment is visible while personal answers,
notes, and outcomes are preserved. Conflicting same-owner Git edits remain
the editors' responsibility, not a new automatic merge feature.

### 3. Invitation packaging and startup

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

Next is the preparation format and its contract/test increment, not a workaround
in shared plan answers. Under [0025](../decisions/0025-contracts-and-tests.md),
contracts and acceptance scenarios need human approval and implementation needs
independent authorship. This document supplies the scope and expected behaviors;
it does not report those tests as written or verified.
