# Plan validator contract

Status: draft by an agent. Not reviewed by a person. Component: `plan/validator`. Issue: [#49](https://github.com/8thlight-daveloper/onboard-kit/issues/49).

The plan validator checks whether the answers in a plan are acceptable. `Validate` checks every answer stored in a plan, for example after a hand edit. `CheckValue` checks one proposed value for one question, so the editor can refuse a bad answer before it applies it. Both are pure functions. They read no file, keep no state between calls, and change no argument. See [0010](../../../docs/decisions/0010-plan-editor.md), [0035](../../../docs/decisions/0035-settled-and-deferred-answers.md), [0045](../../../docs/decisions/0045-defaults-are-suggestions.md), [0050](../../../docs/decisions/0050-section-switch.md), the [domain model](../../../docs/reference/domain-model.md), and the [variables reference](../../../docs/reference/section-template-schema/variables.md).

This contract covers the answer checks only. The checks between sections (references, `after`, cycles, timing, conditions) and the call to the section validator are other work, and they report other codes. A caller of `Validate` gets all of them in one list, and this contract judges only the problems whose code starts with `answer.` or `secret.`.

## Interface

`Validate(plan) -> list of Problem` and `CheckValue(plan, question, value) -> list of Problem`.

- `plan` is a Plan, as stored. The validator reads `plan.Variables`, `plan.Answers`, and `plan.Sections` (each section's `id`, `enabled`, `includeIf`, `timing`, `after`, and `variables`). It reads no other field.
- `question` is a question id, `<scope>.<name>`, such as `plan.team_state` or `access.repo_url`. It is split at the first dot. Any text is accepted, and an id that names nothing gives a problem.
- `value` is a Value, as a person or a driver would submit it. A value of no kind is allowed and is treated as unanswered.
- Both results are empty when nothing is wrong. There is no error return.

**Where the other components come from.** [0020](../../../docs/decisions/0020-binary-structure-and-regenerative-components.md) puts the validator after the question finder and the secrets check and lets it depend on both. `New()` takes no argument, so the validator uses the real components. It asks the finder for the active questions of the plan and the secrets check for each text. The expected problems in the fixtures therefore rest on the [finder contract](../../question/finder/contract.md) and the [secrets contract](../../secrets/contract.md), and they use only what those contracts fix.

## Rules

**Which questions can take an answer.** A question is active when the question finder returns it for the plan. That is a planning question of a section that is on and included, whose `askIf` holds. The plan's own variables are active unless their `askIf` fails. A variable with `audience: onboardee` is not active. A question can be declared and not active.

**What `Validate` checks.** Every answer stored in `plan.Answers`, plan scope and each section, in three groups:

1. A stored answer for a question that is active gets the answer checks below, and then the secrets check.
2. A stored answer for a question that is declared and not active gets only the secrets check. It was given for a question that may be asked again, and the plan file holds it as it is, so type, options, and pattern are not judged. The secrets check is, because the plan file holds the text.
3. A stored answer that no variable declares gets the problem `answer.unknown-question` with severity `warning`, and then the secrets check. A name that no variable of its scope declares, and a section id the plan does not hold, both count. This is a hand edit, and it never blocks assembly.

A variable's `default` is not an answer, and the validator never reads it. A required question with no settled or deferred answer is not a problem here. It is a blocker, and the finder's `Blockers` reports it ([0010](../../../docs/decisions/0010-plan-editor.md), [0045](../../../docs/decisions/0045-defaults-are-suggestions.md)).

**What `CheckValue` checks.** One value for one question, as if the plan stored it. The answers already stored in the plan are not checked, and they take part only in deciding which questions are active.

1. When no variable declares the question, the result is one problem, `answer.unknown-question`, with severity `error`. The id is read as written, so an id with no dot, and an empty id, are unknown.
2. When a variable declares it and it is not active, the result is one problem, `answer.inactive-question`, with severity `error`. The editor asks the finder, and this puts the same answer in the validator for any other caller.
3. Otherwise the result holds the answer checks for the value, and then the secrets check.

For a given active question and a given value, `CheckValue` gives the same problems that `Validate` gives for a plan that stores that value for that question.

**Answer checks.** They run on one value for one active question. The type of the variable is `string` when it has none.

| Step | Rule |
| -- | -- |
| Nothing to check | A value of no kind and a Deferred value get no answer check, whatever the type. Unanswered text, an empty list, and a list of only empty text get none either. Whitespace is text, so a text of only spaces is checked like any other text |
| Kind | The kind the question takes: a list when the variable is a `command` or has `multiple` true, a flag for a `boolean`, text for everything else. Another kind is `answer.wrong-kind`, and no other answer check runs |
| Options | A `choice` takes text that equals one of its `options` exactly, with case. Anything else is `answer.not-an-option`. The text `deferred` is plain text, so it is not an option |
| Pattern | A `string` with a `pattern` takes text that the pattern matches anywhere in it. The pattern is RE2, so `$` is the end of the text, and `.` does not match a line break. No match is `answer.pattern-mismatch`. The message holds the variable's `patternHint` when it has one |
| Person | A `person` answer is a text line, or a list of lines when `multiple` is true. Each line is split at commas. The first part is the name. The parts after it that are not empty after trimming are the ways to reach the person. A line with fewer than two gets `answer.thin-contact`. A list gets one for each such item, with the item's position in `item`. A list of people with exactly one non-empty item also gets `answer.single-person` |

`options` limit only a `choice`, and `pattern` applies only to a `string`. A string that lists options takes any text, and a person is not matched against a pattern. A pattern that does not compile is skipped here, because the section validator reports it.

**Secrets.** The secrets check runs on the stored or submitted value, whatever the question's type and whatever the answer checks said, unless the value is of no kind. It runs on text and on each list item, as the [secrets contract](../../secrets/contract.md) says. Its problems come through with its `code`, `severity`, `message`, and `detail` as they are, and the validator fills in `question` and `section`. A secrets problem is an `error`. No problem repeats the secret or a part of it.

## The problems

| Code | Severity | When | Detail |
| -- | -- | -- | -- |
| `answer.wrong-kind` | `error` | The value is not the kind the question takes | `expected` and `found`, each `text`, `flag`, or `list` |
| `answer.not-an-option` | `error` | Text for a `choice` is not one of its options | none |
| `answer.pattern-mismatch` | `error` | Text for a `string` does not match its pattern | none |
| `answer.thin-contact` | `warning` | A person line has fewer than two ways to reach the person | `item`, the position of the item as text, only for a list |
| `answer.single-person` | `warning` | A list of people holds exactly one person | none |
| `answer.unknown-question` | `warning` in `Validate`, `error` in `CheckValue` | No variable declares the question | none |
| `answer.inactive-question` | `error`, only in `CheckValue` | The question is declared and not active | none |
| `secret.token-prefix`, `secret.url-credential`, `secret.private-key`, `secret.random-string` | `error` | The secrets check found a secret | `item` for a list item, as the secrets contract says |

Every problem has a non-empty `message` for a person. The wording is not specified.

- `question` is the question id. For a stored answer it is `<scope>.<name>`. For `CheckValue` it is the id as given.
- `section` is the id of the plan section that holds the question, and it is empty for a plan variable and for a section the plan does not hold.
- A warning never stops an answer. An error does, as the [domain model](../../../docs/reference/domain-model.md) says.

**Order.** `Validate` lists the problems by question id, compared as bytes, from low to high. Two answers never share an id. For one question the order is: the answer checks, in the order of the table, then the secrets problems in item order. For `answer.unknown-question` the problem comes before the secrets problems. The order does not depend on the order of `plan.Sections`, or on the order the answers are held in.

## Properties

These hold for every plan the contract tests use.

- **Pure.** The plan and the value are the same after the calls as before.
- **No memory.** The same call gives the same result on a shared validator, whatever calls came before, forward or in reverse order, however often it is repeated, and when calls run at the same time.
- **Same plan, same result.** Rotating or reversing `plan.Sections` does not change a result.
- **New lists.** Changing a result, including a problem's `detail`, changes neither the plan nor a later result.
- **Agreement.** For every stored answer of an active question, `CheckValue` with that value gives the problems `Validate` gives for that question. For every stored answer of another question, `Validate` gives at most secrets problems and `answer.unknown-question`.
- **Empty plan.** A plan with nothing in it gives no problems. `CheckValue` on it gives `answer.unknown-question`.

## Side effects and plan access

None. It reads the plan's variables, the answers, and the content of the sections that the finder needs. It returns no plan. This matches the row for `plan/validator` in [components](../../../docs/architecture/components.md).

## Not covered

- Which problems `Validate` adds once the checks between sections and the section validator exist. Their codes do not start with `answer.` or `secret.`.
- A required question with no answer. It is a blocker.
- Whether a variable's `default` fits the type, options, and pattern. The section validator reports it.
- A pattern that does not compile, a `type` that is none of the five, and options listed for a `choice` that has none.
- Duplicate section ids, and duplicate variable names in one scope.
- A section clause or a question that a condition names and that does not exist. That is a check between sections.
- The wording of a message, apart from the pattern hint.
- A Text that is not valid UTF-8.

## Decisions to know

`hard` choices change what a template author writes, what a plan file holds, or what a person sees in the interface. They come first.

1. **Stored answers of questions that are not active are not checked, except for secrets.** An answer to a question hidden by its `askIf`, in a section that is off or left out, or for an onboarding question is kept in the plan and may become active again, so a wrong type or option there is not reported. The secrets check still runs, because the plan file holds the text. Raised by "It keeps its stored answer, is not asked, and does not block" and "Turning a section off keeps its answers" in 0010 and 0050, against "so a secret added to a plan by hand is reported as a problem" in 0010. Cases: `stale-answers`, `secrets`. `hard`
2. **A stored answer that no variable declares is a warning, and a submitted value for such a question is an error.** A typo in a hand edit would otherwise be ignored without a word. Raised by "unknown question id ... Severity `error` or `warning`" in the table of 0010. Cases: `unknown-answers`, `check-value`. `hard`
3. **`CheckValue` rejects a question that is declared and not active.** The editor then needs no activity check of its own, and a hidden question cannot be given an answer by mistake. Raised by the sequence in 0010, where the editor asks the finder before it asks the validator, and by the list of problem classes, which names only an unknown id. Case: `check-value`. `hard`
4. **A wrong kind, a value outside the options, and a pattern mismatch are errors. The person findings are warnings.** An error stops the answer. Raised by "a value of the wrong type" as an error example in 0010, "shown when a value does not match" for `patternHint`, and "a person with fewer than two gets a warning" in variables.md. Cases: `wrong-kinds`, `choice-options`, `patterns`, `people`. `hard`
5. **Options limit only a `choice`, and a pattern applies only to a `string`.** ACME lists options on `team_state` and sets no type, so its answer is free text. Raised by "The allowed values of a `choice`" and "A regular expression a `string` must match" in variables.md. Cases: `options-limit-only-choice`, `patterns`, `acme`. `hard`
6. **A person has too little contact when fewer than two parts after the name are not empty, and a list of exactly one non-empty item warns about a single point of failure.** Empty parts and empty items are not counted, and the name is not checked. Raised by "the comma-separated parts after the name" and "exactly one person gets a warning" in variables.md, which do not say how an empty part counts. Cases: `people`, `check-value`. `hard`
7. **The kind a question takes follows `multiple` and the type.** A list for a `command` and for any variable with `multiple`, a flag for a `boolean`, text otherwise. A `multiple` string therefore takes a list. Raised by "held as a list" for `multiple` and "Yes, as a list" for `command` in variables.md, which name `person` as the only `multiple` type. The finder offers lists the same way. Cases: `wrong-kinds`, `valid-answers`. `cheap`
8. **Empty text, an empty list, and a list of empty text are unanswered for every type, and get no answer check.** So empty text for a boolean is not a wrong kind. Whitespace is text and is checked. Raised by "nothing, empty text, or an empty list" in 0035 and "a single space is a legitimate answer" in 0040. Cases: `deferred-and-empty`, `choice-options`. `cheap`
9. **The text `deferred` is plain text.** For a choice it is not an option, and for a boolean it is the wrong kind. Only the Deferred kind is accepted for every type. Raised by "The quoted `"deferred"` is the Text answer with that word" in 0035. Case: `deferred-word`. `cheap`
10. **`CheckValue` is the name of the one-value function, not `ValidateAnswer`.** `ValidateAnswer` was also proposed. 0010 names `CheckValue` in its surface and in its sequences. Raised by "CheckValue(plan, id, value) []Problem" in 0010. Case: every `checks` list. `cheap`
11. **`CheckValue` stops at an unknown or inactive question and runs no secrets check.** The value is refused for that reason alone. Raised by the order in the 0010 sequence. Case: `secrets`. `cheap`
12. **The validator builds the finder and the secrets check itself, with no argument.** The contract tests therefore run the real components. Raised by the dependency order in 0020, which lets the validator depend on both, and by the wiring test rule, which allows one `New()` call. All cases. `cheap`
13. **A pattern mismatch message holds the variable's `patternHint`.** Raised by "shown when a value does not match" in variables.md. The message is otherwise free. Cases: `patterns`, `check-value`. `cheap`
14. **The problems come sorted by question id as bytes.** The docs give no order, and a map of answers has none. Raised by the same plan giving the same result in 0015. Cases: `problem-order`, `secrets`. `cheap`
15. **`section` is the id of a section the plan holds, and the details are `expected`, `found`, and `item`.** Raised by the Problem fields in the domain model, which leave both open. Cases: most. `cheap`
16. **A pattern is read as RE2, so `$` means the end of the text.** A text that ends in a line break does not match a pattern that ends in `$`. Raised by "`pattern` uses RE2 syntax" in variables.md. Cases: `patterns`. `cheap`

## Wrong implementations tried

Each row names a plausible wrong implementation and the cases or tests that fail it. Sixty variants were run against the tests, and each one fails at least one named test.

| Wrong implementation | Caught by |
| -- | -- |
| The kind is not checked, or a `multiple` variable or a `command` takes text, or a boolean takes the text `true` | `wrong-kinds`, `deferred-word`, `check-value` |
| A wrong kind reported as a warning, or `expected` and `found` swapped | `wrong-kinds` |
| A choice is not checked, is compared without case, or accepts the text `deferred` | `choice-options`, `deferred-word`, `check-value` |
| Options enforced on a string, or on a variable with no type | `options-limit-only-choice` |
| A pattern must match the whole text, is read with `(?s)` or `(?m)`, is ignored, or is applied to a person | `patterns`, `check-value` |
| The `patternHint` is left out of the message, or a pattern mismatch is a warning | `patterns`, `check-value` |
| A Deferred value is checked, or the text `deferred` is skipped as if it were Deferred | `deferred-and-empty`, `deferred-word` |
| Empty text, an empty list, or a list of empty text is checked, or only spaces count as unanswered | `deferred-and-empty`, `choice-options`, `check-value` |
| Answers of a hidden, onboarding, switched off, or left out question are checked | `stale-answers`, `check-value` |
| An unknown stored answer is not reported, or is an error in `Validate`, or a submitted unknown question is a warning | `unknown-answers`, `check-value` |
| The secrets check is skipped for inactive or unknown answers, or skipped in `CheckValue` | `secrets`, `check-value` |
| A secrets problem has no `question` or no `section`, loses its `item`, or repeats the secret in its detail | `secrets`, `check-value` |
| The secrets problems come before the answer problems | `secrets`, `problem-order` |
| `CheckValue` accepts an inactive question, or also runs the secrets check on an unknown question | `check-value`, the agreement test |
| A required question with no answer is reported | `not-problems` |
| A default is checked as if it were an answer | `not-problems` |
| A person with one way to reach them is fine, empty parts count, or the parts are not trimmed | `people`, `check-value` |
| The `item` of a thin contact starts at one, is missing, is counted among the non-empty items, or appears on a single person | `people`, `check-value` |
| A single-person warning for two people, or for a list with one empty item, or an error | `people`, `check-value` |
| The problems are not sorted, or sorted the other way, or depend on the order of the sections | `problem-order`, `secrets`, the section order test |
| `section` always empty, or set to `plan` for a plan variable | `wrong-kinds`, `choice-options`, `check-value` |
| The plan or the value is changed | The not-changed test |
| A result remembered from an earlier call | The no-memory test |
| A `detail` map shared between calls | The new-lists test |
