# Section resolver contract

Status: draft by an agent. Not reviewed by a person. Component: `section/resolver`. Issue: [#63](https://github.com/8thlight-daveloper/onboard-kit/issues/63).

The section resolver fills the variables in a section's text from the answers and the directory. It takes a Section, the plan's answers, the plan's directory, and the id of the plan's team, and returns a new Section with its text filled, and a list of the variables it could not fill. It is a pure function. It reads no file, reads no clock, keeps no state between calls, and does not change its arguments. See [variables](../../../docs/reference/section-template-schema/variables.md), [steps](../../../docs/reference/section-template-schema/steps.md), [0030](../../../docs/decisions/0030-self-contained-plan-file.md), [0035](../../../docs/decisions/0035-settled-and-deferred-answers.md), [0080](../../../docs/decisions/0080-directory.md), and the [domain model](../../../docs/reference/domain-model.md).

The document assembler calls the resolver once for each active section, after the section selector. Nothing resolved is stored. The plan keeps `$name` in its sections.

The name boundary follows [0040](../../../docs/decisions/0040-blank-text-and-variable-name-length.md). Defaults are suggestions under [0045](../../../docs/decisions/0045-defaults-are-suggestions.md) and [0120](../../../docs/decisions/0120-defaults-require-acceptance.md), which supersedes the directory-default fallback in 0080. Gap audience follows [0055](../../../docs/decisions/0055-variable-audience.md). This increment preserves the existing marker text and warning severity; context-dependent caller severity remains separate work.

## Interface

`Resolve(section, answers, directory, team) -> (section, problems)`.

- `section` is a Section, as it is held in a plan.
- `answers` is the plan's Answers value. The resolver reads only `answers.sections[section.id]`. It never reads `answers.plan` or another section's answers.
- `directory` is the plan's Directory, with its contacts, roles, and teams. It may be empty.
- `team` is the id of the plan's team, or the empty string when the plan has none. The caller finds it, as the answer to the plan variable of type `team`. The resolver does not look for it in the answers. The directory and `team` are read only for the variables of type `contact`, `role`, and `team`.
- The first result is a Section with the same fields as the input. Only the text fields listed under What is filled can differ.
- The second result is the variables the text used that had no value to fill, one Problem each. It is empty when there are none. A caller can build a gap table from it.

The resolver does not fail. Any Section and any Answers give a result.

## What is filled

A `$name` is filled in these fields, and only in these.

| Level | Fields |
| -- | -- |
| Section | `title`, `summary`, `owner`, `verification`, `body` |
| Step | `name`, `notes`, `owner`, `open`, `directory`, each item of `run`, each value of `env` |
| Check | each item of `run`, `ask` |

These are left as held, even when they hold `$`.

- Names and identifiers: `id`, `parent`, `timing`, `after`, `requires`, and a step's `id` and `timing`. A `parent` is the id of another section of the plan, as [0070](../../../docs/decisions/0070-sub-sections.md) says. The resolver keeps it as held, even when it holds `$`, and does not look for the section it names.
- `escalation` on a section and on a step. Its items name variables and are not text.
- The keys of a step's `env`, and a check's `env`, which is a variable name.
- `includeIf`, and every part of every variable the section declares: its description, example, default, options, pattern, hint, and `askIf`. A declaration is not resolved.
- `enabled`, and the section clauses in a condition. They are not text.
- `source` and `unknownFields`.

The returned Section keeps every declared variable, so it still describes its own questions. Each variable comes back whole, with its `example` as held. An `example` is an illustration for the person answering. It is not an answer and not a default, so it never fills a `$name` and is never read for one.

## Reading the text

A string is read from left to right, once.

| Text | Becomes |
| -- | -- |
| `$$` | `$` |
| `$` and a name that the section declares | The variable's text, see What a variable becomes |
| `$` and a name that the section does not declare | Left as written, with no problem |
| Any other `$`, such as `$5`, `$Repo_url`, `${name}`, `$ `, and a `$` at the end | Left as written, with no problem |

A name is a lowercase letter followed by at least two lowercase letters, digits, or underscores: `[a-z][a-z0-9_]{2,}`. A one or two character run such as `$x` or `$ab` stays literal, with no problem, in prose and in `run`. A name ends at the first other character, and it takes every character it can, so `$repo_url.git` is the variable `repo_url` and then `.git`, and `$repo_urlx` is the name `repo_urlx`. A name can hold `0` and `9`, so `$item0` and `$item9` are names, and a declared `item` does not fill them. An upper case letter ends a name, so `$itemA` is the variable `item` and then `A`. So do `/`, `:`, `` ` ``, `{`, `@`, and `[`. A name that ends at the end of the text is a name. A declared name is one in the `name` of a variable in `section.variables`. A name that only the answers hold, or only the plan holds, is not declared. A variable named `enabled` is an ordinary variable, because the switch of a section is the field `enabled` of the section and not a variable.

`$$` is read first at each place, so `$$name` is `$name` as text and is not filled, `$$$name` is `$` and then the variable, and `$$$` is `$` and then a lone `$`, which reads as `$$`.

Filling is one pass. What a variable becomes is written as it is and is not read again. A stored answer that holds `$other` or `$$` comes out as held.

A `$` or `$$` at the very end of a field is read by the same rules, so `cost $` stays `cost $` and `price $$` becomes `price $`. A field that is only `$` or `$$` is read the same way.

The resolver does not report a `$` that is not a variable. The section validator does.

## What a variable becomes

For a declared variable the resolver takes the answer stored for it in the section's own scope, then follows these steps.

1. A settled answer is used. For a variable of type `contact`, `role`, or `team`, an answer is used only when it names something in the directory, see Directory variables.
2. A deferred answer is a deferred value. A default does not replace it.
3. An unanswered answer gives no value. A default is never read for resolution, regardless of its kind, the variable type, or the audience. A deferred default does not make an unanswered variable deferred.
4. Explicitly accepting a suggestion stores an answer. The resolver uses that stored answer, even if it equals the default; later changes to the default cannot change the resolved content.
5. No value in a variable with `audience: onboardee` becomes `<name>`, with the variable's name, and reports an unanswered gap.
6. No value in any other variable becomes `_unanswered_` and reports an unanswered gap. A variable with `audience: onboarder` is such a variable, as is one with no `audience`.

Settled, deferred, and unanswered are as [0035](../../../docs/decisions/0035-settled-and-deferred-answers.md) defines them. Text is settled when it has at least one character. A Flag is always settled. A List is settled when it has at least one item that is not empty. The Text `deferred` is a plain string, and it is settled. A value of no kind is unanswered. Nothing is checked against the variable's type, options, or pattern.

An unanswered answer is nothing stored, empty text, an empty list, or a list of only empty items. For a directory variable it is also an answer that names nothing in the directory. The resolver does not evaluate `askIf`. A variable whose question is not asked is read like any other, so one that the text uses and that has no answer is a gap, whether or not it has a default.

A variable that has a value is written as follows.

| Value | Written as |
| -- | -- |
| Text | The text, as held, with its spaces and line breaks |
| Flag | `true` or `false`, in lowercase |
| List | The items in order, joined with one space |

A list is written as held, so an empty item takes its place in the join. The Deferred, unanswered, and onboarding results are these three markers, written exactly.

| Result | Written as | Problem |
| -- | -- | -- |
| Deferred | `_deferred_` | `gap.deferred` |
| Unanswered | `_unanswered_` | `gap.unanswered` |
| Onboardee, no value | `<name>` | `gap.unanswered` |

A variable with no `type` is a `string`.

## Directory variables

A variable of type `contact`, `role`, or `team` holds ids. A `contact` or `role` answer is one id as Text, or a List of ids when the variable is `multiple`. A `team` answer is one id as Text. Ids are matched exactly, so `Sam-Okafor` and a contact's name are not the id `sam-okafor`. Contact, role, and team ids are separate, so a role id given to a `contact` variable names nothing. The type decides the lookup. A `string` variable that holds a contact id or a contact name is written as held.

**Ids.** An id that is not in the directory counts as unanswered, as if it were not stored. In a List the unknown ids are skipped and the known ones keep their order. A List with no known id is unanswered. An unanswered directory variable follows the steps above: it reports an unanswered gap, written `<name>` for an `onboardee` variable and `_unanswered_` otherwise. A known default id never replaces an unknown answer id or a missing answer. Role-holder lookup still uses the selected role's team override or organization-wide holders; that is directory lookup, not default acceptance.

**The team.** The team of the plan is the Team in `directory.teams` whose id is `team`. When `team` is empty or names no Team, the plan has no team. The `team` variable of a section is not the plan's team. It is looked up by its own answer.

**Resolving a role.** A role has holders, which are contact ids in order. When the plan has a team and the team's `holders` has an entry for the role, the team's entry is the role's holders, even when the entry is an empty list. Otherwise the role's own `holders` are used. A holder that is not a contact in the directory is skipped, and the order of the rest is kept. A role with no holders left is vacant. A team's entry does not add to the role's holders and does not fall back to them, so an entry that is empty or holds only unknown ids makes the role vacant for that team.

**Rendering a directory value.**

| Value | Written as |
| -- | -- |
| A contact that has a `slack` | `Name, slack` |
| A contact with no `slack` and an `email` | `Name, email` |
| A contact with neither | `Name` |
| A role with holders | Its holders as contacts, in order, joined with `; ` |
| A vacant role | `Role name (vacant)`, with the `name` of the directory role |
| A team | The `name` of the Team |
| A List of ids | The id's texts in order, joined with `; ` |

The `slackId` is never written. A vacant role is text and not a gap. It adds no Problem. The assembler reports it, see [0080](../../../docs/decisions/0080-directory.md). A vacant role keeps its place in a List of roles. Nothing written from the directory is read again, so a `$` in a contact's name or handle comes out as held. Deferred and `required` work as for other types.

## A command in `run`

A `run` item that is exactly `$name`, for a declared variable of type `command` that has a list as its value, is replaced by the items of the list, in order, each as its own item of `run`. This holds in a step's `run` and in a check's `run`. The items are written as held and are not read again.

In every other place a command is prose. A `run` item such as `--cmd=$test_command` or `$build && $test_command` takes the items joined with one space. An exact `$name` for a variable that is not a `command`, such as a `string`, a `choice`, or a `boolean`, gives one item and never splits on spaces, even when its value is a List, which is then joined with one space. An item is exactly `$name` only when it is a `$` and a name and nothing else, so `#name`, `!name`, and ` name` are prose and come back as written. An exact `$name` for a `command` whose value is Text, not a list, gives one item, the text as held, and never splits on spaces. An exact `$name` whose value is deferred, unanswered, or an onboarding blank gives one item, the marker. An exact `$name` for a `contact`, `role`, or `team` variable gives one item, the text under Directory variables, with its `; ` and its spaces and never split. That holds for a List, a vacant role, and a team.

## Problems

A problem is reported for a declared variable when its marker is written, `_deferred_`, `_unanswered_`, or `<name>`, at least once in a filled field. A variable that the text does not use is never reported, and neither is one that is used only in a field that is left alone. A variable is reported once, however often it is used. Problems come in the order of the section's `variables`. A variable declared twice is not covered.

| Code | When |
| -- | -- |
| `gap.deferred` | A variable the text uses has a deferred value |
| `gap.unanswered` | A variable the text uses has no value, of either audience, including a directory variable whose answer names nothing in the directory |

Each Problem has the severity `warning`, `question` set to `<section id>.<name>`, and `section` set to the section's id. Its `detail.audience` is the declared audience, or `onboarder` when the audience is absent or empty. This warning severity is a compatibility default, not a context-dependent caller policy. An id that names nothing in the directory gives `gap.unanswered`, as an unanswered variable does. A vacant role gives no Problem. The resolver does not decide whether a gap blocks the caller's workflow. Required-question validation and context-dependent caller severity are outside this increment. The `message` is for people and its wording is not specified. Other `detail` entries are not specified.

## Properties

These hold for every input the contract tests use.

- **Pure.** The section, the answers, and the directory are the same after the call as before, including every list and map inside them. A holders list is not sorted or changed in place.
- **Nothing shared.** The returned Section shares no list, map, or pointer with the input. Changing it changes nothing in the input.
- **A copy of the switch and the clauses.** The returned Section holds the same `enabled` as the input, absent when absent, `false` when `false`, and `true` when `true`, and it is a new pointer. The resolver never turns a section on or off, and never reads the switch. A condition in `includeIf` or in a variable's `askIf` comes back as held: a question clause and a section clause, with the section id and the `enabled` of the section clause kept, `false` as well as `true`. Neither shares anything with the input. See [0050](../../../docs/decisions/0050-section-switch.md).
- **No memory.** The same call gives the same result whatever calls came before, forward or in reverse order.
- **Text only.** Every field that is not listed under What is filled is equal to the input, `parent` included. A section with a `parent` comes back with that `parent`, and one with none comes back with none.
- **Nothing to fill.** A section that holds no `$name` for a declared variable comes back equal to the input, with no problem. A list with no items is not told apart from an absent one. A step's `env` is, so an absent `env` stays absent and an empty one stays empty.
- **Zero values.** The zero Section with the zero Answers gives the zero Section and no problem.

## Side effects and plan access

None, and it reads section content, the section's own answers, the directory, and the team id only. It writes nothing but the returned copy. This matches the row for `section/resolver` in [components](../../../docs/architecture/components.md).

## Not covered

- Whether the directory is valid: a duplicate id, a holder that is not a contact, a team entry for a role that does not exist. Two entries with one id are not covered. Those are for the validator.

- That the plan has at most one `team` variable, and how the caller finds the team.

- An answer of the wrong shape for a directory variable, such as a List for a `team`, or a Flag. A contact with an empty name, and a role with an empty name.

- The same contact written twice in a List, or held by two roles in one List.

- Reporting a vacant role. The assembler does.

- The format of the document, such as Markdown escaping of `<name>` and `_deferred_`, one person to a line, and the gap table. Those are the assembler's.

- Whether the answers fit the types, options, or patterns, whether a required question is answered, and whether a `$` is valid. Those are for the validators.

- Two variables with the same name in one section.

- A Flag for a `string`, a Flag for a `command`, or another mismatch of type and value, except Text for a `command` and the Lists that A command in `run` covers. The value is written by its kind, as the table says.

- Text of only whitespace.

- A name that does not follow the schema's name rule, in a declaration.

- Which sections are active. The caller passes the section it wants filled.

- That a `parent` names a section of the plan, that no chain of parents loops, and that an `includeIf` names only ancestors. The resolver is given one section and never sees the others, and it reads only the section's own answers, so an ancestor's answer is not read here. The plan validator's structure check covers the parent.

## Decisions to know

These are choices the docs did not settle. Each says what the resolver does, what raised it, and how far a change would reach. `hard` means a change touches what template authors write, what a plan holds, or what a person sees. `cheap` means a change touches one fixture and nothing outside the component. The directory choices come first.

D1. **An id that names nothing counts as unanswered without fallback.** It reports `gap.unanswered` for either audience, with `_unanswered_` or `<name>` according to audience. Defaults never supply replacement ids. Source: 0120 supersedes the fallback in 0080. `hard`.
D2. **A vacant role is text and not a gap.** It is written `Role name (vacant)` and adds no Problem, because the assembler reports it as `directory.role-vacant`. A vacant role keeps its place in a List of roles. Raised by: "No holder is a gap, shown like any other gap and never as an empty name or a guess". `hard`.
D3. **A team entry replaces the role's holders and never merges.** An entry that is empty, or holds only unknown ids, makes the role vacant for that team. Raised by: "the newcomer's team, if the team override names holders for it, otherwise the role's own holders". `hard`.
D4. **A List keeps its known ids in order and skips unknown ones.** A List with no known id is unanswered. The same contact twice is written twice. Raised by: "A `multiple: true` contact or role question holds an ordered list". `hard`.
D5. **Every list of directory values joins with `; `, a role's holders and a List of ids alike.** A role in a List of roles contributes its holders, so the semicolons run on across roles. Raised by: "The exact forms are not decided". `hard`.
D6. **A directory value in `run` is one item.** An exact `$name` for a `contact`, `role`, or `team` variable is never split and never spliced, even for a List. Raised by: "A variable of type `command` is written `[$name]` and expands to its arguments", which names no other type. `hard`.
D7. **The type decides the lookup.** A `string` variable that holds a contact id or a contact name is written as held. Raised by: "An answer stores an id and nothing else" for the three directory types, which names no lookup for any other type. `hard`.
D8. **The team parameter and a team variable are separate.** Roles resolve with the `team` parameter. A `team` variable is written from its own answer. An unknown or empty `team` is no team. Raised by: "The team of a plan is the answer to the plan variable of type `team`". `cheap`.
D9. **Ids are matched exactly and by kind.** Case matters, a contact's name is not an id, and a role id given to a `contact` variable names nothing. Raised by: "Ids are lowercase slugs". `cheap`.
D10. **Escalation stays as held.** Its items name variables and are not text, so a `contact` or `role` name in `escalation` is not filled here. Raised by: "`escalation` and `owner` references may name `contact` or `role` variables". `cheap`.

The choices below were made before the directory.

01. **The markers wording.** A deferred answer is written `_deferred_`, an unanswered onboarder variable is written `_unanswered_`, and an onboarding variable with no value is written `<name>`. Raised by: "It renders as a visible gap with an owner and timing" and "The assembled document shows it as a blank, `<name>`." The ACME document shows `_deferred_` and `<account_name>`, and the docs do not give the wording for the unanswered one. The assembler does not find a marker in text. It reads the problems. `hard`.
02. **Unanswered is a visible marker and a gap, even with a default.** The variable is not left as `$name`, and it is not blank. Raised by: "Every variable is resolved or reported as a gap." `hard`.
03. **Only stored answers resolve content.** A missing or deferred answer never authorizes using a suggestion. Source: 0045 and 0120. `hard`.
04. **A list is joined with one space.** A List of a `command` is written in one string, in a sentence, a field, or a `run` item that is not exactly `$name`. Raised by: "joins with spaces in prose" for `command` in variables.md. `cheap`.
05. **An onboardee variable is a blank only when it has no value.** Stored settled and deferred answers follow the general rules. The unanswered case becomes `<name>` and reports `gap.unanswered` with audience `onboardee`. Source: 0055; marker wording is preserved for compatibility. `hard`.
06. **Which fields are filled.** Section and step text, `run`, and `env` values are filled. Ids, `timing`, `after`, `escalation`, `includeIf`, variable declarations, `source`, and unknown fields are not. A step's `name`, `open`, and `directory` are filled even though the docs name only `owner`, `run`, `env`, `notes`, and `body` as places for a variable. Raised by: "A variable ... is written `$name` where text uses it" and "`owner` ... May contain `$name`". `hard`.
07. **An answer is not read again.** An answer that holds `$` comes out as typed. Raised by: "Variables fill at assembly." `hard`.
08. **Only an exact `$name` item splices.** A command inside a longer `run` item is joined with spaces, and only a `command` variable splices. Raised by: "A variable of type `command` is written `[$name]` and expands to its arguments". `cheap`.
09. **A name the section does not declare is left as written, with no problem.** The validators report it. Raised by: "A `$name` the section does not declare is a lint error, even when a plan variable has that name." `cheap`.
10. **Problems are the model's Problem.** The gaps come back as `gap.deferred` and `gap.unanswered` warnings, only for variables the text uses, once each, in the order of the section's variables. No new type is added to the model. Raised by: "perhaps a list of variables it could not fill, so a caller can build a gap table". `cheap`.
11. **The resolver reads only the section's own answers and does not evaluate `askIf`.** Raised by: "Filling reads one scope". `cheap`.
12. **A default of any kind is only a suggestion.** Empty, settled, and deferred defaults do not change answer state. A stored answer equal to a suggestion remains independent of later default changes. Source: 0045 and 0120. `hard`.
13. **A `parent` is an identifier, so it is not filled.** A `$` in it stays. Raised by 0070, which names the resolver as having nothing to do for sub-sections, and by the rule that identifiers are left as held. Case: `sub-section`. `cheap`.

## Wrong implementations tried

Each row names a plausible wrong implementation and the groups of cases that fail it. How the contract was checked, and with what scores, is recorded in the issue, not here.

| Wrong implementation | Caught by |
| -- | -- |
| Drop, reorder, resolve placeholders in Requires, or share its backing list | [Requires preservation and isolation](requires_test.go) |
| A stored answer changes when a suggestion changes, a deferred suggestion becomes an answer, or a suggestion replaces an unknown directory id | `TestOnlyStoredAnswersResolveRegardlessOfSuggestions`, from 0035, 0045, 0055, and 0120, across every variable type and both audiences |
| A gap omits its audience, or an omitted audience is not normalized to onboarder | The fixture problem assertions and `TestOnlyStoredAnswersResolveRegardlessOfSuggestions`, from 0055 |
| A three-character name is skipped, a short run is filled from undeclared answers, or a longer name is truncated to a declared prefix | `TestShortRunsAreLiteralAndThreeCharacterNamesResolve`, from 0040 and Reading the text, in prose and command arguments |
| A field not filled: title, summary, owner, verification, body, a step's name, notes, owner, open, directory, run, an env value, a check's run or ask | `every-field`, and `list-in-prose`, `command-in-run`, `deferred` |
| A field filled that is left alone: escalation, a step's escalation, a check's env, an env key, a variable's description or default, an unknown field, `includeIf`, `source` dropped | `left-alone`, `gap-report`, `single-pass` |
| Reads the plan's answers as a fallback, takes any or the first section's scope | `own-scope`, `unanswered` |
| A name that is not declared filled from the answers | `undeclared` |
| A name that is not declared dropped, or a lone `$` dropped | `undeclared`, `dollar-signs` |
| `$$` kept as two signs, or `$$name` filled | `dollar-signs` |
| A name that ends at an underscore or a digit, the shortest match, upper case folded, or `${name}` read | `dollar-signs`, `every-field` |
| A second pass over the filled text, or a default filled in place | `single-pass`, `dollar-signs` |
| A flag written in capitals or as empty text, text trimmed | `value-kinds`, `defaults` |
| A list joined with a comma, `and`, or `then` | `list-in-prose`, `command-in-run` |
| A command not spliced, spliced from a string variable, spliced inside a longer item, spliced for a `string`, first item dropped, or not spliced in a check | `command-in-run`, `single-pass`, `defaults` |
| A default silently used for a missing, invalid-directory, or deferred answer | `defaults`, `deferred`, `onboarding`, `directory-defaults`, `directory-unknown-ids` |
| Empty text, an empty list, or a list of empty items treated as settled, a Flag `false` treated as unanswered, or an unaccepted `false` default used | `defaults`, `value-kinds` |
| An empty or deferred default used | `defaults`, `deferred` |
| Another wording for a marker, the same marker for deferred and unanswered, or `$name` left for an unanswered one | `deferred`, `gap-report`, `onboarding`, `defaults` |
| The Text `deferred` treated as a deferred answer | `deferred` |
| A gap not reported, reported as an error, with no section, with a bare question, or with the wrong code | `deferred`, `defaults`, `gap-report`, `onboarding` |
| A variable the text does not use reported | `gap-report`, `left-alone`, `nothing-to-fill` |
| Problems in order of use, sorted by name, or one for each use | `gap-report`, `deferred` |
| An onboardee gap omitted, its audience missing, its stored answer ignored, or a deferred value written as an unanswered blank | `onboarding`, the gap-audience assertions |
| An `env` map made when there was none | `command-in-run`, `defaults`, `gap-report` |
| The answers changed, or an input list or map changed | The inputs-unchanged test |
| The result sharing a list, map, or pointer with the input | The shares-nothing test |
| A cache by section id | `command-in-run`, `defaults`, `deferred` |
| An `example` dropped when a variable is copied, written from the answer or the default, or read for `$name` | `variable-examples`, the example test, and the shares-nothing test |
| The variables dropped from the result | `command-in-run`, `defaults`, `deferred` |
| After the first build: a read one character past a `$` that panics at the end of a field, for `$`, `$$`, and `$$$` | `dollar-at-end` |
| After the first build: a name that stops before `0`, `9`, `a`, or `z`, takes a character just outside those ranges or an upper case letter, starts at a fixed place, or fails at the end of the text | `name-edges` |
| After the first build: an exact `$name` test that stops at the second character when it is a digit or an underscore, so a command such as `$e2e_test` is not spliced | `command-values` |
| After the first build: a command held as Text split on spaces, or a `string` with spaces split or spliced, in a step or in a check | `command-values`, `command-in-run` |
| After the first build: a condition's `isSettled` pointer shared with the input or dropped, in `includeIf` or `askIf` | `settled-conditions`, and the tests that change the result's `isSettled` targets and check the input |
| A section's switch dropped, turned on, or shared by pointer with the input | `switch-clauses`, and the shares-nothing tests |
| A section clause dropped from `includeIf` or from `askIf`, or its section id lost | `switch-clauses`, and the switch-and-clauses test |
| A section clause read as a question clause, or its `enabled` of `false` lost or flipped | `switch-clauses`, and the switch-and-clauses test |
| A section clause list shared with the input, so that changing the result changes the input | the switch-and-clauses test |
| A `parent` dropped, emptied, filled like text, or taken from an answer | `sub-section`, `nothing-to-fill`, the parent test |
| A variable named `enabled` left as written, written as a marker, or taken for the switch | `switch-clauses`, `value-kinds` |
| A contact written with its slack id, with the email before the slack, or with both | `directory-contact-forms`, `directory-roles` |
| A list of directory values joined with a comma or a space | `directory-roles`, `directory-run`, `directory-fields` |
| A team entry ignored, an empty entry falling back to the role's holders, or an entry of unknown ids falling back | `directory-roles`, `directory-team-override` |
| The team taken from the first team, from the name of the team, or no team giving no holders | `directory-roles`, `directory-no-team`, `directory-unknown-team`, `directory-team-override` |
| An unknown team giving no holders, or a team with no entry giving no holders | `directory-unknown-team`, `directory-plain-team` |
| An unknown holder written as its id, a vacant role written empty or without its name, or a vacant role dropped from a List | `directory-roles`, `directory-team-override`, `directory-run` |
| A vacant role reported as a gap | `directory-roles`, `directory-team-override` |
| Holders sorted, or sorted in place in the directory | `directory-roles`, `directory-fields`, the inputs-unchanged test |
| An unknown id written as the id, an unknown id in a List voiding the list, or only the first id used | `directory-unknown-ids`, `directory-run`, `directory-roles` |
| An unknown directory id replaced by a default, or a directory suggestion silently accepted | `directory-unknown-ids`, `directory-defaults` |
| A deferred answer replaced by the default | `directory-deferred` |
| A contact id matched by name or without case, a role id read as a contact, a contact id read as a role | `directory-unknown-ids` |
| A team variable written from the plan team or as its id | `directory-team-name` |
| A directory List spliced into several `run` items, or one item split on spaces | `directory-run`, `directory-fields` |
| A `string` variable looked up in the directory | `directory-string-unchanged`, `directory-empty` |
| A contact's name or handle read again, or its `$$` halved | `directory-contact-forms` |
| A command spliced for a `string`, `choice`, or `boolean` variable that holds a List | `types-and-audience`, `list-in-run` |
| An item taken for an exact name when a character before the name sorts below `$` | `types-and-audience` |
| A variable with `audience: onboarder` written as an onboarding blank | `types-and-audience` |
| A condition's `is`, `isNot`, or `isNotOneOf` value sharing its list with the input | `condition-lists`, and the shares-nothing tests |

## Bootstrap preservation

[0250](../../../docs/decisions/0250-bootstrap-content-and-projection.md) supplies
the source for the independently specified [preservation tests](bootstrap_test.go).
Resolving text preserves the section's held Bootstrap value, whether true or false, while carrying out the real answer substitution. Never infer the marker from answers or mutate it during resolution.

## Requires preservation

[0260](../../../docs/decisions/0260-prerequisites-and-action-order.md) supplies
the source for [Requires tests](requires_test.go). Requires holds identifiers,
not prose: keep every exact item, order and repeat, including dollar signs.
Do not resolve it from answers or infer it from After. Preserve Bootstrap while
performing actual prose substitution. The existing nothing-shared promise
includes Requires in both directions.
