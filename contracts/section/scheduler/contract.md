# Section scheduler contract

Status: agent-authored contract; independent implementation integrated.
Component: `section/scheduler`. Package: `scheduler`.

The section scheduler returns the global action order chosen by
[0260](../../../docs/decisions/0260-prerequisites-and-action-order.md), not the
selector's contiguous grouped reference order. It checks structure before
selection and keeps explicit prerequisites ahead of their dependents.

## Interface and dependencies

`Schedule(plan) -> (sections, problems)`. `New()` takes no arguments.
See [api.go](api.go). The result contains all active activities, not just bootstrap
activities or currently verifiable work. There is no record argument and no
outcome/hash/resolution step. Returned sections are held model values, not text.

Dependency position 3 permits the real
[plan structure validator](../../plan/structurevalidator/contract.md), the real
[section selector](../selector/contract.md), model, and this component's own
contract. The scheduler constructs these dependencies itself, without injecting
alternate selectors or a caller-supplied rank map. Use their complete contracts,
including the selector's Requires applicability increment, not older behavior.

The [binding](wire_test.go) calls the implementation. Before source integration,
nonempty action-output and error-output cases must fail a nil-returning stub;
empty-output and property cases may pass it.

## Validate the full original plan first

Call structure validation on the complete held plan before filtering by switches,
conditions, bootstrap, dates, or any other applicability result. When it returns
problems, return all those problems in their specified order and no sections,
even if every offending activity would currently be inactive. Do not retain a
partial usable sequence, mask a pure cycle as a mixed cycle, reduce the problem
list, repair an edge, or choose one duplicate declaration.

Codes, severity, location, detail, messages, sorting, repeated-reference handling,
and semantic compact-JSON member encoding follow the structure validator's
contract exactly. No extra scheduler diagnostic is introduced in this increment.
Empty/nil problem and section lists are equivalent.

Structural success is not full authoring validation. This component does not
check variables, steps, valid slug spelling, secrets, directory integrity,
condition reference legality, answer validity, or outcome uniqueness. Callers
still use the ordinary full plan validator at the authoring/runtime boundary.
Inputs otherwise in scope cannot produce a post-validation graph deadlock:
the validated declared Requires/effective-After union is acyclic, and activity
filtering only removes nodes/edges. Do not invent a scheduler-only repair rule.

## Select on the full original plan

After structural success, call the real selector with that same complete original
plan. Its result supplies active sections as held and their unique grouped
reference ranks, numbered in result order. Own conditions and section clauses
therefore see all original stored answers and all original held switches;
defaults remain suggestions under
[0120](../../../docs/decisions/0120-defaults-require-acceptance.md).

Requires activity removes candidates transitively, while an active but untried,
failed, or deferred activity stays active. The scheduler does not read outcomes
or turn waiting-for-verification into inactivity. Parent/Requires applicability
is not reimplemented as a recursive outcome walk. Do not build a reduced plan
and call the selector again, or filter to marked work before selecting/ranking.

The rank is in the complete unprojected selected reference list, including active
unmarked activities. Later bootstrap projection happens only after this global
schedule. The ID fallback below is retained from 0260, but unique real reference
ranks mean that final tie cannot independently distinguish valid inputs here.
Do not expand the API or fabricate equal dependency ranks merely to test it.

## Build the active action graph

Create one node per selected section. Every distinct Requires target contributes
an edge from target to dependent. Repeats count once for graph indegrees, but
the returned Requires slice remains exactly held. Successful applicability means
all these targets are selected too.

Also retain effective legacy After edges between selected siblings with the same
effective planning timing. Roots are siblings. The predecessor named by After
points to the holder. Effective timing is a nonempty held own Timing; otherwise
inherit through Parent, defaulting at a root to anytime. Use the complete original
parent definitions for this computation. Only selected endpoints contribute
action edges. Do not forward hints through an inactive section, promote an
ignored cross-parent/cross-timing hint to a global edge, or turn After into a
verification/applicability obligation. An edge present in both Requires and
After still contributes only once.

Parent is containment and timing inheritance only, never an execution edge.
A parent may require its child. Separate subtrees may interleave. Planning
timings are goals, never graph edges, authorizations, verification requirements,
or grounds to reject a real prerequisite with a later target window.

## Choose the next free activity

Repeatedly choose among all unplaced nodes with no remaining predecessor.
Compare this tuple, in order, and choose the smallest:

1. Bootstrap true before false.
2. Effective planning timing: before-day-one, week-one, days-8-to-30,
   days-31-to-90, anytime.
