# Agent instructions

Read first: `README.md`, `docs/architecture/README.md`, `docs/stack.md`, and the ADRs in `docs/decisions/`.

## Where decisions go

- An ADR in `docs/decisions/` records a decision the contracts or acceptance tests depend on. Number ADRs by tens so a missed one can slot in between.
- `docs/stack.md` records language, libraries, and tooling. Change a row there and move the old row to Retired. Do not write an ADR for a stack change unless it forces a test change.
- Every choice must keep the tool a single self-contained file. See `docs/decisions/0005-self-contained-install.md`.

## Naming

Name a unit for what it acts on and what it does: noun, then role. A name that needs "and" is two units. In docs, write plain spaced words, for example "module selector". In Go, the parent directory carries the noun and the package carries the role: `internal/module/selector`, package `selector`. No underscores or mixed caps in package names. Follow the linter, with no exceptions. Spell words out or use a common short form such as `config`. Do not coin abbreviations like `cfg`.

## Components

- A component is one package under `internal/`. Rebuilding a component means deleting its folder and regenerating it from its contract, its architecture doc, and the dependency rules in `docs/decisions/0020-binary-structure-and-regenerative-components.md`.
- A rebuild touches only `internal/<component>/`.
- Never edit `contracts/`, `acceptance/`, or the domain model in `internal/model/` as part of a rebuild. Stop and ask a person.
- Never write or change the tests that judge your own rebuild.
- Respect the dependency rules. Do not add a dependency they do not allow.
- Write core components as pure functions over immutable values. Never alter an argument, hold state between calls, or do input or output. Return a new value. Drivers and the entry point do input and output. See `docs/decisions/0015-functional-core-and-imperative-shell.md`.
- State what each component reads and writes, in the plan and outside it. See the side effects and plan access tables in `docs/decisions/0020-binary-structure-and-regenerative-components.md`. Stay inside the declared set.

## Content rules

- Keep client names, URLs, repo names, people, credentials, and contract details out of the repo. See `CONTRIBUTING.md`.
- No model runs during assembly. Output is deterministic.
