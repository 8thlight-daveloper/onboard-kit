# Preparation store contract

Component: `preparation/store`.

## Boundary

The store is the file side of the Preparation tab in
[0520](../../../docs/decisions/0520-preparation-screen.md). The screen lists
the people who have a preparation, loads one person, and saves one reviewed
preparation. The store does the file work and nothing else. It carries over the
save rules of [0340](../../../docs/decisions/0340-guided-preparation-save.md) and
[0350](../../../docs/decisions/0350-guided-support-arrangements.md) and the
filesystem boundary of [0220](../../../docs/decisions/0220-preparation-command.md).
It carries over none of the prompts, menus, exit codes, input budget or display
of the retired text guide.

The [API](api.go) has three operations on one `Store`: `List`, `Load` and
`Save`. `New()` takes no dependencies. Calls hold no state, so a `Loaded` value
from one store value works with another. Every call takes explicit arguments
and reads no global, environment, clock, network or process.

Position: an imperative driver at position 7, after `tui`, wired by the binary
at position 8, as [0340](../../../docs/decisions/0340-guided-preparation-save.md)
placed the retired driver. It may depend on `plan/decoder`,
`preparation/decoder`, `preparation/locator`, `model`, the standard library and
the pinned `golang.org/x/sys/unix` facilities. It needs no encoder, editor,
validator or reviewer. The `tui` package does not import it. The walk command
passes it in, as [0520](../../../docs/decisions/0520-preparation-screen.md)
says for file access.

Reads: its arguments, the selected plan file, one person's selected preparation
file, directory entries and file metadata under the selected root. It never
reads a record, another person's preparation body, or anything outside the root
and the plan file.

Writes: only the selected preparation, the private parent directories that were
missing, and its own temporary file in the preparation's directory. It never
writes the plan, a record, another person's file, or anything else. A failed or
refused call may have created a temporary file and removes it. A failed call
after consent may leave newly created empty directories, and says so.

Plan access: the plan is read through the plan decoder and never changed. Every
`model.Plan` and `model.Preparation` the store returns is a fresh value that
shares no memory with the store or with another call.

## Interface

```text
List(ListRequest) ListResult        Root, Plan                  -> Code, People
Load(LoadRequest) LoadResult        Root, Plan, Account         -> Code, Loaded
Save(SaveRequest) SaveResult        Loaded, Reviewed, CreateDirectories
                                                                -> Code, Committed, CreatedDirectories
```

`Loaded` is an interface so each implementation can keep its own snapshots. It
reports `Account`, `Destination` (slash separated, relative to the root),
`Plan`, `Preparation`, `Exists`, `Bytes` and `DirectoriesMissing`. Absent
preparation reports `Exists` false, no bytes, and an identity-only
`Preparation` with schema version 1, kind `preparation`, the account and the
plan ID. Existing preparation reports the decoded value and the exact bytes
read. `Bytes` returns a copy. Changing a returned value changes nothing the
store holds. A `Loaded` holds no open file or directory descriptor, and neither
does any call when it returns, whatever its code: every refusal, success,
unchanged and stale outcome of `List`, `Load` and `Save` leaves the process with
the descriptors it had before the call. The saved snapshots are what `Save` rechecks.

Codes are stable strings. A result carries a code and no display text.

