# Plan refresher contract

Status: draft by an agent. Not reviewed by a person. Component: `plan/refresher`. Issue: none yet.

The plan refresher brings a personal plan up to date with a newer org plan. It takes the cached org plan, the newly fetched org plan, the personal plan, and the origin to record. It returns the new personal plan, a report of what changed, and problems. It is a pure function. It reads no file and no network, reads no clock, keeps no state between calls, and does not change an argument. See [0075](../../../docs/decisions/0075-refreshing-a-personal-plan.md), [0080](../../../docs/decisions/0080-directory.md), [0055](../../../docs/decisions/0055-variable-audience.md), [0060](../../../docs/decisions/0060-personal-plan.md), [0035](../../../docs/decisions/0035-settled-and-deferred-answers.md), the [domain model](../../../docs/reference/domain-model.md), and [0015](../../../docs/decisions/0015-functional-core-and-imperative-shell.md).

The org plan owns the content, and the newcomer owns their own answers. A refresh therefore needs no conflict resolution. The org plan's content replaces the personal plan's content. For each question the audience says whose answer wins. The refresher uses the model and nothing else, so `New()` takes no argument.

## Interface

`Refresh(base, theirs, ours, origin) -> (plan, report, problems)`.

- `base` is the cached org plan, the one the personal plan was last refreshed from. The refresher reads only its answers.
- `theirs` is the newly fetched org plan. It must have an `id`.
- `ours` is the personal plan.
- `origin` is what the driver wants recorded: `Source`, `ContentHash`, and `RefreshedOn`. Its `PlanID` is ignored. The plan id always comes from `theirs.id`.
- `plan` is the new personal plan.
- `report` is a `Report`, defined in [api.go](api.go) as this component's output. It has seven lists of text: `NewQuestions`, `RemovedQuestions`, `AnswersTaken`, `AnswersKept`, `AnswersToReview`, `NewSections`, and `RemovedSections`. The question lists hold question ids, `<scope>.<name>`, such as `plan.team` and `access.repo_url`. The section lists hold section ids.
- `problems` is empty when the refresh ran. A refusal gives exactly one problem.

## Rules

### Refusing

A refresh is refused in two cases.

1. `theirs.id` is empty. The problem is `refresh.missing-plan-id`.
2. `ours.origin.planId` is not empty and is not `theirs.id`. This is another organization. The problem is `refresh.different-plan`, with detail `expected`, the id in ours, and `found`, the id in theirs.

When theirs has no id, only the first problem is reported, whatever ours holds. A personal plan that has never been refreshed has an empty `ours.origin.planId`, and the refresh accepts it. The ids of `base` and of `ours` itself are not checked.

A refusal returns `ours` as it was, an empty report, and the one problem. The origin is not updated. Both problems have severity `error` and a message for a person, and leave `question` and `section` empty. The wording is not specified.

### The new plan

| Part | Comes from |
| -- | -- |
| `directory`, `variables`, `sections`, and the plan's `unknownFields` | `theirs`, as held. The organization owns the directory, so the personal plan's directory is replaced as a whole. No entry of ours or base is kept or merged in, and a theirs with no directory leaves the result with none. A section's `source`, `enabled`, steps, unknown fields, and its place in the list come with the section. A hand edit to the content of ours does not last, and neither does a field of ours that theirs does not have |
| `schemaVersion`, `id`, and `provenance` | `ours`, as held. The personal plan's own id stays |
| `origin` | `planId` from `theirs.id`, and `source`, `contentHash`, and `refreshedOn` from the argument |
| `answers` | The rules below |

### Questions and audience

A question is a variable of the plan, with id `plan.<name>`, or a variable of a section, with id `<section id>.<name>`. A plan declares the questions of its `variables` and of its sections. The audience of a question is the `audience` of its variable in `theirs`, and it is `onboarder` when none is set. Ours and base are never asked for an audience. Two variables with one name in one scope are not covered.

