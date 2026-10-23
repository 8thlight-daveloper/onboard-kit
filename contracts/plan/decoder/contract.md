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
- It reads every known field into the field of the same name in the domain model, including the plan's `directory` into its Directory, and its `planUrl` and `resultsPath` into its URL and ResultsPath.
- It accepts a `kind` of `plan` and drops it, and it refuses any other kind. See Kind.
- It keeps unknown fields on a plan and on a section. It keeps them as plain data.
- It reports the format guards of 0030: the spellings of booleans, numbers, and null, anchors and aliases, and duplicate keys.
- It does not validate. A variable of type `colour`, a check with two of `run` and `ask`, two sections with one `id`, a `timing` that is not a timing, and an answer to a variable that does not exist all pass through unchanged. The plan validator judges them. It does not fill defaults: an absent `type` stays empty, and an absent `enabled` on a section stays none.

## Version

The decoder reads `schemaVersion` before anything else, and the supported version is 1.

| Found | Problem | Path | Detail |
| -- | -- | -- | -- |
| Absent, or not a canonical integer: quoted, a decimal, a boolean, a list, `null`, any number spelling listed under Numbers, or an integer that does not fit in 64 bits | `version.invalid` | `schemaVersion` |  |
| An integer below 1 | `version.too-old` | `schemaVersion` | `found`, `supported` |
| An integer above 1 | `version.too-new` | `schemaVersion` | `found`, `supported` |
| Two `schemaVersion` keys, whatever their values | `format.duplicate-key` | `schemaVersion` |  |

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
| `format.wrong-kind` | The plan has a `kind` whose text is not `plan`. See Kind. A repeated `kind` is judged at each key | The `kind` key |

`format.wrong-kind` has `detail.found`, the text of the kind as written, and `detail.expected`, which is `plan`. Its message names the kind found and says that a plan is expected. A value of the wrong type is not read, so nothing inside it is reported. The value of a duplicate key is read like any other. A value is checked for its spelling before its type, so a plain `yes` where a boolean belongs is `format.boolean-spelling`, not `format.wrong-type`. A canonical number where a boolean or a list belongs, and a plain `true` or `false` where a string belongs, is `format.wrong-type`. A string belongs in a list of strings and in the items of a List Value, so a plain `true` there is `format.wrong-type` too.

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

Under [0260](../../../docs/decisions/0260-prerequisites-and-action-order.md),
requires is a known section list of strings, independent of legacy after.
Read it into Requires, never UnknownFields, using every existing list and
string guard. Keep exact strings, file order and repetitions; absent, null,
empty values and empty lists give nil. Do not resolve placeholders, validate
references or manufacture Requires from After. The exact lowercase key is
section-only: root requires and case variants are preserved unknown data;
a step's requires is unknown. Raw repeated keys are errors even when values
agree. [Requires fixtures](testdata/requires.json) and
[tests](requires_test.go) pin reading, problem paths/order, separate-result
isolation and reuse across inputs independently of the encoder.

Under [0250](../../../docs/decisions/0250-bootstrap-content-and-projection.md),
bootstrap is a known optional boolean on saved sections. Read it into Bootstrap:
absent, empty, null and false are false; only canonical plain true is true.
Apply the existing boolean spelling/type, tag/style, anchor/alias, duplicate-key,
path and source-order problem rules. Preserve the section's other fields.
Do not retain a recognized marker in UnknownFields or infer it from timing,
conditions, answers, defaults, source, parents or children. Root bootstrap and
case-variant section keys remain ordinary preserved unknown data. A step's
bootstrap is unknown, not a marker. [Bootstrap fixtures](testdata/bootstrap.json)
and [tests](bootstrap_test.go) independently pin these distinctions.

