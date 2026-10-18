# Section hasher contract

Status: draft, for a person to approve. Component: `section/hasher`. Issue: [#62](https://github.com/8thlight-daveloper/onboard-kit/issues/62).

The section hasher turns a Section into a short text, its content hash. The plan instantiator records the hash in the section's `source` when it copies a template into a plan, and the deferred update action compares hashes to tell an edited section from an untouched one. The hasher is a pure function. It reads no file, reads no clock, keeps no state between calls, and does not change its argument. See [0030](../../../docs/decisions/0030-self-contained-plan-file.md), the [domain model](../../../docs/reference/domain-model.md), and [0015](../../../docs/decisions/0015-functional-core-and-imperative-shell.md).

0030 decides four things, and this contract keeps them:

- The hash is of the section's structured content in a fixed order, not of raw text, so reformatting does not change it.
- It excludes `source`.
- Variables are unresolved, so answers do not change it.
- The canonical form is a contract protected by golden fixtures. It is not a git blob sha.

The domain model adds that the hash is "computed over everything except `source`".

## Interface

`Hash(section) -> text`.

- The result is always 64 characters, each one of `0` to `9` or `a` to `f`.
- There are no problems to report. The hasher writes a hash for any Section it is given. See Not covered for the values it cannot write.

## Algorithm

The hash is SHA-256 of the canonical form, written as lowercase hexadecimal. A person must approve this choice. See Open questions.

- SHA-256 is in the standard library of every language we might rebuild in, so a rebuild in another language writes the same hash.
- It has one standard text form, 64 lowercase hex digits, which fits the plan file's `hash` field with no prefix or quoting.
- The hash detects change. It does not defend against an attacker. A shorter hash would also detect change, but it leaves less room if plans or template sets grow, and the length saves nothing a person reads.

## Canonical form

The canonical form is UTF-8 text in JSON syntax, with no whitespace outside strings. Its job is to be one text for one content, and never the same text for two contents. It is separate from the plan file's YAML on purpose, so a change in how the encoder writes a plan cannot change a hash. It reuses the encoder's key order only so that the two read alike.

A Section is written as an object. Its keys come in this order, and `source` is not one of them.

| Type | Key order |
| -- | -- |
| Section | `id`, `title`, `summary`, `includeIf`, `timing`, `after`, `owner`, `escalation`, `verification`, `body`, `variables`, `steps`, `unknownFields` |
| Variable | `name`, `description`, `type`, `multiple`, `options`, `pattern`, `patternHint`, `required`, `default`, `phase`, `askIf` |
| Condition | `all` |
| Clause | `question`, `answer` |
| AnswerTest | `is`, `isNot`, `isOneOf`, `isNotOneOf`, `isAnswered` |
| Step | `id`, `name`, `notes`, `owner`, `timing`, `escalation`, `run`, `open`, `directory`, `env`, `checks` |
| Check | `run`, `ask`, `env` |

### Fields

- A key is written as a string, then `:`, then its value. Keys and items are separated by `,`. An object is `{` and `}`, and a list is `[` and `]`.
- A key with a zero value is omitted. Zero means an empty string, `false`, a list with no items, a map with no entries, a Value of kind none, a nil pointer, or an AnswerTest with every field zero.
- Two kinds of field are written even when they are empty. A condition that is present is written, as `{}` when it has no clauses, so a present empty `includeIf` differs from an absent one. A `isAnswered` that is set is written, including `false`.
- A list with no items is the same as an absent list, whether it is nil or empty. The Go model can hold both, and they hash alike.
- An item of a list is always written, as `{}` when every field of a variable, clause, step, or check is zero. Dropping it would shorten the list.
- A value is written as it is held. The hasher fills no default, so a section with no `timing` and a section with `timing: anytime` hash differently, and so do a variable with no `type` and one with `type: string`.
- Nothing is validated. A check with two of `run` and `ask`, a `timing` that is not a timing, and a clause with two tests are all written as held.
- A list is written in the order the section holds it, and repeats are kept. The order of steps, variables, and every other list is content. The hasher never sorts a list, reverses it, or removes an item.
- A map is written as an object with its keys in ascending byte order of their UTF-8 text. This covers a step's `env` and the keys inside `unknownFields`. An entry is written even when its value is empty.

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

Every other character is written as it is, including `<`, `>`, `&`, U+007F, U+2028, and every character outside ASCII. The string is written exactly as held. It is not trimmed, folded, wrapped, or normalized, so whitespace inside a string is content, and a body with two spaces differs from one with one. "Reformatting" in 0030 means how the plan file lays out the same content, such as quoting, block style, and key order.

### Values

A Value is written by its kind, and the other fields of the Value are ignored.

| Kind | Written as |
| -- | -- |
| Text | A string. The Text `unknown-yet` is an ordinary string |
| Flag | `true` or `false` |
| List | A list of strings, `[]` when empty |
| Unknown | `{"unknown":true}` |
| None | The key is omitted. As an item of `isOneOf` or `isNotOneOf`, it is `null` |

A Value of kind Text with no text is `""`, so a `default` of the empty string differs from no `default`. A Flag that is `false` is written, so `default: false` differs from no `default`.

### Conditions

A Condition is an object with `all`, a list of clauses. A Clause is an object with `question` and `answer`. An AnswerTest is an object with the keys in the table, each written by the rules above.

### Unknown fields

The unknown fields of a section are written as one object, under the key `unknownFields`, after every other key. They are not written beside the known keys, so an unknown field named `title` cannot be mistaken for the title. `unknownFields` is omitted when the map has no entries. Unknown fields are kept in a plan, so they are part of the content, and 0030 says plan files may carry fields from a newer tool.

Inside it, a value keeps its kind.

| Value | Written as |
| -- | -- |
| null | `null` |
| String | A string |
| Boolean | `true` or `false` |
| Integer (`int`, `int64`) | Plain decimal digits, with `-` when negative, such as `5` and `-3` |
| Decimal (`float64`) | Plain notation, never an exponent, with the fewest digits that read back as the same number, and `.0` added when the text has no `.`, such as `2.0`, `1.5`, `0.00001`, and `100000000000000000000.0` |
| List | A list of its items in order, `[]` when empty |
| Map | An object with keys in ascending byte order, `{}` when empty |

A key present with an empty list or an empty map is written. It differs from an absent key. A decimal `2.0` differs from the integer `2` and from the string `"2"`.

### Source

`source` is not written. Two sections that differ only in `source`, or where one has a `source` and the other has none, have the same hash. The same holds for the contents of the `source`: its `template` and its `hash`.

## Properties

These hold for every section the contract tests use.

- **Deterministic.** Any number of calls with the same Section give the same hash, on one hasher or on a new one.
- **Stateless.** A hasher that has hashed other sections gives the same hash as a new one.
- **Inputs unchanged.** The section, its lists, and its maps are the same after the call as before.
- **Shape.** The result is 64 lowercase hex digits.
- **Content only.** Two sections hash alike exactly when their canonical forms are the same text. A change to any field of any type except `source` changes the hash.

## Side effects and plan access

None, and it reads the section content only. It reads no plan field, no answer, and no `source`. This matches the row for `section/hasher` in [components](../../../docs/architecture/components.md).

## Not covered

- Comparing a hash with a recorded one, and deciding what a difference means. That is the update action.
- Writing the hash into `source`. That is the plan instantiator.
- The values of an unknown field that are not in the table under Unknown fields, such as NaN, infinity, other integer widths, `float32`, and a value that is not data. The result for them is not specified.
- A string that is not valid UTF-8, and Unicode normal forms. Two strings that look alike but are not the same code points hash differently.
- Whether a section is valid. The hasher writes any section.
- A version marker in the canonical form. See Open questions.

## Open questions

A person decides each of these. The contract takes the first reading.

1. **Algorithm and text form.** 0030 says only "a hash". Taken: SHA-256 as 64 lowercase hex digits. Another reading: a prefix such as `sha256:`, which lets a later algorithm be told apart from this one, or a hash cut to 16 or 20 digits, which is easier to read in a diff.
2. **Defaults.** The domain model says `timing` defaults to `anytime`, `type` to `string`, and `phase` to `planning`. Taken: the hasher fills no default, as the decoder does not, so an absent value and its default hash differently. Another reading: fill them, so a template that adds `timing: anytime` is not seen as changed. That needs the hasher to know the defaults, which then belong to its contract.
3. **Whitespace inside strings.** 0030 says reformatting does not change the hash. Taken: that is about the file's layout, and a space, a tab, or a line ending inside a string is content. Another reading: fold or trim whitespace, or turn CRLF into LF, so an editor that changes line endings does not change a hash. That would make a real change to a body invisible.
4. **A present empty condition.** `includeIf: {}` is a Condition with no clauses, which holds for every plan, as an absent one does. Taken: they hash differently, since the model tells them apart. Another reading: treat them as one.
5. **Empty and absent lists.** Taken: the same, for every list and map of the model, and different for a key in an unknown field. Another reading: tell them apart. The shared test code that reads fixtures turns `[]` into a nil list, and a decoder may do the same, so the difference would depend on how a section was built.
6. **Unknown fields.** 0030 does not say whether they count. The domain model says the hash covers everything except `source`, so they are included. The effect: a plan opened and saved by a newer tool that adds a field to a section makes that section look edited to an older tool.
7. **List order.** 0030 says "a fixed order". Taken: that is the order of keys, and the order of every list is content. Another reading: sort the lists whose order has no meaning, such as `after` and `escalation`. The contract does not know which those are.
8. **A version marker.** Any change to the canonical form changes every hash recorded in every saved plan, so it needs an ADR and a plan for old plans. Taken: no marker. Another reading: start the canonical form with a version, such as `section-hash-1`, so a later form can be told from this one and an update can say that the two hashes are not comparable.
9. **Key order.** Taken: the order of the encoder's tables, as an independent choice. A change to the encoder's order never changes a hash. A person may prefer sorted keys.

## Wrong implementations tried

While drafting, 64 wrong hashers were written as scratch work outside the repo, each with one defect, and the contract tests failed every one. A reference hasher written from this page alone passed every test on its first run, which showed that the hand-written canonical texts agree with the rules. Run against a stub that returns an empty string, every test that exercises the output fails. Only the tests of manifest shape, golden consistency, and determinism pass. Two survivors led to fixtures: an `env` entry with an empty value, and a decimal that a library writes with an exponent. A third, a hasher that trims one string field, led to the trailing-space changes in `every-field`. No second independent build or differential run was made, because the issue is a draft.

| Wrong implementation | Caught by |
| -- | -- |
| `source` included, or only its template | `source-ignored`, `full` |
| Unknown fields dropped, or written beside the known keys | `unknown-fields`, `unknown-key-order`, `field-boundaries`, `every-field` |
| Unknown map keys in iteration order, in case-insensitive order, or not escaped | `unknown-fields`, `unknown-key-order`, the repeat-call test |
| A list sorted, reversed in place, or stripped of repeats | `list-order`, `full` |
| The section changed during the call, a list sorted in place or an unknown map written to | The inputs-unchanged test, `unknown-fields` |
| Empty lists and maps written when they are not nil | The empty-and-absent test, `omitted-zero` |
| Empty lists, maps, or null dropped inside unknown fields | `unknown-shapes`, `unknown-fields` |
| Body trimmed, folded, or CRLF turned into LF | `body-whitespace`, `strings`, `every-field` |
| One field left out, such as `verification`, a step's `directory`, a check's `env`, a variable's `patternHint`, or `isNotOneOf` | `every-field`, `full` |
| A string field trimmed, such as an `env` value | `every-field` |
| An `env` entry with an empty value dropped, or `env` keys unsorted | `full`, the repeat-call test |
| Two fields run together, so `a` and `bc` equal `ab` and `c` | `field-boundaries`, `every-field` |
| A JSON library's escaping of `<`, `>`, `&`, U+2028, `\b`, `\f`, or non-ASCII | `strings`, `unknown-fields` |
| A control character in upper-case hex, DEL escaped, or a newline written as `\u000a` | `strings`, `full` |
| `false` or an empty text `default` treated as absent | `value-kinds`, `values` |
| The Text `unknown-yet` written as the Unknown kind | `value-kinds`, `values`, `full` |
| A Value written by its non-zero fields and not by its kind | The by-kind test |
| A decimal `2.0` written `2`, an integer as a string, or an exponent written | `unknown-shapes`, `unknown-fields`, the number test |
| `int` not supported, or `int` and `int64` hashed differently | The number test |
| A present empty `includeIf` or `askIf` the same as an absent one, or an absent one written as `null` | `absent-and-default`, `basic` |
| A default filled for `timing` or `type` | `absent-and-default`, `basic`, `values` |
| A zero step, variable, check, clause, or list item dropped | `every-field`, `field-boundaries`, `value-kinds` |
| A zero answer written as `{}`, a none item of `isOneOf` dropped, or `isAnswered: false` dropped | `values`, `full` |
| Keys in alphabetical order, or two keys of a clause or a variable swapped | `full`, `values`, `basic` |
| SHA-1, a truncated hash, or upper-case hex | The shape test, every golden case |
| Whitespace, indentation, or a final newline in the canonical form | Every golden case |
| State kept between calls, such as a cache by `id` | Every case that has two sections with one `id`, the repeat-call test |
| The time written into the hash | The repeat-call test, every golden case |
