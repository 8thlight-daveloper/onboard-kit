# Catalog materializer contract

Pure creation under ADR0410,0030,0100,0110,0120,0150,0270 and0320.
The API, fixtures, tests and constructor binding are one contract.
No implementation, passing check, reviewed save, update or retirement is supplied.

## API and declared effects

`New()` constructs the actual catalog reader, section loader, unchanged section
hasher, plan authoring validator and plan validator. It accepts no dependencies,
catalog bytes, manifest, filesystem or alternate semantic policy. `Create` takes
only the request in api.go. Empty selection is meaningful. No old Plan, Record,
preparation, recipient, outcome, clock, account or unknown manifest is an input.

Position5 permits model, this contract, catalog reader1, section loader1,
section hasher1, authoring validator2, plan validator4 and pure standard library.
Do not call another component, legacy instantiator, renderer, finder, encoder,
editor or driver directly. No parsing, hashing, directory-validation or answer
validation algorithm is copied into this component.

Reads: explicit creation context, selected instances and bindings, compiled
catalog definitions, the exact authored preparation relations, and actual
dependency results. Writes: a new returned plan, advisory guidance and diagnostics
only. No input changes or filesystem, environment, process, network, clock,
preparation, record, personal event or saved-plan operations. Calls are independent,
deterministic, safe to overlap and hold no mutable selection state.

## Finite phases and precedence

Each phase completes before the next. Own failures return exactly one diagnostic
and the zero Result. A delegated phase returns its actual ordered batch, with
the wrappers below, and zero Result when any problem has severity error.
Warnings alone permit proceeding. Never return a partial plan or partial guidance.
Any failure discards earlier successful phases' warning batches and returns
only the failing own diagnostic or actual failing dependency batch.
Successful diagnostics are authoring warnings followed by final-plan warnings,
in their actual dependency order, followed by guidance warnings if any. Empty
diagnostic slices may be nil or empty.

01. Preflight all caller-held cardinalities, then text byte limits, then UTF8.
    These guards include ignored payload fields and unselected Value payloads.
    They precede semantic lookup or expanding validation. No truncation follows.
02. Require nonempty PlanID, retaining exact supplied text. PlanID is an opaque
    identity, not a section slug; spaces and printable Unicode are preserved.
    Copy Provenance exactly, including empty fields; read no clock or build value.
03. Check every permanent section ID in request order. It must match
    `[a-z0-9][a-z0-9-]*` and exclude plan,yes,no,y,n,true,false,on,off,null.
    Reject the second occurrence of an exact repeated permanent ID even when
    source and bindings agree. IDs are never trimmed, casefolded or synthesized.
04. Check repeated binding Names within each exact scope: plan bindings first,
    then each instance's binding list in request order. A second occurrence
    refuses even when Value agrees. Names in different scopes are independent.
    Lookup has not yet selected a declaration or interpreted an answer.
05. Read each selected SourceID exactly through the actual catalog reader, in
    request order. Preserve its failure problems with stage source and that
    instance's sourceID path. Unknown lookup is not a nearest match or path.
    A known entry must have carrier section-template. Preparation-guidance and
    tool-help refuse, including invitation-prepare, invitation-start and
    onboarding-resume. All59 current section-template sources are admitted;
    repository and reading examples are not an allowlist. Repeat any admitted
    source with different permanent IDs, including application-local or working-hours.
06. Load the exact selected content with the actual section loader. Any problem
    refuses. The loaded ID must exactly equal SourceID. Refuse a loaded definition
    unless Parent is empty, Requires/After have no elements, IncludeIf is nil,
    every variable AskIf is nil, Source/Enabled are nil and UnknownFields has no
    entries. A present empty condition is not absent. This initial graph-free
    profile admits all current sources; future relationship mapping is separate
    work. Do not rewrite references, drop edges, flatten children or derive IDs.
07. Copy every loaded field and nested value unchanged except ID and Source.
    Set ID to PermanentSectionID. Hash this final-ID unresolved definition with
    the actual unchanged hasher, before attaching Source. Set Source.Template
    to original SourceID and Source.Hash to that actual hash. Do not hash answers,
    resolve placeholders, modify step IDs or use the original-ID hash baseline.
    Sort resulting sections by exact permanent-ID UTF8 bytes. Preserve Variables,
    Steps, Runs, Checks and every other ordered list inside each section.
