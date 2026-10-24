# Invitation reviewer contract

Status: agent-authored contract against accepted agent directions0380/0390;
implementation pending. Public API, prose and independent tests are not an
implementation or human approval.

## Interface, effects and dependencies

New has no arguments and constructs actual permitted dependencies. Inspect,
Review and Begin have the signatures in api.go. Position5 permits model, its own
contract and actual invitation decoder, plan decoder/authoringvalidator/validator,
preparation decoder/validator/locator, record decoder/reconciler/joiner/encoder
and section walker. No preparation reviewer, firststeps assembler, driver,
substitute semantic algorithm or caller-injected validator exists.

Read only supplied bytes, filesystem-presence facts, explicit content choice
and supplied event. Return owned values; change no argument. No filesystem,
configuration, environment, network, process, terminal, clock, randomness,
consent, journal, publication or recovery operation. A returned Publication is
not evidence that anybody began or wrote anything. The starter establishes
physical truth, user consent and safe source/path guards under0360.

Incoming/Proposal are plain constructible data, not authenticity tokens or proof
of prior calls. Carry exact bounded sources, derive all content policy using
actual components and check cached projections. No Valid/authenticated flag,
hidden provenance state, callback or hash-only approval check is permitted.

## Diagnostics and output states

A Diagnostic wraps a fixed Stage, fixed safe Display and unchanged model.Problem.
Preserve every actual dependency Code, Severity, Message, Question, Section and
Detail, including allowed extra detail keys. Copy storage; never inject phase
keys into borrowed detail maps. Diagnostic wrappers are the provenance boundary.
Actual dependency fields are diagnostic DATA, not a terminal-display safety
promise: old codecs may include untrusted found text, arbitrary unknown-key
paths or parser text. Display carries the fixed safe repair meaning below;
the starter applies bounded literal-safe terminal framing and never blindly
prints dependency Message/Detail/Question. Only own diagnostics have the
no-value-echo promise. Display is not produced by sanitizing a raw error string.
The normative [admission rules](admission.md) define exact resource contexts,
weights, caps, placement and fixed diagnostics. They add no estimator API.
Within an actual phase preserve dependency problem order. Never globally sort
or deduplicate different phases. Own diagnostics have severity error except the
explicit previous-resource comparison warning; they have empty
Question/Section, exact fixed Message from the table, and Detail exactly
`{"path":PATH}`. They never repeat rejected values, notes, timestamps or secrets.
For an own diagnostic Display is exactly its fixed Message. Dependency Display
is exactly the Stage table, independently of the raw problem's code/severity/text:

| Dependency Stage | Fixed Display |
| -- | -- |
| inspect.envelope | Ask your onboarding owner for a readable invitation archive. |
| inspect.plan-decode | Ask your onboarding owner for a readable invitation plan. |
| inspect.plan-authoring | Review the invitation plan declarations with its owner. |
| inspect.plan-validation | Review invitation plan integrity and organizational answers. |
| inspect.preparation-decode | Ask your onboarding owner for readable invitation preparation. |
| inspect.record-decode | Ask your onboarding owner for a readable invitation record seed. |
| inspect.locator | Ask your onboarding owner for valid record destinations. |
| inspect.preparation-validation | Review the preparation against this invitation. |
| previous.plan-decode | The previous plan copy cannot be read for comparison. |
| previous.plan-authoring | Review the previous plan declarations before using them for comparison. |
| previous.plan-validation | Review the previous plan integrity before using it for comparison. |
| review.preparation-decode | Keep local preparation intact and request help reading it. |
| review.preparation-validation | Ask your onboarding owner to repair unavailable local support. |
| review.record-decode | Keep the local record intact and request help reading it. |
| review.reconciliation | Keep the local record intact and resolve its plan binding. |
| review.effective-validation | Review stored personal answers against the incoming plan. |
| review.walker | Resolve the session's activity or history problem before opening. |
| review.record-encode | Keep the local record intact and resolve its unwritable content. |
| begin.record-encode | Resolve unwritable record content before beginning. |

On a fatal phase return its diagnostics only, with zero operation result; earlier
successful-phase warnings are not returned as a partial result. A dependency
phase is fatal if any problem has severity error; include that phase's whole
problem list, even if it also has warnings. Encoder problems are all fatal under
its real contract. No other phase runs after fatality. Successful warnings are
kept in the explicitly named result channels. An empty diagnostic list and nil
are equivalent.

