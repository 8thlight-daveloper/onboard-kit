# 0075. Refreshing a personal plan

Status: accepted

Builds on [0030](0030-self-contained-plan-file.md), [0055](0055-variable-audience.md), and [0060](0060-personal-plan.md). It replaces the "running again" part of 0060.

## Context

An organization publishes an org plan. Each newcomer has a personal plan made from it. The organization changes its plan over time: a new system to get access to, a new contact, a corrected step. Newcomers need those changes without losing their own answers, and without anyone being told something stale. A contact who changed roles must not stay in someone's file.

"Update" has meant two different relationships.

| Relationship | Who acts | Conflicts |
| -- | -- | -- |
| A newer template set into the org plan | An onboarder | Real. The onboarder may have edited a section by hand. This is the three-way problem 0030 defers |
| The org plan into a personal plan | A newcomer | Almost none, because the two sides own different things |

This ADR is about the second.

## Decision

### Ownership is disjoint

- The **onboarder** owns the content: sections, variables, the directory, and the answers to onboarder questions.
- The **onboardee** owns their own answers to onboardee questions and their progress.

Nobody edits the other side's part. So a refresh needs no conflict resolution in the usual case.

### The personal plan stays a full plan file

A personal plan holds a full copy of the org plan's content, plus the newcomer's answers and progress. It opens with no network, as 0030 requires. The org plan it was last refreshed from sits beside it as the cached org plan.

### A refresh is a pure function

`refresh(base, theirs, ours)` takes the cached org plan (`base`), the newly fetched org plan (`theirs`), and the personal plan (`ours`). It returns a new personal plan and a report of what changed. It reads no file and no network. The driver fetches and writes.

| Change in the org plan | What the refresh does |
| -- | -- |
| Content: sections, variables, the directory | Taken from `theirs` |
| An answer to an onboarder question | Taken from `theirs`. The org wins |
| An answer to an onboardee question that the newcomer has not changed from `base` | Taken from `theirs`, so a changed pre-seed reaches them |
| An answer to an onboardee question that the newcomer changed | Kept. The newcomer wins |
| A new onboardee question | Appears unanswered |
| A question that disappeared | The answer is kept but hidden, not deleted, in case it returns. See pruning below |
| An answer that is no longer valid, such as a choice option that was removed or a team that no longer exists | Marked as needing review and asked again. Never kept silently |
| A new section or step | Appears as not done |
| A step the newcomer finished, in a section whose content changed | Flagged as changed since it was finished |
| A changed contact, role holder, or team | Applies at once, because answers store ids and names are resolved when read (0080) |

The three-way comparison is what tells a newcomer's own change from an old pre-seed. Without `base`, a pre-seed the organization later corrected could not be told from a choice the newcomer made.

### What the refresh does, exactly

Decided, and pinned by the plan refresher's contract.

- **Content from `theirs`.** Sections with their `source` and `enabled` switch, the variables, the directory, and the plan's unknown fields. The personal plan keeps its own `id`, `schemaVersion`, and `provenance`. Its `origin` is set from the new origin the driver passes in, with the `planId` taken from `theirs`.
- **Each question's audience is read from `theirs`.** A question with no audience is an onboarder question.
- **Onboarder answers** are taken from `theirs`. An answer `theirs` does not hold clears ours.
- **Onboardee answers** are taken from `theirs` when ours equals `base`, and kept otherwise. Equal means the same kind and content. List order counts, and `""`, `[]`, and `false` each differ from absent. An empty `base` makes every answer the newcomer has count as changed, so a missing cache never overwrites a newcomer.
- **A kept onboardee answer that is no longer valid** is removed and listed for review. Invalid means the wrong kind for the variable, a `choice` value no longer in its options (matched exactly), or a `contact`, `role`, or `team` id that is not in `theirs`' directory. For a `multiple` list only the missing ids are dropped, and the question is still listed for review. Deferred and empty answers are always valid. An answer taken from `theirs` is not checked.
- **Hidden answers** are answers for a question `theirs` does not declare. They stay untouched, and come back if the question returns. A question that moves to another scope is a removed question and a new one.
- **Refusal.** The refresh returns the personal plan unchanged, with problems, when `theirs` has no `id` (`refresh.missing-plan-id`) or its `id` differs from `ours.origin.planId` (`refresh.different-plan`). A first refresh has no `planId` yet and is accepted.
- **The report** lists new questions, removed questions, answers taken from `theirs`, answers kept, answers to review, new sections, and removed sections, as sorted ids with no repeats. A list that lost some ids names the question and not the ids.
- **Refreshing again** with the same `theirs` changes only the origin.

