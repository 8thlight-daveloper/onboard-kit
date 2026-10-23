# Plan structure validator contract

Status: agent-authored contract; independent implementation integrated.
Component: `plan/structurevalidator`. Package: `structurevalidator`.

The plan structure validator reports ambiguous section identities and impossible
section relationships. It follows [0260](../../../docs/decisions/0260-prerequisites-and-action-order.md),
[0070](../../../docs/decisions/0070-sub-sections.md), and
[0110](../../../docs/decisions/0110-reject-ambiguous-plans.md).
It does not select activities or produce an action sequence. A scheduler must
refuse to return a usable sequence when these structural errors exist.

## Interface

`Validate(plan) -> list of Problem`. `New()` takes no arguments.

The API is [api.go](api.go). Dependency position 1 permits only the domain model
and this component's own contract. The core is pure, reads no clock or files,
and keeps no state. It reads only each section's ID, Parent, Requires, After,
Bootstrap, and Timing. It writes no plan field or external state. It reads no
answers, conditions, switches, outcomes, prose, directory, or unknown data.
All declared relationships are checked, regardless of current activity.

The [binding](wire_test.go) calls the implementation. Before source integration,
error-output tests must fail a binding that returns no problems; valid-input
and property tests may pass that stub.

## Identities before relationships

Compare exact held section IDs, without normalization, trimming, or case folding.
IDs in this contract are valid UTF-8; invalid Go string byte sequences are
outside its scope, as in the section hasher contract.
Two or more declarations of one ID give exactly one `plan.duplicate-section-id`
for that ID, even when their contents agree. If any duplicate exists, return
only these duplicate diagnostics, sorted by section ID. Do not choose a graph
winner, resolve references, or report downstream relationship errors in that
call. This matches the existing duplicate code and shape in the
[plan validator contract](../validator/contract.md).

A unique empty ID is not rejected here. ID spelling and valid slug syntax are
other authoring checks. Parent's empty string always means no parent; an empty
string inside Requires or After is an ordinary exact reference, not omission.

## Parent and prerequisite graphs

An empty Parent is a root. A nonempty Parent must name a declared section other
than the holder itself. Report unknown and self references independently of
other defects. Check the remaining known, nonself Parent edges for cycles.
Report one parent-cycle problem per strongly connected component (SCC) with
two or more members. A descendant leading into a cycle is not a member and
gets no extra cycle diagnostic merely for reaching it.

Requires is a set of exact target IDs for relationship checks. Each target must
exist and differ from the holder. Repeat references are ordinary repeated data,
not duplicate identities. Report unknown or self targets once per distinct
target, at its first held list index. Remove these invalid edges from further
graph checks. Report one requires-cycle problem per remaining SCC with at least
two members. Self references have their own problem and never also create an
SCC problem. The same no-double-report rule applies to Parent and After self
references.

Every direct known Requires target of a Bootstrap section must itself have
Bootstrap true. Report each violating distinct target at its first Requires
index. This direct check establishes transitive bootstrap closure: do not
silently mark a target, treat before-day-one as bootstrap, or add a requirement
from Parent. A marked child under an unmarked parent is valid when it does not
require that parent. A parent requiring its child is a valid prerequisite DAG
unless some explicit Requires edges create a real cycle.

Parent and Requires are different graphs. Do not combine Parent with execution
edges, require a verified parent, or reject a valid cross-subtree DAG because
whole parent subtrees cannot remain contiguous. Timing values are goals, not
prerequisites. A week-one activity may require a days-31-to-90 activity.

## Effective legacy After constraints

After is deprecated order-only data, not a prerequisite. An unknown After target
or self target still gets its own diagnostic, as the
[ordering reference](../../../docs/reference/section-template-schema/ordering.md)
requires, even though the selector defensively ignores it. Repeats report once
at the first list index. Remove invalid edges from graph checks.

A known nonself After reference contributes an effective ordering edge only
when both sections have valid parent chains, the same exact Parent (including
two roots), and the same effective Timing. Effective Timing is the held nonempty
Timing; otherwise inherit through Parent to a root, defaulting to anytime.
A missing, self, or cyclic parent anywhere in a section's ancestor chain makes
its grouping unusable. Omit only After edges involving such sections. Still
check their raw unknown/self After references and all independent Requires
relationships; still check unrelated valid After groups.

