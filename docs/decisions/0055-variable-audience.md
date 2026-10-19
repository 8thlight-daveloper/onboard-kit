# 0055. Variable audience

Status: proposed

Builds on [0035](0035-settled-and-deferred-answers.md) and [0045](0045-defaults-are-suggestions.md). It replaces `phase`, described in [variables](../reference/section-template-schema/variables.md) and in [0010](0010-plan-editor.md), and a line of [0030](0030-self-contained-plan-file.md). It works with [0060](0060-personal-plan.md).

## Context

Two kinds of person touch a plan. The onboarder builds it. The onboardee is the person being onboarded. A person who helps an onboardee acts for them, so they count as the onboardee. Approvals in an organization's own systems happen outside the tool.

A variable's `phase` is `planning` or `onboarding`. That one word stands for three separate ideas: who answers, when they answer, and where the answer is kept. It also forces rules that no longer hold.

- A plan variable may not use `phase: onboarding`, "because the plan holds one answer for everyone". That assumes one plan for everyone. With one plan for each onboardee, a plan-wide question such as "which team are you on?" can be asked of the onboardee. Its answer chooses which sections they see.
- 0030 says newcomer answers are outside the plan. They will live in the onboardee's own plan file, see 0060.

## Decision

A variable has an **audience**: who answers it.

| Audience | Who | Example |
| -- | -- | -- |
| `onboarder` | The person who builds the plan. This is the default | `discipline`, `repo_url`, `tracker_grantor` |
| `onboardee` | The person being onboarded | `account_name`, `team` |

- **Scope and audience are separate.** A variable is declared in the plan or in a section, and either one can have either audience. The rule against onboarding plan variables is removed.
- **`phase` is removed.** What it meant follows from the audience.
- **The finder takes an audience.** It lists the questions for one audience. The consultant's editor lists onboarder questions. The onboardee's TUI lists onboardee questions, and shows the onboarder's answers without letting the onboardee change them.
- **Blockers are per audience.** A plan is ready to hand over when its required onboarder questions are settled. An onboardee's plan is complete when its required onboardee questions are settled. A required onboardee question does not block the onboarder.
- **Conditions are unchanged.** A condition on an unanswered question is false, so sections that depend on an onboardee answer stay out until the answer exists. The evaluator needs no change.
- **A gap names its audience.** The section resolver reports every unanswered or deferred variable the text uses, with its audience in the problem. Whether that is a warning is the caller's decision, because only the caller knows if the plan is an org plan or an onboardee's. This replaces the rule that an onboarding variable with no value is not a gap.

## Why

- Audience is the real question. Timing and storage follow from it.
- One variable record serves both people and one set of rules, so the multi-team case needs nothing special.
- The resolver stays a pure function. It does not need to know which kind of plan it is looking at.

## Consequences

- **The domain model changes.** `phase` becomes `audience`, with `onboarder` the default. This needs `slow-change`. As in 0035, nothing exists outside this repository, so the schema version stays 1.
- **Contracts**, each its own small increment: the plan decoder and encoder, the section loader, the section hasher (the audience is part of a variable), the section resolver (gaps and markers), and the question finder and plan editor when they are drafted.
- **Docs.** `variables.md`, the schema README, the glossary, 0010, 0030, the architecture docs, and the ACME example take the new field.
- **ADR 0030.** The line "newcomer answers are outside the plan" is superseded by this ADR and 0060.

## Not decided

- The marker the document shows for a missing value. A proposal is `[deferred: name]`, `[unanswered: name]`, and `[to fill in: name]`, one shape with a reason and the variable's name. The resolver contract settles it.
