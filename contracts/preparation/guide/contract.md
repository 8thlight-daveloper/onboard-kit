# Preparation guide contract

Component: `preparation/guide`. Agent-authored, not human-reviewed.

## Boundary

The [API](api.go) receives arguments after the exact `prepare guide` pair,
an input reader and two writers. `New()` takes no dependencies and constructs
the actual plan decoder and validator, preparation decoder, encoder, locator,
editor and reviewer. This position-7 driver may use their contracts, model and
standard-library filesystem facilities and the existing pinned
`golang.org/x/sys/unix` filesystem facilities, not another driver or the legacy
TUI. Unix facilities are permitted, not a required publication algorithm; this
does not add a dependency or change the pure-core dependency rules.
The source decisions are [0340](../../../docs/decisions/0340-guided-preparation-save.md)
and its arrangement extension [0350](../../../docs/decisions/0350-guided-support-arrangements.md).

Reads: explicit arguments and interaction, the selected regular plan and
preparation, safe filesystem metadata and actual dependency results. Writes:
stdout/stderr and, only after current reviewed save consent, the selected
preparation, private required directories and owned publication temporary files.
Do not read the record, environment, clock or network, execute processes, save
answers, generate invitations, or touch any other file.

Arguments require exactly one `--plan FILE` and `--account ID`; `--root DIR`
is optional once, otherwise `.`. Exact separate-token options and `--name=value`
forms are accepted in any order, including mixed forms. Values are nonempty.
Separate-token values may not begin `--`; equals-form values may. Ordinary
single-hyphen values are values, not options. Reject positional arguments,
unknown/repeated options and missing values. Initial `--help` or `-h` wins over
following arguments and returns the exact help fixture with exit 0, empty stderr,
no input read and no filesystem inspection. Help elsewhere is ordinary invalid
syntax. Invalid arguments return usage without input or filesystem access.

The exact help text is the following four lines, including the empty second
line and one final LF. The help fixture contains these same bytes:

```text
Usage: onboard-kit prepare guide --plan FILE --account ID [--root DIR]

Stage recipient details and support arrangements, review, and explicitly save preparation.
The plan and onboardee record are not changed. Export is a separate command.
```

Read bounded regular UTF-8 plan/preparation through actual decoders. A source
may contain at most 1,048,576 bytes, inclusive; never read an unbounded special
file. Require an existing real nonsymlink root and nonsymlink path traversal;
selected source files may not be symlinks. Plan validation errors refuse; warnings
alone do not. Derive expected account from arguments and plan identity from the
validated original plan. A missing or empty original plan ID refuses with
guide.plan; do not create person-specific identity with an absent plan ID.
Resolve the preparation through the actual locator's
resultsPath/account rule. Do not open the resulting record path.

Reject NUL and every literal `..` path component in explicit plan/root arguments,
even when cleaning would appear to keep the path inside a root. Reject a plan
argument ending in a separator. Do not silently clean traversal or resolve a
symlink to make an unsafe selected path acceptable. An absolute plan path is
allowed, with nonsymlink traversal; root remains the separate explicit existing
directory. Tests resolve the host's temporary-directory symlink before selecting
their sandbox, rather than weakening source-path rules for macOS `/var`.

Absent preparation becomes an in-memory schemaVersion 1, kind preparation,
explicit account and original planId. Existing preparation must decode. Apply
the actual editor's identity/ambiguity guards by probing SetRecipient with an
empty recipient, discard the probe result and retain the original value. This
must not erase held fields or reject ordinary stale recipient references before
they can be repaired. Wrong identity and repeated arrangement IDs refuse.
Populate editor context with all declared exact contact, team and section IDs,
including inactive sections, without inferring an answer or role.

## Interaction grammar

All displayed lines end LF. No ANSI sequences, raw-terminal operations, screen
clearing, implicit menu selections or automatic saves. Write this startup block:

```text
Preparation draft — not saved.
Onboarders maintain preparation; onboardees own their record.
Shared-file readers can read these details. Terminal scrollback can retain them.
Do not enter secrets or unnecessary personal information.
Interaction limit: 1048576 input bytes per run, including line endings. No partial save.
```

Then display `Account: Q(account)` and `Destination: Q(relative preparation path)`.
Here Q is the literal quoting rule below, not the text `Q(...)`.
At each Draft prompt emit exactly:

