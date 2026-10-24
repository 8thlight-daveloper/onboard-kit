# Plan authoring validator contract

Status: agent-authored contract; implementation pending.

`plan/authoringvalidator` checks declarations in a decoded whole plan. It does
not judge stored answers, readiness, record history or preparation. The new
start caller first runs this component, then the actual existing plan validator
when declaration errors are absent. Success from both establishes the specified
runtime authoring profile; success from either alone does not.

## Interface and boundary

`New().Validate(model.Plan) -> []model.Problem`. `New()` has no arguments and
constructs the actual secrets checker. Position 2 permits model, this contract,
the standard library and the actual position-1 secrets component only. No file,
environment, clock, network, process, answers, record or persistent state is read.
Do not call finder, evaluator, resolver, selector or another validator. No model
values are returned, repaired, normalized or filled from defaults.

Read the plan's variables, directory declarations and all sections' declarations,
regardless of Enabled, stored answers, conditions, prerequisites or audience.
Parent is read only for ancestor eligibility. Requires, After and Bootstrap are
not judged. Plan identity, schema/version, source/hash, provenance, origin,
resultsPath, URL, unknown fields and stored Answers are ignored. Unknown fields
are never traversed for authoring or secret findings.

The existing plan validator retains exact duplicate section/variable/step IDs,
multiple plan-team variables, graph relationships, directory duplicate/holder/
role integrity, stored answers and stored-answer secrets. Emit none of its
diagnostic codes. This component supplies missing syntax/declaration semantics,
not a second decoder, graph validator or general content linter.

## Diagnostics

Every local diagnostic has `Severity:error`, an exact code from the table,
`Message` exactly the table's repair sentence and `Detail` exactly
`{"path": PATH}`. An empty condition is a warning; unused onboardee questions
are errors. Delegated secrets preserve
their real message and detail as below.
No message repeats rejected text, patterns, defaults, environment values or
condition values. Question is `plan.NAME` or `SECTION.NAME` only for a variable
or its condition/default with a legal nonreserved name and legal scope. Section
is the section ID only when legal and nonreserved, otherwise empty. Invalid
identifiers are never copied into metadata; indexed paths locate their repair.

PATH is an indexed model-field path: `variables[0].name`,
`sections[1].variables[2].askIf.all[0].answer.is`,
`sections[0].steps[1].checks[2].env`, or `directory.contacts[0].name`.
Indices are zero-based decimal without padding. A step environment map value
uses `sections[I].steps[J].env[N]`, where N is the zero-based rank among exact
map keys sorted bytewise, not serialized order. Neither rejected keys nor
environment values enter diagnostic paths. A check env is a
single `.env`. A diagnostic about a whole condition uses its condition path;
cycle/unused diagnostics use the variable's path. Section prose uses lowercase
field names (`title`, `summary`, `verification`, `body`, `owner`). List fields
append `[N]`. Return all independent diagnostics, sorted by bytewise PATH then
Code, Question, Section, Severity and sorted-key compact JSON Detail. Exact ties
retain multiplicity. Nil and empty problem lists are equivalent.

| Code | Repair sentence |
| -- | -- |
| `authoring.section-id` | Use a nonreserved lowercase section slug. |
| `authoring.step-id` | Use a lowercase step slug or omit the identifier. |
| `authoring.required-prose` | Supply nonblank required section prose. |
| `authoring.variable-name` | Use a nonreserved variable name with at least three characters. |
| `authoring.variable-type` | Choose a supported variable type. |
| `authoring.variable-audience` | Choose onboarder or onboardee audience. |
| `authoring.multiple` | Allow multiple values only for contact or role variables. |
| `authoring.choice-options` | Supply at least one nonempty choice option. |
| `authoring.pattern` | Use a valid RE2 string pattern. |
| `authoring.default` | Supply a default that fits its declaration. |
| `authoring.timing` | Choose a supported timing goal. |
| `authoring.directory-id` | Use a lowercase directory identifier. |
| `authoring.directory-name` | Supply a nonblank directory display name. |
| `authoring.dollar` | Escape the dollar sign or use a declared local variable. |
| `authoring.owner-type` | Reference a contact or role variable in owner text. |
| `authoring.escalation` | Use one local contact or role variable reference. |
| `authoring.environment-name` | Use an uppercase environment variable name. |
| `authoring.check` | Supply exactly one meaningful check operation. |
| `authoring.plan-ask-if` | Remove askIf from the plan variable. |
| `authoring.condition-empty` | Remove the empty condition or supply its clauses. |
| `authoring.clause` | Supply exactly one question or section-switch clause. |
| `authoring.question-reference` | Reference an existing unambiguous declared question. |
| `authoring.section-reference` | Reference an existing declared section switch. |
| `authoring.include-scope` | Gate a section only on plan or ancestor questions. |
| `authoring.predicate` | Supply a supported meaningful answer predicate. |
| `authoring.predicate-value` | Use a settled scalar value accepted by the referenced question. |
| `authoring.predicate-list` | Use at least two values without a boolean or all-options list test. |
| `authoring.contradiction` | Remove mutually incompatible condition tests. |
| `authoring.condition-cycle` | Remove the question eligibility dependency cycle. |
| `authoring.unused-onboardee` | Use this onboardee question in section content or eligibility. |

