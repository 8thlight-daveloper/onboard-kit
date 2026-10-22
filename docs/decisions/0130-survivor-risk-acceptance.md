# 0130. Accept the remaining survivor review risk

Status: accepted for this batch

Makes a scoped exception to the human-only survivor acceptance rules in
[0025](0025-contracts-and-tests.md), [0100](0100-behavior-judgment.md), and
`AGENTS.md`.

## Context

The remaining mutation review is delaying work on the application. The user
explicitly delegated acceptance of the remaining survivors to the behavior
judge, asked to record agent rather than human approval, and chose to address
defects when they appear in use instead of finishing this review first.

## Decision

The judge accepts the current batch as deferred testing risk. These entries
stop blocking the hardening commits and application work. This is not a finding
that every mutation is equivalent, harmless, or outside the supported inputs.
Missing probes, unspecified behavior, and known test gaps remain such.

The delegation is human-authorized. Each survivor's acceptance is agent-only;
there is no individual human approval. Keep that distinction in the accepted
lists. Do not apply this exception automatically to future survivors.

The inventory is 224 survivors across 12 components, against source at
`69e65b8` with the pending approved hardening fixtures:

| Component | Agent-accepted risks | Identity evidence |
| -- | -- | -- |
| Section resolver | 45 | Available full gomu report |
| Plan validator | 29 | Available full gomu report |
| Question finder | 37 | Available full gomu report |
| Plan editor | 20 | Available full gomu report |
| Record joiner | 7 | Historical inventory |
| Record splitter | 2 | Historical inventory |
| Record reconciler | 21 | Historical inventory |
| Secrets | 5 | Historical inventory |
| Condition evaluator | 11 | Historical inventory |
| Section hasher | 22 | Historical inventory |
| Section selector | 12 | Historical inventory |
| Plan instantiator | 13 | Historical inventory |

The exact membership and mutation descriptions are recorded in
`tools/mutation/accepted/<component>.txt`. Preserve existing human approvals.
The 131 identities derived from available full reports become active accepted
keys. The 93 historical identities remain comments: their risks are accepted,
but they must not suppress a different mutation after the ordinal-key fix.
Matching those identities is deferred, not a prerequisite for moving on.

## Boundaries

No mutation is applied to production code by accepting it. Existing tests and
checks remain mandatory. Keep already approved hardening fixtures. Do not
weaken tests, change product behavior to match a mutant, or hide execution
errors and timeouts as accepted survivors.

The default-acceptance rule in [0120](0120-defaults-require-acceptance.md) and
the ambiguity rule in [0110](0110-reject-ambiguous-plans.md) remain product
requirements. Their implementation work is separate from this review waiver.
Independent test and implementation authorship and other protected-artifact
rules remain unchanged.

## Consequences

Ship application improvements with known limits in mutation coverage. A later
defect needs a regression test and a fix, and any related exception must be
reconsidered. An adjusted mutation score excludes accepted risks; it must not
be presented as evidence that the tests now detect them.

Working reviews and raw reports remain in ignored `output/` and `.scratch/`.
This decision and the accepted-list entries carry the durable scope and
approval provenance. No fresh full mutation run is claimed for this decision.
