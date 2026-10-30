# Catalog edit guide contract

Status: independent proposed four-part shell contract under0410,0440,0450 and0580.
No product source, promotion, execution or approval evidence is supplied.

## API, dependencies and effects

New() takes no arguments and constructs actual catalog editor6, catalog
materializer5, plan decoder2, plan encoder1, authoring validator2 and plan
validator4. Position7 permits only those, model, own contract, standard library
and already pinned x/sys/unix native facilities. No driver dependency, injected
filesystem/semantic component, alternate catalog, clock, record or preparation.
Binary8 routes exact catalog edit before command/catalog. Other routes unchanged.

Reads explicit argv/actions, bounded selected plan bytes, selected path/native
controls, pure dependency results and borrowed stream results. Writes streams
and only after reviewed save one owned same-parent publication temporary and
the explicitly selected existing plan. Descriptor-only locks are not persistent
files. No directories, backups, manifests, registration, preferences, account
discovery, record/preparation reads/writes, clock/environment identity, network,
process/shell/Git, command/check execution or personal events/outcomes.

## CLI and stable admission

Exactly --plan FILE or --plan=FILE; no duplicates, unknown/positional arguments,
empty path, --out, overwrite flag or implicit file discovery. Separate value
cannot start --; first-equals permits it literally. Initial -h/--help wins over
all later data before filesystem/dependency/input access. Later help is usage.
After initial-help: at most2 argv entries/8192 total UTF8 bytes, decoded and
anchored path4096 bytes each, valid UTF8/no NUL. Violations use edit.usage exit2.
Help is exact testdata/help.txt, stderr empty. Help write failure edit.output.

Anchor relative FILE once at cwd. Hold checked nonsymlink directory capabilities;
reject literal .., missing/non-directory/symlink components, trailing separator,
empty filename or symlink/nonregular destination. Dot/redundant separators may
collapse only after safe traversal of material components. No path expansion.

Selected parent: effective UID, any retained GID and no group/other WRITE
permission.0755/0700 are admitted; do not require0700 or effective-group parent.
Parent ACLs are not inspected. Directories above the selected parent are
traversed without symlinks but are not inspected for owner, mode, ACL, flags,
attributes or change; entries coming and going there never refuse a save. A local
user with write access above the selected parent is outside the claimed threat
under0580. Preserve exact selected parent owner/GID/mode/ACL/xattr/flag state; never
chmod/chown it, remove ACLs or normalize product paths to qualify.

Selected original: regular, exactly one native link, effective UID/GID, exact
mode0600, no ACL and no nonzero native file flags. Extended attributes must be
absent, except that Darwin may have exactly com.apple.provenance. Capture its
complete opaque byte value, including presence versus absence and empty values,
through the held descriptor. Never parse, synthesize, redact or log its value.
Every other name remains unsupported and refuses edit.path before startup/input.
Existing65536-byte representation and262144-byte aggregate metadata caps apply
before allocation. Inspection failure is not absence. Refuse uninspectable
or unsupported controls, inherited-privacy uncertainty, special bits, broader
mode, hardlinks, nonregular or absent files with edit.path before input/startup.
This is not universal support for existing repository plans. Do not silently
make an original more private or strip unsupported metadata. Own temporary
controls may be established before writing its data, never on unrelated files.

Hold an original native snapshot: device/inode, file kind, links, UID/GID/mode,
size, mtime/ctime, native flags, exact ACL/xattr facts and selected parent
identity/controls. Read at most1048576 source bytes, with a
one-byte excess probe, through the held nonsymlink regular descriptor. Compare
before/after fstat controls and namespace identity; any instability refuses.
Each native ACL/xattr representation is capped65536 bytes before allocation;
combined captured metadata262144 bytes. No unsupported fact is assumed empty.

Decode actual bytes. On any error emit safe P and edit.data, no input. Run actual
editor.Edit(original,nil) as bounded immutable admission, then actual encoder
on the returned original proposal to establish representability, then actual
decoder on that encoding and compare modeled values under ordinary codec
roundtrip equivalence: only nil versus empty known maps/slices may normalize,
not optional pointer presence; unknown
numeric kind/value, unknown null/empty data, exact strings, sequence order,
conditions, Source/baselines, answers and explicit booleans must survive.
Do not implement a second parser or drop metadata to obtain agreement. This
comparison uses decoded ordinary values, not raw YAML style/comments. Unknown
int/int64 dynamic width equivalence is allowed; float64 remains distinct.
Refuse edit.fidelity if representable modeled content is lost or changed.
Initial encoder errors or encoding over1048576 bytes use edit.fidelity (P for
actual errors only), before roundtrip decoding; roundtrip decode errors also
use edit.fidelity with safe P. Never decode an oversized generated encoding.
Actual editor guards precede expanding whole-plan validation. Actual authoring
then actual plan Validate are called on the admitted held proposal; preserve
actual errors and safe projections, not shadow validation. No startup/input on
initial errors. Keep exact original raw bytes unchanged for review/stale guards.

