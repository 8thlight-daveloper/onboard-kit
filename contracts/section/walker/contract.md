# Section walker contract

Status: agent-authored contract; independent implementation integrated.
Component `section/walker`, package `walker`, dependency position 4.

Sources: ADR 0085, 0090, 0140, 0180, 0250, 0260, 0270; principles 0100,
ambiguity rule 0110 and default rule 0120. Paths are relative to `contracts/section/walker`.

## Interface and composition

`New()` takes no arguments and constructs the real scheduler, question finder,
section resolver and section hasher. No substitute dependency, outcome map,
caller-supplied hash, clock, audience, timestamp or path argument exists.
`Walk(effectivePlan, record) -> (View, problems)` uses the API's component-owned
derived view types, not new saved model fields.

The caller has already composed the effective plan with authorized preparation
and record answers under their existing precedence and ownership contracts.
The walker reads record.Outcomes only. It must not join record answers again,
read preparation independently, authorize identity, infer first use, or compare
the record's last-reconciled whole-plan reference hash with current content.
Those are driver/other-component responsibilities. Full plan authoring validation
is a caller precondition; the scheduler still checks structural validity itself.

## Refusal, with no partial view

Call the real scheduler on the complete unchanged effective plan first. Return
all its structural problems unchanged and no view fields when it fails. Never
filter by outcomes, activity, bootstrap, or dates before that validation.

On structural success, inspect all held outcomes. For each exact Section string
that appears more than once, report one `walk.duplicate-outcome` error, with
Section that exact ID, empty Question, Detail exactly `{"field":"outcomes"}`,
and Message exactly `Keep one recorded outcome for this section.` Sort these
problems by exact Section UTF-8 bytes. Include identical duplicate values and
duplicates for inactive or absent sections. Return an empty view on any such
error, without resolver gap warnings or a partially usable question/action list.
This component neither selects a winner nor deletes history. Structural refusal
takes precedence; duplicate-outcome diagnostics are only returned after structural
success. Distinct case/whitespace identities do not collide defensively.

No additional malformed-outcome diagnostics are invented here. Outcome kind and
timestamp validity, record account/plan identity, secrets, directory integrity,
and invalid variable declarations belong to existing decode/authoring/runtime
validation. Unknown enum values and invalid Go UTF-8 are outside the successful
view guarantee. Empty unique exact IDs follow the scheduler's defensive scope.

## Active sections and waiting

Return exactly the actual scheduler's sections in its full global action order.
No second schedule, reference-order replacement, bootstrap-only filtering,
outcome-derived activity, or contiguous subtree requirement is allowed.
For every row, CurrentHash is the real hasher's result on the unresolved original
section, including custom source-less sections. It is never a hash of resolved
text, Source.Hash, a plan hash, or a modified copy with new metadata stripped.

WaitingOn contains each distinct direct Requires target that lacks a held
OutcomeVerified outcome. Keep the first occurrence order of the original
Requires slice. Preserve Requires itself, including order and repeats.
Failed, deferred and absent outcomes do not satisfy the gate. Verified still
satisfies it when its hash comparison is unknown or different. Do not recursively
recheck the prerequisites of a verified required section to invalidate its
historical verification. A dependent's own verified outcome does not clear its
WaitingOn for a new attempt. Do not reset that held outcome either.

Parent, effective After, timing goals, answers, first use, invitation data and
assignments supply no verification evidence. Waiting is not inactivity, failure,
or a reason to hide the dependent's questions/instructions. No `CanVerify` field
is promised: question blockers and resolver gaps are information, not newly
invented additional prerequisite gates in this increment. A future recorder
rechecks current inputs at the explicit write boundary.

## Resolved text, questions and support

Resolve each scheduled section through the actual resolver, passing the
complete effective Answers and Directory, not a reduced plan's data. Find the
team by the plan variable whose Type is team, regardless of name; use its stored
Text answer only when it names a directory team. Otherwise pass empty team.
Do not use a default, an arbitrary variable named team, or a first directory team.

SectionView.Section is that resolved Section. Its Escalation list stays held
as required by the resolver. SectionView.Escalation is a separate rendered list,
available for support on any row, not only after failure. Render each item as
the actual resolver fills Body with the same ID, Variables, answers, directory
and team. A minimal carrier Section is one permitted composition, not an extra
public API or an observable call-count requirement. Keep each result in held
item order including repeats. Items not declaring a valid variable
reference remain literal under normal resolver rules. Vacant roles remain
`Role name (vacant)`, not empty or silently replaced by team contacts. Never
resolve the already-resolved text twice or mutate the original definition.

