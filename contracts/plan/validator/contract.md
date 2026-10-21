# Plan validator contract

Status: draft by an agent. Not reviewed by a person. Component: `plan/validator`. Issue: [#49](https://github.com/8thlight-daveloper/onboard-kit/issues/49).

The plan validator checks whether the answers in a plan are acceptable. `Validate` checks every answer stored in a plan, for example after a hand edit. `CheckValue` checks one proposed value for one question, so the editor can refuse a bad answer before it applies it. Both are pure functions. They read no file, keep no state between calls, and change no argument. See [0080](../../../docs/decisions/0080-directory.md), [0075](../../../docs/decisions/0075-refreshing-a-personal-plan.md), [0055](../../../docs/decisions/0055-variable-audience.md), [0010](../../../docs/decisions/0010-plan-editor.md), [0035](../../../docs/decisions/0035-settled-and-deferred-answers.md), [0045](../../../docs/decisions/0045-defaults-are-suggestions.md), [0050](../../../docs/decisions/0050-section-switch.md), the [domain model](../../../docs/reference/domain-model.md), and the [variables reference](../../../docs/reference/section-template-schema/variables.md).

This contract covers the answer checks, the integrity of the plan's directory, and two plan-wide id checks (duplicate section ids, more than one team variable). The checks between sections (references, `after`, cycles, timing, conditions) and the call to the section validator are other work, and they report other codes. A caller of `Validate` gets all of them in one list, and this contract judges only the problems whose code starts with `answer.`, `secret.`, or `directory.`, and the two codes `plan.duplicate-section-id` and `plan.duplicate-team-variable`.

## Interface

`Validate(plan) -> list of Problem` and `CheckValue(plan, question, value) -> list of Problem`.

- `plan` is a Plan, as stored. The validator reads `plan.Directory`, `plan.Variables`, `plan.Answers`, and `plan.Sections` (each section's `id`, `enabled`, `includeIf`, `timing`, `after`, and `variables`). It reads no other field.
- `question` is a question id, `<scope>.<name>`, such as `plan.team_state` or `access.repo_url`. It is split at the first dot. Any text is accepted, and an id that names nothing gives a problem. A name may be empty, so `main.` names a variable with an empty name. An id with no dot names nothing, even when its text is the scope of a variable with an empty name. The scope `plan` always means the plan's own variables, even when a section has the id `plan`, and such a question has no `section`.
- `value` is a Value, as a person or a driver would submit it. A value of no kind is allowed and is treated as unanswered.
- `CheckValue` reads the directory only to look up the ids a value names. It never reports a problem about the directory itself or about the plan as a whole.
- Both results are empty when nothing is wrong. There is no error return.

**Where the other components come from.** [0020](../../../docs/decisions/0020-binary-structure-and-regenerative-components.md) puts the validator after the question finder and the secrets check and lets it depend on both. `New()` takes no argument, so the validator uses the real components. It asks the finder for the active questions of the plan, for both audiences, and the secrets check for each text. The expected problems in the fixtures therefore rest on the [finder contract](../../question/finder/contract.md) and the [secrets contract](../../secrets/contract.md), and they use only what those contracts fix.

## Rules

**Which questions can take an answer.** A question is active when the question finder returns it for the plan with the empty audience, which means both. That is a question of either audience, in a section that is on and included, whose `askIf` holds. The plan's own variables are active unless their `askIf` fails. A variable with `audience: onboardee` is active on the same terms as one with `audience: onboarder`, and its answer is checked by the same rules. A question can be declared and not active.

**What `Validate` checks.** Every answer stored in `plan.Answers`, plan scope and each section, in three groups:

1. A stored answer for a question that is active gets the answer checks below, and then the secrets check.
2. A stored answer for a question that is declared and not active gets only the secrets check. It was given for a question that may be asked again, and the plan file holds it as it is, so type, options, and pattern are not judged. The secrets check is, because the plan file holds the text.
3. A stored answer that no variable declares gets the problem `answer.unknown-question` with severity `warning`, and then the secrets check. A name that no variable of its scope declares, and a section id the plan does not hold, both count. This is a hand edit, and it never blocks assembly.

**Plan-wide and directory checks in `Validate`.** These do not depend on the answers or on which questions are active, and none of their problems has a `question`.

1. Two or more sections with the same `id` give one `plan.duplicate-section-id` for that id, however many sections share it. What the answer checks do for the questions of such sections is not specified.
2. Two or more variables of type `team` in `plan.variables` give one `plan.duplicate-team-variable` for the plan, however many there are. A `team` variable in a section does not count, and neither do variables of other types.
3. Two or more entries of one directory list that share an id give one `directory.duplicate-id` for that list and id, however many entries share it. The three lists are separate: a contact, a role, and a team may share an id.
4. Each holder id in a role's `holders` that is not the id of a contact gives `directory.holder-not-a-contact`. So does each holder id in each list of a team's `holders`. A holder id is compared exactly, with case, and a role id is not a contact id. A team whose `contact` is not empty and is not the id of a contact also gives this code, once for the team, with the detail `team-contact` in place of `holder`. The same exact comparison applies, so a role id, a team id, a different case, and a text of only spaces are not contacts. A team with no `contact`, or an empty one, is fine, and a team's `contact` never counts as one of its holders. A wrong `contact` and a wrong holder of the same team are two problems.
5. Each key of a team's `holders` that is not the id of a role gives `directory.team-role-unknown`. The holders listed under that key are still checked by rule 4.
6. A role whose own `holders` list is empty or absent gives `directory.role-vacant`. A team's override does not change this: the role is reported as vacant for the organization even when a team names holders for it, and a team's empty override for a role that has its own holders is not reported. A role whose holders all name no contact is not vacant here, because rule 4 reports them.
7. A contact whose `email` is not empty and has no `@`, or has a space (U+0020), gives `directory.email-malformed`. Nothing else about the address is checked.
8. A contact whose `email`, `slack`, and `slackId` are all empty gives `directory.contact-unreachable`. A contact with a malformed email has an email, so it is reachable.

Rules 4, 6, 7, and 8 give one problem for each entry, holder, or team contact that has the defect, and an entry listed twice gives the problem twice. Two teams that name the same wrong contact give two problems. The format of an id (the lowercase slug) is not checked.

A variable's `default` is not an answer, and the validator never reads it. A required question with no settled or deferred answer is not a problem here. It is a blocker, and the finder's `Blockers` reports it ([0010](../../../docs/decisions/0010-plan-editor.md), [0045](../../../docs/decisions/0045-defaults-are-suggestions.md)).

**What `CheckValue` checks.** One value for one question, as if the plan stored it. The answers already stored in the plan are not checked, and they take part only in deciding which questions are active.

1. When no variable declares the question, the result is one problem, `answer.unknown-question`, with severity `error`. The id is read as written, so an id with no dot, and an empty id, are unknown.
2. When a variable declares it and it is not active, the result is one problem, `answer.inactive-question`, with severity `error`. The editor asks the finder, and this puts the same answer in the validator for any other caller.
3. Otherwise the result holds the answer checks for the value, and then the secrets check.

For a given active question and a given value, `CheckValue` gives the same problems that `Validate` gives for a plan that stores that value for that question, except that a `directory.unknown-id` problem is an `error` in `CheckValue` and a `warning` in `Validate`.

**Answer checks.** They run on one value for one active question. The type of the variable is `string` when it has none.

| Step | Rule |
| -- | -- |
| Nothing to check | A value of no kind and a Deferred value get no answer check, whatever the type. Unanswered text, an empty list, and a list of only empty text get none either. Whitespace is text, so a text of only spaces is checked like any other text |
| Kind | The kind the question takes: text for a `team`, whatever its `multiple`. Otherwise a list when the variable is a `command` or has `multiple` true, a flag for a `boolean`, text for everything else. A `contact` or `role` is therefore text, or a list when `multiple` is true. Another kind is `answer.wrong-kind`, and no other answer check runs |
| Options | A `choice` takes text that equals one of its `options` exactly, with case. Anything else is `answer.not-an-option`. The text `deferred` is plain text, so it is not an option |
| Pattern | A `string` with a `pattern` takes text that the pattern matches anywhere in it. The pattern is RE2, so `$` is the end of the text, and `.` does not match a line break. No match is `answer.pattern-mismatch`. The message holds the variable's `patternHint` when it has one |
| Directory id | A `contact`, `role`, or `team` answer is an id, and it is looked up among the directory's contacts, roles, or teams, by exact comparison with case and with no trimming. A list is one id for each item, and an empty item is skipped. An id that is not found, including the text `deferred` and a text of only spaces, is `directory.unknown-id`. A list gets one for each such item, with the item's position in `item`. A wrong kind stops this check |

`options` limit only a `choice`, and `pattern` applies only to a `string`. The directory id check applies to the three directory types only. A string that lists options takes any text, and a `choice` is not matched against a pattern. A pattern that does not compile is skipped here, because the section validator reports it. The options and the pattern judge text only. A variable with `multiple` that holds a list gets neither check, whatever its type, and each item of the list is left unjudged. The section validator rejects `multiple` on a `choice` and on a `string`.

**Secrets.** The secrets check runs on the stored or submitted value, whatever the question's type and whatever the answer checks said, unless the value is of no kind. It runs on text and on each list item, as the [secrets contract](../../secrets/contract.md) says. Its problems come through with its `code`, `severity`, `message`, and `detail` as they are, and the validator fills in `question` and `section`. A secrets problem is an `error`. No problem repeats the secret or a part of it.

## The problems

| Code | Severity | When | Detail |
| -- | -- | -- | -- |
| `answer.wrong-kind` | `error` | The value is not the kind the question takes | `expected` and `found`, each `text`, `flag`, or `list` |
| `answer.not-an-option` | `error` | Text for a `choice` is not one of its options | none |
| `answer.pattern-mismatch` | `error` | Text for a `string` does not match its pattern | none |
| `directory.unknown-id` | `warning` in `Validate`, `error` in `CheckValue` | A `contact`, `role`, or `team` answer names an id that is not in the directory | `kind`, which is `contact`, `role`, or `team`, and `item` for a list. Never the id itself |
| `directory.duplicate-id` | `error` | Two or more entries of one directory list share an id | `kind` and `id` |
| `directory.holder-not-a-contact` | `error` | A holder of a role, or of a team's override, is not a contact. Or a team's `contact` is not a contact | For a holder: `role`, `holder`, and `team` for a team's override only. For a team's contact: `team` and `team-contact`, the id it names, and no `role` or `holder` |
| `directory.team-role-unknown` | `error` | A team overrides a role that does not exist | `team` and `role` |
| `directory.role-vacant` | `warning` | A role has no holder of its own | `role` |
| `directory.email-malformed` | `warning` | A contact's email has no `@` or has a space | `contact`, the contact's id |
| `directory.contact-unreachable` | `warning` | A contact has no email, slack, or slackId | `contact` |
| `plan.duplicate-section-id` | `error` | Two or more sections share an id | none. The id is in `section` |
| `plan.duplicate-team-variable` | `error` | Two or more plan variables have type `team` | `variables`, their names in declaration order, joined with a comma |
| `answer.unknown-question` | `warning` in `Validate`, `error` in `CheckValue` | No variable declares the question | none |
| `answer.inactive-question` | `error`, only in `CheckValue` | The question is declared and not active | none |
| `secret.token-prefix`, `secret.url-credential`, `secret.private-key`, `secret.random-string` | `error` | The secrets check found a secret | `item` for a list item, as the secrets contract says |

Every problem has a non-empty `message` for a person. The wording is not specified.

- `question` is the question id. For a stored answer it is `<scope>.<name>`. For `CheckValue` it is the id as given. It is empty for the plan-wide and directory problems.
- `section` is the id of the plan section that holds the question, and it is empty for a plan variable and for a section the plan does not hold. For `plan.duplicate-section-id` it is the duplicated id. The directory problems and `plan.duplicate-team-variable` have none.
- A `directory.unknown-id` problem never repeats the id the answer named, in its message or its detail, because the text may be a secret.
- **The two severities of `directory.unknown-id`.** In `Validate` it is a `warning`: the stored answer counts as unanswered and the plan stays usable. In `CheckValue` it is an `error`: the tool never creates an answer that names an id the directory lacks, so the value is refused. The difference is on purpose, as it is for `answer.unknown-question`.
- A warning never stops an answer. An error does, as the [domain model](../../../docs/reference/domain-model.md) says.

**Order.** `Validate` lists the problems by question id, compared as bytes, from low to high, and an empty question id is the lowest. So the plan-wide and directory problems, which have no question, come first, before every problem about an answer. Among them the order is by `code` as bytes, and then by these keys, compared as bytes, with an absent detail key before any value. So the problems about a team's contact, which have no `role`, come before the problems about a holder, and they come in the order of their `team`, then their `team-contact`:

| Code | Keys, in order |
| -- | -- |
| `directory.contact-unreachable`, `directory.email-malformed` | `contact` |
| `directory.duplicate-id` | `kind`, `id` |
| `directory.holder-not-a-contact` | `role`, `team`, `holder`, `team-contact` |
| `directory.role-vacant` | `role` |
| `directory.team-role-unknown` | `team`, `role` |
| `plan.duplicate-section-id` | `section` |
| `plan.duplicate-team-variable` | none, there is one |

Two answers never share an id. For one question the order is: the answer checks, in the order of the table, then the secrets problems in item order. For `answer.unknown-question` the problem comes before the secrets problems. The order does not depend on the order of `plan.Sections`, on the order of the directory's lists or of any holders list, or on the order the answers are held in.

## Properties

These hold for every plan the contract tests use.

- **Pure.** The plan and the value are the same after the calls as before.
- **No memory.** The same call gives the same result on a shared validator, whatever calls came before, forward or in reverse order, however often it is repeated, and when calls run at the same time.
- **Same plan, same result.** Rotating or reversing `plan.Sections` does not change a result. Reversing the directory's three lists and every holders list does not change one either.
- **New lists.** Changing a result, including a problem's `detail`, changes neither the plan nor a later result.
- **Agreement.** For every stored answer of an active question, `CheckValue` with that value gives the problems `Validate` gives for that question, apart from the severity of `directory.unknown-id`. For every stored answer of another question, `Validate` gives at most secrets problems and `answer.unknown-question`.
- **Empty plan.** A plan with nothing in it gives no problems. `CheckValue` on it gives `answer.unknown-question`.

## Side effects and plan access

None. It reads the plan's directory, variables, answers, and the content of the sections that the finder needs. It returns no plan. This matches the row for `plan/validator` in [components](../../../docs/architecture/components.md).

## Not covered

- Which problems `Validate` adds once the checks between sections and the section validator exist. Their codes are not the ones this contract judges.
- A required question with no answer. It is a blocker.
- Whether a variable's `default` fits the type, options, and pattern. The section validator reports it.
- A pattern that does not compile, a `type` that is none of the known types, and options listed for a `choice` that has none.
- Duplicate variable names in one scope, and the answer checks for the questions of a section whose id is shared.
- The format of a directory id, an empty id, the same holder listed twice in one list, and whether a slack handle or a slack id is well formed.
- Whether a team variable is answered, and a template that gives a `team` variable `multiple`. A template check reports those.
- Whitespace in an email other than the space character.
- The wording of the directory messages.
- A section clause or a question that a condition names and that does not exist. That is a check between sections.
- The wording of a message, apart from the pattern hint.
- A Text that is not valid UTF-8.

## Decisions to know

`hard` choices change what a template author writes, what a plan file holds, or what a person sees in the interface. They come first.

1. **A stored answer that names an id the directory lacks is a `warning` in `Validate`, and a submitted value that names one is an `error` in `CheckValue`.** The code is `directory.unknown-id` in both. A stored answer counts as unanswered and never blocks the plan. A submitted value is refused, because the tool never creates such an answer. Raised by "an answer that names an id that is not in the directory" in the warnings of Integrity, and "The tool never creates such an answer" in 0080. Cases: `directory-answers`, `directory-check-value`. `hard`
2. **A role is vacant when its own `holders` list is empty or absent, whatever the teams do.** A team that names holders for it does not clear the warning, and a team's empty override for a role that has holders is not reported. Raised by "a role with no holder" in Integrity and "No holder is a gap" with the team override in 0080. Case: `directory-warnings`. `hard`
3. **A `team` answer is always one text, even when the variable says `multiple`.** The wrong kind is `answer.wrong-kind`. Raised by "A `team` question's options are the directory's teams" in 0080, which gives a list only to `contact` and `role`. Cases: `directory-answers`, `directory-check-value`. `hard`
4. **Only the variables of `plan.variables` are counted for the team variable.** Two of them give one `plan.duplicate-team-variable` with their names in `variables`. A team variable in a section, or two in one section, is not a problem here. Raised by "the answer to the plan variable of type `team`" and "at most one" in the shared decisions, and "`includeIf` with `isOneOf` on the team answer" in 0080. Cases: `two-team-variables`, `team-variable-scopes`. `hard`
5. **Ids are unique within a list, and the three lists are separate namespaces.** A contact, a role, and a team may share an id, and a holder or an answer is read against one list. A repeated id gives one `directory.duplicate-id` for that list, whatever the number of entries. Raised by "each entry with a permanent id" and "a duplicate id" in 0080. Case: `directory-duplicates`. `hard`

6. **Stored answers of questions that are not active are not checked, except for secrets.** An answer to a question hidden by its `askIf`, or in a section that is off or left out, is kept in the plan and may become active again, so a wrong type or option there is not reported. The secrets check still runs, because the plan file holds the text. Raised by "It keeps its stored answer, is not asked, and does not block" and "Turning a section off keeps its answers" in 0010 and 0050, against "so a secret added to a plan by hand is reported as a problem" in 0010. Cases: `stale-answers`, `secrets`, `audience-answers`. `hard`
7. **A stored answer that no variable declares is a warning, and a submitted value for such a question is an error.** A typo in a hand edit would otherwise be ignored without a word. Raised by "unknown question id ... Severity `error` or `warning`" in the table of 0010. Cases: `unknown-answers`, `check-value`. `hard`
8. **`CheckValue` rejects a question that is declared and not active.** The editor then needs no activity check of its own, and a hidden question cannot be given an answer by mistake. Raised by the sequence in 0010, where the editor asks the finder before it asks the validator, and by the list of problem classes, which names only an unknown id. Case: `check-value`. `hard`
9. **A wrong kind, a value outside the options, and a pattern mismatch are errors.** An error stops the answer. Raised by "a value of the wrong type" as an error example in 0010 and "shown when a value does not match" for `patternHint`. Cases: `wrong-kinds`, `choice-options`, `patterns`. `hard`
10. **Options limit only a `choice`, and a pattern applies only to a `string`.** ACME lists options on `team_state` and sets no type, so its answer is free text. Raised by "The allowed values of a `choice`" and "A regular expression a `string` must match" in variables.md. Cases: `options-limit-only-choice`, `patterns`, `acme`. `hard`
11. **A team's `contact` that is not a contact gives `directory.holder-not-a-contact` once for the team, with the detail `team` and `team-contact`.** The same code as a bad holder, because 0080 lists them in one row. The detail key `team-contact` says what it is, and the problem has no `holder`, so a person and a tool can tell it from a team override's holder. It sorts before the holder problems of the same code, by `team` and then `team-contact`, because an absent `role` sorts first. A team with no contact is fine, and the contact never counts as a holder. Raised by "a team's `contact` names an id that is not a contact" in Integrity and "never stands in for a role holder" in 0080. Case: `team-contact`. `hard`
12. **The kind a question takes follows `multiple` and the type.** A list for a `command` and for any variable with `multiple`, a flag for a `boolean`, text otherwise. A `multiple` string therefore takes a list. Raised by "held as a list" for `multiple` and "Yes, as a list" for `command` in variables.md. The section validator rejects `multiple` on any type but `contact` and `role`, and this validator reads `multiple` as it is. The finder offers lists the same way. Cases: `wrong-kinds`, `valid-answers`. `cheap`
13. **Empty text, an empty list, and a list of empty text are unanswered for every type, and get no answer check.** So empty text for a boolean is not a wrong kind. Whitespace is text and is checked. Raised by "nothing, empty text, or an empty list" in 0035 and "a single space is a legitimate answer" in 0040. Cases: `deferred-and-empty`, `choice-options`. `cheap`
14. **The text `deferred` is plain text.** For a choice it is not an option, and for a boolean it is the wrong kind. Only the Deferred kind is accepted for every type. Raised by "The quoted `"deferred"` is the Text answer with that word" in 0035. Case: `deferred-word`. `cheap`
15. **`CheckValue` is the name of the one-value function, not `ValidateAnswer`.** `ValidateAnswer` was also proposed. 0010 names `CheckValue` in its surface and in its sequences. Raised by "CheckValue(plan, id, value) []Problem" in 0010. Case: every `checks` list. `cheap`
16. **`CheckValue` stops at an unknown or inactive question and runs no secrets check.** The value is refused for that reason alone. Raised by the order in the 0010 sequence. Case: `secrets`. `cheap`
17. **The validator builds the finder and the secrets check itself, with no argument.** The contract tests therefore run the real components. Raised by the dependency order in 0020, which lets the validator depend on both, and by the wiring test rule, which allows one `New()` call. All cases. `cheap`
18. **A pattern mismatch message holds the variable's `patternHint`.** Raised by "shown when a value does not match" in variables.md. The message is otherwise free. Cases: `patterns`, `check-value`. `cheap`
19. **The problems come sorted by question id as bytes.** The docs give no order, and a map of answers has none. Raised by the same plan giving the same result in 0015. Cases: `problem-order`, `secrets`. `cheap`
20. **`section` is the id of a section the plan holds, and the details are `expected`, `found`, `item`, and the directory keys named in the table of problems.** Raised by the Problem fields in the domain model, which leave both open. Cases: most. `cheap`

21. **A pattern is read as RE2, so `$` means the end of the text.** A text that ends in a line break does not match a pattern that ends in `$`. Raised by "`pattern` uses RE2 syntax" in variables.md. Cases: `patterns`. `cheap`
22. **A question is active when the finder returns it for both audiences, and the answer checks do not depend on the audience.** An onboardee answer is checked like an onboarder answer, in `Validate` and in `CheckValue`, and an onboardee question that is hidden, switched off, or left out is inactive like any other. 0055 says a question is active by its own condition, whoever answers it, and does not say which list the validator reads. The editor lists the onboarder questions and the TUI the onboardee questions, and a hand edit can hold both. Cases: `audience-answers`, `stale-answers`, `acme`, `check-value`. `hard`
23. **A stored answer to an onboardee question is not stale, and it is checked.** The old rule treated it as inactive. Cases: `audience-answers`, `stale-answers`. `hard`
24. **An onboarder question that depends on an onboardee answer is active when that answer is stored, even if the stored answer is invalid.** Whether the answer fits its question is a problem of its own, and it does not change which questions are active. Case: `audience-answers`. `cheap`
25. **The plan-wide and directory problems have no `question` and sort first.** They are ordered by `code` as bytes and then by the keys in Order, not by the order of the directory's lists. Raised by the existing order rule, which has no place for a problem with no question. Cases: `directory-order`, `directory-holders`, `directory-warnings`. `cheap`
26. **The directory id check follows the other answer checks.** It applies to active questions only, to onboardee questions like any other, and not to the stored answers of a hidden or switched off question. A wrong kind stops it. Raised by "counts as unanswered everywhere" in the shared decisions. Cases: `directory-answers`. `cheap`
27. **An id is looked up exactly.** Case matters, spaces are not trimmed, the text `deferred` is an id like any other text, and a text of only spaces is an unknown id, while an empty text and an empty list item are unanswered. Raised by "a single space is a legitimate answer" in 0040 and the exact ids of the shared decisions. Cases: `directory-answers`, `directory-check-value`. `cheap`
28. **`directory.unknown-id` never repeats the id in its message or its detail.** Its detail holds `kind` and, for a list, `item`. Raised by the rule that no problem repeats a secret, since an answer to a contact question may hold one. Cases: `directory-answers`, `directory-check-value`. `cheap`
29. **An email is malformed when it is not empty and has no `@` or a space (U+0020), and a contact is unreachable when email, slack, and slackId are all empty.** A malformed email still makes a contact reachable. Raised by "an email with no `@` or with a space" and "a contact with no way to reach them" in 0080. Case: `directory-warnings`. `cheap`
30. **A role whose holders all name no contact is not vacant, and a holder id is compared exactly.** The holders are an error already, so a second problem would repeat it. Holders under a team key for a role that does not exist are still checked. Raised by "a holder that is not a contact" in 0080. Case: `directory-holders`. `cheap`
31. **A duplicated section id gives one `plan.duplicate-section-id` for the id, with the id in `section`.** Raised by "ids are unique within the plan" in 0075. Case: `duplicate-sections`. `cheap`
32. **A list for a `multiple` string or `choice` is not judged against the options or the pattern.** The code reads text only there, and the contract never said what a list does. Raised by "The allowed values of a `choice`" and "A regular expression a `string` must match" in variables.md, which speak of one value, and by lint rejecting `multiple` on these types. Cases: `choice-options`, `patterns`. `hard`
33. **The scope `plan` is the plan's own variables, and an id with no dot is unknown.** The id is read as written. Raised by "It is split at the first dot" in the Interface. Case: `question-ids`. `cheap`

## Wrong implementations tried

Each row names a plausible wrong implementation and the cases or tests that fail it. Each one fails at least one named test.

| Wrong implementation | Caught by |
| -- | -- |
| The kind is not checked, or a `multiple` variable or a `command` takes text, or a boolean takes the text `true` | `wrong-kinds`, `deferred-word`, `check-value` |
| A wrong kind reported as a warning, or `expected` and `found` swapped | `wrong-kinds` |
| A choice is not checked, is compared without case, or accepts the text `deferred` | `choice-options`, `deferred-word`, `check-value` |
| Options enforced on a string, or on a variable with no type | `options-limit-only-choice` |
| A list for a `multiple` choice is flagged as not an option, or a list for a `multiple` string is matched against the pattern as empty text | `choice-options`, `patterns` |
| An id with no dot is read as scope with an empty name, or a section with the id `plan` takes over the plan's questions | `question-ids` |
| A pattern must match the whole text, is read with `(?s)` or `(?m)`, is ignored, or is applied to a `choice` | `patterns`, `check-value` |
| The `patternHint` is left out of the message, or a pattern mismatch is a warning | `patterns`, `check-value` |
| A Deferred value is checked, or the text `deferred` is skipped as if it were Deferred | `deferred-and-empty`, `deferred-word` |
| Empty text, an empty list, or a list of empty text is checked, or only spaces count as unanswered | `deferred-and-empty`, `choice-options`, `check-value` |
| Answers of a hidden, switched off, or left out question are checked | `stale-answers`, `check-value`, `audience-answers` |
| Answers of onboardee questions are not checked, or `CheckValue` calls an onboardee question inactive | `audience-answers`, `stale-answers`, `check-value`, `acme` |
| Onboardee answers get only some answer checks, such as the options but not the kind or the pattern | `audience-answers` |
| An onboardee plan variable is not checked | `audience-answers` |
| Activity read from the finder for one audience only, so onboarder questions or onboardee questions are inactive | `audience-answers`, `check-value` |
| An onboarder question that depends on an onboardee answer is inactive, or an onboardee question that depends on an onboarder answer | `audience-answers` |
| A hidden, switched off, or left out onboardee question is checked, or is not inactive in `CheckValue` | `audience-answers` |
| An onboardee problem has a different severity from the same onboarder problem | `audience-answers` |
| The secrets check is skipped for onboardee questions | `secrets`, `audience-answers` |
| An unknown stored answer is not reported, or is an error in `Validate`, or a submitted unknown question is a warning | `unknown-answers`, `check-value` |
| The secrets check is skipped for inactive or unknown answers, or skipped in `CheckValue` | `secrets`, `check-value` |
| A secrets problem has no `question` or no `section`, loses its `item`, or repeats the secret in its detail | `secrets`, `check-value` |
| The secrets problems come before the answer problems | `secrets`, `problem-order` |
| `CheckValue` accepts an inactive question, or also runs the secrets check on an unknown question | `check-value`, the agreement test |
| A required question with no answer is reported | `not-problems` |
| A default is checked as if it were an answer | `not-problems` |
| The problems are not sorted, or sorted the other way, or depend on the order of the sections | `problem-order`, `secrets`, the section order test |
| `section` always empty, or set to `plan` for a plan variable | `wrong-kinds`, `choice-options`, `check-value` |
| The plan or the value is changed | The not-changed test |
| A result remembered from an earlier call | The no-memory test |
| A `detail` map shared between calls | The new-lists test |
| `directory.unknown-id` is an error in `Validate`, or a warning in `CheckValue` | `directory-answers`, `directory-check-value`, the agreement test |
| An id is compared without case, or trimmed first, or a text of only spaces is skipped | `directory-answers`, `directory-check-value` |
| An id is looked up in every list, so a role id fits a contact question | `directory-answers`, `directory-check-value` |
| Empty list items are flagged, or the `item` starts at one | `directory-answers`, `directory-check-value` |
| The unknown id is in the message or the detail, or the `kind` is missing | `directory-answers`, `directory-check-value`, `directory-order` |
| The directory id check runs for a hidden or switched off question, or after a wrong kind, or the secrets problems come before it | `directory-answers`, `directory-check-value`, `directory-order`, the agreement test |
| A `team` that says `multiple` takes a list, or a single contact or role takes a list, or a multiple contact takes text | `directory-answers`, `directory-check-value` |
| A duplicate id is reported for every use, or the three lists share one namespace | `directory-duplicates` |
| Holders of a role or of a team override are not checked, a role id counts as a holder, or the holders of a team key for an unknown role are skipped | `directory-holders`, `directory-order` |
| A team override for an unknown role is not reported | `directory-holders`, `directory-order` |
| A role is vacant when a team override is empty, or is not vacant when a team names holders for it | `directory-warnings`, `directory-holders` |
| An email is malformed when it is only `@`, or when it is empty, and a slack id does not make a contact reachable, or a malformed email makes a contact unreachable | `directory-warnings`, `directory-answers` |
| Team variables are counted in sections, or one problem for each extra team variable, or the problem is a warning | `two-team-variables`, `team-variable-scopes`, `directory-order` |
| A team's contact is not checked, is a warning, or has another code | `team-contact` |
| A team's contact is reported with the `holder` key, without `team`, or without `team-contact` | `team-contact` |
| A team with no contact or an empty one is reported, or a valid contact is reported | `team-contact`, `directory-answers` |
| A team's contact is compared without case, trimmed, or against the roles or the teams as well | `team-contact` |
| Two teams naming the same wrong contact give one problem, or a duplicated team id gives one | `team-contact` |
| A team's wrong contact hides the problems of its override, or a wrong holder hides the contact problem | `team-contact` |
| The team contact problems keep the order of the directory, sort after the holder problems, or sort by `team-contact` before `team` | `team-contact`, the directory order test |
| A duplicated section id is reported for every extra use, has no `section`, or sorts in another order | `duplicate-sections`, `directory-order` |
| The directory problems keep the order of the directory, come after the answer problems, or sort by severity or by other keys | `directory-order`, `directory-holders`, `directory-warnings`, the directory order test |
| `CheckValue` also reports the directory problems | `directory-check-value`, the agreement test |
