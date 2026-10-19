# Escalation

Part of the [section template schema](README.md).

Escalation says who to ask when something is blocked. It is built from `person` variables, which are ordinary [variables](variables.md#person), and one field, `escalation`, on a section or a step.

## Roles are variables

A template names a role by declaring a `person` variable with `multiple: true`, such as `source_control_grantor`. The consultant answers it once while planning, with the people who can fill the role. Two people can fill one role, which is the fix for a grantor who is out of office. Templates never assume a particular person exists.

## The `escalation` field

`escalation` is an ordered list of `$name` references to `person` variables that the same section declares. It goes on a section and on a step. A step inherits its section's list when it has none.

```yaml
variables:
  - name: source_control_grantor
    type: person
    multiple: true
    description: Who grants source control access
    required: true
  - name: access_backup
    type: person
    multiple: true
    description: Who to ask when the grantor is away
escalation: [$access_backup]
steps:
  - id: source-control
    name: Source control
    owner: $source_control_grantor
    escalation: [$access_backup]
```

A section uses only the variables it declares, so a step cannot fall back on another section's list. If two sections declare a variable with the same name, the finder asks it once.

## The general path is its own section

"If you are stuck" is a standard core template, always included. It declares `general_contacts`, a required `person` variable with `multiple: true` for who can help with anything. Where a consultancy sends people to a client, its template set adds a second template that declares a `consultancy_contacts` variable for who at the consultancy can help. A set for full-time employees omits it. No tool code knows about either side.

Each blocked step points to its own list. The document points to the "If you are stuck" section for everything else.

## A worked case

A newcomer needs access to the organization's source control.

1. The step's own list names the grantor's backup. The grantor is away.
2. The section's list names the onboarding owner.
3. The "If you are stuck" section names someone who can help with anything, and the consultancy's contacts, who point to someone else to try.

The newcomer is unblocked without anyone having been left as the only name on the page.

A multi-step approval is steps with different owners. One step is the newcomer submitting a request. The next is the manager approving it, with an `ask` check, and a third is the grantor granting it. Each step has its own `escalation`.

## What the validators check

| Problem | Severity | Example |
| -- | -- | -- |
| Unknown reference | Error | `escalation` names a variable the section does not declare, or one that is not a `person` variable |
| Single point of failure | Warning | A step's list resolves to fewer than two distinct people |
| One person | Warning | A `multiple` person variable has exactly one person |
| Thin contact | Warning | A person has fewer than two ways to be reached |
| None given | Error | A required `person` variable has no person and is not `deferred` |
