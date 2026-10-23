# 0330. Read-only preparation review

Status: accepted by agent judgment; contract and implementation pending

Builds on [0140](0140-person-specific-preparation.md),
[0150](0150-onboarding-library-and-preparation-flow.md),
[0160](0160-personal-data-and-preparation-approval.md),
[0180](0180-first-onboarding-use.md),
[0280](0280-first-steps-help-and-sources.md), and
[0300](0300-explicit-invitation-export.md).

## Context

An onboarder can edit preparation and explicitly export an invitation. A guided
flow should first let them inspect a proposed preparation without saving it or
creating an archive. Calling an existing editing command as a preview would
save changes before review. Calling export would publish an archive. Neither
is a read-only review interface.

Preparation validation, required onboarder answers and first-steps rendering
answer different questions. A valid preparation can lack an onboarding owner.
A guide can honestly describe missing help while export requires an explicitly
designated usable pre-access contact. One ready flag would hide those limits.

## Decision

Add a pure preparation reviewer before the guided editing driver. It receives
an original plan, a proposed preparation, an explicit account and ordered typed
help references. The caller supplies an authoring-validated, unambiguous plan.
No account or help selection is inferred from a file, directory order, login,
email, unknown field, default or previous call.

Return four separate results: preparation problems, onboarder blockers,
first-steps text and first-steps problems. There is no ready flag, changed
preparation, record, archive or save approval in the result.

1. Validate preparation through the actual preparation validator. Derive its
   context from the original plan's identity and all declared contact, team and
   section IDs, plus the explicit account. Supply no record identity. On any
   preparation problem, return all those problems unchanged and leave the
   other three results empty. Do not manufacture a partial guide.
2. For valid preparation, obtain the actual question finder's onboarder blockers
   from the unchanged original plan. Required newcomer questions do not become
   onboarder blockers. Defaults remain suggestions; explicit deferrals retain
   their existing meaning. Preparation team and assignments are not answers.
3. Independently obtain the actual first-steps assembler's text and problems
   from the same original plan, preparation and explicit help context. Preserve
   its refusal, warning order and exact guide bytes. Keep blockers even when
   the guide refuses, and keep a permitted guide even when blockers exist.

Do not merge these result fields or reinterpret a warning as preparation
invalidity. Missing optional information remains governed by the existing
validator and renderer. A successful preview does not establish owner/help
readiness, pre-access reachability, available filesystem destinations, archive
size, identity confirmation or permission to send an invitation. The existing
export command remains the authority for its separate explicit operation.

### Ownership and dependencies

Preparation reviewer is at dependency position 6. It may depend on its own
contract, model, actual preparation validator at 1, question finder at 3 and
first-steps assembler at 5. It must not depend on drivers, editors, encoders,
record components, filesystem loading or a replacement renderer. Context and
result projections belong in the contract, not the saved domain model.

Reads: explicit account/help context; original plan fields needed by actual
validation-context projection, question finding and first-steps rendering; and
preparation fields read by validation and the permitted first-steps projection.
Writes: no input field, file or outside state. Return owned result containers
and diagnostics. No record read, clock, environment, network, command execution,
saved session state or default acceptance. Repeated and concurrent calls must
not influence each other.

### Keep later work separate

This leaf neither stages edits nor defines a guided command protocol. A later
driver must specify explicit keep/change/clear choices, cancellation, reviewed
saving and truthful postcommit errors before implementation. Existing editor
operations replace whole values; a driver must not clear omitted fields by
passing a sparse recipient. Calling several existing editing commands is not
one atomic reviewed save.

Do not add a discipline field, persistent approval marker, date-confirmation
bit, activity-selection list or preparation-to-answer binding here. Existing
formats do not represent those rules. Do not collect a personal delivery
address or an actual-start date. Packaging and preview record no first use.
Selection, repeated resource identities, preparation relationships, confirmed
import and record-backed walking remain separate increments.

## Rationale and alternatives

P1/P3 of [0100](0100-behavior-judgment.md) favor useful preparation and visible
next actions. P2/P4 preserve explicit ownership and avoid writes during review.
P5 keeps pre-access help useful without pretending it proves access. P6/P7/P8
keep planning targets neutral and avoid speculative personal fields or redundant
questions. This local read-only composition adds no stored-format, privacy,
hash-compatibility or ownership change.

Keeping only explicit editing/export commands is safe but leaves no reusable
review boundary for the guided flow. Building the complete interaction and
publisher first would mix preview with commit behavior. Copying export gates
into a second readiness implementation would invite divergence. The four-field
review is smaller and makes its limits explicit.

## Verification and delivery

Draft an independent four-part contract before building the component. Pin
literal guide bytes from the existing first-steps normative goldens, exact
blocker IDs from declared questions and preparation diagnostic expectations
from their contract. Do not call the implementation to compute expectations.

Include valid review, honest missing optional sources, unaccepted defaults,
explicit deferrals, audience separation, wrong identity, invalid arrangements,
ordered help, guide refusal beside retained blockers and original-plan team
selection despite a different preparation team. Verify immutable inputs, owned
results, deterministic shared calls and concurrent use. A null stub must fail
every meaningful output case before a separate author builds the component.

These are planned checks, not completed verification. This decision specifies
the bounded composition; it does not claim a guided UI, reviewed save, new
export capability or first-use recording has been delivered.