A string field takes text. A boolean field takes a plain `true` or `false`. A section's `enabled` is a boolean field that is none when it is absent, empty, or null, and a Flag pointer otherwise, so `enabled: true` is kept as on and is not the same as absent. A list field takes a list of strings. A key whose value is null, empty, or absent is the zero value. A list field with no items, empty or null or absent, is nil. A list keeps its items in file order and keeps repeats. The decoder never sorts a list, reverses it, or removes an item, whether it is a list of text, of values, or of sections, steps, checks, variables, or clauses. The List kind of a Value is different: an empty list there is a List with no items, not None.

A section's `parent` is a string field, read exactly like `title`: a quoted or plain string is text, and absent, empty, or null is the empty string, which is a root section. A plain `deferred`, a plain date, and a quoted `yes` are text. A plain `yes`, `~`, or number is a spelling problem and a plain `true`, a list, or a map is `format.wrong-type`, at `sections[n].parent`. It sits in any key order. Its key is exact: `Parent`, `parents`, and `x-parent` are unknown fields. See [0070](../../../docs/decisions/0070-sub-sections.md).

A plan's `id` is a string field, read exactly like `title`: a quoted or plain string is text, and absent, empty, or null is the empty string. It sits in any key order. Its key is exact: `ID` and `x-id` are unknown fields. A plain UUID is text. See [0075](../../../docs/decisions/0075-refreshing-a-personal-plan.md).

A plan's `directory` is a map of three lists: `contacts`, `roles`, and `teams`. It is read into the plan's Directory, and the lists keep the file's order. See [0080](../../../docs/decisions/0080-directory.md) and the Directory section below.

A plan's `planUrl` and `resultsPath` are string fields, read exactly like `title`. A quoted or plain string is text, and absent, empty, or null is the empty string. They go into the plan's URL and ResultsPath, and sit in any key order. A plain `https://git.example/plan.obk.yaml`, a plain `records/{account}/record.obk.yaml`, a plain date, and a plain `deferred` are text. A quoted `"yes"`, `'123'`, and `"{account}/record.obk.yaml"` are text, and case and spaces are kept as written, so nothing is trimmed, lowercased, or changed. A plain `yes`, `~`, or number is a spelling problem and a plain `true`, a list, or a map is `format.wrong-type`, at `planUrl` or `resultsPath`. A template that starts with `{` is written in quotes, because a plain one is a flow map: `resultsPath: {account}` is a map, so it is `format.wrong-type` at `resultsPath`, and `{account}/record.obk.yaml` with text after the brace is a syntax error, `file.not-yaml`. The decoder does not check that a URL is a URL or that a template holds `{account}`. Their keys are exact: `planURL`, `PlanUrl`, `ResultsPath`, `results_path`, and `x-planUrl` are unknown fields of the plan and are kept. A variable, an answer, or a field of a section with one of these names is ordinary. See [0090](../../../docs/decisions/0090-plan-and-onboarding-record.md).

A plan's `origin` is a map of four string fields: `planId`, `source`, `contentHash`, and `refreshedOn`. Each is read like `title`, the date as the text written, with no check that it is a date. A plain `deferred` is text in each. An absent key is the empty string. An `origin` that is absent, null, empty, or holds only empty values is the zero Origin, because the model cannot tell them apart. An `origin` that is not a map is `format.wrong-type` at `origin`, and nothing inside it is read. Its keys are exact, in any order. An unknown key in it is `format.unknown-field`, as in `provenance`. The format guards apply the same: an anchor on `origin` or on a value, a repeated key, and a merge key are reported at their paths. A variable, an answer, a section, or an unknown field named `id` or `origin` is ordinary, and so is `origin` as a key of a section. The decoder does not check that `refreshedOn` is a date, that `planId` is a UUID, or that `source` is a path or a URL.

A variable's `example` is a string field, read like its `description`: a quoted or plain string is text, and absent, empty, or null is the empty string. A plain `deferred` is text there, because a string is not a Value. It sits beside `description` and in any key order, in a variable of the plan and in a variable of a section. The decoder never copies it to or from another field. Its key is exact: `Example` and `examples` are unknown fields.

