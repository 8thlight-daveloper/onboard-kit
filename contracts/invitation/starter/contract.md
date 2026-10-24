# Invitation starter contract

Status: complete independently authored contract under0360/0380/0390 and0400;
implementation pending. Runtime format/interaction choices are not inherited
legacy behavior. Independent fixtures and the constructor binding are complete;
this status claims no passing starter implementation or native process-death evidence.
The profile includes terminal.md and storage.md in full.

## API and effects

New takes Runtime from api.go, constructs the actual invitation reviewer, and
returns Command. Nil Runtime functions use the actual system clock/configuration
functions. Non-nil functions replace only those external facts; they do not inject
semantic dependencies, a filesystem backend, fault callbacks, phase hooks or
publication policy. Clock returns exact Timestamp text; the default samples
time.Now once and formats UTC RFC3339Nano. The actual reviewer validates supplied
new-event syntax. Clock is borrowed and called once only after complete accepted
begin/fresh path checks when the actual Proposal needs an event; never otherwise.
Other seams replace os.Getwd, os.Getenv, os.UserHomeDir and os.UserConfigDir.
No callback is closed or mutated. Run does not close borrowed streams.

Run receives arguments after the start command, bounded literal input and terminal
writers. It returns the process exit code. Position6 permits model, own contract,
actual invitation reviewer5, standard library and existing pinned x/sys/unix
filesystem facilities. No driver imports or direct semantic dependency bypass.
Do not call encoders, locator, walker or validators separately from the reviewer.
No commands/checks/URLs/actions, model runs, networking, git, export or answer/
outcome writes occur. Reads/writes are only0360's selected invitation/root/home,
fixed four participants, bounded registrations/recovery data and streams.

## Finite workflow

1. Parse exact fresh/resume CLI; initial help returns without callbacks, input
   or filesystem access. Select exact home precedence and anchor relative paths
   once, with no fallback/tilde/environment interpolation. Safely open existing
   root/home-or-ancestor/archive capabilities, native nonblocking ordered locks,
   bounded registry and pending state; no pre-consent object creation.
2. Fresh entry calls actual Inspect on bounded archive, resolves/checks matching
   registrations and selected physical destination continuity, then displays
   literal-safe identity/privacy/root/destination notice and reads exact confirm
   ACCOUNT or cancel. Identity accepts no publication or defaults.
3. Read bounded selected plan/preparation/record body snapshots through the
   selected safe capabilities; retain all original physical/absence guards.
   Call actual Review. If needed, ask use-incoming or cancel and rerun Review
   with that explicit choice. Show complete actual actions/reconciliation/
   unavailable-comparison/stale-support/privacy/recovery review. Before offering
   begin, freeze purposeful literal-safe terminal session bytes \<=8MiB from the
   actual Review session/support/guidance. A size refusal never reads begin.
   Finish every display write before reading any queued begin bytes.
4. Only complete begin consents. Recheck all selected identities, bytes, security,
   absent names, namespace aliases and registry assumptions. Bootstrap private
   home if absent, reacquire final exclusion in fixed order and recheck again.
   Sample Clock only if needed; call actual Begin and freeze its final artifacts
   and event alongside the already bounded session display. The event is a
   separate frozen scalar, not a reason to rerender the session. The display cap
   is a journal admission check, not a heap proof.
5. Privately publish complete immutable pending payload before any participant
   changes. Fixed role order is local-plan, preparation, session-metadata, record.
   Journal source preimages/final proposals/physical guards; no raw ZIP, typed
   Review/session JSON, unknown-map duplication, transcript or diagnostic PII.
   Every replacement has an owned temporary identity claim before its personal
   bytes/target publication; retain/no-op roles still have exact guards.
6. Verify all intended roles, then publish explicit completion receipt. Record
   equality alone is never completion. Clean personal pending/owned temporaries,
   then nonpersonal claims/owner intent, keeping receipt last. Completed owner
   edits are not a reason to replay old proposals. No rollback of owner changes.
   Leave the verified empty private fixed import-pending container idle; no final
   rmdir and no adoption of unknown children or unclaimed staging.
