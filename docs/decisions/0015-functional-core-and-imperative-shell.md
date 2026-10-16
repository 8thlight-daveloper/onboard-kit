# 0015. Functional core and imperative shell

Status: proposed

Relates to [#24](https://github.com/8thlight-daveloper/onboard-kit/issues/24). Underlies [0010](0010-plan-editor.md) and [0020](0020-binary-structure-and-regenerative-components.md).

## Context

The other ADRs already assume this style. Plans are values in 0010, components declare their side effects in 0020, and tests run in parallel. Nowhere is the style stated as a rule. A component rebuilt from its contract has to know what it may not do, and the contract tests have to check it.

## Decision

The tool is a functional core inside an imperative shell.

**The core** is every component except the drivers and the entry point. Core components follow these rules:

- **Values are immutable.** No function alters an argument. A function that needs a different value returns a new one.
- **Functions are pure.** The same inputs give the same output. A function reads nothing but its arguments.
- **No hidden state.** No package-level mutable state, and no state held between calls. State the core needs is a parameter, and it comes back as part of the result.
- **No clock, no randomness, no environment.** When a function needs the time, the caller passes it in.
- **No input or output.** The core never reads or writes files, the network, or the terminal, and never runs commands.
- **Expected failures are values.** A problem, a blocker, or an unacceptable value is a returned result. A panic means a bug.

**The shell** is the drivers and the entry point. It does everything the core must not:

- Reads and writes files, and reads the embedded modules.
- Reads the clock and the environment, and passes what the core needs.
- Asks the person, shows results, and parses and prints the stdio protocol.
- Holds the current plan, and replaces it with the plan the core returns.
- Writes the plan file and the document, only after a confirmed action or an explicit command.

The shell contains no rules about questions or content. If a decision depends on the content of a plan, it belongs in a core component.

### Sharing

A returned value may share unchanged parts with its input. Nothing is ever mutated, so sharing is safe. Where a language cannot guard against mutation, the rule is enforced by test, below.

### How it is enforced

- **Inputs unchanged.** Every contract test takes a deep snapshot of its inputs before the call and compares afterward. A function that mutates an argument fails.
- **Deterministic.** Every contract test runs each call twice and compares the results byte for byte.
- **Declared plan access.** After any call, every plan field outside the declared set is identical to the input. See [02-components](../architecture/02-components.md#side-effects-and-plan-access).
- **No input or output in the core.** The dependency check forbids core components from importing the filesystem, network, process, or terminal packages. The exact list belongs in [stack.md](../stack.md).
- **No shared mutable state.** A lint forbids package-level variables that can change.

## Consequences

- Tests run in parallel with no setup, because nothing is shared.
- A test can fork from a shared plan and try many next answers, and the review screen can go back by keeping an earlier plan.
- A rebuilt component cannot affect another component through hidden state, so a rebuild stays inside its folder.
- Deterministic assembly, our first principle, holds by construction and is checked by test.
- Copies cost memory and time. A plan holds tens of questions and a few dozen sections, so this does not matter.
- The shell stays thin and easy to read, and the stdio adapter and the TUI differ only in how they ask and show.

## Alternatives considered

- **A mutable core with discipline.** Cheaper to write. It brings back undo logic, locks, and tests that cannot fork, and nothing stops a rebuilt component from reaching through a shared pointer. Rejected.
- **Objects that hold their own state.** The earlier session design. Once the plan carried all the state, the object added a name and an identity and no behavior. See [0010](0010-plan-editor.md).
- **A library of immutable data structures.** Real immutability, at the cost of a dependency and an unfamiliar API. Not needed at this size. Revisit if plans grow large or copying shows up as a cost.

## Open questions

- Which standard library packages does the dependency check forbid in core components? A stack decision.
- Does the property test snapshot cover unexported fields, or only the serialized form? The serialized form is the contract, and it is enough to start.