Review has exactly three outcomes: fatal means zero ReviewResult plus diagnostics;
choice-needed means PlanChoiceNeeded true, nil Proposal and no diagnostics;
complete means false, nonnil Proposal and no diagnostics. Choice-needed is not
a provisional publication. Begin never accepts an incomplete reconstruction.

| Own Code | Fixed Message |
| -- | -- |
| invitation.plan-id | Supply an invitation plan identity. |
| invitation.seed | Use an identity-only seed matching the received invitation. |
| invitation.first-steps | Supply nonempty UTF-8 invitation guidance. |
| invitation.incoming | Use an incoming projection consistent with its archive. |
| invitation.local | Supply consistent local presence observations. |
| invitation.source-size | Keep each local source within the supported byte limit. |
| invitation.plan-choice | Choose undecided or use-incoming explicitly. |
| invitation.known-record-missing | Restore the registered record instead of creating another seed. |
| invitation.record-account | Use the record for this invitation account. |
| invitation.record-plan | Use this plan's record or explicitly bind an unbound record. |
| invitation.record-owned-answer | Keep onboarder-owned answers out of the personal record. |
| invitation.outcome-kind | Use verified, failed or deferred recorded outcomes. |
| invitation.history-time | Preserve a real UTC time for interpreted history. |
| invitation.proposal | Use the complete proposal derived from its exact source evidence. |
| invitation.event | Supply one supported begin event only for a new record. |
| invitation.output-size | Keep the encoded record within the supported byte limit. |
| invitation.resource | Use a smaller onboarding plan or fewer repeated values. |
| invitation.previous-resource | Previous plan comparison is unavailable within the resource limit. |

Exact Stage names and path locations are specified by the phases below; these
are finite literal strings, not dynamic account/plan/section/path projections.

## Inspect

Inspect performs these stages in order, stopping as described above:

1. `inspect.envelope`: actual invitation decoder on original archive bytes.
   Its2MiB raw and1MiB aggregate payload limits apply unchanged. Returned role
   payloads remain exact, including noncanonical but accepted archive framing.
2. `inspect.plan-decode`: actual plan decoder on Files.Plan.
3. `inspect.plan-admission`: full admission with incoming Plan, zero Record,
   nil Preparation and raw archive length, before expanding plan dependencies.
4. `inspect.plan-authoring`: actual authoring validator on the complete plan.
5. `inspect.plan-validation`: actual plan validator on that unchanged plan.
6. `inspect.plan-identity`: empty Plan.ID gives invitation.plan-id at `plan.id`.
7. `inspect.preparation-decode`: actual preparation decoder.
8. `inspect.record-decode`: actual record decoder on the identity seed bytes.
9. `inspect.payload-admission`: full incoming Plan/seed/Preparation admission
   using raw archive length, before semantic seed inspection.
10. `inspect.seed`: the strict profile below, with own seed diagnostics.
11. `inspect.locator`: actual locator on Plan.ResultsPath and envelope account.
12. `inspect.preparation-validation`: actual preparation validator using the
    envelope account, validated plan ID, every declared contact/team/section ID
    including inactive sections, and nil Record context. No stale allowance.
13. `inspect.first-steps`: zero bytes or invalid UTF8 gives invitation.first-steps
    at `firstSteps`. Preserve all other whitespace/control bytes as document data.

Actual decoder/authoring/plan warnings from successful phases enter Incoming.Warnings
in phase order. No warning alone refuses. Successful Incoming holds an owned copy
of exact Archive, all actual decoded InvitationFiles payload bytes, PlanID,
Reference and locator Paths. Reference.ID is Plan.ID, Hash is lowercase hexadecimal
SHA256 of exact received Files.Plan including its final LF if any, and URL is
Plan.URL exactly. Do not re-encode, normalize or fetch the plan for this hash.

Seed defects are independently checked in this order, one invitation.seed per
bad field: Account at `record.account`; Plan.ID at `record.plan.id`; Hash at
`record.plan.hash`; URL at `record.plan.url`; nonempty StartDate at `record.startDate`;
nonempty FirstOpenedAt at `record.firstOpenedAt`; any answer entry in Answers.Plan
at `record.answers.plan`; any answer entry in any Answers.Sections inner map at
`record.answers.sections`; any outcome at `record.outcomes`; any unknown root
entry at `record.unknownFields`. Account must equal envelope account; ID/Hash/URL
must equal the derived Reference exactly. Actual record decoder supplies schema1/
kind/known-field representation guards. Empty maps/lists, including named empty
section answer maps, carry no answer entries and are permitted. A key holding
ValueKindNone is still an entry and refuses if representable, but the actual
record decoder may already reject raw null/empty answer syntax. Seed-only
fixtures must use decoder-valid Deferred/false/text answer entries; raw-null
fixtures judge the codec phase rather than pretend to reach this guard.
No data is erased to create a seed.

