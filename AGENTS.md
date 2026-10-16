# Agent instructions

Read first: `README.md`, `docs/architecture/README.md`, `docs/stack/README.md`, and the ADRs in `docs/decisions/`.

## Where docs go

| Folder | Holds | Pace |
| -- | -- | -- |
| `docs/decisions/` | ADRs: what we chose and why | Slowest. A new ADR supersedes an old one |
| `docs/reference/` | Exact definitions of formats, such as the section template schema, with worked examples | Changes with the format |
| `docs/architecture/` | Views of the current system: components, data flow, sequences | Changes as components change |
| `docs/stack/` | Language and tooling, one file per topic, including the code layout | Fastest |

Docs in `architecture/`, `reference/`, and `stack/` are named for their topic, with no numbers. Only ADRs are numbered.

## Where decisions go

- An ADR in `docs/decisions/` records a decision the contracts or acceptance tests depend on. Number ADRs by tens so a missed one can slot in between.
- `docs/stack/README.md` records language, libraries, and tooling. Change a row there and move the old row to Retired. Do not write an ADR for a stack change unless it forces a test change.
- Every choice must keep the tool a single self-contained file. See `docs/decisions/0005-self-contained-install.md`.

## Naming

Name a unit for what it acts on and what it does: noun, then role. A name that needs "and" is two units. In docs, write plain spaced words, for example "section selector". In Go, the parent directory carries the noun and the package carries the role: `internal/section/selector`, package `selector`. No underscores or mixed caps in package names. Follow the linter, with no exceptions. Spell words out or use a common short form such as `config`. Do not coin abbreviations like `cfg`.

## Components

- A component is one package under `internal/`. Rebuilding a component means deleting its folder and regenerating it from its contract, its architecture doc, and the dependency rules in `docs/decisions/0020-binary-structure-and-regenerative-components.md`.
- A rebuild touches only `internal/<component>/`.
- Never edit `contracts/`, `acceptance/`, `architecture.yaml`, or the domain model in `internal/model/` as part of a rebuild. Stop and ask a person.
- Never write or change the tests that judge your own rebuild.
- When you find that a contract conflicts with another document, is incomplete, or looks wrong, stop and report it. Do not edit the contract unless a person authorizes that specific change.
- Contract and test rules are in `docs/decisions/0025-contracts-and-tests.md`.
- When building a component, first run its contract tests against a stub and confirm every test fails. After they pass, run the mutation tool and report surviving mutants as contract gaps. Do not edit the contract.
- When asked to draft a contract, write all four parts, cite the source of each fixture, list plausible wrong implementations and show that the tests fail them, and do not build the component in the same run.
- Write acceptance scenarios as data under `acceptance/scenarios/`. Never write acceptance tests in implementation code or import implementation packages.
- Respect the dependency rules. Do not add a dependency they do not allow.
- Write core components as pure functions over immutable values. Never alter an argument, hold state between calls, or do input or output. Return a new value. Drivers and the entry point do input and output. See `docs/decisions/0015-functional-core-and-imperative-shell.md`.
- State what each component reads and writes, in the plan and outside it. The rule is in `docs/decisions/0020-binary-structure-and-regenerative-components.md`, and the current tables are in `docs/architecture/components.md`. Stay inside the declared set.

## Where files go

Put throwaway notes, drafts, and your own working files in `.scratch/`. Put generated plans and documents in `output/`. Git ignores both, except `output/.gitkeep`. Do not commit anything from either, and do not write them anywhere else in the repo.

## Checks

Run `just check` before you commit, from the Nix dev shell. Do not write ad hoc scripts to check links or formatting, because `just lint-links`, `just fmt-check`, `just lint-md`, and `just lint-prose` already do. Never skip a failing check or edit a tool config to make one pass without asking a person.

## Content rules

- Keep client names, URLs, repo names, people, credentials, and contract details out of the repo. See `CONTRIBUTING.md`.
- No model runs during assembly. Output is deterministic.

## Issues

- Do not create, edit, or close an issue unless a person asks.
- Use the template in `.github/ISSUE_TEMPLATE/` for the kind of work: decision, contract, build, or integration. `gh issue create --template` works only interactively. When not interactive, remove the front matter from the template file, fill in the sections, and pass the result with `--body-file -`. Add `--label`, `--parent`, and `--blocked-by` as needed.
- Link to documents. Do not copy their text into an issue. Done criteria are checks a machine can run.
