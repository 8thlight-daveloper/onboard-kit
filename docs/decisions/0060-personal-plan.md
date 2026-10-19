# 0060. Personal plan

Status: accepted

Builds on [0030](0030-self-contained-plan-file.md), [0005](0005-self-contained-install.md), [0015](0015-functional-core-and-imperative-shell.md), and [0055](0055-variable-audience.md).

## Context

An onboarder builds a plan for an organization. Each onboardee answers their own questions in the TUI, on their own machine. The organization may add a system later, such as AWS access, and ask everyone to run the tool again to pick it up. An onboardee can also change teams, which changes what applies to them.

## Decision

### Two kinds of plan file

Both are plan files in one format. A file name says which kind it is.

| Term | File | Written by |
| -- | -- | -- |
| **Org plan** | `<name>.org.obk.yaml` | The onboarder. Published at an HTTPS URL or handed over as a file |
| **Personal plan** | `<name>.obk.yaml` | The tool, on the onboardee's machine. It holds their answers |

A personal plan is self-contained, as 0030 requires. It holds the org plan's sections and the onboarder's answers, with the onboardee's own answers. It opens with no network.

### Identity

An org plan has an `id`, a UUID that the tool generates when the plan is first made. The core does not make random values, so the driver generates it and passes it in, as it does the date. A personal plan records the `id` of its org plan, the URL it came from if there was one, and the hashes below. This is how a personal plan finds its org plan again.

### Starting and running again

- **Starting.** `onboard-kit start <path or https URL>` reads the org plan and writes the personal plan. It then opens the TUI on the onboardee's questions.
- **Running again.** The tool refreshes the org plan from the URL, or reads the stored one when offline or when no URL was given. It compares the org plan with the personal plan and applies what changed: new sections, changed sections, changed onboarder answers. It keeps the onboardee's answers and asks the questions that are new or unanswered. This is the update action that 0030 deferred.
- **Network access** is in the driver. It accepts HTTPS only. The core reads no network and no file. A host that needs credentials is out of scope.

### Where the files live

Under a per-user folder for state, as the platform's convention says: `$XDG_STATE_HOME/obk` on Linux, and `~/Library/Application Support/obk` on macOS. Both files for one org sit side by side. The file name is a short name for the organization, taken from its `organization` variable when it has one and otherwise from the `id`. The `id` is always what matches a personal plan to its org plan. The folder rule is a stack doc detail and can change.

### Locked content

What the onboarder owns is the org plan's sections, its plan variable declarations, and the answers to onboarder questions. The onboardee does not change them in the TUI. The tool detects a change made by hand, with two hashes the personal plan records:

- **Each section's `source.hash`**, as 0030 defines it, which detects an edit to a section.
- **An answers hash** over the onboarder's answers in the org plan, which detects an edit to those.

A mismatch is an edit, and the tool replaces that content from the org plan and says so. This protects against accidents. It is not a security boundary, because the file belongs to the onboardee.

### A team change

The team is an answer. The onboardee edits it, and sections for the old team become inactive. Their answers stay stored, as with a section that is switched off ([0050](0050-section-switch.md)), so nothing is lost.

## Why

- One file format keeps the selector, the resolver, and the assembler unchanged. A personal plan is a plan.
- A rollout is an update of the org plan. No second mechanism is needed.
- Hashes already carry the "was it edited" question in 0030, so the lock needs only an answers hash.

## Consequences

- **The domain model changes.** A Plan has an `id`. Provenance gains the org plan's id, its URL when there is one, and its answers hash. This needs `slow-change`.
- **New components**, each its own contract:
  - A plan updater: a pure function from an org plan and a personal plan to a new personal plan and a report of what changed.
  - A hash of the onboarder's answers, in or beside the section hasher.
  - A `start` command in the driver, with the fetch.
- **Docs.** 0030's lines on updates and on newcomer answers, the schema docs, the architecture docs, and the glossary change. The update conflict question in 0030 is now needed for the first release.
- **Versioning.** As in 0035, nothing exists outside this repository, so the schema version stays 1.

## Out of scope

- Running the steps and checks, and the rule that each command is shown and confirmed before it runs. That needs its own decision before the runtime is built.
- Progress records for steps and checks, and telemetry.
- Credentials for a private URL.
- How an update shows a conflict beyond what 0030 says.

## Not decided

- Whether the TUI shows the onboarder's answers in a read-only view or hides them.
- The exact file name rule when the `organization` variable is absent or not safe in a file name.
