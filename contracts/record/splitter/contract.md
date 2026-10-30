# Record splitter contract

Status: draft by an agent. Not reviewed by a person. Component: `record/splitter`.

The record splitter takes the newcomer's answers back out of the joined plan, so a driver can save the record and leave the plan copy alone. It is a pure function. It reads no file, writes no file, reads no clock, keeps no state between calls, and changes none of its arguments. See [0090](../../../docs/decisions/0090-plan-and-onboarding-record.md), [0075](../../../docs/decisions/0075-refreshing-a-personal-plan.md) on hidden answers, [0055](../../../docs/decisions/0055-variable-audience.md) on audience, the [domain model](../../../docs/reference/domain-model.md), and [0015](../../../docs/decisions/0015-functional-core-and-imperative-shell.md).

## Interface

`Split(plan, joined, previous) -> record`.

- `plan` is the plan as saved, with any pre-seeded answers.
- `joined` is the joined plan after the newcomer's edits. It is the plan with the record's answers put over the plan's answers, and then edited.
- `previous` is the record as last saved.
- The result is a new Record. There are no problems to report. Any three values give a record.

## The result

The result is `previous` with its `answers` recomputed. Every other field is copied from `previous` without change: `schemaVersion`, `account`, `startDate`, `firstOpenedAt`, the plan reference, `outcomes` in their order, and the unknown fields. Nothing but `answers` is read from `plan` or `joined`. Their `id`, `planUrl`, `resultsPath`, `origin`, and content never reach the record.

FirstOpenedAt is copied exactly, including empty or unusual text, while answers change. Splitting records no first-use event, infers nothing from `startDate`, and repairs no timestamp. See [0230](../../../docs/decisions/0230-first-use-record-compatibility.md).

## Which answers the record holds

Take each answer in `joined`, as a question id (scope and name, such as `plan.team` or `access.account_name`) and a value. The id is **declared** when the joined plan's variables name it: the plan's variables for scope `plan`, and the variables of the section with that id for a section scope. A name declared in another scope does not count. When a scope names one question more than once, the first declaration decides its audience, and the variables after the repeat are still read. A section id is any text, including empty text, and it names a section scope apart from the plan scope. An answer to a question that is not declared is a **hidden answer**.

The record holds the answer when one of these holds.

- **(a)** The question is declared with audience `onboardee`, and either the answer differs from the plan's answer for that question, or `previous` already held that key.
- **(b)** The question is not declared, and `previous` held that key.

Otherwise the answer is left out. In particular:

- **An onboarder answer is never in the record.** A question declared with audience `onboarder`, or with no audience, which means `onboarder`, is left out, even when `previous` held an answer for it.
- **An answer that follows the plan is not in the record.** A declared `onboardee` answer that equals the plan's pre-seed and was not held before is left out. It follows the plan, so a pre-seed the organization changes reaches the newcomer.
- **An answer the previous record held stays explicit.** It stays when it now equals the plan's pre-seed, so the newcomer's choice is not undone by a later change of the pre-seed.
- **A hidden answer stays only when `previous` held the key.** A hidden answer that `previous` did not hold is left out.
- **A cleared answer is left out.** An answer that is absent from `joined` is absent from the result, even when `previous` held it. A Value of kind none in `joined` or in `plan` counts as absent.
- **The value written is the one in `joined`.** It is never the value `previous` held. This holds for a hidden answer too.

### Differs

The answer differs from the plan's answer when the plan has no answer for that question, or when the two Values are not equal. Two Values are equal when they have the same kind and the same content.

- Text equals Text with the same characters. The text `deferred` and the value Deferred are different. `""` is a Text.
- Flag equals Flag with the same truth value.
- List equals List with the same items in the same order. An empty list equals an empty list.
- Deferred equals Deferred.
- Values of different kinds are never equal, even when both are empty: `""`, `false`, `[]`, and Deferred are four different values.
- A held `""`, `false`, or `[]` is an answer. It differs from a plan that has no answer.

### Absent stays absent

- When no answer is kept, `answers` has no `plan` map and no `sections` map. The result never holds an empty map that the answers did not need.
- A section whose answers are all left out is not in `sections`. When no section is left, `sections` is nil.
- `outcomes` and the unknown fields that `previous` holds as nil stay nil. So does a nil list or map inside the unknown fields, at any depth.

