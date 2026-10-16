# 0025. Contracts and tests

Status: accepted

Relates to [#24](https://github.com/8thlight-daveloper/onboard-kit/issues/24), [#20](https://github.com/8thlight-daveloper/onboard-kit/issues/20). Builds on [0015](0015-functional-core-and-imperative-shell.md) and [0020](0020-binary-structure-and-regenerative-components.md).

## Context

A component can be deleted and rebuilt only if something permanent says what correct means, and that something does not depend on the component's code. [0020](0020-binary-structure-and-regenerative-components.md) sets the structure: permanent contracts and acceptance tests beside disposable components. This record says what those tests are, who writes them, how we know they are strong enough, and how a rebuild is run and reviewed.

The folder layout and file names that carry this out in the current language are in [code-layout](../stack/code-layout.md). The tools are in [testing](../stack/testing.md).

## Decision

### Two kinds of tests, and what they judge

| Kind | Level | Lives in | Judges |
| --- | --- | --- | --- |
| Contract test | One component boundary: fixture in, expected out | `contracts/<noun>/<role>/` | A component and any rebuild of it |
| Acceptance test | Whole binary: a plan or a scripted exchange in, observable results out | `acceptance/` | The assembled system |

We do not call these evals. In this repo, "eval" is reserved for checks that a model judges, which may appear in the authoring loop. Assembly has no model in it, so its checks are ordinary deterministic tests.

Rules for both:

- **Behavior only.** Assert on outputs. Never on internals or code structure.
- **Independent authorship.** The agent that rebuilds a component does not write or edit that component's contract tests, fixtures, or acceptance tests. A person approves every change to them.
- **Properties as well as goldens.** Two runs produce identical bytes. Output holds no secrets and no client specifics. Every placeholder is resolved or reported as a gap.
- **Seed fixture.** [`docs/reference/examples/acme`](../reference/examples/acme/README.md) becomes the first acceptance fixture. Its plan, `acme.obk.yaml`, is the lasting part. Its assembled document, `acme.md`, is a transitional golden.
- **Parallel by default.** Tests run in parallel, use a per-test temporary directory for files, and take an injected clock. No shared mutable state. Any unordered iteration that reaches output is sorted. Golden regeneration is a separate explicit command and never runs as a side effect of a test. The settings for the current language are in [testing](../stack/testing.md).

### Assembled documents are transitional

The intended end state is the plan file and the onboarding helper flow. The assembled document goes away. So nothing long-term may key on assembled content.

- **Lasting contracts** are plan-level: the finder, validator, editor, selector, resolver, hasher, migrator, and instantiator, and acceptance tests that assert on plans and on `stdio` responses.
- **Transitional contracts** are the document assembler and any golden-document acceptance test. Keep them thin, mark them transitional in their contract, and delete them when the assembler goes.
- **Gap table, review marker, and output styles** exist only in the assembled document. They are not plan features, and nothing else depends on them.
- The helper that replaces the document reads the plan through the same components, the finder and the selector, so the lasting contracts already cover what it needs.

### Acceptance tests are language-neutral

Acceptance tests must keep working whatever language the tool is written in, and must not tie the implementation to the language of the tests.

- **Scenarios are data.** Each scenario is a file under `acceptance/scenarios/`. It names the files it starts with, the arguments and the lines it sends, and what it expects back: responses, exit code, and any files written. Plan fixtures are plain plan files.
- **The runner is replaceable.** A small program reads the scenarios, starts the built binary, and compares results. It is the only language-specific part, it holds no knowledge of the tool, and the scenario format is its whole contract. We can replace it, or run a second one in another language, without touching a scenario.
- **Tests never import the implementation.** They see the binary only through its arguments, standard input and output, exit code, and files. This is already true of the `stdio` protocol and the `assemble` command.
- **The scenario format is protected,** like the scenarios, because the runner and every scenario depend on it. Matching rules, such as comparing JSON by subset and ignoring key order, belong to that format.

Contract tests are different on purpose. They call a component in process and are written in the implementation language. Their fixtures are plain data, so only the harness and the interface binding change if the language does.

### Writing contracts

A contract has four parts: a plain description of the behavior, fixtures, contract tests, and the interface binding.

- **An agent drafts, a person approves.** A separate agent run writes the first version of every part, so nobody starts from a blank page. A person reads all of it and gives feedback before anything is built against it. The run that drafts a contract does not build the component.
- **Fixtures cite their source.** Each fixture names where it comes from, such as the ACME example, the schema reference, or an ADR, so a reviewer can check it against something trusted.
- **The implementing agent reads the contract and does not edit it.** If it finds a conflict, a gap, or a mistake, it stops and reports it, in its result and in the pull request. It edits a contract only when a person authorizes that specific change.
- **A contract must tell right from wrong.** One agent writes the spec and the tests together, so a shared misunderstanding could pass review. Two checks guard against it. Both are steps of the workflow, not optional requests.
  - **Drafting.** Before a person reviews, the drafting run lists a few plausible wrong implementations per component, such as ignoring `any`, breaking tie order, or mutating an input. It shows that the contract tests fail each one, and adds fixtures where one slips through. These wrong implementations are scratch work and are not committed.
  - **Building.** The implementing agent works red, then green. Before writing code it runs the contract tests against a stub, and every test must fail. After the tests pass, it runs an automated mutation tool and reports surviving mutants as contract gaps, in its result and in the pull request. The tool generates the mutants, not the agent, so the agent does not grade itself. The agent does not edit the contract.
- **Transitional contracts stay thin.** The document assembler's contract is written last and kept small.

### Rebuilding a component

1. Open a branch. Delete the component's folder.
2. Regenerate it from its contract, the architecture doc for that component, and the dependency rules in [0020](0020-binary-structure-and-regenerative-components.md).
3. Run contract tests and acceptance tests.
4. CI checks that the change touches only that component's folder.

Any change to the contracts, the acceptance tests, `architecture.yaml`, or the domain model is a separate change with human approval. A rebuild that needs a contract change stops and asks.

### Review

A person reviews a rebuild by reading the contract and acceptance results, plus any change to tests, not the regenerated code. The review passes if every test passes and no test changed without approval. Reading the code stays optional.

CI enforces:

- The folder-scope rule for rebuild changes.
- Human approval on changes to contracts, acceptance tests, `architecture.yaml`, and the domain model.
- Dependency check and architecture tests, run before every commit.
- The full test suite.
- `onboard-kit lint` on every change to templates.

The dependency check tool is a stack choice, listed in [dependency check](../stack/dependency-check.md).

## Consequences

- A weaker model rebuilding one component sees one folder, one interface, and one fixture suite, so the debt it can add is bounded by the folder.
- The quality of the contracts decides the quality of the rebuilds. Thin contracts let a rebuild drift without failing a test. Writing and reviewing contracts becomes the main design work.
- Reviewers rely on the test results being trustworthy. Independent authorship of tests is what makes that safe, so the approval rule is not optional.

## Alternatives considered

- **One agent writes the contract and the component.** Fastest. A shared misunderstanding passes review, because the tests were written to fit the code.

## Open questions

- Which mutation tool do we use? A stack choice, see [#33](https://github.com/8thlight-daveloper/onboard-kit/issues/33).
- Do we keep both contract-strength checks after the first builds? Decide after seeing what they catch.
