# Remaining admission cap witnesses

Independent hand derivation from the frozen admission.md/model and public codec, preparation-validator and authoring-validator contracts. Neither reviewer implementation nor any design/production calculator was read. No code was executed or formatted by this author. Root owns compilation, formatting, actual outcomes and survivor replays. Source hashes only construct consistent invitation inputs through existing helpers; no semantic/score oracle comes from a component output.

## Successful public cap pairs

All cap values are inclusive. Positive controls below require complete successful Inspect or Review, not merely the absence of a specific resource code. Negative controls require one fixed invitation.resource at the named early/cheap stage and a zero operation result.

| Test/input | Independent count and nonmasking argument |
| -- | -- |
| Holder list1024/1025, both Role.Holders and a Team.Holders value | One known contact c, legal roles and optional known team. Repeated c is ordinary valid holder data. One list has exactly the target count; total remains below8192. No candidates or rendered directory variables. Each string item has W65, few fixed roots, score below2MiB, occurrences/entries below1100. Team overrides name declared roles; vacant organization roles only warn. |
| Combined holders8192/8193, role lists and team override lists separately | Partition into eight lists of1024, then a ninth list with one c only in the negative. Every individual list fits1024. At most nine roles/one contact/one team; all directory caps fit. Known entries below8300, forest occurrences below8500, no unknown fanout, score below8MiB. |
| Each Value.List4096/4097 | One command root variable with a repeated x list default; defaults remain suggestions. Nonempty repeated command arguments are valid, with no secrets. Total items below32768, one Q, no sections/conditions/directory. Default selected ValueCharge at4096 is69696; Base plus one candidate/default pool and source gives score below5MiB. |
| Combined Value.List32768/32769 | Eight distinct-name command declarations, each default list4096; negative adds a ninth declaration with one x. All individual lists fit. W of list strings is at most65*32769; candidate charge for each full4096-item declaration is about140040, including its own same-name default pool. All roots/source plus Questions keep score below35MiB. Occurrences below33000, known entries below33000, Q at most9. |
| Each options1024/1025 | One choice root declaration, repeated nonempty x options explicitly allowed by the authoring contract. No default/answer, so repeated options are not a duplicate identity or an all-options predicate. Options scalar charge and duplicated Candidate/Group options terms keep score below3MiB. |
| Combined options16384/16385 | Sixteen distinct-name choice declarations with1024 options each; negative adds a seventeenth with one x. Every individual list fits1024. Per-full-candidate Options=66560 appears twice, plus fixed terms. Score remains below32MiB, nodes/entries below17000, Q at most17, no conditions/Value.List items. |
| Known entries65536/65537 | Add one valid directory contact to literal minimal P. Known entries in P are Sections1+Variables1+Contacts1=3. Selected valid preparation has Arrangements1 plus K repeated valid helper entries. Therefore K65532 gives65536 and K65533 gives65537. No holder/Value.List/options/clauses caps apply to preparation helper lists. Entire forest has fewer than66000 occurrences and depth below10. Local source is below700000 bytes; strings have W65 each; complete effective score is below40MiB. Positive Review must preserve all65532 helpers and exact retained preparation bytes. Negative refuses review.preparation-admission/local.preparation before its validator. The same entry count remains valid after real joining because the identity seed has no answer entries. |

For these controls, declaration/default type rules, meaningful optional fields, exact contact namespaces and nonblank required prose are satisfied. Role vacancies are explicitly nonfatal warnings. Names are unique and short; repeated values are counted separately, never deduplicated for the policy. Every source stays below the existing envelope/local limits, and every unknown-data count/depth stays far below its unrelated cap. These are charged-score bounds, not RSS assertions.

## Cap acceptance followed by an explicit semantic refusal

These controls intentionally do not claim usable full-session positives. Their lower inputs must pass the targeted guard and reach the named later semantic phase; upper inputs must refuse earlier for resource admission. The public APIs make this distinction observable without a private estimator seam.

### Total W occurrences131072/131073

