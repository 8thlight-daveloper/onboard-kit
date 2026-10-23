# Additive acceptance ZIP file expectations

Status: implemented acceptance-runner format.
The matching external-package tests are in
`acceptance/runner/zip_assertions_test.go`. No product component imports are used.

## Syntax

```yaml
expect:
  files:
    invitation.zip:
      zip:
        entries:
          - name: plan.obk.yaml
            same-as: acceptance/testdata/invitation-plan.obk.yaml
            method: store
            mode: '0600'
            modified: '1980-01-01T00:00:00Z'
```

This is a new file assertion, not an extraction action or the invitation decoder.
All existing file assertions, source copies, output matching, absence and unchanged
checks retain their current behavior. A `zip` assertion cannot be combined with
`same-as`, `json` or `absent` in that same FileExpectation, including absent false.
This restriction is additive for ZIP; it does not change old combinations.

`zip` is a mapping with exactly one required key, `entries`, a nonempty sequence
of entry mappings. Every entry requires exactly these five nonempty string fields:
`name`, `same-as`, `method`, `mode`, `modified`. Unknown keys at either level are
parse errors. Missing, null, empty, wrong-kind entries/fields are parse errors.
Scalar metadata must actually be text, not a numeric/boolean YAML scalar.

Mapping keys in this new assertion must be unambiguous. Repeating the outer
`zip` key, an inner `entries` key, or any entry metadata key is a parse error,
even when the repeated values are identical. Do not select a first/last value
or silently merge them. This rule applies only to the new ZIP assertion shape;
it does not change the existing duplicate-key policy for legacy-only file
assertions. Repeated entry names in separate sequence items are different:
they remain valid literal metadata under the exact ordered-list rule below.

Method is exactly lowercase `store` or `deflate`. Mode is exactly four octal
digits matching `0[0-7]{3}`, including `0000`; no special mode bits are implied.
Modified uses the current Go `time.Parse(time.RFC3339Nano, value)` acceptance
semantics, not the outcome recorder's stricter new-write profile. In particular,
do not claim strict RFC grammar while accepting the parser's single-digit hour
or comma fraction spellings. Parser-rejected text or invalid calendar dates are
parse errors. Compare actual ZIP modification time
with that instant, not with the spelling or location name of the timestamp.
Equivalent offset representations are equivalent expected instants.

`name` is literal ZIP metadata. Compare exact strings in central-directory order;
do not clean, case-fold, treat it as a filesystem path, infer a namespace, or
deduplicate it. Names that contain separators, parent segments or repetitions are
allowed in this generic expectation when explicitly listed. Product invitation
entry-name/duplicate policy belongs to the separate invitation decoder contract.
The list length and ordering are exact, including duplicated names if explicitly
expected. An entry may not be missing or extra. Do not compare via a name map.

`same-as` is a repository fixture path. Parse rejects a path that is not local
under `filepath.IsLocal`, including absolute/parent-escaping paths and empty text.
At Run, resolve the repository root and fixture symlinks, require the actual
fixture target to remain inside that root, and read only an existing regular
fixture file. A symlink escaping the root is a scenario failure, not permission
to read the outside file even when its bytes would match. A safe interior symlink
may name a regular fixture. Missing fixtures fail with useful context.

## Inspection and resource bounds

Route ZIP expectations to a bounded read path before the runner's generic file
read. Check the actual regular archive file with a capped read of at most
2 MiB plus one byte; reject over 2 MiB before unbounded reading/allocation.
Checking `len(os.ReadFile(...))` only after an unlimited read does not satisfy
this route. The bound is inclusive. Existing non-ZIP assertions are unchanged.

Parse central-directory metadata without extracting any entry. Require exactly
the expected entry count and ordered names, supported exact compression method,
regular file type, exact mode bits (including absence of special flags), and
the expected modification instant. Unix/macOS raw type attributes may not disguise
an unsupported type as regular through a lossy convenience conversion. For Unix
creator systems 3 and 19, a nonzero raw type must identify a regular file; zero
type is an unspecified ordinary file. Other creator systems keep their supported
ZIP type interpretation, not a fabricated universal Unix interpretation.

Stream every matched payload fully, checking read/close errors and its actual
CRC against the central checksum, even when that checksum is zero. Compare exact
payload bytes with the safe repository fixture, including empty payloads. Bound
the aggregate actual uncompressed bytes across entries to 1 MiB, inclusive, with
at most one overflow-detection byte beyond the remaining budget. Never trust
declared sizes as allocation/read allowances. A decoder/parser/checksum failure,
truncation, unsupported method, symlink, directory or other unsupported entry type
is an assertion failure, not a panic, successful partial prefix or extraction.
This is not a complete ZIP-conformance or authenticity assertion beyond the
explicit fields, byte comparison, type, checksum and resource bounds here.

Harmless ZIP prefix/comment/extra metadata is not byte-canonicality policy: the
assertion checks the listed observable metadata and payloads. A usable directory
returned alongside `archive/zip.ErrInsecurePath` is still inspectable literal
metadata; environment-dependent extraction warnings do not turn a non-extracting
checker into an entry-name policy. Never execute a prefix or payload.

The failure identifies the step and `file <archive> zip`, or `file <archive>`
when the archive itself is missing. Its message identifies the failed expectation
and relevant entry/field, without dumping payload bytes. Wording is not frozen.

## Independent tests and review boundary

Tests create small ZIP fixtures with the standard-library writer, then give the
runner opaque bytes as strings through the existing repositoryFiles helper and
fake binary's copy action. Expected payloads are independent literal strings;
expected names/order/method/mode/time are independent literals. No product ZIP
encoder/decoder or new runner Go API is imported. On the old runner, a valid new
assertion must reach the explicit Parse-success assertion and fail because ZIP
is unknown, not because a new exported type or method does not compile.

Wrong implementations to guard against: ignore ZIP assertions; use name maps or subset counts;
ignore mode/time/method/type; accept CRC zero without comparing it; stop before
EOF or compare only a prefix; allocate from declared sizes; unbounded archive
read; ignore aggregate bytes; read expected fixtures outside the repository;
extract entry paths; or break existing assertion kinds. Independent external-
package tests define the runner's expected behavior.
