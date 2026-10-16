# Placeholders

Part of the [section template schema](README.md).

A placeholder is a named blank in a section, written `$name`, filled from an answer when the document is assembled. Each section's placeholder is its own question, with its own description, type, and default. Values are per section. See [0030](../../decisions/0030-self-contained-plan-file.md).

## Syntax

- A name starts with a lowercase letter, then continues with lowercase letters, digits, or underscores, and ends at the first other character. The pattern is `[a-z][a-z0-9_]*`.
- `$name` parses as a plain string in both YAML flow and block context, and it shows in GitHub's Markdown preview. The other delimiters we tried either fail to parse or load wrong data silently. `{{ name }}` becomes a nested mapping if left unquoted.

### Dollar signs

A `$` must be followed by a valid name or by another `$`. Anything else is a lint error, so a typo such as `$Repo_url` cannot pass as plain text.

| Write | Renders |
| --- | --- |
| `$repo_url` | The section's `repo_url` value |
| `$$5` | `$5` |
| `$$STAGING_API_TOKEN` | `$STAGING_API_TOKEN` |

Environment variables are the common case. A shell variable such as `$HOME` or `$STAGING_API_TOKEN` in prose is not a placeholder, because uppercase does not match the name rule, so write it with `$$`. Commands are argument lists with no shell, so a variable never expands inside `run`. To require or set one, see [environment variables in steps](steps.md#environment-variables).

## Unknown values

`unknown-yet` lets a person move past a question they cannot answer yet and come back to it later. The plan stays valid and assemblable in the meantime. It renders as a visible gap with an owner and timing, so the missing value is not lost, and a user interface can list these questions on the review screen. It is accepted for every type, as its own entry and not as free text. It satisfies a required question.

## Fields

| Field | Meaning |
| --- | --- |
| `name` | The `$name` the section uses |
| `description` | What the question asks, in words |
| `type` | `string` (the default), `boolean`, `choice`, or `command` |
| `options` | The allowed values of a `choice` |
| `pattern` | A regular expression a `string` must match |
| `patternHint` | The expected format in words, shown when a value does not match |
| `required` | Whether the question must have a value. Defaults to false |
| `default` | The value used when none is given. It must satisfy the type, options, and pattern |
| `phase` | `planning` (the default) or `onboarding`. See below |
| `askIf` | A [question condition](conditions.md): ask this question only when other answers match |

## Types

| Type | Allowed entries | Free entry |
| --- | --- | --- |
| `string` | Any text, checked against `pattern` if there is one | Yes |
| `boolean` | `true` or `false`, lowercase | No |
| `choice` | One of `options` | No |
| `command` | An argument list. It expands as a list inside `run`, and joins with spaces in prose | Yes, as a list |

- `pattern` uses RE2 syntax, a portable subset with no backreferences or lookaround, so every language can run it.
- A URL check belongs in `pattern`. A clone address such as `git@host:org/repo.git` is not a standard URL, so a built-in `url` type would reject it. The secrets check also looks inside a URL for an embedded credential.

## Phase

`phase` says which process asks the question.

- **`planning`** is asked while the consultant builds the plan, and the value is stored in the plan.
- **`onboarding`** is asked of each newcomer while they onboard. The plan does not hold the value, because each person has their own. Planning skips it: it is not asked and does not block. The assembled document shows it as a blank, `<name>`. Storing and asking for these values waits for the onboarding runtime, [#28](https://github.com/8thlight-daveloper/onboard-kit/issues/28).
