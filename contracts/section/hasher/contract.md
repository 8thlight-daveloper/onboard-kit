# Section hasher contract

Status: approved. Component: `section/hasher`. Issue: [#62](https://github.com/8thlight-daveloper/onboard-kit/issues/62).

The section hasher turns a Section into a short text, its content hash. It is a pure function. It reads no file, reads no clock, keeps no state between calls, and does not change its argument. See [0030](../../../docs/decisions/0030-self-contained-plan-file.md), [0050](../../../docs/decisions/0050-section-switch.md), the [domain model](../../../docs/reference/domain-model.md), and [0015](../../../docs/decisions/0015-functional-core-and-imperative-shell.md).

## Purpose

The plan instantiator records the hash in a section's `source` when it copies a template into a plan. The deferred update action compares hashes to tell a section edited by hand from one that changed upstream.

The principle: **the hash changes only when the section's behavior changes.** A change that has no effect must not change it, because a false alarm tells a consultant a section was edited when nothing happened.

The hash covers the section's definition:

- Its text: title, summary, owner, verification, and body.
- Its questions: the variables it declares, with their descriptions, examples, types, options, patterns, and defaults.
- Its steps and checks.
- Its conditions, `includeIf` and each `askIf`, with their question clauses and their section clauses.
- Its `parent`, `bootstrap`, `timing`, `after`, `requires`, `escalation`, and unknown fields.

It does not cover answers, because answers live in the plan and not in the section. It does not cover `source`, and it does not cover the section's switch `enabled`, which is the plan's own setting. It detects an edit to a section and a change to its template. It does not detect that someone answered a question.

0030 sets the base: the hash is of the section's structured content in a fixed order, not of raw text, so reformatting does not change it. It excludes `source`, and 0050 adds the switch. Variables are unresolved. The canonical form is a contract protected by golden fixtures, and it is not a git blob sha. The domain model says the hash is "computed over everything except `source`". This contract also leaves out `enabled` as 0050 decides, and adds the rules below that make "no effect" precise.

## Interface

`Hash(section) -> text`.

- The result is always 64 characters, each one of `0` to `9` or `a` to `f`.
- There are no problems to report. The hasher writes a hash for any Section it is given. See Not covered for the values it cannot write.

## Algorithm

The hash is SHA-256 of the canonical form, written as lowercase hexadecimal, with no prefix and no version marker. A later change to the canonical form changes every recorded hash, and that is accepted.

## Canonical form

The canonical form is UTF-8 text in JSON syntax, with no whitespace outside strings. Its job is to be one text for one behavior, and never the same text for two behaviors. It is separate from the plan file's YAML on purpose, so a change in how the encoder writes a plan cannot change a hash. It uses the encoder's key order, as an independent choice: a change to the encoder's order never changes a hash.

A Section is written as an object. Its keys come in this order, and `source` and `enabled` are not among them.

| Type | Key order |
| -- | -- |
| Section | `id`, `title`, `summary`, `includeIf`, `parent`, `bootstrap`, `timing`, `after`, `requires`, `owner`, `escalation`, `verification`, `body`, `variables`, `steps`, `unknownFields` |
| Variable | `name`, `description`, `example`, `type`, `multiple`, `options`, `pattern`, `patternHint`, `required`, `default`, `audience`, `askIf` |
| Condition | `all` |
| Clause | `question`, `answer`, `section`, `enabled` |
| AnswerTest | `is`, `isNot`, `isOneOf`, `isNotOneOf`, `isSettled` |
| Step | `id`, `name`, `notes`, `owner`, `timing`, `escalation`, `run`, `open`, `directory`, `env`, `checks` |
| Check | `run`, `ask`, `env` |

### Fields

- A key is written as a string, then `:`, then its value. Keys and items are separated by `,`. An object is `{` and `}`, and a list is `[` and `]`.
- A key with a zero value is omitted. Zero means an empty string, `false`, a list with no items, a map with no entries, a Value of kind none, a nil pointer, a Condition with no clauses, or an AnswerTest with every field zero. A zero field has no effect, so it is the same as an absent one, whether a list is nil or empty.
- An `isSettled` that is set is written, including `false`, because `false` is a test. A clause's `enabled` is written by the rule under Conditions.
- An item of a list is always written. A clause, step, or check with every field zero is `{}`, and a variable with every field zero is `{"type":"string","audience": "onboarder"}`, because of the defaults. Dropping an item would shorten the list.
- Nothing is validated. A check with two of `run` and `ask`, a `timing` that is not a timing, and a clause with two tests are all written as held.
- A map is written as an object with its keys in ascending byte order of their UTF-8 text. This covers a step's `env` and the keys inside `unknownFields`. An entry is written even when its value is empty.

### Defaults

The hasher fills three defaults before it writes, so an absent field and its default hash alike. The values are the domain model's.

| Field | Default |
| -- | -- |
| A section's `timing` | `anytime` |
| A variable's `type` | `string` |
| A variable's `audience` | `onboarder` |

Nothing else is filled. A step with no `timing` or `escalation` inherits them from its section, and the hasher leaves them as held. A step with `timing: anytime` therefore differs from a step with none, because the first does not follow its section. A `required` or `multiple` that is `false` is zero and omitted, so it already hashes like an absent one. A value that is not the default differs.

### Prose

Prose is text that people read. Whitespace noise in it has no effect, so nine fields are normalized before they are written: a section's `title`, `summary`, `body`, and `verification`, a step's `notes`, a variable's `description`, `example`, and `patternHint`, and a check's `ask`. An `example` is one illustration of a good answer, shown to the person answering. It is not an answer or a default, and it is written and normalized like `description`: omitted when it is empty after the rule, so a variable with no example and one with an empty or whitespace-only example hash alike. The rule, in this order:

1. Replace each CR LF, and then each lone CR, with LF.
2. Remove spaces and tabs at the end of every line.
3. Remove all leading and trailing whitespace of the whole string.

Whitespace in step 3 means exactly three characters: space (U+0020), tab (U+0009), and LF (U+000A). A no-break space, a form feed, and every other character that looks like space are not whitespace here.

Nothing else changes. Whitespace inside a line, the indentation of a line, and the blank lines between paragraphs are content, because they can change how Markdown reads. A normalized string with no characters left is empty and is omitted.

Every other string is written exactly as held: a section's `parent`, a `pattern`, a `run` item, an `env` value, a `directory`, an `open`, an option, an `id`, a `name`, an `owner`, a `timing`, a default or test value, and every string inside an unknown field. A trailing space in a pattern or a run item can change what it does.

### Strings

A string is written in double quotes. Inside the quotes, these characters are written as an escape.

| Character | Written as |
| -- | -- |
| `"` | `\"` |
| `\` | `\\` |
| Newline | `\n` |
| Carriage return | `\r` |
| Tab | `\t` |
| Any other character below U+0020 | `\u00` and two lowercase hex digits, such as `\u0001` and `\u001f` |

Every other character is written as it is, including `<`, `>`, `&`, U+007F, U+2028, and every character outside ASCII. Keys are written as strings.

### Lists and sets

Four lists are sets: a section's `after` and `requires`, and `isOneOf` and `isNotOneOf`. Their order and repeats have no effect. After holds legacy reading-order hints, Requires holds explicit verification prerequisites under [0260](../../../docs/decisions/0260-prerequisites-and-action-order.md), and the two tests check membership. A set is written sorted in ascending byte order, with repeats removed. The items of `after` and `requires` are sorted by the bytes of the exact id. The items of the two tests are sorted by the bytes of their canonical text, as the Values section gives it, so the Text `"a"`, the list `["x"]`, `null`, `true`, and `{"deferred":true}` come in that order.

Every other list is written in the order held, and its repeats are kept, because its order is behavior: `steps`, `variables`, `checks`, `run`, `escalation` (an ordered list of who to ask), `options`, a Variable's list default, a Clause list in `all`, and every list inside `unknownFields`. The hasher never sorts or reverses these, and never removes an item.

### Values

A Value is written by its kind, and the other fields of the Value are ignored.

| Kind | Written as |
| -- | -- |
| Text | A string. The Text `deferred` is an ordinary string, and differs from the Deferred kind |
| Flag | `true` or `false` |
| List | A list of strings in the order held, `[]` when empty |
| Deferred | `{"deferred":true}` |
| None | The key is omitted. As an item of `isOneOf` or `isNotOneOf`, it is `null` |

A Value of kind Text with no text is `""`, so a `default` of the empty string differs from no `default`. A Flag that is `false` is written, so `default: false` differs from no `default`. These are values, not prose, and they are never normalized.

### Conditions

A Condition with clauses is an object with `all`, a list of clauses. A Clause has two forms, a question clause and a section clause (0050), and an AnswerTest is an object with the keys in the table, each written by the rules above. A Condition with no clauses is zero, so an empty `includeIf` or `askIf` hashes like an absent one.

- A question clause is an object with `question` and `answer`, as before.
- A section clause is an object with `section`, the id as held, and `enabled`. Its canonical form is `{"section":"id","enabled":true}`, and with a switch that is off, `{"section":"id","enabled":false}`. `enabled` is a test, as `isSettled` is, so `false` is written and not omitted. It is also written when it is `true` on a clause with no `section`.
- A clause is never both forms in a plan, and the hasher does not check it, as it checks nothing else. A clause is written with its keys in the order of the table: `question` and `answer` by the rules for a zero field, `section` when it is not empty, and `enabled` when the clause has a `section` or when `enabled` is `true`. So a clause that holds both kinds of key is written with all of them.
- The id in `section` is written exactly as held. A section clause differs from a question clause with the same text.

### Unknown fields

The unknown fields of a section are written as one object under the key `unknownFields`, after every other key. They are not written beside the known keys, so an unknown field named `title` cannot be mistaken for the title. `unknownFields` is omitted when the map has no entries. Adding an unknown field is a behavior change for the tool that knows it, so it changes the hash. Nothing inside an unknown field is normalized, filled, sorted, or removed.

| Value | Written as |
| -- | -- |
| null | `null` |
| String | A string, exactly as held |
| Boolean | `true` or `false` |
| Integer (`int`, `int64`) | Plain decimal digits, with `-` when negative, such as `5` and `-3` |
| Decimal (`float64`) | Plain notation, never an exponent, with the fewest digits that read back as the same number, and `.0` added when the text has no `.`, such as `2.0`, `1.5`, `0.00001`, and `100000000000000000000.0` |
| List | A list of its items in the order held, `[]` when empty |
| Map | An object with keys in ascending byte order, `{}` when empty |

Inside an unknown field, a key with an empty list or an empty map differs from an absent key. The hasher cannot know what it means to a newer tool, so it does not call it a zero. A decimal `2.0` differs from the integer `2` and from the string `"2"`.

### Bootstrap

Under [0250](../../../docs/decisions/0250-bootstrap-content-and-projection.md),
Bootstrap is section content, not an answer or a plan-owned switch. Write
`"bootstrap":true` after parent and before timing when true; omit it when false.
This preserves all hashes for sections without a true marker. Do not infer it
from timing, source, conditions, or another section. Its value is independent
of Enabled, which stays excluded. The hand-written canonical strings and
independent SHA-256 constants in [bootstrap.json](testdata/bootstrap.json) pin
both marked hashes and old unmarked hashes. Changing Bootstrap must also pass
the existing every-field property without any exception.

### Requires

Under [0260](../../../docs/decisions/0260-prerequisites-and-action-order.md),
Requires is section content. Write its sorted, deduplicated set after After
and before Owner. Omit nil and empty Requires, preserving all old hashes.
Do not combine it with After, infer it from Parent or Bootstrap, or move it
into unknownFields. Exact target changes change the hash; case, whitespace,
and empty IDs are not normalized or validated here. Self-references and unknown
targets are hashed as held, without consulting a plan. Each list is sorted on
a copy, including when After and Requires share backing storage.

The hand-written canonical strings and independent SHA-256 constants in
[requires.json](testdata/requires.json) pin these rules, their position relative
to Bootstrap, and the distinction from legacy After and unknown fields.
The existing every-field property applies without a Requires exception.

### Parent

`parent` is the id of another section of the same plan, 0070. It is written like any text identifier: exactly as held, never normalized, trimmed, or case-folded, because it is an identifier and not prose. A parent with a trailing space, or in another case, differs. It is written right after `includeIf` and before `bootstrap` and `timing`, and it is omitted when it is empty, so an empty `parent` hashes like an absent one and every hash written before `parent` existed stays the same. A section that names itself as its parent is written as held. The hasher does not check that the parent exists, and it does not read the parent section.

### Source and switch

`source` is not written. Two sections that differ only in `source`, or where one has a `source` and the other has none, have the same hash. The same holds for the contents of the `source`: its `template` and its `hash`.

The switch `enabled` of a section is not written either. Two sections that differ only in `enabled`, or where one has `enabled` true or false and the other has none, have the same hash. A section clause inside a condition is a different thing. It is part of the condition, and it is written.

## Properties

These hold for every section the contract tests use.

- **Deterministic.** Any number of calls with the same Section give the same hash, on one hasher or on a new one.
- **Stateless.** A hasher that has hashed other sections gives the same hash as a new one.
- **Inputs unchanged.** The section, its lists, and its maps are the same after the call as before. Sorting a set happens on a copy.
- **Shape.** The result is 64 lowercase hex digits.
- **Behavior only.** Two sections hash alike exactly when their canonical forms are the same text. A change to any field of any type except `source` and a section's `enabled` changes the hash, except a change that the rules above call no effect: a default written out, a zero field, noise at the edge of a prose field, and the order or repeats of a set.

## Side effects and plan access

None, and it reads the section content only. It reads no plan field, no answer, and no `source`. This matches the row for `section/hasher` in [components](../../../docs/architecture/components.md).

## Not covered

- Comparing a hash with a recorded one, and deciding what a difference means. That is the update action.
- Writing the hash into `source`. That is the plan instantiator.
- The values of an unknown field that are not in the table under Unknown fields, such as NaN, infinity, other integer widths, `float32`, and a value that is not data. The result for them is not specified.
- A Value whose kind is not one of the five kinds, such as a negative number. The result for it is not specified.
- A string that is not valid UTF-8, and Unicode normal forms. Two strings that look alike but are not the same code points hash differently.
- Whether a section is valid, including whether its `parent` names a section of the plan or names an ancestor and not a cycle. The hasher writes any section. The plan validator's structure check does that, in a later contract.
- Whether two different texts mean the same to Markdown. Only the whitespace rule above is applied.

## Decided by a person

These are not open. A rebuild keeps them.

1. The principle: the hash follows behavior, not text.
2. SHA-256 as 64 lowercase hex digits, with no prefix and no version marker in the hash.
3. An empty list or map hashes like an absent one for every known field. Inside an unknown field it does not.
4. Unknown fields on a section are in the hash, as held.
5. The key order is the encoder's, as an independent choice.

Decided from the docs, as rules above: the three defaults, the nine prose fields and their rule, an empty condition as absent, and the four sets.

## Wrong implementations tried

Each row names a plausible wrong implementation and the groups of cases that fail it. How the contract was checked, and with what scores, is recorded in the issue, not here.

| Wrong implementation | Caught by |
| -- | -- |
| Drop Requires, retain repeats or held order, move its key, merge it into After, or normalize its IDs | [Requires hash fixtures](testdata/requires.json) |
| Sort Requires by escaped JSON text rather than raw ID bytes | `raw-id-bytes-not-json-escape-order` in [Requires hash fixtures](testdata/requires.json) |
| Write empty Requires, mutate its backing storage, or cache a Requires hash by section ID | Requires omission fixtures and the Requires input-isolation and repeat-call tests |
| Drop Bootstrap, emit false, or move true away from its canonical key position | [bootstrap hash fixtures](testdata/bootstrap.json) and the every-field property |
| `parent` left out of the hash | `parent`, `parent-differs`, the parent test, `every-field` |
| `parent` written in another place, such as after `timing` or in the order of the key sort, or under another key | `parent`, the parent test |
| `parent` trimmed, normalized as prose, or case-folded | `parent-differs`, the parent test, `every-field` |
| An empty `parent` written as `"parent":""`, or told apart from an absent one | `parent-empty`, the empty-parent test |
| `parent` written only when it differs from the section's own `id`, or only when it exists in the plan | `parent-differs`, the parent test |
| `parent` written into `unknownFields`, or an unknown field named `parent` taken for it | `parent-differs` |
| `source` included, or only its template | `source-ignored`, `full` |
| A section's `enabled` written, as true, as false, or only when false | `enabled-ignored`, `switch-enabled-ignored`, `switch`, `full`, the switch test |
| A section's `enabled` read as a zero so that an absent one hashes unlike a true one | `enabled-ignored`, the switch test |
| A section clause dropped, or written as an empty clause `{}` | `switch`, `section-clauses`, `full`, the section clause test |
| A section clause written as a question clause, or `section` written under the key `question` | `switch`, `section-clauses`, the section clause test |
| A section clause written with `enabled` only when true | `switch`, `section-clauses`, the section clause test |
| A section clause written with `enabled` only when false, or with it first | `switch`, `section-clauses`, the section clause test |
| `section` and `enabled` swapped in a clause, or written in the order of the key sort | `switch`, `full`, the section clause test |
| A section clause written with an empty answer, `"answer":{}` | `switch`, `full`, the section clause test |
| A section clause that names one id the same as another, or its `enabled` ignored | `section-clauses`, `switch`, `full` |
| A section clause moved between `includeIf` and `askIf`, or its place in the list lost | `section-clauses` |
| `enabled` of a clause ignored when `section` is empty, or a clause with both kinds of key written as one kind | The held-clause test, `every-field` |
| Unknown fields dropped, inlined, normalized, or an empty list in them dropped | `unknown-fields`, `unknown-shapes`, `field-boundaries`, `exact-strings`, `every-field` |
| Unknown map keys in iteration order, in case-insensitive order, or not escaped | `unknown-fields`, `unknown-key-order`, the repeat-call test |
| A list in an unknown field, a list default, or `options` sorted | `list-order`, `unknown-fields` |
| `escalation`, a step's `escalation`, `run`, `steps`, or `variables` sorted | `list-order`, `full` |
| `after`, `isOneOf`, or `isNotOneOf` kept in the order held, or sorted with repeats kept | `sets`, `sets-order`, `full`, `values` |
| A set sorted by its text without quotes | `sets`, `sets-order`, `values` |
| `after` sorted in place, or any other change to the argument | The inputs-unchanged test |
| Timing, type, or audience default not filled | `defaults-explicit`, `omitted-zero`, `basic`, the model-defaults test |
| A default filled on a step, or the section's `escalation` copied to a step | `defaults-differ`, `defaults-explicit`, `basic` |
| CR LF or a lone CR not normalized | `prose-whitespace` |
| Trailing spaces or tabs kept, the whole-string trim missing, only the trailing newline trimmed, or leading blank lines kept | `prose-whitespace`, `omitted-zero` |
| Inner whitespace, indentation, or blank lines collapsed, or a no-break space or form feed trimmed | `prose-content`, `prose`, `strings` |
| One of the nine prose fields not normalized, or `example` not normalized | `prose-whitespace`, `example-whitespace` |
| `example` left out of the hash, or written when empty or when only whitespace | `example`, `example-differs`, `example-empty`, `every-field` |
| `example` written in another place than right after `description`, or under another key, or on another variable | `example`, `example-places` |
| `example` written only when `description` is set, or in place of `description` | `example`, `example-places` |
| `example` trimmed beyond the prose rule, or its inner spacing, case, or line break collapsed | `example-differs`, `example`, `example-whitespace` |
| A whitespace-only prose field written as `""` | `omitted-zero` |
| A `run` item, `pattern`, option, `directory`, `owner`, name, default value, or id normalized or trimmed | `exact-strings`, `every-field`, `strings` |
| An empty `includeIf` or `askIf` told apart from an absent one | `omitted-zero`, the empty-conditions test |
| Empty lists and maps written when they are not nil, or `required: false` written | The empty-and-absent test, `omitted-zero`, `basic` |
| One field left out, such as a step's `directory`, a check's `env`, or `isNotOneOf` | `every-field`, `full` |
| An `env` entry with an empty value dropped, or `env` keys unsorted | `full`, the repeat-call test |
| Two fields run together, so `a` and `bc` equal `ab` and `c` | `field-boundaries`, `every-field` |
| A JSON library's escaping of `<`, `>`, `&`, U+2028, `\b`, `\f`, or non-ASCII | `strings`, `unknown-fields` |
| A control character in upper-case hex, DEL escaped, or a newline written as `\u000a` | `strings`, `full` |
| `false` or an empty text `default` treated as absent | `value-kinds`, `values` |
| The Text `deferred` written as the Deferred kind | `value-kinds`, `values`, `full` |
| A Value written by its non-zero fields and not by its kind | The by-kind test |
| A decimal `2.0` written `2`, an integer as a string, or an exponent written | `unknown-shapes`, `unknown-fields`, the number test |
| `int` not supported, or `int` and `int64` hashed differently | The number test |
| A zero step, variable, check, clause, or list item dropped, or a none item of `isOneOf` dropped | `every-field`, `field-boundaries`, `sets`, `values` |
| A zero answer written as `{}`, or `isSettled: false` dropped | `values`, `full` |
| Keys in alphabetical order, or two keys of a clause or a variable swapped | `full`, `values`, `basic` |
| SHA-1, a truncated hash, or upper-case hex | The shape test, every golden case |
| Whitespace, indentation, or a final newline in the canonical form | Every golden case |
| State kept between calls, such as a cache by `id` | Every case that has two sections with one `id`, the repeat-call test |
| The time written into the hash | The repeat-call test, every golden case |