| Code | Meaning |
| -- | -- |
| `store.listed`, `store.loaded` | Success of `List` and `Load` |
| `store.saved` | Preparation published. `Committed` is true |
| `store.unchanged` | Nothing written. The held file keeps its bytes, inode, mode and times |
| `store.path` | Root, destination or account refused, or a symlink, non-directory, special file or alias met on the way |
| `store.source` | Plan or preparation unreadable, over 1,048,576 bytes, not UTF-8, or refused by its decoder |
| `store.plan` | The plan has no ID |
| `store.identity` | The preparation is for another account or another plan |
| `store.loaded-state` | `Save` got a `Loaded` that this store did not issue, or nil |
| `store.draft` | Reviewed bytes are empty, do not decode, or are for another account or plan |
| `store.too-large` | Reviewed bytes exceed 1,048,576 bytes |
| `store.stale` | A file changed after `Load`. Reload and review again |
| `store.directories-missing` | Directories must be created and `CreateDirectories` is false |
| `store.volume` | The volume that will hold the file, whether the root's or one mounted at a directory on the way to the destination, cannot keep the file private, because it cannot answer the access control question (FAT and exFAT, such as a USB stick). `Save` refused and nothing was written |
| `store.commit` | Refused before publication. Original files are intact |
| `store.cleanup-refused` | Refused, and the owned temporary file could not be removed. `Committed` is false |
| `store.cleanup-saved` | Published, and the owned temporary file could not be removed. `Committed` is true |

`SaveResult.Committed` is true exactly for `store.saved` and
`store.cleanup-saved`. `CreatedDirectories` is true when this call created at
least one directory, whatever the code.

## Load

Phases, stopping at the first failure. A refusal returns a nil `Loaded` and
changes nothing, creates nothing, and writes no temporary file.

1. The root must be an existing real directory. A symlink as root or in any
   root component, a NUL, a `..` component, or an empty root gives `store.path`.
   A root that is a pipe or another special file gives `store.path` and the call
   does not block. A relative root is accepted and resolved against the process's
   working directory at the time of each call, with `.` components ignored; the
   other rules apply to it unchanged. A control byte other than NUL, such as
   `\x01` or a newline, is an ordinary name byte and is not refused.
   A root is an existing real directory whether or not the caller may search it,
   so a root such as `<directory>/.` for a directory with mode 0600 or 0400 passes
   this phase, and a missing plan then gives `store.source` from `Load` and `List`.
   A directory name that starts with a dash, a space or another punctuation mark
   is an ordinary name as well.
2. Read the plan. It must be a regular file with no symlink in its path, no
   `..` component, no NUL, and no trailing separator. A relative plan is accepted
   and resolved as a relative root is. Control bytes other than NUL are ordinary
   name bytes. Read at most 1,048,576
   bytes and check one extra byte, so exactly that size loads with every byte and
   one byte more does not. Valid UTF-8 is required. Any failure, or any plan decoder error,
   gives `store.source`. A plan with no ID gives `store.plan`. Plan validation
   is not the store's job.
3. Ask the locator for the preparation path with the plan's results path and the
   account. Any locator problem, including an invalid account, gives `store.path`.
4. Walk the destination under the root without following symlinks. A missing
   directory or file is not an error: the preparation is absent, and
   `DirectoriesMissing` is true when a directory is missing. A symlink at any
   component, a component that is not a directory, a destination that is not a
   regular file (directory, pipe, device), or a destination that is the same file
   as the plan gives `store.path`. Opening a pipe must not block, whether the pipe
   is the destination or stands where a directory should be, such as the
   `records` directory or the person's directory.
5. Read an existing preparation with the same bounds as the plan. Failure to
   read, size, UTF-8 or any preparation decoder problem gives `store.source`. A
   regular file that cannot be opened for reading, such as one with mode 000,
   is a failure to read and gives `store.source`, not `store.path`.
6. A decoded account or plan ID that differs from the requested account or the
   plan's ID gives `store.identity`, whether the held value sorts above or below
   the requested one. The store never repairs identity.

Repeated arrangement IDs are not the store's concern. The editor refuses edits
on them, as its [contract](../editor/contract.md) says.

## Save

`Save` takes the loaded state and the reviewed bytes. The screen encodes the
draft with the preparation encoder, shows the user what those bytes mean, and
passes exactly those bytes. The store writes the bytes it was given and never
re-encodes, so the reviewed bytes are the saved bytes.

Phases, in this precedence:

