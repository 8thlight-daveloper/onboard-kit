# Invitation command contract

Status: agent-authored contract under ADR0300; implementation pending.
Component `command/invite`, package `invite`; filesystem driver, dependency
position 6. Its public interface is Run(args, stdout, stderr); New takes no
arguments. Args starts with `export`, without binary name or `invite`.

Use real plan/preparation decoders and encoders, plan/preparation validators,
question finder, preparation locator, record encoder, firststeps assembler and
invitation encoder, model and this component's own contract. Dependencies are
constructed internally, not injected stand-ins. Driver composition does not
replace their rules. No import/start/record writer/joiner, network, action runner,
shell, clock, Git, credential or download operation.

Reads: argv; only explicit plan/preparation regular files; their complete values
for actual validation/encoding; explicit help references; and metadata/handles
needed to safely traverse those paths and the explicit destination parent.
Writes: supplied stdout/stderr and one new explicit ZIP plus an owned exclusive
temporary file in its existing parent. No saved plan/preparation/record, source
inode, local index, preference, environment, directory tree, alternate output or
other person's artifact changes. No cwd change or environment-derived identity.

## Arguments

Required scalar flags: `--plan`, `--preparation`, `--account`, `--out`, and
`--preaccess-help-contact`. Each appears exactly once with a nonempty value.
`--help-contact` and `--help-role` repeat and retain appearance order across both
kinds. Preaccess-help-contact contributes its contact to GeneralHelp at that
flag's appearance position and asserts that the listed business route is usable
before company access for this export. Ordinary repeated references are deduped
by the actual firststeps contract, not rejected as repeated flags. No guessing
owner/help/account or copying an assertion from a filled field/default.

Support `--flag value` and `--flag=value`. In separate spelling a following token
beginning `--` is a missing value; equals spelling permits such a literal value.
Values are not trimmed or case-folded. Unknown operation/flag, missing/empty
values, repeated scalar flags even when equal, positional tokens and a bare `--`
are usage errors. No flags from prepare, implicit dates, overwrite switch,
existing-record argument, recipient delivery address, or implied export mode.
Parse all argv before file effects. Bad account syntax is data refusal after
syntax parsing but before reading files. Enforce the invitation account-name
budget before constructing account-derived archive names.

Initial `--help`, or `export --help` immediately after the operation, takes
precedence over remaining arguments and performs no file inspection. It returns
0, empty stderr and exactly this stdout line:

```text
Usage: onboard-kit invite export --plan FILE --preparation FILE --account ID --out FILE --preaccess-help-contact ID [--help-contact ID | --help-role ID ...]
```

## Paths and bounded input

Paths are trusted explicit argv selections, never archive entry paths. Relative
argv paths resolve against the process's existing cwd without changing it.
Absolute paths are allowed. Permit a `.` component with its ordinary lexical
meaning; reject every `..` component, NUL, a trailing separator for a file, and
an empty path. Do not expand environment text or tilde, infer a repository, clean
traversal, or resolve a symlink into an accepted alternate path. Repeated interior
separators are ordinary filesystem separators and may be collapsed lexically.
Initial separators retain their platform absolute-path meaning. This is not
the locator's distinct strict slash-template syntax.

Traverse every nontrivial component using stable nonsymlink directory handles.
Each source's final object must be a nonsymlink regular file; refuse FIFO,
device, socket, directory and dangling symlink without blocking. The output
parent must already be a nonsymlink directory and remain the held publication
parent. Its leaf must be absent, including no dangling symlink. Reject output
aliases of either source path/object; an existing hardlink destination is existing
output, never an inode to truncate. Add no separate input-to-input alias phase:
an input of the wrong kind follows the stated decoder refusal order. No root
discovery or archive extraction.

The command's symlink rule also applies to ambient argv aliases such as macOS
/var. The caller may explicitly select a real resolved path; the command does
not silently weaken traversal. Tests use an independently resolved sandbox for
successful controls and create deliberate symlinks separately.

Read each source through a bounded reader: at most 1048576 bytes plus one
sentinel byte, rejecting overflow rather than truncating. Metadata is not a
size proof. Capture the opened regular-file snapshot; no cross-file transactional
snapshot or protection from hostile same-owner content mutation is promised.
Each source may independently reach its input limit, but the real final four
canonical payloads must fit the invitation encoder's aggregate 1048576 budget.

