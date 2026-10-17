# Plan encoder contract

Status: draft, for a person to approve. Component: `plan/encoder`. Issue: [#39](https://github.com/8thlight-daveloper/onboard-kit/issues/39).

The plan encoder turns a Plan into the bytes of a plan file. It is a pure function. It reads no file, writes no file, and does not change its argument. The same Plan always gives the same bytes. See [0030](../../../docs/decisions/0030-self-contained-plan-file.md), the [domain model](../../../docs/reference/domain-model.md), and [0015](../../../docs/decisions/0015-functional-core-and-imperative-shell.md).

## Interface

`Encode(plan) -> (bytes, problems)`.

- When every value in the plan can be written, `problems` is empty and `bytes` is the canonical file.
- When a value cannot be written, `bytes` is empty and `problems` lists each one. A problem has severity `error`.

## The canonical file

The same rules apply everywhere. The goal is that two people, or two versions of the tool, never write different bytes for the same plan, so plan files diff cleanly in version control.

### Document

- UTF-8, LF line endings, exactly one newline at the end, and no trailing spaces on any line.
- No tabs, no `---` or `...` markers, and no comments.
- Block style throughout, indented two spaces. A list under a key is indented two spaces, and each item starts with `- `.
- Never an anchor, alias, tag, or merge key.

### Keys

A key is written in the order the tables below give. A key with a zero value is omitted. Zero means an empty string, `false`, an empty or absent list or map, or a Value of kind none. A condition test written `isAnswered` is the exception: it is written whenever it is set, including `false`. A map entry, such as an answer, is always written, even when its value is `false`. An answers section with no answers is omitted. Inside an unknown field, an empty map is written `{}` and an empty list is written `[]`, because that data belongs to someone else and is kept as it is. Two exceptions: `schemaVersion` is always written, and `provenance` is omitted only when both of its fields are empty.

| Type | Key order |
| -- | -- |
| Plan | `schemaVersion`, `provenance`, `variables`, `answers`, `sections`, then the unknown fields |
| Provenance | `templateSet`, `generatedOn` |
| Variable | `name`, `description`, `type`, `multiple`, `options`, `pattern`, `patternHint`, `required`, `default`, `phase`, `askIf` |
| Condition | `all` |
| Clause | `question`, `answer` |
| AnswerTest | `is`, `isNot`, `isOneOf`, `isNotOneOf`, `isAnswered` |
| Section | `id`, `title`, `summary`, `includeIf`, `timing`, `after`, `owner`, `escalation`, `verification`, `body`, `variables`, `steps`, `source`, then the unknown fields |
| Source | `template`, `hash` |
| Step | `id`, `name`, `notes`, `owner`, `timing`, `escalation`, `run`, `open`, `directory`, `env`, `checks` |
| Check | `run`, `ask`, or `env`. Exactly one |
| Answers | `plan`, `sections` |

A key is written by the same rules as a string, so a key such as `n` is quoted. Keys of a map are written in ascending byte order. This covers the plan's answers by variable name, the sections' answers by section id, a step's `env`, and unknown fields.

Unknown fields are written after the known keys, as siblings of them, in ascending byte order. An unknown field whose key is a known key of that type is a problem.

### Scalars

- An integer is written in plain decimal. A boolean is `true` or `false`, lowercase.
- A string is written plain only when all of these hold. Otherwise it is written in double quotes.
  - It is not empty, and does not start or end with a space.
  - Its first character is an ASCII letter, `$`, `/`, or `<`. A string that starts with any other character, such as `ü`, is quoted, and its characters are still written as they are.
  - It has no control character and no newline.
  - It does not contain `: ` or ` #`, and does not end with `:`.
  - It is not, ignoring case, one of `y`, `n`, `yes`, `no`, `true`, `false`, `on`, `off`, `null`, `inf`, `nan`.
  - It is not the Text value `unknown-yet`.
- Inside double quotes, `\`, `"`, newline, tab, and return are written `\\`, `\"`, `\n`, `\t`, and `\r`. Other control characters are written `\xNN`. Every other character is written as it is.
- A date such as `2026-10-15` starts with a digit, so it is always quoted.

### Multi-line strings

A string with a newline is written as a literal block when all of these hold, and otherwise in double quotes.

- No line has trailing whitespace or a tab.
- No character is a control character, other than the newlines.
- The first line is not empty and does not start with a space.
- The text ends with no newline or with exactly one.

The marker is `|-` when the text has no final newline, and `|` when it has exactly one. A block's lines are indented two spaces more than their key, and an empty line inside the text is an empty line in the file.

### Lists

A list is written as a block list, one item per line, even for short lists of words. A list that is directly inside a list is written `- - item`. An empty list is omitted.

### Answers

| Value | Written as |
| -- | -- |
| Text | A string, by the string rules |
| Flag | `true` or `false` |
| List | A block list of strings, or `[]` when the list is empty, wherever a Value is written |
| Unknown | The plain word `unknown-yet` |

A Value of kind none inside answers is a problem: `encode.empty-answer`.

### Unknown field values

The unknown fields of a plan or section hold maps, lists, strings, booleans, integers, decimal numbers, and null. They are written by the same rules, with null as `null`.

In Go, an integer is an `int` or an `int64`, and a decimal number is a `float64`. A decimal is written in plain notation, with `.0` added when it has no fractional part, so `2.0` stays a decimal. NaN, infinity, every other number kind, and every other kind of value are problems: `encode.unsupported-value`. The path of a value inside a list ends in the item's position, such as `x[0]`.

## Problems

| Code | When |
| -- | -- |
| `encode.empty-answer` | A Value of kind none is an answer, or is an item of `isOneOf` or `isNotOneOf` |
| `encode.empty-item` | A variable, section, step, or condition clause in a list would write no keys. Dropping it would shift the list, so it is a problem |
| `encode.ambiguous-check` | A Check has none of `run`, `ask`, `env`, or more than one |
| `encode.unknown-field-collision` | An unknown field has the key of a known field. It is reported once, and its value is not examined |
| `encode.unsupported-value` | An unknown field holds a value that cannot be written |

A problem names the place it was found in `detail.path`. The path is the place's address in the plan, written with dots for fields and a number in brackets for a list item, such as `sections[1].steps[0].checks[0]`, `answers.plan.organization`, or `answers.sections.s.name`. An unknown field at the plan level is its key alone, such as `answers`. Problems come in the order their places occur in the canonical file.

## Properties

These hold for every plan the contract tests use.

- **Deterministic.** Two calls with the same plan give identical bytes.
- **Inputs unchanged.** The plan is the same after the call as before.
- **Well formed.** The output ends in exactly one newline, has no `---`, and has no line with trailing whitespace.

## Not covered

- Reading plan files, which is the decoder.
- Writing the bytes to a file, which a driver does.
- Checking that a plan is valid. An encoder writes any plan it can, and the plan validator judges it.
- Decimal numbers written in exponent form, and decimals too large or small for plain notation.

## Wrong implementations tried

While drafting, these wrong encoders were written as scratch work outside the repo, and the contract tests failed each one. A reference encoder written from this page alone passed every test, which showed that the fixtures agree with the rules. One fixture error was found that way: the key `n` must be quoted.

| Wrong implementation | Caught by |
| -- | -- |
| Map keys in iteration order, not sorted | The golden cases that have maps, and the determinism test |
| `yes`, `no`, `on`, and other reserved words written plain | `strings`, `key-order`, `unknown-fields` |
| Lists in flow style, `[a, b]` | `key-order`, `acme` |
| Zero fields written, such as `required: false` | `omit-zero`, `multiline`, and others |
| The argument changed during the call | The inputs-unchanged test |
| No final newline | Every golden case, and the well-formed test |
| `True` and `False` for booleans | `answers`, `conditions` |
| Unknown fields dropped | `unknown-fields`, and the collision cases |
| Keys in alphabetical order, not canonical order | Every case with more than one key |
| A date written plain | `minimal`, `key-order`, `strings` |
| Every multi-line string as a literal block | `multiline`, the well-formed test |
| A string with two final newlines as `\|+` | `multiline` |
| An empty answer written as `null`, no problem | `empty-answer` |
| The Text `unknown-yet` written plain | `answers`, `strings` |