### Answers

The result starts from the answers of `ours`. Each question that `theirs` declares is then decided, using three answers: `t` in theirs, `o` in ours, and `b` in base. An absent answer and an answer of no kind are the same, and mean unanswered.

| Audience | Case | Result |
| -- | -- | -- |
| Onboarder | Always | `t`. When theirs has no answer, the answer is removed from the plan |
| Onboardee | `o` equals `b` | `t`. The newcomer has not changed it, so a changed pre-seed reaches them. When theirs has no answer, the answer is removed |
| Onboardee | `o` differs from `b`, and `o` is absent | Stays unanswered. The newcomer cleared it |
| Onboardee | `o` differs from `b`, and `o` is valid | `o`, kept |
| Onboardee | `o` differs from `b`, and `o` is not valid | The answer is removed, and listed for review |

An answer from theirs is not checked. It is the organization's, and the plan validator judges it.

**Equal answers.** Two answers are equal when both are absent, or when they have the same kind and the same content. Text is compared as text. A Flag is compared as a flag. A list is compared item by item, in order, so `[a, b]` and `[b, a]` differ. The Text `true` differs from the Flag `true`. The Text `deferred` differs from Deferred. Empty text, an empty list, and the flag `false` each differ from absent. The base is read for answers only. With an empty base, every answer of the newcomer differs from it, so none is replaced.

**Valid answers.** Only an onboardee answer that is kept is checked, against the variable in `theirs`.

- An answer that is not settled is valid. That is Deferred, and per [0035](../../../docs/decisions/0035-settled-and-deferred-answers.md) empty text, an empty list, and a list of only empty text.
- A settled answer must be the kind the variable takes. That is a list for a `command` and for a variable with `multiple`, a flag for a `boolean`, and text for anything else.
- A `choice` takes text equal to one of its `options`, with case. With `multiple`, every item that is not empty must be an option.
- A `contact`, `role`, or `team` variable takes an id from the directory in `theirs`. See below.
- Nothing else is checked. A pattern is not checked, and `options` on a variable that is not a `choice` are not checked. The plan validator reports them.

**Directory answers.** A `contact`, `role`, or `team` answer stores ids ([0080](../../../docs/decisions/0080-directory.md)). Only a kept onboardee answer is checked, as for other types, and the check reads the directory of `theirs`. The directory of ours and of base are not read.

- Kind. A `contact` or `role` takes text, or a list when `multiple`. A `team` takes text, and its `multiple` is ignored. A flag fits none of them, and a list does not fit a single `contact`, `role`, or `team`. A wrong kind is not valid. An answer that is not settled is valid, as above.
- Id. Text must equal the id of an entry of the matching list: a `contact` names a contact, a `role` a role, a `team` a team. An id of another list is missing. The match is exact, with case, so a name or `Sam-Okafor` is not the id `sam-okafor`. Text that is not an id of that list is not valid.
- A list is checked item by item. An empty item is not an id and is left alone, as for the options of a `choice`. Each non-empty item that is not an id of the list is dropped and the others keep their order. When some remain, the shortened list replaces the answer. The question is listed in `AnswersToReview`, not in `AnswersKept`, and the report does not carry the dropped ids. When no non-empty item remains, the answer is removed and listed in `AnswersToReview`.
- An answer from theirs, an answer that the newcomer has not changed, and a hidden answer are not checked, so one that names an id the directory of theirs does not list is still taken or left as it is.

**Hidden answers.** An answer that `theirs` does not declare a question for is left exactly as it is in ours. That holds for a question that is gone, a section that is gone, and an answer that no plan ever declared. Such an answer is not deleted and is not checked. When theirs declares the question again, the answer is an ordinary stored answer, so an onboardee answer that differs from base is kept and an onboarder question takes theirs. An answer in theirs for a question theirs does not declare is not copied. A question that moved from one scope to another is not a move. The old question is gone and its answer is hidden, and the new question is new and starts from `t` as any new question does.

