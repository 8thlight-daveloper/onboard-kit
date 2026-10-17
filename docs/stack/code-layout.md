# Code layout

Part of the [stack](README.md). The structure is decided in [0020](../decisions/0020-binary-structure-and-regenerative-components.md) and [0025](../decisions/0025-contracts-and-tests.md) in language-neutral terms. This file maps it to the current language, Go. A language change replaces this file and leaves those ADRs in place.

## Mapping

| Neutral term | In Go |
| -- | -- |
| Component | One package under `internal/<noun>/<role>`. The directory carries the noun and the package carries the role |
| Domain model | `internal/model`. Go's `internal/` rule keeps it from being imported outside the Go module |
| Entry point | `binary/`, holding a thin `main` |
| Interface file | `api.go` in the contract package, holding the interface and its data types |
| Contract tests | `<role>_test.go` in the contract package |
| Wiring test | `wire_test.go`, the only file in a contract package that imports its implementation, through one constructor call |
| Fixtures | `testdata/` folders, which the Go tool ignores when it builds |

## Folders

```text
internal/
  model/                     # permanent: domain model (Section, Step, Variable, Question, Plan), imports nothing from the project
  plan/
    instantiator/            # component: plan instantiator
    editor/                  # component: plan editor, see 0010
    validator/               # component: plan validator, see 0010
    migrator/                # component: plan migrator, deferred until the first breaking change, see 0030
    decoder/                 # component: plan decoder, bytes to a Plan
    encoder/                 # component: plan encoder, a Plan to canonical bytes
  condition/
    evaluator/               # component: condition evaluator, askIf and includeIf
  question/
    finder/                  # component: question finder, see 0010
  section/
    selector/                # component: section selector
    resolver/                # component: section resolver
    hasher/                  # component: section hasher
    loader/                  # component: section loader
    validator/               # component: section validator, checks one section's structure
  template/
    linter/                  # component: template linter, checks template text against a rules file
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
  testdata/                  # permanent: template sets, plans, golden documents
  scenarios/                 # permanent: scenario files, language-neutral
  runner/                    # replaceable: reads the scenarios and drives the built binary
binary/                      # thin main, the entry point
```

## Naming and calls

- Package names follow the linter: lowercase, no underscores or mixed caps. See `AGENTS.md`.
- A call uses the package name, so the noun is visible only in the import path: `editor.Submit(plan, id, value)`, `finder.Questions(plan)`. Role names are unique today, so no call is ambiguous. If two roles ever clash, or a local variable shadows one, alias the import, for example `planeditor "…/internal/plan/editor"`.
- Go has no nested namespaces, so `plan.editor.Submit` cannot be written.

## Contracts in Go

- The interface lives in the contract package, not in the component, so deleting a component does not delete its specification. The component imports its own contract package to implement it.
- `wire_test.go` makes the permanent folder depend on the disposable one, but only in a test file. A rebuild that changes the entry point fails to compile there.
- Contract packages may import only the model, so a type that several components use has to live in `internal/model`. The whole folder is protected.
