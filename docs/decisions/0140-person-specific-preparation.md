# 0140. Person-specific preparation

Status: accepted; implementation pending

[0160](0160-personal-data-and-preparation-approval.md) records approval of the
bounded preparation shape and the personal-data principle. Start-date semantics
are reviewed separately from that approval.

Extends [0090](0090-plan-and-onboarding-record.md) with a third artifact.
Supersedes its rule that all onboarder writes go to the shared plan: preparation
for one person goes to that person's preparation file instead. Shared plan
content and the newcomer's record keep their existing owners.

## Context

An onboarder needs to assign a pairing partner, arrange a deployment to shadow,
name a reviewer, and prepare access for one newcomer. These are not shared
organization facts or the newcomer's answers. They can also change after an
invitation has been sent.

Clear field ownership would let both people edit one file without intentionally
changing the same values. Git merges text, however, so nearby edits and changes
to formatting can still conflict. The invitation already carries several files;
we do not need one shared writable file to deliver one person's onboarding.

## Decision

Use three logical artifacts, stored as separate YAML files:

| Artifact | Owner | Holds |
| -- | -- | -- |
| Organization plan | Onboarder | Shared sections, instructions, directory, and organization or team facts |
| Person-specific preparation | Onboarder | Assignments and arrangements for one newcomer, including pairing and deployment support, reviewers, and access preparation |
| Onboarding record | Onboardee | Their answers, confirmed identity, outcomes, and notes |

The organization plan is shared across newcomers. Only preparation and the
record are person-specific. The tool reads the three together, but saves an
edit only to the artifact its owner controls. An assignment or an access request
is not evidence that the newcomer completed an activity or obtained access.

The invitation carries preparation alongside the plan, the seeded record, and
the first-steps document. Later preparation updates do not require a first
check-in from the newcomer and must not overwrite their answers or outcomes.
Updates travel through the chosen file-sharing workflow; this introduces no
background synchronization or automatic Git operations.

Separate files remove the routine same-file merge problem between the
onboarder and onboardee. They do not promise conflict-free Git use. Concurrent
edits by two onboarders to a preparation file remain ordinary Git conflicts
for those onboarders to coordinate and resolve. That tradeoff is accepted.

Ownership describes supported tool behavior, not access control. Repository
permissions still govern who can read or edit files. Splitting files does not
make personal information private.

## Alternatives considered

### One writable file with owner-specific fields

This preserves logical ownership but leaves both roles editing the same text
file. Separate files avoid that routine source of Git conflicts without a
merge service or a new delivery mechanism.

### One SQLite database per person

Keep this as an option, not the storage choice for this release. The
[Turso per-user database article](https://turso.tech/blog/give-each-of-your-users-their-own-sqlite-database-b74445f4)
describes isolating each user's data in its own database rather than sharing
all users' rows in one database. A per-person onboarding boundary fits that
approach, but it does not decide how two people exchange updates.

An embedded SQLite database can work locally and offline without a hosted
service. SQLite documents both
[application-file and server-mediated uses](https://sqlite.org/whentouse.html).
Copying independently edited database files through Git would not provide
meaningful row-level merging. Local transactions do not reconcile divergent
copies on different machines.

A hosted per-person database could coordinate writes through an authoritative
service. It would also require authentication, authorization, hosting, backup,
and an explicit offline synchronization and conflict policy. It would not
remove the need for clear data ownership.

Revisit this option when immediate shared updates, private submissions, or
reporting needs justify those costs. Preserve the ownership boundaries if
storage changes. Any implementation must still satisfy the standalone and
offline requirements of [0005](0005-self-contained-install.md); a database
service must not become necessary to start or continue onboarding offline.

## Delivery and verification

This records the storage and ownership direction, not a completed feature.
Define the preparation schema, identity and path rules, update behavior, and
invitation entry validation before implementation. Any personalized pre-seeds
need explicit precedence rules; do not invent a first-wins or last-wins merge.
Existing archive safety limits and identity confirmation remain requirements.

Contract and acceptance increments must demonstrate that:

- Two newcomers can have different assignments without changing their shared
  organization plan.
- Updating one person's preparation changes neither another person's files nor
  the first person's answers, notes, or outcomes.
- Saving newcomer answers or outcomes leaves preparation and the plan alone.
- An invitation supplies enough local information to start without organization
  access, and a missing assignment remains visible rather than guessed.

No contracts, domain-model fields, archive parser, or implementation change as
part of this decision. Their changes need the separate approval and independent
authorship required by [0025](0025-contracts-and-tests.md).
