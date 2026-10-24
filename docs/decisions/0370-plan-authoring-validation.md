# 0370. Validate declarations before record-backed onboarding

Status: accepted by agent judgment; independent contract and implementation available

Builds on [0010](0010-plan-editor.md), [0040](0040-blank-text-and-variable-name-length.md),
[0055](0055-variable-audience.md), [0070](0070-sub-sections.md),
[0080](0080-directory.md) and [0360](0360-confirmed-invitation-start.md).

## Context

The record-backed walker requires authoring validation before it is called.
The current plan validator checks stored answers, directory integrity,
declaration identity collisions and structural relationships. Its contract
explicitly leaves other declaration checks for later work. The section loader
and plan decoder preserve semantically invalid declarations; successfully
reading or rendering a plan does not establish that it is valid.

The planned section validator and template linter have no contracts or
implementations. A section-only checker would not establish whole-plan question
references, ancestor eligibility or cross-question condition cycles. The new
start driver must not invent its own declaration checker or claim that existing
partial validation meets the walker's precondition.

## Decision

Add one pure component, plan/authoringvalidator, over the complete decoded Plan.
Its API is New().Validate(Plan), returning diagnostics without changing the plan.
It has position2 and may depend only on model, its own contract and the actual
position1 secrets checker. Its exact codes, paths, ordering, suppression rules
and literal examples belong in the independent contract before implementation.

The new start caller runs this declaration checker first. Only after its errors
are absent does it run the existing real plan validator, before calling finder,
resolver or walker. Both validation stages must succeed; neither alone is called
full validation. Required unanswered questions remain ordinary blockers, not
invalid declarations. Warnings do not silently become answers or readiness.

### Declaration scope

Check every declared scope, including disabled sections. Cover legal identifiers,
required section prose, variable types and audiences, supported timing names,
multiple-value eligibility, choice declarations, RE2 patterns and declared
default semantics. Cover interpolated local references, owner/escalation types,
check-operation shape, environment names, condition references, legal IncludeIf
scope, finite predicate contradictions and question-eligibility cycles.

Check directory identifier spelling and display-name declarations. Existing
duplicate IDs, holders, team overrides, email/reachability and stored-answer
checks remain with the real plan validator. Parent/Requires/After/Bootstrap
structural diagnostics likewise stay with its actual structure dependency.
Do not duplicate those diagnostics or select a duplicate declaration as a winner.

Read parent relationships only to resolve ancestor eligibility safely. Malformed
or ambiguous context must terminate and suppress dependent judgments rather
than invent an ancestor, loop or replace the existing graph diagnostics.

Defaults are validated suggestions, never inputs to applicability or answers.
Pass declared defaults and individual step environment values to the real
secrets checker. Preserve its diagnostic meaning and add safe location metadata;
do not copy its scanner or claim detection of every secret or client-specific
fact in arbitrary prose and unknown fields.

### Preserve existing specific rules

IncludeIf may name plan questions and strict ancestor questions under0070;
the old plan-only reference-table row is superseded. Ordinary cross-section
AskIf remains supported subject to cycle and predicate checks. A plan-variable
AskIf is an authoring error; the finder's defensive handling is unchanged.

Under0040, short dollar forms remain literal and reachable reserved names remain
reserved. Do not reinterpret a short literal as an undeclared variable. Do not
invent required owners, descriptions, optional step fields or filled answers.
Repeated ordinary arguments, options and escalation entries are not identities.

Under0055, an onboardee question must serve section content or eligibility.
Validate its use through real interpolation, section conditions and the necessary
question-dependency closure. Unused onboardee declarations are errors at this new
boundary, not warnings that allow an unnecessary question into the new session.
Do not silently filter the existing finder's output or reject unused onboarder
declarations. Disabled content still counts as authored use.

Preserve input bytes and model values, unknown data and nil/empty shapes.
Validation performs no repairs, normalization, default acceptance, outcome writes
or date enforcement. Source metadata, plan history, runtime answers, preparation
and records are outside this component's declaration analysis.

### Diagnostics and dependency boundary

Diagnostics identify a field and a repair without repeating rejected values.
Invalid identifiers must not be echoed through Question/Section metadata.
Map-entry locations use a stable sorted-key ordinal rather than repeating an
untrusted environment key. The contract defines that ordinal precisely.
Secret messages remain actual dependency results, not newly invented wording.

Reads: declaration fields, directory declarations, local/qualified reference
context and parent links needed for ancestor analysis. Writes: none. Results
are owned diagnostics. No input mutation, persistent state, clock, filesystem,
environment, network, process, record or preparation access is permitted.

This component is not the missing general template linter. Link checking,
prose style, arbitrary client-specific content detection and template-set
authoring remain separate work. No model field, saved format or hash changes.
Adopt this checker initially at the new start boundary; do not quietly tighten
the contracts or behavior of legacy commands and unchecked core functions.

## Alternatives and rationale

A section validator followed by a separate whole-plan checker adds two units
before satisfying the caller precondition. A driver-local checker duplicates
policy inside publication code. A declaration-only model-only component would
leave the sourced default/environment secret checks for another caller layer.
The selected whole-plan unit reuses the actual scanner with one explicit
dependency and keeps storage outside the core.

P1/P3 of [0100](0100-behavior-judgment.md) favor useful questions and actionable
authoring errors. P2/P4 and [0120](0120-defaults-require-acceptance.md) forbid
silently accepting suggestions or repairing saved work. P6 keeps goals separate
from invalidity. P7 favors safe diagnostics and real credential checks; P8
favors the smallest reusable prerequisite rather than another validation framework.

0040 assigns certain checks to the loader, while its current protected contract
explicitly passes them through. This increment implements declaration validation
without silently changing that loader. The discrepancy remains separate contract
reconciliation work, not a claim of existing loader enforcement.

## Verification and delivery

Independently specify all four contract parts before building. Pin valid complete
and incomplete plans, every declaration failure, disabled declarations, defaults
that remain unaccepted, ancestor IncludeIf, cross-section AskIf, cycles,
ambiguous-reference suppression, useful-question closure and actual secret
composition. Expected diagnostics are literal independent data, never generated
from the implementation that they judge.

Run stub controls, deliberate broken implementations, native mutation, a second
independent build and differential checks under the normal hardening process.
Declare architecture/access boundaries before source work. Successful existing
checks or this direction commit are not evidence that the new checker works.
