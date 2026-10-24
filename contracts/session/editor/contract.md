# Session editor driver

## 1. Purpose, API and effects

New native position8 imperative driver for ONE explicit answer, follow-plan or
verified/failed/deferred record change. Binary9 routes `session edit` and passes
its raw remaining arguments to api.go Run. New(clock Clock) constructs actual
session reviewer7, actual plan decoder2 and preparation locator; it stores Clock
without calling it. The public API imports only io, not model or another
contract. Now returns exact injected event text; production entrypoint supplies
the actual UTC RFC3339Nano wall-clock adapter. Nil Clock refuses non-help valid
requests without filesystem/input. No global-clock fallback or product flag.

Reads bounded native root/home/registry/pending/ancestors/selected plan,
preparation, record. Writes only cooperative lock effects, one bounded private
staging directory/payload AFTER save, then one atomic record replacement. Never
changes plan/preparation/registration/firstOpenedAt/reference/history except the
actual pure selected change. No other driver, process, check, command, URL,
environment lookup, CWD discovery/chdir, network, automatic import/reconciliation,
record creation, identity guessing, default acceptance, consent shortcut or
unbounded retry. Borrowed streams and Clock are never closed/mutated.

Agent-selected proposed contract, not human approval or execution. Existing
read-only start/legacy walk remain unchanged. Pure dependencies grant no native
trust; this driver separately proves selected context and explicit consent.

## 2. Exact native/interaction contract

### CLI and reports

Arguments are exactly `--root ABS --home ABS --session KEY`, each once in any
order, separate values only. No equals, short flags, `--`, unknowns, duplicates,
positionals, missing/empty values or values starting `--`. ABS must be UTF8,
absolute, <=4096 bytes, no NUL or literal `..` component. Safely traverse . or
redundant separators before normalizing, never normalize away a symlink. KEY is
exactly64 lowercase hex. `--help` ALONE prints testdata/help.txt and no stderr,
filesystem, input or Clock. Syntax failure precedes every runtime fact/I/O
except writing its fixed report. Existing ROOT/HOME/session are required.

Normally accepting stderr gets ONE compact JSON object and LF, ordered
status/code/message/saved, no dynamic values. Saved is the actual publication
fact, never consent or intended outcome. Reports/messages below are exact:

| Exit | status | code | message | saved |
| -- | -- | -- | -- | -- |
|0|cancelled|edit.cancelled|Session change cancelled; no record was written.|false|
|0|unchanged|edit.unchanged|The record is unchanged.|false|
|0|saved|edit.saved|The selected record change was saved.|true|
|2|refused|edit.usage|Use an explicit root, state home and registered session key.|false|
|1|refused|edit.context|Select a safe unchanged registered session context.|false|
|1|refused|edit.pending|Finish or repair the pending import first.|false|
|1|refused|edit.source|Read the bounded existing session files before editing.|false|
|1|refused|edit.review|Resolve the displayed session problem before saving.|false|
|1|refused|edit.input|Use complete valid displayed actions within the input limits.|false|
|1|refused|edit.output|Required output could not be completed; no record was written.|false|
|1|refused|edit.limit|The selected session or display exceeds the supported limit.|false|
|1|refused|edit.commit|The record was not replaced; private staging may need repair.|false|
|1|saved|edit.repair|The record was saved, but cleanup or reporting needs repair.|true|

Exit/Result.Saved agree with report. Final report failure returns1; after saved
attempt exactly one edit.repair fallback, never rollback/resample/replay. Help
output failure uses edit.output without input or filesystem. Real cleanup/handle
failure after rename outranks later output/report errors. Ordinary EOF before
save cancels; input errors/invalid/partial action never consent. Failure after
rename always reports saved/repair, including a failed final stdout success line.

### Before identity: metadata only

Use public `contracts/invitation/starter/storage.md` JSON/registration/KEY,
native capability, security and bounded-registry rules as NORMATIVE TEXT, not a
starter implementation/dependency. Open real ROOT/HOME nofollow, native effective
UID/group/ACL-safe existing private HOME. Acquire nonblocking root/home descriptor
locks, deduplicated by unsigned device/inode and acquired in that numeric order.
Unsupported/busy lock refuses. No PID/token mutex/lockfile or expiry. Check
HOME/import-pending FIRST: absent or verified safe EMPTY idle only. ANY child,
including complete receipt, .next, unknown or malformed state, refuses preserved
with edit.pending. Never read potentially mixed session bodies or auto-clean.

