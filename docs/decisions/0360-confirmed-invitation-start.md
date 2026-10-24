# 0360. Confirm an invitation before starting a local session

Status: accepted by agent judgment; confirmed startup and read-only session implemented

## Context

An invitation carries an organization plan, onboarder-owned preparation,
an identity-only record seed and the invitation's first-steps document.
Opening it is not ordinary extraction. The recipient must recognize the
identity, retain existing work and deliberately begin onboarding before the
tool creates a first-use event or writes personal data.

[0090](0090-plan-and-onboarding-record.md) requires local plan
persistence, existing-record reconciliation and an opened-record index.
[0140](0140-person-specific-preparation.md) separates plan,
preparation and record ownership. [0230](0230-first-use-record-compatibility.md)
defines first use; archive inspection is not first use. This decision narrows
the next supported start flow without introducing general synchronization.

The [record joiner](../../contracts/record/joiner/contract.md) deliberately overlays
all record answers regardless of audience. The [reconciler](../../contracts/record/reconciler/contract.md)'s
explanation that organization answers win does not match that joiner. Keep both
pure contracts unchanged here: this new driver rejects misowned answers instead
of silently filtering stored work.

## Decision

### One confirmed, record-backed beginning

The supported entry is:

```text
onboard-kit start --invitation FILE --root DIR [--home DIR]
```

The command requires an explicit invitation and selected record root. It does
not search for an archive, extract arbitrary entries or execute archive text.
The eventual contract defines exact argument, prompt and diagnostic bytes.

The flow is inspection, identity confirmation, complete action review, then
explicit `begin`. Identity confirmation accepts the exact account and plan
identity for this operation. It is not consent to publish files, accept default
answers or record first use. A different local plan requires an explicit choice
to use the incoming plan before the final review; cancellation remains safe.

The final review names artifact roles and intended actions: install or replace
the local plan, retain or install preparation, create or reconcile the record,
and register the local session. It shows actual reconciliation changes,
unavailable comparisons, stale support, privacy exposure and the recovery-data
location. It explains that `begin` starts a read-only onboarding session, not a
verification run or an assertion that any activity is complete.

After accepted begin and recoverable publication, enter an actual record-backed
walker view. It displays activities, resolved instructions, questions and
available support, including waiting or blocked work. It does not write answers
or outcomes, run checks or shell commands, provision access, export data, or
infer completion from preparation, dates or previous verification history.
Question audience and ownership remain visible. The invitation's supplied
first-steps document is identified as invitation-time guidance, not a freshly
validated description of retained local support.

This supported walker session has an explicit prerequisite: independently
contract and build the missing authoring-validation increment required by the
walker. The current plan validator excludes variable type/audience/pattern/
default/condition-reference authoring checks and is not full validation. Do not
call it full validation, weaken the walker's precondition or implement a shadow
generic validator inside this driver. Until the prerequisite is delivered,
an import-only preview cannot claim this supported beginning or create its
first-use event. Existing export/guide validation stages are unchanged.

Rendering first steps, confirming identity, viewing the proposal or cancelling
does not begin onboarding. Review output must complete successfully before
reading a queued `begin`, including into an input buffer. EOF or partial consent
does not authorize writes. No state directories, registration entries, lock
files, journals or temporary files are created before accepted begin.

### Validate roles, not an arbitrary ZIP filesystem

Use the actual invitation decoder and its fixed four roles and resource bounds
from 0240. Decode and validate the embedded plan, preparation and record seed
with their real components. The account is exact and case-sensitive, follows
0210, and agrees across routing and payloads. Plan identity must be nonempty.
Seed plan identity and exact embedded-plan-byte hash must match 0300. The seed
has no legacy start date, answers, outcomes, unknown root data or first-use event.
Its reference URL must match the embedded plan URL exactly; it is not fetched
or treated as authority.

The seed is only a proposal for a genuinely new record. It never replaces an
existing record, repairs malformed history or fills an existing missing event.
Archive checksums and digest-shaped names are integrity/lookup aids, not sender
authentication. This flow makes no provenance claim beyond the visible content.

Use the actual locator for record and preparation paths, not archive entry paths
or a second interpretation of `resultsPath`. Reject traversal, symlinks,
nonregular participants and ambiguous aliases. Bounds apply before allocating
or interpreting content. Resolve selected paths through checked directory
capabilities; preserve original bytes and filesystem identities for rechecks.

### Existing local work

An identical local plan is retained byte-for-byte. A different local plan is
replaced only after explicitly accepting the incoming plan and reviewing that
replacement. Do not use timestamps or hashes to guess which plan is newer.
Use an available valid previous plan for reconciliation comparisons. If it is
unavailable, say that previous question/activity comparisons are unavailable;
an empty comparison is not evidence of no changes.

