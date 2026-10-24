# Invitation admission rules

This is the normative resource-policy part of the reviewer contract under0390.
No public estimator API or new dependency exists. Full admission applies the
cheap guards below and then the complete score. Cheap admission applies every
shape, occurrence, depth, collection and explicit cardinality guard, but no
Render/Questions/score calculation. Both read complete held data, including
disabled declarations, repetition, hidden answers and inactive payload fields.
Refusal never truncates, selects activities, accepts defaults or repairs data.

The promise is conservative pre-expansion admission, not a bound on heap, RSS,
regex/parser work, elapsed time or every implementation's returned-view bytes.
Group/message text is intentionally unspecified by some dependency contracts.
Weights are policy charges, not allocator sizes. Source decoding retains its
existing contracts; admission does not prove a parser memory bound.

## Context, phase and diagnostics

Use model Plan P, Record R, optional pointer to Preparation T and source byte
count E. These roots and their zero/nil slots always count as specified below.

| Stage | Kind and roots | E | Path |
| -- | -- | -- | -- |
| inspect.plan-admission | Full incoming P, zero R, nil T; immediately after plan decode, before authoring | Exact raw archive length | plan |
| inspect.payload-admission | Full incoming P, decoded seed R and incoming T; after both actual decodes, before seed semantics/locator/preparation validation | Exact raw archive length | invitation |
| previous.plan-admission | Full previous P, zero R, nil T; after previous decode, before either previous validator | Exact previous source length | previousPlan |
| review.preparation-admission | Cheap incoming P, zero R, selected T; immediately after retained preparation decode, before its validator | Not used | local.preparation |
| review.record-admission | Cheap incoming P, decoded existing R, selected T; immediately after record decode, before identity/ownership/history/reconcile/join | Not used | local.record |
| review.effective-admission | Full actual joined effective P, actual proposed R, selected T; after real reconcile/join, before effective validation/walker | Raw archive plus each present local body once | effectivePlan |

Do not move the old codec/semantic phases otherwise. Incoming payload admission
can therefore refuse before an independently invalid seed is reported. Missing
local preparation uses already admitted incoming T; no retained-preparation
decode/cheap phase occurs. Missing local record uses already admitted seed and
needs no existing-record decode/cheap phase. Later full effective admission
still checks every selected root. Cheap forests share counters across P/R/T;
do not reset counters between roots or omit incoming P. Zero R is an actual
zero Record root, not absence of that root. T is nonnil whenever preparation
has been selected. Begin repeats all the same source-derived phases.

All full/cheap failures except previous use code invitation.resource, severity
error, exact Message and Display `Use a smaller onboarding plan or fewer repeated values.`,
empty Question/Section, and Detail exactly path from this table. Return zero
operation result, with just this one phase diagnostic. No cap name, rejected
value, calculated score or guessed heap usage is exposed. Previous failure
instead uses invitation.previous-resource, severity warning, exact Message and
Display `Previous plan comparison is unavailable within the resource limit.`,
empty identifiers and Detail exactly `{"path":"previousPlan"}`. It stops only
previous validation, sets Comparison.Available false/Reason resource, and enters
Comparison.Problems after any successful preceding previous decode warnings.
It does not bypass explicit plan choice or erase the previous byte preimage.

E counts raw archive bytes plus each present local plan/preparation/record body
once, even when identical to another source. Never add decoded archive payloads
again. Existing raw2MiB archive/decompressed1MiB/local1MiB bounds remain earlier
guards. E can reach5MiB. Previous admission uses only the previous file length.

## Saturation

B=67108864 and X=67108865. All nonnegative sums/products saturate at X. Zero
times anything is zero; otherwise check `a > X/b` before multiplying. An operand
already at X stays exceeded under positive multiplication. Check count/length
conversions. No overflow, floats, rune lengths or allocating predicted products.
An empty sum is zero. B is admitted; X refuses. Every additive overlap below is
intentional; subtract nothing unless explicitly stated.

## W: exact model occurrences

Follow every declared field, zero/nil fields and inactive Value payloads included.
Each struct, scalar, map, slice or pointer occurrence costs64. Add string byte
length; bool payload1; known integer/ValueKind, int64 or float64 payload8. A
struct adds W of each field; a slice adds W of every element; a map adds W of
each key and value once. No extra map-entry slot is charged. A nonnil pointer
adds W of its target; a nil pointer adds nothing beyond its64. Shared values
are counted once per occurrence, never deduplicated by identity/backing storage.

