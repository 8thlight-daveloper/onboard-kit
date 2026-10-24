# Catalog creator contract

Status: independent proposed four-part contract; implementation and execution
pending. Sources: proposed0430,0410,0030,0100,0110,0120, the existing materializer
and encoder contracts, and0340/0350 plain interaction/publication conventions.

## API, context and effects

New() takes no arguments and constructs the actual catalog materializer and plan
encoder. Run receives arguments after the exact catalog create pair and borrowed
input/stdout/stderr. No injected semantic dependencies, filesystem backend,
alternate catalog, renderer, clock, generated identity or semantic policy.

Position6 permits only this contract, model, catalog materializer5, plan encoder1,
standard library and the already pinned x/sys/unix native filesystem facilities.
Do not call another driver, legacy instantiator, validator, finder or hash/parser
algorithm copied from a core component. Binary8 dispatches this exact pair;
command/catalog retains list/show/export and its own grammar.

Reads: explicit argv/actions, actual compiled materializer results, actual encoder
results, selected output-parent/path/native access metadata and writer results.
Writes: borrowed terminal streams and, only after current reviewed save, the one
explicit new plan plus owned same-parent publication temporary resources. No
directory creation, existing-plan read/edit, personal preparation/record read or
write, environment/clock/network/process/shell/Git operation, invitation/export,
outcome, event, default acceptance, transcript or persistent session state.

The materializer request starts with exact explicit PlanID, zero Provenance,
zero Directory, no PlanVariables/PlanBindings and an allocated empty Instances
list. Each accepted draft change creates a new request; there is no existing
plan input or hidden template selection. Required unanswered facts are allowed.
This driver does not add directory or plan-variable authoring. Invalid supplied
directory IDs in local bindings still fail actual dependency validation.

## CLI and output-path admission

Exactly one --id ID and --out FILE, separate or first-equals spelling, any order.
No positional/unknown/repeated/missing/empty options; separate values cannot
start --, equals values can. No single-dash aliases except initial -h. Initial
-h/--help wins over all later data without input, dependency invocation or
filesystem inspection. Later help is usage. Help bytes are testdata/help.txt.

After initial-help handling, cap argv at4 entries and32768 held UTF8 bytes.
Decoded ID and output path each fit4096 UTF8 bytes. Reject invalid UTF8/NUL in
either; ID otherwise stays exact opaque text, including spaces/printable Unicode.
These CLI violations all use create.usage before filesystem/input access.

Resolve an output path against the current working directory once if relative.
The selected current directory/path components must already safely name real
nonsymlink directories. Reject literal .. components, empty filename, trailing
separator, symlink components, non-directory parent, missing parent, or existing
destination of any kind including dangling symlinks. Dot/redundant separators may
be collapsed only after safely traversing their material components. Absolute
paths are allowed. Anchored path fits4096 bytes. No tilde, environment, URL,
percent-escape or shell expansion. Inspect with held nonsymlink directory
capabilities, not an unchecked reopen string. Path refusal is create.path.

Complete CLI/path admission, build the initial empty request with actual Create,
and retain a stable selected-parent native identity/absence guard before startup.
Any initial materializer error uses create.data with safely projected Problems;
this is not a reason to ask a question or create a file. No preconsent temporary,
directory, lock file, manifest or output file exists. Descriptor-only inspection
is permitted and all owned descriptors close before return.

## Input and safe display

Read one current complete logical line without reading any byte of the next
line, including into buffers. Each payload fits65536 UTF8 bytes excluding LF or
complete CRLF; bare terminal CR is payload. Total consumed bytes fit1048576,
framing included. Before any read with no remaining budget, refuse create.input;
unread EOF/cancel does not outrank budget exhaustion. A complete save/cancel
within budget is acted on without probing trailing input. Partial EOF is never
an action or field value; ordinary valid EOF before save cancels the draft.
Invalid UTF8, oversize partial line, input error or zero-byte/nil-error read uses
create.input. A positive count paired with EOF completes a LF/CRLF-framed line;
a nonEOF error vetoes that read, even if it returned a complete action. No
trimming, case folding, default menu selection or global q/save command.

