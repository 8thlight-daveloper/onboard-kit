# 0020. Binary structure and regenerative components

Status: accepted

[0260](0260-prerequisites-and-action-order.md) extends the dependency order for
prerequisite scheduling and outcome operations. [0270](0270-outcome-definition-hashes.md)
moves the record reconciler to position 2 with a section hasher dependency;
its current contract and architecture view carry that boundary.

Builds on [0015](0015-functional-core-and-imperative-shell.md), [0010](0010-plan-editor.md) and [0030](0030-self-contained-plan-file.md). The contract and test rules are in [0025](0025-contracts-and-tests.md).

## Context

We want to rebuild any component of the tool from scratch when it proves inadequate, and to limit the technical debt a rebuild can add, even when a smaller model does the work. The idea follows Chad Fowler's [Regenerative Software](https://aicoding.leaflet.pub/3majnyfydzs2y): interfaces, behavior, and evaluations are permanent, and implementations are replaceable.

That needs three things:

1. Boundaries that outlive any implementation.
2. Tests that define correct behavior independently of the code, see [0025](0025-contracts-and-tests.md).
3. Rules that keep a rebuild inside its own folder.

We also need to decide how the tool is packaged. The options were one binary, a wrapper that calls separate binaries, and separate libraries.

## Decision

### One binary, components as packages

- One binary, `onboard-kit`, built from `binary/`. Logic lives in component packages. The entry point is thin glue.
- Each component is one package, and a component is the unit of rebuild. A rebuild deletes the folder and regenerates it.
- The domain model holds the data types that several components pass to each other, such as Section and Plan. It changes slowly and is not a component. It is never deleted and rebuilt. Changing it needs an ADR.

Commands, with `assemble` and `stdio` defined in [0010](0010-plan-editor.md):

| Command | Does |
| -- | -- |
| `onboard-kit` | Starts the TUI walkthrough |
| `onboard-kit assemble --plan <file>` | Assembles a document from a plan. Experimental |
| `onboard-kit stdio` | Exposes the plan components over stdin and stdout, for black-box tests. Experimental |
| `onboard-kit lint` | Checks templates: structure and hygiene. Runs in CI on every change to templates |

The deferral of non-interactive generation is replaced by `assemble`, as decided in 0010.

### Permanent and disposable

| Permanent, changed deliberately | Disposable, rebuilt freely |
| -- | -- |
| The domain model | Each component's folder |
| `contracts/` | Unit tests inside a component's folder |
| `acceptance/` |  |
| `architecture.yaml` |  |
| `docs/decisions/` |  |

Documents under `docs/architecture/` and `docs/reference/` change faster than decisions. They describe the current components and formats, and change with them.

`binary/` holds no logic and is rebuilt with whatever it wires.

### Layout

Folders by role, not by language:

| Folder | Holds | Changes |
| -- | -- | -- |
| Component folders, grouped `<noun>/<role>` | One component each, for example `section/selector` | Rebuilt freely |
| Domain model folder | The shared types: Section, Step, Variable, Question, Plan | By ADR only |
| `contracts/` | One folder per component, mirroring the component folders. Holds the interface file, fixtures, contract tests, and a wiring test | A person approves |
| `acceptance/` | Fixtures, language-neutral scenarios, and a replaceable runner | A person approves |
| `binary/` | The thin entry point | Rebuilt with whatever it wires |

The component list, with each one's responsibility, is in [components](../architecture/components.md). The folder names and file names for the current language are in [code-layout](../stack/code-layout.md), so a change of language does not touch this record.

Components are named for what they work on or produce, and what they do. The directory carries the noun and the package carries the role. The naming rule is in `AGENTS.md`.

Types shared by several components live in the domain model folder, and components hold behavior only. Contract folders may import only the model, so a type that several components use has to live there. The whole model folder is protected, and nothing else lives in it.

Each contract also carries a language-neutral description of the component's behavior beside its fixtures. The interface file is the binding for the current language. The interface lives in the contract folder, not in the component, so deleting the component does not delete its specification. The component imports its own contract to implement it.

