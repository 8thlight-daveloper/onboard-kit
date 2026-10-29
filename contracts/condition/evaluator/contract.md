# Condition evaluator contract

Status: approved. Component: `condition/evaluator`.

The condition evaluator decides whether a condition holds, given the answers and the section switches it is passed. It is a pure function. It reads no file, keeps no state, and does not change its arguments. A condition is used for `askIf` and `includeIf`. See [conditions](../../../docs/reference/section-template-schema/conditions.md), the [domain model](../../../docs/reference/domain-model.md), [0015](../../../docs/decisions/0015-functional-core-and-imperative-shell.md), and [0050](../../../docs/decisions/0050-section-switch.md).

## Interface

`Holds(condition, scope, answers, switches) -> bool`.

- `condition` is a Condition. The caller passes one that exists.
- `scope` is the scope the condition is written in: a section id, or `plan`. A bare question name is read in this scope.
- `answers` is an Answers value, as stored.
- `switches` is a Switches value: a map from section id to whether that section is on. A section missing from the map is on. A nil map means every section is on.

The result is `true` when the condition holds and `false` otherwise. It depends only on the condition, the scope, the answers, and the switches it is given. It is the same whether or not the questions the condition names are active, and whether or not the sections they belong to are included or on. There is no error and no problem list.

## Rules

**Finding the answer.** A question id is `<scope>.<name>` or a bare `<name>`. An id with a dot is split at the first dot. `plan.<name>` reads `answers.plan[name]`, and `<section>.<name>` reads `answers.sections[section][name]`. A bare name is read in `scope`, so it means `<scope>.<name>`. A bare name never falls back to another scope. A missing section, a missing name, and zero maps all mean the question has no answer. The evaluator never fails on an id.

**Settled answer.** An answer is unanswered, deferred, or settled, as [0035](../../../docs/decisions/0035-settled-and-deferred-answers.md) defines. The evaluator treats only a settled answer as something a test can match, and `isSettled: false` holds when the answer is deferred or unanswered. A question has a settled answer when the answers hold a value for it that is settled:

- Text is settled when it has at least one character. The Text `deferred` is a plain string and differs from the Deferred kind, so it is settled.
- A Flag, `true` or `false`, is always settled.
- A List is settled when it has at least one item that is non-empty Text. A list with no items and a list of only empty strings are not settled.
- Deferred and a value of no kind are not settled.

A declared default is not in the answers, so it never counts.

**Equal values.** Two values are equal when they have the same kind and the same content. Text matches exactly, with case. Text `true` and Flag `true` are different values, because different kinds are different values.

**Lists.** The docs define no test that compares lists, and `any` and `none` are deferred. Only `command` variables and `contact` and `role` variables with `multiple: true` hold lists. An `includeIf` names plan variables only, and those hold one value each. So list tests are not supported yet, and a later operator will define them. Until then, every test except `isSettled` is false when the answer is a List, or when the test value is a List, or when any item of `isOneOf` or `isNotOneOf` is a List. `isSettled` still applies to a List answer.

**Section clause.** A clause has one of two forms. A question clause has `question` and `answer`. A section clause has `section` and `enabled`, and holds when the switch of the section named equals `enabled`. The switch of a section is `switches[section]`, and is on when the section is missing from the map. So an unknown section id is the same as a section that is on, and a nil map is the same as an empty one. The id is used as written. It is not read in `scope`, and a scope prefix is never added. A section clause reads only the switches. It never reads the answers, so an answer stored under the name `enabled` is an ordinary answer and does not set a switch. A question clause never reads the switches, so its result is the same whether the section it names is on or off. A clause with both kinds of key is not given to the evaluator: the decoder and the loader report it.

**Tests.** Every test except `isSettled` is false when the question has no settled answer, so when it is deferred or unanswered, before any negation. When it has one, and no List is involved:

| Test | True when |
| -- | -- |
| `is` | The answer equals the value |
| `isNot` | The answer does not equal the value |
| `isOneOf` | The answer equals at least one value of the list |
| `isNotOneOf` | The answer equals no value of the list |
| `isSettled` | `true`: the answer is settled. `false`: it is not, so it is missing, Deferred, or empty |

A test is set when its field is present. A test that is not set is not evaluated: it neither holds nor fails, and only the tests that are set decide the clause. A one-item `isOneOf` is evaluated as written. Several tests in one answer must all hold.

**Condition.** A condition holds when every clause holds, whichever form it has, and the clauses may be mixed in any order. A condition with no clauses holds, as no `includeIf` holds. A clause holds when all of its tests hold.