08. Build the unbound plan specified below. Run actual authoring Validate first.
    Errors refuse with its whole batch. Only without those errors run actual
    plan Validate on this unbound plan, so duplicate declarations and directory
    ambiguities are rejected before any binding declaration is selected. On
    error preserve its whole batch. Its successful warnings are not emitted
    twice; the final-plan validation below supplies the final warning batch.
09. Resolve plan bindings against exactly one plan declaration, then instance
    bindings against exactly one declaration in that new section. No qualified
    names, cross-scope fallback, source-ID lookup or same-name propagation.
    Unknown names refuse. Only Audience empty or onboarder permits a binding;
    onboardee refuses even when Value is None or Deferred. Invalid audiences
    and duplicate declarations already failed real validation. Missing bindings
    are allowed, including required declarations. Do not fill defaults or examples.
10. Install exact bound Value structs into fresh scoped maps, including None,
    Deferred, false, empty text/list and inactive payload fields. Do not normalize
    Value payloads or drop explicit keys. With all accepted scope bindings held,
    call actual plan CheckValue for every supplied binding: plan list then
    instance lists, in request order. Return the first failing call's complete
    ordered problems. This uses submitted-answer semantics: an unknown directory
    target is an error even though Validate treats a stored unknown ID as warning.
    Unknown ValueKind is an own shape error before calling dependencies with an
    unsupported value. A None/Deferred still requires authorized known ownership.
11. Run actual plan Validate on the complete bound plan. Errors refuse. It checks
    the whole resulting value; do not treat required unanswered questions as
    refusal, call a second shadow validator, infer readiness or create blockers.
12. Collect advisory preparation using the exact59 rows in the normative
    catalog-preparation-relations reference, plus invitation-prepare for every
    selected permanent instance. Empty selection returns no guidance. Each guide
    appears once, guides and distinct reasons are sorted by exact byte order,
    and every causing permanent ID is retained. Read each guide using the real
    catalog reader and require preparation-guidance carrier. Return its exact
    Markdown bytes. Do not infer links from prose, variables, titles or goals.
    Welcome retains both office and remote advice without selecting personal mode.

Successful Result.Plan has SchemaVersion1, exact PlanID and Provenance, copied
Directory and PlanVariables, canonical-ID Sections, and only explicitly bound
Answers. Origin is exactly zero. URL/ResultsPath are empty; UnknownFields is nil.
No axis variables, default sections, answer defaults, Source metadata in unknown
fields, selection manifest, recipient, event or outcome is invented. Creation
does not mutate an existing plan and is not a save. Later reviewed editing,
identity reservations/retirement and independent outcomes remain roadmap work.

New result collection shape is exact: Sections, Variables, all three Directory
lists, Answers.Plan, Answers.Sections and Guidance are allocated empty when they
have no entries. Section answer scopes exist only for instances with at least
one explicit binding. Copied nested held data retain nil/empty shape, ordering,
repetitions and all unselected Value fields. Normalize only those outer newly
created containers. Every returned mutable object is independent of the request,
other instances from the same source and earlier/later results. Diagnostics'
Detail maps and Markdown buffers are owned too.

## Bounded caller admission

These local caps constrain the new creation operation only. They do
not alter codecs, the legacy instantiator, or invitation0390 admission and do
not claim a heap, elapsed-time, serialized-output or future-invitation guarantee.
All bounds are inclusive, checked before expansion; nonnegative sums saturate
at bound+1. Count repeated occurrences separately, not deduplicated values.

| Held request counter | Maximum |
| -- | -- |
| Instances | 256 |
| Plan variable declarations | 1024 |
| All explicit bindings, including plan and every instance | 8192 |
| Each directory Contacts/Roles/Teams; their combined count | 2048;4096 |
| Each role/team holder list; all holder occurrences | 1024;8192 |
| Each team Holders map; all team Holders map entries | 2048;8192 |
| Each options list; all options | 1024;16384 |
| Each Value.List; all Value.List elements | 4096;32768 |
| Each condition All; all clauses | 256;4096 |
| Each IsOneOf/IsNotOneOf; their combined elements | 256;16384 |
| All known collection entries combined | 65536 |
| Each text field, map key or list item, including ignored payloads | 1048576 bytes |
| PlanID, SourceID and PermanentSectionID individually | 4096 bytes |
| All held text occurrences combined | 8388608 bytes |

