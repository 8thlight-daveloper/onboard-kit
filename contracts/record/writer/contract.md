# Record answer writer

## 1. Purpose, reads and writes

`record/writer` applies one explicit newcomer answer intent to a held record.
Pure deterministic stateless core, position6. Reads the shared Plan, original
Record and Intent. Writes only a returned Record's selected answer key and the
minimal answer containers necessary for that change. No file, environment,
clock, network, process, execution, registry, terminal or serialization effects.
Never mutates an argument or an earlier result.

Under0090 a never-given answer follows an authorized shared preseed; an
explicitly given answer belongs to the newcomer even when equal to that
preseed. Under0120 Default/Example remain suggestions. `submit` and
`follow-plan` are different explicit intents. This does not change generic
record splitter semantics, invent a previous record, reconcile a session, or
introduce a record schema/tombstone. The caller owns identity, consent, complete
session validation/admission, visibility, and native publication.

## 2. API and exact behavior

The complete model-only interface is api.go. `New()` has no arguments and
constructs the actual record joiner and actual plan editor. Success returns a
separately owned full Record and actual ordered warnings. Any error returns
the zero Record and a separately owned ordered Problem batch. Never return a
partially changed record as a usable error result. Nil/empty Problems both mean
no problems; all other model nil/empty shapes follow the rules below.

The caller supplies bounded acyclic model data of the public representable
unknown-value kinds, valid UTF8, and an authoring-valid current shared plan.
This writer does not establish session identity, validate old record history,
prune hidden answers or judge unrelated stored answers. It defensively specifies
the target lookup failures below even if those particular declarations are
malformed. Invalid unknown object kinds/cycles, unsupported proposed ValueKind
or plan enums (apart from the specified audience guard), and unbounded
in-memory resources are not a new serialization/admission contract.

Stop at the first failing phase, in this exact order:

1. Operation must exactly equal submit or follow-plan. Otherwise return one
   own operation error. For follow-plan the supplied Value must DeepEqual the
   all-zero `model.Value{}`, including nil List; otherwise one own follow-value
   error. Never silently discard supplied active or inactive Value fields.
2. Resolve the exact Question at the first literal dot: scope plan means only
   Plan.Variables, every other scope means exactly one Section with that exact
   ID. Within that scope exactly one Variable must have the exact remaining
   name. Missing dot, missing/ambiguous scope or variable, or malformed empty
   scope/name gives one own question error. No trimming, case folding,
   lexicographic scope classification, or lookup in a different section.
   The variable's Audience must be exactly onboardee; empty audience defaults
   to onboarder and is refused, as is explicit onboarder or any other value.
   This declaration lookup does not grant activity permission.
3. For submit, reject the editor's unanswered cases: ValueKindNone, Text with
   empty Text, List with no items or only empty strings. Other fields do not
   turn those into settled values. Whitespace is text, false is settled, and
   Deferred is explicit. Return one own empty error. There is no clear-to-blank
   operation; instruct follow-plan or Deferred instead.
4. Join the ORIGINAL shared Plan and ORIGINAL Record with the actual joiner.
   Call actual editor Submit on that effective plan, exact Question and actual
   submitted Value (or zero Value for follow-plan). Return the complete actual
   Problems unchanged in order/fields. Refuse only if any has severity error;
   warnings alone accompany the successful write. Never shadow CheckValue,
   finder activity, directory/pattern/options/secret checks, or Join semantics.
5. On success, submit copies the actual editor's accepted target Value into the
   exact target record key. Presence is explicit even if equal to the shared
   preseed, Default, Example, or already-held value. No other effective-plan
   answers enter the record. Follow-plan deletes only that exact record key;
   a missing key is unchanged-data success. If deletion of a present section
   key empties its section map, remove that section entry. Never delete an
   originally empty OTHER section map. Retain the enclosing Answers.Plan and
   Answers.Sections nil/allocated-empty shape unless allocation is necessary
   to add the selected key. Removing the last key does not nil an allocated
   outer map. A missing-key follow-plan does not allocate or prune anything.