Bound registry to128 entries,16MiB total metadata and2MiB each, requesting at
most129 entries before refusal; every entry counts. Verify strict metadata,
KEY framing, fixed private sessions/KEY two-file namespace, selected key and
original registered root native continuity. Matching old roots must still exist
at original spelling and correct native identity; no guessed move/reseed.
Other valid account/root entries are retained; inspect metadata/native relevant
namespaces, never their record/preparation bodies. Reject ambiguous/colliding
actual participants even for different case accounts or nested roots. ROOT/HOME
nesting alone is NOT a collision. Genuine native aliases require actual same
device/inode AND identical destinations/namespace; keep original stored spelling
and key. Symlinks are never aliases. Different valid disjoint contexts are not
selected or disclosed as progress for this explicit selected-key route.

An entry is matching only when BOTH its decoded plan ID and account exactly
equal the selected trusted identity. Every such old root, including another
matching context, must exist safely at its original spelling and satisfy the
existing native continuity rules. Selected ROOT/source requirements are unchanged.

For an otherwise valid NONmatching registry entry, demonstrable native ENOENT
absence of its stored root or a final record/preparation target is not by itself
a reason to block this selected edit. Validate all metadata/key/private-registry
requirements and prove relevant actual destination/bookkeeping namespaces remain
nonaliasing/disjoint using safely inspected existing prefixes, parent identities,
literal remaining names and stable absence guards. Keep those proofs fresh at
the existing identity/save/publication recheck boundaries. No synthetic nil/
zero object guard, root-string inequality, shared missing sentinel or lexical
fallback proves absence or disjointness. Same missing basename under genuinely
distinct native parents is not itself an alias.

Only actual ENOENT through a safely inspected no-follow prefix is absence here.
A symlink, dangling symlink, ELOOP, ENOTDIR, permission error, unsupported/failed
native inspection, unknown type or ambiguous namespace is NOT absence; preserve
and refuse edit.context before identity/input when safety cannot be established.
No existing root/participant safety or applicable continuity requirement is
relaxed by this narrow allowance. Actual hardlink/case/ancestor collisions still
refuse, even if a different unrelated final participant is absent.

Retain unrelated registration plan/metadata bytes, native controls, root objects
and absence exactly. Never create/repair/prune unrelated paths, read their body
as selected content, choose them as a winner, expose their account/plan/key/root/
progress in this route, or borrow their first event. Identity and both consents,
selected three-source checks, exact record-only publication and Clock0 for an
answer are unchanged. The allowance does not authorize broad reading of other
record/preparation bodies or automatic registry repair.

Initial metadata gives exact nonempty trusted plan/account, relative record and
preparation destinations, registered plan and metadata paths. Validate bounded
relative path shape without repairing it; verify participant/ancestor identities
and nonaliasing, no FIFO/socket/device/symlink or actual hardlink collisions among
plan/preparation/record/metadata/staging. Before identity read only metadata/native
facts, not source bodies. Display terminal.md identity with paths labeled
REGISTERED and current locator binding explicitly not yet checked. Release locks
before every interactive wait. Hold only safely owned capabilities needed for
later continuity; release all owned descriptors on every return.

### Identity, actual source and action

Only exact `confirm Q(account)` plus complete LF/CRLF consents to reading the
selected three bodies. Q and all terminal/input rules are in terminal.md.
Reacquire ordered locks and recheck pending/registry/parents/native identities
and exact metadata bytes after identity; changed state refuses, not a new
unreviewed identity. Read each selected existing regular body <=1048576 bytes,
extra-byte bounded read, O_NONBLOCK plus fstat before body read, nofollow held
directory capabilities, exact pre/post bytes/native/security guards. Total3MiB.
No unbound/missing record, synthesized prep, bootstrap or raw-hash-as-live-seal.

Now call actual plan decoder on bounded raw plan ONLY to obtain ResultsPath;
call actual locator with it and exact metadata account. Require its returned
relative paths exactly match selected metadata and all native destination
namespaces. Failure uses edit.review for actual codec/locator errors or
edit.context for differing stored destinations. No other YAML parser, invented
path interpolation, fallback to metadata after locator refusal or hidden move.
Use actual reviewer Inspect with all3 exact bodies and trusted identity. Its
fatal errors refuse after safe problem display; its resource error maps edit.limit.
Keep old FirstOpenedAt/On/reference opaque under its contract. Release locks;
fully build/display actual current session before any action byte.

One exact action from terminal.md, no retries/blank defaults. Call actual Propose
with frozen raw Source and explicit intent. Display complete selected-target
before/after change and actual all-field Session projections, warnings, privacy,
ownership/follow-plan behavior, current hash/prerequisites, canonicalization and
private staging notice BEFORE reading save. Verified is the person's statement,
not a check result. Outcome proposed On says sampled after save, never sentinel.
Full fatal actual diagnostics refuse, no usable partial proposal or hidden gate.

### Save freshness and event