7. Display frozen actual record-backed activities/instructions/questions/history/
   available support and labeled original invitation guidance, then read-only
   exit loop. It cannot answer, verify, accept defaults or execute instructions.
   Postcompletion display/input/report failure remains completed-with-repair.
8. Resume requires exact token/root/home, no ZIP. Inspect bounded owner/pending/
   claim/receipt shapes with strict keys/base64/lengths/digests before allocation.
   Show frozen actions and exact resume or cancel; receipt means clean or cancel.
   Resume rolls forward only matching guarded preimage/proposal states; never
   calls reviewer or Clock. Clean only removes attributable owned remnants.

## Persistent shapes

HOME/sessions/KEY/ contains plan.obk.yaml and strict bounded session.json.
KEY is the normative exact domain-separated U64BE-length-framed SHA256 of plan ID,
account and selected absolute root spelling. Scan bounded metadata identities/
root capabilities to reuse genuine aliases; the digest is not authentication.
No central index or automatic move/reseed exists. Missing known records/roots,
changed root objects/destinations and ambiguous identities refuse intact.

HOME/import-pending/ contains exactly owner.json, pending.json, publication.json
and complete.json when their phases require them. One transaction per selected
home. Pending has exactly four role entries, exact before/proposed blobs and
guards, final event and frozen sessionDisplayBase64. Base64 is strict canonical;
session display decoded cap8MiB, role blobs combined10MiB, decoded total18MiB,
pending raw32MiB. Metadata2MiB; small owner/claim/receipt16KiB each. No arbitrary
participant list or unbounded phase log. Exact JSON keys/order/ownership and
cleanup formats are in storage.md; exact terminal bytes/input are in terminal.md.
No typed Review JSON or unknown-map recovery codec is introduced.
An existing empty fixed import-pending container is idle only after safe opening
and verified native ownership, mode0700 and ACL absence. It may be reused after
consent; no pre-consent creation or unknown-child adoption is permitted.

## Safety

Adopt preserve-or-refuse native controls, including explicit-empty ACL versus
absence and inherited grants. Private mode0600/0700 alone is insufficient.
Check before writing personal bytes; never strip parent ACLs or modify unrelated
objects. Nofollow capability-relative traversal, bounded regular reads, exact
source/path/security rechecks, no-replace new publication and conservative aliases
apply even to no-op roles. Guarantee ordinary I/O/process-interruption recovery,
not power-loss durability, hostile-owner CAS or cross-volume group atomicity.

The8MiB frozen display replaces the runtime report's earlier typed session JSON
proposal by root direction. Private same-volume per-role staging directories
shield payloads while preserving original final controls. These are normative
concrete0400 choices, not permission to weaken0360 or claim delivered recovery.

## Runtime facts and selection order

New stores Runtime function choices but invokes none. Help/usage invoke none.
For valid mode, explicit --home avoids every home/environment function. Otherwise
query LookupEnvironment(OBK_HOME) once; nonempty wins. If empty, Darwin calls
UserConfigDirectory once and appends obk. Linux queries XDG_STATE_HOME once;
nonempty appends obk, otherwise calls UserHomeDirectory once and appends
.local/state/obk. Other operating systems refuse this native filesystem profile.
Directory helper error/empty result refuses without fallback. Default helpers
are actual OS functions; their internal OS home/configuration lookup is not a
second injected environment resolver.

If any selected archive/root/home path is relative, call WorkingDirectory once
and anchor all relative paths there without chdir. No call is needed when all
are absolute. Reject invalid UTF8, NUL, every literal .. component and an archive
path ending in a separator. Permit . and redundant separators only by safely
traversing every material component nonsymlink before collapsing those spellings.
Do not expand tilde, variables, URL/percent escapes or repair symlinks. Path bounds
apply after anchoring too. Root must already be a directory. Home may be missing;
read-only native traversal obtains its nearest existing ancestor without creating
anything. Actual symlink aliases are not accepted as harmless spelling aliases.

After context selection, acquire nonblocking process-lifetime descriptor locks
on record root and existing home, or nearest existing home ancestor. Deduplicate
equal native objects and acquire in unsigned numeric device/inode order. Busy or
unsupported locks refuse; no retry loop, PID file or expiry. Inspect existing
pending/receipt state before reading a new ZIP. Fresh invocation with recoverable
state reports recovery-required, with its known token; it cannot auto-resume.

