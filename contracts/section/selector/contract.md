# Section selector contract

Status: draft by an agent. Not reviewed by a person. Component: `section/selector`. Issue: [#66](https://github.com/8thlight-daveloper/onboard-kit/issues/66).

The section selector finds the active sections of a plan and puts them in reading order. It is a pure function. It reads no file, reads no clock, keeps no state between calls, and does not change its argument. See [0030](../../../docs/decisions/0030-self-contained-plan-file.md), the [domain model](../../../docs/reference/domain-model.md), [optional sections](../../../docs/reference/section-template-schema/optional-sections.md), [ordering](../../../docs/reference/section-template-schema/ordering.md), and [0015](../../../docs/decisions/0015-functional-core-and-imperative-shell.md).

A plan holds every template as a section. A section is active when its `includeIf` holds, or it has none, and its `enabled` answer, or default, is true, or it has no `enabled` variable. Nothing is stored. The selector computes the set from the plan whenever it is asked.

## Interface

`Select(plan) -> list of Section`.

- `plan` is a Plan, as stored. The selector reads `plan.Sections` and `plan.Answers`. It reads nothing else.
- The result is the active sections in reading order. It is empty when no section is active.
- There is no error and no problem list. The selector evaluates the plan as written. The validators report plans that cannot do their job.

## Rules

**Whether a section is active.** Both gates must pass.

1. **`includeIf`.** A section with no `includeIf` passes. Otherwise the selector asks the condition evaluator, `Holds(includeIf, scope, plan.Answers)`, with the section's `id` as the scope, because the condition is written in the section. The answers are passed as stored. A declared default is never added to them. A condition that does not hold, whether the answer is wrong, deferred, or unanswered, puts the section out. The meaning of each test is the evaluator's, see [its contract](../../condition/evaluator/contract.md).
2. **`enabled`.** A section with no variable named `enabled` passes, and any stored answer named `enabled` for it is ignored. A section with that variable passes when its toggle is true, found like this:
   - The answer is `plan.Answers.Sections[id]["enabled"]`. It counts only when it is a Flag. The answer of another section, and a plan answer named `enabled`, never count.
   - A stored answer that is not a Flag counts as no answer. That covers Deferred, empty Text, Text such as `"true"`, and a List.
   - With no counted answer, the variable's `default` is the toggle when it is a Flag. A default of another kind counts as no default.
   - With neither, the toggle is false and the section is out.

A section that fails either gate is out. The two gates do not depend on each other, and neither depends on any other section.

**Reading order.** The order uses `timing`, then `after`, then `id`.

1. **Timing.** A section with no `timing` is `anytime`. Sections go in the order `before-day-one`, `week-one`, `days-8-to-30`, `days-31-to-90`, `anytime`. Every section of one timing comes before every section of a later timing. Nothing else moves a section across that line, so `after` never does.
2. **Inside one timing,** the sections are placed one at a time. The next section is the one with the smallest `id` among those that are free. A section is free when every id in its `after` is either not an active section of this timing, or already placed. These ids are ignored: an id of an inactive section, an unknown id, an id of a section in another timing, and the section's own id. An ignored id never links two other sections. Repeats in `after` mean the same as one.
3. **A cycle.** When sections remain and none is free, the section with the smallest `id` among those remaining is placed next, as if its unmet `after` ids were not listed. Then the rule above goes on. No active section is dropped.
4. **Ids** compare by their bytes, so `a` comes before `a-1`, then `a-b`, then `a1`, then `ab`, then `b`.

The order depends only on the active sections. It does not depend on their order in `plan.Sections`.

**What is returned.** Each returned section equals the plan's section with that `id`, field for field, including `source`, `unknownFields`, and fields that are empty. No default is filled in, so a section with no `timing` comes back with none. The result is a new list. Changing, reordering, or replacing an item of it does not change the plan. The returned sections may share their inner lists and maps with the plan, because no component changes a value after it is made.

## Properties

These hold for every plan the contract tests use.

- **Pure.** The plan is the same after the call as before.
- **No memory.** The same plan gives the same result on a shared selector, whatever calls came before, forward or in reverse order.
- **Same plan, same order.** Rotating or reversing `plan.Sections` does not change the result.
- **Values as held.** Each returned section is equal to the plan's section with the same `id`.
- **Empty plan.** A plan with no sections gives no sections.

## Side effects and plan access

None. It reads section content (`id`, `includeIf`, `timing`, `after`, and the `enabled` variable) and the answers. It reads no other plan field. This matches the row for `section/selector` in [components](../../../docs/architecture/components.md).

## Not covered

- Duplicate section ids in one plan.
- A `timing` that is not one of the five names, and an `id` that is not a lowercase slug.
- A section with more than one variable named `enabled`.
- Whether `includeIf` names only plan variables, whether a question it names exists, and whether `enabled` is a `boolean` variable. The validators report these. The selector passes the condition to the evaluator as written.
- Whether the plan has a timing conflict. The validator reports it.
- Whether a section's variables, steps, or any other content is valid.
- The active questions of a section. That is the question finder.
- Filling in `$name` in a section. That is the section resolver.

## Decisions to know

The docs were silent, or could be read two ways, in these places. A choice marked `hard` changes what a consultant sees or the order they read in. A choice marked `cheap` touches one fixture and nothing outside the component.

1. **`after` orders sections only inside one timing, and timing wins.** A section that lists a section of a later timing, or an earlier one, is not linked to it. Raised by "Sort the active sections by `timing`" followed by "Within that, place each section after every section it lists". The validator reports the conflict. Cases: `after-across-timings`, `after-across-timings-loop`. `hard`
2. **Ties are broken by the free section with the smallest `id`, and ids compare by bytes.** This is not the same as visiting sections in id order and pulling each one's dependencies in first. With `a` after `c` and a free `b`, the order is `b`, `c`, `a`, and not `c`, `a`, `b`. Raised by "Where nothing constrains two sections, the one with the smaller `id` comes first". Cases: `after-moves-later`, `id-order`. `hard`
3. **A stored `enabled` answer that is not a Flag counts as no answer, so the default applies.** That includes Deferred, even though the docs say `deferred` is not accepted for `enabled`, and Text such as `"true"`. A default that is not a Flag counts as no default. Raised by "No answer. The `default` applies" and "`deferred` is not accepted for it". Cases: `enabled-deferred`, `enabled-not-a-flag`. `hard`
4. **A section that names an inactive section in `after` ignores it, and the inactive section does not pass its own `after` on.** Raised by "A listed section that is not active is ignored". In `x` after `y`, `y` after `z`, with `y` out, `x` and `z` are free of each other. Cases: `after-section-left-out`, `after-section-turned-off`. `hard`
5. **A cycle never drops a section.** When none is free, the smallest remaining `id` goes next. So with `a` after `b` and `b` after `a`, and a free `c`, the order is `c`, `a`, `b`. With `a` after `c`, and `b` and `c` after each other, the order is `a`, `b`, `c`. Raised by "A cycle is reported by the validator" with no order given. Cases: `after-cycle-pair`, `after-cycle-plan-order`, `after-cycle-with-dependent`. `hard`
6. **An unknown id, and a section's own id, in `after` are ignored.** Raised by the same list of validator problems. Cases: `after-unknown-id`, `after-itself`. `hard`
7. **`includeIf` is read with the section's id as its scope.** So a bare name in it means a question of that section. A valid `includeIf` has only `plan.` names, so the scope matters only for a plan the validator rejects. Raised by "A bare name is a question in the same section". Case: `include-if-scope`. `cheap`
8. **The selector takes a whole Plan and uses the sections and the answers only.** A plan variable's `default` does not count as an answer for `includeIf`. Raised by "A declared `default` is not stored until a person selects it". Case: `include-if-ignores-defaults`. `cheap`
9. **The result holds the plan's sections as they are, in a new list.** Nothing is filled in or normalized, and the sections are not copied deeply. Raised by the lack of any statement in the docs. Cases: `full-section`, and the tests of the plan and the result list. `cheap`

## Wrong implementations tried

Each row names a plausible wrong implementation and the groups of cases that fail it. How the contract was checked, and with what scores, is recorded in the issue, not here.

| Wrong implementation | Caught by |
| -- | -- |
| `includeIf` ignored | `include-if-holds`, `include-if-unanswered`, `acme` |
| `includeIf` read in the plan scope | `include-if-scope` |
| A deferred or unanswered answer lets `isNot` hold, or a default is used as an answer | `include-if-unanswered`, `include-if-deferred`, `include-if-ignores-defaults` |
| `enabled` ignored | `enabled-answers`, `enabled-defaults`, `include-if-and-enabled` |
| The `enabled` answer read from the plan scope, or from another section | `enabled-answers`, `enabled-per-section` |
| The default ignored, or a missing default means on | `enabled-defaults`, `enabled-per-section` |
| Deferred switches the section off | `enabled-deferred` |
| Text `"true"` counts as a toggle | `enabled-not-a-flag` |
| Plan order kept | `plain-sections`, `timing-order`, `id-order` |
| Timing ignored, reversed, or `anytime` placed first | `timing-order`, `timing-default`, `repeated-timings` |
| An omitted `timing` is not `anytime` | `timing-default` |
| `after` beats timing | `after-across-timings`, `timing-order` |
| `after` ignored | `after-moves-later`, `after-chain`, `after-several` |
| Ids compared without case or without hyphens | `id-order` |
| Sections visited in id order, each pulling its dependencies in first | `after-moves-later`, `after-several`, `after-repeated-entries`, `after-cycle-pair` |
| An inactive, turned off, or unknown id in `after` blocks the section | `after-unknown-id`, `after-section-left-out`, `after-section-turned-off` |
| An inactive section's own `after` passed on | `after-section-left-out`, `after-section-turned-off` |
| A section's own id in `after` blocks it | `after-itself` |
| A cycle drops sections, or places the largest id | `after-cycle-pair`, `after-cycle-plan-order`, `after-cycle-with-dependent` |
| A default timing filled in, or `source` dropped, in the result | `full-section`, `timing-default`, the test of each returned section |
| The plan's own list returned | The test of the result list |
| An argument changed: the order of the plan's sections | The plan-unchanged test, the result-list test |