Compute these effective constraints on the complete declarations before any
activity filtering. Enabled and IncludeIf cannot hide a latent cycle or conflict.
There is no separate cross-timing After error under 0260. Valid non-sibling or
different-effective-timing hints stay inert; do not promote them to global edges
or call them true prerequisites. This reconciles the older timing-conflict
wording in the selector and schema without changing the grouped reference order.
This component does not validate the Timing enum; nonempty held timing text is
compared exactly for grouping, not converted into an execution permission.

Report one after-cycle problem per SCC of effective After edges with at least
two members. For every effective After edge whose predecessor is nonbootstrap
and dependent is bootstrap, report bootstrap-after-nonbootstrap at the dependent's
first After index. Do not report that error for an inert cross-group hint.

## Combined ordering conflicts

Execution edges point from prerequisite/predecessor to dependent. Form their
union from known nonself Requires edges and effective After edges only. Parent
containment and timing comparisons add no edges. Identical edges appearing in
both lists are one graph edge, not a conflict.

For each union SCC with two or more members, inspect its induced pure Requires
and After graphs. If either contains a cycle already reported above, do not
add an ordering-conflict problem for that SCC. The existing pure-cycle problem
already rejects it; repairing it may reveal a remaining mixed conflict on a
later call. Otherwise report exactly one ordering-conflict for that SCC: the
pure graphs are each acyclic but cannot coexist. Never label this a true
requires-cycle. A pure cycle in another union SCC must not suppress an unrelated
mixed conflict. This is diagnostic noise suppression, not permission to accept
an impossible union or discard an authored edge.

## Diagnostics

All problems have severity error, empty Question, and exactly the Detail keys
below. Section identifies the holder, except an SCC uses its smallest exact
member ID. No byte offset or plan.Sections index is reported: these are model
relationships, not serialized locations. `item` is the zero-based first held
list index in ordinary decimal. `members` is a compact JSON array of exact
member IDs, sorted by ascending raw UTF-8 bytes, without duplicates. Any valid
JSON string escaping is allowed: decoded member IDs, not one spelling of their
escapes, are the observable identity. For example literal less-than and its
Unicode JSON escape represent the same member. The JSON array must remain
compact, valid, and deterministic within an implementation. This representation
is unambiguous even for unchecked unusual valid UTF-8 IDs.

Messages are the fixed repair instructions below. They never include an
unknown reference value, title, body, answer, default, note, or unknown data.
Declared section IDs may occur only in Section and the SCC members location.
The field and first index identify a bad reference without copying potentially
sensitive arbitrary target text into diagnostic messages or detail.

| Code | Detail | Message |
| -- | -- | -- |
| `plan.duplicate-section-id` | none | Give each section a unique ID before checking relationships. |
| `plan.parent-unknown` | `field: parent` | Choose an existing section as parent, or remove parent. |
| `plan.parent-self` | `field: parent` | Remove the section's reference to itself as parent. |
| `plan.parent-cycle` | `field: parent`, `members` | Remove a parent link within the listed cycle members. |
| `plan.requires-unknown` | `field: requires`, `item` | Choose an existing prerequisite, or remove this requires item. |
| `plan.requires-self` | `field: requires`, `item` | Remove the section's reference to itself in requires. |
| `plan.requires-cycle` | `field: requires`, `members` | Remove a requires edge within the listed cycle members. |
| `plan.bootstrap-requires-nonbootstrap` | `field: requires`, `item` | Make this prerequisite bootstrap, or correct the bootstrap claim. |
| `plan.after-unknown` | `field: after`, `item` | Choose an existing section, or remove this after item. |
| `plan.after-self` | `field: after`, `item` | Remove the section's reference to itself in after. |
| `plan.after-cycle` | `field: after`, `members` | Remove an after edge within the listed cycle members. |
| `plan.bootstrap-after-nonbootstrap` | `field: after`, `item` | Remove this ordering hint, or correct the bootstrap claim. |
| `plan.ordering-conflict` | `field: ordering`, `members` | Change a requires or effective after edge within these members so both orders can coexist. |

