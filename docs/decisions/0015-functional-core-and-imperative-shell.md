# 0015. Functional core and imperative shell

Status: accepted

Underlies [0010](0010-plan-editor.md) and [0020](0020-binary-structure-and-regenerative-components.md).

## Context

Components in this tool are built to be deleted and regenerated from a contract ([0020](0020-binary-structure-and-regenerative-components.md)). Whoever regenerates one, a person or an agent, needs only that component's folder, its contract, and the dependency rules. For that to work, everything a component can do must be visible at its boundary. A component that reads a hidden clock, keeps state between calls, or reaches into a shared object has behavior the contract cannot state and the tests cannot judge.

## Decision

The tool is a functional core inside an imperative shell. This is how every component is written, and it is a rule, not a preference.

### Why

- **Regeneration.** A pure function is fully described by its inputs and outputs. A contract that lists inputs and expected outputs can therefore specify the whole component, and a rebuild that passes it is as correct as the contract is strong. The contract checks in 0020 test that strength. Hidden state would put behavior outside the contract.
- **Discrete pieces.** A function that alters nothing and holds nothing can be replaced, reordered, or tested without disturbing any other component. One component cannot affect another except through the values it returns.
- **Clean responsibilities.** Each component answers one question about a plan. Deciding what is asked, what is acceptable, and how an answer applies are separate functions with separate reasons to change. Doing input and output is a different responsibility, and only the shell has it.
- **Small context for an agent.** To change a component, an agent needs its contract, its inputs, and its declared outputs. It does not need to trace shared state, call order, or side effects elsewhere. The context that makes a change safe is small and fixed.
- **Testing.** Pure functions need no mocks, clocks, or setup. A contract fixture is plain data: inputs and expected outputs. Tests run in parallel and can fork from a shared plan. Checks that arguments are unchanged and that repeated calls match are cheap, and they let tests judge a rebuild without trusting it.
- **Visible boundaries.** What a component may read, write, and import is declared and checked: side effects and plan access in [0020](0020-binary-structure-and-regenerative-components.md#side-effects-and-plan-access), imports by the dependency rules. A component that breaks the style fails a test or a dependency check instead of quietly widening its reach.

### Rules

**The core** is every component except the drivers and the entry point. Core components follow these rules:

- **Values are immutable.** No function alters an argument. A function that needs a different value returns a new one.
- **Functions are pure.** The same inputs give the same output. A function reads nothing but its arguments.
- **No hidden state.** No package-level mutable state, and no state held between calls. State the core needs is a parameter, and it comes back as part of the result.
- **No clock, no randomness, no environment.** When a function needs the time, the caller passes it in.
- **No input or output.** The core never reads or writes files, the network, or the terminal, and never runs commands.
- **Expected failures are values.** A problem, a blocker, or an unacceptable value is a returned result. A panic means a bug.

**The shell** is the drivers and the entry point. It does everything the core must not:

- Reads and writes files, and reads the embedded templates.
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
- **Declared plan access.** After any call, every plan field outside the declared set is identical to the input. See [components](../architecture/components.md#side-effects-and-plan-access).
- **No input or output in the core.** The dependency check forbids core components from importing the filesystem, network, process, or terminal packages. The exact list belongs in [dependency check](../stack/dependency-check.md).
- **No shared mutable state.** A lint forbids package-level variables that can change.

## Consequences

- A contract can specify a whole component, so a component can be deleted and regenerated against it.
- An agent changing a component reads its folder, its contract, and the dependency rules, and nothing else.
- Tests run in parallel with no setup, because nothing is shared.
- A test can fork from a shared plan and try many next answers, and the review screen can go back by keeping an earlier plan.
- A rebuilt component cannot affect another component through hidden state, so a rebuild stays inside its folder.
- Deterministic assembly, our first principle, holds by construction and is checked by test.
- Copies cost memory and time. A plan holds tens of questions and a few dozen sections, so this does not matter.
- The shell stays thin and easy to read, and the stdio adapter and the TUI differ only in how they ask and show.

## Alternatives considered

- **A mutable core with discipline.** Cheaper to write. It brings back undo logic, locks, and tests that cannot fork, and nothing stops a rebuilt component from reaching through a shared pointer. Rejected.
- **Objects that hold their own state.** A session object holding the plan plus methods. Once the plan carries all the state, the object adds a name and an identity and no behavior, and its hidden state sits outside any contract. See [0010](0010-plan-editor.md).
- **A library of immutable data structures.** Real immutability, at the cost of a dependency and an unfamiliar API. Not needed at this size. Revisit if plans grow large or copying shows up as a cost.

## Open questions

- Which standard library packages does the dependency check forbid in core components? A stack decision.
- Does the property test snapshot cover unexported fields, or only the serialized form? The serialized form is the contract, and it is enough to start.