```text
Draft — not saved. Choose recipient, arrangements, review, or cancel:
```

Only exact named actions are recognized. Invalid or empty actions print
`Choose one of the displayed actions.` and repeat the current prompt. `cancel`
is recognized only at Draft and Review, not a global shortcut. EOF before a
complete save action cancels with no disk changes; partial final input is never
an action or value. Input errors return input failure, not cancellation.

`recipient` copies the whole accepted recipient to a local form. Emit the nine
field lines below in order with the currently held form value quoted, followed
by the Form prompt. Redisplay these lines whenever returning to Form:

```text
1. Name: Q(value)
2. Company email: Q(value)
3. Planned start date (goal): Q(value)
4. Team ID: Q(value)
5. Manager contact ID: Q(value)
6. Onboarding owner contact ID: Q(value)
7. Buddy contact ID: Q(value)
8. Work location: Q(value)
9. Time zone: Q(value)
Recipient form — not saved. Choose 1-9, apply, or back:
```

`back` discards the entire local form and returns to Draft. A field number emits
`Field LABEL: Q(value)` using the corresponding label above without its number,
without its trailing colon, then `Choose keep, change, clear, or back:`.
`keep` and `back` return to Form
unchanged; `clear` deliberately replaces that field with empty text and returns
to Form. For team field 4, `change` first prints `Available teams:` followed by
each declared team in original order as `ID: Q(id) Name: Q(name)`; if empty print
`No declared teams.`. For contact fields 5, 6 and 7, print `Available contacts:`
and the corresponding declared contact lines, or `No declared contacts.`.
These lists do not select the first entry or normalize an ID. Other fields do
not print an unrelated directory. Then `change` emits
`Enter the new single-line value (empty input retries):`.
Consume the next complete nonempty input line literally, including leading and
trailing spaces and words such as save, q, clear or cancel. Empty input repeats
that value prompt, not a clear or default. Keep existing multiline values exactly;
this operation cannot create new multiline values. No field-level normalization,
default selection or ID inference. The whole intended form can repair two stale
references together.

`apply` passes the whole form to actual SetRecipient. Success replaces accepted
draft and returns to Draft. Any problems keep the form available, preserve the
accepted draft and disk, print `Recipient not applied.` followed by safely
quoted severity/code/path triples, then redisplay Form. Problems are not terminal statuses.
They never echo dependency message/detail values or rejected personal input.

### Arrangement menu and exact selection

`arrangements` opens the held draft's arrangement list. Every time the
Arrangements menu is entered or repeated, print `Arrangements — draft not saved.`
then, in current list order, one summary per item:
`N. ID: Q(id) Section: Q(section) Kind: Q(kind) Status: Q(status)`.
Number N is one-based decimal. An empty list prints `No arrangements.`. Then:

```text
Arrangements — not saved. Choose new, edit, remove, or back:
```

`back` returns to Draft. Neither list positions nor new generated values become
saved identities. `edit` and `remove` print `Choose arrangement number or back:`.
Only a canonical ASCII decimal ordinal in 1..N selects the corresponding held
item; no leading zeroes, sign, whitespace, truncation, overflow, ID matching,
first-item fallback or empty-line selection. `back` returns to Arrangements.
Invalid input uses the existing fixed invalid-action line and repeats that
selection prompt. Selecting from an empty list offers only back.

### Whole arrangement form

`new` opens a zero arrangement in a local form. `edit` opens a copy of the whole
selected item. Its permanent ID is immutable in edit mode, including a malformed
or empty held ID. Show `Permanent ID: Q(id) [new]` or `[fixed]` respectively,
then these eight fields in order:

```text
1. Section ID: Q(value)
2. Kind: Q(value)
3. Accountable owner contact ID: Q(value)
4. Assigned contacts:
5. Reference: Q(value)
6. Target date (goal): Q(value)
7. Preparation status: Q(value)
8. Note: Q(value)
```

After the `4. Assigned contacts:` heading, show each ordered held helper as
`   N. Q(id)`, or `   None.`. Show all other values, including empty strings.
Redisplay the complete form whenever returning to it. New-mode prompt:

```text
Arrangement form — not saved. Choose id, 1-8, apply, or back:
```

Edit-mode prompt omits the id action:

```text
Arrangement form — not saved. Choose 1-8, apply, or back:
```

