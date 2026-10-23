# Record reconciler contract

Status: draft by an agent. Not reviewed by a person. Component: `record/reconciler`. Issue: none yet.

The record reconciler checks an onboarding record against a new plan. It removes the answers the plan no longer accepts, lists what changed between the previous plan and the new one, and flags current section definitions that differ from recorded outcome hashes. It is a pure function. It reads no file and no network, reads no clock, keeps no state between calls, and does not change an argument. It replaces the plan refresher. See [0090](../../../docs/decisions/0090-plan-and-onboarding-record.md), [0075](../../../docs/decisions/0075-refreshing-a-personal-plan.md), [0055](../../../docs/decisions/0055-variable-audience.md), [0080](../../../docs/decisions/0080-directory.md), [0035](../../../docs/decisions/0035-settled-and-deferred-answers.md), the [domain model](../../../docs/reference/domain-model.md), and [0015](../../../docs/decisions/0015-functional-core-and-imperative-shell.md).

The plan owns the content. The record holds only what the newcomer gave: answers to onboardee questions, outcomes, the account, the start date, and a reference to the plan. A reconcile never changes the plan, and it needs no comparison of three copies. Under [0270](../../../docs/decisions/0270-outcome-definition-hashes.md), it uses the [section hasher](../../section/hasher/contract.md) to compare actual current definitions. `New()` still takes no argument.

## Interface

`Reconcile(record, previous, plan, reference) -> (record, reconciliation, problems)`.

- `record` is the onboarding record.
- `previous` is the plan copy before the refresh. It may be the zero Plan when there is none. It is read only to compare questions and sections.
- `plan` is the new plan copy. It must have an `id`.
- `reference` carries the whole-plan `Hash` and `URL` the driver wants recorded. Its `ID` is ignored. The core computes no whole-plan hash; it computes current section definition hashes for outcome comparison only.
- `record` in the result is the new record.
- `reconciliation` is a `model.Reconciliation` with six lists of text: `NewQuestions`, `RemovedQuestions`, `AnswersToReview`, `NewSections`, `RemovedSections`, and `ChangedOutcomes`. The question lists hold question ids, `<scope>.<name>`, such as `plan.team` and `access.repo_url`. The section lists hold section ids.
- `problems` is empty when the reconcile ran. A refusal gives exactly one problem.

## Rules

### Refusing

A reconcile is refused in two cases.

1. `plan.id` is empty. The problem is `reconcile.missing-plan-id`.
2. `record.plan.id` is not empty and is not `plan.id`. This is another plan. The problem is `reconcile.different-plan`, with detail `expected`, the id in the record, and `found`, the id in the plan.

When the plan has no id, only the first problem is reported, whatever the record holds. A record that has never been reconciled has an empty `plan.id`, and the reconcile accepts it. The ids of `previous` are not checked.

A refusal returns the record as it was, copied, with its old plan reference, an empty reconciliation, and the one problem. Both problems have severity `error` and a message for a person, and leave `question` and `section` empty. The wording is not specified.

### The new record

| Part | Comes from |
| -- | -- |
| `plan` reference | `id` from `plan.id`, and `hash` and `url` from `reference`, as given. An empty `url` is written as empty. The old reference is not kept |
| `answers` | The rules below |
| `outcomes` | The record, as held, in the same order. None is added, removed, or changed |
| `schemaVersion`, `account`, `startDate`, `firstOpenedAt`, and `unknownFields` | The record, as held. Unknown fields keep their nested maps, lists, integers, and decimals |

FirstOpenedAt is preserved exactly on successful reconciliation and refusal, including empty or unusual held text. Reconciliation records no first-use event and infers nothing from `startDate`. See [0230](../../../docs/decisions/0230-first-use-record-compatibility.md).

### Questions and audience

A question is a variable of the plan, with id `plan.<name>`, or a variable of a section, with id `<section id>.<name>`. A plan declares the questions of its `variables` and of its sections. The audience of a question is the `audience` of its variable in `plan`, and it is `onboarder` when none is set. The audience in `previous` is never read. A variable with no type is a string. Two variables with one name in one scope are not covered.

### Answers

The result starts from the answers of the record. Each answer is looked up by its scope and name, `plan` or a section id, among the questions that `plan` declares.

| The answer's question | Result |
| -- | -- |
| Not declared by the plan | Kept exactly as it is. It is a hidden answer, and it is not checked or listed. That holds for a gone question, a gone section, a name that only another section declares, and a question no plan ever declared |
| Declared with audience onboarder | Left alone, not checked and not listed. The joiner puts the plan's answer under an onboarder question, so the record's answer there does not count |
| Declared with audience onboardee, and valid | Kept as it is |
| Declared with audience onboardee, and not valid | Removed, and the question is listed in `AnswersToReview`. A directory list that loses only some ids is shortened instead, and still listed. See below |