1. A `Loaded` that is nil or not issued by this store gives `store.loaded-state`.
   So does a nil pointer of the type this store issues, held in the interface:
   the call never panics.
2. Reviewed bytes over 1,048,576 give `store.too-large`. Exactly that size is
   accepted. Empty bytes, bytes the preparation decoder reports a problem for,
   or bytes whose account or plan ID differ from the loaded ones, whether they
   sort above or below, give `store.draft`. No problem about a stale reference is raised here: stale
   references are decoder-valid and may be saved as a scoped repair.
3. Recheck the files (below). Any observed change gives `store.stale`. The
   recheck comes before the consent check and before any directory is created.
4. If a preparation exists and the reviewed bytes decode to a value deeply equal
   to the loaded preparation, nothing is written and the code is
   `store.unchanged`. The held bytes, inode, mode, owner and times stay. This
   holds when the original is not canonical and when the encoder could not
   encode it. The screen passes `Loaded.Bytes()` when it cannot encode its draft.
   An absent preparation is never unchanged: an identity-only draft is created.
5. If directories are missing and `CreateDirectories` is false, give
   `store.directories-missing`. Nothing is created. The screen asks for consent
   and calls again.
6. Publish (below).

### The recheck

The loaded snapshots are the root directory's identity (device and inode), the
plan's bytes, inode, mode, owner and group, and the preparation's same facts or
its absence. The recheck reopens each by path with the same no-symlink rules as
`Load`. The root's identity is compared even when the plan lives outside the
root and the preparation is absent, so a root replaced by a fresh empty
directory at the same path is stale. A relative root or plan is reopened against
the working directory at the time of `Save`; if that directory moved and a path
no longer opens, the state is stale. It reports a change when any snapshot differs. The same bytes in a new
file count as a change. A mode change counts. An absent destination that now
exists, an existing one that is gone, a path component that became a symlink, a
replaced root, and any failure to reopen all count. This is a cooperative editor
guard, as [0340](../../../docs/decisions/0340-guided-preparation-save.md) says,
not a promise against a hostile process with the same rights.

A changed draft is rechecked twice: before any directory is created, and again
after the temporary file is closed, immediately before publication. An unchanged
draft is rechecked once. A save that cannot pass the second recheck publishes nothing,
removes its temporary file, and leaves the files as the other process left them.
The same loaded state cannot save twice: after a save the files no longer match,
so a second `Save` gives `store.stale` and a fresh `Load` is needed.

### Publication

1. Create each missing directory under the root with mode 0700, with no symlink
   followed. A missing directory is created only after consent and a passing
   recheck. A new directory must keep no group or other access bits and must not
   inherit an access control grant.
2. Create an owned temporary file in the destination's directory with mode 0600,
   exclusive creation and no symlink following. The name is not part of the
   contract. A file in the destination directory that this call did not create is
   never replaced, truncated or removed, and does not by itself make `Save` fail.
3. Write all the reviewed bytes, checking the byte count and the close.
4. Recheck. Then publish: for a new preparation, with an operation that fails if
   the destination now exists; for an existing one, by replacing the selected
   pathname only. The old inode is never truncated or written, so another hard
   link to it keeps its old bytes.
5. Remove the temporary file if it still exists. A file already absent counts as
   cleaned. A replace leaves none to remove.

The saved file's permission bits are the intersection of 0600 and the existing
file's bits, so a new file is 0600, an existing 0644 or 0640 becomes 0600, and an
existing 0400 stays 0400. If the file cannot be saved with those bits, the code is
`store.commit` and the original stays. The store never grants more access than
the existing file had. It never changes the owner or group: if the new file would
not have the existing owner and group, the code is `store.commit`. A mismatch of
group, between the existing file and its directory in either direction, is
refused with `store.commit`, leaves the original file as it was, and leaves no
temporary file.

