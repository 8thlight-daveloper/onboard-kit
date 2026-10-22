# Variables

Part of the [section template schema](README.md).

A variable is a named, typed value that a person fills in once. Where text uses it, it is written `$name` and filled from the answer when the document is assembled, which is why it is also called a placeholder. Each variable is its own question, with its own description, type, and default. See [0030](../../decisions/0030-self-contained-plan-file.md).

## Scope

A variable belongs to the plan or to one section.

- **Section variables** are declared in the section's `variables`. Each is its own variable with its own answer, even when two sections use the same name.
- **Plan variables** are declared in the plan's `variables`. They are the axis variables `team_state` and `discipline`, document details such as `organization` and `team`. The default plan carries the built-in ones. Their values live in `answers.plan`.
- A plan variable is the same type as a section variable, with these limits. It should not use `askIf`, and the validator reports one that does.
- **A section uses only the variables it declares.** A `$name` the section does not declare is a lint error, even when a plan variable has that name. So a section stays self-contained if a plan variable is later removed. The finder still groups questions with the same name, type, and options across scopes, and offers a plan variable's answer as a suggestion. The consultant types `organization` once, and a section's own `organization` variable offers it.
- A question id is `<scope>.<name>`, such as `plan.team_state` or `access.repo_url`.

Each scope declares a variable name at most once, even when the declarations
have identical metadata. Validation reports a duplicate instead of choosing
one, and a submission to that ambiguous question is refused. The same name
in the plan and in a section, or in different sections, is still valid. See
[0110](../../decisions/0110-reject-ambiguous-plans.md).

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
| `description` | What the question asks, in words. Write it as a question the answerer can respond to, and say where the answer appears when that is not obvious |
| `example` | One illustration of a good answer, such as `Engineering manager`. It shows as placeholder text and in the detail of the question. It is never an answer or a default, and nothing fills it in. Part of the template, so a change to it changes the content hash |
| `type` | `string` (the default), `boolean`, `choice`, `command`, `contact`, `role`, or `team` |
| `options` | The allowed values of a `choice` |
| `pattern` | A regular expression a `string` must match |
| `patternHint` | The expected format in words, shown when a value does not match |
| `required` | Whether the question must have an answer. Defaults to false |
| `multiple` | Whether the variable takes several answers, held as a list. Defaults to false. Only `contact` and `role` allow it, and lint rejects the rest |
| `default` | A suggestion offered when the question is asked. It is never an answer and nothing fills it in, see [0045](../../decisions/0045-defaults-are-suggestions.md). It must satisfy the type, options, and pattern |
| `audience` | `onboarder` (the default) or `onboardee`. See below |
| `askIf` | A [question condition](conditions.md): ask this question only when other answers match |

## Types

| Type | Allowed entries | Free entry |
| -- | -- | -- |
| `string` | Any text, checked against `pattern` if there is one | Yes |
| `boolean` | `true` or `false`, lowercase | No |
| `choice` | One of `options` | No |
| `command` | An argument list. It expands as a list inside `run`, and joins with spaces in prose | Yes, as a list |
| `contact` | An id from the directory's contacts. See [directory types](#directory-types) | No |
| `role` | An id from the directory's roles | No |
| `team` | An id from the directory's teams. Single only | No |

- `pattern` uses RE2 syntax, a portable subset with no backreferences or lookaround, so every language can run it.
- A URL check belongs in `pattern`. A clone address such as `git@host:org/repo.git` is not a standard URL, so a built-in `url` type would reject it. The secrets check also looks inside a URL for an embedded credential.

## Directory types

`contact`, `role`, and `team` variables are picked from the organization's directory, which holds contacts, roles, and teams. See [0080](../../decisions/0080-directory.md). An answer is the id of an entry and never a name, so a changed name or a new holder reaches every plan without anyone editing an answer.

```yaml
- name: source_control_grantor
  type: role
  description: Which role grants source control access?
  required: true
- name: general_contacts
  type: contact
  multiple: true
  description: Who can anyone ask for help with anything?
  required: true
```

- **Pick, do not type.** The walkthrough offers the directory's entries. The onboarder adds a missing contact on the Contacts screen. The `team` question's choices are the directory's teams.
- **An unknown id is unanswered.** An answer that names an id the directory no longer has counts as not answered, so a required question blocks until it is filled again.
- **A role resolves to people.** A role gives the contacts who hold it. A team can name its own holders for a role, and those win for that team. The team is the answer to the plan's `team` variable. A role with no holder shows as vacant, and the document lists it as a gap.
- **Order.** A `multiple` question keeps its order. An escalation list is read in that order.
- **Rendering.** A contact shows as `Name, @slack`, or `Name, email` when there is no Slack handle, or just `Name`. Several are joined with a semicolon. A vacant role shows as `Role name (vacant)`.
- **Privacy.** The directory holds real contact details, so the org plan's location must require authentication. Examples and tests use `.example` addresses.

## Audience

`audience` says who answers the question. See [0055](../../decisions/0055-variable-audience.md).

- **`onboarder`** is the default. It is asked of the person who builds and maintains the plan. It holds facts the organization knows and every newcomer shares, such as a repository, who grants access, or who to contact.
- **`onboardee`** is asked of the person being onboarded. It holds what the newcomer knows about themselves, such as their account name, their team, or the kind of work they do, and choices that decide which sections apply to them. Each newcomer's answers live in their own plan file.

The onboarder writes the question and its options, and the onboardee gives the answer. Three questions decide the audience: who knows the answer, whether it varies per person, and whether the onboardee should be able to change it. If not, it is an onboarder variable. Every onboardee question must change what some section says or shows.

Scope and audience are separate: a plan variable or a section variable can have either audience. The assembled document shows an onboardee variable with no value as a blank, `<name>`. Every unanswered or deferred variable the text uses is also reported as a gap with its audience, including onboardee variables. A default remains a suggestion and never removes that gap until a person accepts it. See [0055](../../decisions/0055-variable-audience.md) and [0120](../../decisions/0120-defaults-require-acceptance.md).