Traversal enumerates every string and collection in Request, not a serialized
copy. Count PlanID; Provenance.TemplateSet/GeneratedOn; Directory contacts'
ID/Name/Email/Slack/SlackID, roles' ID/Name/Holders, teams' ID/Name/Contact and
Holders map keys/values; every PlanVariable Name/Description/Example/Type/Options/
Pattern/PatternHint/Audience, Default and AskIf; every binding Name and Value;
each instance SourceID/PermanentSectionID and its bindings. Each Value always
counts Text and every List item independent of Kind. Each clause counts Question,
Section and every Value in Answer.Is/IsNot/IsOneOf/IsNotOneOf; optional bools
contribute no text. No map keys or collection elements disappear from admission.
Unknown ValueKind refuses at the shape phase. Invalid UTF8 anywhere refuses,
including inactive text. Strings are not normalized. No unknown-data root exists.
The combined known-entry counter adds the lengths of every listed request
collection: Instances, PlanVariables, PlanBindings, instance Bindings, all three
directory lists, role Holders, each team Holders map and its value lists,
variable Options, each Value.List, condition All and comparison-value lists.
It counts no scalar fields or map capacity. Map keys also count as text, but
empty team holder lists cannot bypass the map-entry and aggregate-entry guards.

All collection limits precede any string limit/UTF8/semantic failure. A collection
failure returns limit at request. Then field/ID and aggregate byte limits return
limit at request. Then any invalid UTF8 returns text at request. Known-kind
checking follows those guards, plan variable Default before condition values,
plan bindings then instance bindings, preserving held list order. Team holder
maps are traversed by exact key byte order. Own request limits never allocate
predicted output. Current compiled catalog and emitted guide size are the
reader's immutable bounded content, not caller-controlled alternate templates.

## Diagnostics

Every own problem is Severity error, Question/Section empty and Detail exactly
`{"path":Path}`. No rejected IDs, source selector, answer, directory value,
pattern, secret or parser excerpt appears in own Message or Detail. Indices are
zero-based decimal; Path labels use the API fields below in lower camel case.
Own Stage is request, source, profile, binding or guidance as appropriate.

| Code | Stage | Message | Path |
| -- | -- | -- | -- |
| materialize.limit | request | Reduce the selection or held creation data to the supported limits. | request |
| materialize.text | request | Use valid UTF8 creation data. | request |
| materialize.value-kind | request | Choose a supported value kind. | exact Value path |
| materialize.plan-id | request | Supply an explicit nonempty plan identity. | planId |
| materialize.instance-id | request | Supply a nonreserved lowercase permanent section ID. | instances[I].permanentSectionId |
| materialize.duplicate-instance | request | Give each independent activity its own permanent section ID. | instances[I].permanentSectionId |
| materialize.duplicate-binding | request | Supply each scoped binding name once. | planBindings[J].name or instances[I].bindings[J].name |
| materialize.carrier | source | Select a section-template source for an activity. | instances[I].sourceId |
| materialize.source-identity | profile | Repair the catalog definition to match its declared source identity. | instances[I].sourceId |
| materialize.profile | profile | Select a graph-free definition until explicit instance mapping is supported. | instances[I].sourceId |
| materialize.binding-name | binding | Bind an exact variable declared in this scope. | binding Name path |
| materialize.binding-owner | binding | Keep newcomer answers outside the shared organization plan. | binding Name path |
| materialize.guide-carrier | guidance | Repair the catalog preparation guide classification. | guidance |

Source reader problems are wrapped Stage source and the selected sourceId Path;
loader problems are Stage load with the same Path. Delegated authoring,
unbound-plan and final-plan failures have Stage authoring, structure and plan,
with empty wrapper Path. Answer problems have Stage answer and the binding
Value path. Guidance reader problems have Stage guidance, Path guidance.
Preserve each actual Problem.Code/Severity/Message/Question/Section/Detail and
their ordered multiplicity; return independent Detail storage. Do not manufacture
missing-answer diagnostics or reinterpret a warning as completion.

## Independent oracle and hardening boundary

