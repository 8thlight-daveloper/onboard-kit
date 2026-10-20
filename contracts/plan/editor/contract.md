# Plan editor contract

Status: draft by an agent. Not reviewed by a person. Component: `plan/editor`. Issue: [#70](https://github.com/8thlight-daveloper/onboard-kit/issues/70).

The plan editor changes a plan. It is the only component that returns a changed copy of one. `Submit` applies one answer. `SetSectionEnabled` turns one section on or off. Each returns a new plan and a list of problems, and neither alters the plan it is given. Both are pure functions. They read no file, read no clock, and keep no state between calls. See [0010](../../../docs/decisions/0010-plan-editor.md), [0015](../../../docs/decisions/0015-functional-core-and-imperative-shell.md), [0035](../../../docs/decisions/0035-settled-and-deferred-answers.md), [0045](../../../docs/decisions/0045-defaults-are-suggestions.md), [0050](../../../docs/decisions/0050-section-switch.md), and the [domain model](../../../docs/reference/domain-model.md).

The finder says which questions are active, and the plan validator says whether a value is acceptable ([question finder contract](../../question/finder/contract.md), [plan validator contract](../validator/contract.md)). The editor decides what to do with their answer.

## Interface

`Submit(plan, question, value) -> (plan, list of Problem)` and `SetSectionEnabled(plan, section, enabled) -> (plan, list of Problem)`.

- `plan` is a Plan, as stored. `Submit` reads `plan.Variables`, `plan.Answers`, and `plan.Sections`. `SetSectionEnabled` reads `plan.Sections`. Neither reads another field, and both return the other fields as they are.
- `question` is a question id, `<scope>.<name>`, such as `plan.team_state` or `access.repo_url`. The scope `plan` means the plan's own variables. Any other scope is a section id. The id is split at the first dot. Any text is accepted. An id that names nothing gives a problem.
- `value` is a Value, as a person or a driver would submit it. A value of no kind is allowed. It means unanswered.
- `section` is a section id. `enabled` is true or false.
- Both results are a plan and a list of problems. The list is empty when there is nothing to report. There is no error return.
- The editor takes the finder and the validator through `New()` with no argument, as the validator contract does. The contract tests therefore run the real components, and the expected problems rest on the finder and validator contracts and use only what those fix.

## Rules for Submit

**Judging the value.** The editor asks the validator to check the value for the question, against the plan as it was given, and returns every problem the check reports, as they are and in the same order. The editor adds no problem of its own to `Submit`. The checks and their codes are in the [plan validator contract](../validator/contract.md): an unknown question, a question that is declared and not active, a value of the wrong kind, a value outside the options, a pattern mismatch, thin contacts and single people, and secrets. A question is active when the finder returns it for the plan as given, so an answer that reveals a follow-up question does not make that question active in the same call.

**Errors and warnings.**

- When any problem has severity `error`, the answer is not applied. The returned plan equals the plan given, in everything it holds. The problems come back with it, warnings included.
- Otherwise the answer is applied, and any warnings come back with the new plan. A warning never stops an answer.

**Applying a value.** The editor puts the value in the plan's `answers`. A question in scope `plan` is stored under `answers.plan` and a question in a section under `answers.sections.<section id>`, each by variable name. The same name in two scopes is two answers.

- A settled value replaces the stored answer, or adds one. It is stored exactly as given: text as written, with spaces and with the text `deferred` as plain text, a flag, and a list in the order given, including any empty item in it.
- A Deferred value is applied like any other, and it replaces what was stored. A stored Deferred is replaced by a settled value in the same way.
- An unanswered value removes the stored answer, if there is one. Unanswered means a value of no kind, empty text, an empty list, and a list of only empty text, for a question of any type. A question with no stored answer is left as it is. The editor never stores an unanswered value. Text of only spaces is text, and it is stored.
- Every other stored answer stays: the other questions of the same scope, and every answer of every other scope, active or not. Applying an answer does not remove the stored answer of a question that the new answer hides, as [0010](../../../docs/decisions/0010-plan-editor.md) says.
- The plan differs from the argument in `answers` only. Whether a section's map is left in `answers.sections` when its last answer is removed is not specified.

## Rules for SetSectionEnabled

The section switch is a field of the plan section and not an answer ([0050](../../../docs/decisions/0050-section-switch.md)). The editor offers it as its own operation.

- When `section` is the id of a section the plan holds, the returned plan is the argument with that section's `enabled` set, and the result has no problem.
- Turning a section **off** stores `enabled: false` on it. Turning a section **on** removes the field, so that the plan records only the sections turned off. This holds when the section already has an explicit `enabled: true` written by hand: it is removed.
- A section already in the requested state stays in it. Turning on a section that has no field, and turning off a section that is off, change nothing.
- Nothing else changes. The answers of every section stay, including those of the section turned off, and so do the other sections, their order, and every other field of the plan.
- A section id that the plan does not hold, an empty id, and the id `plan` give one problem, `switch.unknown-section`, with severity `error`, and the returned plan equals the plan given. The problem's `section` is the id as given, and its `question` is empty.
- The editor does not report sections that depend on the switch. A question of a section that is off is not active, so a later `Submit` for it is refused with `answer.inactive-question`.

## The problems

| Code | Severity | From | When | Detail |
| -- | -- | -- | -- | -- |
| `answer.*` and `secret.*` | As the validator reports them | `Submit`, from the validator | The value or the question is not acceptable | As the validator reports them |
| `switch.unknown-section` | `error` | `SetSectionEnabled` | The plan holds no section with that id | none |

Every problem has a non-empty `message` for a person. The wording is not specified, and a problem never repeats a secret or a part of it.

## Properties

These hold for every plan the contract tests use.

- **Arguments unchanged.** The plan and the value are the same after the calls as before.
- **Same input, same output.** The same call gives the same result on a shared editor, whatever calls came before, forward or in reverse order, however often it is repeated, and when calls run at the same time.
- **Nothing shared that an edit changes.** The returned plan holds its own `answers`, with their maps and their lists, and its own `sections` list and each section's `enabled`. Changing any of them in the result changes neither the argument nor a later result, and changing the same parts of the argument after the call does not change the result. This holds when the answer is refused too. A list in a submitted value is copied when it is stored. Other parts of the plan, such as a section's variables and steps, may be shared.
- **Only the named change.** A `Submit` that applies changes one entry of `answers` and nothing else. `SetSectionEnabled` changes the `enabled` field of one section and nothing else. In both, every field outside the change, `unknownFields` included, comes back as it was.
- **Examples stay.** Every variable of the returned plan, in `variables` and in each section, keeps its `example` as held, whether the call applies, refuses, clears, or switches. An `example` is an illustration for the person answering. It is not an answer and not a default, and no call changes it.
- **Section order.** The sections of the result are in the order of the argument's sections. Rotating or reversing the argument's sections does not change the answers or the problems.
- **Repeating changes nothing more.** The same call on the result gives a plan equal to the result.
- **Take back.** Clearing an answer that a `Submit` added gives the plan from before. Turning off a section that has no `enabled` field and turning it on again gives the plan from before.
- **New lists.** Changing the result's problems, including a problem's `detail`, changes no later result.
- **Nothing to edit.** On a plan with nothing in it, `Submit` gives `answer.unknown-question`, `SetSectionEnabled` gives `switch.unknown-section`, and the plan stays empty.

## Side effects and plan access

None. `Submit` returns a changed copy that differs only in `answers`, and `SetSectionEnabled` returns a changed copy that differs only in the `enabled` field of one section. Neither alters the plan it was given. The row for `plan/editor` in [components](../../../docs/architecture/components.md) lists `answers` only, and 0010 says the copy "differs only in `answers`", so both need the switch added. The editor imports the finder, the validator, and the model, and nothing else, as [0020](../../../docs/decisions/0020-binary-structure-and-regenerative-components.md) says.

## Not covered

- Whether a section's map stays in `answers.sections` after its last answer is removed.
- Duplicate section ids, and duplicate variable names in one scope.
- A section that holds `enabled: true` and is turned off, and a plan that holds both a `plan` section and plan variables.
- The wording of a message, and the codes the validator will add for checks between sections, which `Submit` does not return.
- A Text that is not valid UTF-8.
- Whether a driver asks for confirmation, shows warnings, or offers a default. Drivers decide that.

## Decisions to know

`hard` choices change what a template author writes, what a plan file holds, or what a person sees in the interface. They come first.

1. **An unanswered value removes the stored answer, and the plan holds no empty answers.** Empty text, an empty list, a list of only empty text, and a value of no kind clear the question. The other choice was to store the empty value. Raised by "Unanswered | Nothing, empty text, or an empty list" in 0035 and "The plan holds only answers" in 0045. Cases: `clear-with-empty-text`, `clear-with-an-empty-list`, `clear-with-a-list-of-empty-text`, `clear-with-no-value`, `clear-a-flag-with-empty-text`. `hard`
2. **The section switch is a second operation, `SetSectionEnabled`, and the plan file records only the sections turned off.** Off stores `enabled: false`, on removes the field, and a hand-written `enabled: true` is removed by a switch to on. Raised by "The editor offers it as an action on a section" and "A section is on unless the plan says otherwise" in 0050. The architecture row for the editor lists `answers` only, and it needs the switch. Cases: `switch-off`, `switch-on-clears-the-field`, `switch-on-clears-an-explicit-true`, `switch-off-over-an-explicit-true`. `hard`
3. **A refused answer returns the plan unchanged and every problem, and a warning comes back with the applied answer.** An error stops the answer, and the problems of the check come back whole, warnings next to the error. Raised by "If an answer has an error, it is not applied, and the returned plan equals the original. A warning is returned with the new plan" in 0010. Cases: `secret-is-refused`, `an-error-and-a-warning-are-both-returned`, `warning-is-applied-and-returned`, `single-person-warning-is-applied`. `hard`
4. **A question that is declared and not active cannot be answered, and cannot be cleared.** A stale answer in a section that is off, or hidden by its condition, stays in the plan until the question is active again, and the editor reports `answer.inactive-question`. Raised by "Is team state an active question?" in the 0010 sequence and "Turning a section off keeps its answers" in 0050. Cases: `hidden-question-is-refused`, `question-of-a-switched-off-section-is-refused`, `clearing-an-inactive-question-is-refused`, `onboarding-question-is-refused`. `hard`
5. **The editor adds no problem codes to `Submit`.** Every problem is the validator's, with its code, severity, `question`, `section`, and detail, because the stdio protocol returns them as they are. An unknown question and an inactive one come from the validator. Raised by "Submit(plan, id, value) (Plan, []Problem)" and the errors table in 0010, which name unknown ids as problems without a code. Cases: all the refused cases. `hard`
6. **An unknown section for the switch is the error `switch.unknown-section`.** The `plan` id counts as unknown, and the problem's `section` is the id as given. Raised by 0050, which says nothing about an id the plan does not hold. Cases: `switch-of-an-unknown-section-is-refused`, `plan-is-not-a-section`, `switch-of-an-empty-id-is-refused`. `hard`
7. **A switch reports no problem about other sections.** Turning off a section that another section's condition names, or whose questions are blocking, gives no warning. Raised by "It is not a question, it never blocks" in 0050. Cases: `switch-off`, `acme-switch-off-a-section`. `cheap`
8. **An answer is stored exactly as given.** The editor does not trim text, drop empty items from a list, or turn the text `deferred` into a Deferred value. Raised by "a single space is a legitimate answer" in 0040 and "The quoted `"deferred"` is the Text answer with that word" in 0035. Cases: `whitespace-text-is-an-answer`, `list-is-stored-as-given`, `the-word-deferred-is-text`. `cheap`
9. **The returned plan is a new plan that equals the original in a refusal, and it shares none of the parts the editor changes.** The answers, the sections list, and each section's `enabled` are fresh in every result. This is a stricter reading of "Plans are values" than the sharing allowed in 0015, so that a driver or a test that changes one plan cannot change another. Raised by "A returned value may share unchanged parts with its input" in 0015 and "No shared mutation" in 0010. Cases: the property test of shared parts. `cheap`
10. **The editor reaches the finder through the validator.** The validator's value check already asks the finder whether the question is active, so the editor makes no second activity check. Raised by the 0010 sequence, which shows the editor asking the finder and then the validator. Cases: the cases of decision 4. `cheap`
11. **Clearing needs an active question like any answer.** An unanswered value for an unknown question is `answer.unknown-question`, and for one that is not active it is `answer.inactive-question`. Raised by the same sequence. Case: `clearing-an-unknown-question-is-refused`. `cheap`
12. **A switch to the state a section already has is not a problem.** The result equals the plan, apart from a hand-written `enabled: true` that is removed. Raised by 0050, which does not say. Cases: `switch-on-when-on`, `switch-off-when-off`. `cheap`
13. **The sections of the result keep the argument's order.** Raised by "differs only in" in 0010. Cases: the section order property. `cheap`

## Wrong implementations tried

Each row names a plausible wrong implementation and the cases or tests that fail it. Forty variants were run against the tests. Each one with a defect in the editor's output fails at least one named test. One variant survives on purpose: it leaves an empty map for a section whose last answer was removed, which the contract leaves open.

| Wrong implementation | Caught by |
| -- | -- |
| An answer with an error is applied anyway | `pattern-mismatch-is-refused`, `hidden-question-is-refused`, `secret-is-refused` |
| A warning stops the answer | `warning-is-applied-and-returned`, `single-person-warning-is-applied` |
| The problems are dropped, or the warnings are dropped when there is an error | `secret-is-refused`, `an-error-and-a-warning-are-both-returned` |
| A refusal returns an empty plan | `secret-is-refused`, `question-of-a-switched-off-section-is-refused` |
| The answer is stored by name only, or a section answer is stored in the plan scope | `same-name-in-a-section-scope`, `new-section-answer`, `command-list-replaces-the-old-list` |
| Empty text, an empty list, or a list of empty text is stored instead of clearing | `clear-with-a-list-of-empty-text`, `clear-a-plan-answer`, `clear-when-nothing-is-stored` |
| Only empty text clears, or only a value of no kind clears | `clear-with-an-empty-list`, `clear-with-a-list-of-empty-text`, `answer-hides-a-question-and-keeps-its-answer` |
| A Deferred value clears the answer, or the text `deferred` is treated as unanswered | `deferred-is-applied`, `the-word-deferred-is-text` |
| Text is trimmed, so a text of spaces is cleared | `whitespace-text-is-an-answer` |
| Empty items are dropped from a list | `list-is-stored-as-given` |
| Clearing removes every answer of the section | `clear-with-an-empty-list`, `clear-with-no-value` |
| Clearing skips the check, so an inactive or unknown question is cleared | `clearing-an-inactive-question-is-refused`, `clearing-an-unknown-question-is-refused` |
| The answers of the other sections are dropped | `question-after-switch-on-takes-an-answer`, `warning-is-applied-and-returned` |
| An inactive or unknown question is applied, because those problems are filtered out | `a-question-of-a-section-out-is-refused`, `unknown-question-is-refused`, `empty-id-is-refused` |
| Provenance changed, or the unknown fields dropped | `new-plan-answer`, `acme-change-the-clone-url`, `list-is-stored-as-given` |
| A section turned on stores `enabled: true`, or keeps a hand-written true | `switch-on-clears-the-field`, `switch-on-when-on`, `switch-on-clears-an-explicit-true` |
| A section turned off clears the answers of the section | `switch-off-when-off`, `acme-switch-off-a-section` |
| Every section is turned off, or the first section is turned off instead | `switch-off`, `switch-leaves-other-sections-alone` |
| The sections are sorted by id, by `Submit` or by `SetSectionEnabled` | `new-plan-answer`, `switch-on-when-on`, the section order property |
| An unknown section gives no problem, or a warning, or no `section` | `switch-of-an-unknown-section-is-refused`, `plan-is-not-a-section`, `switch-of-an-empty-id-is-refused` |
| Turning off a section reports a warning | `switch-off`, `switch-off-when-off` |
| A variable's `example` dropped when the plan is copied, in the plan's variables, in every section, or in one section, after an applied answer, a clear, a refusal, or a switch | `examples-kept-by-a-section-answer`, `examples-kept-by-a-plan-answer`, `examples-kept-by-a-clear`, `examples-kept-by-a-refused-answer`, `examples-kept-by-a-switch-off`, `examples-kept-by-a-switch-on` |
| An argument is changed | The arguments-unchanged test |
| The result shares the answers, the sections list, a stored list, or an `enabled` pointer with the argument | The shared-parts test, the stored-list test, the arguments-unchanged test |
| An earlier result returned on a later call | The same-input test, and most cases |
