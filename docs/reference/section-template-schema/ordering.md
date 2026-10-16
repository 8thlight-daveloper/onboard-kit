# Ordering and dependencies

Part of the [section template schema](README.md).

Sections appear in the document, and are asked about in the walkthrough, in a deterministic order. The order comes from the section's `timing`, then from the dependencies it declares, then from its `id`. There is no numeric position to maintain.

## Timing

`timing` says when the section's work happens. It is one of a closed, ordered set:

| Value | Meaning |
| --- | --- |
| `before-day-one` | Before the consultant starts |
| `week-one` | The first week |
| `first-30-days` | Through the first 30 days |
| `days-30-to-90` | Day 30 to day 90 |
| `anytime` | No fixed time. It is the default when `timing` is omitted |

A step inherits its section's `timing`, and may state its own from the same set.

## Dependencies

`after` is an optional list of template ids. A section lists the sections that must come before it. A listed section that is not active is ignored.

```yaml
id: review-design-system
timing: week-one
after: [access]
```

## How the order is computed

1. Sort the active sections by `timing`, in the order of the table above.
2. Within that, place each section after every section it lists in `after`. Where nothing constrains two sections, the one with the smaller `id` comes first.
3. The result is the same every time for the same plan.

## What the validators check

These are about the relation between sections, so the plan validator reports them. `onboard-kit lint` builds the default plan from the template set and runs the same validator, so the same checks cover templates.

| Problem | Example |
| --- | --- |
| Unknown reference | `after` names an id that is not in the plan |
| Self reference | A section lists itself in `after` |
| Dependency cycle | A is after B, and B is after A. The report names the sections in the cycle |
| Timing conflict | A is after B, but A's timing comes earlier than B's. The report names both sections and their timings |
| Question condition cycle | A question's condition depends on an answer that depends back on it, including across sections |
| Unknown question reference | A condition names a question that no section declares |
| `includeIf` on a section question | An `includeIf` names a question that is not plan-level |

## Question order in the walkthrough

The finder asks the plan-level questions first, then each active section's questions in document order. It never asks a question before the questions its `askIf` condition names. If a condition refers to a later section, the finder moves the dependent question to just after the question it needs. A cycle is reported as a problem and not resolved.