A volume that cannot answer the access control question at all, such as FAT or
exFAT, can honor neither mode 0600 nor an access list. The volume that counts is
the one that will hold the file: the root's, or another volume mounted at a
directory on the way to the destination, such as the person's own directory.
`Save` there gives
`store.volume` instead of `store.commit`, so the caller can tell the person that
the volume cannot keep the file private. The check comes at publication, after
the earlier phases: a stale state, an unchanged draft and a missing consent keep
their own codes. A `store.volume` refusal creates no directory and no
preparation, leaves the held file as it was, leaves no temporary file, and has
`Committed` and `CreatedDirectories` false. `Load` and `List` read such a volume
as any other and are unchanged by it.

Access control lists are preserved or refused, never stripped or broadened. On
macOS an existing preparation with an access control entry, or with an explicit
empty list, either keeps exactly that list on the saved file or the save is
refused with `store.commit` and the original stays. A parent directory whose
inheritable grants would give the new file more access than an ACL-free
preparation had must not pass that grant on. The store refuses with `store.commit`
instead of removing the parent's list. An unchanged call needs none of this and
never writes. No syscall or algorithm is required by these rules.

### Cleanup and reporting

Every path checks the owned temporary file. The report matches the disk:

- Removed or never left: the plain code.
- Not removed before publication (the save was refused for any reason):
  `store.cleanup-refused`, `Committed` false, the destination as it was.
- Not removed after publication: `store.cleanup-saved`, `Committed` true, the
  destination holds the reviewed bytes. Never delete the saved preparation, retry
  without end, or relax access to finish the cleanup.

Cleanup failure takes precedence over the refusal code that preceded it. A
refusal that leaves a temporary file never reports a plain refusal code. A
successful save never reports `store.saved` while the file remains.

A call that returns after consent and a failure may have created empty
directories. `CreatedDirectories` says so. A refusal that created none reports
false and leaves the filesystem exactly as found.

## List

`List(Root, Plan)` returns the people who have a preparation under the root,
without reading any body. It checks the root and plan as `Load` does, with the
same codes. The locator must accept the plan's results path with a probe account,
or the code is `store.path`. The plan needs an ID, or the code is `store.plan`.

A person is listed when the preparation path for their account, as the locator
derives it for the plan's results path, exists as a regular file reached through
no symlink. The empty results path means `records/{account}/record.obk.yaml`.
Walking the template, a directory whose name matches the first directory
component that holds `{account}` yields a candidate: the literal parts around
each token must match, and every token must read the same account. The account
must satisfy the locator. Symlinked or non-directory candidates, directories
without a preparation, a preparation that is a symlink, directory or pipe, and
directory names that are not valid accounts are skipped without error. A
directory name is a candidate only when it equals the template with every token
replaced by one account: a name shorter than the literal parts, a name that
differs from them in any byte, and a name with an extra tail are skipped without
error, and never listed under another name's account. A pipe or other special
file in the place of a candidate is skipped without blocking. A damaged
or foreign preparation body is still listed, because `Load` is the call that
reports it. A missing directory above the candidates is an empty list. A symlink
or non-directory in the fixed directories above the candidates is `store.path`,
and so is a pipe there. A root that is a pipe is `store.path`.

A preparation that exists as a regular file but cannot be opened for reading,
such as one with mode 000, is still listed. `List` reads no body, so it cannot
call the file damaged, but it records the problem on the person. Its `Problem`
is `store.source`, the code `Load` gives for that account. A person whose file opened has an
empty `Problem`, and so does one whose body is damaged or foreign, because `Load`
reports a body.

People come back in ascending byte order of account. Each has its account, its
relative destination and its problem. A listed person can be loaded with the
same arguments, and `Load` refuses with the person's `Problem` when it is not
empty.

## Decisions to know

Each item is tagged `cheap` when changing it later touches only the store, or
`hard` when it changes what the screen or another contract must do.