Q is the existing preparation guide's exact terminal quoting rule: printable
Unicode stays literal; quote/backslash/LF/CR/tab use ordinary short escapes;
other C0/C1 controls, format characters, U+2028/U+2029 and nonprintable runes use
lowercase four-hex escapes or UTF16 surrogate pairs, inside ASCII double quotes.
Q changes display only. No ANSI, raw terminal, cursor move, clear screen, mouse,
terminal-size query, color meaning or hidden global terminal input.

For an LF-containing document, D(label,bytes) emits label: on one line, then
`  | Q(part)` for every LF-split part including the final empty part after a
terminal LF. Valid UTF8 is guaranteed by real core/encoder results. D never
rewrites the encoded plan or guide. Build the complete review using a capped
8MiB builder before displaying it; no allocate-then-measure expansion.

## Draft interaction

Startup is exactly testdata/startup.txt with @ID@/@OUT@ replaced by Q values.
The final startup line and every return to Draft are exactly:

```text
Draft — not saved. Choose add, bind, unbind, remove, review, or cancel:
```

At Draft only those exact actions exist. Empty/unknown actions emit
`Choose one of the displayed actions.` and the same prompt. Cancel here or at
Review discards memory and returns create.cancelled. At value prompts, words
such as cancel/save/remove are literal field text and never actions.

Add emits `Source ID:` then reads one complete line, then `Permanent section ID:`
and reads one complete line. Append a fresh instance with no bindings to a
candidate request and call actual Create. Any error rejects the complete change,
keeps the accepted request/result and prints `Change not applied.` then P below.
Success adopts the entire candidate and prints `Change staged.`. Both return
to Draft. IDs are never inferred, truncated or derived from source/position.
Core duplicate/profile/carrier failures retain their actual ordered diagnostics.

Bind emits `Instance ID:` and reads one line. The ID must exactly match an
accepted draft instance; otherwise emit `Select an existing draft instance.`
and return to Draft without reading a variable/value. For a known instance,
print `Variables: N` and each declaration in its held source order:
`  - Name: Q(name) Type: Q(type) Audience: Q(audience) Required: true/false`.
Display effective empty type/audience as string/onboarder only; this does not
change the request or accept a declaration default. Then ask `Variable name:`
and read one literal line; next prompt is `Value kind: text, flag, list, or deferred:`.

Text asks `Text value (empty text is explicit):` and accepts any complete line,
including empty. Flag asks `Flag value: true or false:`; only exact true/false
sets Flag, otherwise print `Use true or false.` and repeat that value prompt.
List asks `List length: 0-4096:`. Require canonical ASCII decimal 0 or1..4096,
no signs/leading zeroes/spaces/overflow; otherwise print
`Use a canonical count from 0 through 4096.` and repeat. Then prompt
`List item N:` for each one-based item, retaining literal empty items/order/repeats.
Deferred reads no value. An invalid kind prints the fixed invalid-action line
and repeats the kind prompt. No None kind is offered: Unbind removes a key,
whereas encoder refuses a stored None answer. Every constructed Value has only
its selected payload; false, empty text/list and Deferred are explicit values.

Replace the exact existing scoped Binding of that Name if one exists, in place,
or append a new one. This explicit bind action authorizes replacement in the
unsaved draft; it never propagates to other instances or plan scope. Call actual
Create with that whole candidate. Reject atomically on any error as for Add;
otherwise adopt. Unknown names, ownership, kind/options/pattern/secrets and
submitted directory targets are judged only by the actual materializer.

Unbind asks Instance ID and Variable name in that order, with the same unknown-
instance behavior. Remove exactly one existing scoped draft binding by Name;
if absent emit `Select an existing draft binding.` and return to Draft. A valid
removal calls actual Create on the whole candidate and follows Add's atomic
accept/reject behavior. Unbind never writes a None value or fills a default.

Remove asks `Instance ID:`. An unknown ID emits the existing-instance line and
returns to Draft. Remove exactly that never-published instance and its bindings
from a candidate; call actual Create and follow Add's atomic result. Preserve
remaining request order. An ID removed from this unsaved creation may be entered
again; no saved identity/history is being retired or transferred. Editing or
retiring a saved plan remains0410's separately contracted work.

