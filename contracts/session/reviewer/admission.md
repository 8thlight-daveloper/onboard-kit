# Session admission profile

This is a standalone normative caller profile for session/reviewer7, selected
under the current task delegation, not a new public estimator API. No invitation
archive, synthetic ZIP, invitation reviewer or driver is constructed to apply it.
No renderer/finder/validator estimates its own input budget.

## Predecode admission

Each Source body Plan, Preparation, Record is at most1048576 bytes, inclusive.
Check their raw lengths in that order BEFORE any decoder, copying, or derived
model work. Sum E exactly once as len(Plan)+len(Preparation)+len(Record), even
when bodies are byte-identical, maximum3145728. No decompressed/archive term.
Do not infer UTF8 from string conversion; actual source codecs judge raw UTF8
after length admission. A missing/empty body reaches its actual decoder; it is
never replaced by a synthesized body.

Source.Account and Source.PlanID each require valid UTF8 and at most4096 bytes,
before any decode. Empty/invalid trusted identity grammar is a later context
error, not proof that one of the sources may choose a substitute identity.

For Propose and Finalize check all Intent strings before Inspect: Operation,
Question, Section, Kind at most4096 bytes each; Note, Differently, Value.Text
at most65536 each; Value.List at most4096 elements, each at most65536 bytes.
All require valid UTF8, including unused fields. Total bytes over every listed
string and every list element at most262144, inclusive, with overflow-safe
length arithmetic before aggregate allocation. This counts payload even for
None/Deferred or an operation that will be rejected later. Validate in declared
order above, then list order, then total. It is not accepted-answer validation.

Finalize additionally admits its supplied event as UTF8 of at most64 bytes.
Within that bound actual outcome Recorder judges the new timestamp spelling;
this profile never normalizes it. No clock is read or inferred.

Predecode failure returns one own Diagnostic: Stage `source.admission` (body
or identity bounds), `intent.admission`, or `event.admission` respectively;
Code `session.resource`, Severity error, Message exactly
`Use a smaller onboarding session or fewer repeated values.`, empty
Question/Section, Detail exactly `{"path":PATH}`. PATH is `plan`,
`preparation`, `record`, `account`, `planId`, `intent`, or `event` respectively.
Intent details never expose a supplied key, string, cap name or calculated size.
All operation result fields are zero on resource refusal. Source bounds precede
intent bounds, which precede event bounds in Finalize. No truncation follows.

## Frozen postdecode model policy

Normatively incorporate the EXACT frozen policy definitions in the committed
`contracts/invitation/reviewer/admission.md`, headings Saturation, W: exact model
occurrences, Other inclusive caps, Directory and rendering terms, Candidate and
suggestion terms, and Final score. These headings and the enumerated model
fields/caps/formula are part of this proposal by reference, not automatic
adoption of a later edit. A future change to that public policy requires review
of this contract before substitution. Keep the referenced contract present in
every disposable build/test tree. No implementation is referenced.

In particular B=67108864, X=B+1; all sum/product work saturates at X; charge
zero/nil root occurrences and every inactive payload, hidden declaration,
unknown tree and repeated item. All public per-list/combined cardinality,
depth128, occurrence131072/unknown65536 and known-entry65536 guards apply before
score work. Exact W fields and Render/Questions/EligibilityWork/PairWork terms
are unchanged. Score B is admitted, X refuses. No map-order-dependent totals,
default acceptance, field selection, graph interpretation or identity merging.

The session contexts are:

| Stage | Roots at this point | E | Detail.path |
| -- | -- | -- | -- |
| plan.admission | full decoded P, zero R, nil T | all three exact source lengths | plan |
| record.admission | cheap current P, decoded R, nil T | not used | record |
| preparation.admission | cheap current P, decoded R, pointer to decoded T | not used | preparation |
| payload.admission | full current P, decoded R, pointer to decoded T | all three exact source lengths | session |
| effective.admission | full actual Join(P,R), actual R, pointer to T | all three exact source lengths | effectivePlan |
| candidate.admission | full actual Join(P,candidateR), actual candidateR, pointer to T | all three exact source lengths | candidate |

Cheap and full mean exactly the referenced frozen policy, not different model
guards. P/R/T share counters; never reset unknown/depth/node/collection totals
between roots. Full plan admission precedes authoring or finder calls. Cheap
record/preparation admission immediately follows its decode before any identity,
ownership, history loop or value copy. Full payload precedes preparation
validation/Join. Full effective precedes effective validation/Walker. Candidate
admission precedes candidate validation/Walker/encoding. The writer and recorder
receive already admitted effective inputs and separately bounded Intent; after
their one selected change, cheap candidate forest guards precede Join and full
candidate admission precedes expanding consumers. That cheap candidate guard
uses Stage candidate.admission and the same path candidate.

All postdecode failures use the one fixed session.resource Problem/message
above, exact table Stage/path, no cap/score/source value in diagnostics, and no
partial usable result. Return preceding reached dependency warnings followed
by this failure, not diagnostics from stages never reached. Predecode failure
has no preceding dependency diagnostics.

The existing64 event reservation is charged, but a newly supplied outcome
timestamp and changed note/differently may change W(Render) too. Finalize runs
actual candidate admission with the REAL event; it can refuse without output if
that final candidate crosses the limit. A successful preview does NOT promise
final admission for every possible longer event. The shell must report that
honestly and must not stage anything before Finalize succeeds. This is not a
clock-resampling license. Account/reference/first-use are never synthesized.

## Guarantee and limits

This establishes the specified finite raw/admission/fanout profile before the
actual expanding semantic cores. It is not a heap/RSS, decoder parser, RE2 work,
wall-time, terminal-output or native publication bound. Postdecode resource
checks cannot undo decoder allocation; the earlier raw1MiB bound is stated
honestly, not described as proof of parser memory safety. Unknown unsupported
model shapes are not supplied by these actual successful codecs. No reflection
over future model fields silently broadens this frozen policy.