## Review evidence and physical observations

First call actual Inspect on incoming.Archive. Its fatal errors retain their
original Inspect Stage. Then require supplied Incoming to equal the newly derived
Incoming under the equality profile below; otherwise invitation.incoming with
Stage `review.incoming` at `incoming`. Never trust cached payloads/paths/warnings.

Stage `review.local` checks plan/preparation/record Body entries in that fixed
order. Present false with nonempty Bytes gives invitation.local at `local.ROLE`.
Any Bytes longer than1048576 gives invitation.source-size at `local.ROLE`; all
such independent defects are returned in role order, presence before size per
role. Present empty bytes are not absent and later reach the actual decoder.
KnownRecordMissing true with Record.Present true is inconsistent and gives
invitation.local at `local.knownRecordMissing` after body defects.

An unsupported PlanChoice gives invitation.plan-choice at `choice` in Stage
`review.choice`, after local consistency succeeds. A true KnownRecordMissing
with absent record gives invitation.known-record-missing at `local.record` in
Stage `review.known-record`, before requesting a use-incoming decision, not before
the enum guard. The order is incoming consistency, local body/presence checks,
PlanChoice enum validation, known-missing guard, then choice-needed branching.
Thus unsupported choice wins over an otherwise consistent known-missing record.
The bool is a starter-supplied
filesystem fact, not content-validity/ownership/first-use authority. The core
cannot prove registry provenance or detect inode/path aliases from these values.

When local.Plan is present, Bytes differ exactly from received Files.Plan and
choice is undecided, return only the choice-needed result. No later local
decoding/proposal/session work runs. Use-incoming is permitted even when the
local plan is absent or identical, but does not force an unnecessary replacement.

## Previous comparison context

Absent local plan sets Comparison.Reason `absent`, Available false, no problems.
After successful actual previous decode and before either previous validator,
`previous.plan-admission` applies full admission to previous Plan/zero Record/
nil Preparation and that exact previous source length. Failure is comparison
unavailable with Reason `resource` and the exact own warning in admission.md;
retain successful preceding decoder warnings and do not run later validators.
For present plan, perform actual decode, authoring and plan validation with
Stages `previous.plan-decode`, `previous.plan-authoring`, `previous.plan-validation`.
Errors make comparison unavailable, with Reason `decode`, `authoring` or
`validation`, and that failed phase's complete diagnostics in Comparison.Problems.
These are not terminal opening errors. Retain successful preceding warnings in
Comparison.Problems too, in phase order; they describe the previous input only.

After successful validation, empty ID gives Reason `empty-id`; a different ID
gives `different-id`; neither invents a dependency diagnostic. Both are comparison
unavailable, not terminal refusal after accepted incoming replacement. A matching
valid previous plan with no variables and no sections gives Reason `empty-context`
and Available false under the actual reconciler's zero-context rule. Otherwise
Reason `available`, Available true. Supply only the validated available previous
plan to actual reconciler, zero Plan otherwise. Do not forge comparison lists or
claim unavailable means unchanged. The exact original local bytes remain evidence.

## Chosen preparation and support

If preparation is absent, choose validated incoming preparation and preserve its
received bytes for install. Otherwise Stage `review.preparation-decode` uses the
actual decoder. Immediately then `review.preparation-admission` applies cheap
shared-forest guards to incoming Plan/zero Record/selected Preparation, before
the actual validator. Stage `review.preparation-validation` uses the same trusted
incoming context as Inspect, still with nil Record context; record is checked
separately so an explicitly unbound record is not rejected by a prep context.

Only retained preparation permits actual preparation.unknown-team/contact/section
at the0380 exact field/index shapes with a held nonempty `[a-z0-9][a-z0-9-]*`
reference. Section references also exclude plan/yes/no/y/n/true/false/on/off/null.
Inspect each real indexed field, never arbitrary Detail.path prefixes or code
alone. Every problem must be in this allowance; one other problem returns the
entire actual validation-phase diagnostics as fatal. No severity is relabeled.

