# 0340. Guided recipient preparation and reviewed save

Status: accepted by agent judgment; contract authored; implementation pending

Builds on [0140](0140-person-specific-preparation.md),
[0150](0150-onboarding-library-and-preparation-flow.md),
[0160](0160-personal-data-and-preparation-approval.md),
[0190](0190-incremental-preparation-repair.md),
[0210](0210-person-specific-preparation-paths.md),
[0220](0220-preparation-command.md), and
[0330](0330-preparation-review.md).

## Context

The preparation commands save one explicit operation at a time. The pure
reviewer now separates validity, organizational blockers and the actual guide,
but no interaction lets an onboarder propose several recipient changes and
review them before writing. Replaying write commands after questioning would
create several commit points, not one reviewed save.

A new full-screen framework is unnecessary for that first useful interaction.
A plain text conversation can keep existing facts, proposed values and save
consent visible, while supporting ordinary keyboards, screen readers and
independent scripted acceptance scenarios.

## Decision

Add this explicitly bounded operation:

```text
onboard-kit prepare guide --plan FILE --account ID [--root DIR]
```

The first increment stages recipient details only. Preserve every arrangement
and unknown preparation value; do not offer nonworking arrangement controls or
claim the complete invitation flow. Existing explicit preparation commands
remain available for arrangements. Later guided arrangement editing needs a
separate contract increment, not a general UI framework now.

Use a separate preparation guide driver with explicit arguments, input reader,
stdout and stderr. Its New constructor takes no dependencies. Root dispatch
recognizes exactly prepare guide and forwards the streams it already receives.
Existing prepare and invite command interfaces and behavior remain unchanged.
Help returns without reading input or inspecting files. Never obtain input
from a hidden global terminal or invoke another driver to stage changes.

### Stage, review, save

Read the explicit original plan and one account's preparation through the real
decoders, validator and locator. No record is needed or read. The root must
already exist. An absent preparation starts as an identity-only in-memory
proposal; loading creates no file, directory or temporary object. Wrong schema,
kind, account, plan or ambiguous arrangement identity refuses editing. Do not
reset invalid identity or choose a duplicate winner.

Keep the loaded value, exact source-byte snapshots, accepted draft, local
recipient form and reviewed encoded bytes distinct. The form starts from the
whole held recipient. Explicit keep, change and clear choices preserve omitted
fields and deliberately clear selected ones. Menu Enter chooses nothing.
Ordinary words such as save or q are literal text at a value prompt, not global
commands. Keep existing multiline text exactly; this first editor creates
single-line text only. Directory choices use declared exact IDs, never the
first item, inferred role or guessed account.

Allow several local field changes before applying the whole intended recipient
through the actual editor. A refused form changes neither accepted draft nor
disk, and can be corrected or discarded. The existing editor validates the whole
replacement recipient, not sparse changes. Preserve unrelated invalid
arrangements under its scoped repair rules. No defaults, current date, personal
delivery address, actual-start question, discipline field or answer binding.

Review displays the current destination and complete relevant old/new recipient
changes, intentional clears, and the actual reviewer's four results. If
preparation failed, label downstream results not evaluated, not no blockers.
Otherwise retain blockers beside a permitted guide and guide problems beside
blockers. A preview is not export readiness. Missing optional facts remain
optional; dates remain planning goals. Do not infer required arrangements for
activities from catalog names or prose.

Show the actual first-steps text as safely quoted document content, without
changing the reviewer's original string or using a second renderer. The first
recipient-only slice supplies no general-help selection rather than guessing
one. Its guide retains honest missing-help wording. Export still requires its
own explicit help selection and pre-access assertion.

For a changed draft or absent preparation, encode the accepted draft with the
actual preparation encoder and enforce the existing 1 MiB bound before offering
save. An unchanged existing value instead freezes the original bounded bytes
for a no-op, with no regeneration or encoder-validity gate. A decoder-accepted
empty arrangement can be unencodable; merely reviewing it must still permit
the byte-preserving no-op. Both branches recheck original bytes and paths.
Displaying review must
succeed before reading save consent. Recognize an explicit save action only
at that current Review prompt; empty input never saves. Any accepted edit
invalidates prior review. Do not regenerate different bytes after consent.

Cancel and EOF before a complete save action discard memory only and return
an honest no-files-changed result. An unterminated partial input line at EOF
cannot be a save action. Input errors or failed display return failure without
publication. No hidden autosave, transcript, recovery file or approval marker.

### Publish exactly the reviewed preparation

On save, recheck the plan and selected preparation against the loaded bounded
bytes and safe path/identity assumptions. Refuse an observed change or a
formerly absent destination that appeared; require a fresh load and review.
This is a cooperative-editor guard, not a universal compare-and-swap or hostile
same-owner race guarantee. Other onboarders still coordinate edits normally.