A wiring test is the only place a contract touches its implementation, through one constructor call. A rebuild that changes the entry point fails there. This makes the permanent folder depend briefly on the disposable one, but only in a test file. The contract rules are in [0025](0025-contracts-and-tests.md).

### Dependency rules

Each component has a place in a dependency order and may depend only on components earlier in it. Two components at the same position may not depend on each other. If one needs the other, they do not belong at the same position. Architecture tests and a dependency check enforce this.

| Order | Component | May depend on |
| -- | -- | -- |
| 0 | `model` | Nothing from the project |
| 1 | `section/loader` | `model` |
| 1 | `section/resolver` | `model` |
| 1 | `section/hasher` | `model` |
| 1 | `section/validator` | `model` |
| 1 | `template/linter` | `model` |
| 1 | `condition/evaluator` | `model` |
| 1 | `secrets` | `model` |
| 1 | `plan/migrator`, deferred | Nothing from the project, and its own frozen copies of old versions |
| 1 | `plan/encoder` | `model` |
| 1 | `record/encoder` | `model` |
| 1 | `record/joiner` | `model` |
| 1 | `record/splitter` | `model` |
| 1 | `record/reconciler` | `model` |
| 1 | `preparation/encoder` | `model` |
| 1 | `preparation/locator` | `model` |
| 1 | `preparation/validator` | `model` |
| 2 | `section/selector` | `condition/evaluator`, `model` |
| 2 | `plan/instantiator` | `section/hasher`, `model` |
| 2 | `plan/refresher` | `model` |
| 2 | `plan/decoder` | `plan/migrator`, `model` |
| 2 | `record/decoder` | `model` |
| 2 | `preparation/decoder` | `model` |
| 2 | `preparation/editor` | `preparation/validator`, `model` |
| 3 | `question/finder` | `section/selector`, `condition/evaluator`, `model` |
| 3 | `document/assembler` | `section/selector`, `section/resolver`, `model` |
| 4 | `firststeps/assembler` | `document/assembler`, `preparation/locator`, `preparation/validator`, `model` |
| 4 | `plan/validator` | `question/finder`, `section/validator`, `secrets`, `model` |
| 5 | `plan/editor` | `question/finder`, `plan/validator`, `model` |
| 5 | `preparation/reviewer` | `firststeps/assembler`, `question/finder`, `preparation/validator`, `model` |
| 6 | `tui` | `question/finder`, `plan/editor`, `plan/validator`, `document/assembler`, `plan/decoder`, `plan/encoder`, `preparation/decoder`, `preparation/encoder`, `preparation/editor`, `preparation/locator`, `preparation/reviewer`, `model` |
| 7 | `preparation/store` | `plan/decoder`, `preparation/decoder`, `preparation/locator`, `model`. A driver: it reads and writes files. `tui` does not import it. The binary passes it to the screen |
| 8 | `binary` | Wires the components together. Loads a plan through the decoder, saves one through the encoder, and builds a new one with the section loader and the plan instantiator |
| n/a | `contracts/*` | `model`, the shared test code in `contracts/harness`, and its own component, in its wiring test only |

[0520](0520-preparation-screen.md) adds the preparation rows above to the `tui` row. The `tui` package still reads and writes no file. The walk command passes it a preparation saver, as it passes the plan saver.

The document assembler uses the section selector, the section resolver, and the model, and never touches templates. A plan assembles into a document on its own, as [0030](0030-self-contained-plan-file.md) requires. The `section/loader` is used only by the entry point and the `lint` command, to start a new plan, run an update, or check templates. Only drivers and the entry point read or write files outside embedded templates, as the side effects table below shows. The `actions` package rule for v2, where only that package may import `os/exec`, uses the same dependency check.

### Architecture tests are rules as data

The dependency rules are enforced by architecture tests whose rules are data, so a change of language replaces an adapter and not the rules.

