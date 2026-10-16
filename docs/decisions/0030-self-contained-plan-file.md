# 0030. Self-contained plan file

Status: accepted

Relates to [#24](https://github.com/8thlight-daveloper/onboard-kit/issues/24), [#23](https://github.com/8thlight-daveloper/onboard-kit/issues/23), [#6](https://github.com/8thlight-daveloper/onboard-kit/issues/6), [#3](https://github.com/8thlight-daveloper/onboard-kit/issues/3). Builds on [0010](0010-plan-editor.md) and [0020](0020-binary-structure-and-regenerative-components.md).

## Context

A plan ends up at the organization it was made for, in whatever tool version and module set exist later. It may hold content that no module has, because each engagement can add its own sections and questions. If reopening a plan needed the modules it was made from, we would have to decide what happens when those modules are missing or have changed, and a plan could break through no fault of its own.

## Decision

The saved file is the plan. There is no separate configuration file. A plan is named `<name>.obk.yaml`. It holds every standard module as a section, and everything needed to edit and assemble it:

- `schemaVersion`
- `provenance`: the module set version and the generated date
- `questions`: plan-level questions, such as the axis questions team state and discipline
- `values`: the answers, kept by scope. `plan` holds the plan-level answers, and each section has its own scope
- `sections`: all the content, with placeholders left in place

Sketch. Field names are not final, see [#3](https://github.com/8thlight-daveloper/onboard-kit/issues/3):

```yaml
schemaVersion: 1
provenance:
  moduleSet: 7652e0e
  generatedOn: 2026-10-15
questions:
  - id: teamState
    description: Is the team established, or are we the first consultants?
    options: [established, first-consultants]
    required: true
  - id: discipline
    description: Which discipline is this plan for?
    options: [software, design]
    required: true
values:
  plan:
    teamState: established
    discipline: software
  sections:
    clone-and-run-tests:
      repo_url: git@git.acme.example:platform/widget-service.git
sections:
  - id: clone-and-run-tests
    title: Clone the project and run the tests
    source:
      module: clone-and-run-tests
      hash: 9f2c...
    appliesTo:
      teamState: [any]
      discipline: [software]
    order: 20
    placeholders:
      - name: repo_url
        description: Clone URL of the main repository
        pattern: '^(git@|https://)\S+$'
        patternHint: A clone URL, starting with git@ or https://
        required: true
    steps:
      - id: clone
        name: Clone the repository
        run: [git, clone, $repo_url, $repo_dir]
```

### Sections are modules living in a plan

A section has the same shape as a module, with an added `source`. A module file is a section before it is in a plan. A plan holds all of them, with their selection fields (`appliesTo`, `order`) kept and their placeholders unresolved.

- **Active sections are computed.** A section is active when its `appliesTo` matches the axis values and its `enabled` answer, or default, is true. A section with no `enabled` placeholder is always on. See [module schema](../reference/module-schema.md#optional-sections). The section selector computes this from the plan whenever it is needed. Nothing is chosen once and thrown away.
- **Questions are computed too.** The questions asked are the plan-level questions plus every placeholder of every active section whose `askIf` condition holds, if it has one, including the section's own `enabled` toggle. Each is its own question, with its own description, type, and default. A section added by hand brings its questions with it.
- **Placeholders fill at assembly.** Values fill `$name` when the document is assembled. Nothing resolved is stored. A person can change a value or a section's text and get a correct document without the original modules.
- **Inactive sections have no effect.** A hand edit to an inactive section does nothing until team state or discipline makes it active.

### Values are scoped

A question belongs to a scope: the plan, or one section. Its id is `<scope>.<name>`, for example `plan.teamState` or `access.repo_url`. Each scope holds its own values, all in the plan's `values`.

- **No coordination between modules.** Two sections that both declare `repo_url` have two questions and two values. A module author never needs to know another module's names.
- **Asking once is a convenience.** The finder groups questions with the same name, type, and options. A user interface can ask a group once and apply the answer to every member, and a person can split it. The finder also offers `suggestions`: the distinct declared defaults and the values already given for that name in any section, including inactive ones. See [0010](0010-plan-editor.md).
- **Filling reads one scope.** A placeholder takes its own section's value, then that section's default.
- **Newcomer values are outside the plan.** A placeholder with `phase: onboarding` is answered by each newcomer, so the plan holds no value for it. Planning skips it. Where those values live belongs to the onboarding runtime, [#28](https://github.com/8thlight-daveloper/onboard-kit/issues/28).

### The default plan

A new plan starts as the default plan: the axis questions plus every standard module as a section, with no values. The plan instantiator builds it from the loaded modules and the small built-in axis question file. The axis questions are data like every other question.

Modules stay the authored source, one file each. The default plan is derived, never committed.

### Modules are used once

| Needs modules | Does not need modules |
| --- | --- |
| Starting a new plan | Editing values |
| Updating from a newer module set, deferred until after the MVP | Changing team state or discipline |
| | Changing the output style |
| | Editing section text by hand, or adding a section |
| | Assembling the document |

Saved plans keep every section by default, so changing an axis later needs no modules. A later, explicit `prune` action could remove inactive sections for a consultant who wants a tidier file to hand over. It is not part of the MVP.

## Format

YAML, with these guards. The file format outlives any implementation, so it is an ADR decision, not a stack note.

- Booleans are `true` and `false`, lowercase. An unquoted `yes`, `no`, `on`, `off`, `y`, `n`, `True`, or `TRUE` is an error. Quote it if it is text.
- Null is `null` or an omitted key.
- Anchors and aliases are rejected.
- Numbers that parse surprisingly, such as `1.0` or a leading zero, are errors where a string is meant.
- Modules decode strictly and reject unknown fields. Plans decode leniently and keep unknown fields.
- One JSON Schema covers modules and plans, so editors can validate and complete them.
- `onboard-kit lint` enforces these rules on modules. Loading a plan enforces the same rules, so a hand-edited `required: yes` is a problem, not a silent `true`.

Why YAML: modules are authored and reviewed by people, and literal blocks keep prose readable. YAML allows comments. A section copies from a module file into a plan unchanged. Plans in version control produce legible diffs. Few people read a plan end to end, since consultants use the tool and clients read the assembled document, so the choice is driven by module authoring, and the plan inherits it.

## Versioning

`schemaVersion` is an integer, on plans and on module files.

- **Additive, optional fields:** no bump. Plans decode leniently, and an older tool preserves fields it does not know.
- **Breaking change:** bump the version and ship a migration from the old version to the new one. Migrations chain.
- **Plan migrator.** A component with its own contract: an old-version plan in, the current version out. It holds frozen copies of old types, because the current model no longer describes them. An old-version plan and its migrated result are permanent fixtures.
- **Older than supported:** refuse, with a message saying which versions are supported.
- **Newer than the tool knows:** refuse to open it, with a message to upgrade. Guessing risks misreading fields.
- Loading a plan runs the migrator first, then validation.

## Identifiers

| What | Example | Source |
| --- | --- | --- |
| Module id | `access` | Written by the author. Stable, never reused. |
| Source hash | `9f2c...` | Computed by the tool when it copies a module into a plan. |
| Module set version | `7652e0e` | The build's commit sha or tag, embedded at build time. Recorded in provenance. |

The source hash is a hash of the section's structured content in a fixed order, not of raw text, so reformatting does not change it. It excludes `source`. Placeholders are unresolved, so values do not change it either. The canonical form is a contract, protected by golden fixtures. Authors never write a hash into a module file, so it cannot go stale.

A git blob sha would not do. It hashes raw text, it does not match the section's shape, a custom module folder may not be in git, and shas change on rebase or squash.

## Updates

The update action is deferred until after the MVP. Its design is reserved now, because `source` has to be written from the first plan. Without it, old plans have no baseline.

An update compares three things for each section that has a `source`:

| Section now vs recorded hash | New module vs recorded hash | Meaning |
| --- | --- | --- |
| Same | Same | Nothing to do |
| Same | Different | Untouched by hand, safe to replace |
| Different | Same | Edited by hand, no upstream change, keep it |
| Different | Different | Conflict, the consultant chooses |

The hash says whether something changed, and the module set version says where. A section whose hash differs from the recorded one was edited by hand. A newer module whose hash differs from the recorded one changed upstream, between the recorded build and the current one.

- Sections with no `source`, such as engagement-specific ones, are never touched by an update.
- A module removed upstream leaves its section in place, flagged.
- Comparing hashes also makes a "what changed in the template" report possible.

## Ownership after handover

The document is output and the plan is the source. Hand edits go to the plan, and assembling overwrites the document. There is no freeze point. The consultant decides, and the plan may live in version control.

## Output

The assembled document is transitional, see [0025](0025-contracts-and-tests.md). The intended end state is the plan file and an onboarding helper flow, so these features belong to the document and not to the plan.

- Every assembled document carries a gap table with owner and timing, built when assembling, because values can change after a plan is made.
- Output carries the review marker until a person removes it.

## Secrets

The secrets check runs whenever a value is submitted and whenever a plan is validated, so a secret added by hand is caught. See [0010](0010-plan-editor.md).

## Consequences

- Opening a plan cannot fail because modules changed. The module set matters only when starting a plan or running a later update.
- A plan is larger. It carries sections that may not apply to the organization.
- The plan describes its own walkthrough, so any tool version can reopen it.
- The domain model is now the file format in memory. It sits at the slowest rate of change in 0020, and changes to it need an ADR with a migration story.
- The document assembler fills placeholders and uses the section selector and the section resolver. The plan instantiator selects nothing and resolves nothing. 0010 and 0020 are updated to match.
- The ACME example currently stores resolved sections and no `questions`. It needs to hold template sections, an inactive section, and `source` entries.

## Alternatives considered

- **Plans that hold only the selected sections.** The earlier design. Changing team state or discipline then needs the modules again, which meant an injected `expand` function and refusing the change on an opened plan.
- **Answers and a module reference only.** Smallest file. Breaks when the modules are missing or changed, and cannot hold engagement-specific content.
- **Fully resolved sections.** Changing one value would mean hand-editing every place it appears, or needing the modules to regenerate.
- **Store both unresolved and resolved sections.** Two copies drift apart when someone edits one.
- **A git sha as the module hash.** See Identifiers.
- **Prune inactive sections on save.** Smaller file, but changing an axis later needs the modules again. Left as a later, explicit action.

## Open questions

- How does the update action present a conflict, and what does it do by default?
- Where does per-person progress live later? I lean toward a separate file per person, so a shared plan does not collect merge conflicts. Deferred beyond the MVP.
