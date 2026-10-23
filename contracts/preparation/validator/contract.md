# Preparation validator contract

Status: R1 approved by the person; planned-date meaning settled by 0180.
Component: `preparation/validator`.

Authority: [0140](../../../docs/decisions/0140-person-specific-preparation.md),
[0160](../../../docs/decisions/0160-personal-data-and-preparation-approval.md), and
[0180](../../../docs/decisions/0180-first-onboarding-use.md). The approved bounded
shape is described in the [preparation format](../../../docs/reference/preparation-format.md).
Each fixture cites its governing decision and this contract.

## Boundary and interface

`Validate(context, preparation) -> problems`.

The contract consists of this description, testdata/cases.json, validator_test.go,
and api.go with wire_test.go. The types live in internal/model; Context and its
optional record identity are contract input projections. Tests use plain JSON
fixture data and compare existing model.Problem fields. A diagnostic path is
stored in Detail under the key path, with severity error.

The wiring binds the interface to the independently authored implementation.
During a rebuild, run the suite against an empty stub before implementing it.
Valid-input controls may pass against that stub; invalid-input cases must fail.

The context is a trusted projection supplied after validating the plan. It holds
the expected plan id and account, known contact/team/section ids, and optionally
the identity of an existing record. It does not contain record answers, outcomes,
notes, or a filesystem path. The driver must take the expected account from the
explicit command or existing record, not copy it from the untrusted preparation.

Reads: SchemaVersion, Kind, Account, PlanID, Recipient, and Arrangements of
model.Preparation and every context field. UnknownFields is not interpreted. Writes: none,
inside or outside any plan. Returns newly allocated diagnostics only. No input
or output, clock, network, process, saved state, repair, or normalization.

## Validation phases

1. Unsupported schemaVersion produces only preparation.version at schemaVersion.
   Supported version is 1.
2. A kind other than the exact text preparation produces only preparation.kind
   at kind.
3. Identity checks run in account, planId, record.account, record.planId order.
   If any fail, return only those identity problems; do not inspect details.
4. Check recipient fields in plannedStartDate, team, manager, onboardingOwner, buddy,
   workLocation order, then arrangements in input order. Within an arrangement,
   use id, section, kind, status, owner, contacts, reference, targetDate order.

Each problem has a stable code and path and a nonblank useful message. Messages
explain the issue and repair without echoing supplied field values. Paths use
fixed field names and zero-based list indexes, not arbitrary user text. Return
all applicable detail problems in the stated order, without normalizing input.
An empty problem slice or nil both mean valid.

## Identity

Account is nonempty and matches `[A-Za-z0-9][A-Za-z0-9._-]*`. It is compared
exactly, without trimming, case folding, or inferring it from email. An invalid
account produces preparation.account at account; do not add account-mismatch
for the same field. A valid but different account produces
preparation.account-mismatch there.

An empty planId produces preparation.plan at planId. A different nonempty planId
produces preparation.plan-mismatch there. This validator does not impose UUID
syntax on the trusted plan identity.

If context.Record exists, its Account and PlanID must match the expected context.
Report preparation.record-account at record.account and preparation.record-plan
at record.planId, respectively. No record is necessary to prepare an invitation.

## Recipient

Empty optional fields are allowed. Supplied plannedStartDate must be a real calendar
date spelled YYYY-MM-DD, otherwise preparation.date at recipient.plannedStartDate.
Supplied team must be in context.Teams, otherwise preparation.unknown-team.
Supplied manager, onboardingOwner, and buddy must be in context.Contacts,
otherwise preparation.unknown-contact. Supplied workLocation is exactly remote,
on-site, or hybrid, otherwise preparation.location. Use each field's full path.

Name, companyEmail, and timeZone are text, not fetched or inferred. Email delivery,
provider validation, IANA time zone validation, and invitation readiness are not
this component's responsibilities. A missing companyEmail is valid.

## Arrangements

The list may be absent or empty and preserves input order. Each id matches
`[a-z][a-z0-9-]*`; otherwise preparation.arrangement-id. For each occurrence
after the first of a valid id, emit preparation.duplicate-id, even when all
other fields agree. Different ids may name the same section. A section must
name a context.Sections entry, otherwise preparation.unknown-section.

Kinds: support, equipment, office, remote, access, meeting, reading, pairing,
ticket, pr-review, pr-post, pr-merge, deploy-shadow, deploy-lead. Anything else
produces preparation.arrangement-kind. Statuses: pending, requested, arranged,
deferred. Anything else, including an empty status or verified, produces
preparation.arrangement-status. Do not infer a status from filled fields.

A supplied owner or helper must be in context.Contacts, otherwise
preparation.unknown-contact at its field or contacts index. Repeated valid
helper values are allowed. An empty owner is allowed except when arranged;
an empty helper id is supplied-but-unknown, not a valid blank list.

For status arranged, an empty owner produces preparation.incomplete-arrangement.
An empty contacts list also produces that code for pairing, ticket, pr-review,
pr-post, pr-merge, deploy-shadow, and deploy-lead. An empty reference produces
that code for every known kind except support. Unknown kinds do not produce
kind-dependent completeness errors. Whitespace-only owner, contacts, and reference
are not usable content; do not trim them into stored values. Unknown supplied
owner/contact values are diagnosed as unknown, not also as incomplete.

A nonempty targetDate must be a real YYYY-MM-DD date, otherwise preparation.date.
No rule compares a target with today, plannedStartDate, or another target. Dates do not
change status, outcomes, or validity merely because time has passed.

## Properties and exclusions

Repeated calls give equal diagnostics, including messages, and leave every input
unchanged. Diagnostics must not alias input data. No default is accepted and no
arranged state becomes a newcomer outcome. Those write boundaries are also
enforced by the signature: it has no record body or writable output plan.

This typed boundary does not parse YAML, reject serialized duplicate keys,
preserve unknown YAML fields, scan secrets, compose template values, report
all preparation gaps, check invitation readiness, reconcile stale imports, or
perform filesystem writes. Those need separate tests; passing this suite does
not establish those workflows. See the preparation format and the
[invitation flow](../../../docs/architecture/invitation-flow.md) for later increments.

## Wrong implementations tried

| Defect | Observable safeguard |
| -- | -- |
| Accept every preparation | Invalid-input fixtures require diagnostic codes and paths |
| Ignore identity or duplicate arrangement ids | Wrong-account, wrong-plan, record-identity, and repeated-id fixtures require rejection |
| Require a company email or an existing record | Partial-preparation controls remain valid |
| Treat targets as completion deadlines | Calendar-valid targets remain valid independently of the clock |
| Mutate input, echo personal values, or reverse diagnostics | Input-preservation, diagnostic-privacy, and ordered-problem assertions |
| Replace identity equality with lexical ordering | `reject-older-version`, `reject-kind-that-sorts-before-preparation`, and the lexically earlier plan and record fixtures |
| Apply arranged completeness checks to an empty status | `empty-status-does-not-claim-arranged-completeness` |
| Treat an empty contact id as a usable helper | `empty-helper-remains-invalid-even-in-contact-projection` |
| Exempt ticket work from its reference or demand a helper for general support | `arranged-ticket-needs-reference` and `arranged-support-needs-no-helper-or-reference` |
