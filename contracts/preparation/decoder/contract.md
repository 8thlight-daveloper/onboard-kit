# Preparation decoder contract

Component: `preparation/decoder`.

The preparation decoder turns YAML bytes into `model.Preparation`. It reads
only its byte argument and returns a new value and problems. It performs no
input or output, reads no plan or record, and changes no argument. See the
[preparation format](../../../docs/reference/preparation-format.md),
[0140](../../../docs/decisions/0140-person-specific-preparation.md),
[0160](../../../docs/decisions/0160-personal-data-and-preparation-approval.md),
and [0015](../../../docs/decisions/0015-functional-core-and-imperative-shell.md).

## Interface

`Decode(bytes) -> (preparation, problems)`.

Success returns the decoded preparation and no problems. Any problem returns
the zero Preparation, never a partially filled one. Every problem has severity
`error`, a nonblank useful message, and its location in `detail.path` except
for file problems, which have no path. Messages and details must not repeat
supplied scalar values or YAML parser excerpts. Paths identify keys and list
positions; this is not a promise to conceal personal data put into a key.
Diagnostics may name expected fixed fields and the required `kind: preparation`.

## YAML dialect and precedence

Use the [plan decoder](../../plan/decoder/contract.md)'s Reading the file,
Version, Problems, format-guard, Unknown fields, and Properties rules with the
following explicit preparation-specific differences. Read that contract as
part of this one. No plan-only field or answer behavior is introduced here.

1. File checks precede everything else: invalid UTF-8 or syntax anywhere is
   `file.not-yaml`; an empty first document or non-map root is
   `file.not-a-preparation`; a second document is `file.multiple-documents`.
   Keep the inherited ordering: syntax before empty first document, empty
   first document before multiple documents, multiple documents before root type.
2. `schemaVersion` is checked next. The only supported version is canonical
   unquoted integer 1. Missing, null, quoted, noninteger, alias, or out-of-range
   versions give `version.invalid`; smaller integers give `version.too-old`;
   larger give `version.too-new`. Duplicate version keys give
   `format.duplicate-key`. A version problem is the only problem. Its path is
   `schemaVersion`; omit the inherited raw `found` detail for privacy. A good
   version with an anchor produces the ordinary alias problem after kind checks.
3. `kind` is checked before the body. It must be the scalar text `preparation`,
   compared exactly without trimming or case folding. Quotes do not matter.
   Missing, null, or bare empty kind gives `kind.missing`; quoted empty text,
   other scalars, lists, maps, and aliases give `kind.not-a-preparation`.
   A repeated kind gives `format.duplicate-key`, regardless of its values.
   This kind problem is the only problem, at `kind`, and its message includes
   `kind: preparation` without echoing the supplied value. Scalar spelling
   checks do not override kind checks. A valid anchored kind produces ordinary
   `format.alias` at `kind`, alongside body problems in file order.
4. After good version and kind, report all ordinary guards and field problems
   in source order, not alphabetical path order. At the same place retain the
   plan decoder table order: spelling, alias, duplicate, wrong type, unknown
   field, empty item. An anchor is before the node's other problem, as specified
   there. Duplicate values are still read; unknown nested values are not.

The inherited guards cover raw repeated mapping keys at every depth, anchors,
aliases, merge keys, non-scalar keys, boolean spellings, number spellings, and
null spellings, including preserved unknown data. A quoted `"<<"` is an
ordinary key. Repeated merge keys produce alias problems, not duplicate problems.
Explicit tags do not override scalar text/style rules. Dates and timestamps
are text, not time values. `deferred` is always text: preparation has no answers.

## Fields

Known root keys are exactly `schemaVersion`, `kind`, `account`, `planId`,
`recipient`, and `arrangements`. `kind` is stored in `Preparation.Kind`, unlike
the older Record model. `planId` goes into `PlanID`. Other scalar fields map to
the matching model field without normalization, trimming, validation, or copying
from another field.

