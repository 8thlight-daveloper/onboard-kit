# Prior art

Status: first pass, in progress. Tracked in [#30](https://github.com/8thlight-daveloper/onboard-kit/issues/30).

What already exists, what we can learn from it, and how onboard-kit differs. The first pass was shallow: four web searches and two page reads. It did not cover GitHub topic search or package registries. Treat every gap claim below as unconfirmed.

## What we looked at

### Copier

Copier asks questions, renders a template, and saves the answers in `.copier-answers.yml`. Running `copier update` re-renders the template with the saved answers, compares it with the version the project came from, and merges the changes while keeping local edits. Conflicts go to `.rej` files by default, with an inline mode available. The docs warn against editing the answers file by hand, because Copier then assumes those answers produced the project.

What it means for us: this is the closest match to our generation side and to the regeneration question in [#24](https://github.com/8thlight-daveloper/onboard-kit/issues/24). It handles template changes over hand-edited output, which we assumed was too complex to try.

Caveat: docs differ across versions. Check the current one before borrowing anything.

Source: [Copier, updating a project](https://copier.readthedocs.io/en/v9.7.0/updating/)

### Runme

Runme runs the shell commands in fenced code blocks of Markdown files. It has a CLI, a TUI, and a VS Code notebook mode. Blocks can carry names, so `runme run <name>` runs one.

What it means for us: it covers running commands from a document. It has no plan model, owners, timing, or tracking. If our renderer emits named fenced code blocks, documents might run in Runme without our tool.

Caveat: several forks and mirrors exist. Confirm which repository is current.

Sources: [Runme repository](https://github.com/elotl/runme), [CLI docs, via mirror](https://proxy-dev.blitzz.co/proxy/123456/docs.runme.dev/getting-started/cli)

### Atuin Desktop

A block-based editor for runbooks that mix documentation with executable steps, such as terminals, scripts, database queries, and HTTP calls. It supports variables and Jinja-style templating, and syncs through Git. It is in early beta.

What it means for us: it shows demand for documents that run. It is a desktop editor and is not onboarding specific.

Source: [Atuin Desktop docs](https://docs.atuin.sh/desktop/)

### Cortex developer onboarding guide

Describes onboarding in four phases: day one, days 1 to 30, 31 to 60, and 61 to 90. Items are mostly broad, and the access list is itemized. It calls the checklist a living document shaped by feedback, and it asks for a day-90 review with feedback from the new hire, peers, and managers. It does not name an owner for keeping the checklist current, and it does not describe verifying individual items.

We did not reuse its statistics. They come from third parties and we have not checked them.

Source: [Cortex developer onboarding guide](https://www.cortex.io/post/developer-onboarding-guide)

### Microsoft New Employee Onboarding sites

Three SharePoint site templates for pre-onboarding, corporate onboarding, and departmental onboarding, usable alone or together. They include an "onboarding journey", a configurable list of activities paced over 30, 60, and 90 days. Prepopulated pages are marked for the owner to keep, edit, or delete. The maintenance guidance names site owners, training for them, and a content audit schedule, and its KPI table assigns a business owner to each measure.

What it means for us: the activity list is item-level, like the steps in our [ACME example](reference/examples/acme/README.md). The layered templates resemble our core plus variants.

Source: [Microsoft Learn, New Employee Onboarding sites](https://learn.microsoft.com/en-us/sharepoint/provision-neo-hub)

### Port

Port publishes a developer onboarding checklist as vendor content. We did not read it in detail.

Source: [Port, developer onboarding checklist](https://getport.io/blog/developer-onboarding-checklist)

## What we did not find

An open source CLI that reads a versioned onboarding plan, generates it from questions, and tracks completion per person. Search results were mostly vendor blogs, shell script advice, and agent skills that generate or run setup steps. This needs confirming with a deeper search.

## How onboard-kit differs

| Aspect | Existing options | onboard-kit |
| -- | -- | -- |
| Generation | Copier renders templates from answers, with no onboarding content | Templates and answers produce an onboarding plan |
| Execution | Runme and Atuin run commands from documents | A walker runs steps only after confirmation, with a trust model |
| Structure | Hubs hold pages and checklists | Steps carry owner, timing, action, and check |
| Delivery | SaaS portals and SharePoint templates | Open source, with a plan file that lives in the client's version control |
| Maintenance | Advice to audit, rarely tooling | Owners per step, stale flags, provenance, and a diff against the template |

## Borrow, adopt, skip

| Item | Decision |
| -- | -- |
| Copier's answers file and update flow | Study before settling the regeneration rule |
| Runme compatibility through named fenced code blocks | Check whether it comes for free |
| Microsoft's marker for template content that needs review | Matches our "unknown yet" gap markers |
| Cortex and Microsoft guidance on owners, audits, and feedback | Feed the living doc content |
| Depending on Runme, Copier, or Atuin directly | Not decided. License and trust model need review first |

## Still to do

- Search GitHub topics and package registries for onboarding checklists and plans as code.
- Read Copier's update flow in detail.
- Test whether named fenced code blocks run in Runme.
- Confirm licenses for anything we might depend on.