Existing matching-identity preparation belongs to the onboarder. Retain its
exact bytes even when the invitation differs; do not merge fields or choose by
date. Undecodable preparation, wrong identity, duplicate identities or other
ambiguity refuses opening. Ordinary stale references may remain with actual
validator diagnostics. Do not pass invalid retained preparation to a dependency
whose precondition requires validity. Show support unavailable with the reason,
while independently valid record-backed instructions can remain available.
The contract must enumerate this narrow stale-reference allowance from actual
validator codes; it must not classify every validation error as tolerable.

An existing record must decode, have the exact selected account and have either
the incoming plan ID or an empty plan ID. A different nonempty ID refuses.
Binding an unbound record to the incoming plan is an explicit reviewed action.
Use the real reconciler, its real ordered report and proposed record. Preserve
hidden answers, unknown data, deferrals, outcomes and unrelated history under
that contract. Invalid declared answers may be removed only as that reconciler
specifies, with the removal visible before begin. Do not invent a merge.

Runtime history guards are deliberately narrow: outcome sections must be
unique, kinds must be known, and interpreted times must be real UTC timestamps
under 0085. Apply 0230 to an existing nonempty first-use event; malformed events
refuse rather than backfill. Accept held unsorted outcomes; the walker sorts
its returned view, not persistent history. Do not reject or repair a reference
solely because its hash is not in a preferred textual format. New contracts
must specify the exact accepted UTC profile independently.

Before joining, refuse a held record answer whose currently declared variable
has onboarder audience, including omitted/default audience. This applies in
plan and section scopes and includes inactive declarations. Preserve the file
intact; do not delete, filter or reinterpret the key. Hidden answers for variables
not currently declared remain held under 0090 and do not trigger this guard.
The guard follows 0055/0090 ownership and bounds this new driver's accepted input;
it does not change the pure joiner or reconciler. After this guard, use the real
joiner for the effective plan. Preparation values do not become answers.

### First use is the accepted beginning

After complete review, accepted begin and fresh source/path checks, a genuinely
new record receives `FirstOpenedAt` from an injected clock. Validate the sampled
UTC value and freeze it into the pending proposal exactly once. It represents
the accepted beginning, not proof that the entire subsequent display succeeded.

An existing nonempty event is preserved exactly. An existing empty event remains
unknown, including when its answers are empty. Do not derive an actual start
date or backfill from planned dates, filesystem metadata, Git, archive metadata
or the current time. Recovery uses the frozen proposal and never samples again.

### Private session state and known-context checks

Preserve state-home selection precedence: explicit `--home`, nonempty
`OBK_HOME`, then Darwin user configuration directory plus `obk`, or on Linux
nonempty `XDG_STATE_HOME` plus `obk`, otherwise user home plus `.local/state/obk`.
An explicitly empty option is invalid. Empty environment values are absent.
An invalid or inaccessible chosen path refuses; do not fall back to another
home. No supplied symlink is silently followed. Missing first-install directories
may be created only after accepted begin. Final contracts must pin relative
configured-path handling; this draft does not silently tighten legacy behavior.

Store each session under `HOME/sessions/KEY/`, with `plan.obk.yaml` and bounded
`session.json` metadata. The key is a domain-separated, length-framed SHA-256
lookup over exact plan ID, exact account and safely selected absolute root.
The contract defines framing and encoding. Raw identity is never used as a
filesystem path segment. The key is not authentication or proof of ownership.

The private per-session metadata entries themselves form the opened-record
index. No shared central index file is introduced. Entries retain exact identity,
registered root, root filesystem identity, located record/preparation paths and
their binding context. Keep missing-record entries; do not silently prune them.

Before creating a new session, inspect the bounded set of registered entries.
Compare exact account/plan identity and safely opened root objects, not just
path strings or digest keys. A spelling/case alias can reuse an entry only when
its stored path still safely names the same root and the derived destination
namespace agrees. Do not globally casefold accounts or paths. Actual filesystem
name collisions refuse rather than merge distinct accounts.

A stored root path now missing, changed to a different object, or relocated
while another spelling names its old inode is not an implicit move request.
Refuse automatic rebinding/reseeding and retain the entry. Device/inode checks
are cooperative present-object guards, not permanent cross-machine identities.
Detected `resultsPath` or destination changes also refuse, even if the new record
path is absent. A known record that is now missing does not become a fresh seed.
Unknown histories on another machine are outside this local detection guarantee.

All data and bookkeeping participants must be disjoint: invitation input, local
plan, preparation, record, session metadata, pending state and staging objects.
Reject same-inode hard links, incompatible ancestor relationships and aliases
into bookkeeping, including no-op cases. Root/home overlap alone is not an error
if all actual participant namespaces remain safely disjoint.

### One recoverable transaction per state home

This is a checked, recoverable multi-file protocol, not a cross-volume atomic
transaction. Serialize registration and publication for the actual selected
state root and record root, not only the textual session key. Prefer
process-lifetime descriptor locks on already existing safely opened directories,
acquired in a fixed order. Inspection and such locks must create no objects.
Missing state directories are created only after begin under the selected-root
ownership boundary; acquire state-root exclusion before registration/publication.
One pending transaction for the whole state home is the conservative limit.
No PID-age guesses, expiring lock files or automatic stale-lock deletion.
Native lock behavior and races require independent actual-platform tests.

