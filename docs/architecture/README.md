# Architecture

Proposed architecture for onboard-kit. Nothing here is built yet. These docs record our current best guesses so we can argue with them before writing code.

Diagrams use Mermaid, which GitHub renders in Markdown.

## Principles

- **Deterministic assembly.** The same answers and the same modules always produce the same document. No model runs during assembly.
- **Agents author, they do not assemble.** Agents help write and tune modules. A human reviews every change.
- **The plan is the source.** Answers and content save to one self-contained plan file. The document regenerates from it, in any output style.
- **The plan is the lasting artifact.** The assembled document is transitional. The end state is the plan file and the onboarding helper flow, so nothing long-term may key on assembled content.
- **Functional core, imperative shell.** Components are pure functions over immutable values, and never alter an argument. Only drivers and the entry point do input and output. See [0015](../decisions/0015-functional-core-and-imperative-shell.md).
- **Stands alone.** Installing means getting one file. No runtime, package manager, or second download. See [0005](../decisions/0005-self-contained-install.md).
- **Internal tool, client-facing output.** The repo and the tool stay inside 8th Light. Only generated files go to clients.
- **No client specifics in the repo.** Plan files and output live in `output/`, which git ignores.

## Documents

| Doc | Shows |
| --- | --- |
| [01-context.md](01-context.md) | Who and what touches the system |
| [02-components.md](02-components.md) | The components and the direction of their dependencies |
| [03-data-flow.md](03-data-flow.md) | How modules and answers become a document |
| [04-walkthrough-sequence.md](04-walkthrough-sequence.md) | The steps of one run, including reopening a plan |
| [05-module-schema.md](05-module-schema.md) | The shape of a module and a section |
| [06-authoring-loop.md](06-authoring-loop.md) | How feedback becomes a better module |

## Related

- [Decisions](../decisions/) record choices the contracts and tests depend on.
- [Stack](../stack.md) lists the current language and tooling, and what it replaced.
- [Glossary](../glossary.md) defines the terms used here.

## Decisions

| ADR | Decides | Status |
| --- | --- | --- |
| [0005](../decisions/0005-self-contained-install.md) | One self-contained file, no secondary installs | Accepted |
| [0010](../decisions/0010-plan-editor.md) | Question finder, plan validator, and plan editor as plain functions over a plan, and the `assemble` and `stdio` commands | Proposed |
| [0015](../decisions/0015-functional-core-and-imperative-shell.md) | Pure functions over immutable values in the core, input and output only in the shell | Proposed |
| [0020](../decisions/0020-binary-structure-and-regenerative-components.md) | One binary, rebuildable components, contracts, dependency rules, output styles | Proposed |
| [0030](../decisions/0030-self-contained-plan-file.md) | The plan file holds everything, its format, versioning, and hashes | Proposed |

## Open decisions

| Decision | Issue |
| --- | --- |
| Module schema and placeholder syntax | [#3](https://github.com/8thlight-daveloper/onboard-kit/issues/3) |
| Hosting and distribution | [#4](https://github.com/8thlight-daveloper/onboard-kit/issues/4) |

Open questions across all docs are tracked in [#24](https://github.com/8thlight-daveloper/onboard-kit/issues/24). The walking skeleton, [#25](https://github.com/8thlight-daveloper/onboard-kit/issues/25), waits on it.

## Decided

- Language and toolkit are stack notes, see [stack.md](../stack.md).
- The saved file is the plan, a YAML file named `<name>.obk.yaml` ([#23](https://github.com/8thlight-daveloper/onboard-kit/issues/23)).
- A plan holds every module as a section, and the answers decide which are active.
- Output styles are an option on a plan, not module variants. The MVP has Markdown only ([#8](https://github.com/8thlight-daveloper/onboard-kit/issues/8)).
- Non-interactive assembly exists as `onboard-kit assemble`, and `onboard-kit stdio` exposes the plan components over stdio for tests.
- Hosting is deferred until after the MVP.
