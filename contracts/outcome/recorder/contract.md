# Outcome recorder contract

Status: agent-authored contract; core implemented, driver integration pending.
Component `outcome/recorder`, package `recorder`, position 5.

Authorities: ADR 0085 section-level personal statements, 0090 record ownership,
0140 preparation separation, 0180/0230 observed first use, 0260 current verified
prerequisites, 0270 actual definition hashing; principles 0100, 0110, 0120.
[0290](../../../docs/decisions/0290-explicit-outcome-recording.md) fixes the write profile and refusal policy. Read the section walker and section hasher contracts as complete dependencies.

## Interface and boundaries

`New()` takes no arguments. `Record(effectivePlan, originalRecord, Proposal, on)`
returns a new Record and problems. Proposal contains only exact Section, Kind,
Note and Differently. The injected timestamp is separate. No caller hash, clock,
check result, arbitrary target map, I/O callback or stand-in dependency exists.

At position 5 the component may use the real section walker, real section hasher,
model and its own contract only. The driver already composed the effective plan
from owner-approved plan/preparation and record answers. Do not join answers,
interpret preparation, authorize account/plan identity or confirm first use here.
Full authoring validation remains a caller boundary. Duplicate section/graph
checks and outcome uniqueness are still explicitly enforced through the walker.

Each call is an explicit replacement operation, not a patch to an old outcome:
an empty proposed Note or Differently clears that target's previous value. There
is no removal operation, per-step progress, automatic success, clock lookup or
implicit migration. The proposal and on text are supplied by an authorized driver.

## Refusal phases

Never mutate input, even temporarily. Refusal returns a separately owned copy of
the entire original Record, including original outcome order and nil/empty shapes.
Do not return zero Record, sorted history, an updated target or repaired answers.

Evaluate phases in this exact order; a failing phase stops later phases.

1. Validate proposed Kind and injected on independently, collecting both faults
   in Code byte order. No target text is copied into these diagnostics.
2. Refuse unknown held outcome kinds before calling a walker whose successful
   scope excludes those enums. For every distinct exact Section with any unknown
   Kind report one held-kind diagnostic, sorted by Section. The empty, wrong-case
   and whitespace-padded kinds are unknown. Do not inspect or repair old On,
   SectionHash, notes, first use, or unrelated saved fields.
3. Call the real walker on the same full effective plan and original record.
   If it returns errors, return all its problems unchanged in its own order.
   This includes complete graph diagnostics and duplicate outcomes, even for
   inactive or removed sections. Never choose a duplicate winner.
4. Resolve proposed target against exact original section IDs. Unknown target
   gives one unknown-section error. A declared target absent from the walker's
   active rows gives one inactive-section error. No outcome kind can update
   inactive/removed work. Do not auto-enable it or activate a historical outcome.
5. For proposed verified only, refuse when the target row's current WaitingOn
   is nonempty. Emit one prerequisite error for each waiting ID, in first-held
   distinct Requires order. An old verified outcome on the target does not bypass
   a new attempt. Failed/deferred proposals do not require verified prerequisites.

Real walker warnings do not refuse a write or become extra question, gap,
freshness, date or support gates. Return those warnings on success; in phases
4/5 append the recorder errors after the original warnings. Earlier phases do
not return warnings from dependency calls they never make.

All recorder-added problems have severity error and empty Question. Messages
are fixed below. Detail has exactly the listed keys. Section is empty unless
the row specifies a validated declared target or a held outcome identity.
Messages/detail never copy arbitrary proposal values, timestamp text, notes,
answers, differently text or unknown target strings.

| Code | Section | Detail | Message |
| -- | -- | -- | -- |
| `outcome.kind-invalid` | empty | `field: outcome` | Choose verified, failed, or deferred for the new outcome. |
| `outcome.timestamp-invalid` | empty | `field: on` | Supply a valid UTC outcome timestamp. |
| `outcome.held-kind-invalid` | held outcome Section | `field: outcomes` | Repair the unknown recorded outcome kind before recording another outcome. |
| `outcome.section-unknown` | empty | `field: section` | Choose a section declared in this plan. |
| `outcome.section-inactive` | declared target ID | `field: section` | Choose an active section, or change its applicability explicitly first. |
| `outcome.prerequisite-unverified` | declared target ID | `field: requires`, `required: exact required ID` | Verify the required section before recording this section as verified. |

## Proposed new-write timestamp profile

This is the new-write profile of ADR 0290, not an existing codec restriction. Kind is exactly verified, failed or deferred.
Timestamp must use uppercase `YYYY-MM-DDTHH:MM:SS`, optional period plus one to
nine decimal fractional digits, and either `Z` or `+00:00`. All digits are ASCII;
date is valid proleptic Gregorian with four-digit year 0001 through 9999, hour
00–23, minute/second 00–59. No leap second, lower-case separators, comma fraction,
one-digit hour, space separator, whitespace, missing zone, nonzero offset or
unknown-offset `-00:00` is accepted. A leap-day date follows Gregorian rules.
Preserve accepted timestamp text exactly, including fraction precision and the
choice of zero-offset spelling; do not normalize, trim, round or read a clock.