### Shares nothing

The result shares no map, list, or nested value with `plan`, `joined`, or `previous`. That covers the answer maps, the items of a list answer, the outcomes, and the nested maps and lists inside the unknown fields. Changing the result afterwards does not change an argument, and changing an argument afterwards does not change the result.

## The round trip

`Join(plan, Split(plan, joined, previous))` is the joined plan again, for the onboardee answers declared in the plan. The join is the one of 0090: the record's answers put over the plan's, key by key. This holds when `joined` declares the same questions as `plan`, as it does when the newcomer only changes answers. Exactly:

- **Equal.** For every question declared in `plan` with audience `onboardee`, the joined answer after the round trip equals the answer in `joined`. When `joined` has no answer for it, the round trip gives the plan's pre-seed, or no answer when there is none. The one case where that differs from `joined` is the limitation below.
- **Not equal.** The onboarder answers come from the plan: whatever `joined` held for them, the round trip gives the plan's. Hidden answers come back only when `previous` held them, with the value `joined` had.
- **Splitting again changes nothing.** `Split(plan, Join(plan, r), r)` has the same answers as `r`, where `r` is a split result.

### Known limitation

A newcomer who clears an answer that the plan pre-seeds cannot say so in a record. The record cannot hold "no answer", so the cleared answer is left out and the next load reads the pre-seed again. This is decision 1 below.

## Properties

These hold for every case the contract tests use.

- **Deterministic.** The same arguments give the same record, on a new splitter, on a reused one, in any order, and from several goroutines at once.
- **Inputs unchanged.** `plan`, `joined`, and `previous` are the same after the call as before.
- **Shares nothing.** See above.
- **Answers only.** Every field except `answers` equals the field of `previous`.
- **Nothing invented.** Every answer in the result is an answer in `joined`, with the same value.

## Not covered

- Writing the record as bytes, which the record encoder does, and reading one, which the record decoder does.
- Joining, which the record joiner does. The round trip above uses its rule.
- Checking that an answer is valid for its question, and listing hidden answers for review. The reconciler does those.
- A section whose id is `plan`. The contract does not say whether its answers are kept apart from the plan's own.
- Two sections with one id. The contract does not say which section's variables count.
- An audience that is neither `onboarder`, `onboardee`, nor empty.
- A Value of kind none in `previous`. The decoder never makes one.
- Whether an empty but non-nil list or map in `previous` stays empty or becomes nil.

## Decisions to know

1. **hard.** A newcomer cannot clear an answer that the plan pre-seeds. The record drops the cleared answer, and the join reads the pre-seed again at the next load. The alternative is to store Deferred, or a tombstone that means "cleared", when a pre-seeded answer is cleared. That would be an answer value, or a new kind of entry, in the record file, and so a change to the record format and the joiner. Raised by 0090: "An onboardee answer that the newcomer never gave is not in the record. It follows the plan".
2. **hard.** An answer the previous record held stays in the record when it becomes equal to the plan's pre-seed. The alternative is to drop it, which makes the record smaller and lets a later change of the pre-seed reach a newcomer who chose the old value. Raised by 0090: "`answers` holds only the answers the newcomer gave". A person sees the extra line in the record file.
3. **cheap.** An answer held for a question that is now declared with audience `onboarder` is dropped at the next save, and the newcomer's value is lost. The alternative is to keep it as a hidden answer. Raised by 0090: "It never holds an onboarder answer".
4. **cheap.** The audience of a question is read from `joined`, and `plan` supplies only the pre-seeded answers. Raised by the shared decision that the audience is the Variable's `audience` in the plan, when two plans are passed.
5. **cheap.** A hidden answer is written with the value from `joined`, not `previous`. They are the same after a join. Raised by 0075: hidden answers stay untouched.
6. **cheap.** A Value of kind none is treated as no answer. Raised by the model, where an empty Value has kind none and the encoder reports it.
7. **cheap.** `""`, `false`, and `[]` are answers, and equal to a plan answer only when the kinds and content match. 0035 calls `""` unanswered for a condition, but the record keeps what the newcomer typed.
8. **cheap.** Any three values give a record and no problems are reported. Mismatched plan ids and the like are the reconciler's concern.
9. **hard.** When a scope names one question twice, the first declaration decides the audience. The alternatives are that the last one decides, or that a repeat is an error. A plan author who repeats a name with two audiences sees the first one obeyed. The rules in 0090 say nothing about repeats.