After begin, use a private, recognizable application-owned staging namespace
under the chosen state home. Its ownership and recovery-discovery rules cover
interruption before a final manifest exists; unknown entries are preserved and
reported, never removed because their name resembles a temporary file.

Prefer one complete versioned `pending.json` containing bounded base64 exact
preimages and proposals, presence bits, fixed participant roles, checked paths,
source guards, digests, transaction identity and any frozen new first-use event.
Publish the complete immutable pending payload atomically before any participant
write. Incomplete staging can be cleaned only when ownership is established;
otherwise preserve and refuse. ZIP-free recovery begins only once the complete
pending payload is published. Before that boundary, safe cleanup may require
restarting inspection with the original invitation. No private staging becomes
undiscoverable merely because the final pending pathname is absent.

Journal format, item/aggregate byte bounds, ownership discovery and completion
receipt are separate normative contract work. Pending data duplicates personal
notes/history: disclose this in review, store no transcript, retain no completed
backup and allow no unbounded accumulation. Mode bits alone do not prove private
access. Check ACLs, including inherited grants and explicit empty ACLs, before
writing any personal preimage/proposal. Do not strip parent ACLs to proceed.

For every participant, recheck the exact guarded pathname/object and classify
its state as the recorded preimage, intended proposal, retained value or an
unexpected state. Retained preparation is checked but not rewritten. A matching
preimage permits its approved proposal; a matching proposal is already applied.
Unexpected bytes, replacement identity, permissions or access controls refuse
further publication while preserving the owner's work and pending data.
New paths use no-replace publication; existing replacements preserve-or-refuse
required access controls. Per-role encoding/size checks complete before staging.

Publish the local plan and preparation first, session registration metadata next,
and the record last among artifacts. A byte-identical record is a no-op, not
evidence that the transaction reached the record phase. After verifying every
participant's intended state and transaction/phase digests, persist an explicit
bounded completion receipt. Only that verified all-participant boundary is
overall completion. No record-equality shortcut may discard pending work.

Resume is explicit, shows the pending action set and needs no original ZIP once
fully staged. It uses the frozen proposals, not a new reconciliation or clock.
It rolls forward only matching preimage/proposal states. It never rolls back
unexpected owner edits, including after record publication. A complete receipt
prevents later ordinary owner edits being mistaken for an unfinished import.
Readers must notice pending state before using a potentially mixed session.
The exact resume command and receipt cleanup protocol remain contract work.

After completion, remove owned pending preimages/proposals and staging data.
Failure to clean up or display the session reports completed-with-repair, not
cancelled or unchanged. Do not republish or resample on cleanup retry. Before
completion, ordinary failures may leave approved partial publication and private
recovery data; report the actual recoverable state, not “no files changed.”
Close all owned resources before return; do not close borrowed input/output.

The guarantee covers checked ordinary I/O errors and specified process
interruption recovery. It does not claim group atomicity to external readers,
power-loss durability, universal network-filesystem locking, hostile-owner CAS,
cross-volume rollback or resilience to arbitrary corrupted recovery state.

## Boundaries and acceptance work

The driver reads the selected invitation, bounded selected local artifacts,
registered session metadata, pending state and filesystem/access-control
metadata. It reads the selected home configuration and injected clock only at
the stated boundaries. It writes only the selected local plan, located
preparation/record when approved, scoped private session metadata, owned recovery
state and terminal streams. It does not modify the organization source, original
ZIP, unrelated sessions, Git, cloud services or user directories outside these
selected namespaces.

Pure proposal/reconciliation composition and imperative publication remain
separate. Existing codecs, locator, validator, reconciler and walker are used as
real dependencies. No driver calls another driver, no model runs during assembly,
and this decision introduces no preparation-to-question bindings or answer
defaults. State the final dependency/read/write boundary in architecture before
building the component. Sequence authoring-validation contracts/build before
the supported start driver; an ADR selecting this flow is not evidence that
the missing prerequisite already exists.

Independent contracts and whole-binary scenarios must cover: no-write inspection
and queued consent; real reconciled literal records; retained stale preparation;
unbound binding; known missing records and moved results paths; actual alias/case
behavior; private inherited/empty ACLs; two-process exclusion; injected clock
and existing unknown events; record-no-op with unfinished companions; interruption
before/after complete pending publication and between each participant; owner
edits during recovery; completion/cleanup/display failures; and real read-only
walker output without outcome or answer writes. Exception-returning fault tests
are not evidence of killed-process recovery. Test plans are not passing evidence.

## Consequences

Recipients get a real, minimally interactive beginning while existing owner work
remains authoritative and visible. Bounded private recovery increases temporary
PII duplication and implementation/testing cost, but avoids pretending several
files on different volumes can be atomically committed together.

Defer automatic moves, preparation replacement/merge, picker UI, interactive
answers, verification execution, remote synchronization, encrypted fields,
history repair and power-loss durability. Do not advertise those capabilities
as delivered by the import/start slice.