## Phases and readiness

Stop at the first failed phase, with no partial archive. Preserve existing
inputs and unrelated state. Stderr codes below identify the phase without raw
data. Independent warnings accumulate in phase order, keeping each dependency's
held diagnostic order and repeats; do not upgrade them into generic errors.

1. Parse syntax/help; validate exact account syntax/name budget.
2. Check output parent/new leaf, then source path/object/bounded-read safety in
   plan then preparation order. No temporary file yet.
3. Decode plan, then preparation through real decoders. Refuse any decoding
   error. Plan ID must be nonempty. No schema coercion or kind guessing.
4. Actual plan Validate on the complete original plan. Refuse if any error;
   otherwise retain warnings. This includes inactive/unmarked declared graph
   errors through actual structure validation. Do not claim coverage of the
   plan validator's documented missing authoring checks.
5. Actual preparation Validate against explicit account, original plan ID and
   all declared contact/team/section identities, no existing record. Refuse
   its errors; missing optional recipient data/arrangements are not errors.
6. Actual finder Blockers(originalPlan, onboarder). Refuse if nonempty. Required
   onboardee questions alone do not block; unaccepted Defaults do not satisfy
   blockers; explicit Deferred retains current satisfying semantics.
7. Require Recipient.OnboardingOwner to name a known contact with a nonblank
   display name. Require designated preaccess contact to exist with a nonblank
   display name and nonblank Slack or Email route. Blank is empty or all Unicode
   IsSpace. SlackID alone does not qualify. Owner need not equal helper or have
   its own route when the helper supplies one. Do not probe availability.
8. Actual firststeps Assemble with original plan, preparation and appearance-
   ordered explicit GeneralHelp. It validates help and result paths and renders
   actual Bootstrap; refuse any errors, otherwise retain exact text/warnings.
   No prefiltered or preparation-answer-injected plan, private renderer,
   unsupported instructions, default acceptance or fake outcome.
9. Actual canonical plan and preparation encoders. Refuse any problems. Build
   a new schema-version-1 record containing exact Account and reference with
   Plan.ID, Plan.URL and lowercase hex SHA256 of the exact canonical embedded
   plan bytes including final LF. All record answers, outcomes, dates, first
   use and unknown fields are absent. Encode using actual record encoder.
10. Real invitation Encode of these four payloads. Refuse any error or excess;
    no payload omission, truncation, extra entry or compression workaround.
11. Exclusive publication and reporting below.

An empty successful Bootstrap projection alone is allowed with actual neutral
wording. Recipient company email, plannedStartDate, manager, buddy, time zone,
work location and optional arrangements may be absent. Supplied invalid values
still fail validation, and an arranged item retains its completeness rules.
Readiness is no identity confirmation, first use, access check or delivery.

## Typed private reporting

Return 0 for fully reported publication/help, 2 for usage, 1 for every data,
readiness, filesystem, encoding or output-reporting failure. Non-help stdout is
empty before commit. On successful publication stdout is exactly
`Invitation archive written.\n`.

For every non-help result, stderr is one JSON object plus LF. Object key order
and JSON whitespace are free. Its key set is exactly `status`, `code`, `message`,
`problems`, `blockers`. Problems is an array of objects with exactly `code` and
`severity`, projected from applicable dependency diagnostics; never emit their
messages, detail, question, section, parser excerpts or values. Blockers is the
array of exact declared question IDs returned by the finder in a blocker refusal,
in finder order, otherwise an empty array. These question IDs are the one
deliberate diagnostic identity exception: they tell the onboarder what to fix,
without answer/default values. No raw argument, account, unknown selector ID,
path, payload, personal input, OS error or secret is repeated. General help/owner
values belong only inside intentional payloads.

