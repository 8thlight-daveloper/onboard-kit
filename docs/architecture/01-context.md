# Context

Status: proposed

Who and what touches onboard-kit. The tool is internal. The client receives only the generated file.

> [!NOTE]
> **What is a module?**
>
> A module is one YAML file holding a single piece of onboarding content, such as the first week plan or the access checklist. It says when to include it, which placeholders it needs, and what steps it contains. Prose fields hold Markdown text. The tool selects modules from the walkthrough answers and merges them into one plan, which renders as a document. Modules are plain data, so people and agents can write and validate them without special tooling. See [05-module-schema.md](05-module-schema.md) and the [ACME example](examples/acme/README.md).

```mermaid
flowchart LR
    consultant["Consultant (8th Light)"]
    kit["onboard-kit TUI"]
    modules[("Modules, YAML files embedded in the binary")]
    config["Config file"]
    doc["Generated onboarding doc"]
    reviewer["Human reviewer"]
    client["Client onboarding owner"]

    consultant -->|answers questions| kit
    modules -->|selected by tags| kit
    kit -->|saves answers| config
    config -->|regenerates output| kit
    kit -->|writes| doc
    doc -->|read before sending| reviewer
    reviewer -->|approved file| client
```

## Notes

- Modules ship inside the binary, so every consultant on the same version gets the same content.
- The config file and the generated doc contain client specifics. Both are written under `output/`, which git ignores.
- A human reviews every document before it leaves 8th Light. The tool reminds the consultant but cannot enforce it.
- The client never runs the tool. They get a file in the format the consultant chose.

## Open questions

- Who reviews, and what do they check? See [#12](https://github.com/8thlight-daveloper/onboard-kit/issues/12).
- If we host a web version, this picture changes. See [#4](https://github.com/8thlight-daveloper/onboard-kit/issues/4).
