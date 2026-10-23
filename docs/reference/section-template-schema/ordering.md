# Ordering and dependencies

Part of the [section template schema](README.md).

[0260](../../decisions/0260-prerequisites-and-action-order.md) separates the
grouped reference document from the global action sequence. Requires format
preservation, hashing, structural validation, and action scheduling are implemented;
record-backed walkthrough integration remains pending. The grouped order below
is a reference view, not proof that prerequisites have been verified.

Sections appear in the reference document in a deterministic grouped order.
There is no numeric position to maintain.

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

## Prerequisites and legacy hints

`requires` names real prerequisite activities. A dependent stays active but
waits until each direct prerequisite has a unique stored verified outcome.
Failed, deferred, or absent outcomes are insufficient. An inactive prerequisite
makes its dependents inactive. Parent containment supplies no verification
obligation. Dates are planning goals, never permission or failure gates.

Requires works across parent subtrees. The planned action scheduler orders
prerequisites before dependents, even when their timing goals suggest otherwise.
Among available activities it chooses bootstrap first, then effective timing,
grouped reference rank, and exact ID. See 0260 for closure and cycle rules.

```yaml
id: lead-deploy
requires: [shadow-deploy]
```

`after` is a deprecated optional list of reading-order hints. A listed inactive
section is ignored. Only active siblings in the same effective timing group
constrain the grouped reference order. After does not require verification and
must not be silently converted to Requires. New library content uses Requires
for real prerequisites.

```yaml
id: review-design-system
timing: week-one
after: [access]
```

## Sub-sections

A section may name a parent with `parent: <section id>`. A section with a parent is a sub-section. See [0070](../../decisions/0070-sub-sections.md).

```yaml
id: access-request
parent: access
includeIf:
  all:
    - question: access.self_service
      answer:
        is: true
```

- A sub-section is active only if its parent is active, and then by its own `includeIf` and switch. Turning a parent off turns off its sub-sections. A parent that is not in the plan, or a cycle of parents, makes a section inactive.
- It comes directly after its parent, depth first. Siblings are ordered by timing, then `after`, then id. A sub-section with no `timing` takes its parent's for ordering among its siblings. A sub-section never moves a root across a timing line.
- The assembled document writes it one heading level below its parent.
- The plan validator reports a parent that does not exist, a cycle, and an `includeIf` that names anything but a plan variable or an ancestor's question.

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
| Timing conflict, retired by 0260 | Cross-timing After hints are ignored, not errors. Goals never invalidate a real Requires edge. The plan validator delegates these structural rules |
| Question condition cycle | A question's condition depends on an answer that depends back on it, including across sections |
| Unknown question reference | A condition names a question that no section declares |
| `includeIf` on a section question | An `includeIf` names a question that is not a plan variable |

## Question order in the walkthrough

The finder asks the plan variables first, then each active section's questions in document order. It never asks a question before the questions its `askIf` condition names. If a condition refers to a later section, the finder moves the dependent question to just after the question it needs. A cycle is reported as a problem and not resolved.