Allowed stale problems enter Review.PreparationProblems with their original
Stage. Set Support.Available false and leave Recipient zero/Arrangements nil.
Do not pass invalid preparation to valid-input components. Otherwise Available
true, Recipient and every Arrangement known field copied exactly in held order,
including helper repeats; no UnknownFields enters Support. Never convert a
preparation fact into an answer, activity outcome, completion or readiness gate.
InvitationFirstSteps remains original Files.FirstSteps even with stale support.

## Record guards and actual composition

If local record is absent, use the validated seed as the new proposed record.
Otherwise `review.record-decode` calls actual decoder. Immediately after decode,
`review.record-admission` applies cheap shared-forest guards to incoming Plan/
decoded Record/selected Preparation, before any identity/ownership/history
loop, reconciliation or join. In `review.record-identity`,
check account then plan ID, reporting invitation.record-account at `record.account`
and invitation.record-plan at `record.plan.id` when applicable. Empty ID is
permitted and marks the derived bind action; different nonempty ID refuses.
Collect both independent identity defects, in account then plan order, before
refusing this phase; an account defect does not suppress a plan defect.

`review.record-ownership` checks every held answer key against incoming declarations
before reconciliation, even inactive/default-audience ones. For each key currently
declared onboarder, give invitation.record-owned-answer. Paths are `answers.plan[N]`
then `answers.sections[N].values[M]`, ranks among exact stored map keys sorted
bytewise; keys themselves do not enter own diagnostic paths/metadata. No-kind,
empty and deferred values still occupy keys. Undeclared hidden keys are preserved.
Do not delete/filter keys or use a different plan's old audience.

`review.record-history` first checks nonempty FirstOpenedAt, then outcomes in
input order, Kind before On per outcome. Invalid first use/time gives
invitation.history-time at `record.firstOpenedAt` or `record.outcomes[I].on`;
unknown Kind gives invitation.outcome-kind at `record.outcomes[I].kind`.
Return all these independent problems, with no echo of the held text. Historical
UTC is four ASCII year digits0001–9999, valid Gregorian YYYY-MM-DD, uppercase T,
hours00–23/minutes00–59/seconds00–59, optional period followed by one or more
ASCII fractional digits, then exactly Z or +00:00. No rounding, truncation,
lowercase, spaces, comma, leap second, -00:00 or nonzero zone. Every On must be
nonempty; an empty existing FirstOpenedAt remains unknown. StartDate and arbitrary
hash spelling are preserved and not interpreted. Source size bounds fractions.

`review.reconciliation`: actual reconciler with guarded original/seed record,
actual available previous or zero Plan, accepted incoming plan and derived
Reference. Preserve all six real reconciliation lists. No second answer-removal
algorithm is permitted. Then actual joiner on incoming plan and reconciled record.
`review.effective-admission` applies full admission to that actual joined Plan,
proposed Record and selected Preparation, using archive plus each present local
body length once. Only then call effective validator/walker. No rendering charge
uses a fabricated pre-join answer set; cheap source guards are not full scores.
`review.effective-validation`: actual plan validator on that effective plan.
Errors refuse intact, including pattern/secrets failures that reconciliation
does not remove. Warnings enter EffectivePlanWarnings; required missing answers
remain actual blockers. Incoming authoring remains valid because joining changes
answers only. Never silently repair additional answers to make a session usable.

`review.walker`: actual walker on the effective plan and proposed record. Its
duplicate outcomes are therefore checked only after its other preconditions.
Errors return no proposal; successful problems enter SessionProblems. Project
every field exactly: View.Sections to Session.Activities, including full resolved
Section, CurrentHash, WaitingOn, rendered Escalation; every View.Outcomes field
to Session.Outcomes; Questions and Blockers exactly in finder order. Comparison
strings are the actual unknown/same/different values. No bootstrap-only subset,
second scheduling/hash/rendering, history pruning or completion inference.

## Fixed role actions and byte freezing

Artifacts always have three entries ordered plan/preparation/record. Plan action
is install if absent, retain with exact local bytes if identical, replace with
exact received bytes otherwise after use-incoming. Preparation action is install
if absent, retain with exact local bytes otherwise. No prep replacement/merge.