## Input, quotation and staging

Use the creator's public framing/quoting rules as specified here, not its driver
or contract package as a runtime dependency: no
read-ahead,65536 UTF8 payload bytes per logical line (LF/complete CRLF framing),
1048576 consumed bytes total including framing. Never probe beyond remaining
budget. Partial EOF is not a field/action; ordinary valid EOF before save cancels.
Invalid UTF8/oversize/error/zero-progress refuses edit.input. Positive count
plus EOF completes a framed line; nonEOF error vetoes its bytes. No trim,
casefold, global q/save or default choice. Complete save/cancel needs no trailing
EOF probe. Q and D are creator's exact rules, including final empty LF part;
printable Unicode literal, unsafe controls quoted. No ANSI/raw terminal operation.

Startup is testdata/startup.txt with @PLAN@ replaced by Q(anchored path).
Every return to Draft prints exactly:

```text
Edit draft — not saved. Choose add, submit, unset, retire, reactivate, drop, review, or cancel:
```

Unknown/empty actions print Choose one of the displayed actions. then this
prompt. Exact cancel at an action prompt cancels. Field values are always
literal, including cancel/save/drop. No persistent change before consent.

Actions stage typed ordered editor operations, NOT already accepted plan edits.
Display Operation staged; validation occurs at review. on append, then Draft.
Maximum1024 staged operations/256 Adds; reject over-cap action with
Edit limit reached; drop a staged operation before adding another. and no further
field reads. No per-prefix editor call, so Reactivate then Submit can be reviewed
as one atomic batch. Do not infer sources/identities or validate via a mirror.

Add prompts Source ID:, Permanent section ID:, Initial binding count: 0-64:.
Canonical decimal only, invalid count retries with Use a canonical count from
0 through 64. For each binding ask Binding name: then Value kind: text, flag,
list, or deferred:. Text asks Text value (empty text is an exact new binding):.
Flag asks Flag value: true or false: and exact true/false; invalid retries with
Use true or false. List asks List length: 0-4096:, canonical decimal; invalid
retries Use a canonical count from 0 through 4096. Then List item N: (1-based).
Deferred reads no payload. Invalid kind prints the fixed invalid-action line
and repeats kind. Empty/repeated items and names stay exact for actual editor
validation; no None kind is offered. Add's initial bindings are exact materializer
values, not empty-clearing submissions. Zero initial bindings is usable.

Submit prompts Instance ID:, Variable name:, then the same value-kind forms,
except text prompt is Text value (empty text clears an existing answer):.
Unset asks Instance ID:, Variable name: and constructs the named Unset operation.
Retire/Reactivate ask Instance ID: only. Never check activity/ownership locally;
actual editor decides the whole ordered batch. Omitted IDs remain unchanged.
No rename, definition-edit, refresh/prune, delete or plan-scope editing actions.
Help/startup point to catalog list/show for exact source IDs and declarations.

Drop asks Staged operation number:. Accept canonical1..current count, otherwise
print Select an existing staged operation number. and return Draft. Remove only
that in-memory operation; print Operation dropped. It cannot delete a saved
section or release any held ID reservation. All reviews become invalid after
staging/drop/back; never reuse old review bytes or consent.

## Safe diagnostics

P is Problems: N followed in actual order by:
`  - Stage: Q(stage) Path: Q(path) Code: Q(code) Severity: Q(severity) Repair: Q(message)`.
Actual catalog editor diagnostics already have vetted private Stage/Path/Code/
Severity/Message; preserve them exactly and never reconstruct raw metadata.
For other actual dependencies preserve ordered Code/Severity and use stage
decode, encode, authoring, plan, or guidance with Path exactly plan (never raw
Detail.path/Question/Section/Message, which can include private unknown keys).
No filtering/deduplication, invented warnings or readiness interpretation.