Only new mode recognizes `id`. This opens `Field Permanent ID: Q(id)` and the
same keep/change/clear/back scalar interaction as recipient fields. Fields
1, 2, 3, 5, 6, 7 and 8 use their displayed label with the same scalar interaction.
Field 4 opens the helper-list form below. Menu Enter chooses nothing. Scalar
change consumes literal nonempty text, never a global command, normalized
value or default. Keep existing multiline text exactly; newly entered text is
single-line. `back` discards the entire enclosing arrangement form and returns
to Arrangements, including all locally applied helper changes.

Before a section change value prompt, print `Available sections:` and each
declared section in original order as `ID: Q(id) Title: Q(title)`, including
inactive declarations. An empty list prints `No declared sections.`. Before
an owner change, print the existing Available contacts listing. Before a kind
change print `Supported kinds:` then each line in this order:

```text
support
equipment
office
remote
access
meeting
reading
pairing
ticket
pr-review
pr-post
pr-merge
deploy-shadow
deploy-lead
```

Before a status change print `Supported preparation statuses:` then:

```text
pending
requested
arranged
deferred
```

These are available spellings, not inferred selections. Never infer a kind,
status, owner, helper, reference or date from a title, activity, another value,
catalog text or time. Target dates remain goals. Preparation arranged never
records an activity outcome or proves access.

On `apply` in new mode, first compare the proposed ID exactly with every accepted
draft arrangement ID. If one matches, do not upsert that item. Print
`Arrangement not applied.` and the fixed line
`Severity: "error" Code: "guide.arrangement-exists" Path: "arrangements"`,
preserve the local form and accepted draft, and redisplay the form. This is a
driver-local new-mode guard; it does not change the editor's upsert contract.
Otherwise call actual SetArrangement with the complete form. Edit mode always
keeps the selected exact permanent ID, calls actual SetArrangement and replaces
that item in place; it does not append or rename an existing identity.
Any editor problems print `Arrangement not applied.` followed by the unchanged
ordered safely quoted severity/code/path projection, preserve all local proposed
corrections, and redisplay the form. No partial apply or disk write. Success
replaces the accepted draft and returns to Arrangements. Scoped editor rules
preserve unrelated invalid recipient facts and other arrangements; review can
still report them. No whole-file validity, delivery readiness or outcome gate.

### Ordered helper-list form

Copy the enclosing arrangement form's whole Contacts list to a nested local
list, preserving its held absent/empty shape on unchanged apply or back. Merely
opening or accepting an unchanged helper form is not an edit that forces a
rewrite. Removing the final helper, like explicit clear, sets the list absent.
On each entry or repeat, print `Assigned contacts — local list, not applied.`
then each entry in held order as `N. Q(id)`, or `No assigned contacts.`. Prompt:

```text
Helpers — not saved. Choose add, remove, clear, apply, or back:
```

`add` displays Available contacts, then
`Enter the assigned contact ID (empty input retries):`. Append one complete
nonempty literal line exactly as entered. Words such as save, remove, clear and
back are values here, not actions; do not trim, deduplicate or infer a contact.
References are validated only when the enclosing arrangement is applied.
`remove` prints `Choose helper number or back:` and accepts the same canonical
ordinal grammar, removing only that one occurrence. Equal helpers elsewhere
remain in order. Empty/invalid input never means the first or an empty helper.
`back` from the ordinal selection returns to Helpers without removing anything.
`clear` sets only the nested list to absent. Each add/remove/clear returns to
Helpers. `apply` copies the entire proposed list into the local arrangement
form and returns to that form; it does not apply the arrangement or save.
`back` from Helpers discards nested changes, preserving the enclosing form's
previous helper list. Back from the arrangement form later discards both its
scalar and applied helper changes. No menu action has a hidden autosave.

### Staged removal

After explicit ordinal selection for remove, print `Selected arrangement for removal:`
then the complete known-field block defined below, including its selected
ordinal, even for a unique malformed or empty ID. Then:

```text
Removal is staged, not saved. Choose remove or back:
```

`back` returns to Arrangements unchanged. `remove` calls actual RemoveArrangement
with the exact selected ID; unique malformed/empty IDs are valid explicit
removal targets. On success replace the accepted draft, preserve remaining order
and return to Arrangements. On failure print `Arrangement not removed.` and
the ordered safely quoted severity/code/path projection, preserve accepted draft
and return to Arrangements. No missing input chooses an empty ID. Startup still
refuses duplicate IDs before this menu can select a winner. Removal is a staged
preparation edit, not deletion of an activity, contact, answer, outcome or file.

