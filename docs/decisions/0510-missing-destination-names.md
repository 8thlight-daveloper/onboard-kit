# 0510. Missing destinations still need a native nonaliasing proof

Status: agent-selected bounded clarification, not human approval or a
mutation-survivor exception.

## Context

The [session editor](0460-record-backed-session.md) allows an unrelated
registration to remain when its root or destination is genuinely absent. That
allowance preserves an unrelated person's state; it does not establish that two
future destinations are distinct.

Two registrations can have different exact account identities, such as Jules
and jules, while their destinations alias on the actual native volume. Removing
their shared empty account directory makes both paths absent but does not turn
the spelling difference into a valid nonaliasing proof.

## Decision

Keep exact plan and account identity comparisons unchanged. Native filename
equivalence never merges identities, selects a winner or authorizes pruning.

When absent tails share an inspected native parent, require supported native
name-comparison evidence before treating differing components as disjoint.
Refuse ambiguity or an unsupported comparison before displaying identity or
reading input. Preserve genuine disjoint missing branches under the existing
allowance. Native capability inspection must not create test objects in the
product flow or invent facts from process-wide filesystem assumptions.

This applies the existing [nonambiguity](0110-reject-ambiguous-plans.md), privacy
and [preserved-ownership principles](0100-behavior-judgment.md); it does not create
a new account-case policy. Genuinely distinct inspected parent identities do not
need spelling comparison to become distinct. An unsupported comparison remains
a native capability boundary, not permission to guess.

## Acceptance boundary

Independent fixtures establish actual same-object case aliases before removing
only test-owned empty directories. They require a preidentity refusal with no
input, clock, disclosure or native changes. Reversed spelling order must also
refuse. A positive control establishes genuinely distinct account directories
under one parent, removes both empty directories and requires an actual selected
record save while preserving every unrelated registration and absent path.

These checks do not establish arbitrary Unicode collation, every filesystem,
future mounts, hostile-owner safety or crash recovery. The
[driver contract](../../contracts/session/editor/contract.md) carries the rule;
the [command reference](../reference/session-edit.md) records the current native
profile boundary. Exact release verification and a fresh common native corpus
must follow a repair before delivery.