- **`architecture.yaml` holds the rules.** It lists the components, the dependency order, the `noun/role` naming, and the categories of package that core components may not use: filesystem, network, process, terminal, and clock. It names no language. It is protected like a contract.
- **An extractor reads the code.** A thin adapter reads the source and emits a plain graph of each component and what it imports, inside and outside the project. It is the only language-specific part, along with the mapping from each banned category to that language's packages, which is listed in [dependency check](../stack/dependency-check.md).
- **A checker compares.** It tests the graph against the rules, and reports each violation with the import that proves it.
- **Checks that need no parser** run on the file system: the folder layout, the naming, that every component has a contract folder, and that the tables in this ADR and in [components](../architecture/components.md) match `architecture.yaml`.
- **Start with the language's own tooling** for the extractor, because it resolves imports exactly. A generic extractor, such as one built on tree-sitter, is an option if the language changes. See [dependency check](../stack/dependency-check.md).

### Rate of change

Things change at different speeds, and the speed says what a change takes. This is the idea Stewart Brand called pace layers, and that Fowler's regenerative software draws on. We use "rate of change" in our own docs so that "layer" keeps one meaning.

| Rate | What lives there | What a change takes |
| -- | -- | -- |
| Slowest | ADRs, the plan file format | A new ADR. Saved plans exist in the wild, so changes must stay compatible. |
| Slow | The domain model (`model`) | An ADR, because it is the file format in memory |
| Medium | Contracts and acceptance tests | A person approves |
| Fast | Component implementations | Rebuilt freely inside the component's folder |
| Fastest | Template content (YAML) | Normal authoring, reviewed by a person |

A dependency points from a faster thing to a slower thing, never the other way. A slow thing that depends on a fast thing breaks whenever the fast thing changes. The one exception is its wiring test, where a contract imports its own component. It is confined to a test file.

### Components have one job each

Each component has one job and a failure mode of its own, and its contract is not a restatement of its neighbors'. The current list and each component's responsibility are in [components](../architecture/components.md#responsibilities), so changes to a component do not need an ADR.

Variables stay in the plan and fill at assembly, as [0030](0030-self-contained-plan-file.md) decides. This answers an earlier open question: the section resolver is separate from the document assembler, which calls it.

### Side effects and plan access

Two declarations for every component. Side effects are what it does outside the plan. Plan access is what it does with a plan it was given: it reads fields, or it returns a changed copy that differs only in declared fields. No component alters the plan it was given, because plans are values. This follows from the functional style in [0015](0015-functional-core-and-imperative-shell.md).

Each component's contract states both. Each contract also carries a property test: after any call, every plan field outside the declared set is identical to the input.

The current declarations are in [components](../architecture/components.md#side-effects-and-plan-access). They change as components change, so they live in the architecture doc and not in this ADR.

### Output styles

An output style is a way of assembling a plan into a document. It is not a variant of a template, because a plan can hold content that no template has. The document assembler takes the plan and a chosen style, and one plan can produce several. The style is an option when assembling, with an optional stored default. The MVP supports one style, Markdown. Each later style is another small component behind the same contract.

## Consequences

- A weaker model rebuilding one component sees one folder, one interface, and one fixture suite, so the debt it can add is bounded by the folder.
- `model` is the shared dependency. A mistake there ripples everywhere, so it stays small and changes only by ADR.
- There is more up-front structure than a single tree of code. This is the price of cheap rebuilds, and it stays inside one module and one binary.

## Alternatives considered

- **Wrapper binary calling separate component binaries.** Gives hard process isolation. Costs version skew between binaries, IPC for the walkthrough, error and cancellation protocol design, wire versioning, and harder distribution for consultants. Revisit for the `actions` component in v2 or if third-party renderers appear, since isolation there is a security feature.
- **Separate build modules per component** (a module or crate per component, where the language has them). The compiler would enforce the dependency graph, and a rebuild would be a directory swap. Costs tags per template, workspace plumbing, and dependency bump ripples, for independent releases we do not need. Revisit if the layering check proves leaky.
- **Components as packages without a permanent contracts folder.** Least structure. A rebuild deletes the interface along with the code, and the tests that judge it live in the folder being rebuilt.
- **Network services per component.** Rejected for the same reasons as the wrapper binary, plus ports and auth. 0010 covers the server question for the walkthrough.