Collect resolver gap problems across main text and escalation carriers. For a
section retain each (Question, Code) once, ordered by the corresponding variable's
declaration rank, then Code byte order. Append sections' problems in scheduler
order. Severity, Section, Question, Detail.audience and marker behavior follow
the resolver contract; require a nonblank Message without freezing its wording.
Do not invent document-specific vacancy diagnostics or suppress unknown-default
gaps. No problem may copy a note, differently text, personal account value or answer value.

View.Questions and View.Blockers are exactly the real finder's full-plan results
for the empty audience (both), in the finder's own order, not reordered to action
sequence. This preserves question-condition dependencies and ownership metadata.
Find on the same complete effective plan; do not remove waiting or already
verified work first. No synthetic progress/actual-start question is added.

## Complete held outcome history

Return every input outcome once in View.Outcomes sorted by exact Section bytes.
Keep every Outcome field exactly held, including Kind, On, SectionHash, Note and
Differently. The input list itself need not already be sorted. Do not activate
removed/inactive sections to obtain a history row or prune their answers.

For an outcome whose section still exists anywhere in the original plan,
CurrentHash is the actual unresolved section hash, even if currently inactive.
Comparison is unknown when the held SectionHash is empty, same when it equals
CurrentHash exactly, different otherwise. For a removed section, CurrentHash is
empty and Comparison unknown regardless of held hash. Compare arbitrary nonempty
legacy text as held, without case folding or trimming. Unknown is not freshness.
Use honest wording in callers: current definition differs from recorded hash,
not proof that an edit happened after the outcome. No history rewrite or Source
baseline modification is permitted by inspection.

## Purity and ownership

Read plan Sections/Variables/Answers/Directory needed by the actual dependencies
and record.Outcomes. Return derived values only. Read no clock, filesystem,
environment, process, network, config, Git state, planned date or first-use event.
Execute no check, command, environment lookup or message. Keep no mutable state.

Input values remain unchanged. View top-level lists, SectionView slots, WaitingOn,
Escalation, history rows and problem detail maps are new owned storage, independent
of input and later calls. Resolver-returned Sections have its existing deep-copy
guarantee. Question inner data follows the finder's existing immutable-sharing
contract rather than adding an unrelated deep-copy promise. Calls on one instance
may be concurrent. Permuting section declarations or outcome order does not
change the view. Requires and escalation item order are held data, not permutation
invariance claims. Nil and empty result lists are equivalent.

## Independent fixtures and planned wrong adapters

`testdata/cases.json` gives literal semantic action, waiting and history oracles.
Tests also pin an actual current-definition hash from independent canonical text,
history preservation and ambiguity refusal. No expected value is obtained by
calling a production scheduler, selector, resolver, finder or hasher.

`testdata/question-order.json` pins full original question context, cross-section
question dependencies, both audiences and a literal question order different
from global action order. `testdata/structural-errors.json` pins complete exact
multi-problem pass-through even when every declaration is off. The activation
tests also exercise owned nested section storage in both directions, waiting
and escalation lists, diagnostic maps across calls, and concurrent successful
and refused inputs on one walker instance. These are authored tests, not verified
safeguards; independent stub/build/hardening evidence remains required.

| Plausible wrong implementation | Planned distinguishing oracle |
| -- | -- |
| Return grouped order or filter to ready/marked rows | Cross-subtree and waiting action cases |
| Treat failed/deferred/absent as verified, or use freshness as evidence | Gate matrix and unknown/mismatching hash cases |
| Recheck transitive history, clear verified dependents, or use Parent/After/date gates | Historical dependency and containment/order controls |
| Select first/last duplicate outcome, or ignore removed duplicates | Exact refusal cases and reversed input order |
| Use Source.Hash, hash resolved text, skip custom/inactive content | Independent canonical hash test and inactive history |
| Overlay record answers again or infer outcomes from first use | Contradictory record answer and first-use controls |
| Lose removed history or reset notes/times/hashes | All-field history equality assertion |
| Reorder questions into action order or remove waiting questions | Rich literal question-order fixture |
| Drop or reorder structural problems after filtering inactive work | Literal seven-problem structural fixture |
| Resolve escalation privately, drop vacancy or a repeated item | Literal directory/support composition test |
| Mutate inputs, keep state or share result backing data | Purity, nested/list/map ownership and shared-instance concurrency tests |

These adapters have not been executed. Mutation, independent build and differential
evidence belong in a working report, not a count/history in this contract.
