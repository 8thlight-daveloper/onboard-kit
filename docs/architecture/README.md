# Architecture

Architecture and planned increments for onboard-kit. The component views and
contracts distinguish delivered behavior from work that remains. The
[tool overview](../../README.md#commands) describes current commands.

Diagrams use Mermaid, which GitHub renders in Markdown.

## Principles

- **Deterministic assembly.** The same answers and the same templates always produce the same document. No model runs during assembly.
- **Agents author, they do not assemble.** Agents help write and tune templates. A human reviews every change.
- **Owned inputs are the source.** Shared content and organizational answers stay in one self-contained plan. Person-specific preparation and the onboarding record have separate owners under [0140](../decisions/0140-person-specific-preparation.md). Documents regenerate from those inputs; they are not the source of progress.
- **The plan is the lasting artifact.** The assembled document is transitional. The end state is the plan file and the onboarding helper flow, so nothing long-term may key on assembled content.
- **Functional core, imperative shell.** Components are pure functions over immutable values, and never alter an argument. Only drivers and the entry point do input and output. See [0015](../decisions/0015-functional-core-and-imperative-shell.md).
- **Stands alone.** Installing means getting one file. No runtime, package manager, or second download. See [0005](../decisions/0005-self-contained-install.md).
- **Generated files travel.** The onboarder and the onboardee run the tool. Anyone else, such as the organization's onboarding owner, receives generated files and never needs the tool or the repo.
- **No organization specifics in the repo.** Plan files and output live in `output/`, which git ignores.

## Documents

| Doc | Shows |
| -- | -- |
| [context.md](context.md) | Who and what touches the system |
| [components.md](components.md) | The components and the direction of their dependencies |
| [data-flow.md](data-flow.md) | How templates and answers become a document |
| [walkthrough-sequence.md](walkthrough-sequence.md) | The steps of one run, including reopening a plan |
| [authoring-loop.md](authoring-loop.md) | How feedback becomes a better template |
| [invitation-flow.md](invitation-flow.md) | Planned preparation, invitation, and embedded-library increments, with ownership and verification boundaries |

## Related

- [Decisions](../decisions/) record choices the contracts and tests depend on.
- [Reference](../reference/section-template-schema/README.md) defines the template and section format, with a [worked example](../reference/examples/acme/README.md).
- [Stack](../stack/README.md) lists the current language and tooling, and what it replaced.
- [Glossary](../glossary.md) defines the terms used here.

## Decisions

| ADR | Decides | Status |
| -- | -- | -- |
| [0005](../decisions/0005-self-contained-install.md) | One self-contained file, no secondary installs | Accepted |
| [0010](../decisions/0010-plan-editor.md) | Question finder, plan validator, and plan editor as plain functions over a plan, and the `assemble` and `stdio` commands | Accepted |
| [0015](../decisions/0015-functional-core-and-imperative-shell.md) | Pure functions over immutable values in the core, input and output only in the shell | Accepted |
| [0020](../decisions/0020-binary-structure-and-regenerative-components.md) | One binary, rebuildable components, dependency rules, rate of change, output styles | Accepted |
| [0025](../decisions/0025-contracts-and-tests.md) | Contracts, contract and acceptance tests, who writes them, and how a rebuild is run and reviewed | Accepted |
| [0030](../decisions/0030-self-contained-plan-file.md) | The plan file holds everything, its format, versioning, and hashes | Accepted |
| [0035](../decisions/0035-settled-and-deferred-answers.md) | An answer is unanswered, deferred, or settled, and the names of those states | Accepted |

## Open decisions

These decisions are open:

- Hosting and distribution.
- The dependency check extractor and the banned packages.
- Update action conflict handling, deferred past the MVP.
- Whether to keep the contract-strength checks after the first builds.

The [invitation flow](invitation-flow.md) orders the remaining delivery work.

## Decided

- Language and toolkit are stack notes, see [stack](../stack/README.md).
- The saved file is the plan, a YAML file named `<name>.obk.yaml`.
- A plan holds every template as a section, and the answers decide which are active. Any section can be switched off with its `enabled` setting, which is not a question.
- Conditions are data: `askIf` on a variable gates a question, and `includeIf` on a section gates the section. A variant is a template with its own `includeIf` and never overrides a core template. See [conditions](../reference/section-template-schema/conditions.md).
- Sections sort by `timing`, then `after`, then `id`. See [ordering](../reference/section-template-schema/ordering.md).
- Output styles are an option on a plan, not template variants. The MVP has Markdown only.
- Non-interactive assembly exists as `onboard-kit assemble`, and `onboard-kit stdio` exposes the plan components over stdio for tests.
- Hosting is deferred until after the MVP.
