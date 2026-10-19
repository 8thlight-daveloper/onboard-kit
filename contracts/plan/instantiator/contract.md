# Plan instantiator contract

Status: draft by an agent. Not reviewed by a person. Component: `plan/instantiator`. Issue: [#67](https://github.com/8thlight-daveloper/onboard-kit/issues/67).

The plan instantiator builds the default plan when a new plan starts. It takes the loaded templates and the axis variables, and returns a Plan that holds the axis variables and every template as a section, with no answers. It is a pure function. It reads no file, reads no clock, keeps no state between calls, and does not change its arguments. A driver loads the templates, reads the date, and saves the plan. See [0030](../../../docs/decisions/0030-self-contained-plan-file.md), the [domain model](../../../docs/reference/domain-model.md), the [components](../../../docs/architecture/components.md) page, and [0015](../../../docs/decisions/0015-functional-core-and-imperative-shell.md).

The default plan is derived and never committed. The templates are used once, here. Everything after this, editing answers, changing an axis, assembling, works on the plan alone.

## Interface

`Instantiate(templates, axisVariables, templateSet, generatedOn) -> (plan, problems)`.

- `templates` is a list of Section, as the [section loader](../../section/loader/contract.md) returns them, one for each template file.
- `axisVariables` is a list of Variable. They come from the small built-in axis question file, and the instantiator treats them as data.
- `templateSet` is the template set version, such as the build's commit sha or tag.
- `generatedOn` is the date, as text in ISO form. A driver passes it in, because the core never reads a clock.

The constructor, `New()`, takes no argument. The component uses the [section hasher](../../section/hasher/contract.md) and no other component.

When the inputs are acceptable, `problems` is empty and `plan` is the default plan. When they are not, `plan` is the zero Plan and `problems` lists each one. A problem has severity `error` and a non-empty message. A mistake in the inputs is a normal result, not a failure, as the domain model says of a Problem.

## The default plan

| Field | Value |
| -- | -- |
| `schemaVersion` | `1` |
| `provenance` | `templateSet` and `generatedOn` as given |
| `variables` | The axis variables, as given, in the order given |
| `answers` | None. Both maps have no entries |
| `sections` | One section for each template, in id order. See Order |
| `unknownFields` | None |

The plan has no answers, so a default declared on a variable stays a declared default. It is not copied into `answers`.

### Sections

Each section is its template, copied, with `source` set. The copy changes nothing else.

- `source.template` is the template's `id`.
- `source.hash` is the section hasher's hash of the template as given.
- A `source` that the template already has is replaced. The hash does not cover it.
- Nothing is filled in. A template with no `timing` has no `timing` in the plan, although its hash is of the default `anytime`. The same holds for a variable's `type` and `phase`.
- Nothing is resolved, selected, or ordered by `timing` or `after`. `$name` stays in place, and `includeIf`, `askIf`, `after`, and `enabled` are kept as loaded.
- The unknown fields of a section are kept. The loader gives none, and a section from another source may have some.
- Text is not trimmed or normalized. Only the hasher normalizes, and only for its hash.

### Axis variables

The variables are copied as given, in the order given. Nothing is added, filled, sorted, or checked, apart from the name rule under Problems. The instantiator does not know which variables are axis variables and which are document details. If a new plan should also carry `organization`, `team`, or `output_style`, the built-in file holds them.

### Order

Sections are in ascending byte order of `id`, so `10-x` comes before `9-x`, and `a` before `a-b`, before `a1`. After the duplicate check, no two templates have the same id, so no two sections compare equal and a sort that keeps equal ids in place and one that does not give the same result. No test tells them apart, and none could.

The order in which templates are given has no effect, because the driver lists files in whatever order the file system uses, and the same template set must give the same plan file. The stored order is not the document order. The section selector computes that from `timing`, `after`, and `id` when it is needed.

## Problems

Every problem has severity `error`. All are reported, and a problem means the zero Plan.

| Code | When | `section` | `detail` |
| -- | -- | -- | -- |
| `template.duplicate-id` | Two or more templates have the same `id`. The templates may be identical or not | The id | `count`: how many templates have it, as decimal text |
| `variable.no-name` | An axis variable has an empty `name` | none | `index`: its place in the list given, counted from zero, as decimal text |

- One `template.duplicate-id` problem for each id that repeats, in ascending byte order of the id, not one for each extra template.
- One `variable.no-name` problem for each nameless variable, in the order given.
- The template problems come before the variable problems.
- The problem's `question` is empty.
- The `message` is for people and is not specified.

Answers are kept by section id, and an id is how a section is found again, so a plan cannot hold two sections with one id. A variable's name is the key of its answer. An empty id is an id like any other, so two templates with no id are a repeat of the empty id.

## Properties

These hold for every input the contract tests use.

- **Deterministic.** The same arguments give the same result.
- **No memory.** An instantiator that has built other plans gives the same result as a new one, whatever came before.
- **Order of templates.** Giving the templates in any order gives the same plan, or the same problems.
- **Inputs unchanged.** The templates, the axis variables, and everything they hold are the same after the call as before. Sorting happens on a copy.
- **Nothing shared.** No list, map, or pointer in the result is shared with an argument, at any depth, including inside unknown fields. Changing the result changes no argument, and changing an argument after the call changes no result.
- **Whole or nothing.** A problem means the zero Plan. A Plan means no problems. One problem is enough: a single duplicate id, or a single nameless variable, gives the zero Plan and that one problem.
- **Empty and absent are the same.** A list or map that is absent in an argument may be absent or empty in the result, and the same holds for one that is empty. Tests treat a nil list or map and an empty one as equal, at every depth, including inside unknown fields. This matches the [section resolver contract](../../section/resolver/contract.md).

## Side effects and plan access

No side effects. It was not given a plan, so it creates one: every field of the plan, as the row for `plan/instantiator` in [components](../../../docs/architecture/components.md) says. It reads the content of each template to copy and hash it.

## Fixtures

The hashes in `testdata/cases.json` come from the section hasher contract, in one of two ways, and each case cites which.

- Five sections and their hashes are copied from the golden cases of [the hasher contract](../../section/hasher/testdata/cases.json), with their canonical texts.
- Ten more sections have a canonical text written by hand from the rules of the hasher contract. Their hashes are `shasum -a 256` over those texts. They are the four ACME templates and six small sections for the order case.

A test checks that each hash in the manifest is the SHA-256 of its canonical text. No hash was computed by a component.

The templates of the ACME case are the section loader contract's fixtures for the four templates in [the ACME example](../../../docs/reference/examples/acme/README.md). The ACME plan file records `example-not-computed` for the hashes, so its sources cannot be compared.

## Not covered

- Whether a template is valid: its id, its variables, its conditions, its references. The section validator and the plan validator judge that, and `onboard-kit lint` runs them on the default plan. This includes an id that is not a lowercase slug or is `plan`, an empty id on one template, a duplicate variable name, and a reserved or malformed variable name. Such a template is copied and hashed as it is.
- The order of ids that are not lowercase slugs. Ascending byte order is the rule, and the tests use slugs.
- Whether the axis variables are the right ones, and what the built-in file holds. The driver reads that file.
- Whether `templateSet` or `generatedOn` is well formed, or empty. They are written as given.
- Reading, listing, or loading template files, and reading the clock. The driver does.
- The hash itself. The hasher contract defines it, and these tests rely on its golden cases.
- Updating a plan from a newer template set. It is deferred, see 0030.
- Pruning inactive sections. It is deferred, see 0030.

## Decisions to know

The docs left these open, and this contract picked. The `hard` ones change what a new plan file holds or what a person sees. The `cheap` ones touch one fixture.

1. **`hard`. Sections are in id order, not the order given.** The docs say a plan holds "every standard template as a section" and say nothing about order. The same template set should give the same plan file, whatever order the driver lists the files in, so the plan diff stays quiet. The selector orders sections for use.
1. **`hard`. A section is its template with only `source` added.** No default is filled and no text is normalized. 0030 says "a section copies from a template file into a plan unchanged." A template with no `timing` stays without one in the plan file.
1. **`hard`. The plan's variables are exactly the axis variables given.** The glossary says the default plan is "the axis variables and every standard template as a section". The ACME plan also lists `organization`, `team`, and `output_style` as plan variables. The contract does not say where they come from. If they are meant to be in every new plan, the built-in file holds them, or this contract changes.
1. **`hard`. Two templates with one id are an error, and there is no plan.** 0030 says an id is "stable, never reused" and answers are kept by id. Keeping the first or the last would hide a template from the plan without a word.
1. **`hard`. `schemaVersion` is `1`.** The components page marks it as created here, and 0030 says there is one schema version until the first breaking change.
1. **`cheap`. A nameless axis variable is an error.** The docs say nothing about it. The alternative is to copy it and let the plan validator report it. A name is the key of an answer, and the instantiator does not create a plan it knows is unusable.
1. **`cheap`. No templates, or no axis variables, is not an error.** The plan then holds what is left. The docs say "every template", and none is a valid case of every.
1. **`cheap`. A `source` that the template already has is replaced.** The loader never gives one, so this only fixes what a caller who passes one gets.
1. **`cheap`. The result shares nothing with the arguments.** The docs say plans are values and no component changes an argument. A shared list would let a later edit of the result change the loaded templates.
1. **`cheap`. The date and the template set are written as given, unchecked.** Time comes in as a parameter, and the validators judge the plan.

## Wrong implementations tried

While drafting, a reference instantiator and a reference hasher were written from the contract pages alone as scratch work outside the repo. The reference hasher passed every test of the hasher contract. The reference instantiator passed every test here. Then 60 wrong instantiators were written, one defect each. A stub that returns the zero Plan failed every test that checks a result. The tests failed 59 of the 60. One passed: sorting ids without regard to case, which differs only for ids that are not lowercase slugs and is listed under Not covered.

| Wrong implementation | Caught by |
| -- | -- |
| `schemaVersion` 0 or 2 | every plan case |
| The template set or the date fixed in the code, or the two swapped | `values-as-given`, `acme-default-plan` |
| Sections kept in the order given | `acme-default-plan`, `byte-order`, `hasher-goldens`, the template order test |
| Sections sorted by title, by timing, by `after`, by length of id, by number, by case, or in descending order | `byte-order`, `hasher-goldens`, `acme-default-plan` |
| A default filled on a section, or on a variable | `byte-order`, `hasher-goldens`, `values-as-given` |
| Text of a section trimmed | `acme-default-plan` |
| An existing `source` kept | `hasher-goldens` |
| `source.template` set to the title | every plan case |
| The hash of the title only, shortened, or in upper case | every plan case |
| The hash of the template changed first, for example a title with a full stop added | every plan case |
| Unknown fields of a section dropped | `hasher-goldens` |
| Answers filled from the defaults, or an empty answer map for each section | every plan case with a default or a section |
| A plan with unknown fields | every plan case |
| Axis variables with a default type filled, sorted by name, or with the default dropped | `acme-default-plan`, `axis` cases |
| The axis variables dropped when there are no templates | `no-templates` |
| Built-in variables supplied when none are given | `no-axis-variables`, `nothing-given` |
| An empty set of templates refused | `no-templates`, `nothing-given` |
| The template set dropped when there are no templates | `no-templates`, `nothing-given` |
| The axis variables, the sections, or the unknown fields shared with the arguments | the nothing-shared test |
| A shallow copy that shares a list such as `after` | the nothing-shared test |
| The argument sorted in place, or its `source` set | the inputs-unchanged test |
| An axis variable changed while the result is built | every plan case that has axis variables |
| State kept between calls | the no-memory test |
| A duplicate id dropped silently, first or last kept | `duplicate-ids`, `both-problems` |
| One problem for each extra template, or only the first id reported | `duplicate-ids`, `both-problems` |
| The count wrong, no detail, no section, or `question` set | `duplicate-ids`, `both-problems` |
| Duplicate ids reported in the order first seen, or in reverse | `duplicate-ids`, the template order test |
| A duplicate reported as a warning | `duplicate-ids`, `both-problems` |
| A plan returned with the problems, in part or with the variables | `duplicate-ids`, `nameless-axis-variables`, `both-problems` |
| A nameless variable skipped without a problem | `nameless-axis-variables`, `both-problems` |
| The index counted from one, only the first nameless variable reported, or only one with a description | `nameless-axis-variables`, `both-problems` |
| The variable problems before the template problems, or only one kind reported | `both-problems`, `single-duplicate-and-single-nameless` |
| The plan returned when there is exactly one problem, for example by testing for more than one | `single-duplicate-id`, `single-nameless-axis-variable` |
| The index of a single nameless variable counted from one | `single-nameless-axis-variable` |
| An unknown field's nested list or map shared with the argument, or copied only one level deep | the nothing-shared test on `hasher-goldens`, whose `x-nested` and `a` fields hold nested lists and maps |

Found later by the mutation tool, and judged equivalent because empty and absent are the same: dropping the nil checks that keep an absent list or map absent in the result. The tests compare plans with empty lists and maps treated as absent, at every depth.

No second independent build or differential run was made, because the issue is a draft.
