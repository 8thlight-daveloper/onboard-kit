# 0030. Self-contained plan file

Status: proposed

Relates to [#24](https://github.com/8thlight-daveloper/onboard-kit/issues/24), [#23](https://github.com/8thlight-daveloper/onboard-kit/issues/23), [#6](https://github.com/8thlight-daveloper/onboard-kit/issues/6). Builds on [0010](0010-plan-editor.md) and [0020](0020-binary-structure-and-regenerative-components.md).

## Context

A plan ends up at the organization it was made for, in whatever tool version and module set exist later. It may hold content that no module has, because each engagement can add its own sections and questions. If reopening a plan needed the modules it was made from, we would have to decide what happens when those modules are missing or have changed, and a plan could break through no fault of its own.

## Decision

The saved file is the plan. There is no separate configuration file. A plan is named `<name>.obk.yaml` and holds everything needed to edit and assemble it:

- `schemaVersion`
- `provenance`: the module set version and the generated date
- `answers`: team state, discipline, and the default output style
- `questions`: every question the walkthrough asks, with its description, type, options, default, whether it is required, and any condition for asking it
- `values`: the answers to those questions
- `sections`: the content, with placeholders left in place

Sketch. Field names are not final, see [#3](https://github.com/8thlight-daveloper/onboard-kit/issues/3):

```yaml
schemaVersion: 0
provenance:
  moduleSet: example
  generatedOn: 2026-10-15
answers:
  teamState: established
  discipline: software
  outputStyle: github
questions:
  - id: repo_url
    description: Clone URL of the main repository
    required: true
values:
  repo_url: git@git.acme.example:platform/widget-service.git
  tracker_grantor: unknown-yet
sections:
  - id: access
    title: Access
    steps:
      - id: source-control
        name: Source control
        check:
          run: [git, ls-remote, $repo_url]
```

### Sections keep their placeholders

A section holds its text with `$name` placeholders unresolved. Values fill them when the document is assembled. Nothing resolved is stored. This is what lets a person change a value, or edit a section's text by hand, and get a correct document without the original modules.

### Modules are used once

The plan instantiator turns modules and answers into a plan. It selects the modules that apply, copies each one into a section without its selection fields, builds the question list from their placeholders, and records provenance. After that the plan stands alone.

| Needs modules | Does not need modules |
| --- | --- |
| Starting a new plan | Editing values |
| Changing team state or discipline, since that changes which sections apply | Replacing an `unknown-yet` value |
| Updating a plan from a newer module set, deferred until after the MVP | Changing the output style |
| | Editing section text by hand |
| | Assembling the document |

An update, when it exists, is an action the consultant chooses. It compares the plan's provenance with a newer module set and reports what changed. It never runs on load, so opening a file cannot fail because the modules moved on.

### Compatibility

- The plan decodes leniently. A load and save by an older tool preserves fields it does not know. Modules still decode strictly.
- The plan carries provenance, so a later report of what changed in the template is possible.

### Ownership after handover

The document is output and the plan is the source. Hand edits go to the plan, and assembling overwrites the document. There is no freeze point. The consultant decides, and the plan may live in version control.

### Output

- Every assembled document carries a gap table with owner and timing, built when assembling, because values can change after a plan is made.
- Output carries the review marker until a person removes it.

### Secrets

The secrets check runs whenever a value is submitted. Validating a plan runs its values through the same check, so a secret added by hand is caught. See [0010](0010-plan-editor.md).

## Consequences

- The question of what to do with a plan made from an older module set goes away at load time. It returns only as the optional update action.
- The plan is larger and describes its own walkthrough. Any tool version can reopen it.
- The domain model is now the file format in memory. It sits at the slowest rate of change in 0020, and changes to it need an ADR with a compatibility story.
- The document assembler fills placeholders, so it depends on the section resolver. The plan instantiator no longer resolves anything. 0020 is updated to match.
- The ACME example currently stores resolved sections. It needs to keep placeholders and add a `questions` list.

## Alternatives considered

- **Answers and a module reference only.** Smallest file. Breaks when the modules are missing or changed, and cannot hold engagement-specific content.
- **Fully resolved sections.** Changing one value would mean hand-editing every place it appears, or needing the modules to regenerate.
- **Store both unresolved and resolved sections.** Two copies drift apart when someone edits one.

## Open questions

- What does a tool do with a plan whose `schemaVersion` is newer than it knows? Preserve and warn, or refuse?
- How do we design the update action, and how does it treat hand edits? Copier's handling of local edits, in the prior art note, is the place to start.
- Where does per-person progress live later? I lean toward a separate file per person, so a shared plan does not collect merge conflicts. Deferred beyond the MVP.
- How is the condition for asking a question written? This belongs in the module schema ([#3](https://github.com/8thlight-daveloper/onboard-kit/issues/3)).
