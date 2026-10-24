# 0440. Edit held catalog instances without transferring history

Status: accepted agent decision. Verification evidence is recorded separately.

0410 requires existing-plan additions and retirement after the creation core.
Provide one pure catalog editor over an ordinary held Plan and ordered explicit
typed operations. No new saved schema or selection manifest is introduced.
Catalog editor at position6 may depend on actual catalog materializer5, plan
editor5, plan validator4, plan authoring validator2, model and its own contract.
It reads held values and compiled dependency results and returns a fresh proposal
and private diagnostics only. No record/preparation, filesystem, clock,
environment, process, network or outcome operation is an input or effect.

This decision narrows the older plan editor claim that it is the only
changed-plan producer: the separately contracted catalog editor may compose its
operations with explicit catalog-instance additions and retirement. Existing
plan editor API, fixtures, Submit, SetSectionEnabled and directory semantics
remain unchanged. This is not authority for other plan producers or protected
edits.

Reserve every held section ID, including disabled sections, and every held
section-answer scope key, including hidden/nil/empty scopes. An addition cannot
inherit another resource's held definition or facts. Deleted information is not
reconstructed and absent record history is not presumed absent. New independent
resources require distinct explicit IDs. An omitted operation changes nothing.

Add materializes a graph-free source with its final permanent ID through the
actual materializer, preserving its exact initial bindings and source baseline.
Never refresh, rehash, repair or replace an existing section, Source or baseline.
Append new definitions in operation order, preserving all held section order.
Input reorder is not an identity edit; it changes only the retained sequence.

Submit explicitly applies plan/editor's existing scoped submission semantics:
empty submissions clear, false and Deferred remain values, and inactive
questions refuse. Unset is a named None submission, not a bypass of activity or
ownership. Only known onboarder local declarations may be changed. Plan-scope
facts, directory editing and definition editing remain separate operations.
Creation Bindings remain exact materializer values and are not Submit aliases.

Retire delegates SetSectionEnabled(false); Reactivate explicitly delegates true
for that SAME held identity, removing Enabled rather than storing true. Neither
changes descendants, relationships, answers, source baselines or outcomes. A
reactivation is not a new resource identity and never asserts access completion.

Operations run in supplied order on a private candidate. Reactivate then Submit
can repair a newly active fact; Retire then Submit refuses inactive questions.
Retire then Reactivate of one identity is explicit, not ambiguous duplicate
declaration. Repeated Submit to one scope is explicit sequential replacement.
The batch is atomic: any failure yields zero Plan, never an accepted prefix.
Validate input declarations/structure/answers through actual validators before
edits and the complete resulting plan again afterward. Existing invalid plans
require a separate repair path; this flow cannot launder invalid data by retire.
Do not guess or rewrite relationship targets or cascade stored switches.

P2/P4 and0110 protect identities and unrelated work. P3 favors fixed actionable
private repairs and clear sequential semantics. P8 favors reusing actual editor
behavior over a second saved-answer policy. P1 does not waive ownership or
0120's explicit-default rule.0270 forbids baseline/history repair. The pure core
does not compare or transfer outcomes; independent history examples explain the
identity boundary without importing a reconciler into this component.

Still required: a separately contracted shell that loads bounded native input,
shows full original/proposed values and diagnostics, obtains current explicit
consent and safely saves the exact reviewed existing-plan update. This pure
proposal supplies neither reviewed save nor whole-binary editing delivery.