A null value in a step's `env` is the empty string. A map that is present is not nil even when it is empty, and a map that is absent or null is nil. This holds for `answers.plan`, `answers.sections`, a step's `env`, and an unknown field's map. A scope under `answers.sections` is always a map: when it is null or empty, it is kept as an empty map. A `source` that is present is a Source, even when it holds nothing but null. An `includeIf` or `askIf` that is null is none, and one that is present is a Condition, even when it holds no clauses. An `isSettled` that is null is unset.

An empty list or map inside an unknown field is kept as `[]` and `{}`. A null there is kept: as a null under its key, or as a null item of a list.

### Kind

A plan file may carry a top-level `kind`. A plan file and a record file are told apart by it, and this decoder reads plans only. The record decoder has its own rule for the same key. See [0090](../../../docs/decisions/0090-plan-and-onboarding-record.md).

- `kind` is a string field, read like `title`. Absent, null, empty, and `""` are the same as `plan`: a file with no kind is a plan.
- The text `plan` is accepted and dropped. It is not kept as an unknown field, so it is not in the plan's `UnknownFields`, and the encoder does not write it back. A plain `plan`, `'plan'`, and `"plan"` are the same text.
- Any other text is `format.wrong-kind` at `kind`, with `found` the text as written and `expected` `plan`. The text is compared exactly: `Plan` and a quoted `"plan "` with a trailing space are other kinds, and `record`, `template`, a plain `deferred`, and a quoted `"42"` are refused. The problem is one problem among the others, in the order of its place in the file. The rest of the file is still read, so a record file, whose other keys are unknown fields of a plan, gives the kind problem alone, and a file with a kind problem and a spelling problem elsewhere gives both. The plan returned with a problem is the zero Plan, as always.
- A `kind` that is not text is the usual format problem and has no kind to name: a plain `yes` is `format.boolean-spelling`, a plain `123` is `format.number-spelling`, a plain `true`, a list, or a map is `format.wrong-type`, and an anchor is `format.alias` followed by the node's other problem, all at `kind`. A repeated `kind` is `format.duplicate-key` at the second key, and each value is judged on its own, so `kind: plan` then `kind: record` gives the duplicate and then the wrong kind at one place, and `kind: record` then `kind: plan` gives the wrong kind and then the duplicate.
- Under a version problem the `kind` is not read, because a newer file may use anything. A file with `schemaVersion: 2` and `kind: record` gives the version problem alone.
- The key is the plan's alone and is exact. `Kind`, `kinds`, and `x-kind` are unknown fields of the plan and are kept. `kind` on a section is an unknown field of the section and is kept. A variable, an answer, or a section named `kind` is ordinary. In `provenance`, `origin`, the `directory`, and the other maps that cannot keep unknown fields, `kind` is `format.unknown-field` like any other unknown key.

### Directory

A plan's `directory` is optional and sits in any key order, like `origin`. Each of its three keys takes a list of maps, and each item is read as follows. Every field is a string, read exactly like `title`, so a quoted `"42"`, `"yes"`, and `'@sam'` are text, a plain `deferred`, a plain date, and `U01ABC123` are text, and nothing is trimmed, lowercased, or changed. A plain `@sam` or `#it-help` is not valid YAML as a plain scalar, so a file quotes them.

| Item | Keys | Read as |
| -- | -- | -- |
| Contact | `id`, `name`, `email`, `slack`, `slackId` | `ID`, `Name`, `Email`, `Slack`, `SlackID`, all text |
| Role | `id`, `name`, `holders` | `ID`, `Name`, and `Holders`, a list of text |
| Team | `id`, `name`, `contact`, `holders` | `ID`, `Name`, `Contact`, all text, and `Holders`, a map of role id to a list of text |

