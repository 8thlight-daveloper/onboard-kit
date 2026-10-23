# Preparing one person's onboarding

Status: implemented, experimental. Invitation packaging and the guided invitation
flow are not implemented yet.

Preparation is the onboarder's file for one person. It holds supplied details
and arrangements, not the newcomer's answers or proof that they completed work.
See the [three-file format](preparation-format.md).

## Start and inspect

Use a plan with an id and an explicit account. Run from the folder that will
hold the person-specific files, or supply an existing directory with --root.

```sh
onboard-kit prepare init --plan plan.obk.yaml --account newcomer-one
onboard-kit prepare show --plan plan.obk.yaml --account newcomer-one
```

Init refuses an existing preparation. It creates no record and fills in no
recipient details automatically. The usual destination is
records/newcomer-one/preparation.obk.yaml. A custom resultsPath must put the
account placeholder in a directory; the preparation remains the record's sibling.
See [path rules](../../contracts/preparation/locator/contract.md).

Show prints JSON. Recipient information and arrangements are separate from the
record's outcomes. A missing record is normal. A malformed or mismatched record
refuses the combined view, but does not block a preparation edit. Show writes no
files and does not display private answers or preserved unknown data.

## Supply recipient details

```sh
onboard-kit prepare recipient --plan plan.obk.yaml --account newcomer-one --name "Example newcomer" --planned-start-date 2026-11-02
```

Available flags are --name, --company-email, --planned-start-date, --team,
--manager, --onboarding-owner, --buddy, --work-location, and --time-zone.
Team and contact fields use exact IDs declared in the plan. Work location is
remote, on-site, hybrid, or empty. The date is a planning target, not an actual
start date or a deadline. This command records no first-use event.

Supply at least one field. Omitted fields keep their values; an explicitly empty
value clears a field, for example --buddy=. No default becomes an answer, and
these details are not silently copied into newcomer answers.

## Arrange and repair support

The following IDs are examples; use the section and contact IDs in your plan.

```sh
onboard-kit prepare arrangement --plan plan.obk.yaml --account newcomer-one --id first-pairing --section ticket-pairing --kind pairing --status pending
onboard-kit prepare arrangement --plan plan.obk.yaml --account newcomer-one --id first-pairing --owner owner-one --contact buddy-one --reference "Example ticket" --status arranged
onboard-kit prepare remove --plan plan.obk.yaml --account newcomer-one --id first-pairing
```

Arrangement flags are --id, --section, --kind, --owner, --contact, --reference,
--target-date, --status, and --note. An existing ID updates that item in place;
a new ID appends. Supply at least one field besides the ID. Kind and status are
never inferred. The [format](preparation-format.md#arrangements) lists their values
and what is required before an item can be called arranged.

Repeat --contact to replace the entire helper list in the supplied order. Use
--clear-contacts, or --clear-contacts=true, to clear it; do not combine that with
--contact. Other empty flags clear their fields. Remove requires --id; an explicit
--id= can remove one malformed empty-ID item. Duplicate IDs always block edits.

An edit checks the changed recipient or arrangement without blocking on another
item's stale reference. Show lists the remaining problems. A successful edit
does not certify invitation readiness, access, or newcomer completion.

## Ownership and safety

Each edit changes only this person's preparation. It does not read their record,
rewrite the plan, run Git, send an invitation, or provision access. Two onboarders
editing the same preparation still coordinate through their normal Git workflow.

The selected root must already exist and must not be a symlink. Artifact paths
cannot follow symlinks. Input and saved preparation files have a 1 MiB budget.
Writes publish atomically; init never overwrites an existing preparation. New
files use private permissions, but sharing them in Git makes them visible to
repository readers. Business email and linked progress are still personal data.

Successful edits print the relative preparation path. Errors explain which
operation needs repair without repeating personal values. If saving succeeds
but output fails, the saved edit remains; the command reports that distinction.

Flags can use a space or equals. A value beginning -- must use equals, such as
--note=--literal. Only --contact is repeatable; repeated scalar flags are errors.
Use `onboard-kit prepare --help` for the operation list. Exit codes are 0 for
success, 1 for data or filesystem problems, and 2 for usage errors.