`account` and `planId` are text. Strings use the plan decoder's title rule:
absent, bare empty, and null become empty strings; quoted or plain text is
preserved exactly; numbers and noncanonical booleans/nulls are spelling errors;
plain booleans, lists, and maps are wrong types. Quoted numbers and words such
as `"yes"` stay text.

`recipient` is an optional map. Absent, null, and empty map mean the zero
PreparationRecipient. Its exact keys, all text, are:

- `name`, `companyEmail`, `plannedStartDate`, `team`, `manager`,
  `onboardingOwner`, `buddy`, `workLocation`, and `timeZone`.

`plannedStartDate` is stored as `model.Date`, preserving its text; this decoder
does not check calendar validity. There is no actual-start or first-use field
here. See [0180](../../../docs/decisions/0180-first-onboarding-use.md).

`arrangements` is an optional list of maps. Absent, null, and empty list become
nil. Each map has these exact keys:

- Text: `id`, `section`, `kind`, `owner`, `reference`, `targetDate`, `status`,
  and `note`. `targetDate` is stored as Date without calendar validation.
- `contacts`: an optional list of text. Absent, null, and empty list become nil.
  Quoted empty text stays an empty string. Null contact items are
  `format.empty-item`; nested lists/maps and plain booleans are wrong types.

Lists keep input order and ordinary repeats. A null arrangement item is
`format.empty-item`; a non-map item is `format.wrong-type`; an empty map is a
zero Arrangement and is not dropped. Duplicate arrangement IDs pass through:
the preparation validator, not the decoder, rejects domain ambiguity. Repeated
raw mapping keys are different and are rejected here under
[0110](../../../docs/decisions/0110-reject-ambiguous-plans.md).

Unknown root keys are preserved in `UnknownFields` as plain data: string,
boolean, nil, int64, float64, `[]any`, or `map[string]any`. Preserve empty maps,
empty lists, and null list items there. An absent unknown-field map is nil.
Unknown keys in recipient or arrangement maps are `format.unknown-field` at
their own paths, including `x-` keys and differently cased known keys. Do not
read their values for further errors. A value of the wrong type is also not
traversed. No default, role lookup, or organization/personal answer merge occurs.

## Properties and limits

Identical bytes produce deeply equal values and problems, including repeated
calls on one decoder instance. Input bytes remain unchanged, including a UTF-8
byte order mark. Returned nested maps, lists, contacts, and arrangements from
different calls share no mutable state. LF, CRLF, comments, flow style, quoted
strings, block strings, key reorderings, and a missing final newline are valid.

The preparation validator owns identities, known references, statuses, kinds,
date validity, and arranged-item completeness. No filesystem paths, archive
layout, clocks, credentials, semantic PII detection, or readiness assessment is
part of decoding. The plan decoder's other YAML exclusions still apply.
Canonical encoder files that pass the structural and version/kind rules decode
to their represented model values; input formatting need not round-trip.

## Wrong implementations tried

| Wrong implementation | Caught by |
| -- | -- |
| Always return zero, with no problems | Every output fixture and every problem fixture |
| Reject every file | `minimal`, `all-fields`, `unknown-values`, `unvalidated` |
| Use loose YAML-to-struct decoding without guards | `boolean-spelling`, `duplicate-root`, `nested-unknown`, `number-spelling` |
| Drop kind, account, planId, recipient fields, or arrangement fields | `all-fields`, `strings`, `key-order` |
| Sort, merge, or discard repeated arrangements or helpers | `list-order`, `unvalidated` |
| Fill missing information or validate semantic fields while decoding | `minimal`, `unvalidated`, `zero-values` |
| Drop unknown data or convert every number to float64 | `unknown-values` |
| Stop after one problem or sort diagnostics by path | `problem-order`, `duplicate-nested`, `same-place` |
| Echo a supplied personal value or parser excerpt | `private-type`, `private-kind`, `private-syntax` |
| Judge kind before version or body before kind | `version-before-kind`, `kind-before-body` |
| Share retained lists or maps between calls | Result-isolation property |

Each fixture cites its independent source in `testdata/cases.json`. The tests
do not import an implementation except through the constructor binding.
