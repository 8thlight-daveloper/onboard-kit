# Invitation decoder contract

Component: `invitation/decoder`. Agent-authored contract; not human-reviewed.

## Behavior and interface

`Decode(bytes) -> InvitationFiles, problems` reads the bounded invitation ZIP
envelope of [0240](../../../docs/decisions/0240-invitation-archive-envelope.md).
The [API](api.go), [tests](decoder_test.go), [fixtures](testdata/cases.json),
and [binding](wire_test.go) complete the four-part contract. The binding selects
the independently authored implementation's New constructor. Stub checks replace
that binding without changing assertions.

Reads: the supplied bytes only. Writes: none. No plan access, extraction,
filesystem, process, network, clock, environment, or artifact-codec calls.
The result holds an account and four opaque payloads; empty or non-UTF-8 bytes
are allowed. YAML validity, cross-content identity, readiness, fresh seed
construction, first-use recording, and import conflict handling belong to
later drivers. Success is not authorization or proof of authenticity.

## Ordered validation

Stop at the first failed phase. Return the zero InvitationFiles, with no
payload bytes or account, and exactly one problem. In entry phases, inspect
central-directory entries in their stored order; payload reads use the canonical
plan, preparation, record, first-steps order regardless of directory order.

1. Refuse more than 2097152 supplied bytes with invitation.too-large and no
   path, before ZIP parsing. The boundary itself is valid.
2. A malformed, truncated, or otherwise unreadable ZIP directory gives
   invitation.not-zip with no path. No panic, partial success, or parser text.
   A usable directory accompanied only by a ZIP library's insecure-path warning
   remains readable: apply the exact name checks below. Environment-dependent
   library path policy must not change these diagnostic phases.
3. Require exactly four central-directory entries. Otherwise return
   invitation.entries with no path before inspecting names or payloads.
4. For each entry, require an exact allowed central-directory name:
   plan.obk.yaml, first-steps.md, records/<account>/preparation.obk.yaml, or
   records/<account>/record.obk.yaml. Account follows the complete ASCII
   pattern `[A-Za-z0-9][A-Za-z0-9._-]*`; the full name fits 4096 bytes. Do not
   clean, normalize, decode, case-fold, or interpret a path. Reject an unknown,
   repeated, overlong, invalid-account, or mismatched-account name with
   invitation.entry-name and no path. Require one of each role. Both account
   segments match exactly, including case. Do not infer account from contents.
5. For each entry in directory order, require a regular file, not a directory,
   symlink, socket, device, FIFO, or other declared special type. Otherwise
   invitation.entry-type at its fixed role path. Regular files with ordinary
   permission differences remain valid. A slash suffix is not a regular file.
6. For each entry in directory order, reject general-purpose encryption bits
   0, 6, or 13 with invitation.encrypted at its role path. Accept only Store
   (method 0) or Deflate (method 8); otherwise invitation.compression there.
   Encryption takes precedence over method on the same entry.
7. Check declared uncompressed lengths individually and in aggregate, using
   overflow-safe arithmetic. More than 1048576 bytes gives invitation.too-large
   with no path before any payload is opened. Never allocate from unchecked
   compressed sizes, uncompressed sizes, ratios, or ZIP64 declarations.
8. Read each payload to EOF through a bounded reader. Enforce the same actual
   aggregate budget, require exact declared versus actual lengths, and require
   successful CRC, data descriptor, and compressed-stream checks. CRC zero is
   not a signal to skip verification: nonempty data with a different checksum
   must fail even when the advertised CRC is zero. An actual
   budget excess gives invitation.too-large with no path. An unreadable or
   corrupt entry, checksum error, truncated stream, or size mismatch gives
   invitation.entry-data at its fixed role path. If the ZIP reader rejects a
   malformed stream before yielding its excess bytes, entry-data is valid;
   never return more than the budget or a partial result.

Fixed role paths are plan, preparation, record, and firstSteps, not archive names
containing an account. Every problem has severity error, a useful nonblank
message explaining repair, empty Question and Section, and Detail containing
only path when specified. No supplied names, account, payloads, or underlying
parser excerpts are included. Empty or nil problems mean success.

## Accepted noncanonical ZIPs

The reader accepts Store and Deflate, any entry order, ordinary permission
bits, timestamps, bounded comments and extra fields, and supported data
descriptors. Central-directory names are the routing authority; a different
local-header name is inert metadata and need not match. Never extract either
name. Prefixes and trailing bytes are accepted when the ZIP remains structurally
readable within the raw limit. Other unrecognized metadata is not retained.
This is not a byte-canonicality or all-ZIP-variants certification promise.
Archive permissions, comments, and dates never become onboarding facts.

## Fixtures and properties

Each case cites its authority. The fixtures are tiny synthetic payloads, not
usable invitation documents. A record-layout fixture helper constructs ZIP
bytes directly from declarative entries, preserving repeated names, explicit
corruptions, and untrusted declared lengths. Store CRC constants and a complete
small ZIP golden independently check that helper. Deflate controls use the
standard compressor only for test transport construction, not for expected
payloads, which are hand-specified.

Assert exact account and bytes, expected diagnostic code/path, useful private
messages, no partial failure value, determinism, unchanged inputs, isolated
returned slices/maps, and no retained state. Nil and empty payloads are both
zero-length contents; their Go representation need not be distinguished.
Test exact boundaries without huge allocations. A rebuild first runs against
a stub and every output or problem fixture must fail.

## Wrong implementations tried

| Wrong implementation | Caught by |
| -- | -- |
| Return empty success or reject every input | Output and error fixtures |
| Ignore extra or duplicate entries | Exact entry count and name cases |
| Normalize traversal, case variants, or account spellings | Entry-name cases |
| Treat symlinks as harmless regular payloads | Entry-type cases |
| Ignore encryption flags or unsupported methods | Entry encoding cases |
| Trust a declared size or skip CRC verification | Overflow, size, checksum cases |
| Require the writer's exact order or metadata | Noncanonical valid controls |
| Parse payload YAML or record a first-use timestamp | Opaque payload controls |
| Return partial content or leak a supplied name | Failure and privacy checks |

The working report identifies actually exercised wrong adapters separately
from this required coverage and later production mutation evidence.
