# 0160. Personal data and preparation approval

Status: accepted; enforcement and preparation implementation pending

Builds on [0100](0100-behavior-judgment.md),
[0140](0140-person-specific-preparation.md), and
[0150](0150-onboarding-library-and-preparation-flow.md).

## Context

The person approved R1's bounded preparation shape and R3's choice to keep
personal invitation-delivery addresses out of shared preparation by default.
They also asked for personal data to be treated as personally identifiable
information (PII) as a product principle. R2's start-date semantics are a separate
choice, not permission to overwrite dates in a newcomer's record.

[0170](0170-planned-and-actual-start-dates.md) records the subsequent approval of
separate planned and actual start dates; existing values still need an explicit
compatibility rule before any migration.

[0180](0180-first-onboarding-use.md) subsequently replaces the actual-start
question with a first-use timestamp while preserving the planned date.

## Decision

### Treat personal data as PII

Treat data about an identifiable person as PII for product design. This includes
names, personal and company email addresses, account identifiers, and linked
assignments, dates, answers, progress, and notes. A technical account id or a
company address is not automatically nonpersonal. This is a product rule, not
a claim about legal classification or legal compliance.

Collect and retain only what serves a stated onboarding purpose. Explain where
it is saved and who can read it. Authorship and file ownership are not privacy
controls; shared repository readership and the existing visibility notice apply.

Keep personal email accounts, personal phone numbers, home addresses, and similar
private contact details out of the normal onboarding flow and artifacts. An
exception needs an explicit purpose and a deliberate visibility and retention
decision. Do not add those fields just because a future integration might use them.

The invitation builder produces a ZIP; it need not collect a personal delivery
address. The onboarder can send the archive through an approved external channel.
This narrows 0150's earlier proposal to ask for a delivery address and R3's
transient-collection option: do not ask for personal contact details by default.

Company email and other needed business contact information may remain visible
in the existing shared files. They are still PII, not fields that the current
tool can promise to conceal from repository readers. Use a time zone and remote,
on-site, or hybrid arrangements where possible. A rough region may be recorded
when needed to coordinate work; do not collect a precise location or address
for that purpose or infer location from network data.

Do not echo personal input unnecessarily in diagnostics or logs. Intentional
display of a needed contact is different from copying their address into an
error. Examples and fixtures remain fictional. Credentials stay out of every
onboarding artifact.

This does not ban information needed by the existing shared-record workflow or
promise private storage. It introduces no tracking, transmission, automatic
deletion, encryption, or expanded readership. Future metrics need the same
purpose and visibility rules. Removing names alone does not establish that
a small team's data cannot identify someone.

### Record field-level encryption for a later version

Keep selective field encryption as a future option, not an implemented privacy
control or a requirement for this release. The
[encrypted-results option in 0085](0085-walking-sections-and-recording-outcomes.md)
already discusses notes; a future design may also protect selected business
contact and other personal fields. Record which fields stay readable and which
readers can decrypt each protected field.

Before adding it, decide recipient keys, recovery, changes in onboarder access,
offline operation, and how unresolved encrypted values appear in reports and
templates. Encryption must not require sending data to a hosted service.
It does not erase old plaintext from Git history or revoke copies already read.
Do not use possible future encryption to justify collecting private details now.

### Approve bounded preparation

Use explicit recipient details and named arrangements, not an unrestricted
overlay of plan or record fields. Each arrangement has a stable id, activity
reference, kind, owner, assigned contacts, work or resource reference, optional
target, status, and note. Pending, requested, arranged, and deferred describe
preparation, never newcomer outcomes. Arranged work requires the relevant
owner, support, and reference.

The detailed R1 proposal is approved for the domain-model and contract review
increment. Preserve separate accounts, plans, and owners, rejection of duplicate
identities, and existing newcomer work. Source the contract fixtures to this
decision. Implementation and verification remain separate under
[0025](0025-contracts-and-tests.md).

R1 does not settle R2's date semantics or authorize the test author to implement
the judged component. The existing startDate field must not be relabeled as
confirmed actual start merely because it contains a date.

## Verification and delivery

The validator must reject the wrong person, wrong plan, duplicate arrangement
ids, and unsupported statuses without leaking supplied personal values into
diagnostics. Pending arrangements and absent company email remain valid partial
preparation, not completed newcomer work.

Later invitation and driver scenarios must prove that personal contact details
are not requested by the normal flow or emitted in artifacts or logs, that
needed business contacts have clear shared visibility, and that preparation updates leave
newcomer answers and outcomes untouched. A validator-only suite cannot establish
those end-to-end boundaries. Free text can contain personal data; this decision
does not claim automatic recognition or removal of every such value. This ADR
makes no enforcement claim.
