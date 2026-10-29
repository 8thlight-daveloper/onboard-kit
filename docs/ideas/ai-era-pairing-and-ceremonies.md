# Pairing and team practices when a tool writes code

Status: idea, 2026-10-29. The entries it produced are in
[0550](../decisions/0550-developer-onboarding-catalog-additions.md).

## Idea

Classic pairing teaches by watching someone type. When a tool writes part of the
code, typing is not what a newcomer needs to learn. What remains is specifying
the work, judging what came back, explaining it, and knowing the team's rules.
Each section that teaches one of these should have an observable outcome and a
named teammate, because the teammate is how the team's culture is passed on.

## Principles

- **Specify.** Turn a ticket into acceptance criteria and limits that an agent can
  act on, then compare the result with the intent.
- **Judge.** Run and read the generated change and its tests, and show what was
  checked.
- **Explain.** Explain the change to a coworker before it goes to review, so
  "I cannot explain it" does not reach the reviewer.
- **Know the rules.** Which tools are approved, what code and data may go into
  them, how secrets are handled, and whether a change discloses AI use. This
  must come first, and the other sections require it.
- **Learn from failure.** A teammate shows one time the tool was wrong. Sharing
  a failure builds calibrated trust.
- **Keep the tool neutral.** No vendor names in built-in content. The onboarder's
  answers carry the specifics.

## Ideas considered

- **Navigator and driver pairing** uses the existing `ticket-pairing` entry twice
  with distinct ids. No new entry.
- **Verify a generated change** as its own entry was merged into the pre-review
  walkthrough, because running and reading the change is part of explaining it.
- **Read the team's instructions for coding agents and propose one correction.**
  The same feedback loop as reporting an incorrect onboarding step, aimed at the
  team's agent setup.
- **Mob on ticket writing and on prioritization.** The newcomer observes first,
  then contributes. They overlap general meetings, but they have an outcome of
  taking part. Whether they are in every new plan depends on whether the team
  holds such sessions at all, so they are an opt-in group.
- **Standup and retrospective** are easy first wins. A standup entry did not
  exist in the catalog. The retrospective entry did.

## Open questions

- Whether `ai-tool-policy` stays in the default core for organizations that do
  not allow these tools. Today it is in, because it also states a ban.
- How a team sees whether the pairing and review sections happened, beyond a
  recorded outcome.