| Code | Status | Fixed message |
| -- | -- | -- |
| invite.usage | refused | Use invite export with all required explicit flags. |
| invite.account | refused | Supply a valid explicit invitation account within the path budget. |
| invite.output-path | refused | Choose a new archive in an existing nonsymlink directory. |
| invite.output-exists | refused | Choose a new archive pathname; existing output is never replaced. |
| invite.plan-read | refused | Supply a bounded local nonsymlink regular plan file. |
| invite.preparation-read | refused | Supply a bounded local nonsymlink regular preparation file. |
| invite.plan-decode | refused | Correct the plan file's format before export. |
| invite.preparation-decode | refused | Correct the preparation file's format before export. |
| invite.plan-identity | refused | Give the organization plan a nonempty identity before export. |
| invite.plan-invalid | refused | Inspect the plan with the supported stdio validate workflow and correct its validation errors before export. |
| invite.preparation-invalid | refused | Correct the preparation identity or details before export. |
| invite.onboarder-blocked | refused | Supply or deliberately defer the required onboarder answers before export. |
| invite.owner | refused | Name the onboarding owner in preparation with a usable display name. |
| invite.preaccess-help | refused | Designate a known named business contact with a displayable pre-access help route. |
| invite.firststeps | refused | Correct the help selection or first-steps inputs before export. |
| invite.plan-encode | refused | Correct plan values that cannot be encoded before export. |
| invite.preparation-encode | refused | Correct preparation values that cannot be encoded before export. |
| invite.record-encode | refused | The identity-only record seed could not be encoded. |
| invite.archive | refused | Keep all invitation payloads within the archive limits. |
| invite.publish | refused | The new archive could not be safely published; existing files were preserved. |
| invite.cleanup-refused | refused | No archive was published, but temporary invitation data may remain. Restore access to the selected output directory and check exporter-owned temporary data before retrying. |
| invite.cleanup-written | written | The archive was written, but temporary invitation data may remain. Keep the archive; restore access to the selected output directory and check exporter-owned temporary data. |
| invite.written | written | Pre-access help availability has not been tested; deliver through an approved channel. |
| invite.result-output | written | The archive was written, but its result could not be fully displayed. |
| invite.help-output | refused | The usage text could not be displayed. |

For a phase refusal, problems contains prior warnings followed by all diagnostics
from that failed dependency, in its order, projected without private fields.
Path/readiness/usage/publication errors add no synthetic model diagnostics.
Successful warnings include real plan and firststeps warnings in that order.
No warning authorizes changing a payload or asserting tested availability.

## Publication and reporting failures

Finish all validation and byte construction before creating a private mode-0600
exclusive temporary regular file beside the selected destination. Check every
write and close. Publish atomically with a no-replace operation relative to the
held parent; never use a replacement-capable rename, unlink an existing leaf,
truncate a source inode, follow a symlink or create a directory. An output that
appears after preflight is still refused with invite.output-exists. Concurrent
valid exports to one absent leaf must yield exactly one publication and one
refusal; final bytes equal one complete canonical archive, never a mixture.

Remove owned temporary files on completed ordinary failure or success. Leave
no partial final file on a precommit error. The original leaf remains absent or
exactly as the concurrent creator left it. This is not crash durability or a
universal hostile same-owner race guarantee. Do not infer private repository
visibility from local output mode.

Under ADR0310, check owned temporary cleanup rather than discarding its result. If the OS
prevents removing the owned temporary entry, return 1 and report the repair
need. Before publication, use invite.cleanup-refused with status refused and
empty stdout; retain the actual absent or independently created final leaf.
After publication, use invite.cleanup-written with status written; keep the
complete committed archive and still attempt the ordinary written stdout line.
Both reports keep the existing five keys and accumulated dependency warnings,
add no synthetic model diagnostic, and disclose no temporary basename, raw
path, account, OS error or other private value. Never delete a committed archive
to make a refusal report true. Cleanup failure takes priority over a secondary
stdout/report-writer failure; if stderr is writable, retain the cleanup code.
An unwritable stderr cannot be promised a diagnostic, but still requires exit1.
Absence of the owned temporary entry counts as cleanup, including ENOENT.
Do not clear inherited ACLs, change directory permissions, alter unrelated
entries, or retry indefinitely to force cleanup. Normal temporary mode0600
setting does not authorize stripping its inherited access controls. This is
truthful ordinary OS-error handling, not a crash-cleanup or hostile-host promise.

Atomic publication is the commit point. Afterwards write success stdout then
the stderr report. If stdout fails without a cleanup error, keep the archive, return 1 and attempt the
fixed invite.result-output report with status written. Stderr failure returns
1 and keeps any committed archive; it cannot guarantee an unwritable diagnostic.
Do not retry indefinitely or roll back publication. Help stdout failure has no
file effects and returns 1, attempting invite.help-output. Before commit, report
writer failure also returns 1 without filesystem effects. Partial supplied-writer
output on an I/O error is allowed; silent return 0 is not.

