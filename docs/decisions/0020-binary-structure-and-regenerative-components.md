# 0020. Binary structure and regenerative components

Status: proposed

Relates to [#24](https://github.com/8thlight-daveloper/onboard-kit/issues/24), [#20](https://github.com/8thlight-daveloper/onboard-kit/issues/20), [#25](https://github.com/8thlight-daveloper/onboard-kit/issues/25). Builds on [0015](0015-functional-core-and-imperative-shell.md), [0010](0010-plan-editor.md) and [0030](0030-self-contained-plan-file.md).

## Context

We want to rebuild any component of the tool from scratch when it proves inadequate, and to limit the technical debt a rebuild can add, even when a smaller model does the work. The idea follows Chad Fowler's [Regenerative Software](https://aicoding.leaflet.pub/3majnyfydzs2y): interfaces, behavior, and evaluations are permanent, and implementations are replaceable.

That needs three things:

1. Boundaries that outlive any implementation.
2. Tests that define correct behavior independently of the code.
3. Rules that keep a rebuild inside its own folder.

We also need to decide how the tool is packaged. The options were one binary, a wrapper that calls separate binaries, and separate libraries.

## Decision

### One binary, components as packages

- One binary, `onboard-kit`, built from `binary/`. Logic lives in `internal/` packages. The `main` is thin glue.
- Each component is one package under `internal/`, and a component is the unit of rebuild. A rebuild deletes the folder and regenerates it.
- `internal/model` holds the domain model: the data types that several components pass to each other, such as Module, Section, and Plan. It changes slowly and is not a component. It is never deleted and rebuilt. Changing it needs an ADR.

Commands, with `assemble` and `stdio` defined in [0010](0010-plan-editor.md):

| Command | Does |
| --- | --- |
| `onboard-kit` | Starts the TUI walkthrough |
| `onboard-kit assemble --plan <file>` | Assembles a document from a plan. Experimental |
| `onboard-kit stdio` | Exposes the plan components over stdin and stdout, for black-box tests. Experimental |
| `onboard-kit lint` | Checks modules against the schema |

The deferral of non-interactive generation is replaced by `assemble`, as decided in 0010.

### Permanent and disposable

| Permanent, changed deliberately | Disposable, rebuilt freely |
| --- | --- |
| `internal/model` | `internal/<component>` |
| `contracts/` | Unit tests inside `internal/<component>` |
| `acceptance/` | |
| `docs/architecture/`, `docs/decisions/` | |

`binary/` holds no logic and is rebuilt with whatever it wires.

### Layout

```
internal/
  model/                     # permanent: domain model (Module, Section, Step, Placeholder, Question, Plan), imports nothing from the project
  plan/
    instantiator/            # component: plan instantiator
    editor/                  # component: plan editor, see 0010
    validator/               # component: plan validator, see 0010
    migrator/                # component: plan migrator, see 0030
  module/
    loader/                  # component: module loader
  question/
    finder/                  # component: question finder, see 0010
  section/
    selector/                # component: section selector
    resolver/                # component: section resolver
    hasher/                  # component: section hasher
  secrets/                   # component: secrets check
  document/
    assembler/               # component: document assembler, plan to document in an output style
  tui/                       # component: terminal driver
contracts/
  <noun>/<role>/             # mirrors internal/, for example contracts/section/selector
    api.go                   # permanent: interface and data types
    testdata/                # permanent: fixtures and expected results
    <role>_test.go           # permanent: contract tests, behavior only
    wire_test.go             # binds the tests to the implementation
acceptance/
  testdata/                  # permanent: module sets, plans, golden documents
  *_test.go                  # permanent: drive the built binary
binary/                      # thin main, the entry point
```

Components are named for what they work on or produce, and what they do. The directory carries the noun, the package carries the role. In prose these are the module loader, section selector, section resolver, section hasher, question finder, plan instantiator, plan validator, plan editor, plan migrator, and document assembler. The naming rule is in `AGENTS.md`.

Types shared by several components live in `internal/model`, and components hold behavior only. Contract packages may import only the model, so a type that several components use has to live there. The whole `internal/model` folder is protected, and nothing else lives in it.

Each contract also carries a language-neutral description of the component's behavior beside its fixtures. The Go interface in `api.go` is the binding for the current language. See [stack.md](../stack.md).

The interface lives in `contracts/.../api.go`, not in the component. Deleting the component then does not delete its spec. The component imports its own contract package to implement it.

`wire_test.go` is the only place a contract package touches its implementation, through one constructor call. A rebuild that changes the entry point fails to compile there. This makes the permanent folder depend briefly on the disposable one, but only in a test file.

### Dependency rules

Each component has a place in a dependency order and may depend only on components earlier in it. Two components at the same position may not depend on each other. If one needs the other, they do not belong at the same position. Architecture tests and a dependency check enforce this.

| Order | Component | May depend on |
| --- | --- | --- |
| 0 | `model` | Nothing from the project |
| 1 | `module/loader` | `model` |
| 1 | `section/selector` | `model` |
| 1 | `section/resolver` | `model` |
| 1 | `section/hasher` | `model` |
| 1 | `secrets` | `model` |
| 1 | `plan/migrator` | `model`, and its own frozen copies of old versions |
| 2 | `question/finder` | `section/selector`, `model` |
| 2 | `plan/instantiator` | `section/hasher`, `model` |
| 2 | `document/assembler` | `section/selector`, `section/resolver`, `model` |
| 3 | `plan/validator` | `question/finder`, `secrets`, `model` |
| 4 | `plan/editor` | `question/finder`, `plan/validator`, `model` |
| 5 | `tui` | `question/finder`, `plan/editor`, `document/assembler`, `model` |
| 6 | `binary` | Wires the components together. Loads a plan through the migrator, and builds a new one with the loader and the plan instantiator |
| n/a | `contracts/*` | `model`, and its own component in `wire_test.go` only |

The document assembler uses the section selector, the section resolver, and the model, and never touches modules. A plan assembles into a document on its own, as [0030](0030-self-contained-plan-file.md) requires. The `module/loader` is used only by the entry point, to start a new plan or run an update. Only drivers and the entry point read or write files outside embedded modules, as the side effects table below shows. The `actions` package rule for v2, where only that package may import `os/exec`, uses the same dependency check.

### Rate of change

Things change at different speeds, and the speed says what a change takes. This is the idea Stewart Brand called pace layers, and that Fowler's regenerative software draws on. We use "rate of change" in our own docs so that "layer" keeps one meaning.

| Rate | What lives there | What a change takes |
| --- | --- | --- |
| Slowest | ADRs, the plan file format | A new ADR. Saved plans exist in the wild, so changes must stay compatible. |
| Slow | The domain model (`model`) | An ADR, because it is the file format in memory |
| Medium | Contracts and acceptance tests | A person approves |
| Fast | Component implementations | Rebuilt freely inside the component's folder |
| Fastest | Module content (YAML) | Normal authoring, reviewed by a person |

A dependency points from a faster thing to a slower thing, never the other way. A slow thing that depends on a fast thing breaks whenever the fast thing changes. The one exception is `wire_test.go`, where a contract imports its own component. It is confined to a test file.

### Assembly components

Each component has one job and a failure mode of its own, and its contract is not a restatement of its neighbors'.

| Component | Does | Does not |
| --- | --- | --- |
| Module loader | Reads module files and decodes them strictly, enforcing the format rules in 0030 | Select or fill |
| Section selector | Finds the active sections of a plan: `appliesTo` matching against the axis values, `any`, enabled state, sort by `order`, tie-break on `id` | Fill values, read modules |
| Section resolver | Fills the placeholders in one section from the plan's values: `$name`, `$$`, command lists, `unknown-yet` gaps, undeclared or missing names | Choose sections, collect gaps across sections |
| Section hasher | Computes the canonical content hash of a section, for its `source` | Anything else |
| Plan instantiator | Builds the default plan from the loaded modules and the built-in axis questions: copies every module into a section, records `source` and provenance with an injected clock | Select, fill placeholders, ask questions, validate input |
| Question finder | Finds a plan's active questions with their current values or defaults, and the required ones with no value: the plan-level questions plus the placeholders of active sections, joined by name | Change a plan, check values |
| Plan validator | Checks whether values are acceptable: type, options, the secrets check. Checks a whole plan, or one value | Decide what is asked, change a plan |
| Plan editor | Returns a copy of a plan with one answer applied, using the finder and the validator | Decide what is asked, know about modules, or how a plan is built or filled |
| Plan migrator | Brings a plan of an older schema version to the current one, one version at a time | Anything else |
| Document assembler | Turns a plan into a document in an output style. Uses the selector for active sections and the resolver to fill values, and builds the gap table with owner and timing | Read modules, ask questions |
| Secrets check | Decides whether a value looks like a token | Anything else |

Placeholders stay in the plan and fill at assembly, as [0030](0030-self-contained-plan-file.md) decides. This answers the open question in [#3](https://github.com/8thlight-daveloper/onboard-kit/issues/3): the section resolver is separate from the document assembler, which calls it.

### Side effects and plan access

Two declarations for every component. Side effects are what it does outside the plan. Plan access is what it does with a plan it was given: it reads fields, or it returns a changed copy that differs only in declared fields. No component alters the plan it was given, because plans are values. This follows from the functional style in [0015](0015-functional-core-and-imperative-shell.md).

Each component's contract states both. Each contract also carries a property test: after any call, every plan field outside the declared set is identical to the input.

The current declarations are in [02-components](../architecture/02-components.md#side-effects-and-plan-access). They change as components change, so they live in the architecture doc and not in this ADR.

### Output styles

An output style is a way of assembling a plan into a document. It is not a variant of a module, because a plan can hold content that no module has. The document assembler takes the plan and a chosen style, and one plan can produce several. The style is an option when assembling, with an optional stored default. The MVP supports one style, Markdown. Each later style is another small component behind the same contract. This answers [#8](https://github.com/8thlight-daveloper/onboard-kit/issues/8).

### Two kinds of tests, and what they judge

| Kind | Level | Lives in | Judges |
| --- | --- | --- | --- |
| Contract test | One component boundary: fixture in, expected out | `contracts/<noun>/<role>/` | A component and any rebuild of it |
| Acceptance test | Whole binary: plan or scripted answers in, document out | `acceptance/` | The assembled system |

We do not call these evals. In this repo, "eval" is reserved for checks that a model judges, which may appear in the authoring loop. Assembly has no model in it, so its checks are ordinary deterministic tests.

Rules for both:

- **Behavior only.** Assert on outputs. Never on internals or package structure.
- **Independent authorship.** The agent that rebuilds a component does not write or edit that component's contract tests, fixtures, or acceptance tests. A person approves every change to them.
- **Properties as well as goldens.** Two runs produce identical bytes. Output holds no secrets and no client specifics. Every placeholder is resolved or reported as a gap.
- **Seed fixture.** `docs/architecture/examples/acme` becomes the first acceptance golden.
- **Parallel by default.** Tests run in parallel with race detection on, use a per-test temporary directory for files, and take an injected clock. No shared mutable state at package or module level. Any unordered iteration that reaches output is sorted. Golden regeneration is a separate explicit command and never runs as a side effect of a test. The commands for the current language are in [stack.md](../stack.md).

### Rebuilding a component

1. Open a branch. Delete `internal/<component>/`.
2. Regenerate it from the contract package, the architecture doc for that component, and the dependency rules.
3. Run contract tests and acceptance tests.
4. CI checks that the change touches only `internal/<component>/`.

Any change under `contracts/`, `acceptance/`, or `internal/model` is a separate change with human approval. A rebuild that needs a contract change stops and asks.

### Review

A person reviews a rebuild by reading the contract and acceptance results, plus any change to tests, not the regenerated code. The review passes if every test passes and no test changed without approval. Reading the code stays optional.

CI enforces:

- The folder-scope rule for rebuild changes.
- Human approval on changes to `contracts/`, `acceptance/`, and `internal/model`.
- Dependency check and architecture tests, run in `just pre-commit`.
- The full test suite with race detection on.

The dependency check tool is a stack choice, listed in [stack.md](../stack.md).

## Consequences

- A weaker model rebuilding one component sees one folder, one interface, and one fixture suite, so the debt it can add is bounded by the folder.
- The quality of the contracts decides the quality of the rebuilds. Thin contracts let a rebuild drift without failing a test. Writing and reviewing contracts becomes the main design work.
- `model` is the shared dependency. A mistake there ripples everywhere, so it stays small and changes only by ADR.
- There is more up-front structure than a single package tree. This is the price of cheap rebuilds, and it stays inside one module and one binary.
- Reviewers rely on the test results being trustworthy. Independent authorship of tests is what makes that safe, so the approval rule is not optional.

## Alternatives considered

- **Wrapper binary calling separate component binaries.** Gives hard process isolation. Costs version skew between binaries, IPC for the walkthrough, error and cancellation protocol design, wire versioning, and harder distribution for consultants. Revisit for the `actions` component in v2 or if third-party renderers appear, since isolation there is a security feature.
- **Separate build modules per component** (a module or crate per component, where the language has them). The compiler would enforce the dependency graph, and a rebuild would be a directory swap. Costs tags per module, `replace` or `go.work` plumbing, and dependency bump ripples, for independent releases we do not need. Revisit if the layering check proves leaky.
- **Components as packages without a permanent contracts folder.** Least structure. A rebuild deletes the interface along with the code, and the tests that judge it live in the folder being rebuilt.
- **Network services per component.** Rejected for the same reasons as the wrapper binary, plus ports and auth. 0010 covers the server question for the walkthrough.

## Open questions

- Who writes the first contracts for components that do not exist yet? The walking skeleton ([#25](https://github.com/8thlight-daveloper/onboard-kit/issues/25)) is the natural moment. Proposal: a person writes or approves each one, an agent may draft them, and the run that builds a component cannot edit them.
