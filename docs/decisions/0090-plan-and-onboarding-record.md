# 0090. The plan and the onboarding record

Status: accepted

Builds on [0030](0030-self-contained-plan-file.md), [0055](0055-variable-audience.md), [0080](0080-directory.md), and [0085](0085-walking-sections-and-recording-outcomes.md). It replaces the personal plan of [0060](0060-personal-plan.md), and it replaces the part of [0075](0075-refreshing-a-personal-plan.md) that keeps the personal plan as a full copy of the org plan.

## Context

0075 kept each newcomer's file as a full plan file: a complete copy of the org plan with their answers in it. That made the file open offline and the refresh a pure function over three plans.

0085 puts those files in a shared repository so that an onboarder can see who has verified what. A folder of such files has three problems.

- Every file repeats the whole org plan, so a folder of N people holds N near-identical copies and is noisy to browse.
- A copy changes only when its owner refreshes, so the files drift apart, and a report cannot trust them for content.
- What a person actually produced is a small part of each file.

## Decision

### Two artifacts

| Artifact | Holds | Authored by | Where it lives |
| -- | -- | -- | -- |
| **The plan** | Sections, variables, the directory, the onboarder's answers, `id`, `planUrl`, `resultsPath` | The onboarder | The organization's repository, changed by review. Each machine keeps a **plan copy** |
| **The onboarding record** | The newcomer's answers, the outcomes of 0085, their notes, and a reference to the plan | The newcomer | At `resultsPath`, in the repository or folder the organization chose |

The word **report** names something else: a derived, read-only view over one or more onboarding records, computed on demand and stored nowhere, see 0085.

The words in older documents change as follows.

| Older word | Now |
| -- | -- |
| org plan | the plan |
| personal plan | the onboarding record, or the record |
| cached org plan | the plan copy |
| `progress` | `outcomes` |
| `origin` | the record's `plan` reference |

### The record's shape

```yaml
schemaVersion: 1
kind: record
account: dnawara
startDate: "2026-10-22"
plan:
  id: 6f1c2d3e-4a5b-4c6d-8e7f-90a1b2c3d4e5
  hash: 5d3c...
  url: https://git.acme.example/onboarding/plan.obk.yaml
answers:
  plan:
    discipline: software
    team: platform
  sections:
    access:
      account_name: dnawara
outcomes:
  - section: access
    outcome: verified
    on: "2026-10-21T14:05:00Z"
    sectionHash: 5d3c...
    differently: The IT desk wanted the request in their channel, not by email
```

- `kind` says this is a record and not a plan.
- `startDate` is the day the newcomer starts. The onboarder gives it when they make the invitation, and the due dates of 0095 count from it.
- `account` is the permanent key. It is in the path, and a fix is an explicit move, see 0085.
- `plan` says which plan the record was last reconciled against: its `id`, the hash of that plan, and where the latest copy lives.
- `answers` holds only the answers the newcomer gave, to onboardee questions. It never holds an onboarder answer, and the plan never needs the newcomer's.
- `outcomes` is the list of 0085.
- A record has no sections and no variables.

### The tool joins and splits

The components keep taking one plan value, so none of them changes. At load, the driver joins the plan copy with the record into one plan in memory. At save, it writes the newcomer's parts to the record and the plan copy is left alone.

- The join puts the record's answers over the plan's answers. A plan may pre-seed an onboardee question, and the record's answer wins.
- An onboardee answer that the newcomer never gave is not in the record. It follows the plan, so a pre-seed the organization changes reaches the newcomer by itself. This is the case 0075 needed a three-way comparison for. The comparison is no longer needed.
- All onboardee writes go to the record, and all onboarder writes go to the plan. An onboarder walks the plan alone, with no record.
- The join and the split are pure functions.

### Where things are stored

- The **plan copy** is in the person's per-user state folder: Application Support on macOS, and the state folder on Linux. It is the person's data and must not vanish, because the emailed copy may be the only one they have until they gain access. It is not a cache that can be deleted.
- The **record** is at `resultsPath`, filled in with the account, under the root the person gave `start` or the working tree `start` finds. Without a `resultsPath` it is in the same state folder.
- **Preferences**, such as "do not tell me again" for a repository, are in the per-user config folder.

### Refreshing becomes reconciling

To refresh, the tool fetches the latest plan from `planUrl` (or reads a path given), replaces the plan copy, and reconciles the record against it. A reconcile never changes the plan.

- An answer in the record that the new plan no longer accepts, such as a removed `choice` option or a contact id that is gone, is removed and listed for review.
- An answer for a question the plan no longer declares stays, as a hidden answer, until the owner prunes it.
- New questions appear as unanswered, and new and removed sections are listed.
- An outcome is flagged when its `sectionHash` differs from the section's current hash, as 0085 says.
- It returns a new record and a report of what changed, and it is a pure function.

The plan refresher of 0075 is retired and rebuilt as this smaller component.

### The invitation is an archive

The onboarder must be able to send a newcomer something that works before the newcomer has any access. The plan's `planUrl` may be behind that access, so the invitation carries the plan itself.