## Wrong implementations tried

Each row is a plausible wrong implementation and the cases that fail it.

| Wrong implementation | Caught by |
| -- | -- |
| Drops, rewrites, or invents firstOpenedAt while saving an answer | `first-use-is-preserved`, `TestFirstOpenedAtSurvivesAnswerEdits` |
| Returns a zero record | `TestCases`, `TestAbsentStaysAbsent`, the round trip test |
| Returns `previous` unchanged | `answer-that-differs-from-the-plan-is-stored`, `cleared-answer-leaves-the-record`, the stored answers test |
| Stores every answer in `joined` | `answer-equal-to-the-plan-is-not-stored`, `equal-lists-are-not-stored`, `deferred-equal-to-deferred-is-not-stored` |
| An answer with no pre-seed counts as equal | `no-pre-seed-means-the-answer-differs`, `onboarder-answers-are-never-stored` |
| Drops a held key that now equals the plan | `held-key-stays-when-now-equal-to-the-plan` |
| Stores onboarder answers that differ from the plan | `onboarder-answers-are-never-stored`, `scopes-are-separate` |
| Keeps an onboarder answer the previous record held | `onboarder-answers-are-never-stored` |
| An empty audience read as `onboardee` | `onboarder-answers-are-never-stored` |
| Audience read from `plan` and not `joined` | `audience-is-read-from-the-joined-plan` |
| A question matched by name across scopes | `scopes-are-separate`, `a-plan-answer-and-a-section-answer-with-one-name-are-two-questions` |
| Drops every hidden answer | `hidden-answer-held-by-the-previous-record-is-kept`, `a-plan-with-no-questions-keeps-only-held-hidden-answers` |
| Keeps every hidden answer | `hidden-answer-the-previous-record-did-not-hold-is-dropped`, `scopes-are-separate` |
| Treats an undeclared answer as an onboardee question | `hidden-answer-the-previous-record-did-not-hold-is-dropped`, `scopes-are-separate` |
| A hidden answer takes the previous value | `hidden-answer-takes-the-joined-value` |
| A cleared answer kept from `previous` | `cleared-answer-leaves-the-record`, `cleared-hidden-answer-leaves-the-record`, the cleared pre-seed case |
| Equality ignores the kind | `kind-decides-equality-of-empty-values` |
| A list compared as a set, or by prefix | `list-equality-counts-order-and-items` |
| `false` and `""` ignored when the plan has no answer | `no-pre-seed-means-the-answer-differs`, `empty-text-and-false-are-answers-and-differ-from-no-answer` |
| A Value of kind none kept | `TestAnEmptyValueIsNoAnswer` |
| Empty maps invented, or an empty section kept | `a-section-whose-answers-all-leave-is-omitted`, `no-answers-gives-no-answers`, `TestNoEmptyMapIsInvented` |
| Starts from a zero record, or takes the plan id and url from `joined` | `every-other-field-is-copied-from-the-previous-record` |
| Drops the outcomes or the unknown fields | `every-other-field-is-copied-from-the-previous-record` |
| Shares the outcomes, a nested unknown field, or a list answer with an argument | `TestChangingTheResultDoesNotChangeTheArguments`, `TestChangingTheArgumentsDoesNotChangeTheResult` |
| Removes answers from `joined` in place | `TestArgumentsAreNotChanged`, `TestChangingTheResultDoesNotChangeTheArguments` |
| The last declaration of a repeated question decides, or a repeat stops the reading of later variables | `the-first-declaration-of-a-repeated-question-decides-the-audience` |
| A nil list or map inside the unknown fields becomes empty | `TestAbsentPartsInsideTheUnknownFieldsStayAbsent` |
| The plan scope read as the empty section id | `a-section-with-an-empty-id-is-a-section-scope` |
