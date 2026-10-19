# Plan decoder contract

Status: approved. Component: `plan/decoder`. Issue: [#38](https://github.com/8thlight-daveloper/onboard-kit/issues/38).

The plan decoder turns the bytes of a plan file into a Plan. It is a pure function. It reads no file, writes no file, and does not change its argument. The same bytes always give the same result. See [0030](../../../docs/decisions/0030-self-contained-plan-file.md), the [domain model](../../../docs/reference/domain-model.md), and [0015](../../../docs/decisions/0015-functional-core-and-imperative-shell.md).

The decoder is the reverse of the [plan encoder](../encoder/contract.md) for every file the encoder writes. It also reads files a person has edited by hand. It is strict about the format guards in 0030 and free about style.

## Interface

`Decode(bytes) -> (plan, problems)`.

- When the file is acceptable, `problems` is empty and `plan` is the Plan.
- When it is not, `plan` is the zero Plan and `problems` lists each one. A problem has severity `error` and a non-empty message.

## What the decoder does

- It reads version 1 and refuses every other version. There is no migration. See Version.
- It reads every known field into the field of the same name in the domain model.
- It keeps unknown fields on a plan and on a section. It keeps them as plain data.
- It reports the format guards of 0030: the spellings of booleans, numbers, and null, anchors and aliases, and duplicate keys.
- It does not validate. A variable of type `colour`, a check with two of `run` and `ask`, two sections with one `id`, a `timing` that is not a timing, and an answer to a variable that does not exist all pass through unchanged. The plan validator judges them. It does not fill defaults: an absent `type` stays empty, and an absent `enabled` on a section stays none.

## Version

The decoder reads `schemaVersion` before anything else, and the supported version is 1.

| Found | Problem | Path | Detail |
| -- | -- | -- | -- |
| Absent, or not a canonical integer: quoted, a decimal, a boolean, a list, `null`, any number spelling listed under Numbers, or an integer that does not fit in 64 bits | `version.invalid` | `schemaVersion` | |
| An integer below 1 | `version.too-old` | `schemaVersion` | `found`, `supported` |
| An integer above 1 | `version.too-new` | `schemaVersion` | `found`, `supported` |
| Two `schemaVersion` keys, whatever their values | `format.duplicate-key` | `schemaVersion` | |

`found` is the integer's text as written, and `supported` is `1`. When there is a version problem, it is the only problem. A newer file may use anything, so the rest of the file is not read. `-0` is a canonical integer below 1. An anchor on a `schemaVersion` value that is a good version is a `format.alias` problem. An anchor or an alias on a bad version is only the version problem, and an alias in place of the value is `version.invalid`.

## Problems

Problems come in the order of their places in the file, and all are reported. The place of a problem is where the last key or list item of its path starts, so a key and its value are one place. Two problems at one place come in the order of the table below. Two with the same code at one place come parent first: a map starts where its first key starts, so a problem about the map comes before one about that key. Two exceptions: a file problem is the only problem, and a version problem is the only problem. The check order is file, then version, then the rest. File problems are checked in this order, and the first one found is the only one: the bytes are not UTF-8, a YAML syntax error anywhere in the file, an empty first document, a second document, and a top level that is not a map. A syntax error in any document wins over every other file problem. A document is empty when it has no node, or when its only node is a plain scalar with no text and no anchor, whatever its tag: no bytes, only comments, `---` with nothing after it, or `--- !!null`. A document that holds `null` as text, or an anchor, is not empty. UTF-16 is not UTF-8.

An anchor is reported wherever it is written, even on a node that is also null or of the wrong type. The anchor problem comes first, and the node's other problem follows it. A problem names its place in `detail.path`, as the encoder does: dots for fields, and a number in brackets for a list item, such as `sections[1].steps[0].checks[0]`, `answers.plan.organization`, or `x-map.a`. A file problem has no path. The path of a duplicate key is the second and later keys. The path of an anchor or an alias is the node it is written on.

| Code | When | Path |
| -- | -- | -- |
| `file.not-yaml` | The bytes are not UTF-8, or are not YAML | none |
| `file.not-a-plan` | The file is empty or holds only comments or a lone `---`, or its top level is not a map | none |
| `file.multiple-documents` | The file has a second YAML document, even an empty one | none |
| `version.invalid`, `version.too-old`, `version.too-new` | See Version | `schemaVersion` |
| `format.boolean-spelling` | A plain value is a boolean spelling other than `true` and `false` | The value |
| `format.number-spelling` | See Numbers | The value |
| `format.null-spelling` | A plain value is `~`, `Null`, or `NULL` | The value |
| `format.alias` | An anchor, an alias, or a merge key `<<` is used. A merge key is one problem at the key, whatever its value. Each merge key is its own problem, and a repeated one is not also a duplicate key. The value of a merge key is not read. A quoted `"<<"` is an ordinary key. An anchor on a key is reported at the key's path | The node, or the key for a merge key |
| `format.duplicate-key` | A map has the same key twice. A quoted key and a plain key with the same text are the same key | The second and later keys |
| `format.wrong-type` | A known field holds a kind the domain model does not allow for it, or a key in any map is not a scalar. See Known fields. A section clause with no `enabled`, see Section clauses | The field, or the list item |
| `format.unknown-field` | A map that cannot keep unknown fields has a key it does not know | The key |
| `format.empty-answer` | An answer, or an item of `isOneOf` or `isNotOneOf`, is null | The answer or item |
| `format.empty-item` | An item of any other list is null, including an item of a List Value. Dropping it would shift the list. An item that is an empty map is not null and is read as a zero value | The item |

A value of the wrong type is not read, so nothing inside it is reported. The value of a duplicate key is read like any other. A value is checked for its spelling before its type, so a plain `yes` where a boolean belongs is `format.boolean-spelling`, not `format.wrong-type`. A canonical number where a boolean or a list belongs, and a plain `true` or `false` where a string belongs, is `format.wrong-type`. A string belongs in a list of strings and in the items of a List Value, so a plain `true` there is `format.wrong-type` too.

## Reading the file

### Style is free

The decoder accepts anything YAML allows, so a hand edit never fails for style.

- Comments, a leading `---`, any indent, any key order, flow style, single and double quotes, folded and literal blocks, and a missing final newline.
- LF or CRLF line endings, and a UTF-8 byte order mark at the start.
- A key is its text as written. The plain keys `on`, `y`, and `1` are the text `on`, `y`, and `1`. The format guards apply to values and not to keys. A key that is not a scalar, which includes an alias used as a key, is `format.wrong-type`. Its path and its place are those of the map that holds it, so it is ordered where that map starts.
- A tag such as `!!str` is ignored. A scalar is read by its text and its style alone, so `!!str yes` is still the spelling `yes`.
- The decoder does not judge the characters of a quoted string, so a string may hold an escaped NUL or bell.
- A plain date such as `2026-10-15` is text, in every place a string is read.

### Scalars

A plain scalar is a value written without quotes. A quoted scalar, and a literal or folded block, is always text, whatever it says: `"yes"`, `"123"`, `"null"`, and `"deferred"` are text. A plain scalar is read by these rules, in this order.

| Plain scalar | Read as |
| -- | -- |
| `true` or `false` | A boolean |
| `yes`, `no`, `on`, `off`, `y`, `n`, `True`, `TRUE`, `False`, `FALSE`, or any other case of these | `format.boolean-spelling` |
| `null`, or nothing at all after the key | Null |
| `~`, `Null`, or `NULL`, and no other case: `nULL` is text | `format.null-spelling` |
| A number | See Numbers |
| Anything else, including `deferred` | Text |

### Numbers

A canonical integer is an optional `-`, then `0` or digits that do not start with `0`. It must fit in 64 bits. A canonical decimal is a canonical integer, a `.`, and one or more digits. So `7`, `-7`, `0`, `0.5`, `-0.5`, and `2.0` are canonical.

Every other plain scalar that a YAML reader could take as a number is a `format.number-spelling` problem, wherever it is. The list is exact.

- A leading zero before more digits: `007`, `-01`, `0123`.
- A leading `+`: `+1`.
- A leading or trailing `.`: `.5`, `5.`.
- An underscore between two digits of the whole part, the fraction, or the digits of a `0x`, `0o`, or `0b` number: `1_000`, `1_0.5`, `1.0_1`, `0x1_F`, `0o1_7`, `1_0e3`. An exponent holds no underscore, so `1e1_0` is text.
- An exponent: `1e3`, `1.5e-3`.
- A `0x`, `0o`, or `0b` prefix: `0x1F`, `0o17`, `0b101`.
- `.inf`, `-.inf`, and `.nan`, in any case.
- A sign, `+` or `-`, in front of any form above. The letters in a prefix or an exponent may be upper case: `0X1F`, `1E3`.
- A canonical decimal too large for a 64-bit float. A decimal too small for one is read as 0, and is not covered.
- An integer that does not fit in 64 bits.

A canonical integer or decimal is also `format.number-spelling` where a string or a Value is read, because the text was meant. `123` in an answer is a problem and `"123"` is text. In an unknown field, a canonical integer is read as `int64` and a canonical decimal as `float64`. A decimal keeps its kind, so `2.0` is a `float64`. A scalar that only starts like a number, such as `12ab`, `1.2.3`, `v1.0`, `0.5.x`, or `2026-10-15`, is text. So are these: `1_`, `1__0`, `_1`, `1_.5`, `0x_1`, `e3`, `1e`, `1e+`, `0x`, `0xZZ`, `0b2`, `0o8`, `.`, `-yes`, `12:30:45`, `1,000`, and digits of other scripts.

### Values

A Value is read in an answer, in a `default`, in `is` and `isNot`, and in each item of `isOneOf` and `isNotOneOf`.

| In the file | Value |
| -- | -- |
| A quoted scalar, block, or plain text | Text |
| `true` or `false`, plain | Flag |
| A plain `deferred` | Deferred |
| A list of strings | List, and an empty list is a List with no items, not none |
| Null in an answer or an item | `format.empty-answer` |
| Null as `default`, `is`, or `isNot`, or an absent key | None |
| A map, or a list holding a list or a map | `format.wrong-type` |

The word `deferred` is the Deferred value only when it is a whole plain Value. Quoted, it is text. As an item of a list, or anywhere a string is read, such as a `title`, it is text.

### Known fields

A string field takes text. A boolean field takes a plain `true` or `false`. A section's `enabled` is a boolean field that is none when it is absent, empty, or null, and a Flag pointer otherwise, so `enabled: true` is kept as on and is not the same as absent. A list field takes a list of strings. A key whose value is null, empty, or absent is the zero value. A list field with no items, empty or null or absent, is nil. A list keeps its items in file order and keeps repeats. The decoder never sorts a list, reverses it, or removes an item, whether it is a list of text, of values, or of sections, steps, checks, variables, or clauses. The List kind of a Value is different: an empty list there is a List with no items, not None.

A null value in a step's `env` is the empty string. A map that is present is not nil even when it is empty, and a map that is absent or null is nil. This holds for `answers.plan`, `answers.sections`, a step's `env`, and an unknown field's map. A scope under `answers.sections` is always a map: when it is null or empty, it is kept as an empty map. A `source` that is present is a Source, even when it holds nothing but null. An `includeIf` or `askIf` that is null is none, and one that is present is a Condition, even when it holds no clauses. An `isSettled` that is null is unset.

An empty list or map inside an unknown field is kept as `[]` and `{}`. A null there is kept: as a null under its key, or as a null item of a list.

### Section clauses

A clause is a map in the `all` list of an `includeIf` or an `askIf`. It has one of two forms, and a key that belongs to the other form is not known to it. See [0050](../../../docs/decisions/0050-section-switch.md).

| Form | Known keys | Read as |
| -- | -- | -- |
| Section clause | `section`, `enabled` | `Section` and `Enabled` |
| Question clause | `question`, `answer` | `Question` and `Answer` |

- A clause is a section clause when it has a `section` key, with any value and in any key order. Otherwise it is a question clause, including a clause with no keys that it knows.
- A clause of one form that holds a key of the other form is not a second kind of problem. That key is `format.unknown-field` at the key, and its value is not read. So `question` and `answer` in a section clause are reported, and so is `enabled` in a clause with no `section`. The format has no code for a mixed clause, and this one is the nearest.
- `section` is a string and `enabled` is a boolean, read by the rules of Known fields. A plain `yes` for `enabled` is `format.boolean-spelling`. A quoted `"true"`, a number, a list, or a map is `format.wrong-type` at the key. A `section` that is a list or a plain `true` is `format.wrong-type` at the key.
- The model cannot tell an absent `enabled` from `false`, and reading it as `false` would switch a section off by accident. A section clause whose `enabled` is absent, empty, or `null` is `format.wrong-type` at the clause. A `~` is the null spelling alone. At one place, `format.wrong-type` comes before `format.unknown-field`, as the table gives.
- An empty `section`, whether `""`, `''`, or null, is read as the empty string, and `enabled` is read as written. The decoder does not validate, so the plan validator judges an id that names no section.

### Unknown fields

A key that is not in the model's field list for a plan or a section is an unknown field. It is kept, with its value, under that plan's or section's `UnknownFields`. Its value is a map, a list, a string, a boolean, a null, an `int64`, or a `float64`, and the format guards apply inside it. In a map that cannot keep them, an unknown key is `format.unknown-field`. The value of such a key is not read, so nothing inside it is reported. Every such key is reported, a repeated one each time, together with `format.duplicate-key` for the repeat. Those maps are `provenance`, `source`, a variable, a condition, a clause, an answer test, `answers`, a step, and a check.

## Properties

These hold for every file the contract tests use.

- **Deterministic.** Two calls with the same bytes give the same result.
- **Inputs unchanged.** The bytes are the same after the call as before.
- **Nothing shared.** Changing a returned Plan changes no other result.

## Not covered

- Reading files, which a driver does.
- Checking that a plan is valid, which is the plan validator. Decoding templates is the section loader.
- Migrating an older version. The migrator is deferred until the first breaking change. See 0030.
- Line and column numbers in a problem. The path names the place.
- `%YAML` directives, and other YAML features that no plan uses.
- A document end marker `...`.
- Escaping in a path. A key that holds a `.` or a `[` makes an ambiguous path.
- A plain scalar that ends in `?` inside a flow list or map, such as `{ask: Done?}`. One YAML library cannot read it. Quote it.
- Whether `Encode(Decode(file))` equals the file. The encoder contract and this one share their canonical fixtures in `contracts/plan/testdata/`, so each half is tested on its own against the same data, and neither test imports the other component.

## Wrong implementations tried

Each row names a plausible wrong implementation and the groups of cases that fail it. How the contract was checked, and with what scores, is recorded in the issue, not here.

| Wrong implementation | Caught by |
| -- | -- |
| Unknown fields dropped, at the plan or at a section | `unknown-fields`, `hand-edited`, `zero-and-null` |
| Plain `yes`, `no`, `on`, `off` read as booleans, or only some cases recognized | `boolean-spelling` |
| Anchors, aliases, or merge keys accepted, inside unknown fields too | `alias` |
| A duplicate key accepted, first or last winning, or ignored inside an unknown map, or `"a"` and `a` told apart | `duplicate-key` |
| Version 0, a negative version, or a version above 1 accepted | `version-zero`, `version-negative`, `version-too-new` |
| A quoted, decimal, hex, or out-of-range version accepted | `version-quoted`, `version-decimal`, `version-hex`, `version-out-of-range` |
| A newer file reported with its other problems | `version-too-new-with-other-problems` |
| Two `schemaVersion` keys, the last one winning | `duplicate-version` |
| A date read as a time value, or reformatted | `hand-edited` |
| Quoted `deferred` read as Deferred, or plain read as text | `answers`, `strings`, `hand-edited`, `zero-and-null` |
| `1.0`, `123`, `0123`, `0x1F`, `1e3`, `.inf`, or an over-long integer read as text or as a number | `number-spelling` |
| `~`, `Null`, `NULL` read as null | `null-spelling` |
| Defaults filled for `type`, `timing`, or `phase` | `conditions`, `zero-and-null`, `acme` |
| Stopping at the first problem, or sorting problems by path | `many-problems`, `duplicate-key` |
| A partly filled Plan returned with problems | every problem case |
| A key `on` or `y` read as a boolean, or flagged | `hand-edited` |
| The input bytes changed, including a byte order mark stripped in place | the inputs-unchanged test |
| State kept between calls | the results-share-nothing test, and the problem cases |
| Unknown numbers all `float64`, all `int64`, or `int` | `unknown-fields`, `hand-edited` |
| An empty list Value read as nil, or a nil map made empty, or the reverse | `answers`, `null-maps`, `answers-plan-only`, `answers-null`, `unvalidated` |
| Only the first document read | `multiple-documents` |
| An unknown key silently dropped from a variable, step, check, condition, clause, `source`, or `provenance` | `unknown-field` |
| A value validated: an unknown variable type refused | `unvalidated` |
| A file refused for no final newline, CRLF, or a byte order mark | `no-final-newline`, the Windows line endings test, the byte order mark test |
| A quoted `"true"` read as a boolean | `wrong-type` |
| A wrong-type map scanned for more problems | `wrong-type` |
| A null item dropped from a list | `empty-item` |
| `source: null` read as an empty Source, or an empty Source read as nil | `zero-and-null` |
| `isSettled: null` read as false | `zero-and-null` |
| A plain `deferred` read as Deferred in a list or in a string field | `zero-and-null` |
| A merge key with a map value read as an ordinary key | `alias` |
| Problems found in a merge key's value | `alias` |
| A syntax error in a later document hidden by the second-document problem, or the top-level type checked first | `parse-error-in-second-document`, `non-map-with-second-document` |
| UTF-16 read as YAML | `utf16` |
| Two `schemaVersion` keys judged by their values | `duplicate-version-bad-values` |
| A signed, upper-case, or very large non-canonical number read as text or a number | `number-spelling` |
| `nULL` read as a null spelling | `null-spelling` |
| A plain `true` in a List Value read as a flag or as text | `wrong-type-answers` |
| A bad version also reported as an alias, or an alias in place of the version followed to its target | `version-too-new-with-anchor`, `version-alias` |
| A null item dropped from a list in an unknown field | `hand-edited` |
| A null `answers.sections` read as an empty map | `answers-sections-null` |
| Problems at one place in the wrong order | `same-place`, `wrong-type-same-place` |
| A list sorted, reversed, or stripped of repeats, in any array field | `list-order` |
| An anchor on a null node dropped, or on the top-level map dropped | `anchored-null`, `anchor-on-root` |
| A key that is not a scalar skipped, or read as the text of its first node | `complex-key`, `complex-key-in-section`, `alias-key` |
| A quoted empty string taken for an empty document | `quoted-empty-document-then-second` |
| A plain text where a condition clause, an answer test, or a step belongs crashes the decoder | `wrong-type-items` |
| A quoted `true`, number, `null`, or empty string in an unknown field read as a boolean, a number, or null | `hand-edited` |
| Entries after a key that is not a scalar skipped | `complex-key-then-problem` |
| A tag changing how a scalar is read | `tags`, `tagged-null-document-then-second` |
| The value of an unknown key read for problems | `unknown-field-values` |
| An underscore ignored in part of a number | `number-edges` |
| A repeated merge key also reported as a duplicate | `merge-keys` |
| A syntax error after an empty document hidden by another file problem | `syntax-error-after-empty-document`, `empty-documents`, `anchored-empty-document-then-second` |
| A second document ignored when it is empty | `multiple-documents-empty` |
| A lone `---` or a file of comments read as a plan | `lone-marker`, `comments-only` |
| A present but empty `includeIf` read as none | `zero-and-null` |
| A canonical number in a list field reported as a spelling | `wrong-type` |
| An anchor on `schemaVersion` hiding the version or the alias | `alias-on-version` |
| A section `enabled` ignored, or kept as an unknown field | `switch`, `enabled-values`, `key-order` |
| An explicit `enabled: true` dropped as the default | `switch`, `enabled-values` |
| An empty or null `enabled` on a section read as off | `enabled-values` |
| A quoted, numeric, list, or map `enabled` read as a boolean, in a section or a section clause | `enabled-wrong-type` |
| A plain `yes`, `no`, `Off`, `TRUE`, or `n` read as the switch | `enabled-spelling` |
| A mixed clause read without a problem, or its form told by the first key | `clause-forms`, `clause-missing-enabled` |
| `enabled` accepted in a question clause | `clause-forms` |
| A section clause with no, empty, or null `enabled` read as `false` | `clause-missing-enabled` |
| The missing `enabled` reported at the `section` key and not at the clause | `clause-missing-enabled` |
| A section clause `enabled: false` read as `true`, or not kept | `conditions`, `switch`, `enabled-values`, `key-order` |
| An empty `section` refused, or a section id that is a plain `true` or a list read as text | `clause-empty-section`, `enabled-wrong-type` |
| An answer or a variable named `enabled` read as the switch | `enabled-values` |
