# 0090. The plan and the onboarding record

Status: proposed

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
account: dnawara
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

### The invitation

The onboarder must be able to send a newcomer something that works before the newcomer has any access. The plan's `planUrl` may be behind that access, so an invitation carries the plan itself.

- An `invite` command makes one file for an account: the plan with a seeded record, in the same joined form that the tool already reads and writes as a plan file.
- On first run, `start` splits it. The plan goes to the plan copy, and a record is created at `resultsPath`.
- The newcomer sees "Is this you?" for the account and confirms, as 0085 says.

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
- **Drivers.** `start` reads the plan copy and the record, joins them, saves the record on every change, and creates the files on first run. A new `invite` command writes the invitation.
- **The existing refresher** and its tests are retired. Nothing is released, so there is no migration.
- **The decisions** 0060 and 0075 carry a note that this replaces them in part.

## Not decided

- The record's file name and a marker that tells a record from a plan, such as a `kind` key or a file extension.
- Whether the joined invitation file needs a marker so that `start` knows to split it.
- How a record that points at a plan the person no longer has gets that plan back.