- An absent, empty, or null `directory`, and a `directory` with no items, is the zero Directory, because the model cannot tell them apart. A null or empty `contacts`, `roles`, or `holders` is nil. A contact, role, or team that is an empty map is read as a zero item. Nothing is filled in: a contact with only an `id` has no name.
- A team's `holders` that is absent or null is nil, and one that is `{}` is an empty map that is not nil. A role id in it is a key like any other, written as text whatever it says, so `on`, `y`, `1`, `yes`, and `null` are the text as written, and a role id keeps its case. Its value is a list of text, and an entry that is `[]` or null stays in the map with no items. That entry means the role has no holder for that team, and a version that drops it changes who holds the role. The order of the keys in the file is not kept, because a Go map has none.
- A team's `contact` is the id of a directory contact that reaches the team, such as a channel or an on-call alias. It is text, read like `title`, and sits in any key order. Absent, null, empty, and `""` are the empty string. It is kept as written, and a quoted `"42"`, `'true'`, and `"yes"` are text, so is a plain `deferred`, and case and spaces are kept. It is never copied into the team's `holders`, and a team's `holders` never changes because of it, so a team with a `contact` and no `holders`, or with `holders: {}`, or with a vacant entry, is read as written. The key belongs to a team only: `contact` in a contact or in a role is an unknown field.
- A list of holders, in a role or in a team entry, a list of contacts, a list of roles, and a list of teams keep their order and their repeats. The decoder never sorts, reverses, or removes an item.
- Unknown keys are not kept. An unknown key in `directory`, in a contact, in a role, or in a team is `format.unknown-field` at the key, and the value of that key is not read. Keys are exact: `Email`, `slack_id`, and `Contacts` are unknown, and so is a key that belongs to another kind of item, such as `holders` in a contact. There is no `x-` exemption. The format guards apply: an anchor, an alias, a merge key, and a repeated key are reported at their paths, and a key that is not a scalar is `format.wrong-type` at the map that holds it.
- A `directory` that is not a map is `format.wrong-type` at `directory`. A `contacts`, `roles`, or `teams` that is not a list is `format.wrong-type` at that key. An item that is not a map, such as a text, a number, a flag, or a list, is `format.wrong-type` at the item, such as `directory.contacts[2]`, and a null item is `format.empty-item`. A field of the wrong kind is `format.wrong-type` at the field, such as `directory.contacts[0].email`, and a spelling problem is reported at the field. A team's `contact` of the wrong kind is `format.wrong-type` at `directory.teams[0].contact`, and a spelling problem, an anchor, and a repeated key are reported there. A team's `holders` that is not a map is `format.wrong-type` at `directory.teams[0].holders`, and a role entry that is not a list is `format.wrong-type` at `directory.teams[0].holders.<role id>`. A holder is a string at `directory.roles[0].holders[1]` or `directory.teams[0].holders.<role id>[1]`, and a null holder is `format.empty-item` there. Nothing inside a value of the wrong type is read.
- A step's `directory` is its working directory, a string, and is unchanged. A `directory` or `x-directory` on a section is an unknown field of that section, as before. Only a top-level `directory` is the plan's.
- The decoder does not validate the directory. A duplicate id, an id that is not a lowercase slug, an empty id, an email with no `@`, a holder that names no contact, a team `contact` that names no contact, and a team entry for a role that does not exist all pass through unchanged. The plan validator judges them. See [0080](../../../docs/decisions/0080-directory.md), Integrity.
- A newer file may use anything, so under a version problem the `directory` is not read.

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

A key that is not in the model's field list for a plan or a section is an unknown field. It is kept, with its value, under that plan's or section's `UnknownFields`. Its value is a map, a list, a string, a boolean, a null, an `int64`, or a `float64`, and the format guards apply inside it. In a map that cannot keep them, an unknown key is `format.unknown-field`. The value of such a key is not read, so nothing inside it is reported. Every such key is reported, a repeated one each time, together with `format.duplicate-key` for the repeat. Those maps are `provenance`, `origin`, `directory`, a contact, a role, a team, `source`, a variable, a condition, a clause, an answer test, `answers`, a step, and a check.

