# Invitation encoder contract

Component: `invitation/encoder`. Agent-authored contract; not human-reviewed.

## Behavior and interface

`Encode(InvitationFiles) -> bytes, problems` packages four opaque payloads for
one explicit account. The [API](api.go), [tests](encoder_test.go),
[fixtures](testdata/cases.json), and [binding](wire_test.go) complete the
four-part contract. The initial binding is an empty stub. A separate contract-only
binding increment selects the independent implementation's New constructor.
The transport model and envelope profile come from
[0240](../../../docs/decisions/0240-invitation-archive-envelope.md), extending
[0090](../../../docs/decisions/0090-plan-and-onboarding-record.md) and
[0140](../../../docs/decisions/0140-person-specific-preparation.md).

Reads: Account, Plan, Preparation, Record, and FirstSteps arguments only.
Writes: none. No plan access, artifact decoding, filesystem, network, clock,
environment, randomness, extraction, permission changes, or first-use event.
Payloads are opaque bytes, including empty and invalid UTF-8. Packaging success
does not certify identities inside payloads, onboarding readiness, private
content, bootstrap wording, authenticity, or import safety.

## Validation and problems

Stop at the first failed phase and return no bytes and exactly one problem:

1. Account must match the entire ASCII pattern
   `[A-Za-z0-9][A-Za-z0-9._-]*`. Do not trim, normalize, case-fold, infer, or
   substitute an account. Otherwise return invitation.account at account.
2. The longest generated name, records/<account>/preparation.obk.yaml, must
   fit 4096 bytes, inclusive. Check the account length before constructing
   names. Otherwise return invitation.path-too-long at account.
3. The aggregate length of the four payloads must fit 1048576 bytes, inclusive.
   Check with overflow-safe arithmetic. Otherwise return invitation.too-large
   with no path. Empty or nil payloads both encode as empty entries.

Every problem has severity error, an explanatory nonblank message naming the
rule and repair, empty Question and Section, and Detail containing only the
fixed path when specified. No supplied account, payload, archive parser
excerpt, or arbitrary path is echoed. Empty or nil problems mean success.

## Canonical bytes

Emit one classic ZIP, with exactly these entries in this order:

1. plan.obk.yaml from Plan
2. records/<account>/preparation.obk.yaml from Preparation
3. records/<account>/record.obk.yaml from Record
4. first-steps.md from FirstSteps

Preserve account and payload bytes exactly. There are no directory entries,
archive comment, entry comments, extra fields, preamble, trailing bytes,
ZIP64 records, or data descriptors. The raw result fits 2097152 bytes.

Each local header has signature 0x04034b50, needed version 20, flags 0,
method 0 (Store), DOS time 0, DOS date 0x0021, and the IEEE CRC-32 of its exact
payload. Compressed and uncompressed sizes both equal the payload byte length.
The name length is its ASCII byte length; extra length is zero. The exact name
then payload follow the header. Integers use little-endian ZIP encoding.

All four central-directory records follow all local records in the same order.
Each has signature 0x02014b50, creator version 0x0314 (Unix, version 20), needed
version 20, the same flags, method, time, date, CRC, sizes, and name as its local
header; extra length, comment length, disk start, and internal attributes zero;
external attributes 0x81800000 (Unix regular file, mode 0600); and the exact
zero-based local-header byte offset. The name follows with no other data.

Finish with one end-of-central-directory record, signature 0x06054b50, disk
numbers zero, both entry counts four, exact central-directory size and offset,
and comment length zero. Do not stamp the current time or encode payload dates
into metadata. Archive modes do not imply actual filesystem privacy.

## Fixtures and properties

Fixture bytes are synthetic profile-only examples, not functional invitations.
Each case cites its source. The smallest complete golden is manually pinned
as hexadecimal bytes; a separate ZIP record-layout oracle constructs larger
boundary expectations with checked fixed metadata. No new decoder is an
encoder oracle. Tests compare complete bytes and inspect entries independently.

Calls preserve input slices and earlier results, share no retained mutable
state, and return independently owned output bytes and diagnostic maps.
Identical inputs yield identical bytes and problems. A failed call does not
poison a subsequent successful call. Success returns no problems; failure
never returns a partial archive. A rebuild first runs against a stub.

## Wrong implementations tried

| Wrong implementation | Caught by |
| -- | -- |
| Return no archive or reject every input | Successful golden cases |
| Drop a payload or change its bytes | Exact complete archive expectations |
| Case-fold or trim the account | mixed-account and invalid-account cases |
| Use current timestamps, Deflate, or data descriptors | Canonical record bytes |
| Reject the inclusive size boundary or allow one extra byte | Payload boundary tests |
| Build an overlong entry name or silently truncate it | Account path boundary tests |
| Echo supplied personal data in errors | Privacy marker tests |

The working evidence report records which wrong adapters were actually run;
the table names the required checks, not completed mutation evidence.
