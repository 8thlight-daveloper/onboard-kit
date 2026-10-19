# Variables

Part of the [section template schema](README.md).

A variable is a named, typed value that a person fills in once. Where text uses it, it is written `$name` and filled from the answer when the document is assembled, which is why it is also called a placeholder. Each variable is its own question, with its own description, type, and default. See [0030](../../decisions/0030-self-contained-plan-file.md).

## Scope

A variable belongs to the plan or to one section.

- **Section variables** are declared in the section's `variables`. Each is its own variable with its own answer, even when two sections use the same name.
- **Plan variables** are declared in the plan's `variables`. They are the axis variables `team_state` and `discipline`, document details such as `organization` and `team`, and settings such as `output_style`. The default plan carries the built-in ones. Their values live in `answers.plan`.
- A plan variable is the same type as a section variable, with these limits. It may not use `phase: onboarding`, because the plan holds one answer for everyone. It should not use `askIf`, and the validator reports one that does.
- **A section uses only the variables it declares.** A `$name` the section does not declare is a lint error, even when a plan variable has that name. So a section stays self-contained if a plan variable is later removed. The finder still groups questions with the same name, type, and options across scopes, and offers a plan variable's answer as a suggestion. The consultant types `organization` once, and a section's own `organization` variable offers it.
- A question id is `<scope>.<name>`, such as `plan.team_state` or `access.repo_url`.

## Syntax

- A name starts with a lowercase letter, then continues with at least two more lowercase letters, digits, or underscores, and ends at the first other character. The pattern is `[a-z][a-z0-9_]{2,}`. A `$` followed by a shorter run, such as `$x`, is not a variable and stays as written. See [0040](../../decisions/0040-blank-text-and-variable-name-length.md).
- `$name` parses as a plain string in both YAML flow and block context, and it shows in GitHub's Markdown preview. The other delimiters we tried either fail to parse or load wrong data silently. `{{ name }}` becomes a nested mapping if left unquoted.

### Dollar signs

A `$` must be followed by a valid name or by another `$`. Anything else is a lint error, so a typo such as `$Repo_url` cannot pass as plain text.

| Write | Renders |
| -- | -- |
| `$repo_url` | The section's `repo_url` value |
| `$$5` | `$5` |
| `$$STAGING_API_TOKEN` | `$STAGING_API_TOKEN` |

Environment variables are the common case. A shell variable such as `$HOME` or `$STAGING_API_TOKEN` in prose is not a variable, because uppercase does not match the name rule, so write it with `$$`. Commands are argument lists with no shell, so no shell expands anything in `run`. The tool fills `$name` in each item before the command runs. To require or set one, see [environment variables in steps](steps.md#environment-variables).

## Deferred answers

`deferred` lets a person move past a question they cannot answer yet and come back to it later. The plan stays valid and assemblable in the meantime. It renders as a visible gap with an owner and timing, so the missing value is not lost, and a user interface can list these questions on the review screen. It is accepted for every type, as its own entry and not as free text. It satisfies a required question, as a settled answer does. An unanswered question does not.

## Fields

| Field | Meaning |
| -- | -- |
| `name` | The `$name` the section uses |
| `description` | What the question asks, in words |
| `type` | `string` (the default), `boolean`, `choice`, `command`, or `person` |
| `options` | The allowed values of a `choice` |
| `pattern` | A regular expression a `string` must match |
| `patternHint` | The expected format in words, shown when a value does not match |
| `required` | Whether the question must have an answer. Defaults to false |
| `multiple` | Whether the variable takes several answers, held as a list. Defaults to false. Only `person` allows it for now, and lint rejects the rest |
| `default` | A suggestion offered when the question is asked. It is never an answer and nothing fills it in, see [0045](../../decisions/0045-defaults-are-suggestions.md). It must satisfy the type, options, and pattern |
| `phase` | `planning` (the default) or `onboarding`. See below |
| `askIf` | A [question condition](conditions.md): ask this question only when other answers match |

## Types

| Type | Allowed entries | Free entry |
| -- | -- | -- |
| `string` | Any text, checked against `pattern` if there is one | Yes |
| `boolean` | `true` or `false`, lowercase | No |
| `choice` | One of `options` | No |
| `command` | An argument list. It expands as a list inside `run`, and joins with spaces in prose | Yes, as a list |
| `person` | One person, written as a text line. See [person](#person) | Yes, as a line |

- `pattern` uses RE2 syntax, a portable subset with no backreferences or lookaround, so every language can run it.
- A URL check belongs in `pattern`. A clone address such as `git@host:org/repo.git` is not a standard URL, so a built-in `url` type would reject it. The secrets check also looks inside a URL for an embedded credential.

## Person

A `person` variable names someone who can help with something. One answer is one text line, written `Name, email, chat handle`, such as `Jane Doe, jane@acme.example, @jane`. With `multiple: true` the answer is a list of those lines, one person each.

```yaml
- name: source_control_grantor
  type: person
  multiple: true
  description: Who grants source control access
  required: true
```

- **Two ways to reach each person.** The prompt suggests an email and a chat handle. The tool counts the comma-separated parts after the name, and a person with fewer than two gets a warning. A warning never stops the answer.
- **A backup.** A required `multiple` variable needs at least one person, or `deferred`. A `multiple` variable with exactly one person gets a warning, because that person may be out of office.
- **Ask once.** Same-name `person` variables across sections group, and an earlier entry is offered as a suggestion, so a person is typed once.
- **Rendering.** In the document the lines appear in order, one per line in a list, or joined with a comma and the word then in a sentence.
- **Privacy.** A plan holds real contact details, and it stays in `output/`, which git ignores. Examples and tests use `.example` addresses.
- **Direction.** Text lines are the MVP. The intended end state is a structured record per person, with named fields, so the tool can check an email address, the onboarding helper can link or message someone, and a person is edited in one place. A form to add a person and a list of people already entered would replace the lines. The declaration stays `type: person` with `multiple: true`, and only the stored answer changes, through a schema version bump and a migration. The open questions and a toolkit spike are in [#46](https://github.com/8thlight-daveloper/onboard-kit/issues/46).

## Phase

`phase` says which process asks the question.

- **`planning`** is asked while the consultant builds the plan, and the value is stored in the plan.
- **`onboarding`** is asked of each newcomer while they onboard. The plan does not hold the value, because each person has their own. Planning skips it: it is not asked and does not block. The assembled document shows it as a blank, `<name>`. Storing and asking for these answers waits for the onboarding runtime, [#28](https://github.com/8thlight-daveloper/onboard-kit/issues/28).
