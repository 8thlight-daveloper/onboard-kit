# Starter storage contract

Status: complete independently authored normative detail under0400;
implementation pending, with no passing implementation or native-death claim.
These new private formats are not legacy compatibility claims.
All paths are derived from selected capabilities and the fixed role mapping,
not an arbitrary journal participant framework.

## JSON, bounds and physical guards

Every protocol file is valid UTF8, one JSON object, no BOM or trailing JSON value.
Reject duplicate decoded object keys before map conversion, unknown/missing keys,
wrong types, unsupported version and malformed Unicode/surrogate escapes. Readers
may accept JSON whitespace/key order; writers use the key order below, no spaces,
and one final LF. Strings use standard quote/backslash and short control escapes;
other C0 characters use lowercase four-digit Unicode escapes. U+2028/U+2029 are
escaped, other valid Unicode and slash/< />/& remain literal. No HTML escaping.
All numbers except literal version1 are strings, never float-converted. Version
is exactly numeric token1, not1.0/1e0. Decimal strings are0 or an unsigned nonzero
digit followed by digits, within uint64; mode is exactly four octal digits.

Binary strings are canonical standard padded base64: reject wrong alphabet,
whitespace, unnecessary/missing padding and nonzero pad bits. Preflight encoded
length and decoded size before allocating; count actual decoded bytes again.
SHA256 is64 lowercase hexadecimal digits over exact bytes, not semantic identity
or authenticity. Unknown protocol data is preserved/refused, never silently
ignored, migrated, pruned or deleted because its name resembles an application file.

Security objects have exact ordered fields `uid,gid,mode,aclFormat,accessPresent,
accessBase64,defaultPresent,defaultBase64`. UID/GID are canonical unsigned strings,
mode is Unix permission/special bits masked07777. aclFormat is darwin-filesec-v1
or linux-posix-v1. Darwin records exact returned extended-security/filesec ACL
bytes; absence differs from a present zero-entry ACL. Its defaultPresent is
false/defaultBase64 empty. Linux records exact access/default ACL xattr bytes
and absence separately. False presence requires empty bytes. Unsupported/failed
native inspection refuses; do not infer absence from mode bits. Native ACL data
is capped65536 decoded bytes per object before allocation; excess refuses.

Every newly created private application object has native UID equal to the
process effective UID and GID equal to its effective GID, exact mode0600 for a
regular file or0700 for a directory, and no access/default ACL. Establish these
controls before personal bytes. Existing fixed private application objects,
including selected state HOME and import-pending/session namespaces, require
effective UID ownership, a GID in the process's effective or supplementary group
set, the same exact private mode/no-ACL profile, and exact guarded ownership
preservation; do not reset a permitted existing group to the effective group.
Failure to read the native group set refuses. These rules do not broaden or
normalize original selected ROOT or participant controls: retain their separately
accepted exact guards, including another permitted owned group, and preserve
those controls on replacement or refuse. Newly created enclosed payloads start
with effective UID/GID, then may receive the exact frozen original controls only
inside the claimed private staging directory as specified below.

An ObjectGuard has exact fields `kind,device,inode,security`; kind is directory
or regular, device/inode canonical uint64 decimal strings. It excludes timestamps,
link count and filename guesses. Preserve fstat identity and security around the
bounded body read; compare exact bytes independently. Nofollow traversal and
all-participant alias checks remain necessary even when guards/digests match.

| Resource | Inclusive bound |
| -- | -- |
| Raw invitation |2097152; actual reviewer enforces decompressed1048576 |
| Each selected plan/preparation/record and final organization artifact |1048576 |
| Each session.json preimage/proposal |2097152 |
| Inspected registry entries / aggregate metadata |128 /16777216 raw bytes |
| Selected/stored absolute or relative path |4096 UTF8 bytes |
| Four-role preimage/proposal blobs combined |10485760 decoded bytes; at most8 present blobs |
| Frozen session display |8388608 decoded bytes |
| Blobs plus display |18874368 decoded bytes |
| pending.json |33554432 raw bytes |
| Each owner.json/publication.json/complete.json |16384 raw bytes |

Check both raw/decoded/aggregate bounds with overflow-safe arithmetic. Do not
allocate decoded blobs from unchecked JSON/base64 lengths. Render display through
a capped builder before offering begin, not a huge allocation followed by len.
Preflight role encodings/security/claim budgets before personal staging. The
limits can refuse an otherwise semantically valid source; no truncation follows.

