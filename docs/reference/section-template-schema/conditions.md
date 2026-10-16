# Conditions

Part of the [section template schema](README.md).

Two fields hold a condition, and both use the syntax below.

| Field | On | Gates |
| -- | -- | -- |
| `askIf` | A variable | Whether that one question is asked |
| `includeIf` | A section or template | Whether the whole section is in the plan's document and walkthrough |

A section has a second gate, an `enabled` variable, see [optional sections](optional-sections.md). `includeIf` is for facts about the engagement, such as the discipline. `enabled` is for a choice the consultant makes.

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
            isAnswered: true
```

A condition is data, not an expression string. Every implementation language reads it the same way, the JSON Schema checks it, and it can grow without breaking old plans.

| Part | Meaning |
| -- | -- |
| `all` | A list of clauses that must all be true. It is the only combinator for now. Add `any` when a real template needs it |
| `question` | The id of the question to test. A bare name is a question in the same section. `plan.<name>` is a plan variable |
| `answer` | The tests on that question's answer. Several tests in one `answer` must all hold |

Answer tests:

| Test | True when |
| -- | -- |
| `is` | The answer is this one value |
| `isNot` | The answer is a real value other than this one |
| `isOneOf` | The answer is one of these values, a list of two or more |
| `isNotOneOf` | The answer is a real value that is none of these, a list of two or more |
| `isAnswered` | `true`: the question has a real value. `false`: it has none |

Rules:

- **More combinators later.** `any` and `none` are not supported yet. Add them together when a real template needs one, with a rule that a condition on an unanswered question is never true, so `none` cannot ask early.
- **Real value.** The question is active and the plan stores a value for it that is not `unknown-yet` and not empty. A boolean `false` is a real value. A declared `default` is not stored until a person selects it, so it does not count.
- **Unanswered means false.** Every test except `isAnswered` is false when the question has no real value, before any negation. So `isNot` on an unanswered question is false, and the follow-up is not asked early.
- **When false.** The question is not active. It keeps its stored value, is not asked, and does not block.
- **Values fit the question.** Tested values must satisfy the type and options of the question they test.
- **Validator problems.** The section validator reports a reference to an unknown question in the same section. The plan validator reports one across sections, and a cycle of conditions anywhere in the plan. See [ordering](ordering.md#what-the-validators-check).
- **Lint errors.** A condition that cannot do its job:
  - `isOneOf` or `isNotOneOf` with a single value. Use `is` or `isNot`.
  - A list test on a `boolean`. Use `is`, or `isAnswered` for both values.
  - A list test that names every option of a `choice`. Use `isAnswered`.
  - Tests that can never hold together, such as `is: a` with `isNot: a`.

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
- **Unanswered means out.** The same rule as `askIf`: every test except `isAnswered` is false until the question has a real value, so a conditional section stays out of the plan until its question is answered.
- **Plan variables only.** An `includeIf` may name only plan variables, written `plan.<name>`. Those are always asked, so whether a section is in never depends on another section's answers, and there are no cycles between sections' inclusion.
- **Variants.** A variant is another template with its own `includeIf`. It never overrides or extends a core template. Where content differs by axis, write two templates with opposite conditions, such as `discipline` `isNot` `design` and a design twin with `is` `design`, and put shared parts in a template of their own, placed with `after`. See [ordering](ordering.md).
