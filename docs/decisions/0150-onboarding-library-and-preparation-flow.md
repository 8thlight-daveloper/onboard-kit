# 0150. Onboarding library and preparation flow

Status: accepted direction; implementation pending

Builds on [0005](0005-self-contained-install.md),
[0085](0085-walking-sections-and-recording-outcomes.md),
[0100](0100-behavior-judgment.md), and
[0140](0140-person-specific-preparation.md).
The target-date rules below supersede 0095's treatment of timing as due dates
for judging newcomer progress. Dates remain useful for planning and support.
[0160](0160-personal-data-and-preparation-approval.md) narrows recipient collection:
do not ask for a personal delivery address in the normal invitation flow.

## Context

The example templates demonstrate a format, not the complete onboarding
library. The release should carry the full catalog, including preparation by
the onboarder, rather than require another download. Preparation must support
both the initial invitation and changes after the newcomer starts.

## Decision

### Embed the full library

Embed all entries in the [onboarding catalog](../reference/onboarding-catalog.md),
including standard first-steps wording. Organization and team configuration
select what applies. Being embedded does not make every activity compulsory
or put every activity in every person's active plan.

Repositories, reading assignments, and other repeated activities need stable
instances for each configured item. Independent outcomes belong to separate
sections, as 0085 requires. Reordering a list must not attach an outcome to
another item. Specify expansion and identity before implementation; list
position alone is not an identity.

Onboarders supply organization-specific facts and instructions. Generic
templates contain no real organization addresses, credentials, or personal
details. Examples and defaults never silently become answers.

### Dates are goals

Timing windows set planning targets, not completion deadlines. Passing a target
does not fail a section, block progress, or change a recorded outcome. Actual
prerequisites and authorization still matter. A date cannot establish that
access exists or someone is ready to lead a deployment.

The newcomer sees neutral status and useful next actions. Onboarder views
identify missing arrangements, who owns them, and whom they block. A late
pairing assignment calls for the onboarder to arrange a partner, not a warning
that the newcomer failed to pair. Do not assign blame when the blocker is unknown.

Before-start targets chiefly describe preparation by the organization. They
must not imply that the newcomer owes work before their engagement begins.

### Include the preparation for an activity

Selecting an activity also surfaces the preparation needed to make it possible.
Pairing needs a partner and suitable ticket. Shadowing a deploy needs an
experienced lead and an opportunity. A newcomer leading a deploy needs an
experienced observer or support contact. Name those roles explicitly rather
than assigning an ambiguous deploy shadow.

Known shared arrangements may satisfy preparation; do not demand a duplicate
request for access already supplied. Unknown assignments remain visible with
an owner and target, not guessed values. An assignment does not verify the
newcomer's activity. Personal arrangements go in the preparation file, not the
shared plan or the newcomer's record.

This specifies a product relationship, not new step-level conditions or an
already implemented runtime. Selection and preparation need approved format
and contract increments.

### Guide invitation preparation

The onboarder follows one preparation flow:

1. Identify the newcomer: name, account identifier, and company email when known.
   Do not require an email account that has not yet been provisioned. The
   onboarder delivers the ZIP through an approved channel outside this flow;
   a personal delivery address is not required or collected by default.
2. Confirm team, discipline, manager, onboarding owner, and buddy as applicable.
3. Explicitly confirm the start date and remote, on-site, or hybrid arrangements.
   Establish known time zone constraints without guessing personal preferences.
4. Review selected activities from organization and team configuration.
5. Prepare equipment, access, badge or remote arrangements, and meeting invitations.
6. Supply known partners, reviewers, work items, and deployment opportunities.
7. Review missing information, separating what is needed to start from preparation
   that can be supplied later. Identify the person responsible for each gap.
8. Preview and generate the invitation with the plan, preparation, seeded record,
   and first-steps document, under 0140's ownership boundaries.

Use the same preparation data for later updates. The newcomer need not have
submitted a record first. Missing company email must not block an invitation
whose purpose includes obtaining that email. Initial instructions and a help
route must work before company access exists.

Generating an archive does not provision accounts, send meeting invitations,
or claim that the recipient completed anything. Future external actions need
their own explicit authorization. 0090's archive safety and identity
confirmation requirements still apply.

## Verification and delivery

Keep decisions, authored content, planned tests, verified tests, and runtime
implementation distinct. This ADR changes no protected contracts or model fields.
Deliver bounded increments under [0025](0025-contracts-and-tests.md):

- A whole-binary scenario must show the full library is available without an
  external template directory or download. Keep explicit custom templates working.
- Selection must omit irrelevant activities and their preparation, while preserving
  shared preparation needed by another selected activity.
- Repeated repository and document instances need independent outcomes and stable
  identity across list reorderings, additions, and removals.
- Invitation and update scenarios must preserve each owner's separate files and
  existing newcomer work, including unknown or deferred preparation.
- A passed target must not turn a verified or untried activity into failure.
  A missing partner must name the arrangement and its owner.

The preparation schema, per-person lookup and precedence, instance expansion,
and invitation update conflicts need explicit contracts before implementation.