Interfaces are transparent. A nonnil interface contributes only its dynamic
value, no wrapper slot; a nil interface is a null occurrence costing64. Unknown
data is only string, bool, null, int64, float64, unknown-value lists or
string-keyed unknown maps as the actual codecs supply. Unsupported shapes
refuse rather than stringify or follow an unbounded/cyclic structure. Nil/empty
maps and slices have equal W; semantic equality still distinguishes their
nested shapes under the main contract.

These are the complete frozen struct/field enumerations; follow field types
from the public model. No derived Question/View/Reconciliation is an input to W.
A future public field needs explicit policy review, not automatic reflected
policy extension:

| Struct | Fields |
| -- | -- |
| Plan | SchemaVersion, ID, Provenance, Origin, URL, ResultsPath, Directory, Variables, Answers, Sections, UnknownFields |
| Provenance | TemplateSet, GeneratedOn |
| Origin | PlanID, Source, ContentHash, RefreshedOn |
| Directory | Contacts, Roles, Teams |
| Contact | ID, Name, Email, Slack, SlackID |
| Role | ID, Name, Holders |
| Team | ID, Name, Contact, Holders |
| Answers | Plan, Sections |
| Variable | Name, Description, Example, Type, Options, Pattern, PatternHint, Required, Multiple, Default, Audience, AskIf |
| Value | Kind, Text, Flag, List |
| Condition | All |
| Clause | Question, Answer, Section, Enabled |
| AnswerTest | Is, IsNot, IsOneOf, IsNotOneOf, IsSettled |
| Section | ID, Title, Summary, Parent, Bootstrap, Owner, Timing, Enabled, Verification, Body, IncludeIf, After, Requires, Escalation, Variables, Steps, Source, UnknownFields |
| Source | Template, Hash |
| Step | ID, Name, Notes, Owner, Timing, Run, Open, Directory, Env, Escalation, Checks |
| Check | Run, Ask, Env |
| Record | SchemaVersion, Account, StartDate, FirstOpenedAt, Plan, Answers, Outcomes, UnknownFields |
| PlanReference | ID, Hash, URL |
| Outcome | Section, Kind, On, SectionHash, Note, Differently |
| Preparation | SchemaVersion, Kind, Account, PlanID, Recipient, Arrangements, UnknownFields |
| PreparationRecipient | Name, CompanyEmail, PlannedStartDate, Team, Manager, OnboardingOwner, Buddy, WorkLocation, TimeZone |
| Arrangement | ID, Section, Kind, Owner, Contacts, Reference, TargetDate, Status, Note |

Whole-forest root depth is1 for each P/R/optional-T occurrence. Every followed
field, element, map key/value or pointer target increments depth1; transparent
interface unwrapping does not. Cap depth128 and total W occurrences131072.
Count each UnknownFields map root and every key/container/scalar/null inside it
toward at most65536 unknown occurrences, across all roots; these also count in
whole-forest occurrences/depth. Stop at an exceeded cap, without exposing a key.

Known collection entries total every known-model map/slice length, including
nested Answers maps, directory lists/holder maps/holder lists, Value.List,
conditions/predicate lists, variables/options, steps/checks/run/env/escalation/
after/requires, outcomes, arrangements/contacts. A parent slice element counts
as one entry; nested slice elements count again in their own collection. Do not
count unknown collection entries here; use their separate node count. Cap65536.

## Other inclusive caps

| Counter | Maximum |
| -- | -- |
| P.Sections length S |256 |
| All plan/section Variable occurrences Q |1024 |
| Steps / checks across P |4096 /8192 |
| All clauses C in every held IncludeIf/AskIf |4096 total;256 per Condition |
| Each IsOneOf/IsNotOneOf |256 elements each;16384 combined Value elements |
| Contacts / roles / teams |2048 each;4096 combined |
| Each Role.Holders or Team.Holders value |1024;8192 combined items |
| Each Value.List |4096;32768 combined items |
| Each Variable.Options |1024;16384 combined items |
| R.Outcomes |4096 |

Value-list caps count every Value occurrence, including inactive payloads in
defaults/answers/condition Is/IsNot/IsOneOf/IsNotOneOf. Conditions include only
held IncludeIf/AskIf values in P, not imagined inherited copies. Repetitions
count; empty conditions have zero clauses but still retain their W. All these
guards precede score calculation. No separate preparation-arrangement cap exists.

## Directory and rendering terms

For list L, `J(L,k)=sum(len(item))+k*max(0,len(L)-1)`, preserving every item.
Compute complete-directory maxima, zero over an empty set:

```text
Contact=max(len(Name)+2+len(Email)+len(Slack)) over contacts
Team=max(len(Name)) over teams
Vacant=max(len(Name)+9) over roles
H=max holder-list length across roles and every team override
Role=max(Vacant,H*Contact+2*max(0,H-1))
Directory=max(Contact,Team,Role)
```