3. Rank in the full selected grouped reference list.
4. Exact section ID by ascending raw UTF-8 bytes.

After placing a node, remove its outgoing constraints and recompute the free
choice. A newly freed earlier-ranked activity can precede an already free later
one. This is not sorting the whole output by the tuple, DFS dependency expansion,
or always keeping a once-created ready queue in insertion order. Prerequisites
and effective After constraints dominate priority because blocked nodes are
never eligible. Bootstrap is an explicit marker; before-day-one does not imply
it, and it is not inherited from a marked parent or inferred from Source.

Every bootstrap prerequisite must itself be bootstrap, and effective After may
not force nonbootstrap before bootstrap, as structural validation enforces.
Thus a marked activity requiring unmarked work is an error, not a valid example
of overriding bootstrap priority. Tests pin priority pressure with valid marked
prerequisites and different timing/rank values instead.

The cross-subtree example in 0260 has grouped rank a,a1,a2,b,b1,b2, A1 requiring
B1 and B2 requiring A2. With equal timing/unmarked nodes, its action order is
exactly a,a2,b,b1,a1,b2. Adding parent-before-child edges or demanding contiguous
subtrees must not reject or reorder that valid graph.

## Results, purity, and scope

Return every selected section exactly once, field for field as held: no filled
timing defaults, sorted/deduplicated Requires data, marker propagation, source
rewrite, or unknown-field loss. The result has fresh section slots, so replacing,
reordering, or changing a returned section field cannot change a plan slot.
Under the existing immutable-value convention, inner slices/maps/pointers may
be shared. Do not impose a new deep-copy requirement for inner section content.
Problems and their detail maps must not alias input state or a later result.

The scheduler is pure and stateless, including concurrent calls on one instance.
It reads no clock, file, process, network, or outcome record. It writes no plan
field or outside state. It reads the section fields needed by structural
validation and selection, held section content for return, and plan answers for
conditions. No hash/record/resolver dependency or new model type is introduced.
Declaration order, repeats of ordinary references, and previous calls cannot
change action order. Never normalize or mutate those original lists to achieve it.

Guaranteed successful ordering covers the selector's existing timing scope:
empty or one of its five known timing values, and valid UTF-8 IDs with unique
identities. A unique empty ID can be a defensive exact identity, as in structure
validation; valid slug authoring belongs elsewhere. Invalid Go UTF-8 and unknown
Timing enum values are outside successful-order guarantees. This is not authority
to add a new enum diagnostic here. Structural errors still follow their own
contract, independent of the active or projected content.

## Independent fixtures and proposed wrong adapters

[cases.json](testdata/cases.json) contains independently hand-specified successful
orders and source citations. [structure-errors.json](testdata/structure-errors.json)
copies literal error oracles from the independently authored structure contract,
not from calls to its implementation. [Tests](scheduler_test.go) never compute
expected orders/problems by calling a selector or validator implementation.
They also check held values, section permutations, shared-instance state and
concurrency, input preservation, fresh section slots, and error-detail isolation.

[Goal priority cases](testdata/goal-priority.json) place a later-goal child before
an earlier-goal root in the grouped reference view. The global action schedule
must still put the earlier goal first. Every pair of known goal windows has a
literal expected order, without fixing invisible numeric rank values.

These wrong adapters are planned, not executed evidence. No implementation,
mutation, independent-build, or differential completion is claimed by this draft.

| Wrong implementation | Intended rejection |
| -- | -- |
| Validate only active/projected sections, ignore some diagnostics, or return a partial sequence with errors | Full original structure-error fixture family and all-disabled variants |
| Return selector reference order or require contiguous parent subtrees | `cross-subtree-global-dag`, `parent-requires-child` |
| Add Parent execution edges or sort whole output by timing | Parent-requires-child, earlier child goal, later-target prerequisite cases |
| Let priorities override blocked prerequisites or cache ready insertion order | Marked prerequisites and newly freed rank fixtures |
| Use global ID instead of reference rank, or forget timing inheritance | Reference rank and inherited planning timing fixtures |
| Treat After as global, forward it through inactive nodes, or count duplicate graph edges twice | Cross-group hints, inactive After, and duplicate-edge fixtures |
| Filter before own condition/switch evaluation or use a default answer | Original-switch context and default/stored answer fixtures |
| Read outcomes or treat before-day-one/parent Bootstrap as a marker | Outcome-shaped unknown data and explicit-marker controls |
| Change held definition fields, mutate input, share result slots, or keep state | Held-value, purity, result-isolation, permutation and concurrent tests |
