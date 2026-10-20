# 0080. Directory

Status: proposed

Builds on [0055](0055-variable-audience.md), [0065](0065-messages-and-evidence.md), and [0075](0075-refreshing-a-personal-plan.md). It answers [#46](https://github.com/8thlight-daveloper/onboard-kit/issues/46).

## Context

People appear as free text today: a `person` variable holds lines such as `Name, email, @handle`. The same person gets typed in several places, nothing checks an email or a handle, and a message step (0065) needs a real handle. Roles are mixed up with people: "Engineering manager" is a job, and who holds it differs by team.

Nothing has been released, so nothing needs migrating.

## Decision

### A directory in the org plan

The org plan gains a `directory` with three lists, each entry with a permanent id.

```yaml
directory:
  contacts:
    - id: sam-okafor
      name: Sam Okafor
      email: sam.okafor@acme.example
      slack: "@sam"
      slackId: U01ABC123
    - id: it-help
      name: IT help desk
      email: help@acme.example
      slack: "#it-help"
  roles:
    - id: engineering-manager
      name: Engineering manager
      holders: [sam-okafor]
    - id: it-contact
      name: IT contact
      holders: [it-help]
  teams:
    - id: platform
      name: Platform team
      holders:
        engineering-manager: [sam-okafor]
```

- A **contact** is a person or a shared contact such as a help channel. `email`, `slack`, and `slackId` are optional. `slack` is for display. `slackId` is for links, because a handle can change and Slack's links need an id.
- A **role** is a named job. Its `holders` are the organization-wide holders, in order.
- A **team** may override who holds a role for that team, also in order.

### Questions pick from the directory

Three variable types replace `person`: `contact`, `role`, and `team`. An answer stores an id and nothing else. The walkthrough offers the directory's entries to pick from, and the onboarder adds a new contact inline. The `team` question's options are the directory's teams, which settles where the team list comes from.

A `multiple: true` contact or role question holds an ordered list. `owner` and `escalation` name `contact` or `role` variables, and the order of an escalation list is meaningful.

### Resolve when read

An answer, a personal plan, and a refresh store ids and never copy a name, an email, or a handle. The name is looked up in the latest directory each time a document is assembled or a message is shown. A changed contact reaches every newcomer on the next refresh (0075), and a stale name can never be stored.

A role resolves in this order: the newcomer's team, if the team override names holders for it, otherwise the role's own holders. The result is an ordered list of contacts.

- **No holder** is a gap, shown like any other gap and never as an empty name or a guess.
- **A team not yet answered** leaves a role resolving to the organization-wide holders if it has any, and is marked as not yet specific to a team.
- **Escalation** lists every step in order. A vacant role is shown as vacant and not skipped.

### Rendering

A resolved role renders as `Engineering manager: Sam Okafor, @sam`. A contact renders as `Sam Okafor, @sam`, with the email when the context asks for it. The assembler and the walkthrough share the rule. The exact forms are not decided.

### Outward actions show the resolution

A message to a role (0065) shows who it resolved to and why: "Engineering manager of Platform team: Sam Okafor, @sam". A wrong team answer is then visible before anything is sent.

### Integrity

The plan validator reports:

- Errors: a duplicate id, an answer that names an id that is not in the directory, a holder that is not a contact, and a team override for a role that does not exist.
- Warnings: a role with no holder, an email with no `@` or with a space, and a contact with no way to reach them.

An email check stays minimal, since valid addresses are varied.

### Scope

The directory holds people, roles, and teams. It is not a place for general per-team settings. A repository per team stays the open question in 0055.

### Privacy

The org plan then holds contact details for the people named. The org plan's location must require authentication, as 0075 says. Every contact field except the id and name is optional.

## Why

- One place to correct a person, and no copy to go stale.
- A role keeps working when the person holding it changes.
- A message or an escalation reaches someone who exists.
- Ids make renames safe, and a vacancy becomes visible.

## Consequences

- **The domain model changes.** A Directory with contacts, roles, and teams. The variable type `person` is removed, and `contact`, `role`, and `team` are added. This needs `slow-change`.
- **Contracts**, each its own small increment:
  - Decoder and encoder: read and write the directory and the new types. The loader reads the new types in templates and carries no directory.
  - Hasher: the new types are part of a variable.
  - Resolver: it gets the directory and the team answer, and fills a reference as its rendered text or as a gap.
  - Assembler: it passes the directory, and renders references the same way.
  - Finder and validator: the new types are choices from the directory, and the integrity checks above.
  - Editor: operations to add, change, and remove a contact, a role, and a team.
  - Instantiator: a new plan starts with an empty directory.
- **ACME** gains a directory. `general_contacts`, the grantors, and the owner role become references.
- **The walkthrough** gets a People screen for the onboarder. The onboardee sees only results.

## Not decided

- The exact rendering forms and whether a template can ask for a single field, such as a contact's email.
- Whether a contact can be removed while an answer still names it, or removal must be refused.
- Importing contacts from a directory service.
- Whether a team can be part of another team.