Old timestamps remain representational data. A malformed old On or FirstOpenedAt
is neither permission to rewrite it nor a gate for an unrelated outcome update;
this component does not interpret those events or compare chronologies. First
use is still exclusively the confirmed new-record driver event of 0180/0230.

## Apply one current explicit outcome

For any accepted kind, find the exact original unresolved Section, not the
walker's resolved copy. Compute its hash with the real section hasher at this
write boundary. Set the new Outcome to precisely Section, Kind, injected On,
that hash, proposed Note and Differently. No Source.Hash fallback, caller hash,
whole-plan reference hash or retained previous outcome hash is allowed. A custom
source-less section is ordinary supported content. Bootstrap and Requires remain
in the current definition. Answers and resolved names are not part of that hash.

Add one outcome if the target was untried, otherwise replace its one entry whole.
Keep every other outcome exactly held, including removed/inactive history, old
bare hashes, notes, differently text, timestamps and previous verified dependents.
Sort the resulting complete Outcomes by exact Section UTF-8 bytes; do not change
the input order. There is still one entry per section, not an append-only event
log. A later record-history format would need a separate decision.

The verified gate uses unique stored verified kinds for direct current Requires,
as supplied by the real current walker. Mismatching or unknown prerequisite
hashes do not make that evidence disappear. A required verified activity may
itself now wait on newly authored work: do not recursively revoke its stored
fact. Parent, After, dates, answers, preparation, first use and file metadata
supply no missing verification. Replacing a prerequisite's outcome with failed
or deferred does not reset previously saved verified dependents, but a later
attempt to verify those dependents must recheck the new record.

## Preservation and purity

Only returned Outcomes may change. Record SchemaVersion, Account, StartDate,
FirstOpenedAt, Plan reference, Answers and UnknownFields remain exactly held.
Preserve hidden answers, Value kinds, list order, nil and empty maps/lists at
every depth; do not reconstruct a record from the effective plan. Return owned
storage for every record map/list/value on success and refusal, including nested
unknown JSON data and Value.List. Changes to a result cannot change either input
or another result; later input changes cannot change a prior result. Return
fresh problem slices and Detail maps too. Immutable strings need no new identity.
Unknown numeric values retain their held numeric kinds, including int, int64
and float64 in Go. A nil unknown map/list is not an empty one. Preserve these
properties through nested maps/lists; a serialization round trip is not a valid
copy when it loses those held kinds or collection shapes.

The component is deterministic, stateless and safe for concurrent calls on one
instance. It reads the full effective plan only through the walker and the
original section definition needed for hashing/target matching, all original
record data for preservation and Outcomes for validation, proposal and injected
timestamp. It performs no filesystem, network, process, environment, Git or
clock operations; sends no message; executes no check and records no step state.
It writes no plan, preparation or outside state. Unknown-field values supported
by record codecs are the preservation scope; cyclic Go graphs are out of scope.

## Independent fixtures and wrong adapters

The manifest gives literal success/refusal oracles with sources. Hash oracles
contain independent canonical JSON and literal SHA-256 digests, not expected
values computed by any production dependency. Tests judge preservation,
isolation, repeat calls, current rechecking and injected timestamp boundaries.
The binding returns the real constructor directly, without copying or translating
its output. Literal calendar boundaries cover every month and leap centuries.

| Plausible wrong implementation | Intended distinguishing test |
| -- | -- |
| Copy Source.Hash, old outcome hash, resolved content hash or a caller map | Independent literal current-definition oracles |
| Require fresh hashes, recurse through a verified prerequisite, or bypass new requirements for a previously verified target | Historical gate/recheck matrix |
| Treat failed/deferred as verified or treat waiting as inactivity | Verified refusal and truthful failed/deferred recording cases |
| Activate removed/off targets, add Parent/After/date/question gates | Target eligibility and non-prerequisite controls |
| First/last-win duplicates or filter graph checks before activity | Complete real walker refusal cases |
| Merge old notes or append duplicate entries | Full replacement, explicit clearing and sorted history cases |
| Backfill first use, rewrite legacy timestamps, or reconstruct record answers | Complete record preservation fixture |
| Share nested maps/lists on success or refusal, keep state, or mutate original history to sort | Isolation, input snapshots and shared-instance concurrency tests |
| Parse a lenient or non-UTC timestamp, round fractions, trim or normalize it | Independent exact new-write timestamp table |

Each row specifies the independent behavior boundary; execution evidence belongs
in a working report, not this contract.