After complete fresh begin, create only necessary home bootstrap dirs privately,
release/reacquire the expanded final root/home lock set in the same order, and
recheck ALL original paths/source bytes/security/registry/pending assumptions.
A competing winner can make this attempt stale/busy; do not continue from old
consent. Approved empty bootstrap dirs may remain and must be reported truthfully.
Only after final checks sample Clock if Proposal.NeedsFirstOpenedAt. A Clock
callback is trusted to supply a timestamp, not a publication permission; its
invalid text is passed to the actual reviewer and cannot be silently normalized.
Never call Clock on cancellation, refusal before this boundary, existing record,
resume or cleanup. Recovery never invokes any reviewer operation.

Hold exclusion through complete receipt/ordinary cleanup and release every owned
handle before displaying/waiting in the frozen read-only session or final report/
return. No filesystem operation follows that release in the session loop. Waiting
for exit does not retain root/home locks; the displayed view is intentionally
the frozen accepted snapshot, not a live updater. Handle-close failure remains
the corresponding cleanup repair. Do not close input/output/Runtime functions.
These locks coordinate supported local processes, not hostile-owner CAS or
universal/network-filesystem guarantees. Native two-process tests are required;
a fake mutex or custom filesystem backend is not equivalent evidence.

## Registered context and physical snapshots

Registry metadata is structural bookkeeping, not sender authentication. Validate
every bounded entry's format/key and private storage; reject ambiguity rather
than choose a winner. Retain unrelated valid entries and never prune an entry
because its record is missing. No broad decode/read of other record bodies.

For every exact matching plan ID/account, safely open its stored root and require
its saved native identity to still match. Missing/replaced/relocated old matching
roots refuse, including when another spelling names the old inode. Do not turn
an unknown/moved context into an unseen seed. For a matching selected root object,
require exact stored/current locator-relative record/preparation strings and
physical namespace agreement, reuse its key/root spelling, and set actual known-
record-missing when its registered record is absent. The actual reviewer refuses
that fact; the starter must not hide it or reinterpret it as first use.

Same identity with a VALID distinct root may form a separate context only when
all actual destination/bookkeeping namespaces are disjoint. Complete review
explicitly lists the other matching contexts and explains that this creates an
independent local record/session, leaving their work untouched. Do not clone
their progress or first event, infer a move, or call it evidence of global first
use. This0400 choice follows P2/P4/P8 and the root-key framing, not a change to
record semantics. Invalid old matching contexts still refuse. Different root
objects with nested/custom-path/hard-link aliases of effective participants
refuse, rather than double-registering. Different accounts/case identities cannot
merge when actual filesystem namespaces collide.

Use bounded regular-file reads, extra-byte checks, O_NOFOLLOW-style traversal and
O_NONBLOCK followed by fstat to reject FIFO/socket/device/directory participants
before body reads. Hold directory capabilities; never reopen roles through an
unchecked absolute string. For all actual participants and bookkeeping, reject
same-inode hard links, case-colliding names and incompatible ancestor relations,
including no-op roles, archive input, registered metadata, state files and temps.
Root/home overlap itself is not an error when actual namespaces remain disjoint.

Snapshot exact presence/bytes, object identity and UID/GID/mode/native ACL controls,
and every relevant parent chain. Immediately after begin and before each relevant
publication operation recheck selected names and guards, even retained/no-op roles.
Unexpected bytes/replacement/security or an appearing absent target refuses
further writes while preserving the owner change. Before pending, no participant
changes. After pending, approved earlier roles may already have changed; never
promise rollback or no-files-changed. Existing hard-link anchors outside selected
participants keep their old bytes/inode on pathname replacement.

Only native supported actual filesystem cases establish these guarantees. Native
ACL/no-ACL/empty-ACL/inherited-grant checks, locks, no-replace publication and
closed-resource evidence require platform tests. A faulting writer proves only
its output boundary; an ordinary exception is not killed-process evidence.

## Fixed reports and precedence