**Where answers are written.** The plan holds no entry for an absent answer. An entry that is removed leaves no entry of no kind behind, and the result may hold no empty map for a section. The result shares no map or list in `answers` with an argument, including when the refresh is refused.

### The report

Each list is sorted in ascending byte order of the whole text, with no repeats. A list with nothing in it may be nil or empty.

| List | Holds |
| -- | -- |
| `NewQuestions` | Questions that `theirs` declares and `ours` does not. A question in a new section is new |
| `RemovedQuestions` | Questions that `ours` declares and `theirs` does not, whether or not they have an answer. Their answers stay hidden |
| `AnswersTaken` | Questions whose answer in the result differs from the answer in ours because it came from theirs. That includes an answer that theirs removed. An answer that is the same is not listed |
| `AnswersKept` | Onboardee questions where the newcomer's answer differs from base, is present, is valid, and stays. It is listed even when theirs holds the same value |
| `AnswersToReview` | Onboardee questions where the newcomer's answer was removed because it is not valid, and `contact` or `role` lists that lost some ids and keep the rest |
| `NewSections` | Sections in `theirs` and not in `ours` |
| `RemovedSections` | Sections in `ours` and not in `theirs` |

A question appears in at most one of `AnswersTaken`, `AnswersKept`, and `AnswersToReview`. A question that is only hidden, or that is unchanged, appears in none of them.

## Properties

These hold for every plan the contract tests use.

- **Pure.** The three plans and the origin are the same after the call as before. Changing the result's answers changes no argument.
- **No memory.** The same call gives the same result on a refresher that has made other calls, in any order, and when calls run at the same time.
- **Idempotent.** Refreshing the result again, with `theirs` as both base and theirs, gives the same plan with only the origin replaced, and a report that lists no new questions, removed questions, answers taken, answers for review, new sections, or removed sections.
- **Refusal changes nothing.** A refused refresh returns ours, an empty report, and its one problem.

## Side effects and plan access

None. It reads all three plans and returns a copy of `ours` with `origin`, `directory`, `variables`, `answers`, the section content, and `source` changed. This matches the row for `plan/refresher` in [components](../../../docs/architecture/components.md). The dependency table in [0020](../../../docs/decisions/0020-binary-structure-and-regenerative-components.md) has no row for this component. It needs only the model.

## Not covered

- The progress of steps and checks, and the flag "changed since it was finished". The model has no progress record yet. Section hashes in `source` are copied from theirs and not compared.
- Whether a directory is valid, such as a duplicate id or a holder that is not a contact. The plan validator does that. A change to a contact applies when names are read.
- Naming the ids a list lost. The report holds question ids only.
- Fetching, the cached copy, the content hash, the id, and the date. A driver does them and passes the origin in.
- Pruning hidden answers.
- Whether a plan is valid, including duplicate section ids and two variables with one name in one scope. The plan validator does that.
- Updating an org plan from a newer template set.
- The wording of messages.

## Decisions to know

The docs left these open. Each is a choice a person can reverse.

