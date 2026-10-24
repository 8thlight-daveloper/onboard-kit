# 0380. Review an invitation as a pure proposal

Status: accepted by agent judgment; contract draft available, resource admission and implementation pending

Builds on [0015](0015-functional-core-and-imperative-shell.md),
[0090](0090-plan-and-onboarding-record.md),
[0140](0140-person-specific-preparation.md),
[0230](0230-first-use-record-compatibility.md),
[0240](0240-invitation-archive-envelope.md),
[0360](0360-confirmed-invitation-start.md), and
[0370](0370-plan-authoring-validation.md).

## Context

Confirmed start needs decisions about seed identity, retained preparation,
record ownership, reconciliation, existing history and the effective onboarding
view. Those depend on content. Putting them inside filesystem publication would
violate 0015 and make recovery a second interpretation of onboarding policy.

Existing components provide important parts, not the complete proposal. Record
decoding is representational. The reconciler does not validate history or string
patterns. The joiner overlays every record answer, regardless of audience. The
walker requires an authoring-validated effective plan and known outcome kinds;
it neither joins answers nor establishes confirmed identity or first use.

## Decision

Add one pure invitation/reviewer component before the imperative invitation
starter. It inspects supplied bytes, proposes a supported opening and finalizes
one supplied begin event. It does not open files, obtain consent, discover a
root, read configuration or a clock, lock, journal, publish, recover a transaction
or call another driver. All transient types belong in its contract API. No saved
model, format, hash algorithm, preparation binding or old core contract changes.

### Three explicit pure operations

New takes no arguments and constructs actual dependencies. The API has:

```text
Inspect(archiveBytes) -> Incoming, problems
Review(Incoming, Local, PlanChoice) -> Review, problems
Begin(Proposal, optionalEvent) -> Publication, problems
```

Inspect consumes bounded ZIP bytes through the actual invitation decoder. Its
owned Incoming contains validated original content, exact payload bytes,
account, exact received-plan reference, locator paths and original invitation
guide. Review consumes this unmodified value plus selected local byte snapshots
and an explicit plan choice. It returns either a choice-needed result or an
owned complete Proposal with the actual session view. A choice-needed result
is not a publication proposal. Begin consumes an unmodified completed Proposal
and an explicitly supplied event, returning owned final publication bytes.

No operation retains earlier arguments or session state. Incoming and Proposal
are owned source-carrying plain data, not authentication, filesystem capabilities,
saved approval tokens or proofs that a prior call happened. They carry the exact
bounded archive and local source evidence and explicit choices needed to derive
their projections. No caller-supplied Valid or authenticated flag is trusted.
Review derives the incoming facts from the carried archive through actual
dependencies and rejects inconsistent claimed projections. Begin deterministically
re-derives inspection and the complete proposal from the same frozen bounded
evidence through actual dependencies, then compares the complete derived content,
role actions and proposal identity with what it received before finalizing an
event. A mismatched projection or action refuses without partial Publication.
The contract pins exact equality and diagnostic precedence; hashes alone do not
substitute for full source/proposal comparison.

This pure recomputation occurs before publication. It reads no current filesystem
state, clock or new input, accepts no replacement evidence implicitly and never
implements a second reconciliation algorithm. The starter separately binds the
reviewed source evidence to its snapshots and consent. Recovery consumes frozen
Publication bytes only, never re-runs inspection/review/reconciliation against
later files. Result ownership tests must establish that changing one returned
value cannot affect any input, other result or later call. Returned maps, slices,
pointers and bytes are owned at every depth, including decoded unknown data.

Local carries presence separately from body bytes for the local plan,
preparation and record; present-empty is not absent. It carries a trusted
known-record-missing fact from the starter's 0360 registry/path checks. It does
not contain caller assertions that content is valid, stale, owned or a new first
use. The shell establishes physical context; the reviewer establishes content
meaning. PlanChoice is exactly undecided or use-incoming; never infer newer.

The contract defines exact data types, codes, paths and phase ordering. Results
keep fatal refusals, actual warnings, allowed stale-preparation problems, real
reconciliation, comparison availability, walker problems and the walker view
separate. There is no ready, exportable, approved or activity-complete result.

### Inspect the received invitation

