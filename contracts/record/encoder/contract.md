# Record encoder contract

Status: draft by an agent. Not reviewed by a person. Component: `record/encoder`. Issue: none yet.

The record encoder turns a Record into the bytes of a record file. It is a pure function. It reads no file, writes no file, and does not change its argument. The same Record always gives the same bytes. See [0090](../../../docs/decisions/0090-plan-and-onboarding-record.md), the [domain model](../../../docs/reference/domain-model.md), and [0015](../../../docs/decisions/0015-functional-core-and-imperative-shell.md).

## Interface

`Encode(record) -> (bytes, problems)`.

- When every value in the record can be written, `problems` is empty and `bytes` is the canonical file.
- When a value cannot be written, `bytes` is empty and `problems` lists each one. A problem has severity `error`.

## The canonical file

The document, scalar, multi-line string, list, and answer rules are the ones of the [plan encoder contract](../../plan/encoder/contract.md), and apply here without change. That covers: UTF-8, LF, one final newline, block style with two-space indent, no markers, comments, anchors, or tags; when a string is plain and when it is quoted, including the reserved words and the Text `deferred`; the escapes inside quotes; literal blocks; block lists in the order held; the written form of each kind of Value; and the unknown field value kinds and number rules. Map keys, including the keys of answers and of unknown fields, are written in ascending byte order. Only what differs or is new is below.

### Keys

A key with a zero value is omitted, as in the plan encoder. Zero means an empty string, an empty or absent list or map, or a Value of kind none. A map entry, such as an answer, is always written. Inside an unknown field, an empty map is `{}` and an empty list is `[]`.

| Type | Key order |
| -- | -- |
| Record | `schemaVersion`, `kind`, `account`, `startDate`, `firstOpenedAt`, `plan`, `answers`, `outcomes`, then the unknown fields |
| Plan reference | `id`, `hash`, `url` |
| Answers | `plan`, `sections` |
| Outcome | `section`, `outcome`, `on`, `sectionHash`, `note`, `differently` |

- `schemaVersion` is always written, as the integer the record holds. The encoder does not judge the version.
- `kind` is always written, right after `schemaVersion`, as the plain text `record`. It is not a field of the Record.
- `account` and `startDate` are plain text fields written by the string rules. A date such as `2026-10-22` starts with a digit, so it is quoted. A `startDate` that is not a date is written as held.
- `plan` is the plan reference. It is omitted when `id`, `hash`, and `url` are all empty. Each empty field of it is omitted on its own. A hash such as `5d3c1f` starts with a digit and is quoted, and one such as `a1b2c3` is plain. A URL is plain when it starts with a letter, `$`, `/`, or `<`.
- `answers` is written as the plan encoder writes a plan's answers: `plan` is a map of variable name to Value, `sections` is a map of section id to a map of variable name to Value, and every map is in ascending byte order. An empty `plan` map is omitted, a section whose map is empty is omitted, and `answers` is omitted when nothing is left.
- `outcomes` is a block list of mappings, in the order the record holds them. The encoder never sorts it, never removes an item, and keeps repeated sections. It is omitted when the list is empty.
- An outcome writes each field by the string rules and omits an empty one. A Timestamp such as `2026-10-22T14:05:00Z` starts with a digit, so it is quoted. `note` and `differently` may hold a newline, and are written by the multi-line rules, as a block indented two spaces more than their key.
- The `outcome` word is written as the text held, by the string rules, with one exception: the word `deferred` is written plain. In an `outcome` it cannot be taken for the answer Deferred. Any other word, including one that is not `verified`, `failed`, or `deferred`, is written as held, with no problem. A validator judges it.

Unknown fields are written after the known keys, as siblings of them, in ascending byte order. An unknown field whose key is one of the eight keys of a Record is a problem. A record has no sections, variables, directory, provenance, or id, so those keys, and any other, are written as any unknown field.

`firstOpenedAt` is optional Timestamp text, written exactly as held after `startDate` and before `plan`. Empty text is omitted. No date or timestamp is validated or inferred, and legacy `startDate` never supplies this field. Its known-key collision is rejected even when the field is empty. See [0230](../../../docs/decisions/0230-first-use-record-compatibility.md).

Every record string and mapping key containing U+2028 or U+2029 is double-quoted and writes those characters as `\u2028` or `\u2029`. This also applies when the text contains ordinary newlines: use quoted escapes, never a literal block. Other text stays exact. This overrides the inherited literal-character rule for record serialization, extending [0200](../../../docs/decisions/0200-preserve-unicode-line-separators.md).

## Problems

| Code | When |
| -- | -- |
| `encode.empty-answer` | A Value of kind none is an answer |
| `encode.empty-item` | An outcome would write no keys. Dropping it would shift the list, so it is a problem. The path is such as `outcomes[1]` |
| `encode.unknown-field-collision` | An unknown field has one of the keys `schemaVersion`, `kind`, `account`, `startDate`, `firstOpenedAt`, `plan`, `answers`, `outcomes`. It is reported once, and its value is not examined |
| `encode.unsupported-value` | An unknown field holds a value that cannot be written |

