# Preparation editor contract

Component: `preparation/editor`. Agent-authored contract; not human-reviewed.

## Behavior and interface

The editor applies one explicit onboarder edit to person-specific preparation.
Its [API](api.go) has three operations:

- SetRecipient(context, current, recipient) returns preparation and problems.
- SetArrangement(context, current, arrangement) returns preparation and problems.
- RemoveArrangement(context, current, id) returns preparation and problems.

The [preparation format](../../../docs/reference/preparation-format.md),
[0140](../../../docs/decisions/0140-person-specific-preparation.md), and
[0160](../../../docs/decisions/0160-personal-data-and-preparation-approval.md)
supply the bounded shape and separate ownership. [0180](../../../docs/decisions/0180-first-onboarding-use.md)
supplies planned-date meaning. The editor reads its Context projection and every
Preparation field. It changes only Recipient or Arrangements in a returned
Preparation. It cannot read or change a plan or newcomer record because neither
body appears in its interface. Unknown top-level data remains unchanged.

Context holds the expected Account and PlanID plus known Contacts, Teams, and
Sections. It is trusted, supplied by a driver after checking the plan. It has no
record, clock, current date, identity inference, or arbitrary overlay. Drivers
must not derive the expected account from untrusted preparation itself.

Every call is pure, deterministic, and leaves every argument and previous result
unchanged. No file, clock, process, network, saved state, or effect outside its
returned value. Callers treat model values as immutable; this is not a promise
to deep-copy arbitrary opaque objects reachable through UnknownFields.

## Guards and scoped validation

Apply these phases, stopping after any phase that produces problems:

1. Check current schemaVersion, then kind, then account and planId using the
   [preparation validator](../validator/contract.md)'s first three phases without
   a record projection. Unsupported version returns only preparation.version;
   wrong kind only preparation.kind. Account and plan identity failures are
   returned together in that order. These guards never repair identity.
2. Scan existing arrangements in input order for repeated exact IDs. Each
   occurrence after the first returns preparation.ambiguous-arrangement at
   arrangements[index].id. This includes malformed or empty repeated IDs:
   an id-only operation cannot distinguish those entries. Reject ambiguity
   before every operation, even an unrelated recipient edit or removal.
3. Apply operation-specific validation below. On failure return the original
   preparation with all its fields unchanged, alongside problems. Never return
   a partially changed value.

A successful edit certifies the requested change, not readiness of the whole
preparation. Validate only the supplied recipient or arrangement, preserving
unrelated existing invalid details without reporting, repairing, or dropping
them. This permits incremental repair: two stale contact references in different
arrangements can be corrected one at a time. Full preparation validation and
invitation readiness remain separate driver steps. Duplicate identity and wrong
person or plan are not unrelated details and always stop the edit.

This bounded repair choice applies P3/P4 in
[0100](../../../docs/decisions/0100-behavior-judgment.md); ambiguity still fails
under [0110](../../../docs/decisions/0110-reject-ambiguous-plans.md). The product
decision is [0190](../../../docs/decisions/0190-incremental-preparation-repair.md).

### SetRecipient

Replace the entire recipient value, not just its nonempty fields. An explicitly
empty field clears that field; an empty recipient clears all recipient details.
Validate supplied fields by the preparation validator's Recipient rules and
diagnostic order. Empty optional details and absent company email remain valid.
Do not validate existing arrangements except the ambiguity guard.

### SetArrangement

Find the exact supplied permanent ID in the existing list. If it exists, replace
the entire item at the same position, clearing any omitted fields. Otherwise
append a new item after every existing arrangement. Never sort, deduplicate,
merge fields, or change any other arrangement.

Validate the supplied item by the preparation validator's Arrangements rules,
including ID syntax, references, explicit kind and status, calendar dates,
and arranged completeness. Paths use the actual replacement index or the index
immediately after the existing list when appending. Preserve repeated helpers.
Unknown kinds do not trigger kind-dependent completeness problems. Empty status
does not infer pending or arranged. Preparation arranged never means a newcomer
outcome. A target date is never compared with today or the planned date.

### RemoveArrangement

Remove the unique item whose ID exactly equals the supplied ID. No match returns
preparation.unknown-arrangement at arrangements. Do not trim, normalize, infer,
or choose a nearby ID. Do not validate the removed item's other fields: removing
one invalid item is a supported repair. A malformed or empty ID is a valid
explicit removal target when precisely one existing item matches it. A future
driver must expose that action clearly rather than treating missing input as
an instruction to remove an unnamed item.

Keep every remaining item in order. Removing the final item returns an absent
arrangements list. Other operations preserve absent versus empty collections
unless their named edit necessarily changes that collection.

## Problems

Use preparation validator codes and paths for identity and changed-data
validation. The editor adds preparation.ambiguous-arrangement and
preparation.unknown-arrangement as specified above. Each problem has severity
error, a nonblank useful message, and exact path in Detail.path. Diagnostic
messages never echo supplied personal values. Paths use fixed field names and
zero-based indices, not supplied IDs. Independent changed-data failures follow
validator order and are all returned. Nil or empty problems means success.

## Fixtures, tests, and binding

[cases.json](testdata/cases.json) contains hand-specified Context, starting
preparation, operations, and expected results or unchanged-result requirements
plus exact code/path pairs. Each case cites its source. An unchanged result on
failure is an explicit invariant, not an expected result derived by another
editor. [editor_test.go](editor_test.go) compares every field, checks input
preservation and diagnostics, and forks edits to protect prior values.
[wire_test.go](wire_test.go) is a stub until an independent implementation.

The editor performs no serialization, secret scanning, privacy inference, file
authorization, invitation conflict handling, first-use event recording, or
newcomer outcome updates. It does not silently accept defaults or manufacture
support that an onboarder has not explicitly supplied.

## Wrong implementations tried

| Defect | Caught by |
| -- | -- |
| Return zero preparation and no problems | Every fixture |
| Return unchanged preparation for every operation | Successful recipient, replacement, append, and removal fixtures |
| Replace recipient by a nonempty-field merge | clear-recipient-explicitly |
| Append a replacement instead of keeping its index | replace-arrangement-in-place |
| Silently ignore a missing removal target | remove-missing-not-success |
| Return only the first failure | invalid-recipient-no-partial-change, duplicate-id-blocks-replace |
| Ignore a lone validation problem | single-invalid-recipient-no-partial-change |
| Reject several duplicates but choose a winner when only two declarations conflict | two-duplicate-ids-block-replace |
