# 0080. Directory

Status: accepted

The resolver's directory-default fallback under Integrity is superseded by
[0120](0120-defaults-require-acceptance.md). Defaults require explicit acceptance
for directory variables too.

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
      contact: platform-team
      holders:
        engineering-manager: [sam-okafor]
```

- A **contact** is a person or a shared contact such as a help channel. `email`, `slack`, and `slackId` are optional. `slack` is for display. `slackId` is for links, because a handle can change and Slack's links need an id.
- A **role** is a named job. Its `holders` are the organization-wide holders, in order.
- A **team** may name a `contact`, the id of a directory contact such as the team's channel, email, or on-call alias. It may also override who holds a role for that team, also in order. A team's contact is for reaching the team. It never stands in for a role holder, so a vacancy stays visible.
- **Ids** are lowercase letters, digits, and hyphens, starting with a letter or a digit. An id is unique within its list. A contact, a role, and a team may share an id, because each list is separate.
- Teams are a flat list, with no nesting, see Scope.

### Questions pick from the directory

Three variable types replace `person`: `contact`, `role`, and `team`. An answer stores an id and nothing else. The walkthrough offers the directory's entries to pick from, and the onboarder adds a new contact inline. The `team` question's options are the directory's teams, which settles where the team list comes from.

A `multiple: true` contact or role question holds an ordered list. `owner` and `escalation` name `contact` or `role` variables, and the order of an escalation list is meaningful.

- **Shapes.** A contact or role answer is one id, or a list of ids when the variable is `multiple`. A team answer is one id and never a list, and `multiple` is allowed only on contact and role. `deferred` is accepted for every type.
- **Choices.** The finder gives each of these questions the directory's entries as choices, an id and a name, in directory order. It leaves `options` empty. An empty directory gives no choices, and the interface says there is nothing to pick yet.
- **The plan's team** is the answer to the plan variable of type `team`. A plan declares at most one such variable. A plan with none, or with no answer or an unknown id, has no team.

### Resolve when read

An answer, a personal plan, and a refresh store ids and never copy a name, an email, or a handle. The name is looked up in the latest directory each time a document is assembled or a message is shown. A changed contact reaches every newcomer on the next refresh (0075), and a stale name can never be stored.

A role resolves in this order: the newcomer's team, if the team override names holders for it, otherwise the role's own holders. The result is an ordered list of contacts.

- **A team's entry replaces and never merges.** It does not add to the role's holders. An empty entry means the role is vacant for that team. It is kept as an empty list in the file and the model, because dropping it would hand the role back to its own holders.
- **A holder that is not a contact** is skipped when resolving. The validator reports it as an error.
- **No holder** is a gap, shown like any other gap and never as an empty name or a guess.
- **A team not yet answered** leaves a role resolving to the organization-wide holders if it has any, and is marked as not yet specific to a team.
- **Escalation** lists every step in order. A vacant role is shown as vacant and not skipped.

### Rendering

Decided, and shared by the resolver, the assembler, and the interface.

- A contact renders as `Name, slack` when it has a `slack`, otherwise `Name, email` when it has an email, otherwise `Name`.
- A list of contacts, and a `multiple` answer, join with a semicolon and a space, in order.
- A role renders as the contacts it resolves to. A vacant role renders as `Role name (vacant)`.
- A team renders as its name.
- The document's title shows the team's name when the plan has a team.
- Escalation is written in the document as `Escalation: a, then b.`, in a section's footer and on a step, in list order. A vacant role stays in the list as vacant. An unanswered item shows as unanswered.
- A vacant role that the document uses is a gap in the gap table, once per question, with code `directory.role-vacant`.

### Outward actions show the resolution

A message to a role (0065) shows who it resolved to and why: "Engineering manager of Platform team: Sam Okafor, @sam". A wrong team answer is then visible before anything is sent.

### Integrity

The plan validator reports these, each with a code.

| Code | Severity | When |
| -- | -- | -- |
| `directory.duplicate-id` | Error | Two entries in one list share an id. Reported once per id |
| `directory.holder-not-a-contact` | Error | A role's holder, a team override's holder, or a team's `contact` names an id that is not a contact |
| `directory.team-role-unknown` | Error | A team override names a role that does not exist |
| `directory.unknown-id` | Warning for a stored answer, error for a submission | An answer names an id that is not in the list for its type |
| `directory.role-vacant` | Warning | A role's own holders are empty, whatever the teams override |
| `directory.email-malformed` | Warning | An email that is not empty and has no `@` or has a space |
| `directory.contact-unreachable` | Warning | A contact with no email, `slack`, or `slackId` |
| `plan.duplicate-team-variable` | Error | Two plan variables of type `team` |
| `plan.duplicate-section-id` | Error | Two sections share an id, which 0075 needs |

A role is vacant in the document when the holders it resolves to for the plan's team are empty. That is a different test from the validator's, which looks only at the role's own holders. The validator warns about a role that has no organization-wide holder even when every team fills it.

An answer that names an id that is not in the directory counts as unanswered everywhere, so a required question blocks until it is filled.

- **The finder** treats it as unanswered. For a `multiple` answer one known id is enough to count as answered. An optional question never blocks. A default or given suggestion is offered only when every id in it is in the directory for the question's type.
- **The resolver** treats it as unanswered too, so it takes a default if the default names something in the directory, and otherwise shows the unanswered gap. In a list it skips unknown ids.
- **The validator** reports a stored one as a warning, and a new submission naming one is refused as an error. The tool never creates such an answer, see removal below.

An email check stays minimal, since valid addresses are varied.

### Editing

The plan editor adds, changes, and removes directory entries. Each edit is one function that returns a new plan.

- **Set** adds an entry at the end, or replaces the entry with the same id whole and in place. It clears nothing. It is refused, with the plan unchanged, when the entry has a malformed id or an empty name, a holder that is not a contact, a team `contact` that is not a contact, or a team override for a role that does not exist. An id never changes once made.
- **Remove** of an id that is not in that list is an error and changes nothing.
- **The interface shows a removal before it applies it.** The editor changes no argument, so a driver runs the removal on the current plan as a dry run, shows what it would clear, and applies it only after a yes.

### Removing an entry clears its references

Removing a contact, a role, or a team is allowed. The same edit clears every answer and every holder entry that named it, and reports what it cleared. A cleared required question is unanswered, so it blocks onboarding until the onboarder fills it. The tool never leaves a dangling reference behind.

- **Answers are cleared by the variable's type.** A `string` answer that happens to hold the same text stays. Answers in every section are cleared, including sections that are off, hidden, or sub-sections.
- **A list answer loses the id.** If nothing is left in it, the answer is removed.
- **Holders.** Removing a contact removes it from every role's holders and every team override, and clears a team's `contact` that named it. Removing a role also removes the team overrides for it. A team override that loses its last holder stays as an empty list, so the role is vacant for that team. Removing a team removes its entry and clears answers that named it.
- **The report** lists the cleared question ids, as `<scope>.<name>`, and the places changed, as `role:<role id>`, `team:<team id>/<role id>`, and `team-contact:<team id>`, each once, sorted. The places that go with the removed entry itself are not listed.

### Scope

The directory holds people, roles, and teams. It is not a place for general per-team settings. A repository per team stays the open question in 0055.

A section meant for several teams uses `includeIf` with `isOneOf` on the team answer.

### In the file

The directory is written after the plan's origin and before its variables, with `contacts`, `roles`, and `teams` in that order, entries in the order given, and a team's override roles sorted. An unknown key anywhere inside the directory is an error, with no exemption for keys that start with `x-`. A plan's own unknown field named `directory`, `id`, or `origin` is refused on writing, because those are known fields now.

### Over stdio

The stdio driver, used by the acceptance tests, has `directory`, `set-contact`, `set-role`, `set-team`, and `remove` with `kind` and `id`. A removal also returns what it cleared. A question's JSON carries its `choices`.

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
- **The walkthrough** gets a Contacts screen for the onboarder. The onboardee sees only results.

## Not decided

- The section selector still reads stored answers when it tests an `includeIf` condition. A section gated on a directory answer that names an unknown id is judged as the answer is stored, while the finder treats the answer as unanswered. This needs a decision.
- A team's `contact` is stored, checked, and edited. Whether the document shows it, and how a template refers to it, such as "message the team", waits for messages (0065).
- Whether a template can ask for a single field, such as a contact's email.
- Importing contacts from a directory service.
