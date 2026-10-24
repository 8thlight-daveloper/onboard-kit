# Catalog editor contract

Status: independent proposed contract under0410 and proposed0440. No product
implementation, adoption or execution evidence. Proposed plan/editor exclusivity
clarification is a prerequisite to promotion, not silently assumed existing law.

## API and effects

New() constructs actual catalog materializer, plan editor, plan authoring
validator and plan validator. No injected components, catalog, filesystem,
clock, semantic policy, parser or hashes. Position6 permits only those earlier
components, model, own contract and pure standard library. Do not call drivers,
record/preparation components, decoder/encoder, refresher or hasher directly.

Edit receives an ordinary in-memory Plan and an ordered typed operation list.
Reads every held field for bounded ownership copying, exact identities,
declarations/answers, operation data, and actual compiled dependency results.
Writes only a returned proposed Plan and diagnostics. No input mutation, saved
file, record, preparation, outcome, clock, environment, network, terminal,
subprocess, persistent reservation database or hidden session state.

Any failure returns exactly zero model.Plan and the failing diagnostic batch.
Never return the original as an apparently successful proposal or an edited
prefix. Caller retains the original unchanged. Success, including zero
operations, returns a deeply independent plan and advisory warnings if any.
Nil/empty diagnostic lists are equivalent. Every mutable result object belongs
to the caller, including unknown maps, nested lists, pointers and diagnostics.

## Finite phases and precedence

1. Bounded admission below precedes expanding dependency calls. Own failure is
   exactly one fixed diagnostic. Plan.ID must be nonempty; retain it byte-for-
   byte, with SchemaVersion, provenance, origin, URL and results path unchanged.
   This is not a decoder, schema upgrade or personal-plan ownership inference.
2. Call actual authoring Validate on the original plan. Errors refuse the whole
   actual batch through safe projection. Only without errors call actual plan
   Validate on the complete original plan. Errors refuse that batch. Warnings
   alone permit edits; initial successful warnings are discarded, since final
   validation reports the resulting state once.
3. Check operation union shape in input order. Reserve EVERY original Section.ID
   and every key of original Answers.Sections, even disabled, hidden, nil or
   empty scopes. Check every Add SectionID against these reservations and prior
   Adds in operation order, before any source lookup or edit. A reserved ID
   cannot be reused after Retire/Reactivate, even for the same source. Do not
   deduplicate Adds. Blank/invalid permanent IDs still go to actual materializer
   unless reservation collision already refuses. No historical IDs are inferred
   from absent records, discarded definitions, prose, URLs, titles or hashes.
4. Start a deep private candidate. Apply operations one at a time as below in
   exact caller order. Each local/delegated error aborts the entire batch. Keep
   only the failing batch, discarding prior successful-operation warnings. No
   full-plan validator runs between operations; actual local dependency checks
   still apply to the candidate at that point.
5. Recheck the resulting known/unknown value bounds, then actual authoring
   Validate and actual plan Validate on the complete resulting candidate, in
   that order. A failing whole batch refuses atomically. Do not validate only
   active/new sections, mask errors by retirement, or guess relationship repairs.
6. Success warnings are successful operation warning batches in operation order,
   then final authoring warnings, then final plan warnings. A phase's actual
   order and multiplicity stay intact. No earlier admission warning duplication.
   Return the candidate with full ownership isolation and no other changes.

Input-invalid plans cannot be repaired through this narrow flow. The separate
definition/data repair flow must address them first. Inactive held answer kinds
remain held under actual validator policy; Reactivate may expose a final error,
and a subsequent Submit in the SAME batch may explicitly repair that answer.

## Typed operation grammar

Every operation has Kind and SectionID. Only these additional fields apply:

| Kind | Additional active fields |
| -- | -- |
| add | SourceID, Bindings |
| submit | Name, Value |
| unset | Name |
| retire, reactivate | none |

