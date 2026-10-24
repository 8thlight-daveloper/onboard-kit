# Starter terminal contract

Status: complete independently authored exact-byte contract under0400;
implementation pending, not executable evidence.
Every displayed line below ends LF. Q/D/L/V are the substitution rules here,
not literal function names in output. No ANSI, pager, terminal capability lookup,
colors, command execution or authored terminal escapes.

## Literal display and lists

Q(s) is ASCII double quotes around exact text. Escape quote/backslash as `\"`/
`\\`, LF/CR/tab as `\n`/`\r`/`\t`; every other C0/C1 control, Unicode format
character, U+2028/U+2029 and other nonprintable rune uses lowercase four-digit
`\u` escapes, supplementary points as UTF16 surrogate-pair escapes. Remaining
printable Unicode stays literal, including ordinary nonASCII letters; no HTML
escaping. This changes display only. Invalid UTF8 from paths/input refuses; the
actual reviewer already establishes source UTF8 representation.

D(label,s) prints `label:` then one `  | Q(part)` line per LF-separated part,
including final empty part when s ends LF. An empty document has one quoted
empty part. This safely presents every original line/control/trailing newline,
not terminal-executable Markdown. L(label,items) prints `label: N`, then each
item as `  - Q(item)` in held order, with repeats and empty strings retained.
Numbers are unpadded ASCII decimal; booleans lowercase true/false.

V(label,value) prints `label: none`, `label: deferred`, `label: flag true/false`,
`label: text Q(text)` or `label: list N` followed by `  - Q(item)` lines. Only
selected-kind payload displays. A Text deferred is visibly quoted text, not
the Deferred value. Never turn a declared default/suggestion into an Answer.

For a Diagnostic list, print its label and decimal count, then one
`  - Severity: Q(severity) Repair: Q(Display)` line per actual entry in order.
Severity is error/warning when exactly those enums; otherwise fixed unknown.
Never print raw Problem.Message/Detail/Question/Section, raw error strings or
parse excerpts. Do not reconstruct the reviewer's semantic error catalogue.
Fatal actual reviewer diagnostics use the list label Opening problems.

## CLI, input and prompts

Run receives arguments after start. Exactly one mode is allowed:

```text
--invitation FILE --root DIR [--home DIR]
--resume-import TOKEN --root DIR [--home DIR]
```

Options may occur in any order, each once, using separate values or equals.
Reject positional/unknown/duplicate/mixed-mode/missing/empty options. Separate
values cannot start `--`; equals values can. TOKEN is exactly64 lowercase hex.
No force/yes/time/org-plan/no-walk option. Initial -h/--help wins before all
other syntax and without input/filesystem/runtime callbacks; elsewhere invalid.
Help prints exactly these four lines and empty stderr:

```text
Usage: onboard-kit start (--invitation FILE | --resume-import TOKEN) --root DIR [--home DIR]

Inspect an invitation, confirm identity, review, and explicitly begin a read-only session.
Existing work is retained or reconciled; no checks, answers, or outcomes are executed or saved.
```

Read one complete current line, never another byte of the next logical line,
including prefetch into a buffer. Payload maximum65536 UTF8 bytes, excluding
complete LF or CRLF framing. Bare terminal CR is payload. Invalid UTF8 and an
oversized partial line refuse even at EOF. Ordinary valid incomplete EOF never
counts as an action. Total consumed bytes maximum1048576, framing included;
do not read a further byte at the limit. Exhaustion precedes an unobserved EOF.
A completed final action at the boundary is valid without probing EOF. Complete
begin may commit, but a subsequent read-only-loop limit failure is opened-with-
repair, not retroactive refusal of that consent. Data returned with io.EOF is
processed under these same rules. Any input error is failure, not cancellation.
No trimming, casefolding or blank action default.

An unrecognized/empty complete action prints exactly `Choose one of the displayed actions.`
then the same state's prompt. Cancel is only an action where explicitly shown.
Every prompt/review write must fully succeed before reading its following action;
short+nil, zero+error and full-count+error all fail. Consent cannot be consumed
after a failed preceding display, even if no publication would have followed.