For each section occurrence s, independently start M=16 and include
`16+len(v.Name)` for every local declaration in its maximum. Read all held
P.Answers.Sections[s.ID] entries including hidden names: include Text length,
Flag bound5 or selected List J(List,1). None/Deferred/unknown Kind add nothing.
N is maximum selected List length there, or zero. Do not read defaults/other
scopes. Duplicate section occurrences each charge their exact-ID answer map.
If any local declaration's exact Type is contact/role/team, include Directory
in M, and if N>0 include `N*Directory+2*(N-1)`. This is finite type-label charging,
not ID lookup, team selection or membership validation.

D(f) counts every ASCII dollar byte, escaped/short/literal dollars included.
F(s) is exactly: section Title/Summary/Owner/Verification/Body; each step
Name/Notes/Owner/Open/Directory, Run items and Env values; each check Run item
and Ask. Check.Env is held data charged in W, not a rendered field.

```text
TextExtra(s)=sum(D(f)*M) over F(s)
RunExtra(s)=64*sum(D(f)*max(1,N)) over step/check Run items only
Escalation(s)=sum(64+len(e)+D(e)*M) over section Escalation items
Waiting(s)=sum(64+len(id)) over s.Requires, keeping repeats
SectionView(s)=W(s)+TextExtra(s)+RunExtra(s)+Escalation(s)+Waiting(s)+192
Render=sum(SectionView(s))+sum(W(outcome)+200) over R.Outcomes
```

W(s) already includes original fields/run slots/escalation/requires; add extras
without subtracting. No rendered step escalation is invented. Waiting charges
all direct Requires occurrences, without deduplication or interpreting outcomes.
192/200 are policy metadata weights, not measured view storage.

## Candidate and suggestion terms

ValueCharge starts64 and adds only selected payload: Text length,5 for Flag,
or `J(List,0)+16*len(List)` for List. None/Deferred/unknown Kind add zero. Unlike
W, this selected-payload term does not charge inactive fields.

Candidates are every plan declaration with literal scope plan, then every
section declaration with held scope s.ID. No activity/audience/order filtering.
Construct bounded name/length tables, never predicted Cartesian products.

```text
Pool[n]=sum(64+ValueCharge(v.Default)+7) over declarations with exact Name n
       +sum(64+ValueCharge(answer)+5) over all held P.Answers entries named n
```

Given entries include plan and all section scopes, hidden/removed scopes too.
Count each entry once. Do not add R.Answers again: actual join already supplied
effective P. R still has its own independent W in Base. No condition values in
Pool. Even absent defaults count; do not deduplicate/filter/group-match. Source
enum byte lengths7/5 are default/given, not invented source question/scope data.

For candidate q, scope/name are exact held strings. Let I=len(scope)+1+len(name),
typeBytes=6 for empty Type else its length, audienceBytes=9 for empty Audience
else its length, and Options=J(q.Options,0)+64*len(q.Options), even for types whose
real finder suppresses options. Current is P.Answers.Plan[name] when scope is
literal plan; otherwise P.Answers.Sections[scope][name]; absent is zero Value.
Resource charging does not make a reserved section ID semantically valid.

```text
Choice[type]=sum(64+len(ID)+len(Name)) over matching complete directory list
GroupCharge=128+len(name)+typeBytes+Options+1
Candidate(q)=256+I+len(scope)+len(name)+len(Description)+len(Example)
            +typeBytes+audienceBytes+Options+GroupCharge
            +ValueCharge(Current)+ValueCharge(Default)+Choice[Type]+Pool[name]
            +64+I
            +(0 for scope plan; else 2*(256+I+len(scope)+len(name)))
Questions=sum(Candidate(q))
```

Choice is zero except exact contact/role/team. Final terms charge a blocker for
every candidate and two deduplicated-gap slots per local declaration, not per
dollar/carrier. GroupCharge is a fixed tuple-size surrogate, not an upper bound
on the actual finder's intentionally unspecified Group text.

## Final score

```text
Base=E+64+W(P)+W(R)+W(optional T)
EligibilityWork=16*Q*C
PairWork=32*(Q*Q+S*S)+8*sum(len(condition.All)^2)
Score=8*(Base+Render+Questions)+EligibilityWork+PairWork
```

The64 event reservation is always charged, including existing records. Existing
FirstOpenedAt's string slot is already part of W(R). Declarations/conditions
come from P only. R/T add their full W and shared cheap guards, not declarations.
Map order cannot change saturated totals. No estimator is allowed to call a
renderer/finder/validator to estimate work. Admitted inputs still call the real
semantic dependencies unchanged. All resource defaults/activity false positives
remain charges only; no authored/stored value is silently changed.
