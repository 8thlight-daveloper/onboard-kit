# 0055. Variable audience

Status: accepted

Builds on [0035](0035-settled-and-deferred-answers.md) and [0045](0045-defaults-are-suggestions.md). It replaces `phase`, described in [variables](../reference/section-template-schema/variables.md) and in [0010](0010-plan-editor.md), and a line of [0030](0030-self-contained-plan-file.md). It works with [0060](0060-personal-plan.md).

## Context

Two kinds of person touch a plan. The onboarder builds it. The onboardee is the person being onboarded. A person who helps an onboardee acts for them, so they count as the onboardee. Approvals in an organization's own systems happen outside the tool.

A variable's `phase` is `planning` or `onboarding`. That one word stands for three separate ideas: who answers, when they answer, and where the answer is kept. It also forces rules that no longer hold.

- A plan variable may not use `audience: onboardee`, "because the plan holds one answer for everyone". That assumes one plan for everyone. With one plan for each onboardee, a plan-wide question such as "which team are you on?" can be asked of the onboardee. Its answer chooses which sections they see.
- 0030 says newcomer answers are outside the plan. They will live in the onboardee's own plan file, see 0060.

## Decision

One org plan serves every role and team in the organization. Variation comes from the onboardee's answers, which choose sections through `includeIf`. It does not come from separate plans. An organization with a very different function, such as legal, may need its own plan, and that is outside what this model stretches to cover.

A variable has an **audience**: who answers it.

| Audience | Who | What it holds | Example |
| -- | -- | -- | -- |
| `onboarder` | The person who builds and maintains the plan. This is the default | Facts the organization knows and every newcomer shares | `repo_url`, `tracker_grantor` |
| `onboardee` | The person being onboarded | Facts the newcomer knows about themselves, or choices that pick what applies to them | `account_name`, `team`, `discipline` |

Three questions decide the audience of a new variable:

1. **Who knows the answer?** The newcomer knows their account, team, and kind of work. The organization knows its URLs, who grants access, and who to contact.
2. **Does it vary per person?** If it does and the person knows it, ask the onboardee.
3. **Should the onboardee be able to change it?** If not, it is an onboarder variable.

The onboarder writes the question and its options, and the onboardee gives the answer. For `discipline` the onboarder lists the kinds of work the organization supports, and each newcomer picks theirs.

Every onboardee question must exist for a reason: some section's `includeIf` or text depends on it. A question nothing uses is not asked.

### What the ACME example holds

| Variable | Audience | Why |
| -- | -- | -- |
| `discipline` | `onboardee` | The kind of work the person will do, software or design. It selects discipline-specific sections |
| `team` | `onboardee` | The group the person joins |
| `account_name` | `onboardee` | The person's own account |
| `organization` | `onboarder` | An organization fact |
| `team_state` | `onboarder` | A fact about the engagement. It stays with the onboarder until per-team answers exist, see below |
| Repositories and commands, grantors, contacts, URLs | `onboarder` | Facts the organization knows |

### Rules

- **Scope and audience are separate.** A variable is declared in the plan or in a section, and either one can have either audience. The rule against onboarding plan variables is removed.
- **`phase` is removed.** What it meant follows from the audience.
- **The finder takes an audience.** It lists the active questions for one audience. The consultant's editor lists onboarder questions. The onboardee's TUI lists onboardee questions, and shows the onboarder's answers without letting the onboardee change them. A question is active by its own condition, whoever answers it, so an onboarder question may depend on an onboardee answer.
- **Blockers are per audience.** A plan is ready to hand over when its required onboarder questions are settled. An onboardee's plan is complete when its required onboardee questions are settled. A required onboardee question does not block the onboarder.
- **An answer already in the file is respected.** An onboarder who knows a person's team can put the answer in before handing the plan over, and the onboardee is not asked again. The onboardee cannot change an answer to an onboarder variable.
- **Conditions are unchanged.** A condition on an unanswered question is false, so sections that depend on an onboardee answer stay out until the answer exists. The evaluator needs no change.
- **A gap names its audience.** The section resolver reports every unanswered or deferred variable the text uses, with its audience in the problem. Whether that is a warning is the caller's decision, because only the caller knows if the plan is an org plan or an onboardee's. This replaces the rule that an onboarding variable with no value is not a gap.

### Ownership below the organization

Some answers are the organization's but differ by team, such as a repository or `team_state`. Today the plan holds one answer per question, so a section per team with an `includeIf` on `team` carries those. The direction after that is a team level of ownership between the organization and the person: a question says it is answered per team, answers are stored per team, and the lookup runs from the person to the team to the organization. A team lead would then own a team's answers. This is not built and not decided in detail. See "Not decided".

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

- Per-team answers: how a question says it is answered per team, where the answers are stored, and who defines the list of teams. The options of the `team` question are the likely source.
- The marker the document shows for a missing value. A proposal is `[deferred: name]`, `[unanswered: name]`, and `[to fill in: name]`, one shape with a reason and the variable's name. The resolver contract settles it.