## Properties

These hold for every file the contract tests use.

- **Deterministic.** Two calls with the same bytes give the same result.
- **Inputs unchanged.** The bytes are the same after the call as before.
- **Nothing shared.** Changing a returned Plan changes no other result.

## Not covered

- Reading files, which a driver does.
- Checking that a plan is valid, which is the plan validator. Decoding templates is the section loader.
- Checking that a `parent` names a section of the plan, is not the section itself, and forms no cycle. A parent that names no section, the section itself, and a cycle all pass through unchanged. The plan validator's structure check judges them.
- Migrating an older version. The migrator is deferred until the first breaking change. See 0030.
- Line and column numbers in a problem. The path names the place.
- `%YAML` directives, and other YAML features that no plan uses.
- A document end marker `...`.
- Escaping in a path. A key that holds a `.` or a `[` makes an ambiguous path.
- A plain scalar that ends in `?` inside a flow list or map, such as `{ask: Done?}`. One YAML library cannot read it. Quote it.
- Whether `Encode(Decode(file))` equals the file. The encoder contract and this one share their canonical fixtures in `contracts/plan/testdata/`, so each half is tested on its own against the same data, and neither test imports the other component.

## Decisions to know

These are the choices the directory change and the team contact change made where the docs were silent.

01. **hard.** A team's `holders` entry that is `[]` or null is kept in the map with no items, and the role is vacant for that team. 0080 says a team "may override who holds a role". Dropping the entry would hand the role back to the organization-wide holders, so a file that says "nobody" would read as "the usual people".
02. **hard.** An unknown key inside `directory`, a contact, a role, or a team is `format.unknown-field`, with no `x-` exemption and no kept data. The task said to treat them like `provenance` and `origin`. A directory a person extends by hand therefore fails loudly.
03. **hard.** A plain `42`, `007`, `yes`, or `true` as an id, a name, or a holder is a spelling or type problem, as for any string. A file quotes them. 0080 says ids are slugs, and a slug such as `42` starts with a digit.
04. **hard.** A team's `contact` is read like any string: a plain `42`, `yes`, or `true` is a spelling or type problem, and a file quotes them. The decoder never checks that it names a contact or is a slug, so `design-oncall` with no such contact passes through. 0080 says a team "may name a `contact`, the id of a directory contact", and puts the check in the validator and the editor.
05. **cheap.** A team's `holders: {}` is an empty map that is not nil, and an absent or null one is nil. The model tells them apart and the encoder writes neither.
06. **cheap.** A null item of a directory list is `format.empty-item`, and an item that is not a map is `format.wrong-type` at the item. This matches the other lists.
07. **cheap.** A contact, role, or team that is an empty map is a zero item and not a problem. This matches an empty section or step.
08. **cheap.** The decoder keeps two items with one id, and does no check of slugs, emails, or holders. 0080 puts these in the plan validator.
09. **cheap.** A team's `contact` is never read as a holder, and an empty or null one is the empty string, the same as an absent one. 0080 says it "never stands in for a role holder".
10. **cheap.** `contact` in a contact or a role is `format.unknown-field`, by the exact-keys rule for the directory.

These are the choices the plan link and kind change made where the docs were silent.