01. `hard`: `Save` takes the reviewed bytes, not a draft value. The screen
    encodes before it asks for consent, so the reviewed bytes are exactly the
    saved bytes by construction. The store needs no encoder. The alternative,
    `Save(loaded, draft value)` with encoding inside, would let a saved file differ
    from what was reviewed whenever encoder output changes between review and save.
02. `hard`: an unchanged draft is a draft that decodes equal to the held
    preparation, not one with identical bytes. Then nothing is written and the
    original bytes stay (0340). The screen can pass `Loaded.Bytes()` when it
    cannot encode its draft.
03. `hard`: the store sits at position 7 as a driver and the screen does not
    import it. The store depends on the plan decoder, the preparation decoder,
    the locator and the model only.
04. `hard`: the store decodes the plan but does not validate it. A plan with
    validation errors can be loaded, as long as it decodes and has an ID. The retired
    guide refused such plans. The walk already validates before it opens the tab.
05. `hard`: stale means any change to the root, the plan or the preparation
    identity, mode, owner, bytes or existence. A same-bytes rewrite into a new
    file, which some editors do on save, makes a loaded state stale.
06. `hard`: a published file has mode 0600 intersected with the existing mode, and
    an ACL on the existing file or inherited from the parent is preserved or the
    save is refused. A person with shared or ACL-protected folders will see
    `store.commit` instead of a widened file.
07. `hard`: a refused save and a committed save are separate cleanup codes
    (`store.cleanup-refused`, `store.cleanup-saved`). The screen must tell the
    user whether the preparation was saved.
08. `cheap`: consent to create directories is a flag on `Save`. `Load` reports
    `DirectoriesMissing` and creates nothing.
09. `cheap`: `List` never opens a body, sorts by account bytes, and skips what it
    cannot treat as a person.
10. `cheap`: results hold codes only. Wording belongs to the screen.
11. `cheap`: duplicate arrangement IDs are the editor's check, not the store's.
12. `cheap`: the temporary file's name and the publication syscalls are free.
    The store sets only where the temporary file lives (the destination's
    directory) and that it is gone when the call returns.
13. `cheap`: the store keeps no state and takes no clock, so two calls at once
    are safe unless they save the same person, which is the stale guard's case.
14. `hard`: a relative root or plan is accepted and resolved against the working
    directory on each call, as the retired guide did. The alternative, refusing
    relative paths with `store.path` and `store.source`, would force the binary
    to make paths absolute before it calls. The safety rules apply to relative
    paths unchanged. A `Loaded` keeps the path as given, so moving the working
    directory before `Save` makes it stale.
15. `cheap`: only NUL is refused in a root or plan path. Other control bytes are
    name bytes, because the filesystem accepts them and a refusal would hide a
    real directory from the person.
16. `cheap`: an identity mismatch is any inequality, not an ordering, so a held
    file or draft for a lower or higher account or plan ID is refused alike.
17. `cheap`: a group mismatch between an existing preparation and its directory
    is refused rather than repaired, so the store never changes a group. The
    owner half of the rule is stated and untested.

## Fixtures and tests

Fixtures are hand-written and each one cites its source in
[testdata/sources.json](testdata/sources.json); list cases cite theirs in
[testdata/cases.json](testdata/cases.json). Expected saved bytes are the fixture
files themselves, never computed by the product encoder. The tests use real
files in a per-test temporary directory (symlinks resolved first), real
hard links, symlinks and pipes, and the real decoders and locator through the
implementation.

