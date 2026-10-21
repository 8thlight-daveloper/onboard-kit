# Record decoder contract

Status: draft by an agent. Not reviewed by a person. Component: `record/decoder`. Issue: none yet.

The record decoder turns the bytes of an onboarding record file into a Record. It is a pure function. It reads no file, writes no file, and does not change its argument. The same bytes always give the same result. See [0090](../../../docs/decisions/0090-plan-and-onboarding-record.md), the [domain model](../../../docs/reference/domain-model.md), and [0015](../../../docs/decisions/0015-functional-core-and-imperative-shell.md).

A record file is YAML in the same dialect as a plan file. This contract is the [plan decoder contract](../../plan/decoder/contract.md) with a different top level. Every rule below marked "as the plan decoder" is the rule of that contract, sections by the same name, and is not repeated here. A rebuild reads both documents.

## Interface

`Decode(bytes) -> (record, problems)`.

- When the file is acceptable, `problems` is empty and `record` is the Record.
- When it is not, `record` is the zero Record and `problems` lists each one. A problem has severity `error` and a non-empty message.

## Rules that are the same

- **Reading the file**: Style is free, Scalars, Numbers, and Values, as the plan decoder. A plain date or timestamp is text. A plain `deferred` is the Deferred value only as a whole answer, and is text everywhere else, including the `outcome` word.
- **Version**: as the plan decoder. The decoder reads `schemaVersion` first, 1 is the only version, and a version problem is the only problem.
- **Problems**: the codes `file.not-yaml`, `file.multiple-documents`, and every `format.*` code, their paths, their order of reporting, and their order at one place, as the plan decoder. One code differs: an empty file, a file of comments, a lone `---`, and a top level that is not a map are `file.not-a-record`, not `file.not-a-plan`. The order of checks is file, then version, then kind, then the rest.
- **Format guards**: spellings of booleans, numbers, and null, anchors, aliases, merge keys, duplicate keys, and keys that are not scalars, as the plan decoder.
- **Answers**: `answers` is read exactly as a plan's `answers`, with the same nil and empty rules, the same `format.empty-answer` and `format.empty-item`, and with `plan` and `sections` as its only known keys. The decoder does not check that an answer names a variable or a section.
- **Unknown fields**: kept as plain data, with the same kinds, and the format guards apply inside them.
- **Properties** and **Not covered**: as the plan decoder.

## Known fields

The top-level keys a record knows are `schemaVersion`, `kind`, `account`, `startDate`, `plan`, `answers`, and `outcomes`. They sit in any order. Keys are exact: `Account` and `startdate` are unknown fields.

| Key | Read as |
| -- | -- |
| `account` | Text, `Record.Account` |
| `startDate` | Text, `Record.StartDate`, with no check that it is a date |
| `plan` | A map of `id`, `hash`, `url`, each text, into `Record.Plan` |
| `answers` | Answers, see above |
| `outcomes` | A list of maps, into `Record.Outcomes` |

Every string field, here and below, is read as a plan's `title` is: a quoted or plain string is text, absent, empty, or null is the empty string, and a plain `deferred` is text. So `account: "42"` is text and a plain `account: 42` is `format.number-spelling`. A hash such as `1e3` or `5d3c1f` is plain text only when YAML cannot read it as another type, and a file quotes one that starts with a digit. A plain `yes` is `format.boolean-spelling`. A list, a map, or a plain `true` is `format.wrong-type`, at the field.

### Plan reference

`plan` is read like a plan's `origin`. An absent, null, or empty `plan`, and a `plan` with only empty values, is the zero PlanReference. Each key is read alone and none is filled from another. A `plan` that is not a map is `format.wrong-type` at `plan`, and nothing inside it is read. A key other than `id`, `hash`, and `url` is `format.unknown-field` at the key, and its value is not read. The decoder does not check that `id` is a UUID, that `hash` is a hash, or that `url` is a path or a URL.

### Outcomes

