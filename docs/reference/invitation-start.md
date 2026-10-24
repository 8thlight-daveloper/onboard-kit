# Invitation opening

The standalone tool can review an invitation ZIP, explicitly begin its opening,
and display a read-only session backed by the person's record. It runs no checks
or instructions and saves no answers or outcomes during that session.

## Open and review

Choose an existing owned record root and an application-state location:

```sh
onboard-kit start --invitation invitation.zip --root workspace --home private-state
```

The root must already exist. Relative paths use the invocation's working
directory. The application-state home is optional; its platform default or
OBK_HOME may be used instead. An invalid explicit home never triggers fallback.
Supported paths and native ownership/access controls are checked rather than
silently repaired.

The tool shows the account, destinations and privacy notice. Enter the displayed
exact identity confirmation to inspect the complete proposal. That confirmation
does not authorize publication. Only a complete `begin` action after review
permits the approved changes. `cancel` or ordinary EOF before begin retains
existing files. A partial action at EOF does not count as consent.

Review distinguishes existing work, incoming content and any explicit plan
choice. Existing local preparation is retained, including stale support that
needs repair. Existing personal answers, notes, outcomes and unknown data are
not replaced from an identity-only invitation seed. Defaults remain suggestions.
Use the same archive, root and home to review a later reopening; do not delete
old state to bypass a conflict.

## Files and first use

After approved begin, the starter publishes the local organization-plan copy,
person-specific preparation, session registration and record. The plan copy and
registration live under the private application-state home; preparation and
record use their declared destinations under the selected root. Unrelated files
and the source archive are not changed. The tool does not run Git, provision
accounts, send invitations or obtain company access.

For a genuinely new record, firstOpenedAt records this confirmed beginning.
It is not an employment start date. Existing unknown or held first-use events
remain unchanged, and recovery never samples a replacement time. Planned dates
remain goals, not a reason to fail or blame the newcomer.

The displayed session separates questions, suggestions, useful instructions,
support, prerequisites and held outcomes. Enter `exit` to leave it. Viewing
does not accept defaults, save a question answer or record verification.

Use the separate [record editing command](session-edit.md) to propose and save
one answer or outcome after opening. Invitation start itself remains read-only
during its displayed session.

Records and recovery copies contain personal data. Anyone who can read those
files can read their contents; file ownership is not encryption. Keep private
state out of shared repositories and exclude credentials and unnecessary
personal details from records and terminal transcripts.

## Recover an interrupted opening

An interrupted approved opening may leave recoverable state. The fixed report
identifies the state and a transaction selector. Use the same root and home:

```sh
onboard-kit start --resume-import TOKEN --root workspace --home private-state
```

This path needs no source ZIP. Review the frozen recovery actions, then choose
the offered exact action: resume incomplete publication, clean completed or
attributable staging, or cancel while retaining recovery state. Do not invent a
token, move the state between roots or remove unknown files to force recovery.
A reported supplied token alone is not proof that recovery exists or is approved.

Recovery completes or cleans this one opening transaction. It is not a general
command for resuming onboarding, editing answers or recording outcomes. Some
cleanup operations intentionally display no session.

Publication is recoverable, not a single atomic change across every file.
There is no rollback, hostile-owner guarantee, power-loss guarantee or universal
filesystem support. The driver supports the specified Darwin/Linux native
profile; actual observation limits remain documented in the
[fixture scope](../../contracts/invitation/starter/testdata/coverage.md).

## Legacy commands

`start --org-plan FILE` and `walk --plan FILE` retain their legacy plan-file
behavior. Do not combine legacy start options with invitation or recovery
options. Initial `start --help` describes the invitation route; explicit legacy
options still select the legacy implementation under
[0420](../decisions/0420-invitation-start-routing.md).

The exact interaction and storage formats are specified in the
[starter contract](../../contracts/invitation/starter/contract.md).