**No reports.** The evaluator does not report a condition that cannot do its job, such as a contradiction, a one-item list, or a reference to an unknown question. The validators and the linter report those. The evaluator evaluates the condition as written.

**Active questions.** The evaluator reads whatever answer is stored for a question id. Which questions are active is the caller's business, and the caller decides which answers to pass.

## Decided by a person

- The caller decides which questions are active. The evaluator does not.
- Text compares with case.
- A section clause tests the switch only, not whether the section's `includeIf` holds ([0050](../../../docs/decisions/0050-section-switch.md)).
- List tests are not supported yet.

## Open questions

None.

## Properties

These hold for every input the contract tests use.

- **Pure.** The condition, the answers, and the switches are the same after the call as before.
- **No memory.** The same call gives the same result on a shared evaluator, whatever calls came before, forward or in reverse order.
- **Zero answers.** With the zero Answers, every test except `isSettled: false` is false.
- **Nil switches.** With nil switches, every section is on, and a case with an empty map gives the same result as with nil.

## Not covered

- Text of only whitespace. Answers are trimmed and blank text is reported before the evaluator sees it, so whether it is settled is not tested.
- An answer with no test set, or a test value of no kind or Deferred. A valid plan never reaches these, because the validators report them, so the evaluator is not asked to give them a meaning.
- A malformed id, such as an empty one, one with a dot at the start or end, or one with two dots. Reading it as no answer is the only safe behavior. Only an unknown section is tested.
- `any` and `none`, and any test that compares lists.
- Inactive sections and ordering.
- A clause with both a question and a section, and a section clause that names no section id. The decoder, the loader, and the validators report them.
- Whether a section named by a section clause exists. The plan validator reports an unknown one.

## Wrong implementations tried

Each row names a plausible wrong implementation and the groups of cases that fail it. How the contract was checked, and with what scores, is recorded in the issue, not here.

| Wrong implementation | Caught by |
| -- | -- |
| `isNot` or `isNotOneOf` true when the question is unanswered | `missing-answer`, `deferred-answer`, `empty-answers` |
| `isOneOf` matched as a prefix or a substring | `one-of-lists` |
| `isOneOf` or `isNotOneOf` reads only the first item | `one-of-lists`, `false-flag` |
| A one-item `isOneOf` rejected | `one-of-lists` |
| `isSettled: false` ignored | `false-flag`, `all-clauses`, `tests-in-one-answer` |
| `isSettled` true for any stored value, so Deferred and empty count | `deferred-answer`, `empty-answers` |
| Deferred treated as a settled answer | `deferred-answer` |
| Empty text or an empty list treated as an answer | `empty-answers` |
| A list of only empty strings treated as an answer | `empty-answers` |
| A Flag `false` treated as no answer | `false-flag` |
| Flags equal whatever their value | `false-flag` |
| Text `true` equal to Flag `true` | `value-kinds`, `false-flag` |
| A list equal to a list | `list-answer`, `list-test-values` |
| A list matched by one of its items | `list-answer`, `list-test-values` |
| A one-item list equal to its Text | `list-answer`, `list-test-values` |
| A list test value ignored when another item of `isOneOf` matches | `list-test-values` |
| `isSettled` rejects a list answer | `list-answer`, `empty-answers` |
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
| A section clause ignored, so it always holds | `section-clause`, `section-clause-with-questions` |
| A section clause that always fails | `section-clause`, `no-switches` |
| `enabled` ignored, so the clause holds when the section is on | `section-clause` |
| `enabled` inverted, so the clause holds when the switch differs | `section-clause`, `no-switches` |
| A section missing from the switches read as off | `section-clause`, `no-switches` |
| A nil map read as every section off, or a crash on nil | `no-switches`, the nil switches test |
| The section id prefixed with the scope | `section-clause-scope`, `section-clause-in-plan-scope` |
| The section id read as a question id in the answers, so an `enabled` answer decides | `section-clause-ignores-answers` |
| A section clause that holds when the section has no stored answers, or fails when it has | `section-clause-ignores-answers` |
| A question clause that fails when the section it names is off | `answer-test-ignores-switches` |
| A bare name that reads nothing when its scope is off | `answer-test-ignores-switches` |
| A mixed list that checks only the question clauses, or only the section clauses | `section-clause-with-questions` |
| A mixed list read as any, so one holding clause is enough | `section-clause-with-questions` |
| Only the first or only the last clause of a mixed list read | `section-clause-with-questions` |
| The switches map changed by the call | The inputs-unchanged test |
