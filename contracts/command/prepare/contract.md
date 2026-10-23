# Preparation command contract

Component: `command/prepare`.

## Boundary

The command is a filesystem driver for an onboarder's explicit preparation
edits. [0220](../../../docs/decisions/0220-preparation-command.md) defines its
behavior. [0140](../../../docs/decisions/0140-person-specific-preparation.md)
defines ownership, [0190](../../../docs/decisions/0190-incremental-preparation-repair.md)
defines scoped repair, and [0210](../../../docs/decisions/0210-person-specific-preparation-paths.md)
defines person-specific paths. The [API](api.go) is Run(args, stdout, stderr),
returning an exit code. New() returns this interface. Args starts with the
operation, without the binary name or `prepare`. Each call is independent.

Reads: arguments; the explicit plan's complete decoded value for validation;
the plan identity, resultsPath, declared contact IDs, team IDs, and section IDs
for context; selected preparation; and, only for show, the selected record.
Writes: stdout and stderr, and only the selected preparation pathname plus its
missing parent directories and temporary publication file. Plan and record
bytes never change. No clock, environment-derived personal facts, network,
subprocess, repository discovery, working-directory change, or model run.

## Arguments and results

Operations are `init`, `show`, `recipient`, `arrangement`, and `remove`. Each
requires `--plan FILE` and `--account ID`; `--root DIR` defaults to `.`. Support
`--flag value` and `--flag=value`, including explicitly empty values. Flag
names are exact. Unknown operations, flags, positional arguments, missing flag
values, or missing required nonempty plan/account/root values are usage errors.
An account with invalid identity syntax is a data error. Repeated scalar or
boolean flags are usage errors, even when their values agree. Repeated
`--contact` values retain their order and repeats. Parse every argument before
filesystem effects. An initial `--help` returns 0 with fixed usage text on
stdout, empty stderr, and no filesystem reads or writes. Each known operation
also accepts `--help` immediately after its name with the same behavior.

Recipient flags are `--name`, `--company-email`, `--planned-start-date`,
`--team`, `--manager`, `--onboarding-owner`, `--buddy`, `--work-location`, and
`--time-zone`. At least one must be explicitly present. Begin with the saved
recipient, overlay only present flags, then pass the whole recipient to the
typed editor. Explicit empty values clear fields; absent flags preserve them.

Arrangement flags are `--id`, `--section`, `--kind`, `--owner`, repeatable
`--contact`, `--clear-contacts`, `--reference`, `--target-date`, `--status`, and
`--note`. Require an explicitly present id and at least one other arrangement
flag. Begin with the uniquely matching arrangement or a zero new arrangement,
overlay present fields, then call the editor. Contacts replace the entire list;
clear-contacts clears it. Combining clear-contacts and any contact flag is a
usage error, even an empty contact. Clear-contacts accepts only its bare switch
or `--clear-contacts=true`; `=false` and other values are usage errors. No
omitted kind or status is inferred for a new item. Existing arrangement order
is preserved; a new ID appends. Never choose a duplicate ID before editor guards.

Remove accepts only the common flags and an explicitly supplied `--id`.
An empty supplied ID is valid removal input for one malformed empty-ID item;
an omitted ID is a usage error. Init and show accept only common flags. Edit
flags for another operation are usage errors. No command creates implicit
answers, outcomes, statuses, dates, directory entries, or invitation readiness.

Return 0 on success, 2 on usage errors, and 1 on invalid data, refused paths,
decode/validation/encoding problems, or filesystem failures. Failures before
output reporting write no stdout. Failures before publication preserve every
existing artifact byte. Stderr is nonblank,
names the failed operation or required preparation action, and explains the
next repair using fixed field or artifact labels. It never contains supplied
personal values, raw paths, parser excerpts, or operating-system error text.
Successful commands write no stderr. Successful writes print exactly the
relative preparation path followed by LF; show prints one JSON object and LF.
Atomic publication is the commit point. If stdout fails after publication,
return 1 with a fixed diagnostic saying preparation was saved but the result
could not be displayed; keep the published preparation. Output reporting
failures may leave partially written stdout. Show and help output failures
return 1 without file effects. Do not roll back a committed edit.

## Loading and editing

