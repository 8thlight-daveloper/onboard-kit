# Independent score fixture arithmetic

Sources: normative `contracts/invitation/reviewer/admission.md`, public model
structs, authoring validator contract, and the existing literal admission
fixtures. No reviewer implementation, resource calculator, or production output
was consulted. Tests are unexecuted; the root agent owns execution.

Baseline independently counted: source E869 consists of Plan101,
Preparation60, Seed134, Guide4 and ZIP overhead570. W(P)2841 consists of
Plan768, Provenance fields128, Origin fields256, Directory fields192,
Answers fields128, schema/ID payload9, Section1229, and pad key/value131.
Section1229 is19 slots of64 plus Bootstrap1 and ID/Title/Summary12.
W(seed R)974 is9 slots576, PlanReference fields192, Answers fields128,
schema payload8 and account/plan-ID/hash payload70. W(selected T)1177 is
pointer64, Preparation512, recipient fields576, schema8 and string payload17.
Thus Base is `869+64+2841+974+1177=5925`. Render is `1229+192=1421`;
PairWork32 makes `8*(5925+1421)+32=58800`.

The existing zero-variable admission fixture has payload score
`58800 + 152*d + 16*m`. The original base plan has101 source bytes.
The coefficient152 is8 times one source dollar, its Plan W occurrence, its
SectionView W occurrence, and the M16 expansion charge. An unknown pad byte
occurs once in source and once in Plan W, giving16.

Adding `variables:\n  - name: xxx\n` adds25 source bytes. Variable W is1102:
17 slots of64, ValueKind8, three booleans3, Name3. Candidate scope is `plan`,
identity length8, group138, default pool135, current/default charge64 each.
Candidate is `256+8+4+3+6+9+138+64+64+135+64+8=759`.
The Q-square pair increment is32. New intercept is
`58800 + 8*(25+1102+759) +32 =73920`.
`73920+152*440000+16*9684=67108864`.
One inert archive byte adds8 and refuses. This reaches absent-type/audience,
group, current, default, pool, default-source and ValueCharge-slot charges.

Adding a local declaration instead adds33 source bytes. Variable W is charged
twice,1102 in P and1102 in SectionView. Scope `task` also has length4, and the
two local gap slots add `2*(256+8+4+3)=542`, making Candidate1301. Intercept is
`58800+8*(33+2204+1301)+32=87136`. M becomes19, so the dollar coefficient176
is8 times `(1 source +1 P +1 view +19 expansion)`.
`87136+176*380000+16*8858=67108864`.
This additionally reaches the marker width and both local gap reservations.

Both declarations are default onboarder, making unused declarations legal.
Walker returns finder's full-plan empty-audience results, meaning both audiences;
finder returns every active declaration even when unused in prose. Therefore
these sessions contain exactly one onboarder question, plan.xxx or task.xxx.
The name `x` from the parent's provisional idea is invalid under authoring's
three-character minimum and is deliberately replaced with `xxx`.

Root options case adds17 source bytes for `    options: [a]\n`,65 W for its
one string element, and130 Candidate for both Options occurrences. Total1696
is106 pad bytes, so padding9578 preserves B. Options on string declarations
are ignored semantically but charged by resource policy. This targets195.

Command-list default adds38 source bytes. W grows137: Type7 and two List
elements of65 each. Selected default charge rises34 (two letters and two16
slots), counted both directly and in Pool. Type grows from implicit6 to7,
counted twice. Candidate delta70. Total `8*(38+137+70)=1960`; add one inert
source byte8 and remove123 pad bytes1968, giving padding9561. This targets188.

The hidden answer source block has46 bytes. W adds531: scope key70, inner
map64, name key67, and Value330. The selected Text value charge65 adds134
to the declaration's pool (`64+65+5`), without changing Current because this
answer is in hidden scope. Total `8*(46+531+134)=5688`; one inert byte brings
it to5696, removed as356 pad bytes, giving padding9328. This targets190/191.

The local hidden-list case adds67 source bytes. Answer W adds676: task key68,
inner map64, hid key67, Value329 plus two string elements74 each. Its name is
different from xxx, so it changes no candidate pool/current. M becomes21 from
the two10-byte items joined with one separator. Intercept increases5944 to93080,
and one inert byte gives93088. Thus
`93088+192*348000+16*12486=B`. This targets175.