## Independent oracles

The API/stub, tests and testdata form the other three parts. Sources are listed
in sources.json. Canonical payloads and guide are literal data; the baseline
hash was computed independently of product source. Tests inspect archive/zip
metadata, CRC32, payload bytes and SHA256 without importing product codecs.
Filesystem tests use actual OS objects and concurrent real Command.Run calls.
Specialized race/fault-injection requirements need independent OS-level evidence;
they must not be claimed verified merely because a simpler case passes.

| Plausible wrong implementation | Independent distinguishing boundary |
| -- | -- |
| Hash raw input YAML rather than the embedded canonical plan | Comment-spelling case and literal baseline digest |
| Drop Plan.URL or copy a plan pre-seed into the record | Literal URL/pre-seed composition payloads |
| Accept an unselected default, reject deliberate deferral, or block onboardee questions | Required/default/deferred/audience literal cases |
| Filter to active Bootstrap before checking relationships | Inactive unmarked unknown-prerequisite refusal |
| Guess help, accept SlackID as a route, or deduplicate by rendered text | Missing/unknown help and ordered mixed contact/role cases |
| Inject preparation team into Bootstrap answers | Original-plan selection and exact unchanged guide case |
| Rewrite input files, replace an existing archive, or follow symlinks | Source snapshots and real filesystem refusal cases |
| Publish by replacement-capable rename | Existing-object refusal and real concurrent export contenders |
| Omit a payload or truncate to evade the archive budget | Exact four payloads and real canonical aggregate boundary |
| Copy personal fields or parser/OS errors into reporting | Literal guide equality and private-marker diagnostics |
| Report success despite a supplied output-writer error | Explicit post-publication and help-writer failure cases |
| Ignore denied temporary unlink and report success | Independently controlled inherited-ACL postcommit cleanup case |
| Report refusal or delete final output after committed cleanup failure | Complete literal ZIP, retained temporary and exact written-status cleanup report |
| Replace cleanup repair need with a secondary stdout error | Same real ACL fault combined with the supplied failing stdout writer |
| Strip inherited ACLs to hide the cleanup fault | Controlled ACL retains the temporary after OS-denied unlink |

These rows define intended distinguishing tests, not execution claims. Applied
wrong-variant evidence and unresolved practical probes belong in working reports.

## Proposed wrong implementations

These are intended discriminators, not executed evidence or a mutation score.

| Defect | Independent draft oracle |
| -- | -- |
| Return exit0 without publishing or reporting | Literal baseline requires typed report, four payloads and exact guide |
| Infer operation/account/output or ignore repeated scalar flags | Usage cases and account-before-file-read case |
| Parse help after inspecting supplied files | Initial/operation help precedence with nonexistent private paths |
| Echo OS/parser/unknown-selector errors | Exact static messages, projected diagnostic key sets and private markers |
| Omit required onboarder questions or accept Default | Literal blocker/default cases with exact question IDs |
| Block required onboardee or intentional Deferred | Explicit successful audience/Deferred cases |
| Treat SlackID as usable pre-access help | SlackID-only helper refusal |
| Guess owner/help from directory order | Missing-owner and unknown-helper cases; explicit selection ordering |
| Inject preparation team or copy preseed into new record | Original-plan and identity-only seed composition checks |
| Hash raw YAML/Source.Hash instead of canonical embedded bytes | Literal baseline/hash plus source-comment variant and independent SHA256 |
| Reimplement firststeps or drop Bootstrap warnings | Full literal guide and exact warning/fragment composition |
| Use configured result paths as ZIP entry names | Exact four-name metadata and custom-destination guide checks |
| Drop preparation/help to fit size limit | Actual aggregate boundary with literal padding preservation |
| Permit symlinks/FIFO or truncate existing output/input inode | Actual object fixtures and byte/mode/link snapshots |
| Precheck existence then replace at publication | Concurrent real commands; deterministic guard-crossing oracle remains pending |
| Accept one byte too much or reject inclusive input cap | Valid exact-size YAML comment input and one-byte-over input cases |
| Stamp time, compress, add entries/comments or metadata | Independent ZIP reader/local/EOCD profile assertions |
| Keep source/call state or change ZIP by output location | Reused command over independent identical source sandboxes |
| Roll back committed ZIP after stdout failure or silently succeed | Public writer failures plus retained full archive checks |