`outcomes` is a list. Each item is a map with the keys `section`, `outcome`, `on`, `sectionHash`, `note`, and `differently`, in any order. `outcome` is read into `Outcome.Kind`, and `on` into `Outcome.On`. Each is a string field.

- The list keeps its order and its repeats. The decoder never sorts it or removes an item. The recorder keeps the outcomes sorted, see [0085](../../../docs/decisions/0085-walking-sections-and-recording-outcomes.md).
- The `outcome` word is any text. The decoder does not check it against `verified`, `failed`, and `deferred`, and does not check `on` or `sectionHash`. A validator judges them later. The word is read as written, so `Verified` stays `Verified`.
- An item that is an empty map is a zero Outcome. An item that is null is `format.empty-item`. An item that is not a map is `format.wrong-type` at the item, such as `outcomes[1]`.
- A key other than the six is `format.unknown-field` at the key, such as `outcomes[0].comment`, and its value is not read.
- `outcomes` that is absent, null, or empty is nil. One that is not a list is `format.wrong-type` at `outcomes`.
- A string inside an item is read like any string field and is not trimmed. A literal block keeps its final newline.

## Kind

`kind` is part of the file and not of the Record. The encoder always writes it. The decoder requires it, and refuses a file whose `kind` is anything but the text `record`.

| Found | Problem | Path | Detail |
| -- | -- | -- | -- |
| No `kind` key, or `kind` null or with nothing after it | `kind.missing` | `kind` | |
| Any other scalar than the text `record`, such as `plan`, `Record`, `record `, `""`, `yes`, or `1` | `kind.not-a-record` | `kind` | `found` is the text as written |
| A list, a map, or an alias | `kind.not-a-record` | `kind` | `found` is empty |
| Two `kind` keys, whatever their values | `format.duplicate-key` | `kind` | |

- The text must be exactly `record`, in lower case, with nothing around it. Quoting does not matter: `"record"` and `'record'` are the text `record`.
- A kind is judged by its text before the spelling rules, so a plain `yes` is the text `yes` and gives `kind.not-a-record`, not `format.boolean-spelling`.
- A kind problem is the only problem. A plan file that has `kind: plan`, or no `kind`, is refused without reading its sections.
- The message of a kind problem is not fixed, and it says what a person needs. For `kind.not-a-record` it names the kind found, and for both it says that a record needs `kind: record`. For example, a plan file given to a command that wants a record is told that the file has kind `plan`.
- A good `kind` with an anchor is `format.alias` at `kind`, and the file is read on. The version rule for an anchor on `schemaVersion` is the same.

## Not covered

- What the plan decoder does not cover, see its contract.
- Checking a record: that the account is a slug, that `startDate` is a date, that an `on` is an RFC 3339 time in UTC, that an `outcome` word is known, that a `section` names a section of the plan, that the outcomes are sorted, that `plan.id` names the plan in hand, and that an answer belongs to an onboardee question. Validating and reconciling a record are other components.
- Joining a record with a plan, which is the joiner.
- Whether a `kind` that is null and also anchored is `kind.missing` or `kind.not-a-record`.
- Whether `Encode(Decode(file))` equals the file. The record encoder will share fixtures with this contract through `contracts/record/testdata/`, and neither test imports the other component.

## Decisions to know