## Fixed registration

HOME/sessions/KEY/ has exactly plan.obk.yaml and session.json for a completed
registration. The two-file namespace is private but organizational plan/record
content remains under its original ownership rules. Pending state is checked
before a registry reader treats potentially mixed registration as usable.

KEY is lowercase SHA256 of these exact bytes, with uint64 big-endian byte lengths:

```text
UTF8("onboard-kit/session-key/v1") || 0x00 ||
U64BE(len(planID)) || UTF8(planID) ||
U64BE(len(account)) || UTF8(account) ||
U64BE(len(registeredAbsoluteRoot)) || UTF8(registeredAbsoluteRoot)
```

Raw identities never form path components. Reused physical aliases keep the
existing key and registered root spelling after continuity/namespace checks.

session.json exact keys/order: `version,sessionKey,planIdBase64,account,root,
rootDevice,rootInode,record,preparation,planSHA256,transaction`. All are strings
except version1. planIdBase64 decodes to exact nonempty UTF8 Plan.ID, at most1MiB.
Account uses actual reviewer identity; root is the registered safe absolute root;
record/preparation are exact locator-relative destinations. rootDevice/rootInode
bind the actual root object; planSHA256 hashes selected incoming plan bytes;
transaction is the transaction that created or last changed these registration
facts. Do not rewrite equivalent metadata merely to insert a new visit token.
planSHA256 describes the last accepted incoming registration context, not a
live-file immutability seal. A subsequently edited/wrong-ID local plan still
follows actual Review's explicit replacement/unavailable-comparison rules;
do not reject it merely because its current bytes differ from metadata's digest.
New/changed registration gets the current transaction token after begin; its
fixed64-character width is reserved during pre-consent encoding preflight.

Scan HOME/sessions boundedly: request at most129 directory entries, not an
unbounded listing then count. Every entry counts, including malformed/unknown
ones; preserve/refuse them. Parse metadata one at a time into bounded identity
projections, never retain128 maximum decoded trees. Verify directory key against
stored identity/root framing. Do not read all registered record bodies or prune
missing entries. Root continuity and destination-conflict rules in contract.md
decide whether an otherwise valid entry is relevant to this opening.

## One discoverable pending namespace

HOME/import-pending is the only transaction namespace for this home. Its four
protocol names are owner.json, pending.json, publication.json, complete.json.
Recognized atomic staging names are .owner.next, .pending.next, .publication.next
and .complete.next. All are private, exclusively created and capability-relative.
Unknown extra entries preserve/refuse. No recursive cleanup or prefix ownership.
Every creation is after complete begin/resume/clean consent as applicable.
The fixed container may remain empty between transactions. Reuse it only after
safely opening it and verifying native ownership, mode0700 and ACL absence.
An empty safe fixed container is idle, not missing-owner recovery state. This
exception never authorizes adoption or deletion of an unknown child, staging
object or nonempty directory. Unsafe/nonempty unattributed state is preserved
and refused. Do not create even the idle container before consent.

Generate transaction as32 cryptographic random bytes/64 lowercase hex only after
fresh begin and final checks. A token is a lookup identifier, not authority.
Acquire actual-root/home exclusion before bootstrap/staging, not token-only locks.

owner.json exact keys/order: `version,transaction,rootDevice,rootInode,
namespace,pendingSHA256,pendingTemporary`. Namespace is its actual directory
ObjectGuard. pendingSHA256 is empty before complete frozen payload exists, then
its exact raw digest. pendingTemporary is null or ObjectGuard for exclusively
created .pending.next. Publish that exact inode/security claim BEFORE writing
personal preimages/proposals into the pending temporary. Only then write, fully
close, verify and no-replace publish complete pending.json. Update owner intent
atomically through .owner.next; no personal blob is stored in owner intent.

An interruption between empty namespace/temp creation and durable attribution
can leave an unattributed EMPTY object. Preserve/refuse it; filename/UID/mode
alone do not prove ownership. No personal bytes may have been written in that
window. Recognizable claimed incomplete staging can be cleaned after explicit
clean consent, then requires reopening the invitation. ZIP-free roll-forward
begins only after a complete, verified pending.json exists. Missing final pending
never makes claimed staging undiscoverable; inspect the fixed namespace first.

## Immutable pending payload

