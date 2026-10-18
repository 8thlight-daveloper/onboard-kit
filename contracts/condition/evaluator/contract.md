# Condition evaluator contract

Status: draft, for a person to approve. Component: `condition/evaluator`. Issue: [#61](https://github.com/8thlight-daveloper/onboard-kit/issues/61).

The condition evaluator decides whether one condition holds, given a plan's answers. It is a pure function. It reads no file, keeps no state, and does not change its arguments. The finder uses it for `askIf` and the selector for `includeIf`. See [conditions](../../../docs/reference/section-template-schema/conditions.md), the [domain model](../../../docs/reference/domain-model.md), and [0015](../../../docs/decisions/0015-functional-core-and-imperative-shell.md).

## Interface

`Holds(condition, scope, answers) -> bool`.

- `condition` is a Condition. The caller passes a condition that exists. A section with no `includeIf` is the selector's case, not the evaluator's.
- `scope` is the scope the condition is written in: a section id for a section's `askIf`, or `plan` for a plan variable's `askIf`. The selector passes `plan` for `includeIf`, which names plan variables only. The evaluator reads a bare question name in this scope.
- `answers` is the plan's Answers, as stored.

The result is `true` when the condition holds and `false` otherwise. There is no error and no problem list, see Rules.

## Rules

**Finding the answer.** A question id is `<scope>.<name>` or a bare `<name>`. An id with a dot is split at the first dot. `plan.<name>` reads `answers.plan[name]`, and `<section>.<name>` reads `answers.sections[section][name]`. A bare name is read in `scope`, so it means `<scope>.<name>`. A bare name never falls back to another scope. A missing section, a missing name, and zero maps all mean the question has no answer. The evaluator never fails on an id.

**Real answer.** A question has a real answer when the answers hold a value for it that is not Unknown and not empty. Empty means a value of no kind, Text with no characters, or a List with no items. A Flag, `true` or `false`, is always real. Text is real when it has a character, so a space is real. The Text `unknown-yet` is a plain string and differs from Unknown, so it is real. A declared default is not in the answers, so it never counts.

**Equal values.** Two values are equal when they have the same kind and the same content. Text matches exactly, with case. A List matches when it has the same items in the same order. Text `true` and Flag `true` are different. A one-item List and the Text of that item are different.

**Tests.** Every test except `isAnswered` is false when the question has no real answer, before any negation. When it has one:

| Test | True when |
| -- | -- |
| `is` | The answer equals the value |
| `isNot` | The answer does not equal the value |
| `isOneOf` | The answer equals at least one value of the list |
| `isNotOneOf` | The answer equals no value of the list |
| `isAnswered` | `true`: the answer is real. `false`: it is not, so it is missing, Unknown, or empty |

A test is set when its field is present. A one-item `isOneOf` is evaluated as written. Several tests in one answer must all hold.

**Condition.** A condition holds when every clause holds. A condition with no clauses holds. A clause holds when all of its tests hold.

**No reports.** The evaluator does not report a condition that cannot do its job, such as a contradiction, a one-item list, or a reference to an unknown question. The validators and the linter report those. The evaluator evaluates the condition as written.

**Active questions.** The evaluator does not know which questions are active. It reads the stored answer for the id. The finder decides what an inactive question does to a condition that names it, and passes answers accordingly. See the open questions.

## Open questions

These are choices the docs do not settle. The contract takes one reading and the fixtures pin it. A person should confirm each.

1. **Active questions.** conditions.md defines a real answer as one for a question that "is active and the plan stores an answer". The component row says the evaluator works "given the real answers of the questions it names". Readings: the evaluator checks activity, or its caller does. Taken: the caller does, because the evaluator gets answers and no variables. Consequence: the evaluator counts a stored answer of an inactive question as real. The finder, which knows which questions are active, removes that answer before the call. **Decided by a person:** the caller decides, so the evaluator stays a small pure function and the orchestrating component owns the knowledge of what is active. The finder contract carries this rule.
2. **Scope of a bare name in a plan variable's `askIf`.** "A bare name is a question in the same section" does not say what a plan variable's bare name means. Taken: the scope is `plan`, as in the Question type, so a bare name reads the plan answers.
3. **List equality.** "The answer is this one value" does not say whether a List is compared in order. Taken: same items, same order.
4. **Empty List.** "Not empty" is stated for the answer. Taken: a List with no items is empty, like empty Text.
5. **Case.** The docs do not say whether Text compares with case. Taken: exact, with case.
6. **Empty `all`.** The docs do not say. Readings: it holds, as an empty conjunction and as no `includeIf` is always included, or it does not hold, as nothing is answered. Taken: it holds.
7. **An answer with no test set.** The docs do not say whether the clause holds. It is not covered, see Not covered.
8. **Kinds.** The docs do not say whether Text `true` equals Flag `true`. Taken: different, because Value variants are different kinds. The lint rule "values fit the question" keeps this from reaching the evaluator in a valid plan.

## Properties

These hold for every input the contract tests use.

- **Pure.** The condition and the answers are the same after the call as before.
- **No memory.** The same call gives the same result on a shared evaluator, whatever calls came before, forward or in reverse order.
- **Zero answers.** With the zero Answers, every test except `isAnswered: false` is false.

## Not covered

- Whether a question is active, and what the finder does with an inactive one. See open question 1.
- An answer with no test set, or a test value of no kind or Unknown. The evaluator is not asked to give these a meaning. The linter and the validator reject them.
- A malformed id, such as an empty one, one with a dot at the start or end, or one with two dots. Reading it as no answer is the only safe behavior. Only an unknown section is tested.
- `any` and `none`, which the docs defer.
- Inactive sections, `enabled`, and ordering, which are the selector's.

## Wrong implementations tried

While drafting, a reference evaluator was written from this page alone as scratch work outside the repo, and it passed every test. Then 30 wrong evaluators were written, one defect each, and the contract tests failed each. The first run found one gap: a list compared only by its first item. It got a fixture.

| Wrong implementation | Caught by |
| -- | -- |
| `isNot` true when the question is unanswered | `missing-answer`, `unknown-answer`, `empty-answers` |
| `isNotOneOf` true when the question is unanswered | `missing-answer`, `unknown-answer`, `empty-answers` |
| `isOneOf` matched as a prefix or a substring | `one-of-lists` |
| `isOneOf` or `isNotOneOf` reads only the first item | `one-of-lists`, `list-answer`, `false-flag` |
| A one-item `isOneOf` rejected | `one-of-lists` |
| `isAnswered: false` ignored | `false-flag`, `all-clauses`, `tests-in-one-answer` |
| `isAnswered` true for any stored value, so Unknown and empty count | `unknown-answer`, `empty-answers` |
| Unknown treated as an answer | `unknown-answer` |
| Empty text or an empty list treated as an answer | `empty-answers` |
| A Flag `false` treated as no answer | `false-flag` |
| Flags equal whatever their value | `false-flag` |
| Text `true` equal to Flag `true` | `value-kinds`, `false-flag` |
| A one-item List equal to its Text | `value-kinds` |
| A list compared by length, by first item, or without order | `list-answer` |
| Text compared without case | `is-and-is-not` |
| Text compared by substring | `one-of-lists` |
| An empty `all` false | `no-clauses` |
| `all` as any: one true clause is enough | `all-clauses` |
| Only the first clause read | `all-clauses` |
| Only `is` read when `is` and `isNot` are both set | `tests-in-one-answer` |
| A bare name read in the plan | `question-ids`, `all-clauses` |
| A bare name that falls back to the plan | `question-ids` |
| The prefix of a question id ignored | `question-ids` and most cases |
| A plan-scope bare name that reads a section | `plan-scope` |
| An argument changed: the answers, or an item of the condition | The inputs-unchanged test |
| A cache keyed by the question ids, so a later call returns an earlier result | The no-memory test, and most cases |