An unknown Kind or a nonzero inactive field refuses edit.operation at
operations[I]. Zero Value means the complete zero struct, not merely KindNone.
Nil and allocated empty inactive Bindings are both permitted. No hidden
metadata, rename, reorder, replace-definition, refresh-source, prune, delete,
directory edit, plan-variable edit or arbitrary field patch operation exists.
Active Value accepts exactly the five public model kinds; unknown Kind refuses
edit.value-kind before calling a dependency with an unsupported value. A Submit
Value's inactive payload fields are counted/copied and passed exactly as held.
An Add binding Value follows the materializer's own kind/admission phases.

Repeated Submit/Unset/Retire/Reactivate operations are explicit sequential
instructions, not duplicate declarations. They may target a preceding Add.
A target declared by a later Add does not exist yet. Repeating Add with one ID
refuses even when a prior Add is later retired. Omission never retires or deletes.

## Add exact independent instances

For each Add, invoke actual materializer.Create with original exact PlanID,
zero Provenance, a copied current Directory, no PlanVariables/PlanBindings,
and exactly one Instance with supplied SourceID/SectionID/Bindings. Do not pass
unrelated existing plan variables into that subproposal: a used onboardee
variable in an existing section would otherwise appear unused in the isolated
new-definition request. Current graph-free catalog sources need no external
plan-variable references. Directory lookup for explicitly supplied bindings
still uses the actual held directory.

Any materializer error refuses with that complete ordered batch. On success,
append its one returned Section, unmodified, to the candidate. Append order is
operation order, not sorted source/ID order; preserve the original prefix order.
Attach only its explicit nonempty returned answer scope, if any, without changing
other scopes or allocating unrelated empty maps. Bindings are exact values,
including None, Deferred, empty text/list, false and inactive payloads, as the
actual materializer supplies. None may be unencodable later; this pure core
does not promise encoding success or silently change it into Unset.

Do not copy subproposal root schema/provenance/origin/directory/maps over the
held plan. Never pass existing definitions through materializer, rehash their
current content, infer their old baseline basis or repair Source. Final-ID
hash and source identity for NEW instances come only from the actual dependency.
Advisory guides are not stored in the Plan or claimed completed; a future shell
may separately expose actual guidance under its own contract.

## Submit and Unset exact local shared facts

The target must be one currently held section, never plan scope. Unknown target
refuses edit.section at operations[I].sectionId. Require an exactly declared
local variable Name. Unknown variable refuses edit.binding at operations[I].name.
The input/full validation gates already reject ambiguous declarations before
lookup; never select a first/last duplicate. Only Audience empty/onboarder
permits changes; onboardee refuses edit.owner even for Unset/None/Deferred.

For Submit call actual plan editor.Submit(candidate, SectionID+"."+Name, Value).
For Unset call that same actual operation with the zero Value. No second local
CheckValue, finder, pattern/option/directory parser or answer-writing policy.
Apply the real returned plan only without errors, retaining actual warnings.

These are existing editor semantics, deliberately not creation Bind aliases:
None, empty Text, empty List and List of only empty strings clear the key;
false, text of spaces, nonempty lists with repeats/empty items and Deferred are
stored as the editor specifies. A literal Text "deferred" is not Deferred.
Inactive/unknown proposed questions still refuse through the real dependency;
Unset cannot clear an inactive held answer. No cross-scope fallback/propagation,
default/example acceptance or personal answer change is authorized.

For deterministic outer shape, when a Submit/Unset actually removes the last
key from a previously nonempty section answer map, remove that one scope entry.
Do not delete preexisting unrelated empty/nil scopes, normalize outer maps,
allocate unrelated maps or alter unselected Value payloads. This resolves only
the existing editor's deliberately unspecified emptied-scope shape in this new
wrapper; it does not alter its contract or stored-answer clearing semantics.

## Retire and Reactivate same identity

Require an existing exact section target, otherwise edit.section. Delegate to
actual plan editor.SetSectionEnabled(candidate, SectionID, false/true).
Retire stores a false pointer. Reactivate removes Enabled, including held true.
Retire of an already retired target and Reactivate of an already-on target are
allowed idempotent operations. They do not promise resource equivalence: the
caller explicitly chooses same identity; new resources must use Add/newID.