First use the actual invitation decoder's 2 MiB raw and 1 MiB aggregate payload
bounds. Decode the plan, preparation and seed through actual codecs. Run actual
plan authoring validation first; only without its errors run actual plan
validation. Warnings alone do not refuse. Require a nonempty original plan ID.
Retain exact raw plan bytes; do not regenerate them for identity verification.

The seed is schema 1 with exact envelope account and incoming plan ID. Its Hash
is lowercase hexadecimal SHA-256 of exactly the received plan payload bytes,
including any final newline, and its URL exactly equals the incoming Plan.URL.
It has empty StartDate and FirstOpenedAt, no answer entries in any scope, no
outcomes and no unknown root entries. Empty collection shapes carry no entries
and are allowed; no nonempty data is erased to make a seed. Record kind/version
and nested-key format guards come from the actual decoder. Routing account is
already syntax-checked by the actual envelope decoder and comparisons remain
exact, with no normalization or inferred identity.

Validate incoming preparation fully with the actual validator, using envelope
account, incoming plan ID and all declared contact/team/section IDs. Any incoming
preparation problem refuses; only retained local support has the stale exception.
Use actual locator results, never ZIP names, for proposed relative destinations.

FirstSteps must have at least one byte and valid UTF-8. Its bound is the actual
envelope aggregate limit. Preserve all whitespace and controls as document data;
do not infer semantic usefulness, help choices, permissions or readiness. Its
literal-safe terminal presentation belongs to the starter. It remains original
invitation-time guidance, not a fresh rendering of retained local preparation.

### Choose local content without recency guesses

Local input bodies are each bounded to 1 MiB before decoding. An absent local plan
uses the incoming bytes. Byte-identical local plan bytes are retained. Differing
local plan bytes require PlanChoice use-incoming before a Proposal exists.
There is no timestamp/hash recency rule.

An available local plan is decoded and receives the same two real validation
stages before it can serve as previous comparison context. A decoded nonempty
ID different from incoming does not supply previous comparison context, but does
not itself refuse an explicitly reviewed incoming-plan replacement under 0360. An
undecodable, authoring-invalid, plan-invalid, absent or empty-ID previous plan
likewise supplies no comparison context; differing bytes still need the explicit
incoming choice before replacement. Preserve those original bytes for the
starter's guard/recovery obligations. Comparison-unavailable is explicit, not
an empty claim that no changes occurred.

The actual reconciler treats a previous plan with no variables and no sections
as zero comparison context. Report that comparison as not computed even if
such a previous file was valid; do not invent different change lists. The
existing contracts' comparison rules stay unchanged.

Existing local preparation must decode and match incoming account/plan. Retain
its exact bounded bytes, including when different from the archive. Never merge
or replace it, even after accepting incoming plan bytes. Validate it against
the incoming declarations with the actual preparation validator and no record
identity projection; record binding is checked separately. All problems refuse
except the following retained-reference profile.

| Actual code | Allowed fixed path shapes | Exact held value requirement |
| -- | -- | -- |
| preparation.unknown-team | recipient.team | `[a-z0-9][a-z0-9-]*` |
| preparation.unknown-contact | recipient.manager, recipient.onboardingOwner, recipient.buddy; arrangements[I].owner; arrangements[I].contacts[J] | `[a-z0-9][a-z0-9-]*` |
| preparation.unknown-section | arrangements[I].section | `[a-z0-9][a-z0-9-]*`, excluding the section reserved IDs |

Section reserved IDs are exactly plan, yes, no, y, n, true, false, on, off, null.
Paths must identify the actual indicated held field and valid existing index;
there is no arbitrary path-prefix allowance. Any other problem, path shape or
malformed/empty reference refuses, even beside permitted problems. No trimming,
deduplication or fallback occurs. In particular, unknown-contact also diagnoses
empty helpers/whitespace and unknown-section diagnoses missing section values;
these are not accepted by code alone. Permitted unknown references mean current
support unavailable, not proven previously valid. Preserve their real ordered
diagnostics separately and do not pass invalid preparation into a valid-input
dependency. If local preparation is absent, install the validated incoming bytes.

### Guard and reconcile the selected record