An answer the record has for a question the plan declares again after it was gone is an ordinary stored answer. When the question is onboardee, it is checked like any other. The plan's own answers never enter the record, and a new question appears unanswered, even when the plan pre-seeds it.

**Valid answers.** The same rule as the [plan refresher](../../plan/refresher/contract.md), as an independent contract.

- An answer that is not settled is valid for every type and kind. That is Deferred and, per [0035](../../../docs/decisions/0035-settled-and-deferred-answers.md), empty text, an empty list, and a list of only empty text. A flag is always settled, `false` too.
- A settled answer must be the kind the variable takes. That is a list for a `command` and for a variable with `multiple`, a flag for a `boolean`, and text for anything else. A `contact` or `role` takes text, or a list when `multiple`, so a text for a multiple one is not valid. A `team` takes text and its `multiple` is ignored.
- A `choice` takes text equal to one of its `options`, with case. With `multiple`, every item that is not empty must be an option, and one bad item makes the whole answer not valid. A choice with no options takes no settled text.
- A `contact`, `role`, or `team` answer is an id ([0080](../../../docs/decisions/0080-directory.md)) of the matching list in the directory of `plan`. A contact names a contact, a role a role, a team a team. The match is exact, with case, so a name or `Sam-Okafor` is not the id `sam-okafor`. The directory of `previous` is never read, and a plan with no directory lists no ids.
- A list of ids is checked item by item. An empty item is left alone. Each non-empty item that is not an id of the list is dropped, and the others keep their order. When some remain, the shortened list replaces the answer, and the question is listed in `AnswersToReview`. When no non-empty item remains, the answer is removed and listed.
- Nothing else is checked. A pattern, `required`, a `default`, and `options` on a variable that is not a choice are not checked.

**Where answers are written.** The result holds no entry for a removed answer. A section whose answers are all removed leaves no empty map. The result shares no map or list with an argument, including when the reconcile is refused. A record with no plan answer map gives a result with no plan answer map. A refusal returns a record equal to the argument in every part, and a map that is absent stays absent.

### Comparing the plans

When `previous` has no variables and no sections, it is the zero Plan for this purpose, whatever else it holds, and `NewQuestions`, `RemovedQuestions`, `NewSections`, and `RemovedSections` are empty, because there is nothing to compare with. A previous plan with variables and no sections, or with sections and no variables, is compared. The answers are still checked and the outcomes are still flagged.

### The reconciliation

Each list is sorted in ascending byte order of the whole text, with no repeats. A list with nothing in it may be nil or empty.

| List | Holds |
| -- | -- |
| `NewQuestions` | Questions that `plan` declares and `previous` does not, of any audience. A question in a new section is new |
| `RemovedQuestions` | Questions that `previous` declares and `plan` does not, whether or not the record has an answer. Their answers stay hidden. A question in a removed section is removed. A question that moved from one scope to another is removed from the old and new in the new |
| `AnswersToReview` | Onboardee questions of `plan` whose answer was removed, and directory lists that lost some ids and keep the rest |
| `NewSections` | Sections in `plan` and not in `previous` |
| `RemovedSections` | Sections in `previous` and not in `plan` |
| `ChangedOutcomes` | The ids of sections whose outcome has a nonempty `sectionHash` that differs from the actual section hasher result for the held section in `plan`. The comparison is exact, with case. Source may be absent, stale, or unrelated and supplies no fallback. An empty outcome hash and an outcome for a section the plan no longer has are not listed. A section with two differing outcomes is listed once. The outcome kind and the section's `enabled` setting do not exempt it |

### Definition comparisons preserve history

Hash the held current Section, before resolution or rendering, not the previous
section, its Source.Hash, rendered text, or the driver's whole-plan reference.
The section hasher's exclusions and canonical rules apply, including omission
of Source and Enabled. Source-only or switch-only changes do not change this
comparison. Other plan context, answers, and directory values are not part of
the section definition hash.

Under [0270](../../../docs/decisions/0270-outcome-definition-hashes.md), a listed
outcome means only that the current definition differs from the recorded hash.
Legacy bare hashes have no basis/version marker; a mismatch does not prove
that editing happened after verification. Compare arbitrary nonempty hashes
as held, without trimming, case folding, repair, or conversion. An empty hash
or removed section makes comparison unknown, not proven fresh. The reconciler
does not produce explanatory UI wording or evaluate prerequisite eligibility.

Neither agreement nor mismatch authorizes rewriting any outcome, its kind,
timestamp, SectionHash, Note, or Differently, or the original Source.Hash
template baseline. Keep all history and FirstOpenedAt exactly, on success and
refusal. Hash differences do not reset verification or create a freshness gate.
Unchecked duplicate outcome input is still defensive comparison data, not a
supported runtime declaration; the runtime validation boundary rejects it.
Hand-written canonical strings and independently pinned digests in
[outcome-hashes.json](testdata/outcome-hashes.json) are the comparison oracle.

