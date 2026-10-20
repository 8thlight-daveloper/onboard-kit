# Question finder contract

Status: draft by an agent. Not reviewed by a person. Component: `question/finder`. Issue: [#68](https://github.com/8thlight-daveloper/onboard-kit/issues/68).

The question finder reads a plan and says what it asks. It returns every active question with the answer stored for it, its declared default, its group, and its suggestions. It also returns the blockers, which are the required questions that have no answer. Both take an audience, the person who answers: the onboarder, the onboardee, or both. It is a pure function. It reads no file, reads no clock, keeps no state between calls, and does not change its argument. See [0055](../../../docs/decisions/0055-variable-audience.md), [0010](../../../docs/decisions/0010-plan-editor.md), [0045](../../../docs/decisions/0045-defaults-are-suggestions.md), [0050](../../../docs/decisions/0050-section-switch.md), [0035](../../../docs/decisions/0035-settled-and-deferred-answers.md), the [domain model](../../../docs/reference/domain-model.md), and the [section template schema](../../../docs/reference/section-template-schema/README.md).

A plan holds every template as a section and every variable of the plan and its sections. The finder works out the active questions from the plan whenever it is asked. It finds the active sections with the section selector and evaluates `askIf` with the condition evaluator.

## Interface

`Questions(plan, audience) -> list of Question` and `Blockers(plan, audience) -> list of Blocker`.

- `plan` is a Plan, as stored. The finder reads `plan.Variables`, `plan.Answers`, and `plan.Sections`. It reads no other field.
- `Question` and `Blocker` are the model types. A Question has `ID`, `Scope`, `Name`, `Description`, `Type`, `Options`, `Required`, `Multiple`, `Audience`, `Answer`, `DeclaredDefault`, `Group`, and `Suggestions`.
- Both results are empty when the plan has no active question. There is no error and no problem list. The finder evaluates the plan as written, and the validators report plans that cannot do their job.
- `audience` is a `model.Audience`: `onboarder`, `onboardee`, or the empty value, which means both. An empty audience lists every active question in one list. A named audience lists the active questions that variables of that audience declare. A driver that wants the questions and the blockers calls both, with the same audience.
- Both functions take the plan and the audience and nothing else.

## Rules

**The questions that can be asked.** In this order, which is the candidate order:

1. Every plan variable, in declared order. Its scope is `plan`.
2. Then, for each section the section selector returns, in the order it returns them, every variable of the section in declared order. Its scope is the section's `id`.

A section the selector leaves out has no candidate. It is out because its `includeIf` fails or its switch is off, as [0050](../../../docs/decisions/0050-section-switch.md) says. A variable of either audience is a candidate. A variable with no `audience` is an onboarder variable. The finder lists no question for a section's switch: the switch is the `Enabled` field of the plan section, and it is not a question, so no `<section>.enabled` question exists. A candidate's id is `<scope>.<name>`.

**Switches.** The finder builds a Switches value from every section of the plan, active or not: the section's `id` maps to false when its `Enabled` is false, and to true when it is true or absent. It passes this to the condition evaluator for every `askIf`.

**Which candidates are active.** A candidate with no `askIf` is active. Otherwise the finder asks the condition evaluator, `Holds(askIf, scope, answers, switches)`, with the candidate's scope as the scope and the answers described next. The candidate is active when the condition holds. An active question keeps its stored answer and is returned. A candidate whose condition does not hold is not returned and does not block. It keeps its stored answer in the plan.

The answers passed to the evaluator hold only the stored answers of the questions found active before this one, whoever answers them. The answer of any other question counts as unanswered: a question whose condition fails, a question of a section that is out, an answer stored for a name that no candidate declares, the question itself, and a question not decided yet. This is the rule in [conditions](../../../docs/reference/section-template-schema/conditions.md): a settled answer is one for a question that is active. A declared default is not in the answers, so it never counts.

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
- `Description`, `Options`, `Required`, and `Multiple` are the variable's own. A variable with `options` and a type other than `choice` keeps its options.
- `Type` is the variable's type, or `string` when it has none. `Audience` is the variable's audience, or `onboarder` when it has none, so it is never empty.
- `Answer` is the answer stored for the question, exactly as stored: a settled value, a Deferred value, empty text, an empty list, or a value of no kind when nothing is stored. It is never filled from the default.
- `DeclaredDefault` is the variable's `default`, exactly as declared, and a value of no kind when it has none. It is not an answer. A question with a default and no answer has no answer and blocks when it is required, as [0045](../../../docs/decisions/0045-defaults-are-suggestions.md) says.

**Groups.** `Group` is a text key. Two returned questions have the same key exactly when they have the same name, the same type, the same options, and the same `multiple`. Options are compared as a set, so their order and repeats do not matter. The description, the pattern, `required`, the default, and the scope play no part, so a plan variable and a section variable with the same name and type share a group. The key is not empty. Its text is not specified, and only equal or different matters. The finder only reports the group. It never merges anything.

**Suggestions.** A question's suggestions are a list built in two parts, defaults first and then given answers. They do not depend on the audience asked for.

1. **Defaults.** For each returned question in the question's group, in the order of the questions, its declared default. The group is made of the questions of the full list, whatever their audience, so the default of a variable in a section that is out, or of a question whose condition fails, is not offered.
2. **Given answers.** The answers stored for the question's name, taken from the plan's answers first and then from each section's answers in the order of the section ids, compared as bytes. Every section's answers count, whether the section is active or not, and whether or not a question for that answer is returned. This includes the answer of the question itself.

A candidate is offered only when it fits the question:

- It is settled. A deferred answer, empty text, and a list with no non-empty item are never offered. Deferred has its own entry in an interface, and it is not a suggestion.
- A `boolean` offers a Flag. A `command`, or any question with `multiple` true, offers a List. Every other type offers Text.
- A `choice` offers only Text that is one of its options.
- A `string` with a `pattern` offers only Text the pattern matches. The pattern uses RE2 syntax and matches anywhere in the text, so a pattern that must match the whole text starts with `^` and ends with `$`.

Two values are equal when they have the same kind and the same content. A list equals a list with the same items in the same order. Among the defaults, a value that equals an earlier one is dropped, and the same holds among the given answers. A value that is both a default and a given answer is listed twice, once with source `default` and once with source `given`. The suggestions of two questions in one group can differ, because the type is the same and the pattern and the options need not be.

**Blockers.** A blocker is an active question that is required and has no answer. The blockers for a named audience are those of its questions, so a required onboardee question does not block the onboarder, and the other way round. The blockers for an empty audience are those of both, in the order of the full list. A required question is satisfied by a settled or a deferred answer, as [0035](../../../docs/decisions/0035-settled-and-deferred-answers.md) says. These do not satisfy it: nothing stored, empty text, an empty list, and a list of only empty text. A false Flag is settled. A default is not an answer. The blockers are listed in the order of the questions, and each holds the question's id. A question that is not active never blocks, whether its section is out, its switch is off, or its condition fails. An answer already stored for a question of either audience counts, whoever put it there.

## Properties

These hold for every plan the contract tests use.

- **Pure.** The plan is the same after the calls as before.
- **No memory.** The same plan gives the same result on a shared finder, whatever calls came before, forward or in reverse order, and however often it is repeated.
- **Same plan, same result.** For every audience, rotating or reversing `plan.Sections` does not change the questions, their suggestions, or the blockers.
- **Blockers agree with the questions.** For every audience, the blockers are the returned questions that are required and neither settled nor deferred, in the same order.
- **A named audience is a filter.** The questions of a named audience are the questions of the empty audience that have that audience, unchanged and in the same order. Their blockers are the blockers of the empty audience for those questions.
- **New lists.** Changing, reordering, or replacing an item of a result does not change the plan or a later result. The lists inside a Question, such as its options and its suggestions, may share their data with the plan.
- **Empty plan.** A plan with nothing in it gives no question and no blocker, for every audience.

## Side effects and plan access

None. It reads the plan's variables, the answers, and the content of the sections it needs: `id`, `includeIf`, `timing`, `after`, `Enabled`, and `variables`. It returns no plan. This matches the row for `question/finder` in [components](../../../docs/architecture/components.md).

## Not covered

- Duplicate section ids, and duplicate variable names in one scope.
- A `timing` that is not one of the five names, a `type` that is not one of the five types, and an `audience` on a variable, or an audience asked for, that is none of the values above.
- A `pattern` that does not compile. The validators report it.
- An answer stored under a section id the plan does not hold, or under a name that no variable of the plan declares. Whether such an answer is offered as a given answer is left open.
- Whether the plan variables, as a whole, are valid. A plan variable with an `askIf` is evaluated like any other, and the plan validator reports it. A plan variable's `askIf` does not change which sections the selector finds active, because the selector reads the stored plan answers.
- A condition that names a question that does not exist, a section clause that names no section, or a clause with both kinds of key. The validators and the decoder report them, and the evaluator reads them as written.
- A default that does not fit its variable's type or pattern. The validators report it, and the finder only decides whether to offer it.
- Whitespace only text as an answer or a default.
- Whether a choice or a boolean answer is valid. That is the plan validator's check.
- The interface's own entries, such as `deferred`. A suggestion is never `deferred`.

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
| A result list shared between calls | The result-list test |
