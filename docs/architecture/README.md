# Architecture

Proposed architecture for onboard-kit. Nothing here is built yet. These docs record our current best guesses so we can argue with them before writing code.

Diagrams use Mermaid, which GitHub renders in Markdown.

## Principles

- **Deterministic assembly.** The same answers and the same modules always produce the same document. No model runs during assembly.
- **Agents author, they do not assemble.** Agents help write and tune modules. A human reviews every change.
- **Config is the pivot.** Walkthrough answers save to a config file. Any output style regenerates from it.
- **Internal tool, client-facing output.** The repo and the tool stay inside 8th Light. Only generated files go to clients.
- **No client specifics in the repo.** Config files and output live in `output/`, which git ignores.

## Documents

| Doc | Shows |
| --- | --- |
| [01-context.md](01-context.md) | Who and what touches the system |
| [02-components.md](02-components.md) | Packages and the direction of their dependencies |
| [03-data-flow.md](03-data-flow.md) | How modules and answers become a document |
| [04-walkthrough-sequence.md](04-walkthrough-sequence.md) | The steps of one run, including regeneration |
| [05-module-schema.md](05-module-schema.md) | The shape of a module |
| [06-authoring-loop.md](06-authoring-loop.md) | How feedback becomes a better module |

## Open decisions

| Decision | Issue |
| --- | --- |
| Module schema and placeholder syntax | [#3](https://github.com/8thlight-daveloper/onboard-kit/issues/3) |
| Hosting | [#4](https://github.com/8thlight-daveloper/onboard-kit/issues/4) |
| Output styles | [#8](https://github.com/8thlight-daveloper/onboard-kit/issues/8) |
| Config file format | [#23](https://github.com/8thlight-daveloper/onboard-kit/issues/23) |

Language and TUI library are settled in [#2](https://github.com/8thlight-daveloper/onboard-kit/issues/2): Go with Bubble Tea and Huh.