### How `start` runs it

- The content hash in `origin` is the hash of the org plan as the encoder writes it. The source is the absolute path or the URL.
- `start` refuses an org plan with no `id`. `new` makes one, a random UUID by default, and takes `--id` for a fixed one.
- The first run makes the personal plan from the org plan, with no `id` of its own, and saves the org plan beside it as the cache. Later runs refresh and then print a short summary of what changed.
- A missing cache counts as an empty `base`.

### Identity and origin

- An org plan has an `id`, a UUID made when the plan is first created. The driver generates it, since the core makes no random values.
- A personal plan records the org plan's `id`, where it came from (a path or an HTTPS URL), and the content hash it was last refreshed from. A different `id` means a different organization, and the refresh refuses.
- Ids of sections, variables, teams, roles, and contacts are permanent. Renaming means a new id. Reusing an id would attach one person's answers to a different question.

### When a refresh happens

`start` refreshes whenever it can reach the origin. When it cannot, it uses the cached org plan and says "org plan as of ⟨date⟩", so an old copy is never silent. A refresh never runs without the person starting the tool.

### Hosting

The org plan is a file at a path or an HTTPS URL. The tool reads it and never publishes it. Versioning and rollback belong to the host. Keeping the org plan in a git repository is strongly recommended: it gives history, review, merging, and rollback. The tool does not require git, and it never runs it. A URL that serves a directory of people's emails (0080) must require authentication. Credentials never go in the plan.

Several onboarders may edit the org plan, so the file must merge cleanly. Lists are sorted by id and keys are in canonical order, as the encoder already does.

### Hidden answers stay until the owner prunes them

A refresh never deletes an answer on its own. Hidden answers would collect over time, so each file's owner can prune them whenever they choose. A prune action lists the hidden answers first and deletes only after a yes. The onboarder prunes the org plan, and a newcomer prunes their own file. A newcomer's hidden answers are not in the org plan, so the onboarder's prune cannot touch them.

### What is not locked

The personal plan holds a copy of org content, and a person could edit it by hand. The next refresh replaces that content from the org plan, so a hand edit does not last. The tool does not try to stop an edit, and it does not trust one. This replaces the answers-hash lock in 0060.

## Why

- Disjoint ownership makes the common case a pure function with nothing to resolve by hand.
- A full copy keeps every existing component unchanged and the file readable on its own. Splitting the file into answers only, and joining it at read time, added a layer and a question about where each write goes, and saved little.
- The cached org plan costs one file and makes the three-way comparison possible.
- Permanent ids make references safe across refreshes.

## Consequences

- **The domain model changes.** A Plan gains an `id`. A personal plan records its origin, the hash it was refreshed from, and the time. This needs `slow-change`.
- **A new component**, the plan refresher, a pure function with its own contract.
- **`start` changes.** Run again, it refreshes, then reports what changed. The TUI shows the report, such as "2 new questions, 1 step changed".
- **The plan validator** gains a check that ids are unique within the plan.
- **0060.** The "running again" and "locked content" parts are superseded. The two-hash lock is dropped.

## Not decided

- The progress record's format, which 0065 introduces.
- How an update from a template set into an org plan shows a conflict. It stays deferred.
- Rolling back to an earlier org plan.