A known registered record that is now missing refuses; it is not a new seed.
Only an unseen absent selected record uses the validated seed proposal. A present
record must decode, match the envelope account and have either incoming Plan.ID
or empty Plan.ID. An empty ID is an explicitly displayed bind action; a different
nonempty ID refuses. Do not infer a move or repair identity.

Before reconciliation/joining, inspect all held answer keys against the validated
incoming declarations, including inactive scopes. A currently declared onboarder
variable, including empty/default audience, must not have a record answer.
Refuse intact instead of deleting/filtering it. Hidden answers with no current
declaration remain held. This applies 0360 without changing the audience-blind
joiner or the reconciler's unchecked defensive behavior.

Existing history is validated without rewriting it:

- Every outcome Section is unique by exact text, including removed/inactive
  history. Obtain the actual walker's duplicate diagnostic after earlier history
  guards pass; never select a winner or silently remove duplicates.
- Kind is exactly verified, failed or deferred. No normalization or guessed kind.
- Every Outcome.On is a real UTC timestamp under the historical-read profile
  below. Nonempty FirstOpenedAt uses that profile too; empty existing FirstOpenedAt
  remains unknown. Legacy StartDate is preserved and not interpreted here.
- Accept held unsorted outcomes, arbitrary held hash spelling, empty hashes and
  removed-section history. Do not require every historical section to exist
  today, infer freshness or change kind/time/hash/note/differently text.

Historical-read profile: exactly four ASCII year digits for Gregorian year 0001
through 9999; valid YYYY-MM-DD; uppercase T; two-digit hours 00–23, minutes/seconds
00–59; optional full stop followed by one or more ASCII fractional digits; UTC
zone exactly Z or +00:00. No spaces, leap second, comma fraction, lowercase
separators, -00:00 or nonzero offset. Source-size bounds limit fraction length;
do not truncate, round or normalize held precision. Fractions beyond nine digits
remain valid historical text. This is an explicit read profile, not a change to
representational codecs or the outcome recorder's new-write rules.

Call the actual reconciler with guarded record, valid available previous plan
or supported zero context, accepted incoming Plan and the exact received-plan
reference. Preserve its full proposed Record and actual Reconciliation report.
Its supported removal/narrowing of invalid declared newcomer answers is the only
answer repair. Hidden answers, outcomes, legacy date, existing first use and
unknown data stay governed by that contract.

Call the actual joiner on accepted incoming Plan plus proposed Record. Preparation
does not become answers. Run actual plan validation on this effective Plan too:
the reconciler does not check string patterns or secrets. Errors refuse intact;
do not remove additional answers to manufacture success. Warnings remain visible.
Required unanswered questions are blockers, never an extra import/refusal gate.

Then call the actual walker with this effective Plan and proposed Record. Return
an exact fieldwise projection of its whole actual view and problems, including
waiting work, questions, blockers and held history. The reviewer contract imports
model only, not the walker contract: define its own transient session/section/
outcome/comparison projection types with every actual walker field represented.
This adds no shared saved-model type and no shadow scheduling or rendering.
Copy nested storage even where the walker's Question contract permits immutable
sharing with input. Do not filter to bootstrap, reconstruct instructions or infer
completion from preparation, dates or mismatching historical hashes.

### Freeze bytes and one begin event

Review produces explicit fixed-role retain/install/replace/reconcile/bind
actions. Local plan retention uses exact original bytes when identical; otherwise
use received incoming bytes after explicit choice. Local preparation retention
always uses its exact original bytes. Absent preparation uses received bytes.

For an existing record, compare the complete decoded original and actual
reconciled proposal, preserving nil/empty shapes in that comparison. If equal,
retain original exact bytes and mark record no-op without an encoder-validity
gate. Otherwise use the actual record encoder and enforce 1 MiB inclusive. This
does not waive earlier record/history/ownership validation. Freeze these bytes
before offering begin. Deterministic prepublication re-derivation must reproduce
those bytes and actions exactly; it cannot accept a different reconciliation or
encoding after consent.