P(diagnostics) is `Problems: N`, followed by the actual ordered entries as
`  - Stage: Q(stage) Path: Q(path) Code: Q(code) Severity: Q(severity) Repair: Q(repair)`.
Do not print Problem.Message/Question/Section/Detail values. Encoder problems
use stage encode and empty wrapper Path. No dependency diagnostics are invented,
filtered or mistaken for readiness. Warnings from successful changes accompany
the next full review rather than being emitted twice.

Repair is a fixed display projection, not validation or an alternative diagnostic.
Use the exact first matching code row below; secret. prefix is the only code
prefix rule. Otherwise use the exact stage fallback. Keep every actual entry,
severity, code, wrapper path and multiplicity in its original order.

| Actual code | Exact repair |
| -- | -- |
| catalog.unknown-id | Choose an exact section-template ID from catalog list. |
| materialize.carrier | Select a section-template source for an activity. |
| materialize.instance-id | Supply a nonreserved lowercase permanent section ID. |
| materialize.duplicate-instance | Give each independent activity its own permanent section ID. |
| materialize.duplicate-binding | Supply each scoped binding name once. |
| materialize.binding-name | Bind an exact variable declared in this scope. |
| materialize.binding-owner | Keep newcomer answers outside the shared organization plan. |
| materialize.limit | Reduce the selection or held creation data to the supported limits. |
| materialize.text | Use valid UTF8 creation data. |
| materialize.value-kind | Choose a supported value kind. |
| materialize.plan-id | Supply an explicit nonempty plan identity. |
| answer.wrong-kind | Use a value kind that matches the declared variable type. |
| answer.not-an-option | Use an exact declared choice option. |
| answer.pattern-mismatch | Use a value matching the declared pattern. |
| directory.unknown-id | Keep this binding unset until its directory target is declared. |
| secret. prefix | Remove credential-like content; supply a nonsecret organization fact. |

| Other actual stage | Exact fallback repair |
| -- | -- |
| request, binding | Correct the indicated instance or scoped binding before review. |
| source | Choose an exact section-template ID from catalog list. |
| load, profile, authoring | Ask the catalog maintainer to repair this definition. |
| structure, plan | Repair the plan definition or scoped facts before review. |
| answer | Use a value accepted by this variable, or leave the binding unset. |
| guidance | Ask the catalog maintainer to repair the preparation guide. |
| encode | Remove or replace the unrepresentable draft value before review. |
| any other stage | Resolve the indicated creation problem before saving. |

These vetted repairs do not repeat rejected input or copy a scanner/pattern/
directory algorithm. The actual dependency remains the only source of problems.

## Complete review and consent

Review reruns actual Create for the accepted request and actual Encode on its
returned plan. Warnings permit review. Any errors emit `Review unavailable.`,
then P for the actual failing batch, and return to Draft without save prompt.
Encoded plan maximum1048576 bytes inclusive; over-limit bytes or capped-display
refusal emits `Review unavailable: supported output limits exceeded.` and returns
to Draft. No truncation/save/partial review follows. Drop any old review cache.

For successful encoding, freeze those exact bytes. Emit in order:

```text
Catalog review — not saved. Saving does not establish invitation readiness.
```

Then Plan and Destination lines as at startup, P(actual materializer warnings),
D(Plan file,exact encoding), `Preparation guidance: N`, and for every actual
guide in actual order: `Guide: Q(id)`, `Instance reasons: N`, each
`  - Q(permanentID)` in actual order, then D(Advisory instructions,exact Markdown).
Empty selection has zero guides. No guide is acted on or serialized into the
plan. Finish with exactly:

```text
Preparation guidance is advisory; no personal arrangement or activity is completed.
Review complete. Choose save, back, or cancel:
```