Outside the duplicate-only phase, collect all independent applicable diagnostics.
Sort by Code as raw UTF-8 bytes, then Section, then Detail.field, then item as a
nonnegative integer (absent before present), then decoded member IDs
lexicographically by raw UTF-8 bytes. Emit each
identical problem once. Reference list order is held author data, so its first
index can change when that list is edited; section-list permutations cannot
change any result. Nil and empty problem lists are equivalent.

## Properties and scope

- Pure: all input fields and backing storage stay unchanged.
- Stateless: one instance gives the same output after other plans and after its
  prior returned problem details have been edited, including concurrent calls
  on that same instance.
- Permutation invariant: reverse, rotate, or reorder section declarations without
  changing diagnostics. No graph winner depends on declaration order.
- Fresh results: changing returned problems or their detail maps changes neither
  an argument nor the next result.
- Finite: invalid references and cycles return problems, never loop.

No full-plan validation is promised here. Other components own variable and step
identity scopes, schema/type guards, valid IDs, condition references/cycles,
answers, secrets, directory integrity, activity cascading, actual scheduling,
outcome verification, and files. No hidden result means validation succeeded at
those other boundaries. This component writes no repaired graph or plan.

## Fixtures and wrong implementations

[cases.json](testdata/cases.json) contains hand-specified exact diagnostics and
valid controls, with a source for each case. No expected graph or cycle is
computed by the component under test. [Tests](structure_test.go) also exercise
section permutations, input preservation, reused instances, ignored plan data,
and mutation of returned problem maps.

[Hardening cases](testdata/hardening.json) independently pin whitespace-bearing
identities, empty After references, bootstrap closure through containment,
complete ancestry despite own timing, exact nonempty timing groups, preservation
of nonempty ignored content, and separate SCCs linked by a one-way edge.

The wrong implementations below were applied in an isolated reference build.
The contract suite rejects them. A second implementation was built from the
contract alone; differential inputs cover exact identities, independent graph
kinds, inherited grouping, bootstrap closure, diagnostic order, and privacy.
Run inventories, limits, and results belong in working evidence, not this
contract. Automated mutation coverage remains separately reported there.

| Wrong implementation | Cases or tests intended to reject it |
| -- | -- |
| Choose first/last duplicate declaration and continue graph checking | `duplicate-identities-stop-graph-checks`, `duplicate-identical-and-conflicting` |
| Treat empty ID as absent instead of exact identity, or sort item indices as text | `unique-empty-id-is-a-valid-exact-target`, `duplicate-empty-identities`, `empty-id-requires-itself`, `reference-indices-sort-numerically` |
| Skip inactive relationships or treat defaults as answers | `disabled-latent-requires-cycle`, `excluded-bootstrap-closure`, ignored-data test |
| Add Parent execution edges or enforce contiguous subtrees | `parent-requires-child`, `cross-subtree-dag` |
| Reject a later-goal prerequisite | `later-goal-prerequisite` |
| Report downstream tails as cycle members or one problem per cycle edge | `requires-cycle-with-tail`, `parent-cycle-with-tail`, `two-parent-components` |
| Merge all cycle kinds or skip every mixed SCC after one pure cycle | `mixed-ordering-cycle`, `disjoint-pure-and-mixed`, `nested-pure-cycle-suppresses-mixed-echo` |
| Treat After as global, or fail to inherit effective timing | `inert-cross-parent-hints`, `inert-cross-timing-hints`, `inherited-after-cycle` |
| Suppress unrelated graph checks after a bad parent | `bad-parent-and-independent-graphs`, `bad-parent-chain-omits-only-affected-after` |
| Infer Bootstrap from dates/parent or silently mark prerequisites | `marked-child-unmarked-parent`, `bootstrap-transitive-closure`, `excluded-bootstrap-closure` |
| Leak raw bad targets or repeat diagnostics for ordinary references | `private-reference-locations`, `repeated-reference-locations` |
| Keep state, mutate input lists, or reuse result maps | State, permutation, purity, and result-isolation tests |
