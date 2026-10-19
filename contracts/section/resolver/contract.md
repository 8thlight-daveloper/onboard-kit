# Section resolver contract

Status: draft by an agent. Not reviewed by a person. Component: `section/resolver`. Issue: [#63](https://github.com/8thlight-daveloper/onboard-kit/issues/63).

The section resolver fills the variables in a section's text from the answers. It takes a Section and the plan's answers, and returns a new Section with its text filled, and a list of the variables it could not fill. It is a pure function. It reads no file, reads no clock, keeps no state between calls, and does not change its arguments. See [variables](../../../docs/reference/section-template-schema/variables.md), [steps](../../../docs/reference/section-template-schema/steps.md), [0030](../../../docs/decisions/0030-self-contained-plan-file.md), [0035](../../../docs/decisions/0035-settled-and-deferred-answers.md), and the [domain model](../../../docs/reference/domain-model.md).

The document assembler calls the resolver once for each active section, after the section selector. Nothing resolved is stored. The plan keeps `$name` in its sections.

## Interface

`Resolve(section, answers) -> (section, problems)`.

- `section` is a Section, as it is held in a plan.
- `answers` is the plan's Answers value. The resolver reads only `answers.sections[section.id]`. It never reads `answers.plan` or another section's answers.
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

- Names and identifiers: `id`, `timing`, `after`, and a step's `id` and `timing`.
- `escalation` on a section and on a step. Its items name variables and are not text.
- The keys of a step's `env`, and a check's `env`, which is a variable name.
- `includeIf`, and every part of every variable the section declares: its description, default, options, pattern, hint, and `askIf`. A declaration is not resolved.
- `source` and `unknownFields`.

The returned Section keeps every declared variable, so it still describes its own questions.

## Reading the text

A string is read from left to right, once.

| Text | Becomes |
| -- | -- |
| `$$` | `$` |
| `$` and a name that the section declares | The variable's text, see What a variable becomes |
| `$` and a name that the section does not declare | Left as written, with no problem |
| Any other `$`, such as `$5`, `$Repo_url`, `${name}`, `$ `, and a `$` at the end | Left as written, with no problem |

A name is a lowercase letter followed by any number of lowercase letters, digits, and underscores, as the schema defines it. A name ends at the first other character, and it takes every character it can, so `$repo_url.git` is the variable `repo_url` and then `.git`, and `$repo_urlx` is the name `repo_urlx`. A name can hold `0` and `9`, so `$item0` and `$item9` are names, and a declared `item` does not fill them. An upper case letter ends a name, so `$itemA` is the variable `item` and then `A`. So do `/`, `:`, `` ` ``, `{`, `@`, and `[`. A name that ends at the end of the text is a name. A declared name is one in the `name` of a variable in `section.variables`. A name that only the answers hold, or only the plan holds, is not declared.

`$$` is read first at each place, so `$$name` is `$name` as text and is not filled, `$$$name` is `$` and then the variable, and `$$$` is `$` and then a lone `$`, which reads as `$$`.

Filling is one pass. What a variable becomes is written as it is and is not read again. An answer or a default that holds `$other` or `$$` comes out as held.

A `$` or `$$` at the very end of a field is read by the same rules, so `cost $` stays `cost $` and `price $$` becomes `price $`. A field that is only `$` or `$$` is read the same way.

The resolver does not report a `$` that is not a variable. The section validator does.

## What a variable becomes

For a declared variable the resolver takes the answer stored for it in the section's own scope, then follows these steps.

1. A settled answer is used.
2. A deferred answer is a deferred value. A default does not replace it.
3. An unanswered answer takes the section's default for the variable, when the default is settled. A default that is deferred gives a deferred value. A default that is not settled, or none, gives no value.
4. No value in a variable with `phase: onboarding` becomes `<name>`, with the variable's name. It is not a gap.
5. No value in any other variable is an unanswered gap.

Settled, deferred, and unanswered are as [0035](../../../docs/decisions/0035-settled-and-deferred-answers.md) defines them. Text is settled when it has at least one character. A Flag is always settled. A List is settled when it has at least one item that is not empty. The Text `deferred` is a plain string, and it is settled. A value of no kind is unanswered. Nothing is checked against the variable's type, options, or pattern.

An unanswered answer is nothing stored, empty text, an empty list, or a list of only empty items. The resolver does not evaluate `askIf`. A variable whose question is not asked is read like any other, so one that the text uses and that has no answer and no default is a gap.

A variable that has a value is written as follows.

| Value | Written as |
| -- | -- |
| Text | The text, as held, with its spaces and line breaks |
| Flag | `true` or `false`, in lowercase |
| List, for a variable of type `person` | The items in order, joined with `, then ` |
| List, for any other variable | The items in order, joined with one space |

A list is written as held, so an empty item takes its place in the join. The Deferred, unanswered, and onboarding results are these three markers, written exactly.

| Result | Written as | Problem |
| -- | -- | -- |
| Deferred | `_deferred_` | `gap.deferred` |
| Unanswered | `_unanswered_` | `gap.unanswered` |
| Onboarding, no value | `<name>` | None |

A variable with no `type` is a `string`.

## A command in `run`

A `run` item that is exactly `$name`, for a declared variable of type `command` that has a list as its value, is replaced by the items of the list, in order, each as its own item of `run`. This holds in a step's `run` and in a check's `run`. The items are written as held and are not read again.

In every other place a command is prose. A `run` item such as `--cmd=$test_command` or `$a && $b` takes the items joined with one space. An exact `$name` for a variable that is not a `command`, such as a `string` or a `person`, gives one item and never splits on spaces. An exact `$name` for a `command` whose value is Text, not a list, gives one item, the text as held, and never splits on spaces. An exact `$name` whose value is deferred, unanswered, or an onboarding blank gives one item, the marker.

## Problems

A problem is reported for a declared variable when its marker is written, `_deferred_` or `_unanswered_`, at least once in a filled field. A variable that the text does not use is never reported, and neither is one that is used only in a field that is left alone. A variable is reported once, however often it is used. Problems come in the order of the section's `variables`. A variable declared twice is not covered.

| Code | When |
| -- | -- |
| `gap.deferred` | A variable the text uses has a deferred value |
| `gap.unanswered` | A variable the text uses has no value and is not an onboarding variable |

Each Problem has the severity `warning`, `question` set to `<section id>.<name>`, and `section` set to the section's id. A gap never blocks, because the plan validator blocks a required question that is unanswered. The `message` is for people and its wording is not specified. `detail` is not specified.

## Properties

These hold for every input the contract tests use.

- **Pure.** The section and the answers are the same after the call as before, including every list and map inside them.
- **Nothing shared.** The returned Section shares no list, map, or pointer with the input. Changing it changes nothing in the input.
- **No memory.** The same call gives the same result whatever calls came before, forward or in reverse order.
- **Text only.** Every field that is not listed under What is filled is equal to the input.
- **Nothing to fill.** A section that holds no `$name` for a declared variable comes back equal to the input, with no problem. A list with no items is not told apart from an absent one. A step's `env` is, so an absent `env` stays absent and an empty one stays empty.
- **Zero values.** The zero Section with the zero Answers gives the zero Section and no problem.

## Side effects and plan access

None, and it reads section content and the section's own answers only. It writes nothing but the returned copy. This matches the row for `section/resolver` in [components](../../../docs/architecture/components.md).

## Not covered

- The format of the document, such as Markdown escaping of `<name>` and `_deferred_`, one person to a line, and the gap table. Those are the assembler's.
- Whether the answers fit the types, options, or patterns, whether a required question is answered, and whether a `$` is valid. Those are for the validators.
- Two variables with the same name in one section.
- A Flag for a `string`, a Flag for a `command`, or another mismatch of type and value, except Text for a `command`, which A command in `run` covers. The value is written by its kind, as the table says.
- Text of only whitespace.
- A name that does not follow the schema's name rule, in a declaration.
- Which sections are active. The caller passes the section it wants filled.

## Decisions to know

These are choices the docs did not settle. Each says what the resolver does, what raised it, and how far a change would reach. `hard` means a change touches what template authors write, what a plan holds, or what a person sees. `cheap` means a change touches one fixture and nothing outside the component.

1. **The markers wording.** A deferred answer is written `_deferred_`, an unanswered variable with no default is written `_unanswered_`, and an onboarding variable with no value is written `<name>`. Raised by: "It renders as a visible gap with an owner and timing" and "The assembled document shows it as a blank, `<name>`." The ACME document shows `_deferred_` and `<account_name>`, and the docs do not give the wording for the unanswered one. The assembler does not find a marker in text. It reads the problems. `hard`.
2. **Unanswered and no default is a visible marker and a gap.** The variable is not left as `$name`, and it is not blank. Raised by: "Every variable is resolved or reported as a gap." `hard`.
3. **A deferred answer wins over a default.** The person chose to answer later, so the default does not fill it. Raised by: "A variable takes its own section's value, then that section's default" and "A deferred answer becomes a gap". `hard`.
4. **People are joined in a sentence only.** A list of `person` answers is joined with `, then ` and written in one string. One person to a line is not offered, because the resolver returns text and the assembler could not split it again. Raised by: "one per line in a list, or joined with a comma and the word then in a sentence". `hard`.
5. **An onboarding variable is a blank only when it has no value.** A settled answer, a default, and a deferred answer in the plan still follow the general rules, and only the unanswered case becomes `<name>` and is not a gap. Raised by: "The plan does not hold the value, because each person has their own." `hard`.
6. **Which fields are filled.** Section and step text, `run`, and `env` values are filled. Ids, `timing`, `after`, `escalation`, `includeIf`, variable declarations, `source`, and unknown fields are not. A step's `name`, `open`, and `directory` are filled even though the docs name only `owner`, `run`, `env`, `notes`, and `body` as places for a variable. Raised by: "A variable ... is written `$name` where text uses it" and "`owner` ... May contain `$name`". `hard`.
7. **An answer is not read again.** An answer that holds `$` comes out as typed. Raised by: "Variables fill at assembly." `hard`.
8. **Only an exact `$name` item splices.** A command inside a longer `run` item is joined with spaces, and only a `command` variable splices. Raised by: "A variable of type `command` is written `[$name]` and expands to its arguments". `cheap`.
9. **A name the section does not declare is left as written, with no problem.** The validators report it. Raised by: "A `$name` the section does not declare is a lint error, even when a plan variable has that name." `cheap`.
10. **Problems are the model's Problem.** The gaps come back as `gap.deferred` and `gap.unanswered` warnings, only for variables the text uses, once each, in the order of the section's variables. No new type is added to the model. Raised by: "perhaps a list of variables it could not fill, so a caller can build a gap table". `cheap`.
11. **The resolver reads only the section's own answers and does not evaluate `askIf`.** Raised by: "Filling reads one scope". `cheap`.
12. **A default that is not settled is no default.** An empty text default and an empty list default do not fill, and a deferred default gives a deferred value. Raised by: "`default` ... The value used when none is given." `cheap`.

## Wrong implementations tried

Each row names a plausible wrong implementation and the groups of cases that fail it. How the contract was checked, and with what scores, is recorded in the issue, not here.

| Wrong implementation | Caught by |
| -- | -- |
| A field not filled: title, summary, owner, verification, body, a step's name, notes, owner, open, directory, run, an env value, a check's run or ask | `every-field`, and `list-in-prose`, `command-in-run`, `deferred` |
| A field filled that is left alone: escalation, a step's escalation, a check's env, an env key, a variable's description or default, an unknown field, `includeIf`, `source` dropped | `left-alone`, `gap-report`, `single-pass` |
| Reads the plan's answers as a fallback, takes any or the first section's scope | `own-scope`, `unanswered` |
| A name that is not declared filled from the answers | `undeclared` |
| A name that is not declared dropped, or a lone `$` dropped | `undeclared`, `dollar-signs` |
| `$$` kept as two signs, or `$$name` filled | `dollar-signs` |
| A name that ends at an underscore or a digit, the shortest match, upper case folded, or `${name}` read | `dollar-signs`, `every-field` |
| A second pass over the filled text, or a default filled in place | `single-pass`, `dollar-signs` |
| A flag written in capitals or as empty text, text trimmed | `value-kinds`, `defaults` |
| A list joined with a comma, or a person list joined with a space, `and`, or a comma, or with `then` after a single person | `list-in-prose`, `command-in-run` |
| A command not spliced, spliced from a string variable, spliced inside a longer item, spliced for a `person`, first item dropped, or not spliced in a check | `command-in-run`, `single-pass`, `defaults` |
| A default ignored, preferred over an answer, or preferred over a deferred answer | `defaults`, `deferred`, `onboarding` |
| Empty text, an empty list, or a list of empty items treated as settled, a Flag `false` treated as unanswered, or a `false` default | `defaults`, `value-kinds` |
| An empty default used, or a deferred default ignored | `defaults`, `deferred` |
| Another wording for a marker, the same marker for deferred and unanswered, or `$name` left for an unanswered one | `deferred`, `gap-report`, `onboarding`, `defaults` |
| The Text `deferred` treated as a deferred answer | `deferred` |
| A gap not reported, reported as an error, with no section, with a bare question, or with the wrong code | `deferred`, `defaults`, `gap-report`, `onboarding` |
| A variable the text does not use reported | `gap-report`, `left-alone`, `nothing-to-fill` |
| Problems in order of use, sorted by name, or one for each use | `gap-report`, `deferred` |
| An onboarding variable reported as a gap, its answer or default ignored, a deferred one written as a blank, or another blank wording | `onboarding` |
| An `env` map made when there was none | `command-in-run`, `defaults`, `gap-report` |
| The answers changed, or an input list or map changed | The inputs-unchanged test |
| The result sharing a list, map, or pointer with the input | The shares-nothing test |
| A cache by section id | `command-in-run`, `defaults`, `deferred` |
| The variables dropped from the result | `command-in-run`, `defaults`, `deferred` |
| After the first build: a read one character past a `$` that panics at the end of a field, for `$`, `$$`, and `$$$` | `dollar-at-end` |
| After the first build: a name that stops before `0`, `9`, `a`, or `z`, takes a character just outside those ranges or an upper case letter, starts at a fixed place, or fails at the end of the text | `name-edges` |
| After the first build: an exact `$name` test that stops at the second character when it is a digit or an underscore, so a command such as `$e2e_test` is not spliced | `command-values` |
| After the first build: a command held as Text split on spaces, or a person list spliced, in a step or in a check | `command-values`, `command-in-run` |
| After the first build: a condition's `isSettled` pointer shared with the input or dropped, in `includeIf` or `askIf` | `settled-conditions`, and the tests that change the result's `isSettled` targets and check the input |