The explicit plan must decode, have a nonempty ID, and pass plan validation
without any error-severity problem. Warnings and unanswered required questions
are allowed. Derive both paths through the locator; invalid custom layouts
never fall back. Context IDs come from all declared entries, including disabled
sections, without accepting defaults or deriving identity from preparation.

Init creates schemaVersion 1, kind preparation, explicit account, and plan ID,
with empty recipient and arrangements. Any existing destination is refused.
Other operations require an existing decodable preparation matching both the
explicit account and plan. Decode/schema/kind/identity failure stops every
operation. Editing uses the editor's duplicate guard and scoped validation,
without full-file semantic validation as a save gate. Preserve unrelated stale
values and unknown root data. Even malformed or mismatched record bytes must
not block an edit: editing does not inspect the record's final component.
Shared parent directories are checked only for preparation path safety.

Show applies complete preparation validation and displays semantic problems
without refusing the view. Duplicate arrangement IDs remain separate items;
they do not select a winner. An existing record must decode and match account
and plan, otherwise fail. A missing record is normal. Do not validate outcome
semantics, reconcile answers, compute changed status, or manufacture progress.

## Show JSON

Object key order and JSON whitespace are free; the key set and values are exact.
The top-level keys are account, planId, preparationPath, recordPath, recipient,
arrangements, preparationProblems, and record. Paths are relative slash paths.
Recipient contains all nine camelCase fields from the preparation format,
including empty strings. Each arrangement contains id, section, kind, owner,
contacts, reference, targetDate, status, and note. Every string is preserved.
Each preparation problem has exactly code, severity, message, and path, in
validator order; path is the validator's Detail.path or empty text. Record has
exactly present and outcomes. Each outcome contains section, outcome, on,
sectionHash, note, and differently in file order. Arrays are always arrays,
including absent contacts, arrangements, problems, and outcomes becoming `[]`.
Exclude unknown preparation/record data, record answers, legacy startDate,
invented actual dates, and first-use events. Outcome notes are intentional
support output; diagnostic privacy rules still apply to errors.

## Filesystem effects

Root must already be a real directory, never a symlink. Every preparation
component, and each record component used by show, must be nonsymlink; final
existing artifacts must be regular files. Plan must be regular and its final
pathname nonsymlink. Use a rooted filesystem capability so concurrent symlink
changes cannot escape root. Reject directories, FIFOs, devices, and sockets
without blocking. No hostile same-owner coordination guarantee is claimed.

Read each input through a bounded reader with a 1 MiB budget plus one sentinel
byte; metadata alone is insufficient. Never write an oversized encoded result.
Finish argument, plan, path, and changed-value validation before creating any
directories. Fresh directories have mode 0700 and new preparation mode 0600.
Existing parent permissions need not change. Encode fully, write a temporary
regular file beside the destination, close it, and publish atomically. Init
must not overwrite a destination appearing concurrently. Edits replace only
the selected pathname and never truncate an inode that may have other hard
links. Cleanup leaves no temporary file after a completed failure or success.
I/O failures may leave newly created empty directories, never a partial final
file. No operation deletes preparation, plan, or record artifacts.

## Independent fixtures and checks

[sources.json](testdata/sources.json) cites every hand-specified fixture.
[prepare_test.go](prepare_test.go) exercises the command through the interface
in per-test temporary roots without changing the working directory.
[wire_test.go](wire_test.go) is the sole implementation import.
Binary scenarios under acceptance/scenarios/prepare-*.yaml exercise ownership
and progress across two accounts. Expected JSON is written independently;
plan and record preservation compares original bytes, not re-encoded values.

## Wrong implementations tried

| Defect | Observable probe |
| -- | -- |
| Return success without writing | Init requires canonical identity bytes and exact path output |
| Reject every operation | Init requires success and a saved preparation |
| Populate recipient from a suggested plan value | Init requires exactly the identity-only preparation bytes |
| Serialize model field names | Exact full-show fixture rejects PlanID instead of planId |
| Drop outcome support notes | Exact full-show fixture preserves the independent outcome note |
| Write null contacts | Exact full-show fixture requires an array |
| Expose private record answers | Exact full-show fixture rejects extra answers fields |
| Truncate an existing preparation inode | Hard-link probe requires the other pathname's bytes unchanged |

The drafting evidence records which probes have actually run; this table does
not claim that an unexecuted test has verified its requirement.