pending.json exact keys/order: `version,transaction,sessionKey,planIdBase64,
account,root,rootGuard,home,homeGuard,record,preparation,firstOpenedAt,
sessionDisplayBase64,roles`. Root/home are safely bound absolute strings; guards
are directory ObjectGuards. Root is the registered root used for path derivation,
not a new implicit move. firstOpenedAt is the actual final Publication scalar,
possibly empty for an existing unknown event; preserve it exactly. Session display
is already bounded literal-safe terminal bytes, not JSON-encoded Review, unknown
maps or a renderer request. Never journal the raw ZIP, transcript or raw Problems.
Decoded display must be valid UTF8 and contain no raw control/format/nonprintable
rune except structural LF; Q-escaped control spellings remain ordinary ASCII.
Reject unsafe display rather than print or sanitize it during recovery. Valid
display is replayed byte-for-byte, not rerendered/reflowed/reescaped. This is a
terminal-safety format check, not a second semantic activity renderer.

roles is exactly four entries in this order, with no duplicates/unknown entries:

| Role | Base and fixed derived target | Actions |
| -- | -- | -- |
| local-plan | HOME/sessions/KEY/plan.obk.yaml |retain/install/replace |
| preparation | ROOT plus exact preparation relative path |retain/install |
| session-metadata | HOME/sessions/KEY/session.json |retain/install/replace |
| record | ROOT plus exact record relative path |retain/create/reconcile/bind |

Each role has exact keys/order `role,action,beforePresent,beforeBase64,beforeSHA256,
beforeGuard,afterBase64,afterSHA256,parents`. after bytes are final authoritative
Publication bytes, except metadata is the independently specified registration.
Before false requires empty bytes/hash and null guard; true requires exact
preimage digest and regular ObjectGuard. Every after payload is present and has
its exact digest. Retain requires before true and exact identical before/after
bytes; it is checked but never replaced. No absent/empty conflation.

parents has one entry per target-parent path component below its base, in path
order: `ordinal,present,guard`. Ordinal is canonical decimal zero-based index;
present true has the exact existing directory guard; false has null guard. Once
an ancestor was absent, later components are absent too. Names derive only from
the fixed role target; no journal-supplied arbitrary path list. Existing ancestor
guards remain mandatory; allowed missing tails authorize only private creation,
not adopting an unexpected future directory. Do not create participant-parent
directories before complete pending publication. State-home bootstrap is the
separate approved empty-directory exception before final home locks/rechecks.

pending.json is immutable once published. Recovery consumes these exact final
bytes/actions/event/display, never Reviewer.Inspect/Review/Begin, a new archive,
clock, changed content merge or semantic repair. The complete raw pending digest
binds subsequent claims and receipt. All individual sizes/digests/schema/source
guards are checked before any recovery role operation.

## Mutable ownership claims, frozen role bytes

Each role uses a PRIVATE staging DIRECTORY in its safely held actual participant
parent, on that volume, named `.obk-start-TOKEN-ROLE`; ROLE uses the fixed spelling
above. Inside is one fixed file `payload`. They are not all under HOME: same-volume
atomic rename is required even when ROOT and HOME are on different volumes. The
0700/no-ACL directory shields payload bytes while final target controls such as
0644/0400 are prepared. Existing entries refuse; no retry/overwrite of unknown
objects. At most one owned staging directory/payload per role, four pairs total.
Review discloses private recovery copies both in HOME and participant parents.

publication.json exact keys/order: `version,transaction,pendingSHA256,roles`.
roles has exactly four entries in fixed order, each `role,newParents,staging,
temporary`.
newParents is an ordered list of `ordinal,guard` for only originally absent
parent components created by this transaction. staging is null or the exact
directory ObjectGuard of that role's exclusively created private staging
directory; temporary is null or the regular ObjectGuard of its exclusively
created payload. Every new parent
is claimed by component ordinal and actual native identity/security before using
it for personal role bytes. Claim staging and payload before personal proposal
bytes. Initial payload controls are0600/no ACL, current UID/GID; final target
controls derive from the guarded preimage for replacements or remain private
for installs. Inside private staging, an owned payload may have either its
claimed initial security or exact intended final target security while controls
are prepared. No arbitrary intermediate security is accepted. Verify final
controls before rename; the claimed native inode remains the publication identity.
No existing ancestor path/security data is repeated here; pending already holds it.
When fixed roles share an originally missing parent, the first creation claim
owns that physical directory. Later roles reuse only that exact prior claimed
path/parent-object chain and native identity/security, not a new unclaimed object
by the same spelling. This covers local-plan/metadata under sessions/KEY and
preparation/record siblings. Do not duplicate ownership or delete shared target
parents as temporary cleanup; they remain legitimate artifact directories.