Every own problem has Severity error, Question equal to Intent.Question,
empty Section and exactly one Detail entry with key path:

| Code | Message | Detail.path |
| -- | -- | -- |
| record.answer.operation | Unsupported answer operation. | intent.operation |
| record.answer.follow-value | Follow-plan must not carry an answer. | intent.value |
| record.answer.question | Choose one declared newcomer question. | intent.question |
| record.answer.audience | Only newcomer answers can be changed here. | intent.question |
| record.answer.empty | Choose an answer, defer, or explicitly follow the plan. | intent.value |

The record outside this selected answer change is EXACT: SchemaVersion,
Account, StartDate, FirstOpenedAt, Plan reference including Hash/URL, all
outcomes including their original order/duplicates/unknown kinds/old timestamp
text, hidden/empty/deferred/non-target answers, unknown maps/lists and their
nil/empty shapes. This component does not call Walker or Recorder, so their
successful-input preconditions do not authorize rejecting or rewriting history
here. Declared onboarder-held keys are preserved but never editable here;
the higher session reviewer separately rejects ownership-invalid sessions.

Follow-plan reveals the shared stored answer when later joined, or no answer
if none is shared. It never promotes a default or example. It cannot promise
unanswered while a shared preseed exists. Submit Deferred is the available
explicit nonsettled override. No parent map cleanup beyond the stated slot rule.

All returned mutable maps, slices, pointers and supported unknown-value trees
belong to the caller, not any input or earlier result, including successful
no-ops. Preserve allocated-empty shape; copying through JSON/YAML is forbidden
as a shape-preserving substitute. Returned Problem Detail maps likewise own
their storage. Concurrent calls share no mutable state.

## 3. Dependencies and architecture

Position6 depends on real record/joiner1 and real plan/editor5, model0 and its
own contract only. No splitter, reconciler, codecs, Walker, driver, clock or
synthetic dependency. The editor supplies actual finder and validator behavior.
The session reviewer7 composes this writer with the outcome recorder5 and
other current pure cores. Driver8 owns I/O, binary9 owns routing. Do not drop
a semantic dependency to retain an old entrypoint position.

## 4. Independent fixtures and wrong-variant verification plan

writer_test.go supplies literal model fixtures; wire_test.go alone imports the
proposed implementation and performs its one constructor call. wire_stub_test.go
under contractstub substitutes a returning-nothing test stub without an
implementation import. Tests use fresh factories and DeepEqual, never a product
encoder, Split, writer output or generated snapshot as the expected oracle.
Public dependency messages are useful/nonblank, not frozen wording; their
Code/Severity/Question/Section/Detail/order are checked independently.

| Wrong implementation to test separately | Distinguishing fixture |
| -- | -- |
| omit new equal preseed as generic diff | EqualPreseedBecomesOwned |
| import unrelated shared answers | EqualPreseedBecomesOwned full record |
| follow-plan means empty Text/Deferred, or removes other scopes | FollowPlanRemovesOnlyExactOverride |
| submit accepts blank as clear | UnansweredNeedsExplicitIntent |
| default/example becomes accepted during unrelated write | SuggestionsNeverBecomeAnswers |
| first/lexical scope selected | ExactScopesAndAudiences |
| skip actual activity/CheckValue or keep only first error | DelegatedRefusals |
| check activity against shared plan, not actual joined answers | ActivityUsesHeldRecordAnswer |
| prune hidden/empty data or sort outcomes | PreservationAndOwnership |
| share maps/list backing arrays/unknown tree or mutable state | PreservationAndOwnership |
| return original usable record on refusal | every refusal fixture |

No wrong variant was executed for this draft. The separate provenance specifies
the authorized future stub/baseline/one-defect verification sequence. No mutant
kill, pass, second build or differential result is claimed by this contract.