1. **hard.** An unknown key in `plan` or in an outcome is `format.unknown-field`, and is not kept. Only the record itself keeps unknown fields, because the model gives `UnknownFields` to the Record and to nothing inside it. 0090 says "Unknown fields are kept on load and written back, as for a plan". A record written by a newer tool that adds a key to an outcome would be refused by an older one, and the version guard does not help when the version is unchanged.
2. **hard.** `account`, `startDate`, the plan reference, and every outcome field are strings, so a plain `42`, `007`, `1e3`, `yes`, or `true` is a problem and a file quotes them. An account such as `42` and a hash such as `1e3` must be quoted by a person who edits by hand. The encoder will quote them. 0090 shows `startDate: "2026-10-22"` quoted and `hash: 5d3c...` plain.
3. **hard.** A file with no `kind` is refused, with `kind.missing`, because a plan file may also carry no `kind`. 0090 says "A record also carries `kind: record`, and a plan carries `kind: plan` or no `kind`."
4. **cheap.** The kind codes are `kind.missing` and `kind.not-a-record`, and the empty-file code is `file.not-a-record`. 0090 says a message "says which it is" and names no codes.
5. **cheap.** A kind is judged by its text, so `kind: yes` is `kind.not-a-record` with `found` `yes`. A kind that is not a scalar has an empty `found`.
6. **cheap.** The version is checked before the kind, and a kind problem is the only problem. This matches the version rule that a newer file may use anything. A plan with a bad version is told about the version.
7. **cheap.** Outcomes keep file order and repeats. 0085 says the recorder keeps them sorted by section, so the decoder does no sorting and no merging of two outcomes for one section.
8. **cheap.** An empty `plan` map is the zero PlanReference, and an empty outcome map is a zero Outcome, the same as other empty maps in the plan decoder.
9. **cheap.** A key such as `sections`, `variables`, `directory`, `origin`, `provenance`, or `id` at the top of a record is an unknown field and is kept, as 0090 says, and its content is read with the format guards only.

## Wrong implementations tried

Each row names a plausible wrong implementation and the cases that fail it. The inherited rules are tested with the same wrong implementations in the plan decoder contract. Here, the rows that matter are the ones this contract adds, and the shared ones that the record fixtures also catch.

| Wrong implementation | Caught by |
| -- | -- |
| A file with no `kind` accepted | `kind-absent` |
| A kind compared without regard to case, or trimmed | `kind-wrong-case`, `kind-trailing-space` |
| `kind: plan` accepted | `kind-plan` |
| The kind checked before the version | `version-before-kind`, `version-too-new` |
| Two `kind` keys judged by the last one | `kind-duplicate-values` |
| A kind problem reported together with the problems of the rest of the file | `kind-plan`, `kind-duplicate`, `kind-alias` |
| A version problem reported together with a kind problem | `version-too-new`, `version-before-kind` |
| `kind` kept as an unknown field | `ada`, `answers`, and every accepted case |
| A message that does not name the kind found, or does not say what a record needs | the kind message test |
| An anchor on `kind` not reported, or an alias for `kind` followed | `kind-anchor`, `kind-alias` |
| A null or blank `kind` mixed up | `kind-null`, `kind-empty`, `kind-blank` |
| A non-scalar `kind` read as missing | `kind-list`, `kind-alias` |
| A plain `kind: yes` reported as a boolean spelling | `kind-spelling` |
| Unknown fields dropped | `unknown-fields` |
| Plan-only keys such as `sections` or `id` refused | `unknown-fields` |
| An unknown key in the plan reference or in an outcome kept or ignored | `unknown-field`, `problem-order` |
| Outcomes sorted, or repeats removed, or an empty item dropped | `outcomes-edges`, `strings` |
| An outcome word validated | `outcomes-edges`, `strings`, `boolean-spelling` |
| A string trimmed | `strings` |
| A null outcome item dropped | `empty-values` |
| A null answers scope read as nil, or an empty `answers` read as non-zero | `answers`, `answers-null`, `answers-empty-map`, `zero-and-null` |
| A plain `deferred` answer read as text | `answers` |
| An old code `file.not-a-plan` used | `empty-file`, `comments-only`, `lone-marker`, `top-level-list` |
| Problems sorted by path, or stopped at the first | `problem-order`, `same-place`, `boolean-spelling`, and the other multi-problem cases |
| A partly filled Record returned with problems | every problem case |
| A canonical number accepted as an account | `number-spelling` |
| Fields swapped or ignored: `startDate` filled from `account`, `on` and `sectionHash` swapped, `plan.id` dropped | `ada`, `key-order`, `strings` |
| A duplicate key accepted inside an outcome | `duplicate-key` |
| State kept between calls, so two results share maps | the results-share-nothing test |
| The bytes changed, including a byte order mark stripped in place | the inputs-unchanged test |