01. **`hard`. The directory is replaced by the directory of theirs, and ids are checked against it.** A refresh never merges directories and never reads the directory of ours or base. A kept contact, role, or team answer that names an id theirs does not list is removed and listed for review. Raised by 0075: "Content: sections, variables, the directory: taken from `theirs`" and "a team that no longer exists: marked as needing review".
02. **`hard`. A multiple list that loses some ids keeps the rest and is listed for review.** The removed ids are not named in the report, since it holds question ids only. A list that loses every id is removed. Raised by 0080: "An answer that names an id that is not in the directory counts as unanswered".
03. **`hard`. A team answer is text even when `multiple` is set.** Raised by the person's decision that a team question takes one id.
04. **`hard`. The report has seven lists of ids and lives in `api.go`.** The model has no report type, and the brief did not allow one. A person may move `Report` into `internal/model`. Raised by 0075: "It returns a new personal plan and a report of what changed."
05. **`hard`. A refresh replaces a section's switch.** The `enabled` setting of a section comes from theirs with the rest of the section, so a newcomer's own switch does not last. Raised by 0075: "Content: sections, variables, the directory: taken from `theirs`", and 0050 calling the switch "the plan's own setting".
06. **`hard`. A kept answer that is no longer valid is removed and listed for review.** Valid means the right kind for the variable and, for a `choice`, a listed option. A pattern is not checked, and an answer that is not settled is always valid. Raised by 0075: "Marked as needing review and asked again. Never kept silently". The brief named options and kind only.
07. **`hard`. A hidden answer comes back when the question returns.** Answers of a gone question or section stay in the file. When theirs declares the question again, the answer takes part like any stored answer. Raised by 0075: "kept but hidden, not deleted, in case it returns".
08. **`hard`. An onboarder question that theirs does not answer is cleared in the personal plan.** The org wins, so removing an answer in the org plan removes it for everyone. Raised by 0075: "Taken from `theirs`. The org wins".
09. **`cheap`. Equality is by kind and content.** Empty text, an empty list, and `false` differ from absent. A list is equal only in the same order. Raised by 0075: "that the newcomer has not changed from `base`".
10. **`cheap`. A moved question is gone and new.** Its old answer stays hidden. The new question is not seeded from it. Raised by 0075: "Ids ... are permanent. Renaming means a new id."
11. **`cheap`. New and removed are judged against ours, not base.** Ours is what the newcomer has. Raised by 0075: "A new onboardee question: appears unanswered".
12. **`cheap`. Audience is read from theirs only.** A question that changes audience follows the new one at once. Raised by 0055: "The onboarder writes the question".
13. **`cheap`. With an empty or stale base, every newcomer answer counts as changed.** The first refresh cannot tell a pre-seed from a choice. Raised by 0075: "Without `base`, a pre-seed ... could not be told from a choice the newcomer made."
14. **`cheap`. A refusal with no theirs id reports only the missing id.** A refusal returns ours unchanged. Raised by the brief: refuse when `theirs.id` is empty.
15. **`cheap`. The plan's own `unknownFields` come from theirs.** `id`, `schemaVersion`, and `provenance` stay as ours. Raised by 0075: "Content ... unknown fields".
16. **`cheap`. `AnswersKept` lists a kept answer even when theirs holds the same value.** Raised by the brief: "answers kept".

## Wrong implementations tried

Each row names a plausible wrong implementation and the cases that fail it.

