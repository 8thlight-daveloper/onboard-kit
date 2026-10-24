# 0390. Admit invitation content before expanding it

Status: accepted by agent judgment; independent contract available, implementation pending

Builds on [0015](0015-functional-core-and-imperative-shell.md),
[0100](0100-behavior-judgment.md),
[0240](0240-invitation-archive-envelope.md),
[0360](0360-confirmed-invitation-start.md),
[0370](0370-plan-authoring-validation.md) and
[0380](0380-invitation-review.md).

## Context

A small plan can repeat a large answer many times. Directory choices and
same-name suggestions can also multiply across questions. A one-megabyte source
limit does not prevent these expansions. Checking the returned view's length
is too late to prevent the allocation. The authoring validator can also expand
inherited condition relationships before rendering begins.

The recipient must be able to inspect an untrusted invitation without exposing
their machine to these known amplification paths. Preserve the actual core
operations rather than writing a second renderer, finder or validator in the
starter. A conservative caller admission rule is a smaller local increment than
adding budgeted variants of every dependency API.

## Decision

Add deterministic resource admission inside the pure invitation reviewer.
It operates on decoded values before the relevant expanding dependencies.
It neither reads memory pressure nor starts a process. It does not select
activities, resolve a placeholder, choose an answer, infer a team or repair data.
This constrains only the new invitation opening boundary. Existing codecs,
saved formats, legacy commands and pure dependency contracts remain unchanged.

### Fixed cardinalities and charged work

Use inclusive limits over complete held declarations, including inactive data,
repetitions, hidden answers and ignored payload fields:

| Counter | Maximum |
| -- | -- |
| Sections | 256 |
| Plan and section variable declarations | 1,024 |
| Steps / checks | 4,096 / 8,192 |
| Condition clauses | 4,096 total; 256 per condition |
| Each comparison-value list / their combined elements | 256 / 16,384 |
| Contacts / roles / teams | 2,048 each; 4,096 combined |
| Each holder list / combined holder occurrences | 1,024 / 8,192 |
| Each Value.List / combined Value.List items | 4,096 / 32,768 |
| Each options list / combined options | 1,024 / 16,384 |
| Held outcomes | 4,096 |
| Known collection entries | 65,536 |
| Model value occurrences / unknown-data occurrences | 131,072 / 65,536 |
| Model traversal depth | 128 |

Use a charged-work threshold B of 67,108,864, not a measurement in allocated
bytes. Every nonnegative addition and multiplication saturates at B+1 without
allocating the predicted result. B is admitted; B+1 refuses. The exact traversal,
field enumeration, deliberate overlaps and terms belong in the independent
[reviewer contract](../../contracts/invitation/reviewer/contract.md).

The score is:

```text
8 * (Base + Render + Questions) + EligibilityWork + PairWork
```

Base charges exact carried source lengths, all held model fields and opaque
unknown data, with fixed occurrence weights and a reserved short first-use event.
Render charges original section/history data plus conservative substitution,
directory-holder/list and command-slot expansion. Count dollar bytes without
parsing placeholder names; escaped or literal dollars may be overcharged.
Questions charges all potential questions, directory choices and same-name
default/given-answer pools without applying activity or suggestion filtering.
EligibilityWork and PairWork charge declaration, inherited-clause and pairwise
work without building the graphs or candidate products themselves.

The source length is the raw archive plus each present local plan, preparation
and record body once. Its existing maximum is five mebibytes, not the one-mebibyte
decompressed envelope limit. Do not count decoded payload bytes a second time.
Defaults contribute only resource charges; they never become answers.

### Placement and exact context

Keep the existing decoder and semantic phase order except for these additions:

1. After incoming plan decode, run full admission using that plan, zero Record,
   nil Preparation and raw archive length. This precedes authoring validation
   and the plan validator's finder calls.
2. After the existing incoming preparation and seed decodes, run full admission
   on that plan, decoded seed and preparation, using raw archive length. This
   precedes seed semantics, locator and preparation validation.
3. After previous-plan decode and before its validators, run full admission on
   that previous plan, zero Record, nil Preparation and its exact source length.
   Failure makes comparison unavailable; it does not prohibit an explicitly
   reviewed incoming-plan replacement or discard the old preimage.
4. Immediately after retained preparation decode, run the cardinality, shape,
   node and depth guards on incoming plan, zero Record and selected preparation,
   sharing counters across those roots. This precedes preparation validation.
5. Immediately after existing record decode, run those same cheap guards on
   incoming plan, decoded record and selected preparation, before identity,
   ownership, history, reconciliation or joining. They do not estimate rendering
   from an unjoined answer set. Absent inputs still have the specified zero/nil
   root weights; the contract fixes their exact treatment.
6. After the actual reconciler and joiner, run full admission on the actual
   effective plan, proposed record and selected preparation, with all carried
   source lengths. Only then call effective plan validation and the walker.
7. Begin repeats admission during its existing source-based re-derivation.
   It reads no new source or clock. The fixed event reservation prevents a
   permitted short event from invalidating a previously admitted review.

Resource refusal may precede a semantic error. It is not validation success.
Return no partial incoming value, proposal, session or publication. Preserve
every source and deliberate choice, including ordinary repeated data.

### Diagnostics and the guarantee

Use fixed, private own resource diagnostics with explicit stage and indexed or
fixed role paths, never rejected values or guessed heap sizes. Fatal incoming,
local or effective admission uses invitation.resource. Previous-plan admission
uses invitation.previous-resource as a warning in the comparison channel and
Comparison.Reason resource. This is an explicit exception to the draft's rule
that every own diagnostic is an error. Exact text, paths and stage spellings
belong in the independent contract.

The guarantee is that an admitted input satisfies the specified conservative
cardinality and fanout policy before expansion. It is not a proof that process
RSS, heap, regex work, parser allocation, elapsed time or every implementation's
returned view stays below 64 MiB. Some dependency text is intentionally unspecified.
The fixed weights and copy margin must not be presented as allocator facts.

False positives are intentional: disabled sections, escaped dollars, unselected
directory roles and suggestions that would be filtered still count. Refuse with
a repair rather than truncate output, prune history, silently lower the budget
or alter applicability. Later hard memory guarantees would need independently
specified bounded leaf operations or isolation, not an undocumented stronger claim.

## Rationale and verification

P3 makes refusal actionable, P4 preserves all work, P2 and 0120 prohibit default
acceptance, and P7 avoids copying personal values into resource diagnostics.
P8 favors one sufficient caller safeguard with an honest limitation. No new
service, installation requirement, model run or generic budgeting framework
is introduced.

Independently specify arithmetic and public input oracles before implementation.
Cover inclusive caps, saturation, unknown depth, inactive/escaped false positives,
directory and suggestion fanout, command slots, graph work, the no-op retained
record and unavailable previous comparison. Derive expected results from literal
inputs and the published formula, never a production estimator.

The complete normative section library with realistic directory data and held
answers must remain a positive control. A small actual rendering probe establishes
reachability without allocating an attack-sized output. Review source placement
as well as refusal outputs: a late check could return the right error after
performing the dangerous expansion. Design calculators and test plans are not
implementation, end-to-end acceptance or a heap guarantee.