Fixed other-dependency repairs: decode → Repair the selected plan's format;
duplicate keys and invalid spellings must be fixed explicitly.; encode →
Preserve or explicitly repair unrepresentable plan data before saving.;
authoring → Repair the selected plan's declarations before editing.; plan →
Repair the selected plan's structure or stored facts before editing.; guidance →
Ask the catalog maintainer to repair the advisory preparation source.
secret. codes override these with Remove credential-like content before saving.
Directory warning codes get fixed actionable overrides in this shell only:
directory.contact-unreachable → Supply an approved business route for the
declared contact.; directory.role-vacant → Name an accountable holder for the
declared role.; directory.email-malformed → Repair the declared business email
address. Do not change core diagnostic bytes; these overrides apply only to
non-editor dependencies. Codes remain useful without private contact details.

## Complete original/proposed review

Review calls actual editor.Edit(original, entire staged list). Errors emit
Review unavailable. then P(failing batch) and Draft; retain operations for Drop,
do not expose a save prompt. On success call actual authoring then actual plan
Validate on complete proposed plan and actual encoder then actual decoder for
the same roundtrip fidelity check. Any errors use Review unavailable.+P then
Draft; fidelity/output caps use Review unavailable: preserve all plan data
within supported limits. then Draft. No truncated/partial proposal permits save.

Encoded save bytes maximum1048576; complete expanded review maximum8388608,
checked with a capped builder before allocation beyond the limit. If actual
proposal equals original under that same modeled roundtrip equivalence, freeze
ORIGINAL raw bytes for save,
mark unchanged and do not canonicalize on no-op save. Otherwise freeze exact
actual encoder bytes. Never regenerate after consent. All semantic/unknown
metadata must roundtrip, not be silently dropped because it is unrelated.

Guidance uses actual materializer only for each staged Add, with original PlanID,
held Directory, zero provenance/no plan variables, and its exact initial bindings.
Do this only after editor/validation success. Merge actual guide IDs and causing
permanent IDs by exact byte order, distinct reasons, retaining exact Markdown.
Same guide ID must have identical bytes or refuse guidance. Do not infer retained
old guidance from stale Source, prose or catalog ID equality. Do not claim all
active work covered. Newly added then retired instances remain included, labeled
as unfiltered new-addition advice. Guidance errors never allow partial review.

Emit in this exact order (D/Q as above):

```text
Catalog edit review — not saved. No personal work or outcomes are changed.
```

Each item here is a separate LF-terminated line: Plan: Q(path); Operations: N;
each operation one `  - N Kind: Q(kind) Instance:
Q(id)` line. Add additionally Source: Q(source), Initial bindings: N and each
Binding: Q(name) Value: Q(kind), followed by exact selected payload text/flag/list
using Text: Q(text), Flag: true/false, List: N plus `  - Q(item)` lines; Deferred
has no payload. Submit uses Variable: Q(name) and Value/payload; Unset uses only
Variable; switches no extra fields. These operation detail lines have two-space
indent. Then D(Original plan bytes,raw original), D(Proposed plan bytes,save bytes),
Save action: replace. or Save action: unchanged; original bytes retained.
P(editor warnings followed by explicit review validation warnings, then guidance
warnings; real order/multiplicity retained). Preparation guidance for newly added
instances only: N, followed by creator-style Guide/Instance reasons/D(Advisory
instructions) blocks. Finish exactly:

```text
Advice includes retired new additions; no arrangement or activity is completed.
Review complete. Choose save, back, or cancel:
```

Every required write, including prompt, must fully succeed before another input
byte. Short/nil, zero/error and full/error are failures; edit.output and no save.
At Review only exact save/back/cancel; unknown prints invalid-action and repeats
only Review prompt. Back invalidates cache and returns Draft. Save at Draft is
invalid. EOF/partial save cancels, leaving exact original file unchanged.

## Consented native replacement

After complete save, acquire nonblocking cooperative advisory exclusive lock
on the selected held parent descriptor, without a lock file or preconsent lock.
Unsupported primitive/inspection refuses edit.commit; busy refuses edit.busy.
Every implementation of this command must use that same parent-inode lock
convention. Existing other tools are not retroactively cooperating writers.

While locked, recheck original selected parent identity/controls and namespace
identity, open current target nonsymlink, compare original device/inode/link/
UID/GID/mode/native metadata and fresh exact complete source bytes, then fstat
again. Absence, replacement, byte or control changes use edit.stale and preserve
the currently observed owner file. Even an equal-byte replacement inode is stale.
Every original attribute name, presence and exact value participates in these
guards. An observed addition, deletion or value change is edit.stale, including
newly unsupported metadata; never repair the observed original to qualify.
No automatic reload/rebase/rename-to-alternative or uncertain observed overwrite.
An unchanged semantic save still performs this check, then returns unchanged
without any temporary, inode replacement or timestamp-setting operation.

