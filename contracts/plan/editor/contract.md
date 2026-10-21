# Plan editor contract

Status: draft by an agent. Not reviewed by a person. Component: `plan/editor`. Issue: [#70](https://github.com/8thlight-daveloper/onboard-kit/issues/70).

The plan editor changes a plan. It is the only component that returns a changed copy of one. `Submit` applies one answer. `SetSectionEnabled` turns one section on or off. `SetContact`, `SetRole`, and `SetTeam` add or replace an entry of the directory. `RemoveContact`, `RemoveRole`, and `RemoveTeam` remove one and clear every reference to it. Each returns a new plan and a list of problems, and none alters the plan it is given. All are pure functions. They read no file, read no clock, and keep no state between calls. See [0010](../../../docs/decisions/0010-plan-editor.md), [0015](../../../docs/decisions/0015-functional-core-and-imperative-shell.md), [0035](../../../docs/decisions/0035-settled-and-deferred-answers.md), [0045](../../../docs/decisions/0045-defaults-are-suggestions.md), [0050](../../../docs/decisions/0050-section-switch.md), [0080](../../../docs/decisions/0080-directory.md), and the [domain model](../../../docs/reference/domain-model.md).

The finder says which questions are active, and the plan validator says whether a value is acceptable ([question finder contract](../../question/finder/contract.md), [plan validator contract](../validator/contract.md)). The editor decides what to do with their answer.

## Interface

`Submit(plan, question, value) -> (plan, list of Problem)`, `SetSectionEnabled(plan, section, enabled) -> (plan, list of Problem)`, `SetContact(plan, contact)`, `SetRole(plan, role)`, and `SetTeam(plan, team)`, each `-> (plan, list of Problem)`, and `RemoveContact(plan, id)`, `RemoveRole(plan, id)`, and `RemoveTeam(plan, id)`, each `-> (plan, Cleared, list of Problem)`. The signatures are in `api.go`, with `Cleared`, which the model holds.

- `plan` is a Plan, as stored. `Submit` reads `plan.Variables`, `plan.Answers`, `plan.Sections`, and, through the validator, `plan.Directory`. `SetSectionEnabled` reads `plan.Sections`. The directory operations read `plan.Directory`, and the removals also read `plan.Variables`, `plan.Sections` (for each section's variables), and `plan.Answers`. No operation reads another field, and each returns the other fields as they are.
- `question` is a question id, `<scope>.<name>`, such as `plan.team_state` or `access.repo_url`. The scope `plan` means the plan's own variables. Any other scope is a section id. The id is split at the first dot. Any text is accepted. An id that names nothing gives a problem.
- `value` is a Value, as a person or a driver would submit it. A value of no kind is allowed. It means unanswered.
- `section` is a section id. `enabled` is true or false.
- `contact`, `role`, and `team` are directory entries as the model holds them. `id` is the id of an entry. `Cleared` has two lists of text, `Answers` and `Holders`, described under the removals.
- Every result is a plan and a list of problems, and the removals add a `Cleared`. The list is empty when there is nothing to report. There is no error return.
- The editor takes the finder and the validator through `New()` with no argument, as the validator contract does. The contract tests therefore run the real components, and the expected problems rest on the finder and validator contracts and use only what those fix.

## Rules for Submit

**Judging the value.** The editor asks the validator to check the value for the question, against the plan as it was given, and returns every problem the check reports, as they are and in the same order. The editor adds no problem of its own to `Submit`. The checks and their codes are in the [plan validator contract](../validator/contract.md): an unknown question, a question that is declared and not active, a value of the wrong kind, a value outside the options, a pattern mismatch, and secrets. A question is active when the finder returns it for the plan as given, so an answer that reveals a follow-up question does not make that question active in the same call.

**Errors and warnings.**

- When any problem has severity `error`, the answer is not applied. The returned plan equals the plan given, in everything it holds. The problems come back with it, warnings included.
- Otherwise the answer is applied, and any warnings come back with the new plan. A warning never stops an answer.

**Applying a value.** The editor puts the value in the plan's `answers`. A question in scope `plan` is stored under `answers.plan` and a question in a section under `answers.sections.<section id>`, each by variable name. The same name in two scopes is two answers.

- A settled value replaces the stored answer, or adds one. It is stored exactly as given: text as written, with spaces and with the text `deferred` as plain text, a flag, and a list in the order given, including any empty item in it.
- A Deferred value is applied like any other, and it replaces what was stored. A stored Deferred is replaced by a settled value in the same way.
- An unanswered value removes the stored answer, if there is one. Unanswered means a value of no kind, empty text, an empty list, and a list of only empty text, for a question of any type. A question with no stored answer is left as it is. The editor never stores an unanswered value. Text of only spaces is text, and it is stored.
- Every other stored answer stays: the other questions of the same scope, and every answer of every other scope, active or not. Applying an answer does not remove the stored answer of a question that the new answer hides, as [0010](../../../docs/decisions/0010-plan-editor.md) says.
- The plan differs from the argument in `answers` only. Whether a section's map is left in `answers.sections` when its last answer is removed is not specified.

## Rules for SetSectionEnabled

The section switch is a field of the plan section and not an answer ([0050](../../../docs/decisions/0050-section-switch.md)). The editor offers it as its own operation.

- When `section` is the id of a section the plan holds, the returned plan is the argument with that section's `enabled` set, and the result has no problem.
- Turning a section **off** stores `enabled: false` on it. Turning a section **on** removes the field, so that the plan records only the sections turned off. This holds when the section already has an explicit `enabled: true` written by hand: it is removed.
- A section already in the requested state stays in it. Turning on a section that has no field, and turning off a section that is off, change nothing.
- Nothing else changes. The answers of every section stay, including those of the section turned off, and so do the other sections, their order, and every other field of the plan.
- Only the named section changes. A switch is the section's own field, so turning a section off stores `enabled: false` on that section and writes nothing on its sub-sections, and turning a section on removes that section's own field and nothing on its sub-sections. A sub-section keeps its own `enabled`, a hand-written `true` included, and it keeps its `parent`. That a sub-section is inactive while its parent is off is the selector's rule and not stored, see [0070](../../../docs/decisions/0070-sub-sections.md). A sub-section can be turned off alone.
- A section id that the plan does not hold, an empty id, and the id `plan` give one problem, `switch.unknown-section`, with severity `error`, and the returned plan equals the plan given. The problem's `section` is the id as given, and its `question` is empty.
- The editor does not report sections that depend on the switch. A question of a section that is off is not active, so a later `Submit` for it is refused with `answer.inactive-question`.

## Rules for the directory operations

The directory has contacts, roles, and teams ([0080](../../../docs/decisions/0080-directory.md)). An id is unique within its kind and not across kinds, so a contact, a role, and a team can share an id. Every operation names its kind and looks only at that kind.

**The three Set operations.** `SetContact`, `SetRole`, and `SetTeam` store the entry they are given.

- An entry whose id no entry of that kind has is added at the end of that kind's list. An entry whose id an entry of that kind has replaces that entry whole and in the same place. A field the new entry leaves out is gone, so a replaced contact with no `email` has none, and a replaced role or team keeps none of its old holders, and a replaced team keeps no `contact` that the new team leaves out.
- The entry is stored as given: the editor does not trim, sort, or deduplicate it, and a holder named twice stays twice. A team override with an empty list is stored as an empty list. It means the role is vacant for that team, and it differs from no override, where the role's own holders apply, as [0080](../../../docs/decisions/0080-directory.md) says in "A role resolves in this order".
- Nothing else changes. Every answer, every section, and every other directory entry stays, and so do the holders and overrides that name the replaced entry's id, because the id does not change. A Set never clears an answer and never reports a warning.
- The editor judges the entry alone, against the directory as given. It reports the faults below. An error of this kind refuses the edit: the returned plan equals the plan given and the problems come back. Faults the validator reports as warnings, such as an email with no `@` or a contact with no way to reach them, are not the editor's.
  - An id that is empty or does not match `[a-z0-9][a-z0-9-]*` from its first to its last character gives `directory.id-malformed`. The pattern is whole, so an id with a trailing line break does not match. A single character matches, a digit may come first, and a hyphen may come last.
  - An empty name gives `directory.name-empty`. A name of only spaces is not covered.
  - Each holder of a role that is not the id of a contact of the plan gives `directory.holder-not-a-contact`. The id of a role or a team is not a contact. A contact added by the same edit does not count, since an edit holds one entry.
  - Each key of a team's holders that is not the id of a role of the plan gives `directory.team-role-unknown`, and each holder in the list under a key that is not the id of a contact gives `directory.holder-not-a-contact`, whether or not the key names a role.
  - A team whose `contact` is not empty and is not the id of a contact of the plan gives `directory.holder-not-a-contact`, with the detail `contact` and no `holder` and no `role`. The id of a role or a team is not a contact, and ids are compared as written. An empty `contact` means the team has none and gives no problem. The contact is kept apart from the role holders: a team's `contact` is never a holder of a role ([0080](../../../docs/decisions/0080-directory.md), "Integrity").
- The problems come in a fixed order. The id comes first, then the name, then the holders of a role in the order given. For a team the `contact` comes after the name and before every override, and then the keys come in ascending byte order. For one key, `directory.team-role-unknown` comes before the problems of its holders, and the holders come in the order given.
- Every problem has `severity` `error`, no `question`, and no `section`. The detail is in the table below.
- Two entries with one id cannot come from a Set. A directory that already has two is not covered.

**The three Remove operations.** `RemoveContact`, `RemoveRole`, and `RemoveTeam` take an id and remove the entry of that kind with it.

- An id that no entry of that kind has, including an empty id, gives one problem, `directory.entry-not-found`, with severity `error`, and the returned plan equals the plan given. The id of a different kind is not found, and ids are compared as written, so `LEE-PARK` is not `lee-park`. `Cleared` is empty.
- Otherwise the entry is removed and there is no problem. The edit also clears what named it. It changes nothing else.
- **A contact** is removed from every role's `holders` and from every list in every team's `holders`, every occurrence, with the other holders left in their order. A list that is left empty stays as an empty list, which for a team means the role is now vacant for the team. The editor does not delete the override, because that would hand the role to its own holders. It also clears the `contact` of every team whose `contact` is the removed id, compared exactly as written. The team stays with its name and its overrides, and it has no `contact` afterwards. The `contact` of a team that names another contact stays.
- **A role** is removed with every override for it in every team's `holders`, the key and its list. The role's own holders go with the role.
- **A team** is removed with its overrides and its `contact`. The contact it named stays in the directory.
- **Answers.** An answer is cleared when the variable it answers has the type of the removed kind (`contact`, `role`, or `team`) and the answer names the id. This covers the plan's answers and the answers of every section, a section that is off or a sub-section included, and a question that is hidden or not active. The variable is the one the plan or the section declares under that name. An answer in a scope with no such variable is left alone. Only the type counts. A string, person, or command answer that holds the same text, an answer of a variable of another directory type that holds the same text, and a Deferred answer are left alone.
  - A Text answer equal to the id is removed. Text is compared as stored, exactly.
  - A List answer loses every item equal to the id and keeps the other items in their order. When the list then has no item of non-empty text, the answer is removed.
  - The editor never leaves an unanswered value behind. A cleared answer leaves no entry of no kind, and a section whose last answer is cleared leaves no empty map in `answers.sections`.
- **`Cleared.Answers`** lists the id of each question whose stored answer was removed or changed, as `<scope>.<name>`, such as `plan.escalation` or `access.approver`. A list that lost an item counts. The list is in ascending byte order, each id once. It is empty when nothing was cleared.
- **`Cleared.Holders`** lists each holder place that the edit changed, in ascending byte order, each once. `role:<role id>` is a role whose own holders lost the contact. `team:<team id>/<role id>` is a team override that lost the contact, or that was removed with its role. `team-contact:<team id>` is a team whose `contact` was cleared because it named the removed contact, once for the team however many overrides it has. In byte order `team-contact:<id>` comes before every `team:<id>/<role>`, because the hyphen sorts before the colon. The places that go with the removed entry itself are not listed: the removed role's own holders, the removed team's overrides, and the removed team's `contact`. Removing a role or a team clears no team `contact`, because a `contact` names a contact and a role or a team with the same id is another entry. It is empty when nothing was changed.
- A refused removal returns the plan as given and an empty `Cleared`.

**Submit and the directory.** A question whose variable has the type `contact`, `role`, or `team` holds an id ([0080](../../../docs/decisions/0080-directory.md)).

- A value that names an id that is not an entry of that kind in the plan as given is refused, and the plan comes back as given. For a list, one unknown item refuses the whole list. The id of another kind does not count: a team's id is not a contact.
- The validator makes this check and the editor returns its problems as they are, as for every other check of `Submit`. The code and the detail are the validator's, and the cases here compare only the severity, the question, and the section. The editor adds nothing of its own.
- A known id is stored. A Deferred value, and a value that clears the answer, are applied as for any type.
- Only the value is judged. An answer already stored that names an id the directory lacks does not stop another `Submit`. A question of any other type is not checked against the directory, so a `string` answer that names no contact is applied.

## The problems

| Code | Severity | From | When | Detail |
| -- | -- | -- | -- | -- |
| `answer.*` and `secret.*` | As the validator reports them | `Submit`, from the validator | The value or the question is not acceptable, an unknown directory id included | As the validator reports them |
| `switch.unknown-section` | `error` | `SetSectionEnabled` | The plan holds no section with that id | none |
| `directory.id-malformed` | `error` | The Set operations | The id is empty or is not `[a-z0-9][a-z0-9-]*` | `kind` (`contact`, `role`, or `team`), `id` as given |
| `directory.name-empty` | `error` | The Set operations | The name is empty | `kind`, `id` |
| `directory.holder-not-a-contact` | `error` | `SetRole`, `SetTeam` | A holder is not the id of a contact, or a team's `contact` is not | `kind`, `id`, `holder`, and for a team also `role`, the key it was under. For a team's `contact`: `kind`, `id`, and `contact`, the value given |
| `directory.team-role-unknown` | `error` | `SetTeam` | A key of the team's holders is not the id of a role | `kind`, `id`, `role` |
| `directory.entry-not-found` | `error` | The Remove operations | No entry of that kind has the id | `kind`, `id` as given |

Every problem has a non-empty `message` for a person. The wording is not specified, and a problem never repeats a secret or a part of it.

## Properties

These hold for every plan the contract tests use.

- **Arguments unchanged.** The plan, the value, and the entry are the same after the calls as before.
- **Same input, same output.** The same call gives the same result on a shared editor, whatever calls came before, forward or in reverse order, however often it is repeated, and when calls run at the same time.
- **Nothing shared that an edit changes.** The returned plan holds its own `answers`, with their maps and their lists, and its own `sections` list and each section's `enabled`. The result of a directory operation also holds its own directory: the lists of contacts, roles, and teams, each role's holders, and each team's holders map with its lists. The entry that a Set stores is a copy of the one it was given. The lists of a `Cleared` and the problems belong to the caller. Changing any of them in the result changes neither the argument nor a later result, and changing the same parts of the argument after the call does not change the result. This holds when the answer is refused too. A list in a submitted value is copied when it is stored. Other parts of the plan, such as a section's variables and steps, may be shared.
- **Sections keep their `parent`.** Every section of the returned plan has the `parent` it had in the argument, when the call applies and when it refuses, for `Submit` and for `SetSectionEnabled`. The editor does not read `parent`, check that it names a section, or move a section for it. The order of the sections is the argument's.
- **Only the named change.** A `Submit` that applies changes one entry of `answers` and nothing else. `SetSectionEnabled` changes the `enabled` field of one section and nothing else. A Set changes the directory and nothing else. A removal changes the directory and the answers it clears, and `Cleared.Answers` names exactly the answers that changed. In all of them, every field outside the change, `unknownFields`, `id`, `provenance`, and `origin` included, comes back as it was.
- **Examples stay.** Every variable of the returned plan, in `variables` and in each section, keeps its `example` as held, whether the call applies, refuses, clears, switches, sets an entry, or removes one. An `example` is an illustration for the person answering. It is not an answer and not a default, and no call changes it.
- **Section order.** The sections of the result are in the order of the argument's sections. Rotating or reversing the argument's sections does not change the answers or the problems.
- **Repeating changes nothing more.** The same call on the result gives a plan equal to the result. For a removal the repeat gives `directory.entry-not-found` and the same plan.
- **Sorted and plain lists.** A `Cleared` holds each place once, in ascending byte order, whatever the order of the sections, the maps, or the lists in the plan.
- **No empty leftovers.** After a directory operation, `answers.sections` holds no empty map and no answer has no kind.
- **Take back.** Clearing an answer that a `Submit` added gives the plan from before. Turning off a section that has no `enabled` field and turning it on again gives the plan from before. Removing an entry that a Set added, when nothing else names it, gives the plan from before.
- **New lists.** Changing the result's problems, including a problem's `detail`, changes no later result.
- **Nothing to edit.** On a plan with nothing in it, `Submit` gives `answer.unknown-question`, `SetSectionEnabled` gives `switch.unknown-section`, each removal gives `directory.entry-not-found`, and the plan stays empty. `SetContact` adds the first contact.

## Side effects and plan access

None. `Submit` returns a changed copy that differs only in `answers`, `SetSectionEnabled` returns a changed copy that differs only in the `enabled` field of one section, a Set returns a copy that differs only in `directory`, and a removal returns a copy that differs only in `directory` and `answers`. None alters the plan it was given. The row for `plan/editor` in [components](../../../docs/architecture/components.md) lists `answers` only, and 0010 says the copy "differs only in `answers`", so the row needs the switch and the directory added. The editor imports the finder, the validator, and the model, and nothing else, as [0020](../../../docs/decisions/0020-binary-structure-and-regenerative-components.md) says.

## Not covered

- Whether a section's map stays in `answers.sections` after `Submit` removes its last answer. A removal from the directory never leaves one.
- A directory that already holds two entries of one kind with one id, and a name of only spaces. Whether an emptied `holders` map of a team is empty or absent.
- Whether the plan's own `answers.plan` map stays when a removal clears its last answer.
- Warnings about the directory: an email, a role with no holder, an answer that names an unknown id already stored. The validator reports them. So is a team `contact` that already names no contact in a plan as given: a Set of another entry and a removal of a different kind do not look at it.
- That a `parent` names a section, that no chain of parents loops, and that an `includeIf` names only ancestors. The plan validator's structure check covers them, and the editor copies a section as held.
- Duplicate section ids, and duplicate variable names in one scope.
- A section that holds `enabled: true` and is turned off, and a plan that holds both a `plan` section and plan variables.
- The wording of a message, and the codes the validator will add for checks between sections, which `Submit` does not return.
- A Text that is not valid UTF-8.
- Whether a driver asks for confirmation, shows warnings, or offers a default. Drivers decide that.

## Decisions to know

`hard` choices change what a template author writes, what a plan file holds, or what a person sees in the interface. They come first.

1. **An unanswered value removes the stored answer, and the plan holds no empty answers.** Empty text, an empty list, a list of only empty text, and a value of no kind clear the question. The other choice was to store the empty value. Raised by "Unanswered | Nothing, empty text, or an empty list" in 0035 and "The plan holds only answers" in 0045. Cases: `clear-with-empty-text`, `clear-with-an-empty-list`, `clear-with-a-list-of-empty-text`, `clear-with-no-value`, `clear-a-flag-with-empty-text`. `hard`
2. **The section switch is a second operation, `SetSectionEnabled`, and the plan file records only the sections turned off.** Off stores `enabled: false`, on removes the field, and a hand-written `enabled: true` is removed by a switch to on. Raised by "The editor offers it as an action on a section" and "A section is on unless the plan says otherwise" in 0050. The architecture row for the editor lists `answers` only, and it needs the switch. Cases: `switch-off`, `switch-on-clears-the-field`, `switch-on-clears-an-explicit-true`, `switch-off-over-an-explicit-true`. `hard`
3. **A refused answer returns the plan unchanged and every problem, and a warning comes back with the applied answer.** An error stops the answer, and the problems of the check come back whole, warnings next to the error. No check of `Submit` reports a warning today, so no case shows one. Raised by "If an answer has an error, it is not applied, and the returned plan equals the original. A warning is returned with the new plan" in 0010. Case: `secret-is-refused`. `hard`
4. **A question that is declared and not active cannot be answered, and cannot be cleared.** A stale answer in a section that is off, or hidden by its condition, stays in the plan until the question is active again, and the editor reports `answer.inactive-question`. Raised by "Is team state an active question?" in the 0010 sequence and "Turning a section off keeps its answers" in 0050. Cases: `hidden-question-is-refused`, `question-of-a-switched-off-section-is-refused`, `clearing-an-inactive-question-is-refused`, `onboarding-question-is-refused`. `hard`
5. **The editor adds no problem codes to `Submit`.** Every problem is the validator's, with its code, severity, `question`, `section`, and detail, because the stdio protocol returns them as they are. An unknown question and an inactive one come from the validator. Raised by "Submit(plan, id, value) (Plan, []Problem)" and the errors table in 0010, which name unknown ids as problems without a code. Cases: all the refused cases. `hard`
6. **An unknown section for the switch is the error `switch.unknown-section`.** The `plan` id counts as unknown, and the problem's `section` is the id as given. Raised by 0050, which says nothing about an id the plan does not hold. Cases: `switch-of-an-unknown-section-is-refused`, `plan-is-not-a-section`, `switch-of-an-empty-id-is-refused`. `hard`
7. **A switch reports no problem about other sections.** Turning off a section that another section's condition names, or whose questions are blocking, gives no warning. Raised by "It is not a question, it never blocks" in 0050. Cases: `switch-off`, `acme-switch-off-a-section`. `cheap`
8. **An answer is stored exactly as given.** The editor does not trim text, drop empty items from a list, or turn the text `deferred` into a Deferred value. Raised by "a single space is a legitimate answer" in 0040 and "The quoted `"deferred"` is the Text answer with that word" in 0035. Cases: `whitespace-text-is-an-answer`, `list-is-stored-as-given`, `the-word-deferred-is-text`. `cheap`
9. **The returned plan is a new plan that equals the original in a refusal, and it shares none of the parts the editor changes.** The answers, the sections list, and each section's `enabled` are fresh in every result. This is a stricter reading of "Plans are values" than the sharing allowed in 0015, so that a driver or a test that changes one plan cannot change another. Raised by "A returned value may share unchanged parts with its input" in 0015 and "No shared mutation" in 0010. Cases: the property test of shared parts. `cheap`
10. **The editor reaches the finder through the validator.** The validator's value check already asks the finder whether the question is active, so the editor makes no second activity check. Raised by the 0010 sequence, which shows the editor asking the finder and then the validator. Cases: the cases of decision 4. `cheap`
11. **Clearing needs an active question like any answer.** An unanswered value for an unknown question is `answer.unknown-question`, and for one that is not active it is `answer.inactive-question`. Raised by the same sequence. Case: `clearing-an-unknown-question-is-refused`. `cheap`
12. **A switch to the state a section already has is not a problem.** The result equals the plan, apart from a hand-written `enabled: true` that is removed. Raised by 0050, which does not say. Cases: `switch-on-when-on`, `switch-off-when-off`. `cheap`
13. **The sections of the result keep the argument's order.** Raised by "differs only in" in 0010. Cases: the section order property. `cheap`
14. **A switch does not cascade in the plan.** Turning a parent off writes its own `enabled: false` and nothing on a sub-section, and turning it on leaves the explicit `true` of a sub-section in place. The selector applies the cascade when it is read, as 0070 says, so a plan file records only the switches a person made. The other choice was to write the switch on every descendant. Raised by 0070 "Switches cascade" and 0050 "the plan records only the sections turned off". Cases: `switch-off-a-parent-leaves-its-sub-sections-alone`, `switch-on-a-parent-leaves-its-sub-sections-alone`, `switch-off-a-sub-section-alone`. `hard`
15. **A question of a sub-section of a section that is off cannot be answered.** The editor relies on the finder, which treats a sub-section as active only when its parent is. Raised by 0070 and the finder's new rule. Case: `question-of-a-sub-section-of-a-switched-off-section-is-refused`. `cheap`

### Decisions for the directory operations

The `hard` ones come first.

16. **A vacated team override stays as an empty list.** Removing a contact that is the only holder in a team's override leaves the override with an empty list, which means the role is vacant for that team. Deleting the override would hand the role to its organization-wide holders, which is a different answer to "who holds it". Raised by "A team may override who holds a role for that team" and "No holder is a gap" in 0080. Cases: `remove-contact-clears-every-holder-and-answer-that-names-it`, `remove-contact-leaves-no-empty-section-map`, `remove-contact-leaves-the-role-and-team-with-the-same-id`. `hard`
17. **Answers are cleared by the type of the variable and not by the text.** A `string` or `command` answer that holds `lee-park` stays when the contact `lee-park` is removed, and a role answer `ops` stays when the contact `ops` is removed. Raised by "An answer stores an id and nothing else" and "clears every answer ... that named it" in 0080. Cases: the three `remove-contact` cases above, `remove-role-leaves-the-contact-and-team-with-the-same-id`, `remove-team-leaves-the-contact-and-role-with-the-same-id`. `hard`
18. **A removal reports the places it cleared as lists, not as a count.** `Cleared.Answers` holds question ids and `Cleared.Holders` holds places such as `role:ops` and `team:platform/ops`, sorted. A driver can show them and count them. Raised by "reports how many places it cleared" in 0080. Cases: every `remove-*` case. `hard`
19. **Answers in sections that are off, hidden, or sub-sections are cleared too.** A reference is a dangling reference whether or not its question is active. Raised by "The tool never leaves a dangling reference behind" in 0080 and "Turning a section off keeps its answers" in 0050. Case: `remove-contact-clears-every-holder-and-answer-that-names-it`, where the answer is in a section that is switched off, and `remove-contact-leaves-no-empty-section-map`, where it is in a sub-section. `hard`
20. **Removing an id that is not there is the error `directory.entry-not-found`, and nothing changes.** The id of another kind is not there either. Raised by "Removing a contact, a role, or a team is allowed" in 0080, which says nothing of an unknown id. Cases: the `remove-*-refuses-*` cases. `hard`
21. **A Set replaces the entry whole and in place, and adds a new one at the end.** A field or holder the new entry leaves out is gone. The other choice was to merge. Raised by "operations to add, change, and remove" in 0080. Cases: `set-contact-replaces-the-entry-with-the-same-id-in-place`, `set-role-with-no-holders-empties-an-existing-role`, `set-team-replaces-in-place-and-drops-the-old-overrides`, `set-team-with-no-overrides-clears-the-old-ones`. `hard`
22. **A Set that would break integrity is refused as a whole.** An empty or malformed id, an empty name, a holder that is not a contact, and a team override for an unknown role are errors, so the plan never holds them through the editor. Raised by the "Errors" of "Integrity" in 0080, which lists them for the validator. Cases: the `set-*-refuses-*` cases. `hard`
23. **An emptied list answer is removed, and so is a list left with only empty text.** A list that loses the id and still holds others stays a list, and empty items that were already in it stay. Raised by "The plan holds only answers" in 0045 and "A `multiple: true` contact or role question holds an ordered list" in 0080. Cases: `remove-contact-clears-every-holder-and-answer-that-names-it`, `remove-contact-removes-a-list-that-holds-only-empty-text`. `hard`
24. **The id pattern is `[a-z0-9][a-z0-9-]*` from end to end.** A digit may come first and a hyphen may come last. The pattern comes from the decisions of the directory work and not from 0080, which says only that an entry has a permanent id. Cases: the `set-contact-refuses-an-id-*` cases, `set-contact-takes-a-digit-first-id-and-a-trailing-hyphen`, `set-contact-takes-a-one-character-id`. `cheap`
25. **A Set judges the entry against the directory as given.** A role cannot hold a contact that the same edit adds, because an edit holds one entry. Raised by "a holder that is not a contact" in 0080. Cases: `set-role-with-a-holder-on-a-plan-with-no-contacts`. `cheap`
26. **The editor fixes the order and detail of the directory problems.** The order is id, name, then holders, and for a team the keys in byte order with the role check before its holders. The detail keys are `kind`, `id`, `holder`, and `role`. 0080 names the faults and not the report. Cases: `set-contact-reports-the-id-before-the-name`, `set-role-reports-id-name-and-holders-in-that-order`, `set-team-reports-overrides-by-role-id-and-the-role-before-its-holders`. `cheap`
27. **A Set reports no warning.** An email with no `@`, a role with no holder, and a contact with no way to reach them are the validator's warnings, and a Set stores such an entry. Raised by "Warnings" in "Integrity" in 0080. Cases: `set-role-with-no-holders-is-a-vacant-role`. `cheap`
28. **A role's removal reports its team overrides as cleared places and not its own holders, and a team's removal reports none.** The removed entry's own contents are not a place that was cleared. Raised by "clears every answer and every holder entry that named it" in 0080. Cases: `remove-role-removes-team-overrides-and-clears-answers`, `remove-role-with-no-reference-changes-only-its-entry`, `remove-team-clears-the-team-answer`. `cheap`
29. **The unknown-id refusal in `Submit` is the validator's, and these cases do not fix its code.** The cases compare the severity, the question, and the section. Raised by "An answer that names an id that is not in the directory counts as unanswered" and "The tool never creates such an answer" in 0080. Cases: the `unknown-*-is-refused` and `*-refuses-a-*-id` cases of `Submit`. `cheap`

### Decisions for a team's contact

The `hard` ones come first.

30. **Removing a contact clears the team contact that names it and leaves the team.** The team keeps its name and overrides and has no `contact` afterwards. The report lists `team-contact:<team id>`. The other choice was to refuse the removal while a team names the contact. Raised by "clears a team's `contact` that named it" and "`team-contact:<team id>`" in 0080. Cases: `remove-contact-clears-the-contact-of-every-team-that-named-it`, `remove-contact-clears-a-team-contact-and-a-role-holder-together`. `hard`
31. **A Set of a team whose `contact` is not a contact is refused as a whole, with `directory.holder-not-a-contact`.** There is no new code. The detail has `kind`, `id`, and `contact`. A team's `contact` is never a role holder, so the detail carries no `holder` and no `role`. Raised by "a team `contact` that is not a contact" in the Set rule of 0080, and by its error table, which gives the same code. Cases: the `set-team-refuses-a-contact-*` cases. `hard`
32. **A removal of a role or a team clears no team `contact`.** A `contact` holds a contact id, and a role or a team that shares the id is a different entry. Raised by "ids are per kind" in 0080. Cases: `remove-role-leaves-a-team-contact-with-the-same-id`, `remove-team-leaves-the-team-contacts-that-share-its-id`. `hard`
33. **A replaced team keeps only the `contact` it is given.** A Set without a `contact` removes the old one, as it removes old overrides. Raised by "replaces the entry with the same id whole" in 0080. Case: `set-team-with-no-contact-drops-the-old-contact`. `hard`
34. **The team `contact` problem sorts after the id and the name and before the overrides.** It follows the order in which the file writes a team. The cases fix this order. Raised by 0080 naming the faults and not their order. Case: `set-team-reports-the-contact-after-the-name-and-before-the-overrides`. `cheap`
35. **The detail key for the team contact is `contact`, and the contact is compared as written.** An empty `contact` is no contact and no fault, and `Lee-Park` is not the contact `lee-park`. Raised by "a key that says it is the team contact" in the shared decisions of the directory work. Cases: `set-team-refuses-a-contact-in-other-case`, and `remove-contact-clears-the-contact-of-every-team-that-named-it`, where a team with `Lee-Park` keeps it. `cheap`
36. **A Set of a contact or a role leaves team contacts alone.** Replacing a contact keeps the id and so keeps every team that names it. Raised by "It clears nothing" in the Set rule of 0080. Case: `set-contact-keeps-the-team-contacts-that-name-it`. `cheap`

## Wrong implementations tried

Each row names a plausible wrong implementation and the cases or tests that fail it. Each one with a defect in the editor's output fails at least one named test. One variant of `Submit` survives on purpose: it leaves an empty map for a section whose last answer was removed, which the contract leaves open for `Submit`.

| Wrong implementation | Caught by |
| -- | -- |
| An answer with an error is applied anyway | `pattern-mismatch-is-refused`, `hidden-question-is-refused`, `secret-is-refused` |
| The problems are dropped | `secret-is-refused` |
| A refusal returns an empty plan | `secret-is-refused`, `question-of-a-switched-off-section-is-refused` |
| The answer is stored by name only, or a section answer is stored in the plan scope | `same-name-in-a-section-scope`, `new-section-answer`, `command-list-replaces-the-old-list` |
| Empty text, an empty list, or a list of empty text is stored instead of clearing | `clear-with-a-list-of-empty-text`, `clear-a-plan-answer`, `clear-when-nothing-is-stored` |
| Only empty text clears, or only a value of no kind clears | `clear-with-an-empty-list`, `clear-with-a-list-of-empty-text`, `answer-hides-a-question-and-keeps-its-answer` |
| A Deferred value clears the answer, or the text `deferred` is treated as unanswered | `deferred-is-applied`, `the-word-deferred-is-text` |
| Text is trimmed, so a text of spaces is cleared | `whitespace-text-is-an-answer` |
| Empty items are dropped from a list | `list-is-stored-as-given` |
| Clearing removes every answer of the section | `clear-with-an-empty-list`, `clear-with-no-value` |
| Clearing skips the check, so an inactive or unknown question is cleared | `clearing-an-inactive-question-is-refused`, `clearing-an-unknown-question-is-refused` |
| The answers of the other sections are dropped | `question-after-switch-on-takes-an-answer` |
| An inactive or unknown question is applied, because those problems are filtered out | `a-question-of-a-section-out-is-refused`, `unknown-question-is-refused`, `empty-id-is-refused` |
| Provenance changed, or the unknown fields dropped | `new-plan-answer`, `acme-change-the-clone-url`, `list-is-stored-as-given` |
| A section turned on stores `enabled: true`, or keeps a hand-written true | `switch-on-clears-the-field`, `switch-on-when-on`, `switch-on-clears-an-explicit-true` |
| A section turned off clears the answers of the section | `switch-off-when-off`, `acme-switch-off-a-section` |
| Every section is turned off, or the first section is turned off instead | `switch-off`, `switch-leaves-other-sections-alone` |
| The sections are sorted by id, by `Submit` or by `SetSectionEnabled` | `new-plan-answer`, `switch-on-when-on`, the section order property |
| An unknown section gives no problem, or a warning, or no `section` | `switch-of-an-unknown-section-is-refused`, `plan-is-not-a-section`, `switch-of-an-empty-id-is-refused` |
| Turning off a section reports a warning | `switch-off`, `switch-off-when-off` |
| A variable's `example` dropped when the plan is copied, in the plan's variables, in every section, or in one section, after an applied answer, a clear, a refusal, or a switch | `examples-kept-by-a-section-answer`, `examples-kept-by-a-plan-answer`, `examples-kept-by-a-clear`, `examples-kept-by-a-refused-answer`, `examples-kept-by-a-switch-off`, `examples-kept-by-a-switch-on` |
| An argument is changed | The arguments-unchanged test |
| The result shares the answers, the sections list, a stored list, or an `enabled` pointer with the argument | The shared-parts test, the stored-list test, the arguments-unchanged test |
| A copy of the sections that drops `parent`, in an applied answer, a clear, a refusal, or a switch | `parents-kept-by-a-sub-section-answer`, `parents-kept-by-a-clear`, `parents-kept-by-a-refused-answer`, `switch-off-a-sub-section-alone`, the parent test |
| A copy that keeps the `parent` of the first level only, or drops it from a section that has a parent that is itself a sub-section | `parents-kept-by-an-answer-of-a-sub-sub-section`, the parent test |
| A switch off that writes `enabled: false` on the sub-sections, or turns off the parent of a sub-section switched off alone | `switch-off-a-parent-leaves-its-sub-sections-alone`, `switch-off-a-sub-section-alone` |
| A switch on that removes the explicit `enabled` of a sub-section | `switch-on-a-parent-leaves-its-sub-sections-alone` |
| An answer of a sub-section of a switched-off section applied | `question-of-a-sub-section-of-a-switched-off-section-is-refused` |
| An earlier result returned on a later call | The same-input test, and most cases |
| A Set adds a duplicate instead of replacing, or moves the replaced entry to the end | `set-contact-replaces-the-entry-with-the-same-id-in-place`, `set-role-replaces-in-place-and-leaves-team-overrides` |
| A Set merges with the old entry, for a contact, a role's holders, or a team's overrides | `set-contact-replaces-the-entry-with-the-same-id-in-place`, `set-role-with-no-holders-empties-an-existing-role`, `set-team-with-no-overrides-clears-the-old-ones` |
| A Set cleans the entry: it trims the name, lowercases the email, sorts or deduplicates the holders, or drops an empty override | `set-contact-stores-the-entry-as-given`, `set-role-stores-the-holders-as-given`, `set-role-adds-a-new-role-at-the-end`, `set-team-keeps-an-empty-override-as-a-vacancy` |
| A Set clears answers that name the id | `set-contact-replaces-the-entry-with-the-same-id-in-place` |
| The id check allows capitals, an underscore, a space, a leading hyphen, or a trailing line break, or is not anchored at the start or the end | `set-contact-refuses-an-id-with-capitals`, `set-contact-refuses-an-id-with-an-underscore`, `set-contact-refuses-an-id-with-a-space`, `set-contact-refuses-an-id-with-a-leading-hyphen`, `set-contact-refuses-an-id-with-a-trailing-line-break` |
| The id check refuses a digit first, a trailing hyphen, or a single character, or accepts an empty id | `set-contact-takes-a-digit-first-id-and-a-trailing-hyphen`, `set-contact-takes-a-one-character-id`, `set-contact-refuses-an-empty-id` |
| An empty name is accepted | `set-contact-refuses-an-empty-name`, `set-role-refuses-an-empty-name`, `set-team-refuses-an-empty-name` |
| A role's holders are not checked, or are checked against the ids of every kind | `set-role-refuses-unknown-holders-in-order`, `set-role-refuses-holders-that-are-a-role-or-a-team` |
| A team's override for an unknown role, or its holders, are not checked | `set-team-refuses-an-override-for-an-unknown-role`, `set-team-refuses-an-override-holder-that-is-not-a-contact` |
| The problems are in another order: the name before the id, holders before the id, the keys in map order, holders before the role check | `set-contact-reports-the-id-before-the-name`, `set-role-reports-id-name-and-holders-in-that-order`, `set-team-reports-overrides-by-role-id-and-the-role-before-its-holders` |
| A refused Set is applied anyway | every `set-*-refuses-*` case |
| A Set stores the list or map it was given | `TestStoredEntryIsNotTheEntryThatWasSubmitted` |
| A removal ignores the team overrides, or the role holders | `remove-contact-clears-every-holder-and-answer-that-names-it`, `remove-contact-leaves-the-role-and-team-with-the-same-id` |
| A vacated override is deleted instead of left empty | `remove-contact-clears-every-holder-and-answer-that-names-it`, `remove-contact-leaves-no-empty-section-map` |
| Answers are cleared by text and not by the type of the variable, or answers with no declared variable are cleared | `remove-contact-clears-every-holder-and-answer-that-names-it`, `remove-contact-leaves-the-role-and-team-with-the-same-id`, `remove-role-leaves-the-contact-and-team-with-the-same-id` |
| The text match ignores case, or a Deferred answer is cleared | `remove-contact-clears-every-holder-and-answer-that-names-it` |
| A list answer is removed whole, loses only its first occurrence, is kept when emptied, or is kept when only empty text is left | `remove-contact-clears-every-holder-and-answer-that-names-it`, `remove-contact-removes-a-list-that-holds-only-empty-text` |
| An empty map is left for a section whose last answer was cleared | `TestClearedAnswersLeaveNoEmptySectionMapAndNoEmptyEntry` |
| Section answers, answers of a switched-off section, sub-section answers, or answers of a hidden question are skipped | `remove-contact-clears-every-holder-and-answer-that-names-it`, `remove-contact-leaves-no-empty-section-map` |
| `Cleared` is in map order, repeats a place, names the question by its name only, leaves out list answers, or spells a place differently | `remove-contact-clears-every-holder-and-answer-that-names-it`, `remove-contact-leaves-no-empty-section-map`, `TestClearedPlacesAreSortedAndListedOnce` |
| A role's removal keeps its overrides, does not report them, or reports the role's own holders | `remove-role-removes-team-overrides-and-clears-answers`, `remove-role-leaves-the-contact-and-team-with-the-same-id` |
| A team's removal keeps the answer, or reports its overrides | `remove-team-clears-the-team-answer`, `remove-team-with-no-answer-changes-only-its-entry` |
| An entry is not removed | every applied `remove-*` case |
| A removal of an unknown id gives no problem, a warning, or accepts the id of another kind | the `remove-*-refuses-*` cases |
| A removal drops `unknownFields`, `provenance`, `origin`, `id`, or a variable's `example`, or reorders the sections | every applied `remove-*` case |
| A directory operation shares the directory or the answers with the argument, or edits the argument | `TestArgumentsAreNotChanged`, `TestResultSharesNothingWithTheArgumentThatAnEditCouldChange` |
| `Submit` does not check the items of a list, checks ids of every kind, checks every text question, skips section questions, or reports a warning | `a-list-with-one-unknown-contact-is-refused`, `contact-question-refuses-a-team-id`, `string-question-takes-text-that-names-no-contact`, `unknown-section-contact-is-refused`, `unknown-contact-is-refused` |
| A Set does not check a team's `contact`, or checks it against the ids of every kind, or ignoring case | `set-team-refuses-a-contact-that-is-not-in-the-directory`, `set-team-refuses-a-contact-that-is-only-a-role`, `set-team-refuses-a-contact-that-is-only-a-team`, `set-team-refuses-a-contact-in-other-case` |
| A Set refuses an empty `contact`, or reports a bad contact as a warning, or with the key `holder` | `set-team-with-no-contact-drops-the-old-contact`, `set-team-adds-a-new-team-at-the-end`, `set-team-refuses-a-contact-that-is-not-in-the-directory` |
| A Set reports the team contact before the name, after the overrides, or only when the id and name are fine | `set-team-reports-the-contact-after-the-name-and-before-the-overrides` |
| A Set drops the `contact` it stores, or keeps the old `contact` of a replaced team | `set-team-with-a-contact-adds-the-team-at-the-end`, `set-team-replaces-in-place-and-changes-the-contact`, `set-team-with-no-contact-drops-the-old-contact` |
| A refused team Set is applied, or leaves the old team changed | `set-team-refuses-a-replacement-with-a-bad-contact-and-keeps-the-old-team` |
| A removal of a contact leaves the team `contact`, clears it and does not report it, or reports it as `team:<id>` or with the team name | `remove-contact-clears-the-contact-of-every-team-that-named-it`, `remove-contact-clears-a-team-contact-and-a-role-holder-together` |
| A removal of a contact clears every team `contact`, or matches ignoring case | `remove-contact-keeps-the-contact-of-a-team-that-names-another-contact`, `remove-contact-clears-the-contact-of-every-team-that-named-it` |
| A removal of a contact deletes the team or its overrides when it clears the `contact` | `remove-contact-clears-the-contact-of-every-team-that-named-it`, `remove-contact-clears-a-team-contact-and-a-role-holder-together` |
| A removal reports the team `contact` once for each override, or out of byte order | `remove-contact-clears-the-contact-of-every-team-that-named-it`, `remove-contact-clears-the-team-contact-and-leaves-the-role-and-team-with-the-same-id`, `TestClearedPlacesAreSortedAndListedOnce` |
| A removal of a role or a team clears the team `contact` that has the same text | `remove-role-leaves-a-team-contact-with-the-same-id`, `remove-team-leaves-the-team-contacts-that-share-its-id` |
| A removal of a team reports its own `contact` | `remove-team-with-a-contact-removes-the-contact-with-the-team`, `remove-team-leaves-the-contact-it-named` |
| A Set of a contact clears the team contacts that name it | `set-contact-keeps-the-team-contacts-that-name-it` |