`review` calls the actual reviewer with explicit account, original plan,
accepted draft and empty GeneralHelp. Never infer or select help. Display:

```text
Review — not saved. Preview is not export readiness.
```

Follow with Account and Destination lines as at startup, then all nine field
change lines in the same order: `LABEL: Q(original) -> Q(draft)` plus
` [cleared]` only when nonempty original becomes empty. For absent preparation,
original recipient is empty. Always show unchanged lines too. Then display the
complete arrangement lists defined below before printing reviewer results. Print
`Preparation problems:` then either `None.` or each result's safely quoted
`Severity: Q(severity) Code: Q(code) Path: Q(path)` in original order. With any preparation problems,
print `Onboarder blockers: Not evaluated.` and `First steps: Not evaluated.`;
do not claim empty blockers. Otherwise print `Onboarder blockers:` then `None.`
or the actual blockers' safely quoted Question as
`Question: Q(question)` in order; Blocker has no section field. Print `First-steps problems:`
then `None.` or severity/code/path triples in order, followed by `First-steps document:`.
Split the actual unchanged guide on LF and emit every resulting logical line
as `| Q(line)`; omit only the synthetic empty piece after a terminal LF.
If the returned guide is empty, print `No document returned.` instead of a
quoted blank line. This is quoted actual document content, not a second rendering or new model.

### Complete arrangement review

Print `Original arrangements:` followed by every loaded arrangement in original
order, then `Draft arrangements:` followed by every accepted proposed arrangement
in draft order. Each empty list prints `None.`. A known-field block is exactly:

```text
Arrangement N:
ID: Q(id)
Section ID: Q(section)
Kind: Q(kind)
Accountable owner contact ID: Q(owner)
Assigned contacts:
Reference: Q(reference)
Target date (goal): Q(targetDate)
Preparation status: Q(status)
Note: Q(note)
```

Directly after Assigned contacts, show each ordered helper as `N. Q(id)`, or
`None.`. Use all fields even when absent: `""` means absent scalar and None.
means absent list. Do not deduplicate, sort, hide unchanged fields or omit removed
items from Original. Quote multiline/unsafe text with Q on one display line.
Unknown preparation values are preserved but never printed. The same complete
block is used by removal confirmation; its arrangement ordinal is the explicitly
selected ordinal from the held list. Ordinals are display/selection only.

For an existing semantically unchanged preparation, freeze original bounded
bytes for no-op, without requiring successful encoding, even with an empty
arrangement. Otherwise encode the accepted draft through the actual encoder;
any problems or result over 1 MiB refuse the review/save attempt with encode
failure. Preserve unknown root data and all arrangements in order. Full
preparation validity is not an extra save gate; a representable scoped repair
may retain unrelated stale arrangements. No write or temporary creation occurs
while constructing or displaying review.

After the full current review and all required output writes succeed, emit:

```text
Review complete. Choose save, back, or cancel:
```

Only at this prompt does exact `save` consent to this displayed snapshot.
`back` returns to Draft; a later accepted edit invalidates the old review.
Never consume queued save input after a failed review display. Consumption
includes reading following save bytes into an input buffer, not only
interpreting them; failed display leaves the passed reader immediately after
the previous completed review action.
A completed
`save\n` followed by EOF is consent; an unterminated `save` followed by EOF is not.
After one successful save or no-op, exit. Do not export or reopen menus.

### Input and literal display

Accept valid UTF-8 logical lines of at most 65,536 bytes excluding one final LF
or CRLF. Preserve all other whitespace. Reject invalid UTF-8 and excess bytes
without echoing them. Do not accept an unterminated final line as an action.
The same UTF-8 and logical-payload size checks apply to an incomplete line at
EOF. Invalid UTF-8, including an incomplete code point, or excess logical
payload returns guide.input before ordinary EOF cancellation. Only a complete
CRLF pair is framing: a terminal CR without LF is part of the payload. Thus
65,536 payload bytes followed by CRLF are valid, 65,536 payload bytes followed
by bare CR and EOF exceed the bound, and 65,535 payload bytes followed by bare
CR and EOF are a valid incomplete line. Below the total budget, a valid
incomplete line still cancels without becoming an action or value.