The local contact case adds22 source bytes for Type and83 for Directory. W
adds14 for the two Type occurrences and406 for Contact: six64 slots plus
ID1/Name20/Slack1. Candidate grows87: Type+Group2, Choice85. Intercept
`87136+8*(105+420+87)=92032`. Contact maximum23
exceeds marker19, so coefficient208. Thus
`92032+208*320000+16*28552=B`. Targets170/172/177.

The local vacant-role case adds19 source Type bytes and65 Directory bytes.
W adds8 for the two Type occurrences and277 for Role (four64 slots plus21
string bytes). Candidate grows81: Type+Group minus4, Choice85. Intercept
`87136+8*(84+285+81)=90736`. Vacant maximum29
exceeds marker19, so coefficient256. Thus
`90736+256*260000+16*28633=B`. Targets171/177. The vacancy is a warning;
supporting an inactive empty role does not invent a selected answer.

Rendered carrier cases move all dollars out of body into a single held field.
A zero Step W is768 (struct plus11 fields); each Check W is256 (struct plus3
fields). Both are counted in Plan W and SectionView W.

Run source overhead29 and one string slot64 yield W1664 across the two copies.
Intercept `58800+8*(29+1664)=72344`, plus inert8 gives72352. The dollar
coefficient is664:8 times source1, P1, view1, text16 and run64. Thus
`72352+664*100000+16*39782=B`. This targets178.

Check Ask overhead47, Step plus Check W1024 copied twice2048 yield intercept
`58800+8*(47+2048)=75560`, plus inert8 gives75568. Thus
`75568+152*440000+16*9581=B`. This targets180.

Step environment overhead40 adds a65-byte key and64-byte empty value to Step,
so W is897, copied twice1794. Intercept
`58800+8*(40+1794)=73472`. Thus
`73472+152*440000+16*9712=B`. This targets181.

Eligibility case starts from root-variable intercept73920. The IncludeIf
source adds74 bytes. Nonnil Condition adds128 plus Clause1175, counted twice
as2606. Clause1175 comprises struct64, Question64, AnswerTest914, Section68,
Enabled65. AnswerTest914 is struct64, two Values329 each, two slices64 each,
and pointer64. Eligibility adds16 and clause-square adds8. Intercept
`73920+8*(74+2606)+24=95384`, plus inert8 gives95392. Thus
`95392+152*440000+16*8342=B`. This targets161, using the explicitly legal
self section-switch IncludeIf clause and no new behavior choice.

Escalation starts from the contact-directory boundary, retaining M23. Source
`    escalation: [$xxx]\n` is23 bytes; its one4-byte string has W68 counted
twice136. Escalation adds `64+4+23=91`. Delta `8*(23+136+91)=2000` is125
pad bytes; padding28427 preserves B. Targets183.

Waiting starts from zero-variable baseline. Requires source adds22 bytes and
the second section source42, total64. Requires element `other` costs69 twice.
Second Section W is1224 (19 slots64, Bootstrap1, ID5, Title1, Summary1),
counted twice2448. Its view metadata adds192, and Waiting adds69. S-square
changes by96. Intercept `58800+8*(64+138+2448+192+69)+96=82184`; inert8
gives82192. Thus `82192+152*440000+16*9167=B`. Targets184.

History adds an81-byte literal outcome block to134-byte record source, so
the present local record contributes215 to E. Its Outcome W is480: seven
slots64 plus section4, kind8 and timestamp20. This adds480 to Base and680
to Render including metadata200. Effective intercept is
`58800+8*(215+480+680)=69800`. One inert archive byte gives69808. Thus
`69808+152*440000+16*9941=B`. A trailing newline changes only source E and
must refuse effective admission. Targets186 while actual outcomes remain one.

Shape audit: public record decoder contract, Outcomes, maps serialized `outcome`
to model Outcome.Kind; `kind` is only the required top-level record discriminator.
The outcome block's source lines are10+18+22+31=81 bytes. Public plan decoder
contract, Known fields, makes section `requires` an ordered string list, and
the schema's ordering example uses `requires: [shadow-deploy]`. The escalation
schema defines an ordered list of local contact/role references and explicitly
shows `escalation: [$access_backup]`; the fixture's local contact xxx follows
that form. Public walker contract permits held outcome hashes to be empty,
yielding comparison unknown, and adds no malformed-outcome diagnostics.

The Flag rendering maximum mutation5 to4 is algebraically invisible: each
section starts M16, and every subsequent operation is max with a nonnegative
operand. Thus M is always at least16 and neither4 nor5 can increase it. This
is an equivalence candidate, not accepted-list authority or a verified reaching
probe. Root must apply and reach that mutation before recording evidence.