- An `invite` command makes one zip file for an account. It takes `--start`, defaults it to today, and prints the date it used. It holds the plan as `plan.obk.yaml`, a record seeded with the account and the plan reference as `records/<account>/record.obk.yaml`, and `first-steps.md`, a document for someone with nothing installed, see below. Both files are in their own canonical shapes, so there is no joined form and nothing to split.
- `onboard-kit start --invitation <file>` reads the archive directly. The newcomer extracts nothing by hand. The tool puts the plan in the plan copy and the record at `resultsPath`.
- The archive is untrusted input. The tool reads only those fixed entries, never a path taken from inside the archive, and refuses an archive of more than a handful of entries or more than 1 MiB uncompressed in all. Video and other large help is linked, not embedded.
- A record that already exists at that path is never overwritten. The tool reconciles it against the plan from the archive instead.
- **The account is confirmed first.** "Is this you?" is the first thing `start --invitation` does, before it writes anything. It shows the account and the path where the files would go. A wrong account may mean the onboarder already provisioned access or told others the name, so on no the tool writes nothing and says to ask the onboarder for a corrected invitation, naming the owner from the directory. A move of the file, see 0085, fixes a mistake found later.

### The first-steps document

The archive always includes `first-steps.md`, for a newcomer who has nothing installed yet. It is assembled for the onboardee, deterministically, with no model.

- It holds the bootstrap sections as the assembler writes them.
- It ends with a block of standard help that the tool supplies and the plan fills in: what to do about a lost plan (ask the onboarding owner for a fresh invitation, or clone the repository and run `start` inside it), a lost record (it is in the repository at its path), a wrong account (the file can be moved), where to get the tool, and who to ask when stuck, from the plan's general contacts and the owner resolved from the directory.
- The standard wording is part of the tool.
- The parts are listed in the [first-steps document](../reference/first-steps-document.md).

### Several records on one machine

A newcomer has several records when they have several plans: the consultancy case of one client after another, a return after leaving, or being a newcomer in one plan and an onboarder in another. A change of team does not need a second record, because the team is an answer in the same record. Inside one plan's repository there is one record per account.

- `start` with nothing given lists the records known on this machine, by organization and team, with progress such as "4 of 7 verified" and the last activity, and the person picks one. They can also start a new one from an invitation.
- Records live in different repositories, so the tool keeps a small local index in the per-user state folder: for each record the person has opened, the plan id and the path. The index is not in the plan or in any repository.
- **A missing file is shown, not removed.** A record whose file is gone appears in the picker as missing, and the person can choose to forget it. The tool never removes an entry on its own, because a missing file may be an unmounted drive or a repository not yet cloned.

### The record's name and marker

- A record is kept at `resultsPath`, which by default is `records/{account}/record.obk.yaml`. The folder per account leaves room for another record for the same person later, such as a transfer to another team, with no renaming. The path is a convention.
- A record also carries `kind: record`, and a plan carries `kind: plan` or no `kind`. The decoder refuses a file of the wrong kind with a message that says which it is, so a command given the wrong file does not guess.

### Getting a lost plan back

A record names its plan in `plan.url`. When the plan lives in the same repository as the records, the url may be a path relative to the repository root. A newcomer on a new machine clones the repository and runs `start` inside it, and the tool finds the record by account and the plan by that path. Without repository access, the onboarder sends a fresh invitation. Getting access is what the bootstrap sections are for.

### Reports read records

A report reads records and the plan, and nothing else. It does not need each person's copy of the content, because the record's `plan.hash` says which plan they were on, and the plan gives the structure and the due dates.

## Why

- The repository holds only what people produced, so a folder of records is small, readable, and free of drift.
- Content and state have separate owners and separate places. That matches the ownership rule of 0075: the plan is the onboarder's, and the record is the newcomer's.
- A refresh no longer overwrites a file the newcomer owns.
- Nothing downstream changes, because the components still take one plan.

## Consequences

- **The domain model changes.** A Record type holds the account, the plan reference, the answers, the outcomes, and unknown fields. This needs `slow-change`.
- **Components**, each its own small increment: a record decoder and encoder, a joiner and a splitter, a reconciler that replaces the plan refresher, and the outcome recorder and section walker of 0085. The plan decoder and encoder lose `origin`, and a plan keeps its `id`, `planUrl`, and `resultsPath`.
- **Drivers.** `start` reads the plan copy and the record, joins them, saves the record on every change, and creates the files on first run, from the plan copy or an invitation. A new `invite` command writes the invitation archive, and `start` reads it with `--invitation`.
- **The existing refresher** and its tests are retired. Nothing is released, so there is no migration.
- **The decisions** 0060 and 0075 carry a note that this replaces them in part.

## Later

- **Offboarding.** Consultants leave every engagement, so ending matters. A newcomer's verified outcomes say what access they were given, so a reversed pass over those sections is a revocation checklist. A report could then say who still holds access to a system. A record could also be closed, so a finished engagement is kept and marked. Nothing here is designed.

## Not decided

- The wording of the standard help in the first-steps document. The parts and their sources are decided, in the [first-steps document](../reference/first-steps-document.md).