## Properties

These hold for every case the contract tests use.

- **Pure.** The record, both plans, and the reference are the same after the call as before. Changing the result's answers, answer lists, outcomes, unknown fields, or reconciliation lists changes no argument.
- **No memory.** The same call gives the same result on a reconciler that has made other calls, in any order, and when calls run at the same time.
- **Idempotent.** Reconciling the result again with `plan` as both `previous` and `plan` gives the same record with only the plan reference replaced. The new reconciliation lists no new questions, removed questions, answers for review, new sections, or removed sections. Its `ChangedOutcomes` is the same as the first, because outcomes are never changed.
- **Refusal changes nothing.** A refused reconcile returns the record, an empty reconciliation, and its one problem.

## Side effects and plan access

None. It reads the record, both plans, and the reference, including actual current section definitions for hashing, and returns a copy of the record with `plan` and `answers` changed. It writes no plan field or outcome. This matches the row for `record/reconciler` in [components](../../../docs/architecture/components.md). Its allowed project dependencies are the model, section hasher, and its own contract under 0270.

## Not covered

- Fetching the plan, the plan copy, the hash, and the url. A driver does them and passes the reference in.
- Pruning hidden answers.
- Whether a plan or a record is valid, such as duplicate section ids or two variables with one name in one scope.
- Which ids a list lost. The reconciliation holds question ids only.
- The wording of messages.
- What happens when a variable is both `boolean` and `multiple`.
- Marking an outcome as stale in the record. The reconciler lists it, and the interface shows it.

## Decisions to know

The docs left these open. Each is a choice a person can reverse.

1. **`hard`. An answer under an onboarder question is left alone.** The record's answer there is not checked or listed, since the joiner puts the plan's answer under it. Raised by 0090: "An answer in the record that the new plan no longer accepts ... is removed and listed for review", with 0055 giving the onboarder the answer to an onboarder question.
1. **`hard`. A zero previous plan lists no new or removed questions or sections.** With no variables and no sections in `previous`, a first reconcile does not report everything as new. Raised by 0090: "New questions appear as unanswered, and new and removed sections are listed."
1. **`hard`. A multiple contact or role list that loses some ids keeps the rest and is listed for review.** A list that loses every id is removed. A multiple choice list with one bad item is removed whole. The same choices as the refresher. Raised by 0080: "An answer that names an id that is not in the directory counts as unanswered".
1. **`hard`. Ids are checked against the directory of the new plan only.** Raised by 0090: "a contact id that is gone, is removed and listed for review".
1. **`hard`. A hidden answer comes back into play when its question is declared again as onboardee.** It is then checked like any answer, so a stale one is removed and listed. Raised by 0090: "stays, as a hidden answer, until the owner prunes it".
1. **`hard`. Compare a nonempty outcome hash with the actual held section definition, including custom sections without Source.** Settled by 0270, not a source-presence inference. An outcome for a removed section remains unlisted and preserved; the section is in `RemovedSections` when `previous` had it. Unknown comparison is not evidence of freshness, and a mismatch is not proof of edit chronology.
1. **`hard`. A refusal gives one problem, the missing id first.** A record with no plan id is accepted as a first reconcile. Raised by the lead's decision on `reconcile.missing-plan-id` and `reconcile.different-plan`.
1. **`cheap`. A team answer is text even when `multiple` is set.** Raised by the person's decision that a team question takes one id.
1. **`cheap`. A multiple contact or role takes a list only.** A text there is not valid. Raised by the refresher fixtures, which the brief said to mirror.
1. **`cheap`. New and removed questions follow every audience.** Raised by the domain model: "Question ids the new plan declares and the previous plan did not".
1. **`cheap`. A moved question is removed and new.** Its answer stays hidden. Raised by 0075: "Ids ... are permanent. Renaming means a new id."
1. **`cheap`. An empty `url` in the reference is written as empty.** Raised by the brief: the reference comes from the driver.
1. **`cheap`. A record with no plan answer map gives a result with none.** Raised by the pure function rule: nothing is added that the record did not hold.
1. **`cheap`. A variable with no type is a string.** Raised by the domain model's default variable type.

## Wrong implementations tried

Each row names a plausible wrong implementation and the cases that fail it.

