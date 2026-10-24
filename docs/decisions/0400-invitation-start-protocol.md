# 0400. Freeze the confirmed invitation opening protocol

Status: accepted by agent judgment; independent starter contract and implementation available

Builds on [0015](0015-functional-core-and-imperative-shell.md),
[0100](0100-behavior-judgment.md),
[0110](0110-reject-ambiguous-plans.md),
[0120](0120-defaults-require-acceptance.md),
[0360](0360-confirmed-invitation-start.md),
[0380](0380-invitation-review.md), and
[0390](0390-invitation-resource-admission.md).

## Context

The pure reviewer can inspect an invitation and return a complete proposal.
It cannot establish consent, publish files, sample a clock, or recover from a
stopped process. The starter needs explicit storage and interaction rules before
independent fixtures and implementation can judge the same behavior.

These are new local protocol choices, not compatibility claims for an existing
stored format. They retain0360's limits: no rollback, group atomicity, hostile
owner protection, power-loss durability, or universal filesystem guarantees.

## Decision

### Keep policy in the reviewer

The starter constructs the actual reviewer. It has no alternate renderer,
validator, reconciler, locator, or first-use policy. Its external fact seams are
clock, environment, working directory, user home and user configuration directory.
Defaults are lazy; help reads none of them. Tests use actual native filesystem
operations, not injected filesystems, semantic dependencies or fault hooks.

Preserve the legacy start route separately. The new routes are explicit
invitation opening and token-based resume, both with an explicit record root
and optional state home. Identity confirmation is not publication consent.
Only a complete begin action after the full review authorizes fresh publication.
Required output must succeed before reading the next action, including queued
input. No default acceptance, answer edit, verification or command execution
occurs in the read-only session.

Sample the clock once only when a new record needs an event, after begin,
fresh source/path checks, state-home bootstrap and final native exclusion.
Pass its exact value to actual Begin. Existing unknown events stay unknown.
Recovery never samples time or calls a reviewer operation.

### Freeze display rather than a second typed model

Before offering begin, serialize the actual reviewed session, support and
original invitation guidance into bounded literal-safe terminal bytes. Preserve
meaningful questions, suggestions, held answers, activities, waiting work,
instructions, escalation and outcome history. Label guidance as invitation-time
content and dates as goals. Never present a suggestion as an accepted answer.

Use a capped builder, not a large allocation followed by a length check.
Freeze at most8MiB. Store the final firstOpenedAt separately after Begin;
it does not require rerendering the prebuilt session. Recovery validates terminal
safety and replays those bytes exactly, without decoding arbitrary typed unknown
maps, reconstructing the view, or retaining a transcript or raw parser details.
Structured diagnostic data is not safe terminal text; display only the reviewer's
fixed repair channel with literal framing.

### Fixed private storage and admission

Register a session under state home using a domain-separated, length-framed
SHA256 key over plan identity, account and registered absolute root spelling.
Metadata also records the root's native identity and exact locator destinations.
The digest is neither authentication nor permission to merge contexts.

One fixed import-pending directory belongs to each state home. Use strict
versioned JSON with exact keys, duplicate-key rejection, canonical base64,
bounded native security guards and exact source digests. No generic participant
framework or unbounded phase log is introduced. The independent contract pins
key order, framing, field sets and all inclusive byte bounds before building.

| Resource | Limit |
| -- | -- |
| Invitation / each selected organizational or record body | 2MiB /1MiB |
| Each session metadata body | 2MiB |
| Registry entries / total raw metadata | 128 /16MiB |
| Path text | 4096UTF8 bytes |
| Four-role preimages and proposals combined | 10MiB decoded |
| Frozen display / combined decoded payload | 8MiB /18MiB |
| Complete pending JSON | 32MiB raw |
| Each owner intent, publication claim or receipt | 16KiB raw |
| Native ACL representation per object | 64KiB decoded |

Preflight raw, decoded, aggregate and claim sizes before personal staging.
Reject excess without truncation. These limits and0390's charged score are not
a process memory or elapsed-time guarantee.

### Claim private objects before writing personal data