Normally accepting stderr receives exactly one terminal JSON object, ordered
status/code/message/transaction and final LF. Strings use storage.md JSON rules.
All messages below are exact. Transaction is empty until known, otherwise the
known64hex recovery token. Never include raw OS/parser text or paths/accounts.

A supplied resume token becomes known after the entire CLI passes syntax
validation. Before any stored transaction has been independently identified,
retain that exact supplied token on a subsequent context refusal, including a
missing or unsafe root, home or recovery namespace. This identifies the caller's
selector; it does not establish that recovery state exists, is authorized or is
complete. Reports derived from an identified stored transaction keep their
existing state/token rules. Syntax-invalid requests report an empty transaction,
even when an earlier argument contained a valid token. Initial help emits no
report. These rules grant no consent and create no recovery state.

| Exit | Status | Code | Message |
| -- | -- | -- | -- |
| 0 | cancelled | start.cancelled | Opening cancelled before begin; no files were changed. |
| 0 | retained | start.recovery-retained | Existing recovery state was retained. |
| 0 | opened | start.opened | Opening completed and the read-only session was displayed. |
| 0 | cleaned | start.cleaned | Completed recovery data was cleaned; no session was displayed. |
| 0 | cleaned | start.staging-cleaned | Owned incomplete staging was cleaned; reopen the invitation to begin. |
| 2 | refused | start.usage | Use one invitation or resume token, an explicit root, and an optional home. |
| 1 | refused | start.source | Read a bounded regular invitation or selected local file. |
| 1 | refused | start.context | Select a safe supported root, state home, and unchanged session context. |
| 1 | refused | start.review | Resolve the displayed invitation review problem before beginning. |
| 1 | refused | start.input | Use complete valid input lines within the displayed limits. |
| 1 | refused | start.output | Required output could not be completed; opening did not begin. |
| 1 | refused | start.limit | The opening or frozen display exceeds the supported limit. |
| 1 | refused | start.commit | Artifacts were not published; approved empty directories or owned staging may need cleanup. |
| 1 | pending | start.recovery-required | An approved opening is pending; resume its transaction before starting another. |
| 1 | opened | start.cleanup-required | Opening is complete; clean its recovery state before starting another. |
| 1 | pending | start.pending | Approved opening is incomplete; retain recovery data and resume this transaction. |
| 1 | pending | start.cleanup-pending | Opening is incomplete and owned recovery cleanup needs repair. |
| 1 | opened | start.cleanup-opened | Opening is complete, but owned recovery cleanup needs repair. |
| 1 | opened | start.display-opened | Opening is complete, but the read-only session could not be fully displayed. |
| 1 | opened | start.input-opened | Opening is complete; the read-only session ended with an input error. |
| 1 | opened | start.limit-opened | Opening is complete; the read-only session reached its input limit. |
| 1 | opened | start.report-opened | Opening is complete, but its final report could not be written. |
| 1 | state-derived | start.recovery-input | Existing recovery state was retained after an input error. |
| 1 | state-derived | start.recovery-limit | Existing recovery state was retained after reaching the input limit. |
| 1 | state-derived | start.recovery-output | Existing recovery state was retained; required output could not be completed. |
| 1 | cleaned | start.report-cleaned | Owned incomplete staging was cleaned, but its final report could not be written. |

Usage before all I/O/facts outranks runtime issues. Safe file/type/body failure is
source; native path/home/alias/lock/registry/guard failure is context; actual
reviewer refusal is review after displaying only its safe diagnostics. Explicit
raw source/output/journal/display limits use limit. Invalid Clock event is actual
reviewer refusal, not a fallback timestamp. Help success alone has empty stderr;
failed help write uses output without invoking facts or filesystem.
Prebegin invalid UTF8, overlong logical input and consumed-input budget exhaustion
use start.input, not start.limit. Completed-session budget exhaustion instead uses
start.limit-opened; recovery-prompt exhaustion uses state-derived start.recovery-limit.

For recovery-prompt input/output failures BEFORE resume/clean consent, use the
recovery-input/limit/output rows, not a fresh no-opening claim. State-derived
means exactly opened for a valid completed receipt, pending for a complete
pending payload without receipt, or retained for attributable incomplete staging.
No deletion or publication follows those failures. Existing completion/pending
truth outranks the fact that this invocation has not yet consented. Malformed or
unattributed protocol state remains preserved/refused with context repair.