| Wrong implementation | Caught by |
| -- | -- |
| Drops, repairs, or invents firstOpenedAt during answer removal or refusal | `first-use-is-preserved`, `TestFirstOpenedAtSurvivesReconciliationAndRefusal` |
| The old url kept when the driver's url is empty, the old hash kept, or the plan id taken from the record | `reference-is-replaced-as-given`, `first-reconcile-is-accepted`, the reconcile-again test |
| A first reconcile refused, or a missing id accepted | `first-reconcile-is-accepted`, `refuse-a-plan-with-no-id` |
| Both problems reported, or the different-plan problem reported first | `missing-id-is-the-only-problem-when-both-apply` |
| A refusal that updates the reference, returns the record itself, or turns an absent map into an empty one | `refuse-a-different-plan`, the refusal test, the changing-the-result test |
| A different plan refused only when the record's id sorts after the plan's | `refuse-a-different-plan`, `refuse-a-different-plan-whose-id-sorts-before` |
| An onboarder answer checked, or an empty audience treated as onboardee | `onboarder-questions-are-left-alone`, `audience-is-read-from-the-new-plan` |
| The audience read from the previous plan | `audience-is-read-from-the-new-plan` |
| Hidden answers deleted, or an undeclared answer listed for review | `answers-without-a-question-are-hidden-and-kept` |
| A returning hidden answer not checked | `a-hidden-answer-is-checked-when-the-question-returns` |
| Deferred, empty text, an empty list, or a blank list judged not valid | `answers-that-are-not-settled-stay` |
| A false flag treated as not settled | `kind-that-does-not-fit-needs-review` |
| The kind not checked, or a flag, list, or text accepted for the wrong type | `kind-that-does-not-fit-needs-review`, `kind-fit-for-contact-role-and-team` |
| Options matched without case, or a choice with no options accepting text | `choice-not-in-options-needs-review` |
| Only the first item of a multiple choice list checked, or an empty item rejected | `multiple-choice-checks-every-item` |
| A pattern or the options of a string checked | `nothing-else-is-checked` |
| A team that takes a list when multiple | `kind-fit-for-contact-role-and-team` |
| Ids checked against the previous directory, any list, or without case | `ids-are-checked-against-the-new-plan-directory`, `id-must-be-in-the-list-for-its-type` |
| A list with a missing id removed whole, not listed, or with its empty items dropped | `multiple-list-drops-only-the-missing-ids` |
| A list of blanks kept after the drop | `multiple-list-drops-only-the-missing-ids`, `no-directory-in-the-plan-means-every-id-is-missing` |
| An answer checked against the variable of another scope | `scopes-are-separate` |
| A removed answer left as an entry of no kind, or a section left with an empty map | the no-empty-entries test, `a-removed-answer-leaves-no-entry-and-no-empty-section` |
| The sections after one that lost every answer skipped, or a result that depends on the order sections are read in | `many-sections-lose-all-their-answers`, the map-order test |
| A plan answer map made for a record that has none | the no-plan-answers test |
| The plan's answers copied into the record | `answers-of-the-plan-never-enter-the-record` |
| New questions only in the plan scope, or removed questions only when answered | `new-and-removed-questions-and-sections`, `answers-without-a-question-are-hidden-and-kept` |
| Lists sorted by scope then name, or not sorted | `new-and-removed-questions-and-sections`, the sorted-lists test |
| A zero previous plan that lists everything, is decided by the id, or is decided by either list being empty | `a-previous-plan-with-no-variables-and-no-sections-compares-nothing`, `a-previous-plan-with-variables-but-no-sections-is-compared`, `a-previous-plan-with-sections-but-no-variables-is-compared` |
| Source.Hash compared instead of actual content, or custom source-less sections skipped | `outcomes-are-kept-and-the-changed-ones-listed`, `manual-edit-baseline-unchanged`, `custom-without-source` in [outcome-hashes.json](testdata/outcome-hashes.json) |
| Reconstruct only old section fields, dropping Bootstrap or Requires before hashing | `bootstrap-and-requires-are-current-content` in [outcome-hashes.json](testdata/outcome-hashes.json) |
| Matching actual content flagged because Source or Enabled changed, or a hash trimmed or case folded | `actual-match-with-stale-source`, `source-only-change`, `switch-only-change`, `uppercase-is-not-repaired`, `whitespace-is-not-repaired` in [outcome-hashes.json](testdata/outcome-hashes.json) |
| Empty hashes or removed sections flagged, outcome kinds exempted, or history repaired on comparison/refusal | The outcome-definition tests, `a-removed-section-keeps-its-outcome`, the refusal tests |
| An outcome changed, dropped, or shared with the argument | `outcomes-are-kept-and-the-changed-ones-listed`, the changing-the-result test |
| Unknown fields dropped, shared, or copied only one level deep | `every-other-field-is-unchanged`, the changing-the-result test |
| Answer lists shared with an argument, or an argument changed | the arguments test, the changing-the-result test |
| State kept between calls | the same-result tests |
| A reconcile that is not idempotent | the reconcile-again test |