Creation-before-claim interruption preserves/refuses unattributed EMPTY objects.
It is not permission to adopt a directory/temp by matching name, owner or mode.
Preflight the16384-byte claim budget conservatively before personal role bytes;
very deep missing tails may therefore refuse. Keep one bounded claim object,
not an append-only phase log. Atomic .publication.next updates are private and
checked against the prior known claim; unexpected replacement/security refuses.

A target at guarded preimage identity/bytes/security permits its frozen approved
write. A target at the claimed payload inode, exact after bytes and intended
FINAL target security is already applied; initial temporary0600 security is not
the final-target comparison when original controls differ. Absent preimage
permits only no-replace publication. Retain requires its
original guard/bytes. Matching bytes on an unclaimed replacement inode are not
applied proof. Unexpected target bytes/identity/security stop further publication
and preserve owner work/pending. Source hard-link anchors not selected as targets
retain old inode/bytes when the selected name is atomically replaced.

Existing ACL-bearing replacements may be conservatively refused if exact control
copying is unavailable; no-op retention needs no replacement ACL capability.
New private dirs/initial payloads require0700/0600, correct ownership and no
access/default ACL before any personal bytes; inherited grants/explicit-empty
ACL are not absence. A payload with final preserved0644 controls is still shielded
inside its private0700 staging directory until publication.
Existing replacements preserve UID/GID/mode/ACL exactly or refuse. Parent ACLs
are never stripped and unrelated objects never chmodded to make publication pass.

## All-role completion and receipt-last cleanup

Publish roles in fixed order: local plan, preparation, metadata, record. Verify
EVERY intended target and every required guard after all operations. A no-op
record or byte equality alone cannot prove that companions were applied.

complete.json exact keys/order: `version,transaction,sessionKey,pendingSHA256,
rootDevice,rootInode,homeDevice,homeInode,
namespaceDevice,namespaceInode,namespaceSecuritySHA256,roles`.
roles has four fixed entries `role,sha256,device,inode,securitySHA256`. The security
hash is SHA256 of canonical Security JSON bytes without final LF. The receipt
binds the exact pending payload and independently verified final identities/bytes/
controls; it does not carry event/display or become sender authentication.
Root/home identity fields bind the safely selected context even after pending
and owner intent were removed. Receipt-only cleanup rechecks safely held root/
home identities and the owned cleanup namespace, not later ordinary root/home
security changes that remain safe and not later role bytes/security. Namespace
security hash protects the application-owned cleanup object itself. Do not
reread session metadata or record/plan/preparation bodies to clean completion.
A safe root
spelling alias is allowed only when it still names the bound object; no implicit
move or replaced root/home/namespace is accepted.
Atomically publish it through .complete.next only after all-role verification.
Receipt failure remains incomplete/pending, even when all role bytes now match.

A valid receipt means cleanup-only. Never classify later ordinary owner edits
as unfinished import, revalidate them against old proposals, republish roles or
resample time. If pending remains, explicit clean may display its frozen session
after cleanup. If pending is already absent, clean succeeds with a distinct
cleaned status and explicitly says no session was displayed. Do not manufacture
lost display bytes or reconstruct content using Reviewer or current artifacts.

Cleanup removes exact attributable private pending/payload remnants and empty
owned role-staging directories first,
then publication/owner claims and staging remnants, and receipt LAST. Keep receipt
through every personal-data deletion so process death cannot turn completion
into an unrecognized missing-pending transaction. After receipt removal leave
the verified empty fixed import-pending directory as the private idle container;
do not perform a final rmdir. Unknown extras are preserved and
reported; no recursive delete. If complete exists with pending absent, receipt
plus selected root/home context is a valid cleanup-only state, not a seed request.

Ordinary errors/process interruption can leave approved empty directories,
partial approved participants and bounded private recovery data. Report that
truthfully. Before published pending, no participant may have changed. After
receipt, cleanup/display failure never becomes cancelled/unchanged. No group
atomicity, rollback, hostile-owner CAS or power-loss durability is promised.