A problem names its place in `detail.path`: dots for fields and a number in brackets for a list item, such as `answers.plan.organization`, `answers.sections.s.name`, `outcomes[1]`, or `x[0]`. An unknown field is its key alone. Problems come in the order their places occur in the canonical file, so answers, then outcomes, then unknown fields, with map keys in ascending byte order.

## Properties

These hold for every record the contract tests use.

- **Deterministic.** Two calls with the same record give identical bytes.
- **Inputs unchanged.** The record is the same after the call as before. That includes the order of the outcomes.
- **Well formed.** The output ends in exactly one newline, has no `---`, and has no line with trailing whitespace or a tab.

## Not covered

- Reading record files, which is the decoder. The decoder requires `kind: record` and refuses any other.
- Writing the bytes to a file, which a driver does.
- Checking that a record is valid. That covers the form of a date or timestamp, whether the `outcome` word is one of the three, whether a section exists in the plan, and the order of the outcomes. A validator and the recorder judge these.
- Decimal numbers written in exponent form, and decimals too large or small for plain notation.

## Decisions to know

1. **hard.** The `outcome` word `deferred` is written plain, `outcome: deferred`, while every other text field writes the text `deferred` in quotes. The shared decisions list `deferred` as an outcome word and the plan encoder quotes it everywhere. The quotes in a plan keep an answer that is the text `deferred` apart from an answer that is deferred. An `outcome` holds no answer, so the quotes would only make the common case look odd. Only the exact word `deferred` is affected. `Deferred` is plain by the ordinary rule.
2. **cheap.** An outcome that would write no keys is `encode.empty-item` at its path, as in the plan encoder. Writing `- {}` or dropping it are the other choices, and dropping shifts the list. The docs say outcomes are written in the order held and do not name this case.
3. **cheap.** An outcome word that is not one of the three is written as held, with no problem. The docs say the decoder reads any text without validating and a validator checks later. The encoder follows the same split.
4. **cheap.** An unknown field named `schemaVersion`, `kind`, `account`, `startDate`, `firstOpenedAt`, `plan`, `answers`, or `outcomes` collides. The docs say unknown fields are written back as for a plan, and the plan encoder reports a key that is a known key. `id`, `sections`, `directory`, `provenance`, and `variables` do not collide, because the shared decisions say a record has no such fields.
5. **cheap.** A section with no answers is omitted from `answers.sections`, and an empty `plan` or `sections` map is omitted, as in the plan encoder. A map entry that holds an answer is always written.
6. **cheap.** A `plan` reference that holds only one field writes that field. It is omitted only when all three are empty.
7. **cheap.** The encoder writes any `schemaVersion` as held, and does not refuse a version other than 1. Refusing is the decoder's guard.

## Wrong implementations tried

Each row is a plausible wrong implementation and the cases that fail it.

| Wrong implementation | Caught by |
| -- | -- |
| Argument or answer maps written in iteration order, not sorted | `ada`, `answers`, `unknown-fields`, `empty-answer`, and the determinism test |
| `kind` never written, or written after `account` and `startDate` | `empty`, `ada`, `identity`, and the kind test |
| An unknown field named `kind` or `schemaVersion` not reported | `collision`, `problem-order`, and the collision-once test |
| `id` as an unknown field reported as a collision | `unknown-fields`, `collision` |
| The `outcome` word `deferred` quoted, or every outcome word quoted | `ada`, `outcome-fields`, `outcome-order`, `quoting`, and the outcome word test |
| The text `deferred` plain in another field or in an answer | `quoting`, `answers`, the answer kinds test |
| Everything quoted | Every case with plain text |
| A date or timestamp written plain | `ada`, `identity`, `outcome-fields` |
| Outcomes sorted by section, in the copy or in the argument | `outcome-order`, `quoting`, the inputs-unchanged test |
| Outcomes with a repeated section dropped | `outcome-order` |
| An outcome's keys in another order | `ada`, `outcome-fields`, `quoting` |
| `differently` never written | `ada`, `outcome-fields`, `multiline` |
| Zero fields written, such as `note: ""` or `plan:` with no fields | `omit-zero`, `empty`, `partial-items`, `plan-url-only` |
| An empty `answers` written, or an empty section in `answers.sections` | `empty`, `omit-zero`, `sections-only` |
| An empty list answer omitted or written in flow style, or booleans as `True` | `answers`, the answer kinds test |
| Every multi-line string as a literal block, or a string with two final newlines as `\|+` | `multiline`, the well-formed test |
| An empty outcome dropped, or written as `{}`, with no problem | `empty-item`, `problem-order`, the file order test |
| An empty answer written, with bytes returned alongside problems | `empty-answer`, and every problem case |
| A colliding field's value examined as well | The collision-once test |
| Unknown fields dropped | `unknown-fields`, the number kind tests |
| A stub that returns nothing | Every golden case |
| Drops or swaps firstOpenedAt, infers it from legacy startDate, or validates timestamp text | `first-use`, `TestFirstOpenedAtIsRepresentationalText` |
| Lets unknown firstOpenedAt shadow the known field, including when empty | `TestFirstOpenedAtCollisionIsRejectedEvenWhenTheFieldIsEmpty` |
| Emits Unicode separators literally or in a literal block | `unicode-separators`, `TestRecordUnicodeSeparatorsUseExplicitEscapes` |