A new record proposal has no FirstOpenedAt during review and explicitly requires
the accepted-begin event. Begin takes that completed Proposal and a supplied
nonempty event only for this case. After the source/proposal re-derivation and
full equality checks above, it inserts the event into a new Record value,
calls the actual record encoder and enforces 1 MiB inclusive. It preserves every
other reviewed value. New-event profile is the same valid date/clock rules as
historical reads, but zone exactly Z and fractional precision at most nine ASCII
digits. Preserve supplied accepted text; no clock, CLI date default or epoch
inference exists in this core. Existing-record Begin requires no event and returns
the already frozen bytes unchanged; a supplied event refuses rather than being
ignored or used to backfill history.

The starter samples its injected clock only after accepted begin and fresh
source/path rechecks when Proposal requires this new event. It journals the
resulting Publication/event before participant writes. Resume uses that frozen
publication and event, never another reconciliation or clock sample. Event time
means the accepted beginning under 0360, not proof every subsequent display
succeeded. Begin's pure result does not itself mean consent or publication.

All per-role output encodings/bounds must be established before publication
staging; encoding failure makes no partial Publication. An absent identity seed
cannot inherit an event from planned dates, the archive, legacy StartDate or a
previous attempt that left no recoverable proposal.

### Dependencies and effects

Invitation reviewer is position 5. It may depend on model, its own contract and
actual invitation decoder, plan decoder, authoring validator at 2, plan validator at 4,
preparation decoder/validator/locator, record decoder/reconciler/joiner/encoder,
and section walker at 4 at their earlier positions. Use standard pure hashing,
UTF-8, calendar and value-copy facilities as needed. Do not call another driver,
copy a core algorithm or introduce a generic replacement validator.

Do not depend on preparation reviewer at 6 or first steps assembler at 5. Their fresh
original-plan preview is not required by this minimum opening and cannot recover
the original invitation's help choices. Selected valid preparation can supply
a bounded known-field support projection; stale preparation supplies unavailable
support plus actual reasons. Unknown values never enter that display projection.

Reads: supplied archive/local bytes and explicit context/choice/event values,
through real semantic dependencies. Writes: no input field or outside state;
returns owned derived values, proposed bytes and diagnostics. No record mutation
occurs until the separate starter publishes approved bytes. No shared plan,
source ZIP, answers/outcomes, index, journal, filesystem or environment operation.

Invitation starter can be position 6, after this reviewer; existing binary wiring at 8
can remain after it and guide at 7. Its argument/interaction, privacy, roots/aliases,
ACLs, locks, source snapshots, journal/receipt bounds and recovery/reporting are
separate imperative contract work under 0360. This decision specifies none of
those filesystem algorithms and makes no group-atomicity or durability claim.

## Rationale and alternatives

One pure boundary reuses existing exact semantic operations and keeps content
policy out of publication. A section of driver-local predicates would conceal
policy inside I/O. Adding a general record validator, second guide renderer,
generic importer framework or several one-use proposal components would expand
the prerequisite without replacing the needed exact composition.

P2/P4 preserve owner work and explicit choices; P3 keeps missing comparisons,
stale support and refusals understandable; P5 keeps the actual offline onboarding
view; P6 prevents planning dates becoming gates; P7 limits diagnostics/display;
P8 favors one sufficient pure unit. Defaults remain suggestions under 0120.

## Verification and remaining work

Independently specify exact APIs, diagnostic phase/order/projection, bounds and
literal proposals before building. Pin actual dependency composition with
handwritten byte/model/view oracles, including strict seed data/hash checks,
all stale-reference code/path/value boundaries, malformed/long-precision UTC
history, misowned inactive answers, hidden work, no-op record bytes, unbound
binding, invalid previous comparison, postjoin pattern/secrets refusal and
actual waiting/blocker/history output. Include a valid incoming plan, wrong-ID
local plan and matching-ID existing record: explicit incoming acceptance permits
the replacement with previous comparison unavailable; the old plan bytes remain
the guarded preimage. A wrong-ID record still refuses. Demonstrate stub and wrong-variant
failures, then an independent implementation and differential checks.

No contract, implementation or passing evidence is supplied by this decision.
It does not resolve starter recovery formats or authorize changing old contracts.
The known reconciler/joiner prose inconsistency is bounded by 0360 ownership
refusal, not silently repaired. The authoring validator prerequisite must be
independently delivered before a supported reviewer/walker opening is claimed.