1. **hard.** A kind that is not `plan` is `format.wrong-kind` at `kind`, with `found` and `expected` in the detail. It is an ordinary problem in file order and not a file problem, so the rest of the file is still read and a version problem still hides it. The task left the code open. A record file given to this decoder therefore says what it is, and its other keys, which are unknown fields of a plan, add no problems.
2. **hard.** An absent, null, or empty `kind` is a plan, and the text is compared exactly, with no trim and no case folding. So `kind:` is the same as no kind, and `Plan` is refused. ADR 0090 says an absent kind is a plan. The record decoder requires the key, so the two decoders do not read the same empty key alike.
3. **hard.** `kind: plan` is dropped and not kept as an unknown field. A hand-written `kind: plan` therefore disappears when a tool decodes and encodes the file, and the encoder never writes it.
4. **hard.** `planUrl` and `resultsPath` are plain strings with no check. A URL that is not a URL, a template with no `{account}`, and a path that holds a `..` pass through. A template that starts with `{` has to be quoted in the file, and the decoder reports an unquoted one as a wrong type or as `file.not-yaml`.
5. **cheap.** A `kind` that is not text is the ordinary spelling, wrong-type, or alias problem and gets no `format.wrong-kind`, because there is no text to name.
6. **cheap.** A repeated `kind` is a duplicate key, and each value is judged, so the first of `kind: record` then `kind: plan` is still refused.
7. **cheap.** `kind` in a section is an unknown field of the section and is kept, and `kind` in `provenance`, `origin`, or the directory is `format.unknown-field`. Only the plan owns the key.
8. **cheap.** `format.wrong-kind` comes last in the order of codes at one place, after `format.alias` and `format.duplicate-key`, because it is judged on the value after the value is read.

## Wrong implementations tried

Each row names a plausible wrong implementation and the groups of cases that fail it. How the contract was checked, and with what scores, is recorded in the issue, not here.

