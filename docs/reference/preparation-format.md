# Preparation format

Status: bounded shape approved under R1; implementation pending.
Date meaning follows [0180](../decisions/0180-first-onboarding-use.md).

## The three files

The organization plan remains shared. Each account has a preparation file owned
by onboarders and a record owned by the newcomer. Approved preparation convention:

```text
plan.obk.yaml
records/<account>/preparation.obk.yaml
records/<account>/record.obk.yaml
```

Keep the record's existing configurable results path. Derive the preparation
path from the same person's record directory, with a fixed sibling filename.
If a custom record filename is that sibling filename, reject the collision.
Never follow an arbitrary archive entry path. This is the path-policy direction;
the preparation validator does not implement filesystem or archive safety.

## Example

All names and identifiers below are fictional placeholders.

```yaml
schemaVersion: 1
kind: preparation
account: newcomer-one
planId: plan-one
recipient:
  name: Example newcomer
  companyEmail: ""
  plannedStartDate: "2026-11-02"
  team: team-one
  manager: lead-one
  onboardingOwner: owner-one
  buddy: buddy-one
  workLocation: remote
  timeZone: ""
arrangements:
  - id: first-pairing
    section: ticket-pairing
    kind: pairing
    owner: owner-one
    contacts:
      - buddy-one
    reference: Example first ticket
    targetDate: "2026-11-05"
    status: arranged
    note: ""
```

`planId` identifies the organization plan, not its current content hash. Updating
the shared plan does not change which person owns preparation. Contacts and
teams reference the plan directory; they are not duplicated personal directories.

## Recipient information

Account and plan identity are required. Recipient details may be incomplete
while an onboarder prepares things. A later invitation-readiness boundary checks
what must be supplied before delivery; it must not demand company email before
the activity that obtains company email.

The recipient fields cover name, company email, planned start date,
team, manager, onboarding owner, buddy, work location, and known time zone.
They are organization-supplied information, not claims about confirmed identity
or the newcomer's preferences. Work location is remote, on-site, hybrid, or
unfilled. Dates, when supplied, are real YYYY-MM-DD calendar dates.

Approved under R3 and the later clarification: do not ask for personal delivery
email, personal phone, or home address in the normal flow. Generate the archive
for the onboarder to deliver through an approved external channel. Company
email remains PII but may be visible in shared files. Prefer time zone and work
arrangements to location; only a needed rough region is appropriate. Exceptions
need explicit purpose, visibility, and retention decisions. Field-level encryption
is recorded for a later version, not present protection. See [0160](../decisions/0160-personal-data-and-preparation-approval.md).

## Arrangements

Each arrangement has its own permanent id and names the section instance it
supports. Several arrangements may support one section. The id does not change
when the list is reordered or the display wording changes. Duplicate ids are
errors, including identical declarations. Repeated contacts are ordinary data,
not duplicate identities.

- `kind` is one of support, equipment, office, remote, access, meeting, reading,
  pairing, ticket, pr-review, pr-post, pr-merge, deploy-shadow, or deploy-lead.
- `owner` is the accountable onboarder contact, not necessarily the helper.
- `contacts` names assigned helpers. For deploy-shadow these are the experienced
  people leading; for deploy-lead these are the people observing or supporting.
- `reference` identifies the ticket, PR, resource, meeting, or opportunity. It is
  text, not a command to execute or a URL the validator fetches.
- `targetDate` is an optional planning target, never a completion deadline.
- `status` is explicitly pending, requested, arranged, or deferred.
- `note` gives work context without secrets or private personal information.

Pending, requested, and deferred arrangements may lack an owner, helper, or
reference. Missing information remains visible in the preparation view. For an
arranged item, owner is required. All kinds except support require a reference.
Pairing, ticket, PR work, and deploy kinds also require at least one helper.
That is the approved minimum for claiming an arrangement is ready, not proof
the newcomer performed the activity or the requested access works.

Unknown references are invalid even for pending items. An empty helper list is
incomplete; a list containing an unknown helper is not valid partial preparation.
Inactive but declared sections remain valid references so disabling an activity
does not discard its arrangements.

## Ownership and updates

Editing preparation changes no shared plan or newcomer record. Onboarder task
status is distinct from newcomer outcomes: arranged is not verified. Reopening
an older invitation must not silently replace newer preparation or personal work.
Same-owner concurrent Git edits are resolved by those editors.

Do not use preparation as an arbitrary map over plan or record fields. The
typed recipient and arrangement fields are the approved bounded first increment.
Declaring template bindings to these fields is a later, explicit format change.

The preparation date is plannedStartDate. Under 0180 there is no
actual-start-date question. The newcomer's record will instead carry firstOpenedAt,
set once at their first confirmed invitation start. Existing record.startDate
values remain unchanged until their compatibility rules are explicitly updated;
do not treat them as proof of first use or employment start. Saved personal
answers retain their existing precedence.

## Serialization and later boundaries

Use the existing YAML dialect and guards. Require kind preparation; reject
duplicate mapping keys before converting them to maps. Preserve unknown fields
at the top level, as existing plan and record formats do, and reject unknown
keys in fixed nested structures. A preparation decoder and encoder need their
own fixtures for those rules, including raw repeated YAML keys and canonical
round trips. They are not tested by the typed
[validator suite](../../contracts/preparation/validator/contract.md).

The approved preparation model and validator contract are staged separately
from implementation. Template bindings and serializer contracts remain separate
increments requiring approval. Before invitation implementation, approve
the bounded archive layout, import conflict handling, path validation, identity
confirmation, and readiness rules. No new runtime behavior is authorized by
this format reference alone.
