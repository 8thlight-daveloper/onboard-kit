---
category: behavior
status: accepted
approval: agent
affects: [command/invite, preparation/reviewer]
builds-on: ['0055', '0140', '0150', '0160', '0240', '0280']
---

# 0300. Explicit invitation export

Builds on [0055](0055-variable-audience.md),
[0140](0140-person-specific-preparation.md),
[0150](0150-onboarding-library-and-preparation-flow.md),
[0160](0160-personal-data-and-preparation-approval.md),
[0240](0240-invitation-archive-envelope.md), and
[0280](0280-first-steps-help-and-sources.md).

## Context

The archive codecs and first-steps preview exist, but neither composes a usable
export. An onboarder needs to package their original plan and one person's
preparation without changing those files, manufacturing personal answers, or
claiming that the recipient has started. Export is a bounded increment toward
the guided invitation and confirmed startup flows, not their replacement.

Directory metadata does not prove that a newcomer can reach a helper before
company access exists. A company address or Slack handle can depend on the
access the invitation is intended to obtain. Do not infer availability, probe
the network, or collect personal delivery addresses to disguise that limitation.

## Decision

### Compose one explicit export

Provide a noninteractive `invite export` operation with explicit plan file,
preparation file, account, and new ZIP destination. Optional typed contact/role
selectors choose general help. Require an explicit pre-access help contact
designation for this export. No inferred account, default date, overwrite flag,
existing-record input, automatic delivery, network fetch, import or first use.

Read bounded local regular files through the real plan/preparation decoders.
Require a nonempty plan identity, valid exact account, matching preparation
identity, valid results-path layout, and no errors from current plan and
preparation validation. Validate complete declared relationships before activity
filtering. Keep the plan validator's documented authoring-coverage limits visible;
calling it is not proof of every possible schema or condition check.

Use the real question finder's onboarder blockers for handover. Required
onboardee questions alone do not block. An unaccepted default is not an answer.
Deliberate deferrals retain the finder's existing semantics and visible gaps;
do not silently require every required answer to be settled.

Missing company email, planned date, manager, buddy, time zone and optional
arrangements do not fail solely because they are absent. Supplied invalid values
still fail existing validation; an explicitly arranged item must satisfy its
existing completeness rules. Arrangement status never becomes verification.

### Require an explicitly designated pre-access help route

Preparation must name a known onboarding owner with a nonblank display name.
The onboarder must explicitly designate a known, named business contact with a
nonblank route the guide can display as usable by this recipient before company
access. `--preaccess-help-contact ID` supplies that designation and includes the
contact in the guide's general help. The owner and helper may be the same person.
Other selected contacts and roles remain optional; a vacant role alone does not
satisfy this requirement. SlackID alone is not a displayable route.

This is an organizational assertion for one export, not machine verification,
a new stored field, a personal answer, or a claim that access works. State that
availability has not been tested. Do not silently supply the assertion from a
default or a filled field. The first-steps preview still permits missing-source
wording; these delivery-oriented gates belong only to export.

### Preserve separate payloads and exact hashing

Encode the untouched original plan and validated preparation with their real
encoders. Set the seed's plan reference hash to lowercase hexadecimal SHA-256
of exactly the canonical plan bytes embedded in this archive, including the
final newline. Copy plan ID and URL exactly. The hash is not authentication.

Create a fresh schema-version-1 record with only account and that plan reference.
Never copy plan/preparation answers, outcomes, unknown data, legacy start date
or firstOpenedAt into it. A pre-seeded organization answer stays in the plan.

Obtain the exact real first-steps preview from the complete original plan,
preparation and explicit help selection. Preserve its warnings and truthful
staged wording. Do not inject preparation into answers or infer bindings from
question names or types; the preparation format explicitly reserves those
bindings for a separate change. Keep any resulting missing information visible.

Pass the four real payloads to the invitation encoder, retaining its fixed
archive names and aggregate size limits. Custom result paths affect displayed
destinations, never ZIP entry routing. No dropped payload, partial archive,
truncation, compression workaround or second renderer.

### Publish only a new complete archive

The explicit output parent must already exist and be a stable nonsymlink
directory. Refuse existing output objects, including dangling symlinks, and
unsafe path components. Do not create a directory tree or replace an input.

Prepare all bytes before publication. Write a private exclusive temporary file
inside the validated parent, check writes and close, then publish with an atomic
no-replace operation relative to the held parent. A replacement-capable rename
does not satisfy this rule. Clean owned temporary files on ordinary failures.
Two concurrent exports to one absent destination must produce one complete
archive and one refusal, never replacement or mixed bytes. Do not claim crash
durability or universal race resistance without a separate tested guarantee.

Success means written, not sent, opened, imported, identity confirmed or started.
The onboarder delivers the ZIP through an approved external channel. Preserve
all input bytes and unrelated local records/state. Read no credentials, execute
no checks, run no Git command and contact no service.

## Rationale and alternatives

P3 and P5 of [0100](0100-behavior-judgment.md) favor an actionable help route
before access exists. P2 and P4 preserve explicit ownership and local changes;
P6 keeps optional planning dates from becoming a gate; P7 avoids collecting
private delivery information or repeating personal values in diagnostics.

A draft-only archive could allow missing help with a visible warning, but would
not establish the pre-access support required for this delivery-oriented export.
Requiring an explicit contact designation is reversible and adds no saved format.
It cannot establish real reachability; the onboarder remains responsible for
that fact. Requiring recipient company email would block invitations intended
to obtain that email. Guessing preparation-to-question bindings would violate
the current ownership and original-plan preview contracts.

## Verification and remaining boundaries

The command driver reads explicit local plan/preparation and selected output
filesystem metadata. It uses actual validation, finder, encoders, locator,
first-steps assembler and invitation encoder; writes only a new explicit ZIP
and owned temporary file; changes no plan, preparation, record or local index.

Independent tests must inspect ZIP metadata and literal payload/hash oracles
without using the product decoder to judge the writer. Cover original versus
raw-source hashing, identity-only seed, explicit help, defaults/deferrals,
onboardee questions, optional missing preparation, inactive structural errors,
size boundaries, no-replace races, symlinks, failure cleanup and input preservation.
Whole-binary scenarios must separately verify command dispatch and exact artifact
operation. Component success alone does not establish those command boundaries.

Import conflicts, recoverable multifile installation, confirmed identity,
first-use recording, prepared answer bindings and guided selection remain
separate work. Full embedded catalog availability is delivered separately under
[0320](0320-embedded-catalog-availability.md). Do not advertise remaining flows as delivered by
this export increment.
