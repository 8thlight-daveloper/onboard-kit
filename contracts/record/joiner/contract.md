# Record joiner contract

Status: draft by an agent. Not reviewed by a person. Component: `record/joiner`. Issue: none yet.

The record joiner puts a newcomer's onboarding record over the plan and gives back one plan. Every other component takes that one plan, so none of them needs to know about records. It is a pure function. It reads no file, keeps no state, and does not change its arguments. See [0090](../../../docs/decisions/0090-plan-and-onboarding-record.md), [0075](../../../docs/decisions/0075-refreshing-a-personal-plan.md), the [domain model](../../../docs/reference/domain-model.md), and [0015](../../../docs/decisions/0015-functional-core-and-imperative-shell.md).

## Interface

`Join(plan, record) -> plan`.

- `plan` is the plan copy, as the driver read it.
- `record` is the newcomer's record, as the driver read it.
- The result is a new Plan. There is no problem list, and the function never fails.

## Rules

**Answers.** The result holds the plan's answers with the record's answers laid over them, in the plan scope and in every section scope.

- A question the record answers takes the record's value, whatever the plan held for it.
- A question only the plan answers keeps the plan's value. A pre-seed the newcomer never answered therefore follows the plan.
- A question only the record answers is added.
- A section scope the record has and the plan does not is added, with the record's answers.
- The plan scope and a section scope are separate. The same name in two scopes is two questions. The record's answer in one scope never touches another.

**A record answer is laid over whatever it is.** A Text, a Flag, a List, and Deferred are laid over alike. Whether the answer is unanswered, deferred, or settled under [0035](../../../docs/decisions/0035-settled-and-deferred-answers.md) does not matter. So the Text `""`, the Flag `false`, an empty List, and Deferred each replace what the plan held, and so does the Text `deferred`. The presence of the key decides. A List replaces the plan's List as a whole. It is never joined to it.

**Hidden answers.** An answer for a section or a question that the plan does not declare is kept, whether it came from the plan or from the record. The joiner does not look at the plan's variables or sections to decide what to keep.

**The joiner does not look at the audience.** A record answer is laid over even when the plan gives that question to the onboarder.

**Absent stays absent.** The result has a map only when the plan has it or the record gives it a key.

- A plan with no answers and a record with none give a plan with no answers.
- A plan with no plan map keeps none when the record has no plan answers. A plan with no section map keeps none when the record has no section answers.
- An empty map in the record adds nothing: no plan map, no section map, and no section for a section id whose map is empty.
- An empty map the plan holds comes back as an empty map. The record's answers are laid into it.

**Everything else comes back as given.** The `schemaVersion`, `id`, `provenance`, `origin`, `planUrl`, `resultsPath`, `directory`, `variables`, `sections`, and `unknownFields` of the result equal the plan's, with nothing added, dropped, reordered, or normalized. A nil list or map stays nil, and an empty one stays empty.

**Only the answers of the record are used.** The account, start date, plan reference, outcomes, and unknown fields of the record do not appear in the result. The plan reference is not compared with the plan's `id`. A Plan has no place for the rest.

**An empty record.** A record with no answers gives a plan equal to the plan given.

**Nothing is shared.** The result holds its own maps, lists, pointers, and values, at every depth. That covers the answers, the variables and their defaults, every condition and its answer tests, the sections, steps, checks, and escalation lists, and the unknown fields. It shares none with the plan or with the record. A caller can change anything in the result without changing either argument, and change anything in either argument without changing the result.

## Properties

These hold for every plan and record the contract tests use.

- **Pure and deterministic.** Two calls with the same arguments give equal plans, and nothing is read from or written to anything outside the arguments.
- **Arguments unchanged.** The plan and the record are equal to what they were before the call.
- **Idempotent.** Joining the result with the same record gives a plan equal to the result.
- **Nothing shared.** Changing any map, list, pointer target, or value of the result changes neither argument, and changing any of those in either argument changes neither the result nor the other argument.

## Not covered