## Declaration checks

Section IDs match `[a-z0-9][a-z0-9-]*` and exclude `plan`, `yes`, `no`, `y`, `n`,
`true`, `false`, `on`, `off`, `null`. Nonempty step IDs use that grammar without
the section reserved set; empty step IDs remain allowed. Do not require a step
name, owner, run, open, notes, check or verification that the schema leaves
optional. Repeated command arguments, options and escalation references remain
ordinary data, not duplicate identities. Sections require title and summary
with at least one non-Unicode-whitespace rune. Optional prose, including blank
description, hint, name, notes, body and verification, creates no required-prose
error and is not modified. Unicode line/paragraph separators count as whitespace.

Variable names match `[a-z][a-z0-9_]{2,}` and exclude `enabled`, `yes`, `true`,
`false`, `off`, `null`. One name error per declaration covers both syntax and
reservation. Empty Type means string; otherwise exactly string, boolean, choice,
command, contact, role or team. Empty Audience means onboarder; otherwise exactly
onboarder/onboardee. No trimming/case folding. Multiple is allowed only on contact
and role. Invalid type suppresses dependent multiple/options/pattern/default
checks, not independent name/audience/AskIf/secret checks.

A choice needs a nonempty Options list with no empty string; one options error
at `.options` covers all defects. Whitespace is an actual option. Repeated options
are allowed. Options on other types are ignored. Compile a nonempty Pattern only
on string variables as RE2; irrelevant patterns/hints are allowed and ignored.
RE2 search semantics apply, not implicit anchors. Invalid pattern or choice
options suppresses dependent default and predicate membership/pattern checks,
not independent kind checks. A patternHint is not required.

A None default means no suggestion, regardless of inactive payload fields.
Deferred is allowed for every otherwise valid declaration. For other defaults,
use only the payload selected by Kind: flag for boolean, list for command or
multiple contact/role, text otherwise. Unknown Kind and mismatched Kind produce
one `authoring.default` at `.default`. Matching string text, including empty or
whitespace, must match any valid string pattern; matching choice text must occur
exactly in Options. Directory text/list items must name the correct declared
namespace exactly; empty items do not name an entry. Empty lists are permitted
suggestions; ordered repeats are valid. A directory lookup against duplicate IDs
does not select a winner: a present duplicate target suppresses membership
judgment for that ID, leaving the existing validator's ambiguity error. Shape
errors still report. Emit at most one authoring.default per declaration at
`.default`, even if several items fail. Default checking never creates or consumes an answer.

Nonempty section/step Timing is exactly before-day-one, week-one, days-8-to-30,
days-31-to-90 or anytime. Empty timing inherits/defaults elsewhere and is valid.
Cross-timing prerequisites are not errors here.

Directory contact/role/team IDs use `[a-z0-9][a-z0-9-]*`; one bad-ID diagnostic
per entry. Each entry requires a display Name containing a non-whitespace rune.
This explicit display-name rule does not trim or rewrite names and is not an
assertion that 0040 enumerated directory names. Holders/contact/email/Slack
integrity, availability and duplicate IDs remain with the existing validator.

## Text, owners, escalation and checks

Scan exactly the resolver's filled fields: section title/summary/owner/
verification/body; step name/notes/owner/open/directory, Run items and Env values;
check Run items and Ask. Do not scan variable metadata/defaults, condition values,
IDs, source, env keys or unknown fields for placeholder syntax.

Read once left-to-right, consuming `$$` first as an escaped dollar. After a single
`$`, a lowercase-letter-started maximal `[a-z0-9_]+` run of length one or two is
literal, as is any digit-started maximal `[a-z0-9_]+` run. A lowercase-started run
of at least three is a placeholder requiring exactly one declaration in the same
section; plan-only declarations do not suffice. All other single-dollar forms
(including trailing dollar, uppercase start, underscore start and `${name}`)
are malformed. A declared prefix does not shorten the maximal run; an uppercase
letter ends it. Emit at most one `authoring.dollar` per scanned field containing
any malformed or unknown placeholder, not one per occurrence. Duplicate local
declarations suppress unknown-reference judgments for that name without choosing
a winner; other dollar errors still report. Escaped/short runs are not uses.