A successful full-score input at the occurrence ceiling is impossible under the published formula: W is at least64 per occurrence, so8*W is already B at131072, before the mandatory positive64 Base reservation, source or any other term. The occurrence cap remains independently useful in the explicitly score-free cheap phases.

Literal minimal P has59 W occurrences: zero Plan23, its Section19 and Variable17. Zero R has14. A nonnil optional Preparation contributes pointer1+zero Preparation17=18. Fixed forest is therefore91. Arrangement has nine fields plus its struct, hence10 occurrences.13098 zero arrangements add130980; one first-arrangement Contacts string adds1: total131072. A second string gives131073. Known entries are only2 plan entries+13098 arrangements+1/2 helpers, below13200; unknown occurrences are the four nil UnknownFields roots, and depth stays below10. Source is below100000 bytes.

The preparation has valid schema/kind, Account other and matching PlanID plan-a. Preparation validator phase3 explicitly says identity failure returns only identity problems and never inspects details. Account other is valid syntax but differs from expected jules, so at the inclusive occurrence ceiling the only later diagnostic is preparation.account-mismatch at account, wrapped Stage review.preparation-validation with its fixed Display. Its useful actual Message and any allowed extra Detail are not frozen. Zero arrangements' invalid IDs/sections/kinds/statuses never produce thousands of diagnostics because this identity phase precedes detail validation. At131073, review.preparation-admission/local.preparation must win instead.

### Condition and predicate caps

Use decoder-valid AskIf on root variables, which the actual authoring contract expressly rejects with exactly one authoring.plan-ask-if per declaration without examining its body. Keeping these held conditions on plan variables avoids counting every condition/value again in Render. The positive is the exact later inspect.plan-authoring refusal; it is not a supported plan or successful opening.

| Cap pair | Input and hand bound |
| -- | -- |
| Per Condition256/257 clauses | One root variable, one condition of repeated section-switch clauses. Total clauses below4096; Q1. Each Clause has18 occurrences and W1177 for Section absent. Score below5MiB and all other caps fit. |
| Clauses total4096/4097 | Sixteen root conditions with256 clauses each; negative adds a seventeenth with one clause. Each list fits256. Clause nodes below74000, known entries below4200. At the upper witness, clause W is4097*1177; EligibilityWork16*17*4097 and PairWork8*(16*256*256+1)+32*17*17 plus all sources/roots/questions keep score below56MiB. |
| IsOneOf256/257 or IsNotOneOf256/257 | One root AskIf with one question clause and one repeated Text x comparison list. Each Value has5 occurrences and W330. Combined comparison elements remain below16384; Value.List item count is zero. Score below2MiB. |
| Combined comparison elements16384/16385 | One root AskIf;64 clauses each hold256 Text x values, alternating IsOneOf and IsNotOneOf. Negative adds a65th clause with one value. Individual value lists stay within256; condition clauses65 stay within256 and global C within4096. Value nodes below82000 plus about1200 clause nodes, known entries below16500. Values contribute at most330*16385 W and clauses at most1177*65; Base/source, one candidate and work terms keep score below50MiB. No Render duplication or condition-value suggestion pool is invented. |

At positive sizes, authoring emits only the forbidden-plan-AskIf problem for each gate declaration. The exact repair is Remove askIf from the plan variable.; wrapper Display is Review the invitation plan declarations with its owner. A single variable yields variables[0].askIf/plan.gate000. For16 declarations, the authoring contract sorts bytewise PATH: the explicit declaration-index oracle is0,10,11,12,13,14,15,1,2,3,4,5,6,7,8,9. Both condition-body resolution and cycles are suppressed by the explicit plan-AskIf rule, so no nonexistent-switch or self-reference diagnostics are guessed.

## Coverage and limits

The proposed five test groups cover actual remaining targets136,138,139/140,141/142,144/145,148/149 and150/151/152 through public Inspect/Review only. They preserve the distinction between a complete positive and a guard-passing source that must later be refused semantically. No private helper tests, dependency injection seam, accepted-survivor entry or component/protected-file edit is introduced. Expected arithmetic is independently specified here; actual test/wrong-variant results remain root work.
