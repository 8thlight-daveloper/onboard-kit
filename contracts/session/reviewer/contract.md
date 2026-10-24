# Session reviewer

## 1. Purpose, effects and authority

Pure position7 core for reviewing and finalizing ONE explicit newcomer-owned
record change in an already selected existing session. api.go defines the whole
model-only API. Source carries three independent raw bodies and explicit trusted
Account/PlanID; it does not carry paths, a selected account guess, filesystem
security, a session key or a permission flag. New() has no arguments and builds
the actual dependencies below. No fake ZIP, clock, filesystem, environment,
terminal, process, check/command execution, network or saved state.

Reads all three source bodies through actual codecs and all model fields needed
by actual validators/joiner/writer/walker/recorder. Returns derived owned values
and, only from Finalize, complete record bytes. It never writes plan,
preparation, registration, a first-use event, reference binding, reconciliation
result or history prune. Reference mismatch is a displayed byte-hash fact, not
automatic reconciliation or a live metadata seal. No driver is called.

This is an agent-selected proposed contract, not human approval or implemented
behavior. It composes0090/0140/0230/0270/0290 and the new explicit-intent writer;
it leaves generic splitter and existing cores untouched. Actual native identity,
pending-state proof, read/save consent, source freshness and atomic publication
belong to driver8/binary9, not a pure value's assertion of trust.

## 2. API, phases and exact behavior

### Common result and diagnostic rules

Any severity error in a fatal channel stops at that phase and returns zero
Session/Proposal/Publication. Collect a dependency's entire reached batch in its
own order; preserve every Problem field. Warnings never become errors. Keep
preceding reached warnings on a later refusal. A successful no-warning diagnostic
slice may be nil or empty. Own errors have Severity error, fixed Message below,
empty Question/Section, and only the specified Detail entries. Do not expose
arbitrary values in own diagnostic metadata. Resource phases and diagnostics are
fully fixed in admission.md, including predecode/intent/event bounds.

### Inspect(Source)

1. Predecode raw/identity admission. Then expected Account must match complete
   ASCII `[A-Za-z0-9][A-Za-z0-9._-]*` and PlanID must be nonempty; either invalid
   context gives one own context error before any codec. Identity is never copied
   from untrusted preparation/record to repair this context.
2. Actual plan decoder, Stage plan.decode; full plan.admission immediately
   after success. Actual plan authoring validator, Stage plan.authoring; then
   actual plan validator, Stage plan.validate. Do not call a successful-input
   semantic core after declaration/graph ambiguity errors.
3. Actual record decoder, Stage record.decode; cheap record.admission immediately
   after success. Actual preparation decoder, Stage preparation.decode; cheap
   preparation.admission immediately after success. Missing/empty bodies must
   produce the actual decoder refusal, never an invented empty model.
4. Exact identity phase, Stage identity: collect, in this order, one shared plan
   ID mismatch, one record account mismatch, one record plan ID mismatch.
   Empty held plan ID is a mismatch, not permission to bind. Compare byte-exact,
   case-sensitive nonempty trusted context, never order/basename/URL/hash.
5. Stage record.ownership: inspect every original record answer key currently
   declared onboarder, including default-audience, inactive, empty and deferred
   answers. Emit one own ownership error per key, ordered by plan-name bytes,
   then section-ID/name bytes. Use indexed paths among ALL held keys: plan[N],
   then sections[N].values[M] as specified below. Undeclared hidden keys stay.
   Prior actual plan validation prohibits ambiguous declaration winners.
6. Stage record.history: for each original outcome whose Kind is not exactly
   verified/failed/deferred emit one own kind error in original index order.
   Establish Walker's explicit enum precondition without revalidating unrelated
   On, FirstOpenedAt, StartDate or arbitrary hash text. No-kind is invalid here.
   Duplicate known outcomes are left to actual Walker's full diagnostic batch.
7. Full payload.admission before preparation validation or Join. Actual
   preparation validator, Stage preparation.validate, using trusted expected
   identity, actual decoded record identity, and all validated plan contact,
   team and section IDs. Apply only the exact stale-reference support allowance
   below. Every other error is fatal as a complete unchanged dependency batch.
