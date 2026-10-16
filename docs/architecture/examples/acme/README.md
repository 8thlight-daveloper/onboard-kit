# Example: ACME Corp

A worked example of the module and plan formats. ACME Corp is a made-up organization, and every host uses the reserved `.example` domain. Nothing here points at a real system.

A consultant joins ACME's Platform team, an established team, as a developer. The first goals are access to the tools, a local copy of the project, and a passing test run.

| Path | Contents |
| --- | --- |
| `modules/access.yaml` | A module that applies to every engagement |
| `modules/clone-and-run-tests.yaml` | A module for software work |
| `modules/review-design-system.yaml` | A module for design work. It is inactive in this plan |
| `acme.obk.yaml` | The plan for this engagement |
| `acme.md` | The plan assembled into a document, written by hand |

## What the files show

- **A plan holds every module.** `acme.obk.yaml` has all three modules as sections, including the design one. Only the sections whose `appliesTo` matches the answers appear in `acme.md`.
- **A module and a section have the same shape.** A section is a module copied into the plan with one added field, `source`. See [0030](../../../decisions/0030-self-contained-plan-file.md).
- **Placeholders stay in the plan.** Sections keep `$name`, and `values` holds the answers. Values fill placeholders when the document is assembled. A placeholder of `type: command` holds an argument list. It expands as a list in `run` and joins with spaces in prose.
- **Questions come from two places.** Plan-level `questions` hold the axis questions and the document details. Each active section's `placeholders` become questions of their own. Nothing is asked for the inactive design section, so its scope holds no values.
- **Values are scoped.** `values.plan` holds the plan-level answers, and `values.sections.<id>` holds each section's own. `access` and `clone-and-run-tests` both declare `repo_url`, so they have two values. A user interface can ask the two once and apply the answer to both.
- **Types and checks.** `repo_url` is a `string` with a `pattern` and a `patternHint`. A `boolean` or `choice` would accept only its options. `unknown-yet` is always allowed.
- **An optional section.** `review-design-system` declares an `enabled` placeholder, so its on/off state is a question like any other. The two software sections have none and are always on when they apply.
- **A question condition.** In `access`, `source_control_grantor` has an `askIf` that asks it only once `repo_url` has a value. See the schema doc for the tests a condition can use.
- **`required` and `default`.** Placeholders can be required, or carry a default. `repo_dir` has the default `project`, and this plan overrides it. `outputStyle` has a default and no value.
- **`phase: onboarding`.** `account_name` is answered by each newcomer, so the plan holds no value for it and planning does not ask it. The document shows it as `<account_name>`.
- **`unknown-yet`** marks a value the consultant did not have yet. It satisfies a required question. The gap table comes from it.
- **Steps are plain YAML.** `run` or `open` is the action. `check` holds `run` or `ask`. A step with `run` and no `check` passes when the command exits with code 0. Commands are argument lists, and no shell parses them.
- **A step inherits `timing`** from its section, when assembling.
- **`source`** records the module a section came from. The hashes here are placeholders, because the section hasher is not specified yet.
- **Prose fields hold Markdown.** `body` and `notes` are literal blocks, hard-wrapped.
- **Booleans are lowercase `true` and `false`.** Quote anything else that would read as a boolean.

Nothing here runs. The schema is not final, see [#3](https://github.com/8thlight-daveloper/onboard-kit/issues/3).