| Wrong implementation | Caught by |
| -- | -- |
| An onboardee answer always taken from theirs, or never taken | `onboardee-answer-changed-is-kept`, `onboardee-answer-not-changed-takes-theirs`, `equality-by-kind-and-order` |
| An onboarder answer treated as an onboardee one, so ours wins | `onboarder-takes-theirs`, `audience-comes-from-theirs` |
| The audience default is onboardee, or the audience is read from ours | `audience-defaults-to-onboarder`, `audience-comes-from-theirs` |
| Equality that ignores list order, or treats Deferred as the text `deferred`, a flag as text, `false` or empty text as absent | `equality-by-kind-and-order` |
| An absent base answer counted as unchanged | `onboardee-answer-changed-is-kept`, `no-base-answers-keeps-every-newcomer-answer` |
| An invalid answer kept, removed without a report, or reported and left in | `choice-no-longer-listed-needs-review`, `kind-no-longer-fits-needs-review` |
| Options matched without case, kind not checked, Deferred or an unsettled answer judged invalid, options checked on a string | `choice-no-longer-listed-needs-review`, `kind-no-longer-fits-needs-review`, `answers-that-are-not-checked-stay` |
| An answer taken from theirs is checked | `only-a-kept-answer-is-checked` |
| Hidden answers deleted, answers of a removed section dropped, or answers of theirs for an undeclared question copied | `removed-question-is-hidden-not-deleted`, `new-and-removed-sections`, `answers-with-no-question-in-theirs` |
| A returning question skipped | `returning-question-brings-its-hidden-answer-back` |
| New questions judged against base | `new-is-judged-against-ours` |
| Removed questions listed only when answered, or removed sections listed as new | `removed-question-is-hidden-not-deleted`, `new-and-removed-sections` |
| A question matched by name across scopes | `scopes-are-separate`, `moved-question-is-gone-and-new` |
| Lists sorted by scope then name, sorted without case, or not sorted | `report-lists-are-sorted-by-whole-id`, the sorted-lists test |
| Unchanged answers listed as taken, or kept only when theirs differs | `nothing-changed`, `onboardee-answer-changed-is-kept` |
| The plan id taken from the argument, the origin not updated or only partly, or the id of ours replaced | `first-refresh-and-the-origin-passed-in`, `later-refresh-replaces-the-origin`, `content-comes-from-theirs` |
| Provenance, variables, sections, the switch, or unknown fields taken from the wrong plan | `content-comes-from-theirs` |
| A missing id accepted, a first refresh refused, or both problems reported | `refuse-an-org-plan-with-no-id`, `missing-id-is-the-only-problem-when-both-apply`, `first-refresh-and-the-origin-passed-in` |
| A refusal that changes the plan, or has the wrong severity or detail | `refuse-a-different-organization`, the refusal test |
| Answers shared with an argument, a taken list shared with theirs, or a list answer that no decision touches shared with ours | The arguments test, the changing-the-result test, `a-taken-list-is-copied`, `hidden-list-answers-are-copied` |
| A removed answer left as an entry of no kind, or a section left with an empty map | The no-empty-entries test, `answers-removed-leave-no-entry-and-no-empty-section` |
| Only some items of a multiple choice list checked, the first item skipped, or an empty item beside valid options rejected | `multiple-choice-checks-every-item` |
| State kept between calls | The same-result tests, `new-question` |
| The directory kept from ours, taken from base, merged, or taken only when it has contacts | `directory-comes-from-theirs`, `directory-removed-by-theirs-is-removed`, `no-directory-in-theirs-means-every-id-is-missing` |
| Ids checked against the directory of ours or base, or the check skipped when theirs has no directory | `id-not-in-theirs-directory-needs-review`, `no-directory-in-theirs-means-every-id-is-missing`, `multiple-list-drops-only-the-missing-ids` |
| Any id accepted from any list, or the contact list used for roles | `id-must-be-in-the-list-for-its-type`, `id-not-in-theirs-directory-needs-review` |
| Ids matched without case | `id-must-be-in-the-list-for-its-type` |
| A missing single id kept, or removed without a listing | `id-not-in-theirs-directory-needs-review`, `id-must-be-in-the-list-for-its-type` |
| A list with a missing id removed whole, listed as kept, not listed, or sorted | `multiple-list-drops-only-the-missing-ids` |
| Empty items dropped as missing, or a list of blanks kept after the drop | `multiple-list-drops-only-the-missing-ids` |
| A team that takes a list when `multiple`, a multiple contact that takes text, a single contact that takes a list, or a flag accepted | `kind-fit-for-contact-role-and-team` |
| An answer of theirs, an unchanged answer, an onboarder answer, or a hidden answer checked against the directory | `only-a-kept-directory-answer-is-checked` |
| A refusal that takes the directory of theirs | `refusal-keeps-our-directory` |
| A refresh that is not idempotent | The refresh-again test |

## Bootstrap preservation

[0250](../../../docs/decisions/0250-bootstrap-content-and-projection.md) supplies
the source for the independently specified [preservation tests](bootstrap_test.go).
Bootstrap is onboarder-owned section content: a successful refresh takes the new organization's held marker, whether true or false. A refused refresh preserves the local marker exactly. Do not treat the marker as a newcomer answer or infer it from dates.