For changed save, create exactly one owned exclusive same-parent temporary.
Before any data, establish effective UID/GID, exact0600, no ACL, zero native
flags and a complete attribute map exactly equal to the held supported original.
If the original carries Darwin com.apple.provenance, set its exact captured
opaque bytes on the owned temporary, then independently inspect the entire map.
Successful assignment alone is insufficient. Unknown temporary attributes or an
inability to establish exact equality refuse edit.commit before data. Never
remove or ignore provenance to qualify. An attribute-free original requires an
attribute-free temporary; unexpected ambient provenance refuses rather than
silently changing metadata. Existing controls outside the owned temporary must
never be changed to qualify.
Do not strip inherited ACLs to qualify; refuse unsupported privacy. Full-write
and close checks required. Inspect temporary security facts and exact bytes;
verify the complete exact supported attribute map after writing and again in
the final temporary/source guard immediately before publication. Never publish
a mismatching map or restore a fresh ambient default. A prepublication failure
uses edit.commit/cleanup-refused; postpublication failures retain the saved
replacement under existing written/report/cleanup rules, never rollback.
freshly recheck guards/source bytes again immediately before atomic same-parent
replacement through held capabilities. Replace ONLY the still-matching selected
file. Retain exact approved parent controls. No in-place truncation or backup.

This is a cooperative present-object replacement, NOT hostile same-UID CAS.
Descriptor locks do not constrain noncooperating writers; there is an unavoidable
check/rename gap under that excluded threat. Do not claim to preserve native
inode, ctime/birth timestamps, raw formatting/comments on changed semantic save,
power-loss durability or all filesystems. Supported content/access metadata is
preserved or refused. No uncertain *observed* namespace change is overwritten.

Before publication failures clean only owned temporary and retain original;
after successful atomic replacement retain the complete saved file even if
cleanup/reporting fails. Never roll back over potentially newer work. Cleanup
failure is explicit before/after status. Unlock/close all owned descriptors;
never close borrowed streams. A failure cannot claim unchanged if replacement
already occurred. All adjacent paths are outside read/write scope, not scanned.

## Terminal reports

One LF-terminated JSON object on accepting stderr, exactly status/code/message,
keys order free. No raw native/parser/input/path/answer values. Save adds no
extra stdout beyond complete review. Initial successful help alone has no stderr.

| Exit | status | code | Exact message |
| -- | -- | -- | -- |
|0|cancelled|edit.cancelled|Cancelled. The selected plan was not changed.|
|0|unchanged|edit.unchanged|The reviewed proposal leaves the selected plan unchanged.|
|0|written|edit.written|The reviewed plan edits were saved.|
|2|refused|edit.usage|Use catalog edit --plan FILE.|
|1|refused|edit.path|Select a safe single-link effective-owned0600 plan in an owned directory without group/other write; unsupported controls require separate review.|
|1|refused|edit.data|Repair the displayed plan data before editing.|
|1|refused|edit.fidelity|All held plan metadata must survive before this file can be edited.|
|1|refused|edit.input|Use complete valid input within the displayed limits.|
|1|refused|edit.output|Required review output failed; no plan replacement was accepted.|
|1|refused|edit.busy|Another cooperating plan edit is saving; reload and review again.|
|1|refused|edit.stale|The selected plan context changed; reload and review again.|
|1|refused|edit.commit|The plan was not replaced; owned temporary cleanup may need repair.|
|1|refused|edit.cleanup-refused|The plan was not replaced; temporary cleanup needs repair.|
|1|written|edit.cleanup-written|The plan edits were saved; temporary cleanup needs repair.|
|1|written|edit.report-written|The plan edits were saved, but reporting failed.|

After replacement cleanup repair outranks report failure. Failed final report
attempts one fixed report-written fallback (or higher-priority cleanup-written).
Keep saved bytes. Before-write refusal report failure returns1 without retrying
actions. Unchanged/cancelled report failure returns1, no file write; no fallback
promises to an unwritable writer. No after-consent semantic dependency calls.

## Four parts and staged evidence

API, contract, independent fixtures/tests and zero constructor binding are the
four-part proposal. Expected save bytes and operation transcripts are authored,
never generated by product encoder/editor. Native fixtures prove stated setup
or name a platform gap, not chmod originals or simulate supported metadata.
Root must establish assertion-based stub-red before implementation and freeze
source before admitting tests. Wrong variants, independent build, streaming
differential, native race/mutation and binary fixtures remain execution duties;
drafting is neither passing evidence nor0410 delivery.
