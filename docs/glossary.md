# Glossary

Plain definitions of the terms used in the architecture docs and decisions. Add a term here before using it in a doc.

## The content

| Term         | Meaning                                                                                                                                                                                                                                              |
|--------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Module       | One reusable piece of onboarding content, written once as a YAML file. It says when it applies, which questions it needs answered, and what text it produces.                                                                                        |
| Placeholder  | A named blank in a module, written `$name`, filled from an answer.                                                                                                                                                                                   |
| Answers      | What the consultant says in response to the questions: team state, discipline, placeholder values.                                                                                                                                                   |
| Section      | A module after it has been selected and filled in, as it appears inside a plan.                                                                                                                                                                      |
| Plan         | The finished, standalone description of one organization's onboarding: sections, answers, question definitions, and a record of where it came from. It is saved as `<name>.obk.yaml` and carries everything needed to reopen it without the modules. |
| Output style | The target format of the document, such as Markdown for GitHub.                                                                                                                                                                                      |
| Gap          | A value not known yet. It shows in the document as a visible to-do with an owner and timing.                                                                                                                                                         |
| Provenance   | A record of where a plan came from: which module set version and when it was made.                                                                                                                                                                   |

## The tool

| Term        | Meaning                                                                                                                                                          |
|-------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Tool        | The `onboard-kit` binary, the one file consultants install. Everything below this heading is a part of it.                                                       |
| Walkthrough | The interactive run where the tool asks questions and builds a plan.                                                                                             |
| TUI         | Terminal user interface. The screen-based walkthrough in a terminal.                                                                                             |
| Session     | The tool's memory of one walkthrough: which questions are open, which answers are accepted, what is still missing. See [0010](decisions/0010-session-engine.md). |
| Driver      | Anything that operates a session: the TUI, a headless command, or a test.                                                                                        |
| Headless    | Running without a person or a screen. A command reads input and writes output.                                                                                   |
| Assemble    | The step that turns a plan into a document in an output style. It is deterministic. No model runs.                                                               |

## The structure

| Term               | Meaning                                                                                                                                                                                                                                                                         |
|--------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Component          | One part of the tool with one responsibility: it takes one kind of input and produces one kind of output, such as the module selector. It is a part of the tool, not of a plan. Each component lives in its own folder and can be deleted and rebuilt.                          |
| Module selector    | Component that picks and orders the modules that apply to the answers.                                                                                                                                                                                                          |
| Section resolver   | Component that fills a module's placeholders and produces a section.                                                                                                                                                                                                            |
| Plan instantiator  | Component that combines selection and filling into a plan.                                                                                                                                                                                                                      |
| Document assembler | Component that turns a plan into a document in an output style. It reads only the plan.                                                                                                                                                                                         |
| Domain model       | The definitions of the data that several components pass to each other, such as Module, Section, and Plan. They live in one folder, `model`, and change slowly.                                                                                                                 |
| Boundary           | The edge of a component: what it accepts, what it returns, and what it hides. A component's contract describes its boundary.                                                                                                                                                    |
| Dependency         | One component using another. The document assembler using the domain model is a dependency.                                                                                                                                                                                     |
| Dependency rules   | The rules that limit which dependencies are allowed, so boundaries hold. Components have a place in a dependency order and may depend only on components earlier in it. The session may depend on the module selector, and the module selector may never depend on the session. |
| Dependency check   | An automatic check that fails the build when a component has a dependency the rules do not allow. It works by reading the imports in the code.                                                                                                                                  |
| Rate of change     | How fast something changes, and what a change takes. Also known as pace layers, a term from Stewart Brand. The domain model changes slowly and needs an ADR. A component's implementation changes freely. A dependency points from the faster thing to the slower thing.                                                              |

## Keeping rebuilds safe

| Term                 | Meaning                                                                                                                               |
|----------------------|---------------------------------------------------------------------------------------------------------------------------------------|
| Rebuild (regenerate) | Delete a component's folder and have it written again from its contract.                                                              |
| Contract             | The written description of what one component must do, together with its fixtures and contract tests. It outlives any implementation. |
| Fixture              | A saved example: a specific input and the output the component must produce for it.                                                   |
| Golden               | A fixture whose expected output is a whole saved document, compared byte for byte.                                                    |
| Contract test        | An automatic check that one component, given a fixture's input, produces its expected output.                                         |
| Acceptance test      | An automatic check that the finished tool, run as a user would run it, gives the right result.                                        |
| Unit test            | A test a component's own author writes for its internals. It is deleted with the component.                                           |
| Protected files      | Files a rebuild may not change without a person's approval: contracts, acceptance tests, and the domain model.                        |
| Walking skeleton     | The thinnest end-to-end version of the whole tool, built first, with each component barely working.                                   |

## Records

| Term       | Meaning                                                                                                                                 |
|------------|-----------------------------------------------------------------------------------------------------------------------------------------|
| ADR        | Architecture decision record. A short document that states a decision and the reasons, for decisions the contracts and tests depend on. |
| Stack note | A row in [stack.md](stack.md) recording the current language or tool. Cheap to change.                                                  |
| Eval       | A check where a model judges the result. Not used for assembly.                                                                         |