Also bound all consumed input to 1,048,576 bytes per Run, counting every action,
value and LF/CRLF byte returned by the passed reader. Before every further read,
if the budget is exhausted, return guide.input without another reader call or
publication. Never prefetch beyond the current complete logical line or consume
more than the remaining budget; input buffering is consumption too. A complete
save LF/CRLF wholly inside the budget remains valid, including when its final
framing byte exactly consumes the budget; do not make a further read to test EOF.
A partial line at exhaustion is not an action, value or consent and returns
guide.input. Refusal takes precedence over unread EOF or queued cancel, even
when the last consumed line was complete. EOF below the budget still cancels;
a reader may return its final byte together with io.EOF: a complete save line
still consents, while a partial line at the exhausted budget still refuses
with guide.input instead of cancelling.
a complete cancel within budget is still cancellation. Once save/cancel/help/
usage exits, unconsumed trailing bytes are not charged or inspected. Do not
truncate, normalize or partially apply a value to fit a budget. Input refusal
preserves original preparation bytes with no hidden partial save.

Q encloses text in ASCII double quotes. Escape double quote and backslash as
`\"` and `\\`, LF/CR/tab as `\n`, `\r`, `\t`; all remaining C0/C1 controls,
Unicode format characters (including bidi controls and zero-width joiner),
U+2028 and U+2029 become lowercase `\u` escapes with four hex digits, or
surrogate pairs for a supplementary code point. Ordinary printable Unicode
remains literal, with no HTML escaping. Thus `Ünal` displays `"Ünal"`, ESC
displays `"\u001b"`, U+202E displays `"\u202e"`, U+200D displays `"\u200d"`,
and a two-line held name displays `"Line one\nLine two"` on one terminal line.
This rule changes display only, never held input or saved encoding. Unknown
values, raw parser/OS messages and dependency diagnostic messages are not shown.

## Publication and reporting

On save, recheck original plan and preparation bytes and safe path/identity
assumptions; any observed change or appearing formerly absent destination
refuses with stale-review. Recheck unchanged existing preparation too. This is
a cooperative-editor guard, not universal hostile same-owner compare-and-swap.
Refuse a preparation output alias of the plan, including an existing hard link.
No source truncation or following symlinks. Existing hard links to preparation
retain their old inode bytes when its selected pathname is replaced.

Only after consent may missing preparation directories be created with private
0700 mode. Use a private 0600 regular temporary, check every write and close,
then atomically publish the frozen reviewed bytes: no-replace publication for
new preparation, pathname replacement for existing preparation. No-op preserves
original bytes, inode and modes. Do not regenerate bytes after consent. Observe
existing restrictive permissions and never strip ACLs or chmod unrelated data.
Mode bits alone do not establish private ACL access. Owned temporaries, new
directories and publication must not inherit grants that weaken required privacy
or relax an existing preparation's ACL absence. Establish safe access controls
before writing reviewed bytes; preserve parent ACLs or refuse rather than
stripping inherited or parent access controls. No specific syscall or publication
algorithm is required by this rule.
If existing access controls cannot be preserved safely, refuse with guide.commit
and preserve the original preparation; an unchanged no-op does not require the
ability to copy or replace its ACL.
Commit-attempt failure may leave empty newly created directories; report that
possibility instead of claiming unchanged filesystem.

Check owned temporary cleanup after success and failure. Already absent means
cleaned. Before commit, cleanup failure preserves destination and reports
cleanup-refused; after commit, retain saved preparation and report cleanup-saved.
Release every owned inspection and publication file/directory descriptor before
Run returns on every path. Do not close the caller's borrowed input or writers.
Never delete committed preparation, retry indefinitely or relax access controls.
Cleanup failure takes precedence over a secondary output failure. This contract
does not promise crash durability, universal races, recoverable multi-file
transactions, or fault injection hooks.

For a normally accepting stderr writer, terminal stderr is exactly one
LF-terminated JSON object with exactly the keys
status, code and message; key ordering is not significant. Fixed values follow.
No dynamic account, path, input, unknown data, parser excerpt or OS error appears
in terminal diagnostics. Help is the sole success with empty stderr. All other
terminal outcomes use the table, including errors before startup:

