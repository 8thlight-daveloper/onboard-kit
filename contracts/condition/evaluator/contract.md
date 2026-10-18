# Condition evaluator contract

Status: draft, for a person to approve. Component: `condition/evaluator`. Issue: [#61](https://github.com/8thlight-daveloper/onboard-kit/issues/61).

The condition evaluator decides whether a condition holds, given the answers it is passed. It is a pure function. It reads no file, keeps no state, and does not change its arguments. A condition is used for `askIf` and `includeIf`. See [conditions](../../../docs/reference/section-template-schema/conditions.md), the [domain model](../../../docs/reference/domain-model.md), and [0015](../../../docs/decisions/0015-functional-core-and-imperative-shell.md).

## Interface

`Holds(condition, scope, answers) -> bool`.

- `condition` is a Condition. The caller passes one that exists.
- `scope` is the scope the condition is written in: a section id, or `plan`. A bare question name is read in this scope.
- `answers` is an Answers value, as stored.

The result is `true` when the condition holds and `false` otherwise. It depends only on the condition, the scope, and the answers it is given. It is the same whether or not the questions the condition names are active, and whether or not the sections they belong to are included or enabled. There is no error and no problem list.

## Rules

**Finding the answer.** A question id is `<scope>.<name>` or a bare `<name>`. An id with a dot is split at the first dot. `plan.<name>` reads `answers.plan[name]`, and `<section>.<name>` reads `answers.sections[section][name]`. A bare name is read in `scope`, so it means `<scope>.<name>`. A bare name never falls back to another scope. A missing section, a missing name, and zero maps all mean the question has no answer. The evaluator never fails on an id.

**Real answer.** A question has a real answer when the answers hold a value for it that is real:

- Text is real when it has at least one character. The Text `unknown-yet` is a plain string and differs from Unknown, so it is real.
- A Flag, `true` or `false`, is always real.
- A List is real when it has at least one item that is non-empty Text. A list with no items and a list of only empty strings are not real.
- Unknown and a value of no kind are not real.

A declared default is not in the answers, so it never counts.

**Equal values.** Two values are equal when they have the same kind and the same content. Text matches exactly, with case. Text `true` and Flag `true` are different values, because different kinds are different values.

**Lists.** The docs define no test that compares lists, and `any` and `none` are deferred. Only `command` variables and `person` variables with `multiple: true` hold lists. An `includeIf` names plan variables only, and those hold one value each. So list tests are not supported yet, and a later operator will define them. Until then, every test except `isAnswered` is false when the answer is a List, or when the test value is a List, or when any item of `isOneOf` or `isNotOneOf` is a List. `isAnswered` still applies to a List answer.

**Tests.** Every test except `isAnswered` is false when the question has no real answer, before any negation. When it has one, and no List is involved:

| Test | True when |
| -- | -- |
| `is` | The answer equals the value |
| `isNot` | The answer does not equal the value |
| `isOneOf` | The answer equals at least one value of the list |
| `isNotOneOf` | The answer equals no value of the list |
| `isAnswered` | `true`: the answer is real. `false`: it is not, so it is missing, Unknown, or empty |

A test is set when its field is present. A one-item `isOneOf` is evaluated as written. Several tests in one answer must all hold.

**Condition.** A condition holds when every clause holds. A condition with no clauses holds, as no `includeIf` holds. A clause holds when all of its tests hold.

**No reports.** The evaluator does not report a condition that cannot do its job, such as a contradiction, a one-item list, or a reference to an unknown question. The validators and the linter report those. The evaluator evaluates the condition as written.

**Active questions.** The evaluator reads whatever answer is stored for a question id. Which questions are active is the caller's business, and the caller decides which answers to pass.

## Decided by a person

- The caller decides which questions are active. The evaluator does not.
- Text compares with case.
- List tests are not supported yet.

## Open questions

None.

## Properties

These hold for every input the contract tests use.

- **Pure.** The condition and the answers are the same after the call as before.
- **No memory.** The same call gives the same result on a shared evaluator, whatever calls came before, forward or in reverse order.
- **Zero answers.** With the zero Answers, every test except `isAnswered: false` is false.

## Not covered

- Text of only whitespace. Answers are trimmed and blank text is reported before the evaluator sees it, so whether it is real is not tested.
- An answer with no test set, or a test value of no kind or Unknown. A valid plan never reaches these, because the validators report them, so the evaluator is not asked to give them a meaning.
- A malformed id, such as an empty one, one with a dot at the start or end, or one with two dots. Reading it as no answer is the only safe behavior. Only an unknown section is tested.
- `any` and `none`, and any test that compares lists.
- Inactive sections, `enabled`, and ordering.

## Wrong implementations tried

While drafting, a reference evaluator was written from this page alone as scratch work outside the repo, and it passed every test. Then 33 wrong evaluators were written, one defect each, and the contract tests failed each. A stub that returns false failed every test that checks a result.

| Wrong implementation | Caught by |
| -- | -- |
| `isNot` or `isNotOneOf` true when the question is unanswered | `missing-answer`, `unknown-answer`, `empty-answers` |
| `isOneOf` matched as a prefix or a substring | `one-of-lists` |
| `isOneOf` or `isNotOneOf` reads only the first item | `one-of-lists`, `false-flag` |
| A one-item `isOneOf` rejected | `one-of-lists` |
| `isAnswered: false` ignored | `false-flag`, `all-clauses`, `tests-in-one-answer` |
| `isAnswered` true for any stored value, so Unknown and empty count | `unknown-answer`, `empty-answers` |
| Unknown treated as an answer | `unknown-answer` |
| Empty text or an empty list treated as an answer | `empty-answers` |
| A list of only empty strings treated as an answer | `empty-answers` |
| A Flag `false` treated as no answer | `false-flag` |
| Flags equal whatever their value | `false-flag` |
| Text `true` equal to Flag `true` | `value-kinds`, `false-flag` |
| A list equal to a list | `list-answer`, `list-test-values` |
| A list matched by one of its items | `list-answer`, `list-test-values` |
| A one-item list equal to its Text | `list-answer`, `list-test-values` |
| A list test value ignored when another item of `isOneOf` matches | `list-test-values` |
| `isAnswered` rejects a list answer | `list-answer`, `empty-answers` |
| Text compared without case | `is-and-is-not` |
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