Literal owner prose is allowed. Every unambiguous placeholder in section/step
Owner must be contact or role; one `authoring.owner-type` per owner field if any
known valid non-directory type is referenced. Unknown types suppress dependent
owner/escalation-type judgments, leaving the declaration's variable-type error;
unknown references instead produce dollar errors. Escalation entries must be
exactly `$NAME` for one local contact/role variable, with no surrounding spaces,
escape, prose or suffix. Emit one escalation error per bad item; do not also
scan escalation as prose. Missing escalation is valid and missing answers are
not authoring errors. Duplicate target declarations suppress dependent escalation
and owner-type judgments without selecting a winner. Repeats preserve meaning/order.

Step Env keys and nonempty check Env use `[A-Z][A-Z0-9_]*`. A check needs exactly
one meaningful operation: Run length > 0, nonblank Ask, or Env != empty. Empty
Run and blank Ask are absent at this model boundary; no attempt is made to infer
serialized key presence. Invalid Env still counts as one supplied operation and
receives the name error, not a missing-operation error. Multiple meaningful
operations yield one check error; independent text/env checks still run. Empty
run arguments and empty env values are valid literal data. Steps may have both
Run and Open; checks never execute during validation.

Pass every declared Default unchanged to the actual secrets check, and each step
Env value as a separate Text value, even after an authoring error or in disabled
sections. Do not scan any other declaration text for secrets. Preserve each real
secret Code, Severity, actual Message unchanged and existing Detail (including
list `item`),
then add `path` for `.default` or the Env entry, and set Question/Section from
its location. Return fresh owned Detail maps; do not rename secret codes or copy
the scanner. The existing validator separately handles stored-answer secrets.

## Conditions, references and cycles

Nil condition is absent. A present condition with zero All clauses gets one
warning at the condition path. Plan-variable AskIf gets exactly one error at
`.askIf`, even if empty; do not validate or use its body for dependency edges.
This diagnoses authoring without changing finder's defensive evaluation.

A clause has one nonempty Question or Section, never both/neither. Section
clauses must have a zero AnswerTest; otherwise one clause error at the clause
path suppresses dependent checks. A section clause references an exact existing
section ID and tests its switch only, not activity/completion; self switches are
permitted and create no question edge. Unknown switch gives section-reference
at `.section`. A question clause uses the variable list at the containing position
for a bare name, even while reporting an invalid section ID such as `plan`.
Qualified IDs split at the first dot. An explicit `plan.NAME` always names actual
plan variables, never a section. Do not concatenate an invalid section ID and
reinterpret a bare reference as qualified syntax. Unknown
scope/name gives question-reference at `.question`. No default/answer is read.
An unknown or ambiguous question suppresses its target-dependent predicate
checks, not independent clause-shape diagnostics.

Section IncludeIf question references must be plan variables or strict ancestor
questions under 0070, never the section's own, siblings', descendants' or unrelated
questions. Bare names therefore do not supply a local IncludeIf escape. A known
out-of-scope question gives include-scope at `.question`, not an unknown error.
AskIf on section variables may reference local, plan or any other section's
questions, including later sections, subject to cycles. The narrower finder
exclusion sentence is not used to ban the documented cross-section use case.

Build no winner for duplicate section IDs or variable names. Still validate each
entry's independent declaration and clause shape. References to an ambiguous
target suppress reference/type/scope diagnostics and dependency edges; the
existing validator supplies the duplicate error. A duplicated enclosing section
ID makes every bare local question target ambiguous, even a missing local name:
suppress its lookup, predicate, scope and edge judgments, not independent clause
shape or explicit plan-qualified references. Duplicate section IDs suppress
inter-section question resolution, ancestor-scope and question-cycle analysis,
not plan-only references or independent local declaration checks. Switch checks
remain independent: preserve existence and finite contradictions for a unique
target even when another ID is duplicated; missing targets still error, while a
duplicated switch target suppresses dependent judgments without choosing a winner.
Missing/cyclic/ambiguous
Parent chains suppress include-scope judgment dependent on that chain; traversal
must terminate and existing structure diagnostics remain sole graph reports.
Self IncludeIf remains forbidden after unique target resolution even under a
missing or cyclic Parent chain: that known self relation needs no ancestor
traversal. Duplicate-target ambiguity still suppresses the dependent judgment.
An unrelated duplicate ID does not suppress this known bare-local self relation
in a uniquely identified containing section. Qualified nonplan references still
follow the earlier global resolution suppression, even when their prefix equals
the containing section ID. Do not infer a qualified self exception before lookup.
Question edges requiring unknown ancestor eligibility are suppressed. Duplicate
source variable declarations contribute no question graph node or unused check.