8. Actual Joiner on original shared Plan and original Record; full
   effective.admission; actual plan validator Stage effective.validate; actual
   Walker Stage walk. Preserve real scope/audience/default semantics, global
   action/question order, unresolved definition hashes, support gaps and direct
   Requires evidence. Never filter waiting/verified work or choose an outcome
   winner before these calls. Missing required questions/blockers are information,
   not new fatal outcome gates; actual structural/value errors still refuse.

Successful Session fields:

- Account/PlanID equal trusted inputs. Reference equals complete held reference.
  PlanReferenceMatches is exact held Hash equality with lowercase SHA256 of raw
  Source.Plan; empty/legacy mismatching hash is false, not an error. Preserve
  Hash/URL in every proposed record; no metadata planSHA256 input or assumption.
- Activities and Outcomes are complete actual Walker projections in its own
  order, with all Section/Outcome fields, CurrentHash, WaitingOn, Escalation and
  Comparison strings exactly unknown/same/different. Current hashes are actual
  unresolved definitions, not source/template/rendered-text hashes.
- Questions and Blockers are actual full-plan both-audience Finder results via
  Walker, including current answers/default suggestions and ownership metadata.
  Deep-copy nested Question internals despite the dependency's weaker immutable
  sharing guarantee. Do not fill unaccepted defaults/examples into answers.
- Support is available only on valid preparation: every known recipient and
  arrangement field copied exactly, held ordering/helper repeats retained,
  UnknownFields excluded from terminal support. Nothing overlays answers.

#### Stale support allowance