`testdata/relations.tsv` literally enumerates all59 source/carrier positives,
their complete expected common/additional guide unions and both repeated-instance
reason IDs. `unions.json` fixes mixed-source ordering, sharing, remaining causes,
welcome conditional advice and deliberate deploy direction. Expectations are
not discovered from runtime inventory or reconstructed from production relations.

`all-guidance.tsv` independently hand-lists the complete nineteen-guide union
and every causing first permanent ID when all59sources are selected together.
It is not calculated from the per-source relation fixture during the test.

`repository-a.canonical` and `repository-b.canonical` are hand-authored final-ID
canonical JSON texts under the unchanged hasher contract. SHA256 over those
literal bytes, excluding the fixture's final LF, is an independent standard
primitive oracle. No product hasher, encoder or loader generates expected values.
`repository.body.txt` is literal unresolved source text with its final LF.
Tests compare the full expected Section and plan/answer shapes, not only hashes.

Future graph-bearing compiled sources, source-ID mismatches and broken guide
carriers are unreachable through today's public request: caller injection is
deliberately unavailable. Source review must check those placement requirements;
a separately changed frozen catalog and corresponding independent fixture are
needed to exercise such a future source. Current compiled local variables are
all onboarder. A supplied onboardee plan declaration is unused under the
current graph-free sources and the actual authoring validator refuses it first;
the later section binding-owner diagnostic needs a valid future source with a
used onboardee local question to become reachable. That defensive ownership
requirement is specified, but its own error code is not claimed reached today.
Do not claim coverage from request
fields that cannot supply it. No runtime testing hook is authorized.

The constructor-only binding names the actual New constructor. Root supplies
an independently maintained failing component stub before an independent build.
Every semantic test requires
nonzero output or exact error, so a zero stub cannot pass by returning no problems.
Input immutability, stateless concurrent reuse and returned-storage isolation are
also judged. Authoring the tests does not authorize this author to build the unit.

| Plausible wrong implementation | Independent rejection boundary |
| -- | -- |
| Restrict repetition to three special sources or silently select all sources | All59 repeated positives and empty-selection exact plan |
| Derive IDs from request position, source or URL | Explicit-ID reverse/add/remove controls and duplicate-instance refusal |
| Hash original ID, attach instance ID as Source.Template or include answers | Literal repo-a/repo-b canonical hashes and binding-independent baseline |
| Resolve body, change step IDs or flatten repeated answer scopes | Full repository Section oracle and exact isolated Answer maps |
| Fill defaults/examples, copy a same-name answer or store newcomer binding | Unbound default control, two-instance single binding and ownership refusals |
| Accept an undeclared binding because Validate would warn | Exact own unknown-name refusal before answer validation |
| Accept duplicate binding/declaration by choosing first/last | Both-order duplicate and delegated declaration ambiguity controls |
| Use Validate's unknown-directory warning for a newly submitted binding | Actual CheckValue directory-error control |
| Drop shared preparation, infer all documentation is reading, reverse deploy roles | 59literal unions and mixed remaining-cause/deploy/welcome fixtures |
| Sort/rewrite held directory/variables or alias nested request values | Full plan equality and bidirectional ownership tests |
| Call plan validation before authoring or run only one validator | Invalid type/pattern before invalid answer, and duplicate directory controls |
| Truncate caps or ignore inactive Value payloads | Inclusive/exclusive instance/text/count guards and invalid inactive UTF8 |
| Accept examples without bindings or reject explicit answers equal to suggestions | TestPlanExamplesRemainUnacceptedWithoutBindings and TestExplicitPlanBindingsRemainDistinctFromExamplesAndDefaults |
| Reject a valid digit-started permanent identity | TestDigitStartedPermanentSectionIDRemainsExplicitAndValid independently changes only the literal canonical ID before SHA256 |
| Return only the first problem from a failing dependency call | TestAuthoringRefusalPreservesTheCompleteOrderedBatch and TestBindingRefusalPreservesEveryMissingContactItem |
| Collapse nested empty collections or share default backing storage | TestNestedAllocatedEmptyCollectionsKeepTheirHeldShapes and TestVariableDefaultListsDoNotAliasCallerOrOtherResults |
| Tighten a team-holder or identity cap, skip held text validation or report the last bad value kind | admission_gaps_test.go isolates exact inclusive/exclusive guards, inactive lists, map keys and earliest traversal paths |

Wrong-control, stub, mutation and independent-build results belong in review
evidence outside this contract.