A question AnswerTest needs at least one set test. Is/IsNot are set when Kind is
not None; IsOneOf/IsNotOneOf are set when their slice is nonnil, including empty;
IsSettled is set when its pointer is nonnil. No tests gives predicate at `.answer`.
List-valued target questions permit only IsSettled. Any other test on such a
target gives one predicate error at `.answer`, suppressing value/list checks.
Unknown target type suppresses dependent value/list/type judgments, leaving its
own declaration error. Every explicit comparison value must be a settled scalar
of the target type,
exact choice membership, valid string pattern and correct directory namespace;
None, Deferred, empty Text, List and unknown kinds are invalid. Whitespace Text
and Flag false are settled. Each bad value gives predicate-value at its value
path. This does not validate stored answers or evaluate the condition.

Each supplied list test needs length >= 2 and cannot target boolean or cover
every distinct option of a valid choice. Emit one predicate-list at that list
field if any such rule fails, suppressing its element-value errors. Repeated
ordinary values are allowed; all-options compares sets. Independent Is/IsNot
tests still validate. Never broaden list comparison support defensively.

Report one contradiction per condition at the condition path for the following
finite catalogue, grouped by resolved question across all clauses: distinct Is
requirements; an Is excluded by IsNot/IsNotOneOf or absent from an IsOneOf;
intersection of positive IsOneOf sets emptied by other positive sets/exclusions;
both IsSettled true and false; IsSettled false together with any comparison;
or all finite boolean/valid-choice possibilities excluded. Conflicting required
switch states for the same Section also contradict. Compare exact scalar Kind
and payload; ignore inactive payload fields. Skip ill-formed, ambiguous or
invalid-predicate clauses for this analysis. Do not solve regex satisfiability,
derive transitive logical implications or compare conditions across fields.

Question eligibility dependencies are directed from a question to every resolved
question named by its AskIf and by its own/ancestor sections' IncludeIf. These
include inherited activity dependencies; switches add no edges. Find every
cyclic strongly connected component, including a self edge, over all declarations
independent of activity. Emit one condition-cycle at each participating variable
path; do not mark downstream noncycle questions. References with known targets
can contribute an edge despite an independent predicate error, but illegal
IncludeIf scope or malformed/ambiguous target cannot. Suppress cycle analysis
when duplicate section IDs prevent unique scope identity. No Parent graph error
is emitted here.

## Useful questions and properties

A declared onboardee variable is used when real unescaped interpolation in a
scanned section field, an exact escalation reference, or a section IncludeIf references
it, or when its answer is a necessary AskIf dependency chain of such a used
question. Count declared disabled sections too. Example/default/description,
held answers and self-only cycles are not uses. An escalation use is exactly
`$NAME` resolving to one local declaration, independent of its type validity.
Known wrong types still error; unknown types suppress dependent type judgments.
Malformed syntax, unknown and ambiguous escalation targets are not uses.
Emit one unused-onboardee error
per uniquely declared, legal-name/audience variable not reached by that closure.
A legal enclosing section ID is not required: continue independent use checking
at that indexed position, omit invalid identity metadata and count valid local
interpolation. Duplicated section scopes still suppress unused judgments.
Do not warn about unused onboarder declarations or filter the finder's output.
This rejects unsupported declarations under 0055 without changing finder output.
It does not replace a required-answer blocker or imply completion.

Input remains deeply unchanged, including nil/empty shapes, unknown maps, stored
answers, condition pointers, default lists and directory maps. Results share no
mutable map/list/pointer with input, dependencies or another call. Mutating a
returned Detail cannot affect input or later results. A shared Validator is
deterministic and safe for concurrent calls, with no call-history cache. Empty
plan gives no diagnostics. Paths follow input order, so permutation invariance
of indexed paths is not promised. Diagnostics never accept/repair content.

## Authority and bounded judgments

Sources: ADRs 0040 (blank prose/short dollars), 0045/0120 (suggestions), 0055
(audience/use), 0070 (ancestor IncludeIf), 0080 (directory), 0110 (ambiguity),
0260 (goals and graph separation); schema variables, conditions, steps,
escalation, reserved names and ordering; domain model; actual secrets, finder,
evaluator, resolver and plan-validator contracts. Fixture sources are separately
recorded in `testdata/sources.md`.

