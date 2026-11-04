---
category: behavior
status: accepted
approval: agent
affects: [catalog/creator, catalog/editguide, catalog/editor, catalog/materializer]
builds-on: ['0030', '0100', '0110', '0120', '0150', '0270', '0320']
---

# 0410. Select explicit catalog instances with stable identities

Builds on [0030](0030-self-contained-plan-file.md),
[0100](0100-behavior-judgment.md),
[0110](0110-reject-ambiguous-plans.md),
[0120](0120-defaults-require-acceptance.md),
[0150](0150-onboarding-library-and-preparation-flow.md),
[0270](0270-outcome-definition-hashes.md), and
[0320](0320-embedded-catalog-availability.md).

## Context

The compiled catalog supplies definitions, not selected organization plans.
A team may need two repository access activities or several reading assignments.
Each needs its own identity, answers and outcomes. A title, URL, list position
or template ID cannot identify these independent instances.

0030's existing default creation includes every supplied template unchanged.
Its instantiator cannot rename repeated copies while preserving their original
catalog identity. Explicit catalog selection is a separate creation path, not
a change to that legacy behavior or to existing saved files.

## Decision

### Keep the plan as the lasting configuration

The caller explicitly selects exact section-template catalog IDs and supplies
a permanent section ID for every instance. Select nothing implicitly. Reject
duplicate instance IDs even when their definitions agree. Guidance and tool-help
carriers cannot become newcomer activities.

The resulting organization plan contains ordinary unresolved sections and exact
scoped answers. It works without its source catalog. Do not add a separate
executable selection manifest, encoded policy in unknown fields, personal plan
overlay, schema field or generic repeat-expression language.

Allow repeated selections of any supported graph-free section template when
each has a distinct permanent ID. The current catalog's section definitions are
graph-free; repository and reading repetitions are required examples, not an
exclusive allowlist. Do not generate instances from list positions, titles,
URLs, answers or definition hashes.

Preserve declaration and step identities inside each independent scope. A future
definition with Parent, Requires, After or condition-reference relationships
needs a separately specified exact instance mapping before renaming. Until
that profile is supported, refuse it explicitly rather than guessing targets,
rewriting substrings or dropping relationships. The contract pins the admitted
shape and validates the whole resulting plan through actual dependencies.

### Bind only deliberately supplied shared facts

Each explicit binding names an exact declared onboarder variable in its new
section scope. Missing audience retains its existing onboarder meaning. Unknown
targets and onboardee bindings refuse. The same variable name in two instances
does not share an answer unless the caller explicitly supplies both values.

Defaults and examples remain suggestions. Do not resolve variables, fill missing
answers, copy personal preparation into answers or manufacture completeness.
Actual answer validation still applies; missing required facts remain visible
to the existing finder. No recipient, arrangement, first-use event, permission
or outcome is created by selecting an activity.

### Preserve source identity and the correct starting baseline

For this new creation path, copy the selected unresolved definition, assign its
explicit permanent section ID, and hash that final-ID definition with the actual
unchanged section hasher. Then attach Source.Template with the original catalog
ID and Source.Hash with that hash. Source is excluded from the hash.

This makes an untouched instance match its own stored baseline. Two copies
with different permanent IDs can have different hashes because the current
canonical form includes the section ID. Answers and Enabled remain excluded.
The hash says nothing about resource identity, access or authorization.

Do not pass renamed copies through the old instantiator and thereby report the
instance ID as its template identity. Do not change its behavior, canonical hash
algorithm, old Source values, existing outcome hashes or saved plans.

A future upstream comparison must construct the candidate with the SAME
permanent ID before comparing its definition to this baseline. That update
operation remains separately contracted work. An older hand-renamed section
may have an original-ID baseline; do not infer its basis from dates or unequal
Source.Template and section ID, repair it silently or claim a mismatch proves
when an edit occurred. Preserve it for explicit conflict review.

### Surface advisory preparation without claiming it happened

Use the exact [catalog preparation relations](../reference/catalog-preparation-relations.md).
Return each relevant guide once, with every causing permanent instance ID.
Sort guide IDs and their instance reasons by exact byte order. Keep shared
guidance when another selected activity still needs it.

These are authored relationships, not inferred links from prose, titles or
variable names. They do not extend0320's reader manifest. Conditional office
and remote advice is not a selected personal work location or a demand for
both arrangements. A later person flow may filter advice using explicit
context; selection does not guess it.

### Preserve identity during later editing

The first pure materializer creates a new proposal; it receives no existing
plan or personal history and must not claim synchronization or retirement.
Existing-plan selection editing and safe reviewed save remain required work.

That later edit flow reserves every held section ID, including disabled ones.
Retiring an instance keeps its definition, Source and scoped facts and sets
Enabled false. A new independent resource needs a new ID. A URL change alone
cannot decide whether a resource moved or was replaced; the owner explicitly
chooses an identity-preserving edit or a new independent activity.

Do not transfer old outcomes, rewrite a record, remove reservations merely
because a request list omits them, or silently repair affected relationships.
Ordinary manual edits and older explicit pruning can remove information; this
flow cannot reconstruct a deleted ID's history or infer a replacement identity.

## Components and effects

The catalog materializer is pure at dependency position5. It may use model,
its own contract, catalog reader1, section loader1, section hasher1, plan
authoring validator2 and plan validator4. Construct the actual dependencies;
do not copy parsers, hash algorithms or validation policy.

It reads explicit creation context, selections and bindings, compiled definitions,
the authored relation data and dependency results. It writes only a returned new
plan, relevant guidance and diagnostics. No input mutation, filesystem, clock,
environment, network, preparation, record or existing-plan access is added.
A later driver owns reviewed save and must declare its effects separately.

## Rationale and verification

P1 and P3 make known shared facts and needed preparation useful. P2, P4 and0110
protect explicit ownership, stable identity and saved work. P5 preserves offline
self-contained plans. P6 keeps targets neutral; P7 keeps personal assignments
out of shared selection configuration. Existing commitments constrain convenience.

Independent contracts must pin selected carriers, exact IDs and scopes, no default
acceptance, duplicate refusal, repeated nonrepository examples, complete ownership
and guidance unions. Hash fixtures use hand-authored final-ID canonical texts
and standard SHA256, never expected values from the product hasher.

Whole-binary scenarios must show offline selection and reviewed save, independent
repeated outcomes across reorder/addition/retirement, unchanged personal files
and preserved old baselines. Creation-only tests do not establish those later
behaviors. This decision records direction, not passing tests or delivery.