Only exact save after completed review consents. Reacquire root/home exclusions;
recheck pending idle state, bounded registry/metadata/context, all relevant native
ancestors/security, participant distinctness, exact3 source bytes AND native
guards, selected locator destinations and fixed staging idle/absence. Any stale
fact refuses edit.context with Clock0, no merge/retry. Metadata planSHA256 remains
historical registration context, not equality with live edited plan bytes.

Now and ONLY when actual Proposal.NeedsEvent call Clock.Now ONCE, convert inside
driver to timestamp and call actual Finalize. Answers/follow-plan Clock0. Actual
invalid clock text refuses through reviewer, no normalization/fallback/resample.
Finalize may refuse actual candidate resource charge; review never guaranteed
every event length would fit. No staging before final successful encoded bytes.
A final model no-op keeps exact raw record bytes, native inode and metadata;
no stage/rename. Display no-op and report unchanged. Every postclock guard is
rechecked again immediately before publication; late change preserves owner work.

### Record-only native publication and persistent metadata

Private fixed directory in actual record parent `.obk-record-edit-KEY`, at most
one file record.new. No journal/backup/arbitrary schema/transaction token/replay.
Create exclusively after save; initial mode0700/current UID/GID/native ACL absent
before personal bytes. Reuse only proven EMPTY safe owned private directory;
unknown/nonempty/linked/uninspectable state refuses preserved, no prefix adoption.
Payload exclusive nofollow create0600/current UID/GID/no ACL, actual held inode,
private directory shielding. Write complete final <=1MiB bytes; verify them.

Freeze exact record UID/GID/mode, native ACL presence/payload (including explicit
empty/inherited), native persistent flags, and ALL xattr names/bytes, not only
known names. Bound names256, name-list65536 bytes, each value65536, aggregate1MiB;
overflow/unsupported/failed native read refuses. Enumeration/read pre/post guards
must be stable. Native ACL is separately inspected where not an xattr. Temporal
stat fields, link count and new inode identity necessarily change on publication;
they are NOT persistent security metadata preservation promises. ROOT/parents
are never chmodded or stripped. Native flags that prohibit safe replacement
strictly refuse rather than clear an owner's immutable/append/security flag.

Darwin automatically created com.apple.provenance is NOT ignored because fresh.
Inside private staging preserve EVERY frozen persistent record attribute/ACL,
UID/GID/mode/flag exactly or refuse. Extra automatic payload attrs may not be
silently dropped/ignored; if exact final source set cannot be established by
permitted native operations, refuse original intact. New private objects can
have automatic non-granting attrs, but no ACL/private-control exception. Never
strip unknown selected source metadata, coerce ACL absence from mode, drop
capability/label metadata or claim ordinary success after conservative refusal.

Before rename verify held final bytes/security/all attrs and fresh original3
bodies/native/context/pending state again under exclusions. Same-volume atomic
rename over exactly selected original record entry; no in-place truncate and
no other artifact rename. A distinct old hardlink anchor keeps original inode,
bytes and metadata. Publication truth begins at successful rename, not reporting.
Pre-rename failure never changes original record; only live actual held staged
inode/namespace proof permits cleanup. Unknown substitutions are preserved.
After success leave safely empty staging directory idle; no recursive cleanup
or final rmdir. Postrename failures Saved=true, report repair, no rollback.

After process death, abandoned nonempty stage is preserved/refused without body
adoption, replay, deletion or fresh event. No automated repair is offered here.
Single pathname atomicity prevents partially published record; this is not
four-file atomicity, power-loss durability, hostile-owner CAS or proof every
timing point was reached. Those limits appear in help. Native real positive
fixtures must actually save with preserved attributes; refusal is a coverage/
delivery gap, not successful newcomer flow.

## 3. Dependencies and architecture

Driver8: actual session reviewer7, actual plan decoder2 solely for postidentity
ResultsPath, actual preparation locator, own model-free contract, standard native
OS and pinned x/sys/unix. Use public native protocol as specification; no private
starter imports/code copying or another driver call. Runtime Clock is external
fact. Model conversion is private driver glue, not public API. Binary9 wires
this and other drivers plus real UTC clock. No position7 peer shortcut.

## 4. Independent fixtures and verification boundary

External Go tests create actual native files/guards/strict registrations from
public schema and independent framed standard SHA256. No fake filesystem,
injected semantic dependency, product flag or product-generated golden. Literal
three-artifact answer/follow/outcome values and native before/after byte/inode/
all-metadata snapshots judge writes. Required display-failure writers implement
both Write and WriteString explicitly through the same faulting path (no
embedded Buffer bypass). Native lock child is test-only real second process.
Stub/production bindings are separate. Binary scenarios REALLY invoke the
provided binary, but proposed generic capture capability is separately reviewed
before those new-format data scenarios can run. Native Run is not binary proof.
No compilation, test, mutant, native crash timing or publication is claimed here.