For new record, action create, Bytes nil, NeedsFirstOpenedAt true, BindsUnboundRecord
false. The nil provisional bytes are not a publication payload. For existing
record, compare complete decoded original and actual reconciled model values
including nested unknown types and nil/empty shape. Equal means action retain
and exact local bytes, with NO encoder call/gate. Otherwise actual record encoder
in `review.record-encode`; bound result1048576 inclusive, invitation.output-size
at `record` in `review.record-size` if exceeded. Action is bind when original ID
empty, reconcile otherwise; BindsUnboundRecord reflects original empty ID even
if other reconciliation changes also occur. Existing NeedsFirstOpenedAt false.

The complete Proposal carries deep-owned exact Evidence and Review. Review
contains derived account/plan/paths, role actions/bytes, comparison/report, event/
binding facts, separate warnings/stale/session channels, actual session/support
and invitation guidance. No decoded Plan/Record claim is accepted from callers.

## Begin and equality

Begin repeats actual Inspect/Review from Proposal.Evidence. Fatal source diagnostics
retain their original Stage. Choice-needed or any inequality of the complete
re-derived Proposal and supplied Proposal gives invitation.proposal at `proposal`
in `begin.proposal`, before inspecting event. Consistently constructed plain values
are accepted without requiring prior calls. No source is implicitly replaced.

Equality compares bytes exactly; scalar/enums exactly; ordered sequences and map
keys/payloads exactly; nil pointer versus present value distinctly; all cached
diagnostic fields including actual messages and fixed Display; all actions, model fields, unknown
values and flags. Only nil versus empty zero-length byte slices and zero-length
top-level derived lists are equivalent: Incoming.Warnings; Review.Artifacts,
PreparationProblems/IncomingWarnings/EffectivePlanWarnings/SessionProblems;
Comparison.Problems; Reconciliation's six lists; Session's four outer lists.
Nested model/unknown list/map shapes remain exact. No unordered set comparison,
hash-only identity, trimming or ignored display field is allowed.

After consistency, new record requires a nonempty supplied event with the same
date/clock syntax as history but Z-only and at most nine fractional digits.
Existing record requires empty event. Any violation gives invitation.event at
`event` in `begin.event`. Never read or infer current time, backfill an existing
empty event, use legacy/planned date or treat a new record as existing because
its seed has empty answers. Re-derive the actual proposed new record from frozen
sources, set only FirstOpenedAt to accepted text, then actual record encoder in
`begin.record-encode` and inclusive1MiB bound with invitation.output-size at
`record` in `begin.record-size`. No partial Publication on error.

Successful Publication.OriginalReview is an owned exact snapshot of accepted
Review, with its provisional new-record bytes still nil. Publication.Artifacts
are the FINAL authoritative bytes/actions; only its new record bytes differ.
Publication.FirstOpenedAt is the inserted event for a new record, or the exact
held existing event including empty. OriginalReview and final Artifacts share
no mutable storage. The starter journals only final artifact/event values and
never asks Begin to re-merge after publication/interruption.

## Ownership and independent verification plan

Every operation preserves all arguments deeply, including presence and nil/
empty shapes. Every returned slice/map/pointer/byte has owned storage independent
of inputs, another result and later calls, including unknown data and the actual
walker's immutable-shared Question internals. Concurrent calls on one New instance
have no history or state effects. Error/choice/complete outputs obey the same
ownership rules. No comments, decoded model payload or personal answer is logged.

Independent fixtures package handwritten literal payloads under the actual
ZIP envelope specification, never derive expected semantic outputs from product
encoders/renderers. Pin exact raw-plan hash/LF effects, layered refusal stages,
warnings, choice-needed, foreign/invalid previous context, retained stale support,
record ownership/history/binding, actual reconciliation and complete walker
projection, no-op encoder refusal bypass, supplied-event matrix, forged cached
projections and deep ownership. Testdata/sources.md records fixture provenance.
The binding selects the independently authored implementation's New constructor
and imports no other implementation. Stub checks replace only this binding with
a test-only fixed nonzero diagnostic stub; assertions remain unchanged. The root
agent executes stub checks and actual dependency composition checks separately;
this draft makes no claim of verified safeguards or a passing implementation.

## Plausible wrong implementations

The tests below specify required safeguards. Execution, independent builds,
wrong-variant evidence and remaining hardening gaps belong in working reports,
not this contract.