## Fresh identity and complete review

After actual Inspect and safe physical context selection, print:

```text
Invitation opening is explicit.
Organization plan and preparation may be shared. Personal records and recovery copies contain personal data.
Terminal output may remain in scrollback. Viewing accepts no defaults and saves no answers or outcomes.
Input limit: 1048576 bytes total; 65536 bytes per logical line.
Account: Q(account)
Plan: Q(planID)
Selected root: Q(root)
State home: Q(home)
Local plan: Q(localPlanPath)
Preparation: Q(preparationPath)
Record: Q(recordPath)
Recovery: Q(home/import-pending)
Identity: choose confirm ACCOUNT or cancel:
```

ACCOUNT in the prompt is actual syntax-validated ASCII account, unquoted; only
exact `confirm ACCOUNT` accepts this displayed account AND plan identity for
this operation. No plan-ID copying is required. Cancel/ordinary EOF returns
start.cancelled and makes no new objects. This is identity only, not begin.

If actual Review needs a plan choice, print exactly:

```text
Local plan bytes differ from the invitation. No recency or merge is assumed.
Plan copy: choose use-incoming or cancel:
```

Only exact use-incoming becomes that reviewer choice; it is still not save.
On a completed Proposal, pre-render the bounded frozen Session block below
before offering begin. Then print `Opening review — not begun.` and these lines:

```text
Account: Q(account)
Plan: Q(planID)
```

Print four roles in storage order as `Role: Q(role) Action: Q(action) Path: Q(path)`.
Metadata action is actual install/retain/replace registration work, not a fifth
semantic artifact. Print `Other matching contexts: N`, then each
`  - Root: Q(registeredRoot) Record: Q(recordRelativePath)` in bytewise stored-key
order. If new registration and N>0, print `This opening registers an independent local context; other contexts are not copied or moved.`
If it also needs a new record event, print `Begin creates an independent record and session here; other matching work remains untouched.`
No older event is copied or presented as global first use.
Print `Previous comparison: Q(reason)` and its safe diagnostic
list labeled Previous comparison problems. Print the six actual reconciliation lists in this order using L: New
questions, Removed questions, Answers to review, New activities, Removed
activities, Different recorded definitions. Never say an empty unavailable
comparison proves no changes or a definition mismatch proves subsequent editing.

Print safe diagnostic lists labeled Incoming warnings, Effective plan warnings,
Support problems, Session problems; Support problems keep actual error severity
even when the complete core proposal remains usable. Print `Support available: true/false`.
Then print exactly one of:

```text
First use: a new begin event will be sampled only after begin.
First use: the existing event is preserved; an empty event stays unknown.
```

Finish review with:

```text
Private recovery copies may temporarily exist in the state home and selected artifact parents.
Recovery: Q(home/import-pending)
Begin enters a read-only onboarding view. It does not verify activities, run commands, or accept answers.
Review complete. Choose begin or cancel:
```

Only complete begin is publication consent. Cancel/EOF here preserves every
source and creates no state/target/temp/lockfile. Review/display errors never
consume queued begin. Pure Begin runs only after accepted begin and fresh guards.

## Frozen actual session block

This entire block is encoded once from the accepted Review BEFORE begin, using
a capped builder. No unknown fields, transcript, opaque question Group or raw
diagnostic payload enters it. Preserve every meaningful actual field enumerated
below; do not silently shorten, omit waiting work or infer completion. The new
event is a separate frozen Publication scalar, not part of this prebuilt block.

Start with:

```text
Read-only onboarding session.
Instructions are displayed, not executed. Dates are goals, not completion evidence.
Defaults and suggestions are not accepted answers.
```

Print `Activities: N`. For each actual activity in order, print `Activity: K`
then scalar Q fields ID, Title, Summary, Parent, Owner, Timing, Verification,
Current definition hash; D(Body,body); L(Requires,requires), L(After,after),
L(Waiting on,waitingOn), L(Escalation,renderedEscalation). These values come from
the actual resolved Section and walker row; no second rendering/hash/schedule.
Bootstrap/Enabled/Source/conditions and declaration metadata are not completion
claims and are not separately dumped. Active questions below carry their data.

