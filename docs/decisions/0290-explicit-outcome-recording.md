# 0290. Explicit outcome recording

Status: accepted by agent judgment; core implemented, driver integration pending

Builds on [0085](0085-walking-sections-and-recording-outcomes.md),
[0090](0090-plan-and-onboarding-record.md),
[0140](0140-person-specific-preparation.md),
[0230](0230-first-use-record-compatibility.md),
[0260](0260-prerequisites-and-action-order.md), and
[0270](0270-outcome-definition-hashes.md).
Defines the write operation separately from the read-only section walker.

## Context

An outcome is the newcomer's explicit statement about one activity. Reading a
plan, showing a changed-definition flag, arranging support or passing a planning
date must not manufacture that statement. A write must check current prerequisites
without rewriting unrelated history or borrowing a template baseline hash.

The runtime needs a precise operation and refusal boundary. The record codecs
preserve timestamp text; that does not decide which newly supplied timestamps
the outcome writer accepts. Likewise, the walker excludes unknown stored outcome
kinds from its successful-view guarantee, so a writer must not invoke it outside
that boundary and silently interpret a new kind.

## Decision

### Replace one explicit current outcome

Use a pure outcome recorder over an already composed effective plan, the original
record, an explicit proposal, and an injected timestamp. The proposal has only
section ID, outcome kind, note and differently text. It supplies no hash.
The authorized driver owns identity confirmation, answer composition, visibility
notices and eventual recoverable publication; the core grants none of them.

Only an active declared section may receive a new outcome. An inactive or removed
section's history remains held, but the writer does not reactivate it to update
an outcome. Unknown targets are refused without echoing arbitrary supplied text.
This is not a history-removal or account-move operation.

The proposal replaces that target's complete outcome. Empty note or differently
text deliberately clears only that target's old text, not unrelated notes.
Add a first outcome or replace its one existing entry, then sort the complete
result by exact section ID. Compute the actual unresolved current section hash
with the real hasher at the write boundary under 0270. Never use Source.Hash,
the old outcome hash, rendered text or a caller-supplied hash.

### Recheck evidence without rewriting history

Call the real walker on the full current inputs. Graph errors and duplicate
outcomes refuse all writes, including duplicates for removed or inactive work.
A new verified outcome requires every direct current prerequisite to have a
unique stored verified outcome. A mismatching or unknown prerequisite hash still
counts; its current-definition flag remains a separate fact. Do not recursively
revoke stored verification or reset previously verified dependents.

Failed and deferred proposals may record a truthful attempt while waiting.
They do not satisfy another activity's prerequisite. The writer adds no hidden
freshness, Parent, After, date, question, support or earlier-action gate. Real
walker warnings remain useful information and may accompany a successful write;
they are not silently converted to errors. Driver question/authoring workflows
remain separate from this explicit recording operation.

### Keep old data and new-write validation separate

Validate the proposed kind and timestamp first, reporting both faults in code
order. Then refuse unknown held outcome kinds before invoking the walker. Report
one problem per affected exact section ID, without choosing a duplicate winner.
Next, pass through the walker's complete errors; finally check target activity
and the new verified gate. Later refusals retain preceding walker warnings.

The three kinds remain exactly verified, failed and deferred, with no trimming
or case repair. Introduce this explicit profile for new injected timestamps:

- Four-digit Gregorian year 0001 through 9999 and a valid calendar date.
- Uppercase T, hours 00 through 23, and minutes/seconds 00 through 59.
- Optional full stop followed by one through nine ASCII fractional digits.
- UTC zone written exactly Z or +00:00.
- No whitespace, leap second, unknown -00:00 offset, nonzero offset, comma
  fraction, lower-case separator or one-digit hour.

This is narrower than every RFC 3339 spelling, not an existing codec restriction.
Preserve accepted text exactly, including fractional precision and zero-offset
spelling. The core reads no clock and makes no chronology or attendance claim.
Do not interpret, normalize or reject unrelated held On or FirstOpenedAt values
during an outcome update. In particular, do not backfill first use or relabel
legacy startDate. The confirmed new-record driver remains responsible for first
use under 0230.

### Preserve the whole record on both paths

Success changes only returned Outcomes. Refusal returns an independently owned
exact copy of the original record, including its original unsorted history and
nil/empty collection shapes. Preserve answers, hidden answers, metadata, plan
reference, first use, unknown data and every unrelated outcome. Never sort or
repair an input temporarily. Returned nested storage and problem maps are owned
independently of inputs and other calls.

## Rationale and alternatives

P2 and P4 of [0100](0100-behavior-judgment.md) favor explicit ownership and local
recoverable changes. P3 favors truthful waiting/support information, while P6
keeps dates from becoming progress punishment. P7 limits diagnostics to the
validated identity and repair information rather than copying notes or answers.

Merging empty proposed notes with old text would make clearing unpredictable.
Updating inactive history through an activity action would imply applicability
the current plan denies. Treating an unknown stored kind as verified or failed
would invent semantics; refusing preserves the data for a distinct repair flow.
The stricter new timestamp profile excludes some uncommon representations but
changes no saved format and destroys no old data.

## Component boundary and verification

Outcome recorder is dependency position 5. It may depend on section walker at 4,
section hasher at 1, model and its own contract. It reads the effective plan,
original record, explicit proposal and injected timestamp. It returns a copied
record and diagnostics, writes no input or outside state, and performs no
filesystem, network, process, environment, Git or clock operation.

Independent fixtures must distinguish current hashing from source/old hashes,
direct from transitive verification, new attempts from held history, warnings
from gates, all refusal phases, exact timestamp spellings and complete ownership
on success and refusal. Binary integration must separately prove explicit action,
record-only publication, preserved plan/preparation bytes and no writes on
cancellation or refusal. A pure recorder does not deliver the runtime by itself.