After begin, before complete pending, an ordinary staging/encoding/close failure
uses commit and may retain only approved empty dirs/owned attributable staging;
no participant was published. After complete pending but before receipt, failure
uses pending; cleanup/owned-handle repair outranks ordinary pending and uses
cleanup-pending. Preserve approved earlier changes and pending, never rollback.
After verified receipt, cleanup/owned-handle repair outranks later display/input/
report failure and uses cleanup-opened. Otherwise failed session output uses
display-opened, invalid input uses input-opened, exhausted budget uses limit-opened.
Cancelled recovery retains its pre-existing state; any actual cleanup/handle
failure must still report the corresponding pending/completed repair, not success.

On final report failure return1. If completed, attempt exactly one fixed
report-opened fallback with the same known token; no loop or new publication.
If a higher-priority completed cleanup error exists, its fallback preserves that
cleanup status/code rather than masking it. A failing writer may already have
accepted a full JSON object; the exact-one guarantee cannot retract its bytes.
No reporting failure changes completion or samples a second time.
An incomplete-staging-only clean has not completed onboarding: failed final
report uses report-cleaned, not report-opened. Recovery not acted on retains its
prior state; no generic report fallback may assert a nonexistent completion.

## Independent fixture boundary

Four-part fixtures specify literal help/prompts/display/protocol bytes,
injected clocks/config-call counts, real reviewer sessions/reconciliation,
source rechecks, queued-consent nonconsumption and native filesystem cases. No
product implementation is supplied. The independent suite is complete and has
not been executed by its author. Killed-process
scenarios must classify stable reached state independently before SIGKILL; finite
retry failures remain named coverage gaps. Contract completion is not verification
of a starter implementation or of native interruption recovery.

## Plausible wrong implementations

These are proposed rejection targets, not claims that variants have been run.
Fixture provenance is in testdata/sources.md. The contract author does not build
the starter or manufacture expected values through a product dependency.