Print `Steps: N` per activity. For each held step print `Step: K`, Q fields ID,
Name, Owner, Timing, Open, Directory; D(Notes,notes), L(Run,run), L(Escalation,
heldStepEscalation). Print `Environment: N`, then `  - Key: Q(key) Value: Q(value)`
in exact bytewise key order. Print `Checks: N`, each `Check: K`, L(Run,run),
D(Ask,ask), and `Environment question: Q(check.Env)`. No check/run/open/env action
is performed or interpreted by displaying it.

Print `Questions: N` in actual finder order. Each `Question: K` has Q fields ID,
Scope, Name, Description, Example, Type, Audience; Required/Multiple booleans;
V(Answer,answer), V(Default suggestion,declaredDefault), L(Options,options).
Print `Choices: N`, each `  - Value: Q(value) Label: Q(label)`. Print
`Suggestions: N`, each `Suggestion: K`, `Source: Q(source)` and V(Value,value).
Print L(Blockers,actual blocker Question strings), preserving actual order.

Print `Recorded outcomes: N` in actual derived history order. Each `Recorded
outcome: K` has Q fields Section, Outcome, On, Recorded definition hash, Current
definition hash, Comparison; D(Note,note), D(Differently,differently). No history
field is rewritten, dropped, reset, verified or date-compared by this display.

Print `Support available: true/false`. False prints exactly `Current support is unavailable; ask the onboarding owner to repair preparation.`
and the safe Support problems list. True prints every known Recipient scalar,
with Q labels Name, Company email, Planned start date (goal), Team ID, Manager
contact ID, Onboarding owner contact ID, Buddy contact ID, Work location, Time
zone. Then `Arrangements: N`; each `Arrangement: K` has Q fields ID, Section,
Kind, Owner contact ID, Reference, Target date (goal), Status; L(Helper contact
IDs,contacts), D(Note,note). Preserve repeats and spaces; no directory fallback,
guessed helper or preparation-to-answer/outcome binding.

Print safe Session problems, then D(Invitation-time guidance,originalFirstSteps).
This label does not claim guidance is newly validated against retained support.
No personal unknown-map data enters this display. Individual scalar/doc/list
labels specify exact casing; K is one-based decimal, no blank separator lines.

After completion, separately print `First opened at: Q(frozenEvent)` (empty
quotes mean unknown), then the EXACT frozen block bytes, then
`Session: read only. Choose exit:`. Exact exit or ordinary EOF ends successfully.
Other complete actions, including cancel, print invalid-action and reprompt.

## Recovery prompts and completion reporting

With complete pending and no receipt, print `Pending opening: Q(token)`,
`Account: Q(account)`, `Plan: Q(planID)`, the four frozen role/action/path lines,
then `Recovery: choose resume or cancel:`. Only exact resume rolls forward.
With valid receipt and pending, show the same frozen context, then
`Completed opening: choose clean or cancel:`; clean never republishes roles.
With receipt but pending absent, print `Completed opening: Q(token)` and
`Session display is no longer staged. Cleanup will not reopen the session.`
then the same clean prompt. Successful cleanup returns start.cleaned, without
pretending a session was displayed. Attributable incomplete staging instead
prints `Incomplete staging: Q(token)`, `Reopen the invitation after cleanup.`,
then `Staging: choose clean or cancel:`. Unknown/unattributed objects refuse
without deletion. Recovery cancel/EOF says existing state was retained.

stderr is exactly one LF-terminated JSON object for normally accepting writers,
keys in order status/code/message/transaction. No dynamic path/account/parser
text. Transaction is empty or known64hex token. Help alone has empty stderr.
Finite codes/messages/status/precedence are fixed in contract.md. A failing
writer may have already accepted bytes; exact-one-object excludes that case.
After completed report failure, make one fixed report-opened fallback attempt,
never loop, republish, resample or call a completed operation cancelled.
