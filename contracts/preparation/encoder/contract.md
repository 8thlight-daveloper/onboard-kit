# Preparation encoder contract

Component: `preparation/encoder`. Agent-authored contract; not human-reviewed.

## Behavior

`Encode(preparation) -> (bytes, problems)` is pure. It reads a preparation value,
including unknown fields, and changes nothing. It reads and writes no external
state. It returns canonical YAML when every value is representable. Any error
returns no bytes, never partial output.

The [preparation format](../../../docs/reference/preparation-format.md),
[0140](../../../docs/decisions/0140-person-specific-preparation.md),
[0160](../../../docs/decisions/0160-personal-data-and-preparation-approval.md), and
[0180](../../../docs/decisions/0180-first-onboarding-use.md) supply the shape and
ownership. The [plan encoder](../../plan/encoder/contract.md) supplies document,
scalar, multiline, list, and unknown-value rules. Those rules apply unchanged:
UTF-8, two spaces, LF, one final newline, no markers, comments, anchors, or tags;
reserved words and digit-leading strings quoted; qualifying multiline text in
literal blocks; sorted mapping keys; lists kept in their original order.

### Known fields

| Structure | Key order |
| -- | -- |
| Preparation | `schemaVersion`, `kind`, `account`, `planId`, `recipient`, `arrangements`, then unknown fields |
| Recipient | `name`, `companyEmail`, `plannedStartDate`, `team`, `manager`, `onboardingOwner`, `buddy`, `workLocation`, `timeZone` |
| Arrangement | `id`, `section`, `kind`, `owner`, `contacts`, `reference`, `targetDate`, `status`, `note` |

Always write schemaVersion as the held decimal integer, including zero and
unsupported versions. Always write kind as the held string, including empty or
non-preparation text. Kind is explicit in the preparation model; do not replace
it with a constant or invent a missing value. The decoder requires preparation;
the encoder preserves its input instead of silently repairing it.

Omit other empty strings and empty or absent lists. Omit recipient when every
field is empty. Known leaf fields except schemaVersion are strings or lists of
strings; Date follows the string rules. Keep nonempty dates, statuses, references,
and identities exactly as supplied, even when semantically invalid. The semantic
validator judges their meaning and references, not this component.

Arrangements remain in order, including repeated identities. Contacts remain in
order, including repeats and empty strings. An arrangement with no writable keys
is an error rather than omitted or written as an empty mapping. An arrangement
with a nonempty contacts list has a writable key even when each contact is empty.
The decoder can read an empty arrangement; that does not make it encodable.

### Unknown values and safety

Write unknown top-level keys after known fields, sorted by raw UTF-8 byte order.
Reject a collision with any of the six known root keys, even when that known
field is empty. Do not inspect a colliding value. Nested maps are map[string]any;
lists are []any; supported scalars are string, bool, int, int64, finite float64,
and nil. Empty maps and lists remain `{}` and `[]`. A whole float64 retains `.0`.
Do not coerce other numeric types, pointers, or arbitrary structs.

Reject invalid UTF-8 in every string and mapping key, rather than replacing bytes.
An invalid mapping key reports one unsupported-value problem at its path and
does not inspect that entry's value. Reject a cycle at the first map or list edge
that revisits a container on the current traversal branch. Shared acyclic values
in different branches are valid. Never panic or loop on unsupported input.

## Problems

| Code | When |
| -- | -- |
| `encode.empty-item` | An arrangement would write no keys |
| `encode.unknown-field-collision` | An unknown top-level key collides with a known key |
| `encode.unsupported-value` | Invalid UTF-8, unsupported unknown-value type, nonfinite number, or a cyclic edge |

Each problem has severity error, a nonblank useful message, and its exact path
in detail.path. Paths use dots and bracketed zero-based indices. Unknown root
fields use their key alone. Problems follow canonical traversal order: known
fields first, then sorted unknown keys, descending lists in input order and
maps in sorted order. Report independent failures, not only the first. After a
collision, unsupported value, invalid key, or cycle, do not descend into that
value. Diagnostics do not repeat supplied values in messages; paths identify the
location. An invalid raw key may appear in its path because it is the location.

## Fixtures and properties

[cases.json](testdata/cases.json) contains hand-specified input and expected
bytes or problem code/path pairs. Each case cites its source. The supplemental
tests cover Go values JSON cannot represent: integer kinds, unsupported types,
nonfinite numbers, invalid UTF-8, cycles, and shared acyclic containers.

Calls are deterministic, preserve all argument data, and share no retained
state. A success returns no problems. A failure returns no bytes. Golden bytes
are the independent oracle, not another encoder or the decoder's output.

The [API](api.go), [tests](encoder_test.go), and [binding](wire_test.go) complete
the four-part contract. The binding is a stub until a separate implementation
increment. File I/O, semantic validity, invitation readiness, privacy scanning,
and first-use recording are outside this component.

## Wrong implementations tried

| Wrong implementation | Caught by |
| -- | -- |
| Return no bytes and no problems | Every fixture |
| Return one generic problem for every input | Successful golden fixtures and exact paths |
| Write kind as a fixed preparation constant | zero, not-semantic-validator, quoting |
| Drop recipient information | recipient-all, not-semantic-validator, quoting, multiline |
| Omit the final newline | Every successful fixture |
| Return no problems for invalid input | empty-items, all-collisions |
| Stop after the first problem | empty-items, all-collisions |