Use rooted nonsymlink filesystem capabilities and regular bounded inputs.
Refuse an output alias of the selected plan. Do not truncate an existing inode,
follow a symlink escape, write outside the selected root or publish oversized
data. After consent, missing preparation directories may be created privately;
preflight and cancellation create none. An I/O failure during the commit attempt
may leave newly created empty directories, which must not be described as an
unchanged filesystem.

Write private temporary bytes, check write and close, then atomically publish
the reviewed encoding. For absent preparation, use no-replace publication.
For existing preparation, replace only its selected pathname, not other hard
links. A semantically unchanged existing preparation is a no-op: preserve its
original bytes rather than reformatting it just because the guide was opened.
An absent identity-only draft may be explicitly reviewed and created.

An accepted scoped repair may save while unrelated stale arrangements remain
invalid, provided the existing encoder can represent the retained data. Report
the remaining problems honestly. Do not add a whole-file-validity or export
readiness gate to saving, or weaken encoder representability rules.

Check owned temporary cleanup on every path. If cleanup fails before commit,
report refusal with repair needed and preserve the destination. If it fails
after commit, retain the saved preparation and report saved with cleanup repair
needed. A secondary output failure cannot turn a committed save into a claim
that nothing changed. An already absent owned temporary counts as cleaned.
Do not strip access controls, change unrelated permissions, delete the committed
file or retry indefinitely. Contracts must pin private codes and reporting
precedence; [0310](0310-invitation-cleanup-failures.md)'s invitation-specific
codes are not silently reused as an existing guide guarantee.

Save exits after one commit or no-op. Export remains a separate explicit command.
Do not generate an archive, seed, outcome, first-use event, index or configuration.
No messages are sent, accounts provisioned or commands executed.

### Accessible and literal interaction

Use line-oriented text menus with stable labels, visible choices and ordinary
Enter-terminated input. No raw terminal mode, screen clearing, cursor movement,
mouse, terminal-size query, color-dependent meaning or minimum-window lockout.
Labels say draft not saved until publication. There is no readiness badge,
completion percentage or overdue failure.

Bound each logical UTF-8 input line to 64 KiB excluding LF or CRLF framing.
Preserve other whitespace; reject oversized or invalid input without echoing it.
Quote displayed scalar values and terminal controls literally, and prefix guide
lines as document content so an authored prompt cannot become a live control.
The contract must pin exact presentation with independent examples. This changes
display only, never held values or saved encoding.

Explain ownership, shared-file readership and terminal scrollback. Intentional
display of selected recipient facts supports editing; diagnostics and status
summaries must not repeat rejected personal input, parser excerpts, raw OS errors
or unknown values. Do not imply private modes conceal shared data or that free
text is automatically scrubbed of secrets or PII.

### Components and access

Preparation guide is an imperative driver at position 7. It may depend on its
own contract, model and actual plan decoder/validator, preparation decoder,
encoder, locator, editor and reviewer at positions below 7. It may not call other
drivers, use the legacy autosaving TUI or copy a core renderer. Binary wiring
is at position 8 for this increment, superseding 0020's original position 7 so
the entry point remains after the components it wires.

Reads: explicit arguments and interaction, bounded selected plan/preparation
bytes and safe path metadata, actual dependency results and supplied writer
outcomes. Writes: stdout/stderr and, only after reviewed save, the selected
preparation plus private required directories and owned temporary publication
resources. No plan or record writes, record reads, environment-based identity,
clock, network, process execution or outside state. The default root is current
filesystem context, not acceptance of a question default.

## Verification and delivery

The [four-part driver contract](../../contracts/preparation/guide/contract.md)
specifies the recipient-only interaction. Independent binary data scenarios
and a separately authored implementation follow. Pin exact prompt grammar, input limits,
literal display, cancellation, save consent, scoped preservation and truthful
commit/cleanup reports. Use frozen literal payloads, not the product encoder
to calculate expected saved bytes. Test real dispatch separately.

Include absent and existing preparation, keep versus clear, two recipient
references repaired together, unrelated stale arrangement retained, wrong
identity and duplicate IDs, literal menu-looking text, control characters,
unknown/malformed records ignored, failed review display before queued save,
stale review bytes, concurrently appearing destinations, symlinks, no-op byte
preservation, and saved-but-not-reported outcomes. Filesystem fault and race
claims need actual reaching fixtures, not a substitute filesystem hook.

This is a bounded direction under P1/P2/P3/P4/P6/P7/P8 of
[0100](0100-behavior-judgment.md), not completed verification. Guided arrangements,
activity selection, repeated resources, preparation-answer bindings, import and
record-backed onboarding remain separate work. No saved format changes here.
