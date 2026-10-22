# 0110. Reject ambiguous plans

Status: accepted

Builds on [0100](0100-behavior-judgment.md),
[0055](0055-variable-audience.md), and
[0090](0090-plan-and-onboarding-record.md).

## Context

Choosing the first or last duplicate definition can change what a plan means
or which answers reach a newcomer's record. Deterministic tie-breaking does
not make ambiguous input acceptable. The user chose validation errors rather
than assumptions about author intent, across the levels of a plan.

## Decision

Reject ambiguous plan definitions. Report an error rather than guessing,
silently choosing a winner, merging conflicting definitions, or discarding
information. Diagnostics identify what the author must fix without repeating
secrets.

Keys must be unique within their namespace. This includes two distinct cases:

- **Serialized mapping keys:** a YAML mapping must not repeat a key, even if
  both values agree. Detect repeats before converting the mapping into a
  representation that keeps only one value. Existing decoder duplicate-key
  rules remain in force, including inside preserved unknown data.
- **Domain identities:** a sequence may contain objects whose names or IDs
  serve as keys. Those identities must be unique in their declared scope.
  In particular, two variable declarations with the same name in one scope
  are invalid, even if their metadata agrees. Duplicate section IDs and
  directory identities follow their existing namespace rules. Apply this
  approach to steps and other named entities where the schema defines an
  identity; specify the namespace in the relevant contract.

The same variable name in distinct valid scopes names different questions and
is not itself an error. Likewise, directory lists have separate namespaces.
Repeated ordinary data is not a duplicate key: a repeated command argument,
instruction, or value remains governed by that field's own contract. Do not
deduplicate lists or reject every repeated string under this principle.

The supported authoring flow must diagnose ambiguous definitions instead of
presenting a first/last-wins fallback as legitimate plan behavior. This decides
validation policy; it does not change the splitter's defensive behavior on
unchecked malformed inputs in the same increment. H2's proposed endorsement
of first-wins as plan policy is no longer an open choice.

## Boundaries

An explicit accepted precedence rule is not ambiguity merely because values
differ. For example, 0090 deliberately puts record answers over plan pre-seeds.
Do not reject that supported override or change its precedence.

Missing optional data, intentional deferrals, and preserved unknown fields do
not automatically become errors. Their existing rules still apply. When the
requirements themselves are ambiguous, the behavior judge reports the conflict
rather than inventing a runtime interpretation. This sharpens 0100's ambiguity
example without changing its protected-artifact approval rules.

## Verification and delivery

Use both code mutations and input mutations. A code mutation can remove a
duplicate check; an input mutation can add a conflicting definition to a known
valid plan. The latter needs an independently specified expected error, not
agreement between implementations that could share the same defect.

Exercise mapping-key repeats at each supported nesting level, semantic
identity collisions within their scopes, both conflicting declaration orders,
matching duplicate metadata, and valid cross-scope or repeated-data controls.
Keep raw YAML cases for mapping duplicates: a map-based generator may erase
the duplicate before the decoder sees it.

Contract increments must specify diagnostic codes, locations, ordering, and
the rejection boundary before implementation. This ADR is not evidence that
every current validator or driver enforces the rule. No domain-model field,
schema format, or component access set changes here. Keep implementation work
separate from the already approved bounded hardening changes.
