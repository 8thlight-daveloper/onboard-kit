# Question finder contract

Status: draft by an agent. Not reviewed by a person. Component: `question/finder`. Issue: [#68](https://github.com/8thlight-daveloper/onboard-kit/issues/68).

The question finder reads a plan and says what it asks. It returns every active question with the answer stored for it, its declared default, its group, and its suggestions. It also returns the blockers, which are the required questions that have no answer. Both take an audience, the person who answers: the onboarder, the onboardee, or both. It is a pure function. It reads no file, reads no clock, keeps no state between calls, and does not change its argument. See [0055](../../../docs/decisions/0055-variable-audience.md), [0010](../../../docs/decisions/0010-plan-editor.md), [0045](../../../docs/decisions/0045-defaults-are-suggestions.md), [0050](../../../docs/decisions/0050-section-switch.md), [0070](../../../docs/decisions/0070-sub-sections.md), [0035](../../../docs/decisions/0035-settled-and-deferred-answers.md), [0080](../../../docs/decisions/0080-directory.md), the [domain model](../../../docs/reference/domain-model.md), and the [section template schema](../../../docs/reference/section-template-schema/README.md).

A question of type `contact`, `role`, or `team` picks from the plan's directory. The finder lists the directory's entries as the question's choices, and it reads an answer that names an id the directory does not hold as no answer.

A plan holds every template as a section and every variable of the plan and its sections. The finder works out the active questions from the plan whenever it is asked. It finds the active sections with the section selector and evaluates `askIf` with the condition evaluator.

## Interface

`Questions(plan, audience) -> list of Question` and `Blockers(plan, audience) -> list of Blocker`.

- `plan` is a Plan, as stored. The finder reads `plan.Variables`, `plan.Answers`, `plan.Sections`, and `plan.Directory`. It reads no other field.
- `Question` and `Blocker` are the model types. A Question has `ID`, `Scope`, `Name`, `Description`, `Example`, `Type`, `Options`, `Choices`, `Required`, `Multiple`, `Audience`, `Answer`, `DeclaredDefault`, `Group`, and `Suggestions`.
- Both results are empty when the plan has no active question. There is no error and no problem list. The finder evaluates the plan as written, and the validators report plans that cannot do their job.
- `audience` is a `model.Audience`: `onboarder`, `onboardee`, or the empty value, which means both. An empty audience lists every active question in one list. A named audience lists the active questions that variables of that audience declare. A driver that wants the questions and the blockers calls both, with the same audience.
- Both functions take the plan and the audience and nothing else.

## Rules

**The questions that can be asked.** In this order, which is the candidate order:

1. Every plan variable, in declared order. Its scope is `plan`.
2. Then, for each section the section selector returns, in the order it returns them, every variable of the section in declared order. Its scope is the section's `id`.

A section the selector leaves out has no candidate. It is out because its `includeIf` fails or its switch is off, as [0050](../../../docs/decisions/0050-section-switch.md) says, or because it is a sub-section that is not active, as the next paragraph says. A variable of either audience is a candidate. A variable with no `audience` is an onboarder variable. The finder lists no question for a section's switch: the switch is the `Enabled` field of the plan section, and it is not a question, so no `<section>.enabled` question exists. A candidate's id is `<scope>.<name>`.

**Sub-sections.** A section may name a parent section, as [0070](../../../docs/decisions/0070-sub-sections.md) says. The finder does not read `parent`. It takes the sections the selector returns, in the order it returns them. So the questions of a sub-section come directly after the questions of its parent, and the questions of a section below the sub-section come after the sub-section's, in the order the selector gives for the whole subtree. A question of a sub-section the selector leaves out is not a candidate, whatever the reason: its parent is out or switched off, its own switch is off, its own `includeIf` fails, its parent is not a section of the plan, or its parent chain is a cycle. Such a question is not returned and does not block, and an answer stored for it stays in the plan. The question of an active sub-section is active by its own `askIf` like any other, and it blocks when it is required and has no answer. Whether a sub-section is active is the selector's rule and not restated here: the finder shows it only through which questions it returns. The selector reads the answers as stored, so a stored answer of a parent's question counts for the sub-section's `includeIf` even when that question is not asked, while an `askIf` still reads only the answers of active questions, as the next paragraphs say.

**Switches.** The finder builds a Switches value from every section of the plan, active or not: the section's `id` maps to false when its `Enabled` is false, and to true when it is true or absent. It passes this to the condition evaluator for every `askIf`. A sub-section's value is its own switch only: a switched-off parent does not change the value of its sub-sections, so a section clause on a sub-section whose parent is off tests that sub-section's own switch.

**Which candidates are active.** A candidate with no `askIf` is active. Otherwise the finder asks the condition evaluator, `Holds(askIf, scope, answers, switches)`, with the candidate's scope as the scope and the answers described next. The candidate is active when the condition holds. An active question keeps its stored answer and is returned. A candidate whose condition does not hold is not returned and does not block. It keeps its stored answer in the plan.

The answers passed to the evaluator hold only the stored answers of the questions found active before this one, whoever answers them. For a directory question the answer passed is the answer when it is found, with the items that are not in the directory removed from a list, and no answer when it is not found. The answer of any other question counts as unanswered: a question whose condition fails, a question of a section that is out, an answer stored for a name that no candidate declares, the question itself, and a question not decided yet. This is the rule in [conditions](../../../docs/reference/section-template-schema/conditions.md): a settled answer is one for a question that is active. A declared default is not in the answers, so it never counts.

**Audience.** A question is active by its own rules, whoever answers it, as [0055](../../../docs/decisions/0055-variable-audience.md) says. So an onboarder question can depend on an onboardee answer and the other way round, and the finder decides activity and order the same way for every audience. The audience then only chooses what is returned. The questions for an empty audience are the active questions as the rules above give them. The questions for a named audience are those of the empty audience whose variable has that audience, in the order they have there. Each question keeps the group, the suggestions, and the answer it has in the full list. A condition never reads the audience.

**Order of deciding.** A question names another question when a question clause of its `askIf` has that question's id. A bare name means the question in the candidate's own scope, and a name with a dot is split at the first dot, as the evaluator reads it. A section clause names no question. The finder decides the candidates one at a time:

1. A candidate is ready when every other candidate it names has been decided. A name that is not a candidate, and the candidate's own id, are not waited for.
2. The next candidate decided is the first ready one in candidate order.
3. When none is ready, the condition names a cycle. The next candidate decided is the first undecided one in candidate order. Its condition is read as written, with every undecided question counted as unanswered. No question is dropped because of a cycle. The plan validator reports the cycle.

**The order of the questions.** The active questions are put in order in a second pass, so a question is never asked before a question its condition names. The first pass only decides which questions are active. The second pass places one question at a time:

1. The first ready active question in candidate order goes next. A question is ready when every other active question it names has been placed. A name that is not an active question is not waited for, and neither is the question's own id.
2. When none is ready, the first unplaced active question in candidate order goes next, as above.

So a question that does not name a later question stays where it is in candidate order, and a question that names a later one moves to just after the last question it needs. A question that names a question that is not asked has no constraint from it. The result is the same for the same plan every time.

**What a question holds.**

- `ID`, `Scope`, and `Name` as above.
- `Description`, `Example`, `Options`, `Required`, and `Multiple` are the variable's own. `Example` is one illustration of a good answer, and it is empty when the variable has none. Each question keeps the `Example` of its own variable, so two questions in one group can hold different examples. It is never an answer, never a default, and never a suggestion. It plays no part in the group key, activity, the order, the suggestions, or the blockers. A variable with `options` and a type other than `choice`, `contact`, `role`, or `team` keeps its options.
- `Type` is the variable's type, or `string` when it has none. `Audience` is the variable's audience, or `onboarder` when it has none, so it is never empty.
- `Answer` is the answer stored for the question, exactly as stored: a settled value, a Deferred value, empty text, an empty list, or a value of no kind when nothing is stored. It is never filled from the default.
- `DeclaredDefault` is the variable's `default`, exactly as declared, and a value of no kind when it has none. It is not an answer. A question with a default and no answer has no answer and blocks when it is required, as [0045](../../../docs/decisions/0045-defaults-are-suggestions.md) says.

**Directory questions.** A question whose type is `contact`, `role`, or `team` picks from the plan's directory, as [0080](../../../docs/decisions/0080-directory.md) says.

- `Choices` holds the entries of the directory list for the type: `contacts` for a contact, `roles` for a role, `teams` for a team. There is one choice for each entry, in the order of the directory. Its `Value` is the entry's id and its `Label` is the entry's name. A plan with no directory, or an empty list for the type, gives no choices.
- `Options` is empty, whatever the variable declares. A variable of any other type has no choices, including `choice`, and its `Choices` is nil, not an empty list.
- The choices are the same for every question of a type, whatever its scope, audience, or answer.
- An id is in the directory when an entry of the list for the question's type has exactly that id. Ids are compared as written, so case and spaces count. A contact id is not a role id, and a role id is not a team id.
- An answer is a Text id, or a List of ids when the question is `multiple`. The answer is **found** when its id is in the directory. A List is found when at least one of its items is in the directory. A Deferred answer is found. Everything else is not found: an id that is not in the directory, empty text, an empty list, a list with no id in the directory, and nothing stored.
- A question whose answer is not found counts as unanswered, as [0080](../../../docs/decisions/0080-directory.md) says. It blocks when it is required. A condition reads it as unanswered.
- `Answer` is still the stored answer as it is. The finder keeps an answer that names an unknown id, so an interface can show it, and it does not remove the unknown items of a list.
- A directory question is active by the same rules as any other question, and the audience, the order, the groups, and the blockers of the other types hold for it.

**Groups.** `Group` is a text key. Two returned questions have the same key exactly when they have the same name, the same type, the same options, and the same `multiple`. Options are compared as a set, so their order and repeats do not matter. The description, the pattern, `required`, the default, and the scope play no part, so a plan variable and a section variable with the same name and type share a group. The key is not empty. Its text is not specified, and only equal or different matters. The finder only reports the group. It never merges anything.

**Suggestions.** A question's suggestions are a list built in two parts, defaults first and then given answers. They do not depend on the audience asked for.

1. **Defaults.** For each returned question in the question's group, in the order of the questions, its declared default. The group is made of the questions of the full list, whatever their audience, so the default of a variable in a section that is out, or of a question whose condition fails, is not offered.
2. **Given answers.** The answers stored for the question's name, taken from the plan's answers first and then from each section's answers in the order of the section ids, compared as bytes. Every section's answers count, whether the section is active or not, and whether or not a question for that answer is returned. This includes the answer of the question itself.

A candidate is offered only when it fits the question:

- It is settled. A deferred answer, empty text, and a list with no non-empty item are never offered. Deferred has its own entry in an interface, and it is not a suggestion.
- A `boolean` offers a Flag. A `command`, or any question with `multiple` true, offers a List. Every other type offers Text.
- A `choice` offers only Text that is one of its options.
- A `contact`, a `role`, or a `team` offers only an answer made of ids that are in the directory for its type: Text for one id, and a List for a `multiple` question in which every item is in the directory. A default or a given answer that names an id outside the directory, or that mixes known and unknown ids, is not offered. A contact id given as the answer of a role question of the same name is not offered to it.
- A `string` with a `pattern` offers only Text the pattern matches. The pattern uses RE2 syntax and matches anywhere in the text, so a pattern that must match the whole text starts with `^` and ends with `$`.

Two values are equal when they have the same kind and the same content. A list equals a list with the same items in the same order. Among the defaults, a value that equals an earlier one is dropped, and the same holds among the given answers. A value that is both a default and a given answer is listed twice, once with source `default` and once with source `given`. The suggestions of two questions in one group can differ, because the type is the same and the pattern and the options need not be.

**Blockers.** A blocker is an active question that is required and has no answer. The blockers for a named audience are those of its questions, so a required onboardee question does not block the onboarder, and the other way round. The blockers for an empty audience are those of both, in the order of the full list. A required question is satisfied by a settled or a deferred answer, as [0035](../../../docs/decisions/0035-settled-and-deferred-answers.md) says. These do not satisfy it: nothing stored, empty text, an empty list, and a list of only empty text. For a directory question an answer that is not found does not satisfy it, as the rules above say. A false Flag is settled. A default is not an answer. The blockers are listed in the order of the questions, and each holds the question's id. A question that is not active never blocks, whether its section is out, its switch is off, or its condition fails. An answer already stored for a question of either audience counts, whoever put it there.

## Properties

These hold for every plan the contract tests use.

- **Pure.** The plan is the same after the calls as before.
- **No memory.** The same plan gives the same result on a shared finder, whatever calls came before, forward or in reverse order, and however often it is repeated.
- **Same plan, same result.** For every audience, rotating or reversing `plan.Sections` does not change the questions, their suggestions, or the blockers.
- **Blockers agree with the questions.** For every audience, the blockers are the returned questions that are required and have no answer, in the same order. An answer is none when it is neither settled nor deferred, and, for a directory question, when it is not found in the directory.
- **A named audience is a filter.** The questions of a named audience are the questions of the empty audience that have that audience, unchanged and in the same order. Their blockers are the blockers of the empty audience for those questions.
- **New lists.** Changing, reordering, or replacing an item of a result does not change the plan or a later result. The lists inside a Question, such as its options and its suggestions, may share their data with the plan.
- **Empty plan.** A plan with nothing in it gives no question and no blocker, for every audience.

## Side effects and plan access

None. It reads the plan's variables, the answers, the directory, and the content of the sections it needs: `id`, `includeIf`, `parent`, `timing`, `after`, `Enabled`, and `variables`. It returns no plan. This matches the row for `question/finder` in [components](../../../docs/architecture/components.md).

## Not covered

- Duplicate section ids, and duplicate variable names in one scope.
- Whether a section's `parent` names a section of the plan, whether the parent chain has a cycle, and whether an `includeIf` or an `askIf` names only an ancestor's questions and plan variables. The plan validator's structure check reports them. The finder reads such a plan as the selector returns it and never hangs.
- A `timing` that is not one of the five names, a `type` that is not one of the five types, and an `audience` on a variable, or an audience asked for, that is none of the values above.
- A `pattern` that does not compile. The validators report it.
- An answer stored under a section id the plan does not hold, or under a name that no variable of the plan declares. Whether such an answer is offered as a given answer is left open.
- Whether the plan variables, as a whole, are valid. A plan variable with an `askIf` is evaluated like any other, and the plan validator reports it. A plan variable's `askIf` does not change which sections the selector finds active, because the selector reads the stored plan answers.
- A condition that names a question that does not exist, a section clause that names no section, or a clause with both kinds of key. The validators and the decoder report them, and the evaluator reads them as written.
- A default that does not fit its variable's type or pattern. The validators report it, and the finder only decides whether to offer it.
- Whitespace only text as an answer or a default.
- Whether a choice or a boolean answer is valid. That is the plan validator's check.
- The interface's own entries, such as `deferred`. A suggestion is never `deferred`.
- Duplicate ids in the directory, a holder or a team override that names no contact or role, and an id that is not a lowercase slug. The plan validator reports them. The finder lists every entry as written and does not read holders.
- Whether a directory answer has the right shape: a List for a question that is not `multiple`, Text for one that is, a List for a `team`, or `multiple` on a `team`. The finder treats Text as one id and a List as its items.
- Whether a section's `includeIf` that names a directory question reads an unknown id as unanswered. The section selector decides, and the finder passes the plan as stored.
- A directory entry with an empty name. Its choice has an empty label.

## Decisions to know

The docs were silent, or could be read two ways, in these places. A choice marked `hard` changes what a person sees in the interface, what a plan file holds, or what a template author can rely on. A choice marked `cheap` touches one fixture and nothing outside the component.

01. **A condition counts only the stored answers of questions that are active.** An answer that is kept in the plan but belongs to a question hidden by its own `askIf`, to a section that is switched off or left out, or to no declared question reads as unanswered. An answer to an onboardee question counts like any other when that question is active. So a follow-up never starts because of a stale answer. Raised by "A condition is true only when every question it tests has a settled answer" in 0010 and by "The question is active and the plan stores an answer for it" in conditions.md, against the evaluator contract's note that the caller decides which answers to pass. Case: `ask-if-inactive-answers`. `hard`
02. **A question moves only past a question that will be asked, and only as far as it must.** The order is candidate order, with a question placed just after the last active question its condition names. Several needs, chains, and a need in an earlier section all work the same way. A question that names a question that is not asked does not move. Raised by "the finder moves the dependent question to just after the question it needs", which says nothing about chains, several needs, or a question that is never asked. Cases: `order-condition-names-later-section`, `order-condition-names-later-question`, `order-ignores-questions-not-asked`. `hard`
03. **The defaults offered are those of the returned questions in the group, and the given answers are those of every scope.** A default in a section that is out, or in a question hidden by its condition, is not offered, but an answer stored there is. Raised by "the distinct declared defaults in its group, plus the answers already given for that name in any scope ... inactive ones included", which says "inactive" only for the answers. Cases: `suggestions-defaults`, `suggestions-given`. `hard`
04. **A value that is both a default and a given answer is listed twice, defaults come first, and given answers go in the order plan, then section id.** Raised by "the distinct declared defaults in its group, plus the answers already given", which counts distinct values for the defaults and gives no order. The Suggestion type has one source, so the two sources stay apart. Cases: `suggestions-default-and-given`, `suggestions-given`. `hard`
05. **A suggestion must fit the question, and deferred and empty values are never suggestions.** A boolean offers flags, a command and a `multiple` question offer lists, a choice offers its options, and a string with a pattern offers matching text. Raised by "For a `boolean` or `choice` only valid values are offered", which says nothing about the other types, and by "`deferred` is always offered, as its own entry", which says nothing about the suggestion list. Cases: `suggestions-valid-values`, `suggestions-settled-only`, `suggestions-lists`. `hard`
06. **A required question with only a default blocks, and the answer is never filled in.** Raised by 0045, "A question with a default and no answer is unanswered". Cases: `defaults-are-suggestions`, `blockers`. `hard`
07. **Questions in a condition cycle are decided in candidate order, each seeing the undecided as unanswered, and none is dropped.** A question that is part of a cycle only stays if its own condition holds under that rule. The questions that still wait on each other go after the ones that do not. Raised by "A cycle is reported as a problem and not resolved". The plan is invalid, so a person sees this only in a plan the validator rejects. Case: `ask-if-cycle`. `cheap`
08. **The group key is unspecified text, and options count as a set.** Two choices with the same options in another order are one group. Raised by "Questions with the same name, type, options, and `multiple` form a group". The model says the key is "shared", not what it holds. Case: `groups`. `cheap`
09. **`Type` is `string` and `Audience` is `onboarder` when the variable says nothing.** The model's Question has both as fields that are not optional. Raised by "`string` by default" and "`onboarder` by default" in variables.md. Case: `kinds-and-defaults`. `cheap`
10. **A pattern matches anywhere in the text.** A pattern that must match all of it carries its own `^` and `$`, as the ACME patterns do. Raised by "A regular expression a `string` must match". Case: `suggestions-pattern`. `cheap`
11. **`Answer` is the stored answer as it is, including empty text, an empty list, and deferred.** Raised by the model's "The stored answer, if any". Case: `answers-as-stored`. `cheap`
12. **Options belong to a question whatever its type, and only a `choice` is limited by them.** The ACME plan lists options on `team_state` without `type: choice`, so its question is a string that carries its options. Raised by "`options` ... The allowed values of a `choice`". Case: `acme`. `cheap`
13. **Plan variables take part like any other variable.** An `askIf` on one is evaluated, and it can be hidden. Raised by "It should not use `askIf`, and the validator reports one that does". Case: `ask-if-tests`. `cheap`
14. **Blockers come in the order of the questions.** Raised by "required questions with no answer" with no order. Case: `blockers`. `cheap`

15. **An empty audience means both, and it lists the questions in one list in the same order as before.** 0055 says the finder takes an audience and lists the active questions for one audience, and does not say how to ask for both. The empty value is the zero value of `model.Audience`, and the plan validator needs the full list. Cases: every case, as the full list. `hard`
16. **A named audience is a filter of the full list, so the order does not change.** A question that waits for a later question stays after it in the full list, and a named list shows its own questions in the order they have there. 0055 says only "lists the active questions for one audience". Cases: `audience-ask-if`, `audience-split`. `hard`
17. **Activity does not depend on the audience asked for.** An onboarder question whose `askIf` names an onboardee answer is active when that answer is stored and the condition holds, even for `Questions(plan, onboarder)`. An answer to a question that is not active reads as unanswered, whichever audience it has. 0055 says "A question is active by its own condition, whoever answers it". Cases: `audience-ask-if`, `ask-if-inactive-answers`. `hard`
18. **Suggestions and groups draw on every question and every stored answer, whatever the audience.** A default of an onboardee question is offered to an onboarder question of the same group, and an answer stored for an onboardee question is a given answer. The model has one group key and one suggestion list for a name. Case: `audience-suggestions`. `hard`
19. **A required question with an answer already in the file does not block, whoever put the answer there.** An onboarder who fills in an onboardee answer settles it. 0055 says "An answer already in the file is respected". Case: `audience-blockers`. `cheap`
20. **An audience that is none of the three values is not covered.** The finder may return no question or treat it as both. `cheap`
21. **A question's `Example` is its own variable's, and nothing else reads it.** It is never an answer, so a required question with only an example blocks, and it is never offered as a suggestion. The model docs do not describe the field yet, so the rule comes from the change that added it. Case: `examples`. `cheap`
22. **A sub-section's `includeIf` sees the stored answer of a parent's question that is not asked, and an `askIf` does not.** A parent question hidden by its own `askIf` that has a stored answer can still open its sub-section, because the selector reads the answers as stored. A question of that sub-section whose `askIf` names the hidden question reads it as unanswered, because the finder counts only the answers of active questions (decision 01). Raised by 0070, "A sub-section's `includeIf` may name questions of its ancestors", which does not say whether a hidden ancestor question counts, against "a settled answer is one for an active question" in conditions.md. Case: `subsection-gate-hidden-parent-answer`. `hard`
23. **The Switches value of a sub-section is its own switch.** A switched-off parent makes its sub-sections inactive but does not turn their values false in the Switches the finder passes, so a section clause still tests the section's own switch. Raised by 0070, "Switches cascade", which says what a switched-off parent does to activity and not to a section clause. Case: `subsection-switch-clause`. `cheap`
24. **Plans with a missing parent or a cycle give the questions of the sections that stay active.** The finder returns what the selector returns and has no error for them. Raised by 0070, "no cycle is possible", which holds only for a plan that passed the validator. Case: `subsection-bad-parent`. `cheap`

25. **A `multiple` directory question is answered when at least one of its ids is in the directory.** A list of one known and one unknown id satisfies a required question, and a list with no known id blocks. The other choice was that every id must be known. The answer keeps both ids, and the interface can show the unknown one. 0080 says an answer that names an unknown id counts as unanswered and does not say what a list that mixes known and unknown ids is. Cases: `directory-multiple`, `directory-ask-if`. `hard`
26. **A condition reads an unknown id as unanswered, and a list as only its known ids.** `isSettled: false` holds for it, and `is` and `isNot` do not, so a follow-up that depends on a contact does not start from an id that is gone. The other choice was that a condition reads the stored answer. 0080 says it counts as unanswered "everywhere" in the sentence about the finder blocker and the resolver, and does not name conditions. It follows decision 01, that a condition reads only what is settled. Cases: `directory-ask-if`, `directory-audience`. `hard`
27. **`Answer` keeps the stored id, and the question blocks or not on its own.** The finder does not blank an unknown id or drop it from a list, so the interface can show what is stored and mark it. The other choice was an empty answer, which would hide the id. The model says `Answer` is "the stored answer, if any". Cases: `directory-unknown-answers`, `directory-multiple`. `hard`
28. **A default or a given answer is offered only when every id in it is in the directory for the question's type.** A list that mixes known and unknown ids is not offered, so the offer is wholly valid, where an answer needs only one known id (decision 25). A contact id stored under the name of a role question is not offered to it. The other choice was to offer the known items of a list. It follows 0045, "a choice offers only valid values", and 0080 does not say. Cases: `directory-suggestions`, `directory-unknown-answers`. `hard`
29. **A directory question has choices and no options, and nothing else has choices.** A variable of type `contact`, `role`, or `team` that declares `options` has them dropped, since the directory is the list. A `choice` and a string with options keep their options and have no choices. 0080 says "the walkthrough offers the directory's entries" and the `team` question's "options are the directory's teams", and the model has a separate `Choices` field. Case: `directory-choices`. `cheap`
30. **Ids are compared as written, and each type has its own list.** `Sam-Okafor` and ` sam-okafor` are not `sam-okafor`, and a role id is no answer to a contact question. 0080 gives ids as permanent and does not say how they are compared. Cases: `directory-unknown-answers`, `directory-contacts-only`. `cheap`
31. **An optional directory question with an unknown id does not block, and empty text and a deferred answer behave as before.** Raised by "so a required question blocks until it is filled" in 0080. Case: `directory-unknown-answers`. `cheap`
32. **A question that is not a directory question has a nil `Choices`, not an empty list.** The docs say only that it has no choices. A nil list and an empty list read the same by length, so a reader could take either. Nil says that the question has no choices at all, and an empty list is left for a directory question whose list is empty, where either is accepted. Every case that holds a question of another type checks it. `hard`

## Wrong implementations tried

Each row names a plausible wrong implementation and the groups of cases that fail it. Each one with a defect in the finder's own output fails at least one named test.

| Wrong implementation | Caught by |
| -- | -- |
| `askIf` ignored, so every question is asked | `ask-if-tests`, `ask-if-scope`, `blockers`, `ask-if-section-clause` |
| A condition reads every stored answer, including those of hidden questions | `ask-if-inactive-answers`, `ask-if-cycle` |
| A condition ignores the answers of onboardee questions | `ask-if-inactive-answers`, `audience-ask-if` |
| An onboarder question that names an onboardee answer is dropped, or the other way round | `audience-ask-if`, `audience-section-choice` |
| Switches not passed to the evaluator, built from the active sections only, or read the wrong way round for a missing `enabled` | `ask-if-section-clause` |
| A question listed for each section's switch | most cases, for example `inactive-sections` |
| Plan variables listed after the sections, or sections in plan order | `reading-order`, `ask-if-section-clause`, `suggestions-given` |
| Questions of every section asked, or of a switched-off section asked | `inactive-sections`, `blockers`, `acme` |
| Onboardee questions left out of the full list, or left out of a named audience that has them | `audience-split`, `acme`, `blockers` |
| Questions sorted by name, or declaration order not kept | `reading-order`, `ask-if-scope` |
| No reordering for conditions | `order-condition-names-later-section`, `order-condition-names-later-question` |
| A dependent moved to the end, or after a whole section | `order-condition-names-later-question`, `ask-if-tests` |
| Activity and order decided in one pass, so a question waits for one that is never asked | `order-ignores-questions-not-asked` |
| A cycle drops its questions, or breaks at the last question, or makes all of them asked | `ask-if-cycle` |
| A question that names itself waits for itself | `ask-if-cycle` |
| A section clause hides the question clauses after it from the order, or counts as naming a question with an empty name | `order-section-clause-then-question-clause`, `order-section-clause-names-no-question` |
| A bare name read in the plan scope, in the condition or in the order | `ask-if-scope`, `order-condition-names-later-question` |
| Default used as the answer, or as a condition's answer | `defaults-are-suggestions`, `acme`, `ask-if-tests` |
| A default satisfies a required question | `defaults-are-suggestions` |
| Deferred does not satisfy a required question | `blockers`, `acme` |
| An empty text, an empty list, or a list of empty text satisfies a required question | `blockers`, `answers-as-stored` |
| A false flag does not satisfy a required question | `blockers` |
| Optional questions block, or hidden questions block | `blockers`, `ask-if-inactive-answers` |
| Blockers reversed or sorted by id | `blockers` |
| Empty or deferred answers hidden in the question | `answers-as-stored`, `blockers`, `acme` |
| Answer, declared default, description, options, `multiple`, `required`, or scope left out | `acme`, `kinds-and-defaults`, `blockers`, `suggestions-defaults` |
| `Type` or `Audience` left empty when the variable has none, or the audience taken from the audience asked for | `kinds-and-defaults`, `audience-split` |
| No given answers, or no defaults, in the suggestions | `suggestions-given`, `suggestions-defaults`, `ask-if-scope` |
| Given answers from active sections only, or from the question's own scope or group only | `suggestions-given`, `suggestions-valid-values` |
| The plan's own answer left out of the suggestions, or the question's own answer left out | `ask-if-scope`, `suggestions-given`, `defaults-are-suggestions` |
| Defaults from every section, or from hidden questions, or from the question alone | `suggestions-defaults` |
| A value that is a default and a given answer listed once, or no distinct at all | `suggestions-default-and-given`, `suggestions-defaults`, `suggestions-given` |
| Given answers before defaults, in map order, or in plan section order | `suggestions-default-and-given`, `suggestions-given` |
| A choice offers values outside its options, a boolean offers text, a list offers text | `suggestions-valid-values`, `suggestions-lists` |
| A `multiple` question of type string or choice offers Text, or a text default of a `multiple` question is offered | `suggestions-lists-and-flags` |
| Empty text or a value of no kind offered for a choice that lists empty text or a pattern that matches empty text | `suggestions-empty-text-fits` |
| Flags equal whatever their value, so a false Flag is dropped after a true one | `suggestions-lists-and-flags` |
| A pattern ignored, or matched against the whole text | `suggestions-pattern` |
| Deferred and empty answers offered | `suggestions-settled-only`, `answers-as-stored` |
| Group key ignores the type, the options, `multiple`, or the name | `groups`, `suggestions-valid-values`, `suggestions-lists` |
| Group key depends on the options' order, the description, `required`, or the scope | `groups` |
| Group key empty | `groups` and most cases |
| A named audience returns every question, or returns the other audience, or the full list when the audience is empty | `audience-split`, the named-audience test |
| A named audience is sorted apart from the full list, or its questions are put in declared order instead of the order of deciding | `audience-ask-if`, the named-audience test |
| The audience asked for changes which questions are active, or which answers a condition sees | `audience-ask-if`, the named-audience test |
| Blockers of a named audience hold the other audience's questions, or those of both | `audience-blockers`, `audience-unanswered-choice`, `audience-ask-if` |
| An empty audience gives the blockers of one audience only | `audience-blockers`, `audience-split` |
| A required onboardee question with an answer in the file, or a deferred one, still blocks | `audience-blockers` |
| A hidden, switched off, or left out onboardee question blocks | `audience-blockers`, `audience-section-choice` |
| Suggestions and groups built from the questions of the audience asked for only | `audience-suggestions`, the named-audience test |
| A section that depends on an onboardee answer is in when the answer is missing, or out when it is given | `audience-section-choice`, `audience-unanswered-choice` |
| An onboardee plan variable is skipped, or an onboardee section variable is skipped | `audience-split`, `audience-section-choice` |
| An argument changed: the answers or the order of the sections | The plan-unchanged test |
| An earlier result returned on a later call | The no-memory test, and most cases |
| Example left out of the question, or empty when the variable has one | `examples` |
| Example taken from another variable of the group, such as the first one or the last one | `examples` |
| Example used as the answer, or as a default, or satisfies a required question | `examples` |
| Example offered as a suggestion | `examples` |
| Example added to the group key, so the questions of a group split | `examples` |
| A result list shared between calls | The result-list test |
| Sub-sections' questions listed after every root's, or the plan's section order used instead of the selector's | `subsection-order`, `subsection-order-timing`, `subsection-grandchild-open` |
| Sub-sections sorted among the roots by timing or id, so a sub-section does not follow its parent | `subsection-order`, `subsection-order-timing` |
| Siblings of a sub-section in plan order, or `after` among them ignored | `subsection-order`, `subsection-grandchild-open` |
| Questions of a sub-section of a switched-off section returned, or only the parent's own switch checked | `subsection-parent-off`, `subsection-grandchild-middle-out`, `subsection-switch-clause` |
| Questions of a sub-section whose own switch is off, or whose own `includeIf` fails, returned | `subsection-own-gates`, `subsection-gate-closed`, `subsection-gate-unanswered`, `subsection-grandchild-own-gates-closed` |
| A sub-section's `includeIf` that names its parent's answer ignored, or read as unanswered | `subsection-gate-open`, `subsection-grandchild-open` |
| A sub-section's `includeIf` that names a parent question that is not asked read as unanswered | `subsection-gate-hidden-parent-answer` |
| A question's `askIf` that names a hidden question reads the stored answer | `subsection-gate-hidden-parent-answer` |
| A sub-section with a missing parent treated as a root, or a section on a cycle or below one returned | `subsection-bad-parent` |
| The Switches value of a sub-section made false by its parent | `subsection-switch-clause` |
| Blockers of a sub-section that is out, or of one below a switched-off section | `subsection-gate-closed`, `subsection-parent-off`, `subsection-bad-parent` |
| Blockers of a sub-section that is in left out | `subsection-gate-open`, `subsection-grandchild-open` |
| Choices sorted by name or id, or the directory order not kept | `directory-choices`, `directory-suggestions` |
| The label is the id, or the value is the name | `directory-choices` |
| Contacts offered for a role or a team, or one list for all three types | `directory-choices`, `directory-contacts-only` |
| Choices for every question, or for a `choice` or a string with options | `directory-choices`, `directory-unknown-answers` |
| An empty list of choices, not nil, on a question that is not a directory question | the choices test, on every case |
| Choices left out for onboardee questions, or built only for onboarder questions | `directory-audience` |
| Ids copied into `Options`, or the declared options of a directory question kept | `directory-choices` |
| Choices missing when the directory has no entry of that type, or from a plan with no directory | `directory-contacts-only`, `directory-empty` |
| An unknown id satisfies a required question | `directory-unknown-answers`, `directory-empty`, `directory-contacts-only` |
| An id of another list satisfies the question: a contact id for a role, a role id for a team | `directory-unknown-answers`, `directory-multiple` |
| Ids compared ignoring case, or after trimming | `directory-unknown-answers` |
| An answer is blanked or trimmed of its unknown ids in `Answer` | `directory-unknown-answers`, `directory-multiple`, `directory-ask-if` |
| A list needs every id to be known, or only its first id, or any non-empty item | `directory-multiple` |
| An optional directory question blocks, or a deferred one blocks | `directory-unknown-answers`, `directory-multiple` |
| No check at all when the directory is empty | `directory-empty`, `directory-contacts-only` |
| A condition reads the stored unknown id, so `is` holds for it or `isSettled` is true | `directory-ask-if`, `directory-audience` |
| A condition ignores every directory answer, so a known id counts as unanswered | `directory-ask-if`, `directory-audience` |
| An unknown id is offered as a default or a given answer | `directory-suggestions`, `directory-unknown-answers` |
| A list that mixes known and unknown ids is offered | `directory-suggestions`, `directory-multiple` |
| A suggestion checked against the ids of every list and not of the question's type | `directory-suggestions` |
| The group key split by scope for a directory question | `directory-suggestions` |
| The choices of one question changed by editing the plan's directory in place | the plan-unchanged test, the named-audience test |
| Repeated options split a group, or empty and absent options split it | `group-option-sets` |
| Options containing commas collide with other option sets in the group key | `group-option-sets` |
| Suggestions sort list items or drop empty items before comparing them | `list-suggestion-content` |
