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

The newcomer's onboarding record holds a top-level `outcomes` list, see [0090](0090-plan-and-onboarding-record.md). The plan never does. Each entry is the outcome of one section.

```yaml
outcomes:
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

A refresh replaces the plan copy and leaves the record's `outcomes` alone. An outcome is **changed since it was recorded** when its `sectionHash` differs from the section's current hash. That comparison is made when the outcomes are read, so the reconciler of 0090 needs no rule beyond keeping the list. An outcome for a section that no longer exists stays, and a prune may remove it.

### `requires` replaces `after`: dependencies are a graph

A section may list other sections it requires, by id.

```yaml
- id: deploy-to-staging
  requires: [aws-access, clone-and-run-tests, environment]
```

- A section comes after the sections it requires, so `requires` also does what `after` did for ordering. The order is `timing`, then `requires`, then `id`.
- A section's verification is out of reach until every section it requires is `verified`. A required section that is only `deferred` keeps the dependent waiting, and the walk says what it waits on. A requirement cannot be skipped. If it is not really required, the onboarder removes it from `requires`.
- A section can be required by several others, so a shared prerequisite is written once. This is a graph and not a tree. `parent` stays for containment: grouping, and the switch cascade of 0070.
- **A section that requires an inactive section is inactive too.** It cascades the way a sub-section follows its parent. If a person does not need AWS access, they do not do the deploy. Something that applies only to some people goes in its own section, as 0065 says. The review shows the reason, such as "Deploy is off because AWS access is off".
- A cycle, and a required id that is not a section, are errors in the section validator.
- There is no migration. Nothing has been released.

### A bootstrap comes first

A section may carry `bootstrap: true`. A bootstrap section can be done with only a public copy of the tool and the plan file that the onboarder sent, such as getting access to the organization's repository.

- A bootstrap section may not require a section that is not a bootstrap section. The section validator checks this.
- The walk shows the bootstrap sections first, and says they can be done before the person has access.
- When they are verified, the tool offers to check for the latest plan. That is the refresh of 0075, run on request from the location where the org plan lives.
- The org plan has a top-level field, `planUrl`, where its latest copy lives, a path or an HTTPS address. It is set when the plan is created, the way the `id` is, and the onboarder can change it. A record keeps the same `planUrl` in its plan reference, so `start` can refresh from the record alone, with no `--org-plan` given. It is a field of the file, and not an answer to a question.
- The assembled document remains the way to send the first steps to someone who has nothing installed.

### A shared repository is stated, not hidden

The newcomer's file holds their answers and notes. If the file sits in a git working tree, anyone with read access to that repository can read them. The tool says so.

- The driver sees that the plan file is inside a working tree by finding a `.git` folder above it. That is a look at the file system. The tool never runs git.
- When it does, the start screen says plainly that the file is in a shared repository and that its answers and notes are visible to anyone who can read it. The person acknowledges it with enter.
- The screen offers "Do not tell me again for this repository". The choice is kept per repository, in the person's own config folder, and never in the plan or in the repository. Silencing one repository does not silence another.
- The header carries a "shared" marker for the whole walk, whether or not the notice was silenced. The note and `differently` fields say that their text is visible to people who can read the repository.

### Where the results live

Results are shared by file. The personal file is self-contained and does not care how it travels, so git is one transport, and a folder on a network share or a synced drive is another.

- **The path comes from the org plan.** The org plan has a top-level field, `resultsPath`, a template such as `records/{account}/record.obk.yaml`. `{account}` is the only variable. A personal file lives at the filled-in path under the root that `start` is given, or under the working tree it finds by looking for a `.git` folder above the current folder. Distinct paths per person mean the files never conflict.
- **The tool places the file and does not run git.** It writes the file at the right path, then prints the exact commands to add, commit, and push it, and offers to copy them. The person runs them. Running git from the tool would be a confirmed action under the rules for running commands (#29), and it is not part of this stage. Protected branches, signed commits, credentials, and merge conflicts make it a product of its own.
- **The onboarder seeds the account.** A principle, soft but real: the onboarder does what they can so the newcomer does less. An `invite` command makes a personalized copy of the org plan for one account, with the account answered and the results path filled in, for the onboarder to send. The newcomer sees "Is this you?" and confirms. The confirmation is the safety against a typo.
- **The account is the key.** It is permanent in the path. A mistake is fixed by an explicit move: the tool offers to move the file to the new path and the answer changes with it. Git sees a delete and an add.
- **Visibility is stated and accepted.** Anyone who can read the repository can read every file in it. The organizations this serves are developer teams where onboarding results are useful to everyone, so one shared repository is acceptable, with the notice above. An organization that wants more uses a repository per cohort or per team, so that readership is the people involved. Sparse checkout is not a control: it limits what the working tree shows and not what a person can fetch, so the tool does not offer it as privacy.
- **A database is not needed yet.** It would add a service to host, with authentication, backups, and an outage that blocks newcomers, and the tool is one self-contained file. What it would buy is write-without-read, live dashboards, and queries. Revisit it when an organization needs write-without-read or outgrows a folder of files. The file format does not change if a later "submit" sends the same file to an endpoint.

**Review belongs to the org plan.** Diffs and pull-request review matter for the org plan, because it changes the process for everyone. A person's own results file needs neither. The simplest reliable arrangement is two places: the org plan in a repository where changes are reviewed, and the results folder where newcomers can push directly. A separate results repository does that with no special rules. Some hosts can also restrict review to one path, with code owners or path-based push rules, but support differs, so the tool does not depend on it.

**Recorded option: encrypted results.** Not built now, because the organizations this serves are developer teams where results are useful to everyone, and encryption works against that. If an organization wants private notes, the first thing to try is encrypting only the sensitive text, `note` and `differently`, to a set of recipients: the onboarders, whose public keys the org plan would list, and the newcomer. Outcomes, times, and section ids stay plain, so everyone can still see who has verified what, and diffs of progress stay readable. A library such as age fits. The costs are key management: a newcomer's key can be lost, a new onboarder cannot read older notes until someone who can decrypt them re-encrypts, and removing a recipient does not revoke what was already pushed. The file loader should be able to read an encrypted wrapper later. A binary format on its own gives no privacy and loses the readable file of 0030, so it is not an option.

A later `report` command would read a folder of onboarding records and answer questions such as who has not verified a section. See [0095](0095-reports.md).

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

- **The domain model changes.** A Section gains `requires` and `bootstrap`, and loses `after`. A Plan gains `planUrl` and `resultsPath`, and the new Record of 0090 holds the `outcomes`. This needs `slow-change`.
- **Contracts**, each its own small increment: the decoder, encoder, loader, and hasher (`requires` and `bootstrap` are content), the section selector (ordering and the cascade), the section validator (cycles, unknown ids, the bootstrap rule), the refresher (outcomes survive), the plan validator (a duplicate or unknown outcome), and two new components: an outcome recorder, and a section walker that gives each active section its state, what it waits on, and its resolved text.
- **The reference docs** for ordering, optional sections, and the schema change for `requires`.
- **The TUI** gains a Process screen for the onboardee, the start notice for a shared repository, and the bootstrap offer to refresh.
- **The drivers** find the working tree, keep the per-repository choice, and write the file at its path. A new `invite` command makes a personalized copy.

## Not decided

Nothing is left open here. How the onboarder sees newcomers' outcomes is a report, and its open questions are in [0095](0095-reports.md).
