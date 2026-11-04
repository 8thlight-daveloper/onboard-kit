---
category: regenerative-process
status: proposed
builds-on: ['0020', '0025']
---

# 0620. Keep decisions as records and current state in checked files

Builds on [0020](0020-binary-structure-and-regenerative-components.md) and
[0025](0025-contracts-and-tests.md). When accepted, it replaces the dependency
table in 0020.

## Context

A regenerative build starts from contracts and the decisions behind them. A
person or an agent has to be able to ask which decisions affect one component,
and trust the answer. Three things in this repository get in the way:

- **A table in a record went stale.** 0020 holds a dependency table of 33
  components. The code has 59. 29 are missing from the table, 3 in it no longer
  exist, and 3 imports in the code break it. Nothing fails when they drift,
  because a table in prose cannot be checked.
- **Status is free text.** The records use about eight spellings of "accepted",
  and several add the state of the build, such as "implementation pending". That
  state goes stale, and a program cannot read any of it.
- **Decisions are hard to find from a component or a category.** 32 of 47
  contracts cite a record, in several styles. Method decisions, product decisions,
  and this project's operations share one list.

## Decision

**A record and a checked file have different jobs.** A decision record is a dated
statement of what was chosen and why. It changes only by its status, and a later
record supersedes it. Current state lives in files that a check reads:
`architecture.yaml`, the contracts, the reference docs, and the stack files. A
record never holds a table or a list that a check could verify. It states the
decision and links to the file that carries the result.

**Every record starts with front matter.** The fields are `category`, `status`,
`approval`, `note`, `affects`, `builds-on`, and the replacement fields. The
exact schema is in the
[decision record front matter](../reference/decision-record-front-matter.md).

**Records have five categories,** and all stay in this repository:

| Category | What it decides |
| -- | -- |
| `architecture` | The shape of the product and its code |
| `behavior` | What the tool does or refuses, whatever it looks like |
| `ux` | How a person meets it: screens, flows, messages, and the command surface |
| `regenerative-process` | The method of building from contracts |
| `operations` | How this project runs: CI, releases, tooling, and one-time exceptions |

Numbers stay one sequence, by tens, so no link breaks.

The `regenerative-process` category marks the decisions about the method, so the
Threnos repository can list them at any time and link to them.

**`affects` links a decision to components.** A `behavior` or `ux` record names
the components it changes, as ids in `architecture.yaml`. It is `[]` only while
the component it describes is not built. An `architecture` record may name them.
A process record names none. No record says `all`. A list is as long as the
decision needs, so a record that 18 contracts cite names 18 components.

**A record holds no implementation state.** "Implemented" and "pending" are
facts about the code, and a check of the code says them better. A record also
holds no counts or scores.

**`just check` reads the records.** It fails on a missing or unknown field, on a
component id that does not exist, on a replacement that is not listed on both
records, on `affects` in a process record, and on a contract that cites a product
record missing from that record's `affects`. It reports without failing a listed
component whose contract does not cite the record, because contracts keep no
history.

## Consequences

- The dependency table in 0020 moves to `architecture.yaml` once a person accepts
  that file. 0020 then gets `partly-superseded-by` 0620, and its table becomes a
  link. 0020's table is older than `docs/architecture/components.md`, which
  already declares the positions and imports of the newer components. The
  declared imports in `architecture.yaml` come from that document, never from
  what the code does, so a difference in the code shows as a violation.
- The 71 existing records now have front matter. Their categories and `affects`
  lists are a first pass that a person reviews. The border cases are 0010, 0360,
  0400, 0420, and 0580.
- Twenty status lines named the state of the build, and that text is gone. The
  code is the place to read it.
- "Where decisions go" in `AGENTS.md` needs a line that points at the schema. A
  person approves that edit.
- A new record is harder to write by a little. It needs a category and an
  `affects` list. A check says what is missing.

## Not decided

1. **When `architecture.yaml` is canonical.** It is a draft until a person accepts
   it. Its 12 command components have no declared imports and are held to their
   kind only.