All review/prompt writes must fully succeed before reading any following byte,
including queued save. A short write, zero+error or full-count+error uses
create.output and publishes nothing. At this Review only exact save publishes;
back invalidates the frozen review and returns to Draft. Unknown/empty actions
emit the fixed invalid-action line and repeat only the Review prompt. Any later
accepted edit needs a fresh complete review; save at Draft is just an unknown
action. No new materialization or encoding occurs after save consent.

## Reviewed exclusive publication

After complete save, recheck the exact selected parent path/native identity and
safe capability chain and require destination still absent. An observed parent
replacement or newly appearing destination uses create.stale; preserve the owner
change and publish nothing. This is a cooperative present-object recheck, not
hostile-owner CAS. No automatic retry, overwrite or alternate destination.

Only now create an exclusive same-parent regular temporary, with effective UID/
GID, mode0600 and no access/default ACL before writing personal/shared facts.
Never strip inherited/parent ACLs or change unrelated controls to proceed;
refuse create.commit if private creation cannot be established. Cap each native
ACL representation at65536 decoded bytes before allocation. No directory creation.
Check full write/short-write/close results and atomically publish no-replace
relative to the held parent, checking current guards immediately beforehand.
Replacement-capable rename alone is insufficient. The final file is exactly
the frozen reviewed bytes with those private controls. No modification of an
existing inode or record/preparation path occurs.
An atomic no-replace collision is create.stale, including when the destination
appears after the final cooperative recheck. Cleanup the owned temporary while
preserving the independently created destination.

Check owned temporary cleanup on success and failure; absent counts as cleaned.
Before publication cleanup failure uses create.cleanup-refused with destination
preserved. After publication keep the complete final file and report
create.cleanup-written; never roll it back. Close owned descriptors before return
and never close borrowed input/output. No claim of crash durability, multi-file
recovery, universal filesystem support or hostile race resistance.

## Terminal reports

Normally accepting stderr gets one LF-terminated JSON object with exactly
status/code/message. Keys may be in any order; no raw path, answer, input, parser
or OS error appears. Help alone has empty stderr. Successful save has no extra
stdout status line; the last stdout is the completed Review prompt.

| Exit | status | code | Exact message |
| -- | -- | -- | -- |
|0|cancelled|create.cancelled|Cancelled. No files changed.|
|0|written|create.written|The reviewed organization plan was created.|
|2|refused|create.usage|Use catalog create --id ID --out FILE.|
|1|refused|create.path|Choose an absent plan file in an existing safe directory.|
|1|refused|create.data|Repair the displayed creation data before saving.|
|1|refused|create.input|Use complete valid input within the displayed limits.|
|1|refused|create.output|Required output failed before save; no plan was published.|
|1|refused|create.stale|The destination context changed; reload and review again.|
|1|refused|create.commit|The plan was not published; owned temporary data may need cleanup.|
|1|refused|create.cleanup-refused|The plan was not published; temporary cleanup needs repair.|
|1|written|create.cleanup-written|The plan was created; temporary cleanup needs repair.|
|1|written|create.report-written|The plan was created, but its report could not be written.|

Input/output failures before publication cannot claim unchanged filesystem if a
consented temporary was created; use commit/cleanup refusal then, not input or
preconsent output wording. After commit cleanup/close repair outranks a report
failure. If final reporting fails after creation, return1 and attempt one fixed
report-written fallback; retain the final plan. A cleanup-written fallback keeps
that higher-priority code. Refusal reporting failure returns1 with no new retry
or alternate destination. Help write failure uses create.output with no facts
or input read. No report is promised to an unwritable stderr writer.

## Four parts and required evidence

api.go is the permanent interface. testdata holds independently authored literal
help/startup/review/plan/transcripts/cases and canonical hash texts. creator_test.go
and native_test.go judge public streams/files only; wire_test.go starts with a
zero stub and later changes only to actual New(). Expected payloads never come
from the product materializer/encoder. Whole-binary scenarios separately judge
root routing, explicit save and unrelated artifact preservation.

Before building, root runs the tests against the supplied zero stub; every
meaningful output/refusal/save case must fail by assertion. The author does not
build this unit. Root then applies the finite wrong implementations in
verification-plan.md, native mutation, an independent build and differential
checks. Drafting supplies no passing or variant-killing evidence.