- Taking the answers back out, which is the splitter, and checking a record against a new plan, which is the reconciler.
- Reading and writing plan and record files, which are the decoders, the encoders, and the driver.
- Deciding which questions are active, or which answers a person may see. The conditions, the audience, and the callers do that.
- A record answer that holds a Value of no kind. The decoder never gives one.
- A record that belongs to another plan. The driver decides that from the plan reference.

## Decisions to know

1. **hard.** The key decides, not the state of the answer. A record answer that is unanswered (empty text, an empty list) or deferred replaces a settled pre-seed in the plan. The docs say the record's answer wins and do not say what an empty one does. The other choice, to skip a record answer that is not settled, would let a pre-seed the newcomer cleared or put off come back.
2. **hard.** The joiner does not filter by audience. A record answer to a question the plan gives to the onboarder is laid over like any other. The docs say the record holds only answers to onboardee questions, which a decoder or validator could judge, and say nothing about the joiner skipping the rest. Filtering would hide a hand-edited answer and make the joined plan depend on the plan's variables.
3. **cheap.** A List answer replaces the plan's List as a whole. The docs say the record's answer wins and do not mention lists.
4. **cheap.** Nothing is created from nothing. A record with empty maps adds no map and no section, and the plan's own empty maps stay. The docs say a record holds only answers the newcomer gave, and a map with no answers is not one.
5. **cheap.** The plan reference of the record is not used and not checked against the plan. The docs say the reference tells which plan the record was last reconciled against, which is for the reconciler and the driver.
6. **cheap.** The result is a deep copy. The docs say the join is a pure function, and a caller who changes the joined plan must not change the plan copy or the record.

## Wrong implementations tried

Each row is a plausible wrong implementation and the cases that fail it.

| Wrong implementation | Caught by |
| -- | -- |
| A stub that returns nothing, or ignores the record | Every case, and the empty record test |
| The plan wins on the same key | `record-wins`, `value-kinds`, `scopes-are-separate`, `ada` |
| The record's plan map replaces the plan's | `hidden`, `one-side`, and the empty record test |
| The record's section map replaces the plan's | `record-wins`, `one-side`, `hidden`, `value-kinds` |
| A record answer that is empty text, an empty list, `false`, or Deferred is skipped (each tried on its own) | `value-kinds` |
| Lists joined instead of replaced | `value-kinds`, and the joined twice test |
| A record answer to an onboarder question skipped | `record-wins` |
| Record answers for an undeclared question or section dropped | `hidden`, `one-side`, `value-kinds` |
| Plan answers for an undeclared question or section dropped | `empty-record`, `other-fields`, `hidden`, `empty-plan-maps` |
| Section answers written into the plan scope, or a plan answer written into a section | `scopes-are-separate`, `section-answers-only`, `record-wins`, `ada` |
| The record's unknown fields added to the plan's | `other-fields` and the answers-only test |
| The `id`, `planUrl`, or origin hash taken from the record's plan reference | `other-fields` and the answers-only test |
| The origin or the unknown fields of the plan dropped | `empty-record`, `other-fields` |
| An empty map made when both sides hold none | `no-answers-anywhere`, `empty-record-maps-add-nothing`, `plan-answers-only`, `section-answers-only` |
| An empty section made from the record's empty section map | `empty-plan-maps`, `empty-record-maps-add-nothing` |
| An empty map of the plan turned to nil, or its empty section dropped | `plan-empty-map-stays`, `empty-plan-maps` |
| The record's maps or lists shared with the result | the nothing shared test, the changing the result test |
| The plan copied shallowly, so the sections, unknown fields, or answer lists are shared | the nothing shared test, the changing the result test, the empty record test |
| One nested part of the plan left uncopied, such as a variable default, an answer test of a condition, a step's escalation list, or a flag pointer | `nested-values` with the nothing shared test and the two changing anything tests |
| An argument changed in place | the arguments not changed test, the changing the result test |
| An empty list laid over as a nil list | `value-kinds` |

## Bootstrap preservation

[0250](../../../docs/decisions/0250-bootstrap-content-and-projection.md) supplies
the source for the independently specified [preservation tests](bootstrap_test.go).
Layering record answers preserves every section's Bootstrap definition value. No record answer changes that marker, and the deep-copied returned sections do not alias input section slots.
