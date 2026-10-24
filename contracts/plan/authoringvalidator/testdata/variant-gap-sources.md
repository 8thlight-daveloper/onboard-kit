# Independent authoring variant gaps

The contract-hardening skill supplies the independent fixture workflow. These
literal plan cases judge public Validate results through the existing
authoringvalidator_test harness. The new Go file imports only testing; the
existing harness supplies the model binding and compares exact diagnostics,
metadata, local repair messages and detail maps. No implementation body, wrong
variant body or implementation-produced expected output was read or generated.
Only variant IDs and descriptions were used to identify candidate gaps.

The JSON model values and expectations are independently hand-authored from
contracts/plan/authoringvalidator/contract.md, its API, the model types and the
secrets contract. Numeric Value kinds follow the domain model: None 0, Text 1,
Flag 2, List 3 and Deferred 4. Every section has required title and summary so
unrelated prose errors do not mask the intended boundary. Cases returning no
diagnostics are controls against overvalidation, rather than standalone proof
that a zero-output validator is correct.

## Sources per case

| Cases | Normative rule and literal oracle |
| -- | -- |
| environment-digit-start-is-invalid | Step Env keys match `[A-Z][A-Z0-9_]*`; a digit cannot start one. Indexed rank avoids echoing its key. |
| multiple-team-still-takes-text-default; multiple-team-list-default-has-independent-shape-error | Only contact/role permit Multiple. Team suggestions always use Text; List is independently the wrong default shape. Path sorting puts default before multiple. |
| active-text-stored-answer-is-not-scanned | This validator ignores stored Answers. The held secret is active Text, so a scanner that wrongly traverses it would produce an observable extra finding. |
| escaped-dollar-alone-is-literal | `$$` is consumed as an escape before considering a following local name. |
| single-letter-dollar-alone-is-literal; two-letter-dollar-alone-is-literal | Lowercase runs shorter than three are literal, with no other invalid dollar in the same field. |
| digit-dollar-alone-is-literal | A maximal digit-started run is literal. |
| trailing-dollar-alone-is-malformed | A lone final single dollar is malformed. |
| role-variable-is-valid-owner | Owner placeholders may name contact or role declarations; no actual directory value or answer is required. |
| qualified-reference-first-dot-keeps-invalid-scope-independent | The reference splits into scope foo and name bar.baz, regardless of a declared invalid scope foo.bar. Independent invalid section ID and unknown question errors both remain. Invalid section metadata is empty. |
| qualified-plan-reference-first-dot-survives-unrelated-duplicate-sections | plan.target.more means plan scope with undeclared name target.more. Duplicate section IDs suppress inter-section judgments, but explicitly do not suppress plan-only references. |
| condition-reference-retains-surrounding-space | Reference lookup is exact; surrounding spaces are not trimmed into a declared plan question. |
| unknown-target-type-suppresses-predicate-membership-and-list-checks | A declared unknown target type has its own variable-type error and suppresses dependent predicate type/value/list judgments. |
| negative-list-excludes-required-scalar | An Is requirement excluded by a valid two-element IsNotOneOf produces one condition contradiction. |
| individual-exclusions-exhaust-choice-domain | Three individually valid IsNot tests exclude every declared distinct choice option. No all-options list predicate masks the contradiction. |
| text-comparison-ignores-inactive-flag; flag-comparison-ignores-inactive-text | Scalar equality compares exact active Kind/payload only. Equal active values with different inactive fields still contradict Is versus IsNot. |
| disabled-section-still-has-question-cycle | Every declaration participates independent of Enabled. Each participant gets one cycle diagnostic at its variable path. |
| descendant-inherits-ancestor-include-dependency | leaf.bravo inherits middle's legal IncludeIf dependency on top.alpha; top.alpha depends on leaf.bravo. Both are cycle participants despite middle declaring no question. |
| duplicate-section-identity-suppresses-all-cycle-analysis | Duplicate section IDs globally suppress cycle analysis, including a local cycle in another uniquely identified section. Existing-validator duplicate codes are never emitted here. |
| duplicate-source-declarations-have-no-unused-judgment | Duplicate source variable names contribute no graph node or unused check. |
| self-only-cycle-does-not-make-onboardee-useful | A self-only AskIf loop has both condition-cycle and unused-onboardee errors at the same path. Code order puts cycle before unused. |
| negative-subset-list-and-scalar-exhaust-choice-domain | A valid negative two-value list excludes a proper subset of three choice options; a separate scalar exclusion removes the last. No positive comparison is present. The finite valid-choice domain is exhausted, so one contradiction belongs to the condition. |
| disabled-own-include-dependency-closes-question-cycle | parent.alpha depends on child.bravo through AskIf; child.bravo depends on parent.alpha through its own section's legal strict-ancestor IncludeIf. The child is explicitly disabled, but all declaration eligibility edges still participate. Each variable gets one cycle error. |

Negative-list domain exhaustion without a positive candidate set and an inactive
section's own IncludeIf edge are distinct from inactive AskIf edges or active
inherited IncludeIf edges. Their expectations follow the finite contradiction
and activity-independent graph rules; no implementation or overlay body supplied
their expected results.

## Remaining message pass-through limitation

The rewrite-secret-message candidate remains unverified, not accepted and not
claimed equivalent. The contract requires preserving the actual dependency
Message. Its secrets dependency contract deliberately does not freeze exact
English, and the authoring tests may not import dependency implementations or
substitute a dependency. Existing nonblank, private, stable-per-code checks do
not distinguish a consistently invented generic message from pass-through.

A direct adapter comparison needs independently obtained actual dependency
output. That oracle is unavailable under this task's restrictions. A keyword or
exact English assertion would manufacture a wording requirement the contracts
do not contain. No such assertion, dependency import or harness substitution was
added.

The authoring API accepts malformed model declarations and must diagnose
independent errors. The first-dot cases exercise this input domain without
changing valid identifier grammar or claiming that every reference spelling
distinguishes alternative splitting implementations.

This document records provenance and distinguishing behavior. Execution results,
mutation findings and approvals belong in the working hardening review.
