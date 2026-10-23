# 0260. Prerequisites and action order

Status: accepted by agent judgment; implementation pending

Builds on [0085](0085-walking-sections-and-recording-outcomes.md),
[0110](0110-reject-ambiguous-plans.md),
[0150](0150-onboarding-library-and-preparation-flow.md), and
[0250](0250-bootstrap-content-and-projection.md).
Reconciles 0085's immediate After replacement and timing-first order with saved
plans. Separates the action sequence from the grouped reference view of
[0070](0070-sub-sections.md). Supersedes 0250's planned nested action-fragment
headings when the prerequisite scheduler is integrated.

## Context

A reading preference is not proof that another activity must be verified.
Containment is not a prerequisite either. Suppose group A has activities A1
and A2, and group B has B1 and B2. A1 requires B1; B2 requires A2. This graph
has no cycle, but no ordering of whole contiguous parent subtrees can satisfy
both edges. A parent that requires its child is another valid graph: the child
can be verified before the parent. Adding parent-before-child execution edges
would invent a cycle.

The existing selector provides a grouped reference order. It cannot stand in
for a general prerequisite graph. Keeping that limitation, silently converting
After, or treating target dates as permission would change the product chosen
in 0085 and 0150. These decisions apply P2, P3, P4, P5, and P6 of
[0100](0100-behavior-judgment.md).

## Decision

### Requires is explicit section content

Add Requires, a list of exact section IDs, to Section. It is optional and empty
when absent. It is not an answer, preference, date, or result. Do not derive it
from After, Parent, titles, selected defaults, or recorded outcomes.

Retain After as deprecated order-only compatibility data. New embedded library
content uses Requires for real prerequisites, not After. Older After-bearing
plans remain readable without acquiring new verification obligations. A later
removal needs its own format and preservation decision. This narrows 0085's
immediate field-removal choice, not its prerequisite requirement.

Both fields serialize independently. Requires appears after After and before
Owner in canonical section key order. Keep the existing list guards, raw-key
duplicate checks, exact strings, list order, and repetitions in codecs. Hash
Requires as a sorted, deduplicated set in that position. Empty and nil lists
are omitted, preserving every old hash without Requires. Repeated identical
references are ordinary repeated data, not duplicate declarations under 0110.
A colliding unknown field is refused even when the held list is empty.

This is an additive optional section field under
[0030](0030-self-contained-plan-file.md), with schema version 1. An older binary
preserving unknown fields does not thereby enforce prerequisites. New runtime
instructions must use the capable binary and its verified commands, not imply
that preservation establishes equivalent behavior.

### Applicability and waiting are different

Evaluate each section's own condition and switch against the complete original
plan. Defaults remain suggestions under
[0120](0120-defaults-require-acceptance.md). A section clause still tests the
named section's switch, not its derived activity or verification state.

Start with all sections whose own gates hold. Repeatedly remove a candidate
whose parent or required section is missing or inactive until no more candidates
are removed. This greatest remaining active set preserves the valid case where
parent A requires child A1: both remain active when both own gates hold. Turning
either off propagates inactivity. Do not mistake mutual applicability support
for a prerequisite cycle; parent edges are not prerequisite edges.

Outcomes play no part in applicability. A failed, deferred, or untried required
activity stays active. Its dependent stays active but waits for verification.
Do not hide that dependent, mark it failed, or erase its saved answers.

### Keep two honest orders

The grouped reference view keeps the selector's existing parent hierarchy and
reading-order behavior. It is a way to browse content, not a promise of execution
order or eligibility. Show explicit prerequisites when presenting such content
as instructions. Existing full-document output without Requires stays unchanged.

The action overview, first steps, and walkthrough use a global topological
sequence over active sections. Every Requires edge places the prerequisite
before the dependent. Parent adds no execution edge and never requires a
verified parent outcome. Among currently free nodes choose, in order:

1. Bootstrap before nonbootstrap.
2. Effective planning timing, inheriting through Parent as before.
3. Rank in the full grouped reference order.
4. Exact section ID in ascending byte order.

Goals break ties among free activities; they never put a real prerequisite
later, create an authorization, or become a failure or verification gate. A
week-one activity may require a later-target activity. An inconsistent goal is
not a reason to reject or bypass that prerequisite. This supersedes treating
such timing differences as dependency errors in the schema ordering reference.

Retain only the legacy After edges that existing reference ordering honors:
references between active siblings in the same effective timing group. Those
effective edges also constrain action order, but never verification or activity.
Do not promote previously ignored cross-group hints into new global constraints.

For the example above, use equal timings and IDs a, a1, a2, b, b1, b2, with
their stated parents and no After hints. Reference rank is a,a1,a2,b,b1,b2.
The specified free-node choices give a,a2,b,b1,a1,b2. Both real prerequisites
are earlier, without requiring a whole subtree to stay contiguous.

### Diagnose impossible relationships

Reject duplicate section identities before resolving graph references. Do not
choose a declaration. Validate declared parent and Requires references and
cycles even for currently inactive sections. Unknown references, self-requiring
sections, and true prerequisite cycles are errors. Duplicate references do not
create additional prerequisite nodes or completion obligations.

