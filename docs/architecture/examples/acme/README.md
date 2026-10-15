# Example: ACME Corp

A worked example of the module and plan formats. ACME Corp is a made-up organization, and every host uses the reserved `.example` domain. Nothing here points at a real system.

A consultant joins ACME's Platform team, an established team, as a developer. The first goals are access to the tools, a local copy of the project, and a passing test run.

| Path | Contents |
| --- | --- |
| `modules/access.yaml` | A module that applies to every engagement |
| `modules/clone-and-run-tests.yaml` | A module selected only for software work |
| `acme.obk.yaml` | The instantiated plan for this engagement |
| `acme.md` | The plan rendered as a document, written by hand |

## What the files show

- **A module is a plan component plus selection fields.** `appliesTo`, `order`, and `placeholders` exist only in modules. Instantiation selects, orders, fills placeholders, and drops them.
- **Placeholders use `$name`.** A placeholder of `type: command` holds an argument list. It expands as a list in `run` and joins with spaces in prose.
- **Steps are plain YAML.** `run` or `open` is the action. `check` holds `run` or `ask`.
- **A step with `run` and no `check`** passes when the command exits with code 0.
- **A step inherits `timing`** from its module. The plan spells it out, so the plan is self-contained.
- **Commands are argument lists.** No shell parses them.
- **`unknown-yet`** marks a value the consultant did not have at generation. The gap table comes from it.
- **Prose fields hold Markdown.** `body` and `notes` are literal blocks, hard-wrapped.

Nothing here runs. The schema is not final, see [#3](https://github.com/8thlight-daveloper/onboard-kit/issues/3).