| File | Covers |
| -- | -- |
| [load_test.go](load_test.go) | Held state, refusal codes and order, size bounds with every byte at the limit, copies, no writes |
| [save_test.go](save_test.go) | Exact bytes, no-op, consent, double save, modes, hard links, foreign state |
| [safety_test.go](safety_test.go) | Symlinks, traversal, special files, plan alias, escape after load |
| [stale_test.go](stale_test.go) | Every stale cause, order against consent, change during the write |
| [list_test.go](list_test.go) | People, order, skips, layouts, refusals, a preparation that cannot be opened, a directory or pipe with mode 0 in its place |
| [cleanup_darwin_test.go](cleanup_darwin_test.go) | Report matches what is left behind, and a save over a file with the user-immutable flag is refused |
| [acl_darwin_test.go](acl_darwin_test.go) | Existing, empty and inherited ACLs on macOS |
| [directories_darwin_test.go](directories_darwin_test.go) | Refused directory creation under an inheritable ACL or in an unwritable directory: no inherited grant is left and `CreatedDirectories` matches the disk |
| [descriptors_darwin_test.go](descriptors_darwin_test.go) | Every `List`, `Load` and `Save` outcome in the refusal tables returns the process to its descriptor baseline |
| [selection_test.go](selection_test.go) | Relative root and plan, control bytes and punctuation-led names in paths, root replaced with the plan outside it |
| [ownership_darwin_test.go](ownership_darwin_test.go) | Group mismatch between a preparation and its directory |
| [volume_darwin_test.go](volume_darwin_test.go) | `Save` on FAT and exFAT disk images made with `hdiutil`, as the root and as a volume mounted on the person's directory: `store.volume`, nothing written, `Load` and `List` unchanged |
| [leftover_test.go](leftover_test.go) | Files the call did not create, found under the temporary names the store used one after another, survive a save that still succeeds |
| [wire_test.go](wire_test.go) | The only import of the implementation, through `New()` |

Two tests use timing. The stale-during-write test fills the reviewed bytes to
1 MiB, watches the directory for the temporary file, and changes a file when it
appears. The cleanup test sets the macOS immutable flag on the temporary file the
same way. Both run 20 attempts and fail if the condition was never reached, and
check every attempt for consistency. They are probabilistic by nature. The tests
that change the working directory and count descriptors run alone, before the
parallel tests, because both are process-wide.

The group mismatch test needs a second group the current user belongs to, and
skips with a message when there is none. The volume test makes an 8 MiB image per
case with `hdiutil`, mounts it without privileges, and always detaches it; it
skips with the reason when `hdiutil` cannot create or attach an image. The
leftover test finds the temporary name by watching the directory during a 1 MiB
save, as the two timing tests do. It repeats that with each name found so far
already occupied, until five names are occupied, then saves again in a fresh
directory holding all five. It fails if no attempt saw a new name. A build with random names passes
it without a collision. The unreadable preparation test skips when run as root.

## Wrong implementations tried

Each row names a class of wrong build and the tests that fail it.

