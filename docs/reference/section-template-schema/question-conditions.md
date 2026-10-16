# Question conditions

Part of the [section template schema](README.md).

A question condition says "ask this question only if other answers match." It gates one question, which is one placeholder. Sections have their own gates: `appliesTo`, and an `enabled` placeholder, see [optional sections](optional-sections.md).

```yaml
placeholders:
  - name: staging_url
    description: Staging URL
    askIf:
      all:
        - question: plan.discipline
          answer:
            is: software
        - question: plan.teamState
          answer:
            isNot: first-consultants
        - question: repo_url
          answer:
            isAnswered: true
```

A condition is data, not an expression string. Every implementation language reads it the same way, the JSON Schema checks it, and it can grow without breaking old plans.

| Part | Meaning |
| --- | --- |
| `all` | A list of clauses that must all be true. It is the only combinator for now. Add `any` when a real template needs it |
| `question` | The id of the question to test. A bare name is a question in the same section. `plan.<name>` is a plan-level question |
| `answer` | The tests on that question's answer. Several tests in one `answer` must all hold |

Answer tests:

| Test | True when |
| --- | --- |
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

