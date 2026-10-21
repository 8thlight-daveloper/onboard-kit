# 0085. Walking sections and recording outcomes

Status: proposed

Builds on [0030](0030-self-contained-plan-file.md), [0050](0050-section-switch.md), [0055](0055-variable-audience.md), [0060](0060-personal-plan.md), [0065](0065-messages-and-evidence.md), [0070](0070-sub-sections.md), and [0075](0075-refreshing-a-personal-plan.md). It answers the progress record that 0065 and 0075 leave open, and it is the first stage of the onboardee runtime, [#28](https://github.com/8thlight-daveloper/onboard-kit/issues/28). It replaces the ordering field `after` with `requires`.

## Context

The onboardee's walk asks questions and stops there. Nothing records whether a newcomer got what they needed, so they cannot see their progress, and the organization cannot see who still lacks access or whose instructions did not work.

The newcomer's side of a process has four parts: what they need, what to do to get it, how to check that they have it, and whether it worked. The plan already holds the first three. A section's title, summary, and body say what the person needs. Its steps are the instructions. Its `verification` text and the checks on its steps say how to tell. The fourth part, the answer to "did it work", is missing.

## Decision

### A process is a section

- The section is the unit the newcomer walks. Its steps are instructions, done in any order the person finds sensible, and they are verified as a group, once, at the end.
- The tool records no state for a single step. A step is not "done" on its own. The section is verified, or it did not work, or the person has put it off.

### The outcome of a section

A personal plan may hold a top-level `progress` list. The org plan never does. Each entry is the outcome of one section.

```yaml
progress:
  - section: access
    outcome: verified
    on: "2026-10-21T14:05:00Z"
    sectionHash: 5d3c...
    differently: The IT desk wanted the request in their channel, not by email
  - section: clone-and-run-tests
    outcome: failed
    on: "2026-10-21T15:30:00Z"
    sectionHash: 91ab...
    note: The tests fail on a fresh clone
```

- `outcome` is `verified`, `failed`, or `deferred`. A section with no entry has not been tried. `deferred` uses the word from [0035](0035-settled-and-deferred-answers.md): the person did what they could and cannot verify yet.
- `on` is a UTC time in RFC 3339 form. A driver passes it in, because the core reads no clock.
- `sectionHash` is the section's content hash when the entry was made, the `source.hash` the section carries.
- `note` is optional text. `differently` is optional text for what the person had to do differently from the instructions. A section can be verified and still have a `differently`.
- Each section has at most one entry, and entries are sorted by `section`.
- The newcomer does not edit the plan's content. A correction goes in `differently`, so it survives a refresh and the onboarder decides what to change. This keeps the ownership rule of 0075.

### A failed section shows who to ask

A `failed` outcome shows the section's `escalation` list, resolved from the directory as 0080 says, with vacant roles shown as vacant. The person can also copy a message to the section's owner, once messages (0065) exist.

### A refresh keeps outcomes and flags what changed

A refresh takes section content from the org plan and leaves `progress` alone. An outcome is **changed since it was recorded** when its `sectionHash` differs from the section's current hash. That comparison is made when the progress is read, so the refresher needs no rule beyond keeping the list. An outcome for a section that no longer exists stays, and a prune may remove it.

### `requires` replaces `after`: dependencies are a graph

A section may list other sections it requires, by id.

```yaml
- id: deploy-to-staging
  requires: [aws-access, clone-and-run-tests, environment]
```

- A section comes after the sections it requires, so `requires` also does what `after` did for ordering. The order is `timing`, then `requires`, then `id`.
- A section's verification is out of reach until every section it requires is `verified`. A required section that is only `deferred` keeps the dependent waiting, and the walk says what it waits on.
- A section can be required by several others, so a shared prerequisite is written once. This is a graph and not a tree. `parent` stays for containment: grouping, and the switch cascade of 0070.
- **A section that requires an inactive section is inactive too.** It cascades the way a sub-section follows its parent. If a person does not need AWS access, they do not do the deploy. Something that applies only to some people goes in its own section, as 0065 says. The review shows the reason, such as "Deploy is off because AWS access is off".
- A cycle, and a required id that is not a section, are errors in the section validator.
- There is no migration. Nothing has been released.

### A bootstrap comes first

A section may carry `bootstrap: true`. A bootstrap section can be done with only a public copy of the tool and the plan file that the onboarder sent, such as getting access to the organization's repository.

- A bootstrap section may not require a section that is not a bootstrap section. The section validator checks this.
- The walk shows the bootstrap sections first, and says they can be done before the person has access.
- When they are verified, the tool offers to check for the latest plan. That is the refresh of 0075, run on request from the location where the org plan lives.
- The org plan records where its latest copy lives, a path or an HTTPS address, so that the copy sent by email knows where to refresh from. The field's name is not decided.
- The assembled document remains the way to send the first steps to someone who has nothing installed.

### A shared repository is stated, not hidden

The newcomer's file holds their answers and notes. If the file sits in a git working tree, anyone with read access to that repository can read them. The tool says so.

- The driver sees that the plan file is inside a working tree by finding a `.git` folder above it. That is a look at the file system. The tool never runs git.
- When it does, the start screen says plainly that the file is in a shared repository and that its answers and notes are visible to anyone who can read it. The person acknowledges it with enter.
- The screen offers "Do not tell me again for this repository". The choice is kept per repository, in the person's own config folder, and never in the plan or in the repository. Silencing one repository does not silence another.
- The header carries a "shared" marker for the whole walk, whether or not the notice was silenced. The note and `differently` fields say that their text is visible to people who can read the repository.

### Sharing is by file

Where the file lives is the newcomer's choice and the organization's convention. An organization that wants to see progress asks newcomers to keep their files in a shared checkout, and `start` can be pointed at it. A later `report` command would read a folder of plans and answer questions such as who has not verified a section. It is a pure function from plans to a table. It is not designed here, and a due date on a section would belong to it.

### The tool runs nothing in this stage

The action runner, [#29](https://github.com/8thlight-daveloper/onboard-kit/issues/29), and its trust rules come later.

- An `ask` check is put to the person.
- A `run` check is shown with its command. The person runs it themselves and says whether it passed.
- An `env` check is shown by name. The person says whether it is set. The tool never reads the value.

A check is advice that the person confirms, and the outcome is the person's own statement.

## Why

- A section already is the process, so the record follows what the content already says.
- One list in the one plan file keeps the plan self-contained, as 0030 requires.
- A section hash makes "this changed since you verified it" a comparison, so the flag can never be stale.
- A graph handles a shared prerequisite, and the cascade keeps a dependency from being optional in disguise.
- Stating the privacy of a shared file at the moment it matters costs one keypress.

## Consequences

- **The domain model changes.** A Section gains `requires` and `bootstrap`, and loses `after`. A Plan gains `progress` and the location of its latest copy. This needs `slow-change`.
- **Contracts**, each its own small increment: the decoder, encoder, loader, and hasher (`requires` and `bootstrap` are content), the section selector (ordering and the cascade), the section validator (cycles, unknown ids, the bootstrap rule), the refresher (outcomes survive), the plan validator (a duplicate or unknown outcome), and two new components: an outcome recorder, and a section walker that gives each active section its state, what it waits on, and its resolved text.
- **The reference docs** for ordering, optional sections, and the schema change for `requires`.
- **The TUI** gains a Process screen for the onboardee, the start notice for a shared repository, and the bootstrap offer to refresh.
- **The drivers** find the working tree and keep the per-repository choice.

## Not decided

- The name of the field that records where the latest org plan lives, and whether it is an onboarder answer or a plan field.
- Whether a person can verify a section early, with a warning, when it is only waiting on a deferred requirement.
- Whether the report is a command or a screen, and how due dates are written.
- The evidence patterns for Teams and email, which 0065 leaves open.