| Exit | status | code | message |
| -- | -- | -- | -- |
| 0 | cancelled | guide.cancelled | Cancelled. No files changed. |
| 0 | saved | guide.saved | Preparation saved. |
| 0 | unchanged | guide.unchanged | Preparation unchanged. Original bytes preserved. |
| 2 | refused | guide.usage | Use prepare guide --plan FILE --account ID [--root DIR]. |
| 1 | refused | guide.source | Cannot safely load the selected plan or preparation. |
| 1 | refused | guide.plan | The selected plan is invalid. |
| 1 | refused | guide.identity | Preparation identity or arrangement IDs are invalid. |
| 1 | refused | guide.path | Cannot safely select the preparation destination. |
| 1 | refused | guide.input | Cannot read a complete supported input line. |
| 1 | refused | guide.output | Cannot display the current draft or review. No files changed. |
| 1 | refused | guide.encode | Cannot encode the reviewed preparation within the size limit. |
| 1 | refused | guide.stale-review | Source files changed. Reload and review before saving. |
| 1 | refused | guide.commit | Preparation was not saved. A commit attempt may have left empty preparation directories. |
| 1 | refused | guide.cleanup-refused | Preparation was not saved. Owned temporary cleanup needs repair; empty preparation directories may remain. |
| 1 | saved | guide.cleanup-saved | Preparation saved. Owned temporary cleanup needs repair. |
| 1 | saved | guide.report-saved | Preparation saved, but its completion report could not be written. |

Dependency decode errors use source; warnings alone do not refuse. Locator,
root, traversal and output-alias selection failures use path. Once a regular
preparation source has been safely selected, bounded body-read failures,
including oversize and invalid UTF-8, use source, not path. A derived filename
does not make its body-read failure a path-selection failure. Locator
problems use path;
editor startup guard problems use identity; plan validation errors use plan.
Absent files are permitted only for the selected preparation. Other unsafe
metadata/open/read failures use source before destination selection, path while
resolving/validating destination, and commit during publication. Invalid root
uses path. Failed writes, including a full byte count with nonnil error, are
failures. Before commit any stdout failure stops before further input. Never
write to stdout after commit; successful status is stderr. If its write fails,
return 1 and make one best-effort report-saved report for a committed save;
do not loop. A failure writing a no-op or cancellation report returns 1 without
claiming a disk mutation. Cleanup status retains precedence if its report fails.
A failing writer may already have accepted some or all bytes; this best-effort
fallback can therefore append a second object. The exact-one-object promise
does not apply to failing writers. No output scheme can retract accepted bytes.

## Fixtures and verification

All fixtures are hand-specified from 0340, existing preparation serialization
examples, editor scoped-repair rules and reviewer outputs. The manifest cites
sources per case. Expected saved bytes are literal YAML, never calculated with
the product encoder. Contract tests use actual files and passed streams; they
do not substitute dependency implementations or inspect product internals.

The test-only binding starts with a zero-output stub for contract-only adoption.
When a separately authored implementation exists, replace only that binding
with its constructor; it is then the sole implementation import. Run the
suite against an empty stub first: every output case must fail. Run separately
authored broken variants and record assertion failures; a table of intended
defects is not evidence of a performed run. Keep execution evidence and remaining
coverage limits in the hardening report, not in this regenerable contract.

## Wrong implementations tried

