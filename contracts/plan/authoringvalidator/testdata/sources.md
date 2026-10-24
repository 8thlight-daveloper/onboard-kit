# Literal fixture sources

All payloads are independently handwritten model JSON, not product decoder,
finder, evaluator, validator, renderer or encoder output. They contain only
invented `.example` routes and a deliberately incomplete private-key header.
That header is a secrets-contract shape, not a real credential. Numeric Value
kinds transcribe `internal/model/value.go`: None 0, Text 1, Flag 2, List 3,
Deferred 4. No YAML presence is claimed after conversion to this model boundary.

## Plan cases

| Cases in cases.json | Independent authority and expected distinction |
| -- | -- |
| empty, ignored-held-data | 0015/0020; scoped interface; unknown/answer history is not declaration input |
| valid-declarations | Variable fields/types, 0045/0120, 0080; missing required answers are blockers, not authoring errors |
| required-prose-slugs, variable-name-enums | 0040; domain Section/Variable; reserved-names reference; explicit indexed privacy rule |
| multiple-choice-pattern | Variables reference and existing plan-validator exclusions; type-specific options/pattern scope |
| directory-declarations | 0080 exact ID spelling; domain Directory required ID/name; draft's explicit nonblank display-name judgment |
| timing-all-declarations | Ordering reference and 0260; all declarations checked without punishing date goals |
| short-escaped-longest-dollars, dollar-fields | 0040; variable syntax; exact resolver filled-field table, not resolver error output |
| owners-escalation | 0080; escalation reference; draft's bounded literal-owner clarification |
| checks-environment | Steps reference; model Check's meaningful-operation boundary; no shell execution |
| condition-shapes | Conditions reference; evaluator exclusions; finder plan-variable AskIf authoring exclusion |
| ancestor-and-cross-ask-valid, include-scope | 0070 supersedes old plan-only row; ordering's later-section AskIf example |
| inactive-cross-cycle, inherited-eligibility-cycle | Conditions/ordering dependency-cycle rule; explicit eligibility graph, not active answer evaluation |
| duplicate-and-graph-delegation, duplicate-sections-suppress-reference-analysis, missing-parent-scope-suppression, parent-cycle-suppresses-scope-only | 0110; actual plan-validator/structure-validator ownership, no duplicate winner or new graph code |
| malformed-qualified-questions | Variables/evaluator first-dot scope rule; nonexistent exact declarations never use a prefix fallback |
| use-closure | 0055 explicitly requires question purpose; disabled uses and AskIf dependency closure are draft's precise rule |
| secret-projection, errors-do-not-hide-declaration-secrets | Actual secrets contract: private header alone suffices, list item indices, inactive/error declarations still stored |

## Default and predicate tables

`defaults.json` independently transcribes each declared default's shape and
expected type/options/pattern/directory rule. Empty and Deferred controls are
explicit draft judgments consistent with 0035/0045/0120, not accepted answers.
The pattern-search control comes from the existing plan-validator semantics.
Directory exact-case/namespace controls come from 0080.
Default secret cases use the secrets contract's fabricated URL-password and
long mixed-case/digit shapes. Inactive-payload controls come from that contract's
Kind-decides rule; the header in an unselected field must not be scanned.

`predicates.json` transcribes condition syntax and evaluator's valid-input
exclusions. Singleton/boolean/all-choice lists come from conditions' lint-error
catalogue. Scalar false, whitespace, unaccepted defaults and list IsSettled
controls preserve 0035/0040 semantics. The contradiction cases use literal finite
sets with visibly empty intersection or excluded singleton; no expected value
is computed by the condition evaluator. Self-switch conjunction tests use 0050's
switch-only meaning and do not invent an activity cycle.
Predicate directory controls use the exact 0080 namespace and 0110 no-winner
rule; pattern controls use authored RE2 search/anchors, not product output.

## Harness and evidence status

The harness compares ordered literal code/severity/path/question/section/detail
and exact local repair sentences. Secrets messages are judged only by their real
contract: nonempty, stable for a code and not exposing the supplied header/key;
no invented dependency English sentence is frozen. Properties cover deep input
preservation, independent result maps, repeated/reversed calls and concurrent
calls on one validator. The fixed nonzero test-only stub is intentionally wrong
for both empty valid plans and the literal bad declarations. It is not product
implementation. Root must independently run stub red, actual dependency checks,
wrong variants, two fresh builds and differential checks. None ran during draft
authorship; test plans are not verified safeguards.

`TestDeepShapesAndCallerHeldReferences` uses two separately constructed literal
model graphs, compared with reflect.DeepEqual rather than serialization. Its
unknown values distinguish int from float64, nil from empty typed slices/maps,
and preserve held condition/switch/source pointers, default/command backing
lists and maps. Mutating one result must not affect a simultaneously held second
result or the input. Caller edits after validation remain visible through the
original input graph. The one expected pattern error is literal and independent.

Fixture decoding rejects unrecognized struct fields and extra JSON values.
UnknownFields remains an intentional arbitrary-data map, not an exemption for
typos in model or expectation field names. JSON syntax was checked with jq only;
the parent must compile/run the strict loader before claiming every mapping has
been executed successfully.

## Indexed scope and purpose regressions

`differential-gaps.json` was independently handwritten from 0110, 0055, 0070,
the indexed diagnostic rules and bounded agent judgments under 0100/P3/P7.
Duplicate-own-scope cases distinguish target-dependent lookup from independent
clause shape and explicit plan scope. Invalid-ID cases preserve independent
question-purpose errors without echoing the rejected ID. Exact escalation use
counts purpose separately from supported wrong-type errors or unknown-type
suppression; malformed syntax is not use. Self inclusion remains forbidden
without consulting a malformed Parent chain. Bare references remain positional
under an invalid section called plan, while explicit plan prefixes name actual
plan variables. Switch cases retain unique-target contradictions and unknown
target diagnostics without selecting a duplicated target.

These are agent-reviewed diagnostic clarifications, not human approval or
implementation-derived expected output. They change neither ownership nor the
defensive contracts of finder/evaluator. `variant-gap-sources.md` separately
records the isolated existing-rule controls in `variant-gaps.json` and the
remaining direct dependency-message oracle limitation.

`second-differential-gaps.json` pins the priority between independently known
bare-local self inclusion under 0070 and global qualified-question suppression
under 0110. A unique child remains its own nonancestor even when its Parent ID is
duplicated. A qualified child reference still follows earlier global resolution
suppression; a duplicated own child scope suppresses both forms. These are
agent-reviewed diagnostic priorities under P3, not accepted invalid plans or
implementation-derived expectations. Structural duplicate errors stay with the
existing validator.
