# Conditions

Part of the [section template schema](README.md).

Two fields hold a condition, and both use the syntax below.

| Field | On | Gates |
| -- | -- | -- |
| `askIf` | A variable | Whether that one question is asked |
| `includeIf` | A section or template | Whether the whole section is in the plan's document and walkthrough |

A section has a second gate, its `enabled` switch, see [optional sections](optional-sections.md). `includeIf` is for facts about the engagement, such as the discipline. The switch is for a choice the consultant makes, and a condition can test it with a section clause.

## Syntax

```yaml
variables:
  - name: staging_url
    description: Staging URL
    askIf:
      all:
        - question: plan.discipline
          answer:
            is: software
        - question: plan.team_state
          answer:
            isNot: first-consultants
        - question: repo_url
          answer:
            isSettled: true
```

A condition is data, not an expression string. Every implementation language reads it the same way, the JSON Schema checks it, and it can grow without breaking old plans.

| Part | Meaning |
| -- | -- |
| `all` | A list of clauses that must all be true. It is the only combinator for now. Add `any` when a real template needs it |
| `question` | The id of the question to test. A bare name is a question in the same section. `plan.<name>` is a plan variable |
| `answer` | The tests on that question's answer. Several tests in one `answer` must all hold |

### A section is the unit of "only if"

There are no step-level conditions. Every step in a section applies together. A step that applies only for some people goes in its own small section, with an `includeIf` on a plan variable or a switch the onboarder can turn off. Text that differs inside a step is a variable the organization fills in. See [0065](../../decisions/0065-messages-and-evidence.md).

### Section clause

A clause can test another section's [switch](optional-sections.md) instead of an answer.

```yaml
askIf:
  all:
    - section: design-review
      enabled: true
```

- `section` is the id of another section in the plan. `enabled` is `true` or `false`.
- The clause holds when that section's switch equals `enabled`. A section whose `enabled` is absent is on.
- It tests the switch only. It does not ask whether that section's `includeIf` holds, so a chain of sections cannot make a cycle through the switches.
- A clause has `section` or `question`, never both.
- The plan validator reports a `section` that names no section in the plan.

Answer tests:

| Test | True when |
| -- | -- |
| `is` | The answer is this one value |
| `isNot` | The answer is a settled answer other than this one |
| `isOneOf` | The answer is one of these values, a list of two or more |
| `isNotOneOf` | The answer is a settled answer that is none of these, a list of two or more |
| `isSettled` | `true`: the answer is settled. `false`: the answer is deferred or unanswered |

Rules:

- **More combinators later.** `any` and `none` are not supported yet. Add them together when a real template needs one, with a rule that a condition on an unanswered question is never true, so `none` cannot ask early.
- **Settled answer.** The question is active and the plan stores an answer for it that is not `deferred` and not empty. The operator `isSettled` tests for it. A boolean `false` is a settled answer. A declared `default` is not stored until a person selects it, so it does not count.
- **Unanswered means false.** Every test except `isSettled` is false when the answer is deferred or unanswered, before any negation. So `isNot` on an unanswered question is false, and the follow-up is not asked early.
- **When false.** The question is not active. It keeps its stored answer, is not asked, and does not block.
- **Values fit the question.** Tested values must satisfy the type and options of the question they test.
- **Validator problems.** The section validator reports a reference to an unknown question in the same section. The plan validator reports one across sections, and a cycle of conditions anywhere in the plan. See [ordering](ordering.md#what-the-validators-check).
- **Lint errors.** A condition that cannot do its job:
  - `isOneOf` or `isNotOneOf` with a single value. Use `is` or `isNot`.
  - A list test on a `boolean`. Use `is`, or `isSettled` for both values.
  - A list test that names every option of a `choice`. Use `isSettled`.
  - Tests that can never hold together, such as `is: a` with `isNot: a`.
- **Lint warnings.** A condition that does nothing but is not wrong. It never blocks.
  - A condition with no clauses, such as `all: []`. It always holds, the same as no condition, so a section or question it was meant to narrow is not narrowed.

## Section conditions

`includeIf` has the same syntax as `askIf`.

```yaml
id: review-design-system
includeIf:
  all:
    - question: plan.discipline
      answer:
        is: design
```

- **No `includeIf` means always included.** A core template has none. A variant has one that narrows by an axis or any other plan-level answer.
- **Unanswered means out.** The same rule as `askIf`: every test except `isSettled` is false until the question has a settled answer, so a conditional section stays out of the plan until its question is answered.
- **Plan variables only.** An `includeIf` may name only plan variables, written `plan.<name>`. Those are always asked, so whether a section is in never depends on another section's answers, and there are no cycles between sections' inclusion.
- **Variants.** A variant is another template with its own `includeIf`. It never overrides or extends a core template. Where content differs by axis, write two templates with opposite conditions, such as `discipline` `isNot` `design` and a design twin with `is` `design`, and put shared parts in a template of their own, placed with `after`. See [ordering](ordering.md).