| Wrong implementation | Caught by |
| -- | -- |
| Drop Requires, keep it as unknown data, sort or deduplicate it, convert After, skip its list guards, or accept repeated raw keys | [Requires fixtures](testdata/requires.json) and [tests](requires_test.go) |
| Drop the known Bootstrap value or accept quoted true as a boolean | [bootstrap fixtures](testdata/bootstrap.json) |
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
| Defaults filled for `type`, `timing`, or `audience` | `conditions`, `zero-and-null`, `acme` |
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
| An `example` key refused as an unknown field, or dropped, in a plan variable or a section variable | `example-values`, `key-order` |
| An `example` read as a Value, so a plain `deferred` is Deferred or a list is accepted | `example-values`, `example-problems` |
| An `example` that is absent, empty, or null read as the text `null` or as the description | `example-values` |
| An `example` swapped with the description, or filled from it | `example-values`, `key-order` |
| A quoted `yes`, `123`, `null`, or `~` in an `example` read as a boolean, a number, or null | `example-values` |
| A plain `yes`, `~`, `123`, or `0x1F` in an `example` read as text | `example-problems` |
| A list, a map, or a plain `true` or `false` in an `example` read as text or as a flag, or a map scanned for more problems | `example-problems` |
| An anchor on an `example` not reported, or its wrong type not reported after it | `example-problems` |
| A repeated `example` key accepted | `example-problems` |
| `Example` or `examples` accepted as `example` | `example-problems` |
| A section `parent` ignored, or kept as an unknown field | `parent-values`, `key-order` |
| A `parent` that is empty or null read as the text `null`, or an absent one filled from another field | `parent-values` |
| A `parent` checked: a missing section, the section itself, or a cycle refused or cleared | `parent-values` |
| A quoted `yes`, `123`, `null`, or `~` in a `parent` read as a boolean, a number, or null, or a plain `deferred` read as Deferred | `parent-values` |
| A `parent` swapped with the `id` or the `title`, or read from `Parent`, `parents`, or `x-parent` | `parent-values`, `key-order` |
| A plain `yes`, `~`, `123`, `0x1F`, `2.5`, or `Off` in a `parent` read as text | `parent-problems` |
| A list, a map, or a plain `true` or `false` in a `parent` read as text or as a flag, or a map scanned for more problems | `parent-problems` |
| An anchor on a `parent` not reported, or its wrong type not reported after it | `parent-problems` |
| A repeated `parent` key accepted | `parent-problems` |
| A plan `id` or an `origin` ignored, or kept as an unknown field | `origin-values`, `origin-strings` |
| An `origin` read for only some of its four keys, or two of them swapped | `origin-values`, `origin-partial`, `origin-strings` |
| An empty `id`, a null or empty `origin`, or an origin of empty values read as a value, or as the text `null` | `origin-empty`, `origin-empty-map`, `origin-blank` |
| `ID`, `Origin`, or an `x-id` key read as the identity | `origin-values` |
| An origin value changed on the way in, such as trimmed or lowercased | `origin-values`, `origin-strings` |
| An unknown key in `origin` dropped, kept, or read from a different case | `origin-problems`, `origin-duplicate` |
| A wrong-type, spelling, or anchor problem not reported in `id` or in an `origin` value | `origin-problems`, `id-spelling`, `origin-wrong-type`, `origin-scalar` |
| An `origin` that is not a map accepted, or its inside read for more problems | `origin-wrong-type`, `origin-scalar` |
| An anchor on `origin` not reported | `origin-anchor` |
| A repeated `id`, `origin`, or origin key accepted | `origin-duplicate`, `origin-problems` |
| The plan's `directory` dropped, or kept as an unknown field, or kept in both places | `directory-shared`, `directory-values`, `directory-strings`, `directory-and-step` |
| An unknown key in the directory, a contact, a role, or a team accepted, dropped, or kept, or an `x-` key exempted | `directory-unknown-field`, `directory-problem-order` |
| The value of an unknown directory key read for more problems | `directory-unknown-field` |
| A team's role entry that is `[]` or null dropped, so the role is no longer vacant for that team | `directory-empty-lists`, `directory-strings` |
| A team's `holders: {}` read as nil, or an absent or null one read as an empty map | `directory-empty-lists` |
| A role's holders, a team's holders, or a directory list sorted, reversed, or stripped of repeats | `directory-list-order`, `directory-values`, `directory-shared`, `directory-strings` |
| Two items that share an id merged, or the second refused | `directory-list-order` |
| `slack` and `slackId` swapped, or `slackId` dropped | `directory-shared`, `directory-values`, `directory-strings` |
| A directory string trimmed, lowercased, stripped of `@`, or filled from another field | `directory-strings`, `directory-unvalidated`, `directory-values` |
| A role id in a team's `holders` lowercased, or read as a boolean or a number | `directory-strings` |
| A quoted `"42"`, `'true'`, `"yes"`, `'null'`, or a plain date or `deferred` in a directory field read as a number, a flag, null, or Deferred | `directory-strings` |
| A directory that is validated: a duplicate id, a bad slug, a bad email, a holder naming no contact, or a team entry for no role refused or cleaned | `directory-unvalidated`, `directory-list-order` |
| A step's `directory`, or a section's `directory`, read as the plan's directory or dropped | `directory-and-step` |
| A directory that is not a map, or a list that is not a list, accepted without a problem, or its inside scanned | `directory-scalar`, `directory-list`, `directory-wrong-type-lists` |
| An item that is not a map accepted without a problem, or crashing the decoder | `directory-wrong-type-items` |
| A null item of a directory list or of a list of holders dropped | `directory-null-items` |
| A directory field or a holder of the wrong kind, or in a spelling, accepted | `directory-field-problems` |
| An anchor, an alias, or a merge key accepted inside the directory, or not reported at a team's `holders` | `directory-anchors` |
| A repeated key accepted inside the directory, including a role id in a team's `holders` | `directory-duplicate-keys` |
| A key that is not a scalar skipped in a team's `holders` or in a contact | `directory-complex-key` |
| A team's `contact` dropped, or kept as an unknown field | `directory-team-contact` |
| A team's `contact` swapped with the `name` or the `id`, or read from `Contact`, `contacts`, `team-contact`, `x-contact`, or `contactId` | `directory-team-contact`, `directory-team-contact-elsewhere` |
| A team's `contact` added to its `holders`, or put in as a holder of a role, or a team's `holders` filled when it has none | `directory-team-contact` |
| A team's `contact` checked against the contacts, or filled, lowercased, trimmed, or cleared when it names none | `directory-team-contact` |
| A quoted `"42"`, `'true'`, `"yes"`, `'null'`, a plain date, or a plain `deferred` in a team's `contact` read as a number, a flag, null, a date, or Deferred | `directory-team-contact` |
| A null, empty, or bare `contact` read as the text `null` or `~`, or refused | `directory-team-contact` |
| A team's `contact` of the wrong kind, a plain `yes`, `Off`, `42`, `0x1F`, or `1e3`, accepted, or reported at the team and not at the `contact` | `directory-team-contact-problems` |
| An anchor or a repeated key on a team's `contact` not reported | `directory-team-contact-problems` |
| `contact` accepted or kept in a contact or in a role, or a value there read for more problems | `directory-team-contact-elsewhere` |
| The directory read in a second pass, so its problems come out of file order | `directory-problem-order` |
| A newer file reported with the problems of its directory | `directory-version-too-new` |
| `planUrl` or `resultsPath` ignored, or kept as an unknown field | `plan-links-values`, `plan-links-quoted`, `plan-links-plain-punctuation` |
| `planUrl` and `resultsPath` swapped | `plan-links-values`, `plan-links-quoted`, `plan-links-numbers-and-dates` |
| `planUrl` or `resultsPath` lowercased, trimmed, or changed | `plan-links-case-and-spaces`, `plan-links-words` |
| An empty or null link read as the text `null` | `plan-links-blank`, `plan-links-blank-null` |
| A link read as an unknown value, so a plain `yes`, `~`, `0x1F`, `12`, a flow map, or a list is accepted | `plan-links-spelling`, `plan-links-spelling-null-and-number`, `plan-links-wrong-type`, `plan-links-wrong-type-collections` |
| An anchor or a repeated key on a link not reported | `plan-links-anchor-and-duplicate` |
| An unquoted `{account}/...` template accepted | `plan-links-unquoted-brace` |
| `planURL`, `PlanUrl`, `ResultsPath`, or `x-planUrl` read as the link | `plan-links-values` |
| `kind: plan` refused, kept as an unknown field, or dropped only when plain | `kind-plan`, `kind-plan-after-body`, `kind-plan-quoted`, `kind-plan-double-quoted`, `kind-unknown-neighbours` |
| An absent, null, or empty `kind` refused, or kept | `kind-empty`, `plan-links-blank-null`, every case with no `kind` |
| An absent `kind` refused as missing | every case with no `kind`, such as `minimal` |
| Any kind accepted, or only `record` refused | `kind-other-word`, `kind-case`, `kind-deferred`, `kind-quoted-number`, `kind-spaces` |
| The kind compared without case, or trimmed | `kind-case`, `kind-spaces` |
| A wrong kind reported with another code, with no path, or with no `found` or `expected` in the detail | `kind-record`, and every `kind-` problem case |
| A message that does not name the kind found, or does not say a plan is expected | the message check in the problem test |
| A wrong kind reported as the only problem, so the others are hidden | `kind-with-other-problems`, `kind-anchor`, `kind-duplicate-plan-then-record` |
| The kind checked before the version, so a newer file gives a kind problem | `kind-version-too-new` |
| A repeated `kind` judged by the last or the first value only | `kind-duplicate-plan-then-record`, `kind-duplicate-record-then-plan` |
| A `kind` that is not text named as a kind, or skipped without a problem | `kind-spelling`, `kind-wrong-type`, `kind-plain-true`, `kind-anchor` |
| A kind problem ordered before the anchor at the same place | `kind-anchor`, `kind-duplicate-plan-then-record` |
| `kind` accepted inside `provenance`, `origin`, or the directory, or refused in a section | `kind-nested-unknown`, `plan-links-values` |
| `Kind`, `kinds`, or `x-kind` read as the kind | `kind-unknown-neighbours` |
| A record file judged for its other keys, or its shape | `kind-record-file` |
