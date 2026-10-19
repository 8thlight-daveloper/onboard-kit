# Reference

Exact definitions of formats the tool reads and writes, with a worked example. These change with the format, which is faster than a decision but slower than the code that uses them.

| Doc | Defines |
| -- | -- |
| [domain-model.md](domain-model.md) | The types that components share, stored and computed |
| [section-template-schema/README.md](section-template-schema/README.md) | The shape of a template and a section, variables, question conditions, and reserved names |
| [acceptance-scenarios.md](acceptance-scenarios.md) | The data format of an acceptance scenario, and how a run is judged |
| [examples/acme](examples/acme/README.md) | A worked example: four templates, a plan, and the assembled document |

The plan file's format and versioning rules are decided in [0030](../decisions/0030-self-contained-plan-file.md), and the `stdio` protocol in [0010](../decisions/0010-plan-editor.md).