Validate effective legacy ordering constraints separately from prerequisites.
A pure effective After cycle is an ordering error. An acyclic Requires graph
and an acyclic After graph may still have an impossible union: B requires A
while A is after B. Report an ordering conflict, not a false prerequisite-cycle
claim, and do not silently discard either authored constraint. Parent contiguity
and timing goals are not graph edges, so cannot manufacture this conflict.
Check the declared effective constraints before current activity filtering so
switching off a section does not hide a latent conflict.

One cycle diagnostic per strongly connected component is enough. Contract
fixtures must specify exact codes, locations, member ordering, deduplication,
and privacy-safe repair messages before implementation. A scheduler returns no
usable sequence when its structural validation fails. The ordinary plan
validator still owns the complete authoring validation boundary.

### Bootstrap is a real first stage

Every direct prerequisite of a bootstrap section must be bootstrap. Check all
declared edges, regardless of switches or dates; this establishes transitive
closure too. Do not silently add markers. A marked child under an unmarked
parent remains valid when it does not require that parent.

An effective After hint placing nonbootstrap work before bootstrap work also
contradicts the first stage and is an ordering error. It is not evidence that
the work is a true prerequisite. The author corrects the hint or the claim that
the activity is pre-access work.

Validate and select from the full plan, schedule the full action graph, then
project bootstrap activities. Never filter the plan before evaluating conditions,
ancestor gates, or relationships. Render action fragments with flat activity
headings and explicit containment context, reusing the real assembler's section
and gap formatting. Do not imply a contiguous tree by nesting interleaved actions,
include unmarked ancestor instructions or gaps, slice finished Markdown, or copy
private formatting into an invitation component.

0250's marker, codecs, hash rules, full-plan context, and absence of implicit
parent/child marking remain unchanged. Its nested action projection is still
unimplemented; replace that planned presentation before activating its contract,
not by building a temporary renderer that immediately needs replacement.

### Verification preserves recorded facts

A new verified outcome requires one unambiguous stored verified outcome for
every direct current Requires target. Failed, deferred, or absent outcomes do
not qualify. Parent, After, dates, file metadata, and defaults supply no missing
evidence. A runtime update rechecks current inputs at the write boundary.

A stored verified outcome continues to count when its definition-difference
flag is set. Show the flag; do not invent a hidden freshness gate. An existing
verified dependent is not retroactively reset if its prerequisite later fails
or the author adds a requirement. A new verification attempt obeys the current
graph. Preserve outcomes, timestamps, notes, differently text, and first use.
Duplicate outcomes are ambiguity errors, never first/last-wins.

The hash used for definition-difference reporting needs a separate explicit
reconciliation of 0030's template baseline and 0085's current-content wording.
Do not silently rewrite Source.Hash or outcome history to resolve it. Close
that boundary before the new runtime claims to report current definition changes.

## Component boundaries

The implementation sequence keeps contract and source authors independent.
All core components are pure, change no argument, and perform no outside I/O.

| Component | Reads and returns | Dependency position and allowed project dependencies |
| -- | -- | -- |
| Plan structure validator, new | Reads section identities, Parent, After, Requires, Bootstrap, and Timing; returns relationship problems only | 1: model |
| Section selector | Reads full-plan answers, conditions, switches, parents, and Requires; returns held active section values in reference order | 2: condition evaluator, model |
| Section scheduler, new | Reads the full plan and actual selector/structure results; returns action-ordered sections and problems, never outcomes | 3: plan structure validator, section selector, model |
| Document assembler | Reads the full plan; returns reference documents or scheduled bootstrap fragments | 4: section scheduler, section selector, section resolver, model |
| Section walker, new | Reads plan and record; returns action, waiting, outcome, and definition views | 4: section scheduler, question finder, section resolver, section hasher, model |
| Outcome recorder, new | Reads plan, record, explicit proposed outcome and injected timestamp; returns a copied record or refusal | 5: section walker, model |
| Plan validator | Composes existing answer/declaration checks and relationship validation; returns problems | 4: plan structure validator, question finder, section validator, secrets, model |

These positions extend [0020](0020-binary-structure-and-regenerative-components.md)
and move the document assembler from position 3 to 4. The architecture view and
each component contract must state their exact read/write sets before building.
The model's Requires slice needs independent copy/isolation tests wherever a
plan or section is returned. Codecs and hasher come before graph/runtime builds.

## Verification and delivery

Pin independent canonical hashes, list preservation, type guards, raw duplicate
keys, zero-held known-key collisions, and slice isolation. Graph fixtures must
include cross-subtree prerequisites, a parent requiring its child, true and mixed
cycles, inactive cascades, unknown references, bootstrap closure, and a later
target prerequisite. Hand-specified orders and outcomes are the oracle, not a
second call to the selector being judged.

Runtime scenarios must distinguish inactive from waiting, preserve historical
verification on content changes, refuse ambiguous outcomes without writes, and
record new verification only after real prerequisites. These are planned tests,
not implemented safeguards. Invitation/import/start, first use, the full embedded
catalog, repeated instances, neutral support views, and exact native-artifact
verification remain required work under the unchanged release roadmap.