Publish a complete immutable pending payload before changing any participant.
It carries exact source preimages and final proposals for four fixed roles:
local plan, preparation, session metadata and record, in that publication order.
Retained roles are guarded too. No byte-equal record shortcut proves completion.

Owner intent claims the exact pending temporary inode and controls before any
personal journal bytes. Each role stages in a private0700 directory on its
actual target parent's volume, with one payload claimed before personal bytes.
Apply preserved final target controls while the payload remains inside that
private directory, then rename. This avoids exposing a0644 replacement before
publication and avoids pretending cross-volume rename is atomic.

Initially private payload controls and the exact intended final controls are
the only permitted security states during preparation. Native ACL absence is
not the same as an explicit empty ACL. Preserve existing controls or refuse;
never strip parent ACLs or change unrelated objects to proceed.

Claim newly created target parents by component ordinal and native identity.
When roles share a missing parent, reuse only its exact earlier transaction
claim. Existing ancestor guards remain in immutable pending. The bounded claim
budget may refuse a very deep destination before personal staging.

An interruption before attribution can leave an empty unclaimed child object.
Preserve and refuse that object. A recognizable filename, UID or mode alone
does not authorize reading, adopting or deleting it.

### Complete explicitly and clean without replay

Verify every final role and required guard, then publish a completion receipt.
Only that all-role boundary establishes completion. Keep the receipt until all
personal pending data, owned payloads and other attributable remnants are gone.
Cleanup is capability-relative and exact; never recursively delete a namespace.

A completed receipt means cleanup-only. Later ordinary edits to role files do
not authorize old proposals to overwrite them. Receipt-only cleanup compares
the selected root/home native identities and owned cleanup namespace, not later
role contents, metadata or ordinary safe root/home permission changes.

If pending display bytes were already removed, successful cleanup reports that
no session was displayed. Do not rerun review, reconstruct lost display bytes,
sample a second event, or call a completed opening cancelled.

Leave the fixed import-pending directory as an empty private idle container
after deleting the receipt. A safely opened empty container may be reused;
do not remove it as the final cleanup step. This removes the normal process-death
window between final receipt deletion and rmdir. It retains no personal backup.
The idle exception applies only to the empty fixed container with verified native
ownership,0700 controls and no ACL, not unknown children or unclaimed staging.
No container is created before begin; unexpected or unsafe state stays intact.

### Distinguish contexts rather than infer moves

For an exact matching account and plan, old registered roots must remain valid.
A missing, replaced or relocated old root refuses rather than reseeding. A safe
spelling alias of the same native root reuses its existing registration only
when destinations and namespaces agree.

A valid distinct root may create a separately reviewed context only when all
actual artifact and bookkeeping namespaces are disjoint. Review names the other
matching contexts and explains that the new record/session is independent.
Never clone progress, infer a move, copy the older event, or claim global first
use. Nested/custom-path/hard-link destination collisions still refuse.

## Rationale and verification

P2 requires deliberate identity, replacement and begin choices. P3 requires
truthful pending/completed reports and visible waiting/support context. P4
preserves owner changes and makes recovery exact rather than a new merge.
P7 favors bounded temporary personal copies, explicit disclosure and no second
unknown-data codec. P8 favors a sufficient frozen display and an empty reusable
container rather than an avoidable manual cleanup question. Privacy and native
ownership constrain that convenience; unknown objects remain untouched.

These decisions are agent judgments, not human approval or implementation
evidence. Independent fixtures must pin exact terminal bytes, protocol fields,
input limits, partial-write handling, callback counts and source rechecks.
Actual platform tests must exercise nofollow traversal, file types, aliases,
ACL absence/empty/inheritance, access preservation, two-process native exclusion,
no-replace publication and handle closure.

Killed-process tests must independently observe a reached persistent state,
then kill and resume without the ZIP. Cover before/after pending publication,
each role, no-op record with unfinished companions, receipt-only cleanup and
the empty idle container. Exceptions or production fault flags do not substitute
for process-death evidence. Unobserved phases and unavailable native platforms
remain named coverage gaps, not passing claims.

The starter contract's four parts, independent implementation, binary wiring
and release verification remain separate work. Catalog selection, stable
repeated templates and record editing remain in the release roadmap.