| Wrong behavior | Independent observable control |
| -- | -- |
| Accept every archive, or refuse every input | TestStrictSeedProfile refuses ten concrete defects; TestLiteralNewSession requires complete actual session and literal publication bytes. |
| Trust cached Incoming identity, payload or destinations | TestIncomingCannotAuthorizeForgedProjection changes each projection while keeping archive evidence fixed. |
| Treat arbitrary local bytes as accepted replacement without choice | TestChoiceAndPhysicalObservationPriority requires choice-needed before previous decoding, then unavailable previous comparison after explicit acceptance. |
| Stop at first identity defect or pick the known-missing guard before enum validation | TestRecordIdentityCollectsBothDefects and TestChoiceAndPhysicalObservationPriority pin ordered independent failures and simultaneous-invalid precedence. |
| Rebuild every existing record or normalize unknown first-use time | TestRetainExactExistingBytesAndNoEvent retains comments/key order and rejects supplied new event; TestNoOpDoesNotEncodeExpandedUnknownString distinguishes retain from bounded changed serialization. |
| Reconcile or render a second shadow algorithm instead of composing actual components | TestLiteralNewSession pins raw plan hash, independently specified unresolved section hash, resolved marker, actual question/blocker and wrapped warning. |
| Accept cached UI/action flags, unsafe Display or forged session as authority | TestBeginReDerivesAllCachedProjection changes every projection family, checks rejection before invalid event, and accepts consistently copied values on a fresh instance. |
| Use a general timestamp parser that normalizes history or accepts any UTC zone for new events | TestHistoricalUTCAndNewEventProfilesDiffer preserves long fractions/+00:00 history but refuses these as new events and rejects malformed Gregorian/case/zone inputs. |
| Check source limits after choice or apply an exclusive byte boundary | TestSourceBoundsBeforeLaterGuards differentiates1048576 from1048577 for each body before the unsupported-choice guard. |
| Share caller/dependency/result backing arrays or retain mutable history between calls | TestDeepOwnershipAndReusedConcurrentInstance snapshots exact nested/nil shapes, mutates independent publication branches, and compares later/concurrent outputs. |
| Treat any stale preparation error as harmless, or turn support assignments into accepted personal data | TestStaleLocalSupportIsRetainedButUnavailable pins actual ordered stale diagnostics and malformed/unrelated refusal; TestValidSupportProjectsEveryKnownFieldOnly pins every known field/repeat and retains the required personal blocker. |
| Let the audience-blind joiner accept organizational record keys | TestCurrentOnboarderKeysRefuseBeforeJoin rejects default/explicit onboarder keys holding false, Deferred and empty text before reconciliation. |
| Reseed an unbound existing record or erase legacy fields | TestUnboundRecordUsesActualBindAndPreservesLegacyData pins handwritten binding bytes and no first-use insertion. |
| Drop hidden answers, silently fix pattern failures, or accept defaults as answers | TestChoicesReconciliationAndDeepQuestionOwnership preserves hidden scope while removing a bad current choice; TestPostJoinPatternFailureRemainsFatal rejects the unreconciled pattern error. |
| Truncate history, make definition mismatch invalidate verification, or choose a duplicate-outcome winner | TestCompleteOutcomeHistoryAndDuplicateRefusal pins full fields, unknown/different comparisons, held source order and actual duplicate refusal; TestHistoryGuardCollectsInHeldFieldOrder pins earlier own guard ordering. |
| Use an exclusive B limit or forget exact carried archive bytes | TestAdmissionExactPublicBoundaryAndPayloadRefusal admits independently derived B and refuses the same models with one inert archive byte at full payload admission. |
| Skip disabled/literal-dollar declarations or expand before early plan refusal | TestAdmissionEarlyPlanPrecedesSemanticValidation requires early resource errors for admitted-size hostile charges, even disabled or independently invalid semantic data. Output alone does not establish check timing; actual dependency placement also needs independent source review. |
| Deduplicate identical present local source bytes | TestAdmissionCountsIdenticalPresentLocalSource requires effective-only refusal after exact-B incoming succeeds. |
| Treat over-budget previous context as fatal, discard preimage or bypass choice | TestAdmissionPreviousResourceWarningAndChoiceOrder pins choice first, warning/reason resource and retained source evidence. |
| Omit preparation pointer/known-root depth, delay cheap guards or reset per-root counters | TestAdmissionLocalDepthGuardsBeforeSemanticErrors pins128/129 and semantic precedence; TestAdmissionSharedUnknownCounterAcrossRoots pins65536/65537 combined unknown occurrences. |
| Budget only rendered body text or truncate candidate lists | TestAdmissionDirectoryChoiceFanoutWithoutLargeRendering refuses512-contact fanout while preserving all200 choices for each of64 questions in its smaller positive control. |