| Wrong implementation | Caught by |
| -- | -- |
| Drop a recipient field, trim literal values, merge clears, or guess the first team/contact | save-all-nine-recipient-fields-literally; clear-all-nine-does-not-guess-directory-defaults; empty-manager-input-retries-before-explicit-id |
| Adopt the empty guard probe, autosave apply, apply a discarded form, or lose corrections after refused apply | stage-name-not-autosave; discard-entire-local-form; repair-two-references-retain-stale-arrangement |
| Treat Draft save, menu Enter, a value word, incomplete EOF or stale review as save consent | save-at-draft-is-not-consent; empty-menu-does-not-save; literal-save-value; literal-q-value; partial-save-eof; TestStaleReviewAndAppearingDestination |
| Drop preparation problems, blockers, warning severity, guide problems or guide text; suppress one reviewer result because another is present | repair-two-references-retain-stale-arrangement; TestActualPlanValidationAndBlockerReview; TestFirstStepsWarningRetainedBesideGuide |
| Skip identity or byte-change guards, ignore a changed or appearing source, or allow oversized input/encoding | TestIdentityAndAmbiguityRefusal; TestStaleReviewAndAppearingDestination; TestSourceSizeBoundary; TestUnencodableSizeExpansionDoesNotPublish |
| Re-encode a no-op, discard arrangements/unknown values, truncate a shared inode, retain a temporary or follow a selected symlink | noop-unencodable-preserves-original; repair-two-references-retain-stale-arrangement; TestLiteralDisplayAndNoopInode; TestSafePathsAndHardlinks; filesystem snapshots in TestRecipientScenarios |
| Ignore full-count writer errors, prefetch consent, lose saved bytes after report failure, omit the fallback or falsely report cancellation after save | TestReviewOutputFailurePrecedesConsent; TestReviewDisplayFailureDoesNotPrefetchConsent; TestPostcommitReportFailureRetainsSavedBytes; create-identity-explicit |
| Strip existing ACLs or restrictive modes, inherit a broader publication ACL, or reveal terminal controls or record content | TestDarwinExistingAccessControlPreservedOrRefused; TestDarwinExplicitEmptyACLIsNotNoACL; TestDarwinInheritedACLDoesNotBroadenPrivatePublication; TestExistingReadOnlyModePreservedOrRefused; TestLiteralDisplayAndNoopInode; TestRecordIsNotReadAndPrivateDirectories |
| Upsert a new colliding ID, rename a selected item, append a replacement or lose unrelated stale data | new-id-collision-never-upserts-existing; edit-id-is-fixed-not-renameable; edit-whole-arrangement-clears-in-place; append-full-arrangement-ordered-repeated-literal-helpers |
| Sort/deduplicate helpers, remove the wrong occurrence, apply a nested helper form globally or lose unchanged source bytes | helper-removal-selects-one-exact-occurrence; helper-back-discards-nested-proposal; arrangement-back-discards-applied-helper-list; unchanged-empty-helper-form-is-byte-preserving-noop; adding-then-removing-final-helper-restores-noop |
| Default missing kind/status, commit removal early or treat blank input as selection of a malformed identity | new-partial-form-has-no-kind-or-status-default; confirmed-removal-still-cancellable-before-save; removal-confirmation-back-preserves-original; remove-unique-empty-id-by-explicit-ordinal |
| Hide arrangement fields/removed values or display authored controls as live prompts | TestCompleteOriginalAndDraftArrangementReview; TestArrangementScalarDisplayRemainsLiteral |
| Read beyond the interaction budget, charge runes instead of bytes, ignore CRLF framing or treat exhausted partial EOF as cancellation | TestTotalInteractionBudget; TestBudgetDoesNotInspectUnusedSuffixAndPreservesBelowLimitEOF |
| Retain an owned inspection descriptor or close borrowed streams | TestOwnedInspectionDescriptorReleasedWithoutClosingBorrowedStreams |
| Apply a valid helper proposal to the accepted draft before the enclosing arrangement apply | TestValidLocalArrangementChangesStayDiscardable |
| Accept a padded, signed, leading-zero, non-ASCII or otherwise noncanonical ordinal because later actions hide early selection | TestNoncanonicalOrdinalsRetryBeforeAnyFurtherInput |
| Count total input as runes and consume queued consent after 1 MiB of multibyte input | TestMultibyteBudgetEndsAtReviewBeforeQueuedSave |
| Apply a later recipient against loaded original preparation, losing accepted arrangement changes | TestRecipientApplyPreservesPreviouslyAcceptedArrangement |
| Hide actual preparation diagnostics or interleave dependency messages with the fixed triple projection | TestDependencyDiagnosticsUseOnlyAdjacentFixedTriples |

The arrangement and interaction-budget fixtures extend these safeguards without
replacing the recipient and native ACL regressions. Their source manifest cites
0350 and the existing editor, encoder, validator and format rules. Apply the same
independent broken-variant workflow to new-mode upsert, unintended renames,
helper loss/reordering, early apply/removal publication, hidden defaults,
unrelated stale-data loss and budget overflow or consent prefetch. Keep actual
execution evidence separate from the normative fixture definitions.

The static plan/preparation hardlink fixture is
a cross-kind alias refusal control: the real plan and preparation decoders
require different kinds, so its shared bytes cannot pass both. It is not a
reaching proof of an otherwise-valid output-alias guard. Filesystem faults and race coverage must be
reported as observed evidence, not inferred from the happy path.