Use only the public invitation reviewer's existing `Chosen preparation and
support` exact allowance, restated here to prevent a guessed code whitelist:
all actual errors must be preparation.unknown-team at recipient.team;
preparation.unknown-contact at recipient.manager/onboardingOwner/buddy or an
actual arrangements[I].owner/contacts[J]; or preparation.unknown-section at an
actual arrangements[I].section. Check real decoded field/index existence,
Detail containing exactly that path, empty Question/Section, and a nonempty
held reference matching `[a-z0-9][a-z0-9-]*`. Section refs additionally exclude
plan/yes/no/y/n/true/false/on/off/null. No arbitrary Detail.path prefix, malformed
index, empty reference, wrong identity, shape, completeness or other validation
problem is allowed. An empty actual problem batch means valid/available support.
If a NONEMPTY batch consists entirely of qualifying stale problems, keep their
original Stage/Problem in Support.Problems, Available=false, zero Recipient and
nil Arrangements. This separate nonfatal support channel retains Severity:error;
do not append those stale errors to the fatal/ordinary []Diagnostic channel or
relabel them warnings. One other error makes the whole actual batch fatal.
Missing preparation is NOT this allowance. No repair/reconciliation occurs.

#### Own diagnostics

| Stage / Code | Exact Message | Detail |
| -- | -- | -- |
| context / session.context | Select an explicit valid account and plan identity. | path: context |
| identity / session.plan | The shared plan does not match this session. | path: plan.id |
| identity / session.record-account | The record account does not match this session. | path: record.account |
| identity / session.record-plan | The record plan does not match this session. | path: record.plan.id |
| record.ownership / session.record-owned-answer | Keep preparation answers out of the newcomer record. | path: answers.plan[N] or answers.sections[N].values[M] |
| record.history / session.outcome-kind | Repair the unknown recorded outcome kind before editing this session. | path: outcomes[I].kind |
| intent / session.intent | Choose exactly one session change. | none |
| proposal / session.proposal | Review this change again. | none |
| event / session.event | An answer change does not use an event timestamp. | none |
| record.roundtrip / session.representation | This record needs a lossless save path. | none |

### Propose(Source,Intent)

Apply raw/intent admission before expensive work, then repeat complete Inspect.
Only then validate a single intent, Stage intent. Operation exactly answer,
follow-plan or outcome; Value.Kind one of the five model kinds. For answer only
Question/Value may be nonzero; for follow-plan only Question may be nonzero; for
outcome only Section/Kind/Note/Differently may be nonzero. All other fields must
DeepEqual their exact zero values (including nil inactive List). Empty target
values reach the actual writer/recorder for their specified refusal; no target
guessing, whitespace repair or multiple changes.

For answer map to real writer6 submit. For follow-plan map to its follow-plan.
Pass the ORIGINAL shared Plan and ORIGINAL held Record, not the joined plan as
shared input and not a fabricated previous record. Stage answer preserves its
complete Problems. This is critical to equal-preseed ownership. Use its returned
Record exactly. NeedsEvent=false.

For outcome call real Recorder5 on the actual joined effective Plan, ORIGINAL
Record and exact intent proposal. Stage outcome. Preview-only injected timestamp
is fixed `2000-01-01T00:00:00Z`, solely to run its real current gate before
asking the driver for consent. NeedsEvent=true. The sentinel is NEVER a claimed
event and never exposed in public Proposal.AfterRecord or After.Outcomes: blank
only the new target's On in those projection fields, with NeedsEvent as the
normative uncommitted label. Preserve every unrelated historical On exactly,
even if it happens to equal the sentinel. Do not blank all matching strings.
The driver must display `sampled after save`, not a blank historical event.

Cheap/full candidate admission and actual effective validation/Walker repeat
for the real recorder candidate WITH its internal preview timestamp; blank the
target event only when constructing the public review projection afterward.
Before/BeforeRecord are original inspected session/record;
After/AfterRecord are real candidate projections, except target preview On as
above. Answer edits never sort history; outcome Recorder's specified complete
history sorting is retained. Sources and intent are owned copies in Proposal.
No encoded candidate is published by Propose. Warnings accompany success.
These candidate calls have exact diagnostic stages candidate.validate and
candidate.walk; their Problems are unchanged, not relabeled or deduplicated
against the earlier effective.validate/walk warning batches. A newly activated
invalid held answer therefore refuses the candidate through actual validation.

### Finalize(Proposal,event)

Predecode/intent/event admission first. Recompute Propose from Proposal.Source
and Proposal.Intent, using actual cores, and compare ALL supplied projection
fields, records and NeedsEvent with its complete result. Ignore nil vs empty
ONLY for derived top-level Session result slices and diagnostic lists where
dependency contracts permit either; do not equate nested model shapes or alter
Source byte slices/Intent. Compare supplied projections in an expected-bounded
way: fixed field/type checks, collection lengths, key membership and string
lengths first; descend only the already admitted rederived expected structure.
Do not deep-copy/traverse an unchecked forged collection, cyclic unknown tree,
or arbitrary interface value before matching it. No supplied projection may
increase the comparison traversal beyond the admitted expected shape. Successful
outputs are built from rederived owned values, not by trusting supplied models.
A mismatch returns session.proposal, not a forged
target patch. Earlier reached errors refuse with their actual batch. The pure
API verifies source-derived consistency, NOT authenticated consent or native
freshness: a caller supplying a wholly self-consistent new Source cannot be
distinguished from legitimate new input. The driver must compare exact frozen
source bytes/guards and successful displayed consent separately.

For answer/follow-plan require empty event, otherwise session.event. Use actual
writer result. For outcome call actual Recorder AGAIN with the exact supplied
event (including empty/invalid text), original effective Plan/Record and intent.
Its full kind/timestamp/gate Problems are authoritative. Never substitute the
sentinel, normalize, generate, resample or borrow FirstOpenedAt. Any nonempty
accepted event is returned exactly as Publication.Event. Answers return empty
Publication.Event. Re-run cheap/full candidate admission with actual final
record, then actual validation/Walker at candidate.validate/candidate.walk.
No stale preview bypass. Finalize returns the complete reached rederivation
warnings followed by any actual finalization warnings/error; no warnings are
invented from supplied projection fields.

If final candidate DeepEquals ORIGINAL decoded record, return independently owned
EXACT Source.Record bytes, Changed=false, without canonical encoding. Otherwise
real record Encoder at Stage record.encode returns canonical bytes (bounded1MiB
before return); real record Decoder at Stage record.roundtrip checks them.
Compare intended and roundtripped records after ONLY this known-container normal
form: empty Answers.Plan→nil; omit empty Answers.Sections scope maps; empty
Answers.Sections→nil; empty Outcomes→nil. No Value entry/Kind/List, unrelated
outcome field/order, unknown map/list/scalar/null, account/reference/timestamp
or metadata is normalized. Unknown empty {} and [] remain data. Remaining
inequality returns session.representation, not silent data loss. Ordinary
documented known empty-container canonicalization is NOT a refusal. Return
canonical owned bytes, Changed=true. No byte/source/result mutation.

## 3. Dependencies and placement

Position7: actual record/writer6; outcome/recorder5; section/walker4;
plan/validator4; plan/authoringvalidator2; actual plan/record/preparation codecs;
preparation/validator; record/joiner1; model and own contract. Standard SHA256
compares raw plan bytes only, never substitutes for real section hasher via
Walker/Recorder. Standalone admission.md defines caller accounting; no assumed
invitation API/dependency. No Split, reconciler, another driver, clock or I/O.
Driver8 and binary9 sit above this full set; entry growth is allowed by0020.

All mutable results, Source copies, questions, sections, unknown trees and
diagnostic maps deep-own storage, including no-ops and refusals. Calls can run
concurrently without shared mutable state. Effects/native security are excluded,
not certified by a successful pure result.

## 4. Independent executable fixture contract

External Go tests and literal raw YAML sources accompany this contract. Only
wire_test.go imports the proposed component for one New call; contractstub uses
a separate returning-nothing stub adapter. Expected bytes/hashes/code/path/order
are authored literals/public definitions, not implementation-generated goldens.
Provenance maps each oracle and the finite planned wrong variants. No execution,
stub failure, test pass, mutant kill, native or shell guarantee is asserted here.

| Plausible wrong behavior | Independent fixture discriminator |
| -- | -- |
| authorize guessed identity or bind empty reference | ExplicitContextAndCompleteIdentityBatch |
| choose a duplicate section/outcome winner | NoAmbiguousPlanOrOutcomeWinner |
| let held onboarder answers overlay or ignore unknown kinds | OwnershipAndUnknownKindGuardAllHeldEntries |
| keep only first actual error or skip candidate validation | ActualCodecAndDeclarationRefusals, IntentIsOneExplicitOperationAndWriterErrorsPassThrough, CandidateValidationRejectsNewlyActiveInvalidHeldAnswer |
| lose equal accepted ownership or silently accept suggestions | EqualPreseedOwnershipAndFollowPlan, SuggestionsAndRawNoOp |
| invent completion/readiness from preparation/questions | EveryKnownSupportFieldAndHelperRepeatIsOnlySupport, MissingQuestionsAreNotAnExtraOutcomeGate |
| use template hash, preview event, or stale prerequisite gate | OutcomePreviewIsUncommittedAndFinalizeUsesActualEvent, OnlyTargetPreviewEventIsBlanked, CurrentDirectPrerequisiteGate |
| trust forged proposal or alias source/result | FinalizeRederivesEverySuppliedProjection, SourceIntentQuestionsAndUnknownTreesAreIndependentlyOwned |
| lose unknown {}, [] or canonicalize an unchanged record | KnownEmptyCanonicalizationPreservesUnknownEmptyData, SuggestionsAndRawNoOp |
| omit source/root resource charge or reject inclusive threshold | ExactChargedBoundaryCountsAllThreeSourcesAndRoots plus independently worked score-oracle.md |
| inspect only active data or enforce budget after expansion | ChargedFanoutIncludesDisabledDeclarations, SectionCapAndUnknownDepthBeforeExpandingCores; source placement still needs independent implementation review |

These are planned discriminators, not a report of executed variants. Provenance
records the finite one-defect verification procedure and remaining evidence gaps.
