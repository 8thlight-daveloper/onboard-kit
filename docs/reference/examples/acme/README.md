# Example: ACME Corp

A worked example of the template and plan formats. ACME Corp is a made-up organization, and every host uses the reserved `.example` domain. Nothing here points at a real system.

A consultant joins ACME's Platform team, an established team, as a developer. The first goals are access to the tools, a local copy of the project, and a passing test run.

| Path | Contents |
| -- | -- |
| `templates/access.yaml` | A template that applies to every engagement |
| `templates/clone-and-run-tests.yaml` | A template for software work |
| `templates/review-design-system.yaml` | A template for design work. It is inactive in this plan |
| `templates/stuck-general.yaml` | A template for every engagement, with a required `contact` variable |
| `acme.obk.yaml` | The plan for this engagement |
| `templates/plan-variables.yaml` | The plan variables that `new` adds to every plan |
| `acme.md` | The plan assembled into a document. `onboard-kit assemble --plan acme.obk.yaml` writes it, and an acceptance scenario checks it |

## What the files show

- **A plan holds every template.** `acme.obk.yaml` has all four templates as sections, including the design one. Only the sections whose `includeIf` condition holds, or that have none, appear in `acme.md`.
- **A template and a section have the same shape.** A section is a template copied into the plan with one added field, `source`. See [0030](../../../decisions/0030-self-contained-plan-file.md).
- **Variables stay in the plan.** Sections keep `$name`, and `answers` holds the answers. Answers fill variables when the document is assembled. A variable of `type: command` holds an argument list. It expands as a list in `run` and joins with spaces in prose.
- **Variables come from two places.** The plan's `variables` hold the axis variables, the document details, and a setting. Each active section's `variables` hold the rest. Each variable becomes a question. Nothing is asked for the inactive design section, so its scope holds no values.
- **Answers are scoped.** `answers.plan` holds the answers to plan variables, and `answers.sections.<id>` holds each section's own. `access` and `clone-and-run-tests` both declare `repo_url`, so they have two values. A user interface can ask the two once and apply the answer to both.
- **Types and checks.** `repo_url` is a `string` with a `pattern` and a `patternHint`. A `boolean` or `choice` would accept only its options. `deferred` is always allowed.
- **An optional section.** Any section can be switched off. The plan section holds an optional `enabled`, true by default, and it is not a question. No section in this plan is switched off. See [0050](../../../decisions/0050-section-switch.md).
- **A directory.** The plan holds a `directory` of four contacts, three roles, and two teams. Answers to `contact`, `role`, and `team` questions are ids, and the document shows the names and Slack handles. `source_control_grantor` and the other grantors are `role` questions, `general_contacts` is a `contact` question with `multiple: true`, and `team` is a `team` question. The Platform team names its own contact, and the Design team names its own holder for the Engineering manager role. The addresses use the reserved `.example` domain. See [0080](../../../decisions/0080-directory.md) and [escalation](../../section-template-schema/escalation.md).
- **The plan is in canonical form.** The tool writes plan files the same way every time: a fixed key order, block lists, and quoted dates and reserved words. This file is the expected output of the plan encoder's contract. See [the encoder contract](../../../../contracts/plan/encoder/contract.md).
- **A question condition.** In `access`, `source_control_grantor` has an `askIf` that asks it only once `repo_url` has an answer. See the schema doc for the tests a condition can use.
- **`required` and `default`.** Variables can be required, or carry a default. `repo_dir` has the default `project`, and this plan overrides it. `onboarding_owner_role` suggests `Engineering manager`, and this plan accepts the suggestion as its answer.
- **`audience: onboardee`.** `account_name` is answered by each newcomer, so the plan holds no answer for it and planning does not ask it. The document shows it as `<account_name>`.
- **`deferred`** marks an answer the consultant chose to give later. It satisfies a required question. The gap table comes from it.
- **Steps are plain YAML.** `run` or `open` is the action. `checks` is a list, and each entry is a `run`, an `ask`, or an `env`. A step with `run` and no `checks` passes when the command exits with code 0. Commands are argument lists, and no shell parses them.
- **Order comes from `timing`.** `access` is `before-day-one` and the other two are `week-one`, so they sort in that order, with ties broken by `id`. No numeric position is stored. See [ordering](../../section-template-schema/ordering.md).
- **A step inherits `timing`** from its section, when assembling.
- **`source`** records the template a section came from. The hashes here are variables, because the section hasher is not specified yet.
- **Prose fields hold Markdown.** `body` and `notes` are literal blocks, hard-wrapped.
- **Booleans are lowercase `true` and `false`.** Quote anything else that would read as a boolean.

Nothing here runs. The schema is not final, see [#3](https://github.com/8thlight-daveloper/onboard-kit/issues/3).