Preserve every definition field, Source and baseline, answer scope, descendant
switch and relationship exactly. No cascade, relationship redirection, removal,
outcome/hash migration, record reconciliation or completion claim. The real
whole-plan final validation remains authoritative after applicability changes.
Retire before Submit can yield answer.inactive-question and roll back the entire
batch. Reactivate before Submit can permit it. Do not reorder either sequence.

## Preservation and ownership

Every unrelated held field stays byte-for-byte/value-for-value and shape-for-
shape, including Source nil/present-empty/stale/arbitrary hashes, locally edited
definition text, conditions, Parent/After/Requires order/repeats, variable
metadata/defaults/examples, steps/checks/environment, Bootstrap, SchemaVersion,
ID, provenance, origin, URL/results path, directory, plan answers, hidden scopes,
unknown metadata and nil versus allocated empty containers/pointers.

Only explicit addition appends Section/its supplied facts; Submit/Unset changes
one authorized scope entry and the specified emptied-scope cleanup; retirement
changes one Enabled. Preserve all other sequence order. Reordering original
Sections changes only the retained prefix sequence, never target identity,
facts, baselines, success or warning contents apart from actual indexed
authoring paths. No stable-ordering claim for indexed paths after permutation.

All mutable result values are independent of inputs, operation payloads, actual
dependency temporaries, earlier/later calls and other new instances. Mutating
result Source, unknown nested values, conditions, submitted lists or directory
maps cannot mutate any other owner. Shared Editor calls are deterministic and
safe to overlap; no process-global remembered reservations or content cache.

## Bounded admission

The accepted held-value graph is ordinary model data plus unknown values made
of map[string]any, []any, string, bool, int, int64, finite float64 and nil, retaining
exact numeric dynamic type and nil/empty shape. No codec roundtrip is a clone.
Unknown unsupported dynamic values refuse edit.held-shape without invoking
methods. Cycles/depth exhaustion refuse edit.limit before recursive expansion;
shared acyclic nodes may be copied independently. Never normalize unknown keys
or treat unknown metadata as executable operation or reservation policy.

Inclusive local limits: operations1024, Adds256, held/result Sections4096,
each held collection8192 entries, all collection entries262144, all text
occurrences33554432 UTF8 bytes, each text1048576 bytes, each operation identity/
Name/SourceID4096 bytes, unknown nesting64. Count held plan and operations
together, including ignored payloads, map keys, repeated occurrences, conditions,
defaults, known and unknown fields. Each map entry counts once, each list item
once; a map's value-list entries also count. Unknown root has depth0; each nested
map/list value increases depth1. Reject a container over its cap before copying
its keys/items. Known recursive models contain no unknown pointers outside the
specified data containers. Saturate sums at bound+1. Existing value aliases do
not exempt repeated occurrences from accounting.

Traversal is struct declaration order from model/API, slice order, map keys in
exact byte order. Limits/unsupported unknown shape/UTF8 failures use the first
encountered guard; field strings are counted before UTF8 judgment at that field.
Apply these same guards to the final candidate before validators. Invalid UTF8
anywhere, including unknown keys/inactive payloads, refuses edit.text. This
admission narrows the new operation only, not codecs or legacy editor policy;
it promises no exact heap/time/serialized-size bound. No truncation follows.

## Private useful diagnostics

Diagnostic has Stage, Path, Code, Severity and fixed Message only. No raw
Question, Section, Detail, parser text, rejected identifier, secret, pattern,
answer or metadata value is returned. Preserve actual dependency Code/Severity,
ordered multiplicity and failing-batch precedence; only safely project metadata
and wording. No dependency diagnostics means no invented diagnostic.

Own diagnostics have severity error and these exact fields:

| Code | Stage | Path | Exact Message |
| -- | -- | -- | -- |
| edit.limit | admission | request | Reduce held plan or edit data to the supported limits. |
| edit.held-shape | admission | plan | Supply ordinary decoded metadata values without unsupported objects. |
| edit.text | admission | request | Use valid UTF8 held plan and edit data. |
| edit.plan-id | admission | plan.id | Select an existing plan with an explicit identity. |
| edit.operation | operation | operations[I] | Use exactly the fields of one supported explicit edit operation. |
| edit.value-kind | operation | operations[I].value | Choose a supported submission value kind. |
| edit.reserved-id | reservation | operations[I].sectionId | Choose a new permanent ID; held definitions and answer scopes reserve their IDs. |
| edit.section | operation | operations[I].sectionId | Select an exact section already held at this operation. |
| edit.binding | operation | operations[I].name | Select an exact local variable declared in this section. |
| edit.owner | operation | operations[I].name | Change only onboarder shared facts; preserve newcomer-owned answers. |

Delegated stages are input-authoring/input-plan, add- plus the materializer's
actual Stage, submit, switch, final-authoring/final-plan. Their wrapper Path is
plan for input/final checks and operations[I] for edit calls; materializer adds
its own indexed Path after a dot when nonempty. Authoring's safe indexed
Detail.path may be appended after plan.; no other raw metadata is projected.
Plan-validator Question/Section/Detail is never concatenated into a path.

Fixed repair messages by exact actual code:

| Code | Exact Message |
| -- | -- |
| catalog.unknown-id | Choose an exact section-template ID from catalog list. |
| materialize.carrier | Select a section-template source for an activity. |
| materialize.instance-id | Supply a nonreserved lowercase permanent section ID. |
| materialize.duplicate-binding | Supply each scoped binding name once. |
| materialize.binding-name | Bind an exact variable declared in this scope. |
| materialize.binding-owner | Keep newcomer answers outside the shared organization plan. |
| answer.inactive-question | Reactivate or satisfy this question's declared applicability before submitting or unsetting. |
| answer.unknown-question | Preserve undeclared held answers for explicit history review; submit only declared local questions. |
| answer.ambiguous-question | Repair duplicate variable declarations before editing this question. |
| answer.wrong-kind | Submit a value kind matching the declared variable type. |
| answer.not-an-option | Submit an exact declared choice option. |
| answer.pattern-mismatch | Submit a value matching the declared pattern. |
| directory.unknown-id | Submit an exact declared directory target, or leave the binding unset. |
| plan.duplicate-section-id | Repair repeated held section IDs before editing. |
| plan.duplicate-variable-name | Repair repeated variable declarations within their scope before editing. |
| plan.duplicate-step-id | Repair repeated step IDs within their section before editing. |
| secret. prefix | Remove credential-like content; supply only nonsecret shared facts. |

All other codes use exactly one stage fallback: input/final authoring →
Repair the indicated held declaration before editing.; input/final plan →
Repair the plan's reported structure or stored facts before editing.; add-* →
Repair the indicated new instance or ask the catalog maintainer.; submit →
Correct the scoped submission before applying this batch.; switch →
Select a valid held section switch before applying this batch.
Code table takes precedence over fallback; only secret. is a prefix match.
Actual validators remain sole semantic authorities; this fixed projection is
not a second validator and never treats a warning as completion/readiness.

## Four parts and required evidence

api.go is the permanent interface proposal. External tests and literal data are
independent of prospective product source. wire_test.go begins with zero stub;
root must establish assertion-based stub-red before source construction. Source
author must not read judging fixtures before first complete source freeze.
Expected new baselines use authored canonical JSON and standard SHA256 only,
never actual materializer/hasher/encoder output as the expected result.

Fixtures cover existing locally edited/stale-baseline/source-less history,
reorder plus new repeated-source addition, retire/reactivate/Submit ordering,
exact new binding versus existing clearing semantics, reserved hidden/disabled
IDs, ambiguity, graph/condition retention and immutable ownership. Plausible
wrong variants and execution obligations are separately recorded. No verified
test/mutation/differential or reviewed-save claim follows from drafting.
