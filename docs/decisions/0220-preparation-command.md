# 0220. Preparation command

Status: accepted by agent judgment

Builds on [0140](0140-person-specific-preparation.md),
[0190](0190-incremental-preparation-repair.md), and
[0210](0210-person-specific-preparation-paths.md).
Applies P2, P3, P4, P6, and P7 of
[0100](0100-behavior-judgment.md).

## Context

The preparation core can represent, validate, and repair assignments but has no
command that an onboarder can use. Editing the shared plan for each person would
break the three-file ownership boundary. A useful command must also distinguish
an arranged opportunity from evidence that the newcomer performed it.

## Decision

Add `onboard-kit prepare` with init, show, recipient, arrangement, and remove
operations. Every operation requires an explicit plan file and account. The root
directory defaults to the current directory; this is a filesystem context, not
an answer. Derive both person-specific paths through the preparation locator.
The plan must have a nonempty identity and no validation errors; warnings and
unanswered questions do not prevent preparation.

Init creates an incomplete preparation with the explicit account and plan
identity. It neither creates a record nor claims invitation readiness. Refuse an
existing destination rather than resetting or adopting it. Do not populate any
recipient facts from defaults, the clock, environment, or another person's file.

Recipient and arrangement edits use explicit flags. Omitted fields keep their
existing values; an explicitly empty value clears a field. Repeated contact
flags replace the helper list in that order, preserving repeats. An explicit
clear-contacts flag clears it; using that flag together with contacts is an
error. Accept that action flag bare or with exactly =true; reject =false instead
of silently saving a no-op. An arrangement's id selects exact replacement in
place or append; it is not a position or display title. Remove requires an
explicitly supplied id and
may remove a unique malformed empty id as allowed by 0190. No edit infers a
kind, status, helper, date, answer, or completion.

Reject repeated scalar or boolean flags, even with matching values. Contact is
the only repeatable flag. In particular, two account or root flags must not
silently select the last person's destination. Help returns success without
reading or writing artifacts. Reject unknown flags, operations, and positional
arguments without echoing their values.

Use the typed editor's identity and duplicate guards and scoped repair rules.
An edit may succeed while unrelated stale references remain; do not label that
result ready for invitation. Before any edit, require the existing preparation
to match the explicit account and plan. Updating preparation reads no newcomer
record and never changes the plan or any record.

Show is read-only. It displays supplied recipient facts, arrangements, validation
problems, and the current record's outcomes in separate labeled JSON fields.
It does not copy preparation into answers or manufacture outcomes. A missing
record is normal before onboarding begins. An existing record must decode and
match both account and plan, otherwise refuse the combined view. Do not expose
record answers or notes unrelated to the requested progress view. Show includes
outcome notes because they are deliberately recorded context for support; the
operator is choosing to display this person's data, not creating a diagnostic log.
Dates retain their honest names, and passing time changes no status.

## Filesystem boundary

The explicitly selected root must already be a real directory. Reject a symlink
root or any symlink component used for a preparation or record path. Reject
non-regular artifact files, including directories and special files. Inspect
files through a rooted filesystem capability so concurrent symlink changes
cannot escape the selected root. This does not promise to coordinate concurrent
editors or defend against another process with unrestricted write access to the
same root. Same-owner coordination remains the editors' responsibility.

Plan input is a separately selected read-only file, never an output destination.
Read regular input artifacts with a 1 MiB byte budget each, checking an extra
byte rather than trusting metadata alone. Do not follow a symlink plan input.
The command never fetches a URL, executes a command, changes the working
directory, discovers a repository, or writes outside its explicit root.

After successful validation and encoding, write a temporary regular file in the
destination directory, close it, and publish atomically. Creation must refuse a
concurrently appearing destination without overwriting it. Replacement changes
only the selected preparation pathname; do not truncate an existing inode that
may have another hard link. Use private file permissions for newly written
preparation and directories; this does not hide data after it is shared in Git.
Refuse an encoded preparation larger than 1 MiB before writing so the command
does not publish a file that its own bounded reader cannot reopen.
Clean up temporary files on failure. Preflight failures create no directories;
an I/O failure may leave newly created empty directories, never a partial final
file. No operation deletes a person's preparation or record file.

Atomic publication is the commit point. If reporting the successful write then
fails, retain the saved preparation and report that saving succeeded but output
failed. Do not attempt a rollback that could discard a later editor's work.
Failures before publication preserve the existing artifact bytes.

## Diagnostics and verification

Usage errors return 2; invalid data, refused paths, and I/O failures return 1;
successful operations return 0. Diagnostics describe the failed operation and
repair without echoing supplied personal values, file contents, or arbitrary
paths from parser or operating-system errors. Intentional show output is distinct
from diagnostics. Successful edits print the relative preparation path only.

Independent component tests must exercise symlinks, identity guards, explicit
clearing, incremental repair, failure preservation, and exact write boundaries.
Binary scenarios must prepare two people, update an assignment after a record
exists, display outcomes separately, and prove the plan and record bytes stay
unchanged. A successful whole-file validation is not evidence of invitation
readiness, which has a later explicit boundary.

The guided invitation flow, archive import, newcomer record editing, template
bindings, and first-use event remain separate ordered increments. This command
is their preparation interface, not a replacement for that work.