| Wrong behavior | Independent assertion boundary |
| -- | -- |
| Treat identity/review/partial begin as publication consent | TestIdentityAndReviewAreNotBegin snapshots every owned object; TestExactIdentityAndCompleteReview pins full decision context. |
| Read ahead through queued begin after a failed review write | TestRequiredDisplayFailureDoesNotConsumeQueuedConsent pins reader position for short, zero-error and full-count-error writers. |
| Fill a default, execute a run/check or hide waiting work | TestPurposefulFullSessionAndSupportRemainReadOnly pins selected answers/suggestions, ordered environment/check/support fields, escaped controls and absence of the owned command sentinel. |
| Resample existing/recovery time or canonicalize a no-op record | TestExistingUnknownEventAndFormattingRemainOwned pins literal bytes/native inode; TestPendingRecoveryUsesFrozenBytesWithoutArchiveOrClock forbids clock use without ZIP. |
| Skip actual four-role publication or reorder/drop the record-backed session | TestFreshBeginPublishesLiteralArtifactsAndActualSession pins literal record/preparation/plan, exact metadata and full frozen session bytes. |
| Trust changed original bytes/inode/access/parent or an appearing absent target | TestFreshGuardsRecheckAfterCompleteBegin and TestRecoveryRechecksRetainedRolesAndContextIdentity independently change each family. |
| Guess ownership from names or recursively clean unknown children | TestEmptyIdleReuseAndUnknownStateRefusal preserves unclaimed empty staging and arbitrary children; TestAttributableIncompleteStagingOnlyCleansAfterConsent distinguishes an exact owner claim. |
| Accept noncanonical JSON/base64, unsafe frozen terminal bytes or missing fields | TestStrictPendingProtocolRefusesWithoutRepairing includes duplicate keys, nonzero pad bits, an actually omitted required field and coherently updated outer digests. TestCanonicalBase64DisplayControl admits the valid counterpart. |
| Infer completion from equal record/all-role bytes | TestAllRetainedRolesStillRequireAnExplicitReceipt keeps the transaction pending until a real receipt. |
| Resume completed proposals over later owner edits or recreate lost display | TestCompletedCleanupNeverReplaysAgainstLaterOwnerEdits pins unchanged malformed later bytes; TestReceiptOnlyCleanupBindsNativeRootAndHome isolates native context rather than ordinary role content. |
| Report an incomplete/completed recovery error as fresh cancellation | TestRecoveryDisplayFailureRetainsStateAndDoesNotReadConsent and TestRecoveryInputBudgetUsesPriorState pin fixed state-derived reports. |
| Hide a successful opening after display/input/report failure | TestPostCompletionWriterAndInputFailuresRemainOpened and TestFinalReportFallbackPreservesCompletionAndToken pin authoritative bytes and exactly one accepted fallback. |
| Follow environment fallback after bad explicit configuration or change cwd | TestConfigurationPrecedenceAndRelativeAnchoring and TestBadConfiguredHomeNeverFallsBack count exact runtime facts and compare the entire namespace. |
| Reseed/move registered work or clone another context's first event | TestRegisteredContextDoesNotReseedOrMove and TestDistinctValidRootsRequireExplicitIndependentContextReview preserve old native objects and pin a different independent event. |
| Release a different account's registered namespace when its final files are absent | TestNativeAccountParentAliasRefusesWithBothRegisteredFinalFilesMissing proves equal native parent objects and refuses before consent; TestDistinctNativeAccountParentsPermitAnIndependentPrivateOpening preserves disjoint-account privacy and permits a fresh opening. |
| Escalate modes or rewrite hard-linked original in place | TestExistingPublicationControlsAndHardlinkAnchor pins0400/0600/0644 exact controls, replacement inode and old hard-link bytes. |
| Treat all resource bounds as one generic failure or charge unused input | TestRawSourceBoundsRefuseBeforePublication, TestPendingRawAndDecodedDisplayBounds and TestTotalInputBudgetAndUnusedTrailingBytes distinguish physical sources, input and completed truth. |
| Let an unrelated malformed entry mask the inclusive registry limit | TestRegistryInclusive128ValidEntries uses independently framed valid128/129metadata contexts with existing native roots and disjoint destinations. |
| Guess newly appeared parents or reject another fixed role's genuine shared-parent claim | TestMissingSharedParentsRequireExactCreationClaims contrasts absent, exactly claimed and unclaimed same-named parents. |
| Require initial0600 after final0400 was safely prepared or trust arbitrary intermediate controls | TestClaimedPayloadInitialAndFinalControlsAreFiniteStates pins the claimed inode in initial, final, already-published and unrelated-mode states. |
| Reset another caller-owned group to the current group | TestAlternateOwnedGroupIsPreserved reaches a real alternate permitted group or reports a concrete native fixture gap. |
| Reject or normalize an existing permitted private app group, or inherit that group into new private objects | TestExistingPrivateStatePermittedGroupIsPreserved preserves exact original HOME/idle guards; fresh publication separately pins effective UID/GID on new objects. |
| Treat explicit empty ACL as absence, strip a held ACL or inherit grants through nominal0600/0700 | TestDarwinExistingACLNoopAndReplacement, TestDarwinInheritedGrantCannotBroadenPrivateStagingOrPublication and TestDarwinIdleExplicitEmptyACLIsNotSafeAbsence use actual native presence/payload facts. |
| Use an in-memory mutex or retain native locks/inspection handles during the session | TestActualTwoProcessExclusionAndSessionRelease and TestDarwinOwnedDescriptorsReleasedBeforeSessionInputAndReturn use actual separate processes and unique owned inode descriptor counts. |
| Drop a valid supplied resume token before context verification, retain it from invalid CLI or report it during help | TestValidResumeTokenIsKnownBeforeContextVerification, TestResumeTokenRemainsUnknownUntilTheWholeCLIIsValid and TestInitialHelpDoesNotReportASuppliedResumeToken pin exact reports and untouched inputs/artifacts. |

The optional TestDarwinObservedProcessDeath layer uses test-binary subprocesses
and test-only environment selection, never production hooks. It SIGSTOPs and
independently classifies reached state before SIGKILL. It reports every required
unobserved phase by name. Passing deterministic seeded-state tests does not fill
those native evidence gaps; see testdata/coverage.md.