| Wrong implementation | Caught by |
| -- | -- |
| Returns nothing for every call | Every test that checks output |
| Follows a symlink as the root, a parent directory, or the plan | TestUnsafeSelectionsAreRefusedWithoutFollowing, TestSaveNeverFollowsASymlinkIntoAnotherDirectory, TestListCases |
| Accepts `..` or NUL in the root or plan | TestUnsafeSelectionsAreRefusedWithoutFollowing, TestListRefusals |
| Skips the stale recheck | TestStaleLoadedStateIsRefusedBeforeAnyWrite, TestExternalEditorBetweenLoadAndSaveIsNotOverwritten, TestSavedLoadedStateCannotSaveAgain |
| Compares bytes only, or skips the plan, or ignores an appearing destination | TestStaleLoadedStateIsRefusedBeforeAnyWrite |
| Rechecks only before the temporary file is written | TestStaleCheckRunsAgainAfterTheTemporaryIsWritten |
| Publishes, then rechecks | TestExternalEditorBetweenLoadAndSaveIsNotOverwritten, TestStaleCheckRunsAgainAfterTheTemporaryIsWritten |
| Creates directories before the recheck, or before consent | TestStaleCheckHappensBeforeDirectoriesAreCreated, TestSaveNeedsConsentBeforeCreatingDirectories |
| Checks consent before staleness | TestStaleCheckComesBeforeTheConsentCheck |
| Skips the recheck for an unchanged draft | TestStaleLoadedStateIsRefusedBeforeAnyWrite |
| Truncates the existing inode | TestSavedFileKeepsOnlyTheSelectedPathname |
| Weakens or copies a mode instead of intersecting | TestSavedModesNeverBroaden |
| Treats a temporary file already absent as a failure | Every test that expects `store.saved` for a replace |
| Rewrites an unchanged draft, by new file or in place | TestUnchangedDraftKeepsOriginalBytesAndInode |
| Treats only byte-equal drafts as unchanged | TestUnchangedDraftKeepsOriginalBytesAndInode |
| Leaves the temporary file after a stale refusal | TestStaleCheckRunsAgainAfterTheTemporaryIsWritten |
| Ignores a cleanup failure and reports the earlier code | TestCleanupReportMatchesWhatIsLeftBehind |
| Creates directories with group or other access | TestSavePublishesExactlyTheReviewedBytes, TestPublishedFileIsNotReadableByOthers |
| Creates directories during `Load` | TestLoadChangesNothing, TestLoadReportsHeldState |
| Reports `CreatedDirectories` wrongly | TestSaveNeedsConsentBeforeCreatingDirectories, TestConsentWithNothingToCreateCreatesNothing |
| Lists by following symlinks, unsorted, or only decodable bodies | TestListCases |
| Skips the size bound or is off by one | TestSourceSizeBoundary, TestSaveAcceptsExactlyTheSizeLimit, TestSaveRefusesReviewedBytesThatAreNotThisPersons |
| Accepts reviewed bytes for another account or plan | TestSaveRefusesReviewedBytesThatAreNotThisPersons |
| Accepts a preparation that is a hard link of the plan | TestUnsafeSelectionsAreRefusedWithoutFollowing |
| Blocks on a pipe | TestUnsafeSelectionsAreRefusedWithoutFollowing, TestListCases |
| Holds the root open after `Load` | TestNoDescriptorOutlivesACall, TestEveryLoadAndListOutcomeReturnsDescriptors |
| Leaks a descriptor on a refusal: missing, malformed or ID-less plan, flat results path, plan hard link, destination that is a directory or pipe, oversize file, or any directory walk | TestEveryLoadAndListOutcomeReturnsDescriptors, TestEverySaveOutcomeReturnsDescriptors |
| Opens directories without requiring a directory, so a pipe as the root, as `records` or as the person's directory blocks | TestUnsafeSelectionsAreRefusedWithoutFollowing, TestListRefusals, TestListCases, TestEveryLoadAndListOutcomeReturnsDescriptors |
| Treats a relative root or plan as absolute | TestRelativeRootAndPlanAreResolvedAgainstTheWorkingDirectory, TestRelativePathsKeepTheSafetyRules, TestRelativeLoadedStateIsStaleWhenTheWorkingDirectoryMoves |
| Refuses a control byte other than NUL in a root or plan | TestControlBytesOtherThanNulAreOrdinaryNameBytes |
| Skips the root identity check when the plan is outside the root and the preparation is absent | TestReplacedRootIsStaleWhenThePlanLivesOutsideIt |
| Saves an existing preparation with a group that differs from its directory's group | TestExistingPreparationWithAnotherGroupIsNeverSavedWithTheWrongGroup |
| Accepts a held preparation or a draft whose account or plan ID sorts below the requested one | TestLoadRefusals, TestSaveRefusesReviewedBytesThatAreNotThisPersons |
| Truncates a file at the size limit | TestLoadReturnsEveryByteOfAFileAtTheSizeLimit, TestSaveAcceptsExactlyTheSizeLimit |
| Panics on, or lists under a second account, a directory name shorter than or differing from the template | TestListCases |
| Returns its internal bytes from `Loaded.Bytes` | TestLoadedValuesAreIndependentCopies |
| Refreshes the loaded state after saving | TestSavedLoadedStateCannotSaveAgain |
| Accepts a plan with no ID | TestLoadRefusals, TestListCases |
| Creates a directory under a parent with an inheritable ACL, or reports `CreatedDirectories` true for a refusal that created none | TestDarwinRefusedDirectoryCreationUnderInheritableACLLeavesNothingGranted |
| Ignores a failed directory creation, or reports it as created | TestRefusedDirectoryCreationInAnUnwritableRootReportsNothingCreated, TestRefusedDirectoryCreationBelowAnUnwritableRecordsDirectoryReportsNothingCreated |
| Leaks a descriptor when a group mismatch refuses a save | TestRefusedGroupMismatchSaveReturnsDescriptors |
| Skips or misreads root path components whose names sort before a dot | TestRootAndPlanBelowDirectoriesWithUnusualNamesAreOrdinary |
| Answers `store.commit` for a volume that cannot answer the access control question, or saves there as if no list existed | TestDarwinVolumeThatCannotKeepAFilePrivateRefusesSaveWithItsOwnCode |
| Judges only the root's volume, so a FAT volume mounted on the person's directory gives `store.commit` | TestDarwinVolumeMountedWhereTheFileIsWrittenRefusesSaveWithItsOwnCode |
| Gives `store.volume` for an unchanged draft, or creates a directory before it refuses | TestDarwinVolumeThatCannotKeepAFilePrivateRefusesSaveWithItsOwnCode |
| Uses one fixed temporary name with exclusive creation and no retry, or stops retrying after the first collision, so a leftover blocks every save | TestSaveNeverTouchesAFileItDidNotCreateAndStillSaves |
| Truncates, replaces or removes a file it did not create that sits under its temporary name | TestSaveNeverTouchesAFileItDidNotCreateAndStillSaves |
| Omits a preparation it cannot open from `List`, or lists it without a problem, or gives a problem that differs from `Load`'s | TestListShowsAPersonWhosePreparationCannotBeOpenedWithTheLoadProblem |
| Lists a directory or a pipe with mode 0 at the preparation path | TestListSkipsANonRegularEntryAtThePreparationPathWhateverItsMode |
| Reports `store.saved` when the replacement of an immutable file fails | TestSaveOverAFileThatCannotBeReplacedIsRefusedWithoutACommit |
| Gives a problem to a person whose body is damaged or whose file opened | TestListCases, TestListShowsThePersonLoadOpens |
| Panics on, or accepts, a typed nil `Loaded` of the store's own type | TestSaveRefusesLoadedStateItDidNotIssue |
| Refuses a root written `<directory>/.` when the directory has no search bit, with `store.path` instead of reaching the plan | TestRootThatCannotBeSearchedIsStillARootWhenWrittenWithATrailingDot |
| Ignores ACLs on the existing file or the parent | TestDarwinExistingAccessControlIsPreservedOrRefused, TestDarwinExplicitEmptyACLIsNotNoACL, TestDarwinInheritedACLDoesNotBroadenPrivatePublication |

## Limits

These rules cannot be forced by a test without a hook in the public API or a
second user, and so are stated and not tested:

- `store.cleanup-saved` cannot be forced, so only `store.cleanup-refused` is
  tested. The cleanup test checks that a reported code matches the disk.
- A mismatch of the owner between an existing preparation and the caller needs a
  second user. The rule is that the store never changes the owner and refuses
  with `store.commit`.
- A new file is published with an operation that fails if the destination now
  exists. A rename in its place differs only in a window of a few microseconds
  after the last recheck.
- A read error on a regular file that was opened and is being read gives
  `store.source`. The call never keeps the bytes read so far as a snapshot, loads
  from them or saves against them. No test can force a read error on such a
  file.
- A plan file directly under `/`, such as `/plan.yaml`, loads like any other
  absolute plan. No test can create one without privilege.
- A short write or a failed close gives `store.commit` before publication or
  `store.cleanup-refused` when the temporary file stays. No test can force
  either.