Bounded judge recommendations for this draft: explicit slug spelling; no new
required optional prose; nonblank directory display names; allowed empty/Deferred
suggestions; repeated ordinary options; literal owner prose with typed real
references; meaningful model-level checks; unrestricted documented cross-section
AskIf; finite contradictions; unused onboardee errors under 0055. P3
favors actionable authoring feedback, P2/P4 preserve exact data and deliberate
answers, P6 avoids date punishment, and 0110 forbids selecting ambiguous winners.
Invalid identifier metadata is omitted and environment paths are ranked to avoid
echoing rejected private text under P7. Secret messages preserve the dependency's
actual nonempty stable-per-code wording, not invented English literals. These are
proposed contract choices, not human approval or verified enforcement.

Known source discrepancies remain reported, not patched: 0040 assigns name/prose
checks to loader while its protected contract passes them through; ordering's
old plan-only IncludeIf row is superseded by 0070; evaluator's defensive list
behavior does not endorse list predicates; finder handles malformed plan AskIf
defensively while authoring rejects it. No protected existing contract changes.

## Wrong implementation distinctions

These are independently specified rejection oracles. Execution evidence belongs
in the working hardening review, not this contract.

| Wrong implementation | Independent catcher |
| -- | -- |
| Accept all or reject all declarations | `empty`, `valid-declarations`, `required-prose-slugs`, all literal invalid cases |
| Reuse old validator or traverse held answers/unknown data | `ignored-held-data`, `duplicate-and-graph-delegation` |
| Skip disabled sections, unknown enums or Unicode blank prose | `timing-all-declarations`, `variable-name-enums`, `required-prose-slugs` |
| Reject optional prose, empty step IDs, literal empty command arguments or repeats | `valid-declarations`, `required-prose-slugs`, `ordinary-repeated-set-values` |
| Treat default as an answer; collapse false/Deferred/empty suggestions; require full regex match | `false`, `deferred`, `empty-string`, `space-pattern-search`, input-ownership property |
| Guess directory namespace/name or normalize case | `contact-no-name-lookup`, `contact-no-namespace-fallback`, `choice-exact-case` |
| Apply options/pattern to irrelevant types or support multiple team | `valid-declarations`, `multiple-choice-pattern`, `team-list` |
| Scan wrong text fields, shorten placeholders, use plan-scope fallback or reject short dollar literals | `dollar-fields`, `short-escaped-longest-dollars` |
| Reject literal owner prose or accept wrong owner/escalation types | `owners-escalation` |
| Ignore check shape, trim env names or report malicious raw keys | `checks-environment`, `TestRejectedIdentifierPrivacy` |
| Allow plan AskIf, unknown references or malformed clauses | `condition-shapes` |
| Ban ancestor IncludeIf or cross-section AskIf; allow self/sibling IncludeIf | `ancestor-and-cross-ask-valid`, `include-scope` |
| Inspect activity instead of declarations or mark cycle tails as members | `inactive-cross-cycle`, `inherited-eligibility-cycle` |
| Pick duplicate winners or repeat delegated graph/holder diagnostics | `duplicate-and-graph-delegation`, `missing-parent-scope-suppression` |
| Allow unsupported list comparisons, singleton/boolean/all-choice lists or Deferred predicate values | `list-target-comparison`, `list-comparison-value`, `singleton-set`, `boolean-set`, `all-choice-options`, `deferred-test` |
| Miss cross-clause contradictions or confuse switch tests with activity | `positive-intersection-excluded`, `different-values-required`, `boolean-exhausted`, `switch-contradiction` |
| Permit useless onboardee declarations, count examples/defaults as use or ignore dependency closure | `use-closure` |
| Skip default/env secrets, skip disabled text, lose list item detail or echo a header | `secret-projection`, `TestSecretsMessagesStableAndPrivate` |
| Share diagnostic maps, mutate nested model data, retain call history or race | `TestInputOwnershipAndRepeatedCalls`, `TestSharedValidatorConcurrentCalls` |
| Hide short/escaped dollar defects behind another error in the same field, trim exact references or inspect inactive scalar payloads | `TestIndependentVariantGaps` |
| Lose negative-list finite exhaustion or disabled/inherited inclusion dependencies | `TestIndependentVariantGaps` |
| Guess a duplicate enclosing scope, hide independent invalid-ID purpose errors or reinterpret a bare name through an invalid plan section | `TestDifferentialGapRegressions` |
| Count only well-typed escalation as purpose, suppress known self inclusion under broken parents or hide unique-switch contradictions | `TestDifferentialGapRegressions` |
| Hide a unique bare self-inclusion error because an unrelated ancestor ID is duplicated, or guess a duplicated own scope | `TestSecondDifferentialGapRegressions` |
