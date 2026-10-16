# Ordering and dependencies

Part of the [section template schema](README.md).

Sections appear in the document, and are asked about in the walkthrough, in a deterministic order. The order comes from the section's `timing`, then from the dependencies it declares, then from its `id`. There is no numeric position to maintain.

## Timing

`timing` says when the section's work happens. It is a name from a closed, ordered set, and each name has a fixed window of days. Days count from the start day, and day 1 is the first day of work.

| Name | Window | Meaning |
| -- | -- | -- |
| `before-day-one` | Up to day 0 | Before the consultant starts |
| `week-one` | Days 1 to 7 | The first week |
| `days-8-to-30` | Days 8 to 30 | The rest of the first month |
| `days-31-to-90` | Days 31 to 90 | The second and third months |
| `anytime` | None | No fixed time. It is the default when `timing` is omitted |

- **The window is part of the name.** It never depends on which other sections are in the plan, so a template means the same thing in every plan.
- **Authors pick a name.** A name reads well in the document, and the window gives a validator and the later onboarding helper numbers to work with: the helper can add the window to a start date to get due dates.
- **Windows do not overlap**, apart from `anytime`, which has none.
- **A finer value** can be added later as an optional explicit window, for example days 12 to 14, without breaking existing templates.

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
| -- | -- |
| Unknown reference | `after` names an id that is not in the plan |
| Self reference | A section lists itself in `after` |
| Dependency cycle | A is after B, and B is after A. The report names the sections in the cycle |
| Timing conflict | A is after B, but A's window starts before B's. The report names both sections and their windows. `anytime` never conflicts |
| Question condition cycle | A question's condition depends on an answer that depends back on it, including across sections |
| Unknown question reference | A condition names a question that no section declares |
| `includeIf` on a section question | An `includeIf` names a question that is not a plan variable |

## Question order in the walkthrough

The finder asks the plan variables first, then each active section's questions in document order. It never asks a question before the questions its `askIf` condition names. If a condition refers to a later section, the finder moves the dependent question to just after the question it needs. A cycle is reported as a problem and not resolved.
