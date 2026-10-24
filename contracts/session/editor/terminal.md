# Session editor terminal grammar

Every structural line is LF. No ANSI, pager, shell, hyperlinks or terminal probes.
Q/D/L/V are EXACT public `contracts/invitation/starter/terminal.md` Literal
display and lists definitions, frozen by this reference, not a starter call.
All source/runtime text goes through Q/D/L/V, including controls/bidi format.

Help is exactly testdata/help.txt. Identity block is exactly:

```text
Record-backed session editing is explicit.
Plan and preparation are organization-owned; answers and outcomes are newcomer-owned.
Records, terminal scrollback and private staging can contain personal data.
Registered account: Q(account)
Registered plan: Q(planID)
Registered root: Q(originalRoot)
State home: Q(home)
Session key: Q(key)
Registered plan file: Q(planPath)
Registered preparation: Q(preparationPath)
Registered record: Q(recordPath)
Current locator binding is checked only after identity confirmation.
Identity: type confirm Q(account) or cancel:
```

Only the EXACT `confirm ` followed by the displayed Q(account) spelling accepts
identity. For newcomer: `confirm "newcomer"`. This is not save. Cancel or ordinary
valid incomplete EOF cancels; malformed/UTF8/oversized/error input refuses.

Initial actual reviewer inspection has two display paths.

If the initial actual reviewer Inspect returns fatal diagnostics, print only
their complete safe Session problems block and fixed repair instruction, then
refuse. Do not print a current-session header, derived session fields, reference
comparison, action menu or save prompt. Consume no action input and sample no
event. A zero failed Session is not an empty successful session. The separate
nonfatal stale-support channel does not trigger this refusal branch.

After successful inspection, print:

```text
Current record-backed session.
Instructions are displayed, not executed. Verified is your explicit statement.
Defaults and suggestions are not accepted answers.
```

Then use the public starter terminal's full Frozen actual session block field
enumeration for Activities/Steps/Questions/Blockers/Recorded outcomes/Support,
starting at Activities. Actual new reviewer fields map one-to-one to those
projections, with no second renderer. OMIT its old read-only header, Invitation-
time guidance and old Session problems list because this API has no archived
guidance or Display field. Instead print all actual method diagnostics as
`Session problems: N`, each `  - Stage: Q(stage) Code: Q(code) Severity: Q(severity)`
followed by fixed `Review the selected source files to resolve any reported errors.`
Do not dump raw Problem.Message/Detail or unknown data. Support.Problems use
the same code/stage projection under label Support problems, retaining severity.
Print `Plan reference matches current bytes: true/false` without rewriting it.
End with exactly:

```text
Choose one change; nothing here runs a command or check.
answer QUESTION text JSONSTRING
answer QUESTION flag true|false
answer QUESTION list JSONARRAY_OF_STRINGS
answer QUESTION deferred
follow-plan QUESTION
outcome SECTION verified|failed|deferred JSONSTRING_NOTE JSONSTRING_DIFFERENTLY
exit
Action: choose one displayed change or exit:
```

Pipe syntax above describes alternatives; literal action lines contain exactly
one choice, not pipes. One ASCII space separates tokens, no extra leading/
trailing whitespace. Nonempty QUESTION/SECTION are UTF8 tokens excluding ASCII
space/tab/CR/LF/NUL. Strict JSON strings/arrays only, no trailing value, invalid
UTF8, unpaired surrogate, object, null or nonstring list item. No trimming,
casefold, implicit Deferred/text conversion, blank/default acceptance or retry.
Empty text/list submits reach actual writer's explicit-intent refusal. `exit`
returns edit.cancelled. Raw syntax/input failures return edit.input.

Actual successful Propose review is:

```text
Selected record change — not saved.
Operation: Q(operation)
Question: Q(question)
Section: Q(section)
```

Answers/follow-plan print V(Before answer,actual before effective Question.Answer),
V(After answer,actual after effective Question.Answer), and
`Stored override after save: true/false` from actual AfterRecord key presence.
For follow-plan print `Following the shared plan may reveal an existing shared answer; it does not promise blank.`
For submit print `Explicit submission stores your answer even when equal to a shared answer.`
Outcome prints Q fields Kind, Current definition hash from actual target row,
D(Note,note), D(Differently,differently), L(Waiting on,actual target WaitingOn),
then `New outcome time: sampled after save` (never empty historical time/sentinel).
Print actual complete After session field enumeration again, so changed activity,
questions/blockers/support/history are not hidden. Its new target On is printed
as Q("sampled after save") only at that target when NeedsEvent=true; old matching
sentinel text remains held. Then diagnostics and exactly:

```text
Only the selected record can change; plan, preparation, registration and first-use data are retained.
Changed records use canonical YAML; a no-op retains exact original bytes.
Private staging can remain after interruption; unknown leftovers are preserved, not replayed or automatically deleted.
Review complete. Choose save or cancel:
```

Only exact save is publication consent. Cancel/ordinary incomplete EOF preserves
everything. Complete final output is exactly `Record saved.` or `Record unchanged.`
and LF, then the fixed stderr report. No post-save input or follow-on action.

Read one logical line bytewise/capability-equivalent with NO read-ahead into its
successor, even into an internal buffer. Maximum65536 UTF8 payload bytes, framing
LF/CRLF excluded; bare terminal CR is payload. Total consumed max1048576 including
framing; do not probe EOF beyond a complete final action. Data+EOF is processed;
data+nonEOF error refuses,0+nil is immediate input error, not an infinite loop.
There are only three reached input lines in this finite route, so total budget
is a defensive bound, not a falsely claimed independently reachable boundary.

Every required display block must finish before the next input byte. Write must
return exact full count and nil error; short+nil,0+error,full-count+error all
fail. If writer implements Flush() error, flush successfully before consuming
the next line. Never bypass an error via embedded Buffer.WriteString or retry a
failed block as though displayed. Borrowed streams/Clock are never closed.
Each separately prebuilt identity/current/review block is capped8388608 bytes
using a bounded builder; no truncation, arbitrary grouping omission or huge
allocation then cap. No session2 source render after save is needed.
