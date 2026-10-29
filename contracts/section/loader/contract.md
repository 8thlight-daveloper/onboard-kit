# Section loader contract

Status: draft accepted by a person. A hardening pass comes before the first build. Component: `section/loader`.

The section loader turns the bytes of one template file into a Section. It is a pure function. It reads no file, writes no file, and does not change its argument. The same bytes always give the same result. A driver reads the files. See [0030](../../../docs/decisions/0030-self-contained-plan-file.md), the [section template schema](../../../docs/reference/section-template-schema/README.md), the [domain model](../../../docs/reference/domain-model.md), and [0015](../../../docs/decisions/0015-functional-core-and-imperative-shell.md).

A template file is one section before it is in a plan. It has the section fields, a `schemaVersion`, and no `source`. The loader reads the same YAML dialect as the [plan decoder](../../plan/decoder/contract.md), with the same format guards and the same problems. This page lists only what differs. Where it says "as the decoder", the decoder contract is the rule, and a change there is a change here.

## Interface

`Load(bytes) -> (section, problems)`.

- When the file is acceptable, `problems` is empty and `section` is the Section.
- When it is not, `section` is the zero Section and `problems` lists each one. A problem has severity `error` and a non-empty message.

A loaded Section has no `Source` and no `UnknownFields`. The file's `schemaVersion` is checked and then dropped, because the domain model keeps it on the plan and not on the Section.

## What the loader does

- It reads version 1 and refuses every other version. See Version.
- It reads every known template field into the field of the same name in the domain model.
- It rejects an unknown key everywhere, including on the section itself. A template keeps no unknown fields. See Unknown fields.
- It reports the format guards of 0030, as the decoder does: the spellings of booleans, numbers, and null, anchors and aliases, and duplicate keys.
- It does not validate. See What passes through.

## The template file

The known keys at each level. A key not listed is unknown.

| Level | Keys |
| -- | -- |
| Template | `schemaVersion`, `id`, `title`, `summary`, `includeIf`, `parent`, `bootstrap`, `timing`, `after`, `requires`, `owner`, `escalation`, `verification`, `body`, `variables`, `steps` |
| Variable | `name`, `description`, `example`, `type`, `multiple`, `options`, `pattern`, `patternHint`, `required`, `default`, `audience`, `askIf` |
| Condition | `all` |
| Clause, question form | `question`, `answer` |
| Clause, section form | `section`, `enabled` |
| Answer test | `is`, `isNot`, `isOneOf`, `isNotOneOf`, `isSettled` |
| Step | `id`, `name`, `notes`, `owner`, `timing`, `escalation`, `run`, `open`, `directory`, `env`, `checks` |
| Check | `run`, `ask`, `env` |

`example` on a variable is one illustration of a good answer, such as `Engineering manager`. The loader reads it exactly as it reads `description`: a plain or quoted string, with the spelling, wrong-type, and empty-value rules of a text field and the same codes and paths, such as `variables[0].example`. A null, empty, or absent `example` is the empty string. It is text and nothing else: it is not a `default`, it is not an answer, it is not trimmed, and a dollar sign in it is kept as written. A variable may hold an `example` with no `description`.

The `env` of a step is a map whose keys are the variable names it sets, so any scalar key is allowed there. Its values are text.

`enabled` on a template is not a key either. A template never declares the section switch, and a template section is on when it is first added (0050). An `enabled` key on the template, on a variable, on a step, or on a check is unknown, with the problem and the path of any other unknown key, and its value is not read. `enabled` is a key of a clause only, see Clauses.

`source` is not a template key. It belongs to a plan, and authors never write a hash into a template (0030, Identifiers). A `source` in a template is an unknown key, with the problem and the path of any other unknown key.

## What differs from the decoder

| Decoder rule | Loader |
| -- | -- |
| The top level is a Plan, and the plan and each section keep unknown fields | The top level is one Section. Nothing keeps unknown fields |
| `format.unknown-field` only in `provenance`, `source`, a variable, a condition, a clause, an answer test, `answers`, a step, and a check | `format.unknown-field` in every map named in the table above, the top level included. An `x-` key is unknown too |
| `file.not-a-plan` | `file.not-a-template` |
| `source`, `provenance`, `answers`, `sections` are known keys | They are unknown keys. A plan file given to the loader gets `format.unknown-field` for each of them |
| A path starts at the plan, such as `sections[1].steps[0]` | A path starts at the template, such as `steps[0]` |

Everything else is as the decoder: the file problems and their order, the Version table, the order of problems and the place of a problem, the paths, Style is free, Scalars, Numbers, Values, Known fields, and the codes `format.boolean-spelling`, `format.number-spelling`, `format.null-spelling`, `format.alias`, `format.duplicate-key`, `format.wrong-type`, `format.empty-answer`, and `format.empty-item`. The order of the decoder's problems table decides two problems at one place, and `format.unknown-field` comes after `format.wrong-type` in it.

## Version

As the decoder, with the same codes, the same path `schemaVersion`, and `supported` equal to `1`. A template with no `schemaVersion` is `version.invalid`. A version problem is the only problem, so a newer file is not read for anything else: its unknown keys and its other problems are not reported. Two `schemaVersion` keys are `format.duplicate-key` and the only problem.

## Problems

| Code | When | Path |
| -- | -- | -- |
| `file.not-yaml`, `file.multiple-documents` | As the decoder. A template file holds one document | none |
| `file.not-a-template` | The file is empty or holds only comments or a lone `---`, or its top level is not a map. A list of sections is not a template | none |
| `version.invalid`, `version.too-old`, `version.too-new` | See Version | `schemaVersion` |
| `format.boolean-spelling`, `format.number-spelling`, `format.null-spelling`, `format.alias`, `format.duplicate-key`, `format.wrong-type`, `format.empty-answer`, `format.empty-item` | As the decoder | As the decoder |
| `format.wrong-type` | Also a section clause with no `enabled`, or with an empty or null one. See Clauses | The clause |
| `format.unknown-field` | A map has a key that the table above does not list. A clause has the keys of its own form only | The key, such as `summery`, `source`, `enabled`, `steps[0].extra`, or `includeIf.all[0].question` |

Problems come in the order of their places in the file, and all are reported. The detail of a problem names its place in `detail.path`. A file problem is the only problem, and a version problem is the only problem. The value of an unknown key is not read, so nothing inside it is reported. A repeated unknown key is reported each time, with `format.duplicate-key` for each repeat.

## Clauses

A clause has one of two forms, and a key that belongs to the other form is not known to it (0050). The loader reads a clause as the plan decoder does, see Section clauses in the [decoder contract](../../plan/decoder/contract.md). The same clause gives the same problem codes and the same paths in both.

| Form | Known keys | Loads into |
| -- | -- | -- |
| Section clause | `section`, `enabled` | `Section` and `Enabled` |
| Question clause | `question`, `answer` | `Question` and `Answer` |

- A clause is a section clause when it has a `section` key, with any value and in any key order. Otherwise it is a question clause, including a clause with no keys that it knows.
- A key of the other form is `format.unknown-field` at the key, and its value is not read. So `question` and `answer` in a section clause are reported, and so is `enabled` in a clause with no `section`. A key counts when it is present, even when its value is null. The clause is read for its own keys as usual, and the problems of the clause come in the order of their places. See open question 5.
- `section` is text and `enabled` is a boolean, with the format guards of the decoder: `enabled: yes` is `format.boolean-spelling`, and `enabled: 1`, a quoted `"true"`, a list, and a map are `format.wrong-type` at `enabled`. A `section` that is a plain `true` or a list is `format.wrong-type` at `section`, and a canonical number is `format.number-spelling` there.
- A section clause whose `enabled` is absent, empty, or `null` is `format.wrong-type` at the clause, such as `includeIf.all[0]` or `variables[1].askIf.all[2]`. The model cannot tell an absent `enabled` from `false`, and reading it so would switch a section off by accident. A lone `section: access` is a problem. A `~` is the null spelling alone and is not also a problem of the clause. When the clause and its first key are one place, the problem of the clause comes before the other problems of that place, as in the decoder: before a `format.unknown-field`, and before a `format.wrong-type` at that key. See open question 6.
- An empty `section`, whether `""`, `''`, or null, is read as the empty string, and `enabled` is read as written. The loader does not check the section id, as it does not check a question id.
- The two forms are the same in `includeIf` and in `askIf`.
- An item of `all` that is not a map, whether plain text or a list, is `format.wrong-type` at the item, such as `includeIf.all[0]` or `variables[0].askIf.all[1]`. An `answer` that is not a map is `format.wrong-type` at `answer`. An `all` that is not a list is `format.wrong-type` at `all`. Nothing inside any of them is read, and the loader reads the other items as usual. The decoder gives the same codes and paths.

## Reading a parent

`parent` is the id of the section that this one is a sub-section of (0070). A template is a section copied into a plan, so a template that names a parent is allowed. The loader reads `parent` exactly as it reads `title`: a plain or quoted string, with the spelling, wrong-type, anchor, duplicate-key, and empty-value rules of a text field and the same codes and paths, such as `parent`. A null, empty, or absent `parent` is the empty string, which is a root section. The text is not trimmed, resolved, or checked: `access` and `$parent` load as written, and a plain `deferred` is text. The loader does not check that the parent exists, that it is not the section itself, or that no chain of parents loops. A template alone cannot know. The plan validator checks them.

## Reading the file

### What is read as the decoder reads it

Style is free: comments, a leading `---`, any indent and key order, flow style, quotes, folded and literal blocks, LF or CRLF, a byte order mark, quoted keys. A plain `true` or `false` is a boolean, the other boolean spellings are problems, and so are the number and null spellings. A quoted scalar is always text. A plain date is text. A key is its text, and a key that is not a scalar is `format.wrong-type`. A null or absent field is the zero value, a list with no items is nil, and a map that is present is not nil. A step's `env` map is nil when absent or null, and a null value in it is the empty string. A present `includeIf` or `askIf` is a Condition even when it holds no clauses. An `isSettled` that is null is unset. A list keeps its items in file order and keeps repeats.

### Values

A Value is read in a `default`, in `is` and `isNot`, and in each item of `isOneOf` and `isNotOneOf`, as the decoder reads it. The plain word `deferred`, as a whole plain Value, is Deferred. In any other place it is text: a `title`, an item of a list of strings, an item of a List Value, a quoted `"deferred"`.

### Placeholders

A `$name`, `$$`, or other dollar sign in a string is text, kept as written. The loader does not resolve, check, or unescape it. The section resolver and the section validator do that. `$Repo_url`, `$5`, and `${x}` all load unchanged.

`run` is always a list of strings, in a step and in a check. A scalar there is `format.wrong-type` at the field, whatever it says: `run: make test`, `run: $name`, and `run: "$name"` alike. The loader does not split it and does not read it as one item. A `command` variable is written `run: [$name]`, which loads as a list of one item, the text `$name` (steps.md).

## What passes through

The loader judges the file's format and not its meaning. The section validator and the plan validator judge meaning, and `onboard-kit lint` runs them. These all load unchanged and without a problem:

- A missing `id`, `title`, or `summary`, and an empty string for any of them. A file with only `schemaVersion` is the zero Section.
- An `id` that is not a slug, or is `plan` or a reserved word. A variable `name` that breaks `[a-z][a-z0-9_]*` or is reserved. Two variables with one name, and two steps with one id.
- A `timing`, `type`, or `audience` that is not in its list. A `type` that is absent stays empty, a `timing` stays empty, and an `audience` stays empty. The loader fills no default.
- A `multiple` on a type that does not allow it, a `default` that does not fit its type, options, or pattern, and a `pattern` that is not valid RE2.
- A variable named `enabled`, of any type. It is an ordinary Variable, with no special reading. The name is reserved (0050, reserved names), and reserved-name checks are the job of the section validator and the plan validator, not the loader. A clause with `question: enabled` is an ordinary question clause.
- A section clause that names a section that does not exist, or the section itself. The plan validator reports it.
- A condition that cannot do its job: a single value in `isOneOf`, tests that cannot hold together, a clause naming a question that does not exist, an `includeIf` that names a section question, and a `askIf` with no clauses.
- A check with none of `run`, `ask`, `env`, or with more than one. A step `env` name that is not uppercase.
- An `after` or `escalation` that names a section or variable that does not exist, or the section itself.
- A secret in any value. The secrets check judges it.

After loading, `includeIf` and `askIf` are `*Condition` values with their clauses in file order. A section clause is a Clause with `Section` and `Enabled` and an empty `Question` and `Answer`. A variable named `enabled` is in `Variables`, in file order, with the `Default` Value that the file gives. Nothing is resolved, ordered, or selected.

## Properties

These hold for every file the contract tests use.

- **Deterministic.** Two calls with the same bytes give the same result.
- **Inputs unchanged.** The bytes are the same after the call as before, with or without a byte order mark.
- **Nothing shared.** Changing a returned Section changes no other result.
- **Whole or nothing.** A problem means the zero Section. A Section means no problems.
- **No source, no unknown fields, no switch.** A loaded Section has a nil `Source`, nil `UnknownFields`, and a nil `Enabled`.

## Not covered

- Reading files, which the driver does, and the folder layout of a template set.
- Checking that a `parent` names a section of the plan, that it is not the section itself, and that no chain of parents loops. That is the structure check of the plan validator, a later contract.
- Checking that a template is valid, which is the section validator. Checking the set of templates is the plan validator on the default plan, and style rules are the template linter.
- Warnings. The loader reports errors only.
- Migrating an older version. The migrator is deferred until the first breaking change. See 0030.
- Line and column numbers in a problem. The path names the place.
- `%YAML` directives, a document end marker `...`, and escaping in a path, as in the decoder contract.
- A plain scalar that ends in `?` inside a flow list or map, such as `{ask: Done?}`. One YAML library cannot read it. Quote it.
- A number too small for a 64-bit float, as in the decoder contract.
- The rest of the decoder's edge rules are inherited by reference and are tested only where a case below names them. The hardening pass finds which of them need a case of their own for a template.

## Open questions

Questions 2 to 4 are settled above so that the tests can run. A person decides whether the choice stands.

1. **Decided: scalar `run`.** A person decided that `run` stays a list and a scalar is refused. The source is the schema reference, [steps](../../../docs/reference/section-template-schema/steps.md), which writes `run: [$test_command]`. The ACME templates follow it.
2. **A `source` key.** A template with a `source` is `format.unknown-field`. Another reading is a code of its own, such as `template.has-source`. The first needs no new code and keeps one rule for every unknown key.
3. **The file code.** `file.not-a-template` is new. The decoder's `file.not-a-plan` names a plan.
4. **`x-` keys.** The decoder keeps them on a plan and a section. A template rejects them. A person may want a place for author notes in a template, which would be an ADR 0030 change and a version bump.
5. **Mixed clause.** The decoder contract has no code for a clause that holds keys of both forms. The loader follows it: the key of the other form is `format.unknown-field` at the key. A person may want a code of its own, such as `format.mixed-clause`. It would be a change to the decoder and the loader together.
6. **Absent `enabled`.** The decoder contract refuses a section clause with no `enabled` as `format.wrong-type` at the clause, because `Clause.Enabled` cannot tell absent from `false`. The loader follows it. A person may prefer a model change that keeps the difference, such as a `*bool`, which would change both contracts.

## Bootstrap

Under [0250](../../../docs/decisions/0250-bootstrap-content-and-projection.md),
bootstrap is also a known optional template boolean, copied into Section.Bootstrap.
Absent, empty, null and false are false; only canonical plain true is true.
Apply the existing boolean spelling/type, tag/style, duplicate-key, alias and
source-order problem rules. Never infer it from timing, source, default answers,
parent or descendants. `Bootstrap` with a capital B, other case variants, and
a step's `bootstrap` remain unknown keys; the exact lowercase root key is the
template marker only.
The [bootstrap fixtures](testdata/bootstrap.json) and [tests](bootstrap_test.go)
pin representational reading independently of product implementation.

## Requires

[0260](../../../docs/decisions/0260-prerequisites-and-action-order.md) makes
requires a known optional template list of strings, read into Requires.
Apply the decoder's Requires list rules at template-relative paths: keep exact
text, list order and repetitions, and read absent, null, empty values and empty
lists as nil. Do not convert After, resolve placeholders or validate the
referenced sections. Case variants and a step's requires remain unknown keys.
Each raw repeated key is an error even when values agree. Independent
[fixtures](testdata/requires.json) and [tests](requires_test.go) pin the reading,
all-item guards, problem paths/order, separate-result isolation and state reuse.

## Wrong implementations tried

Each row names a plausible wrong implementation and the groups of cases that fail it. How the contract was checked, and with what scores, is recorded in the issue, not here.

| Wrong implementation | Caught by |
| -- | -- |
| Ignore Requires, sort or deduplicate it, convert After, skip list/item guards, or accept a raw repeated key | [Requires fixtures](testdata/requires.json) and [tests](requires_test.go) |
| Skip an anchor check on the root mapping or the supported schemaVersion value | anchored root and version [bootstrap fixtures](testdata/bootstrap.json) |
| Drop the known Bootstrap value or accept quoted true as a boolean | [bootstrap fixtures](testdata/bootstrap.json) |
| An unknown key kept or ignored at the top level, or only below it, or only at the top | `unknown-field-top`, `unknown-field-everywhere`, `unknown-field-repeated`, `many-problems` |
| An unknown key accepted in a variable, condition, answer test, step, or check, or `schemaVersion` known inside a step | `unknown-field-everywhere` |
| An `x-` key accepted | `unknown-field-top`, `unknown-field-repeated`, `many-problems` |
| An `enabled` key accepted or read on the template, a variable, a step, or a check, its value read, or `Enabled` set to true or false on a loaded Section | `template-enabled-key`, the no-switch test, `all-fields` |
| A section clause not read in `includeIf`, or not in `askIf`, or read as an empty question clause | `section-clauses`, `all-fields` |
| A section clause with `enabled: false` dropped, or `enabled` read from the wrong clause or the wrong key order | `section-clauses`, `all-fields` |
| An empty or null `section` refused, or read as a question clause, or a section clause told apart by its first key | `section-clauses`, `all-fields` |
| A key of the other form read as a form of its own: `question` or `answer` in a section clause kept, or reported by another code, such as `format.wrong-type` at the clause | `section-clause-mixed` |
| A mixed clause skipped whole, or its other keys not read, or the value of the unknown key read for problems | `section-clause-mixed`, `section-clause-wrong-type` |
| `enabled` in a clause with no `section` accepted or ignored, or read as the start of a section clause | `section-clause-mixed` |
| A section clause with no `enabled` read as `false`, or the problem at the key and not at the clause | `section-clause-wrong-type` |
| An empty or null `enabled` accepted, or a `~` reported twice, or a wrong-typed `enabled` reported again at the clause | `section-clause-wrong-type` |
| The problem of the clause ordered after the problem of its first key | `section-clause-wrong-type`, `section-clause-mixed` |
| A section clause with `enabled: yes`, `1`, `"true"`, a list, or a map accepted, or `section` of the wrong kind accepted, or an extra key in it accepted | `section-clause-wrong-type` |
| A variable named `enabled` refused, or a question clause on `enabled` refused | `unvalidated`, `all-fields` |
| A `source` accepted silently, or read into the Section | `source-key`, `unknown-field-repeated`, the no-source test |
| A plan file's `provenance`, `answers`, or `sections` accepted | `plan-as-template` |
| The value of an unknown key read for problems | `unknown-field-top`, `unknown-field-repeated`, `many-problems` |
| A repeated unknown key reported once | `unknown-field-repeated` |
| A quoted `"<<"` read as a merge key | `unknown-field-top` |
| `yes`, `on`, `True`, or `TRUE` read as a boolean, or not reported | `boolean-spelling`, `same-place`, `tags` |
| A quoted `"yes"`, `"1.0"`, or `"~"` checked as a spelling | `strings`, `style-is-free` |
| `1.0`, `007`, `0x1F`, `1e3`, or a canonical number in a string field accepted, or a date refused | `number-spelling`, `tags`, `strings` |
| `~` or `Null` read as null, or `nULL` flagged | `null-spelling`, `strings` |
| Anchors, aliases, or merge keys accepted | `alias`, `same-place` |
| A duplicate key accepted, or `"K"` and `K` told apart | `duplicate-key`, `unknown-field-repeated` |
| A tag changing how a scalar is read | `tags` |
| A key that is not a scalar skipped | `complex-key` |
| Defaults filled for `type`, `timing`, or `audience` | every case with a variable or a section that omits them |
| The template validated: a missing `id` or `title`, an unknown `type` or `timing`, `id: plan`, an `after` naming the section, a bad variable name, two steps with one id, a check with two keys, an `includeIf` on a section question | `version-only`, `minimal`, `unvalidated` |
| `$$` unescaped, a `$name` resolved or refused, or a trailing newline trimmed | `placeholders`, `acme-clone-and-run-tests`, `acme-access`, `strings` |
| Text trimmed, such as an owner or a directory with spaces | `strings` |
| Quoted `"deferred"` read as Deferred, a plain one read as text in a Value, or read as Deferred or dropped in a string or a list | `deferred`, `all-fields` |
| `example` not read, or reported as an unknown field | `all-fields`, `strings` |
| `example` read into `description` or `default`, or `description` read into `example`, or `example` used when `description` is empty | `all-fields`, `strings` |
| `example` trimmed, resolved as a placeholder, or its text changed | `strings`, `placeholders` |
| `example` not checked as a text field: a spelling not reported, a list, a map, or a plain `true` accepted, or a null read as the text `null` | `boolean-spelling`, `number-spelling`, `null-spelling`, `wrong-type-values`, `null-and-empty` |
| `parent` not read, or reported as an unknown field | `sub-section`, `all-fields` |
| `parent` read into `title` or `after`, or kept only when a section has an `includeIf`, or an `includeIf` question read into `parent` | `sub-section`, `all-fields` |
| `parent` not checked as a text field: a spelling not reported, a list, a map, or a plain `true` accepted, an anchor not reported, a null read as the text `null`, or a repeated key accepted | `boolean-spelling`, `number-spelling`, `null-spelling`, `alias`, `tags`, `wrong-type`, `duplicate-key`, `parent-null` |
| A null or empty `parent` read as text, or refused | `null-and-empty`, `parent-null`, `parent-quoted-empty` |
| `parent` trimmed, resolved as a placeholder, or a plain `deferred` read as Deferred | `strings`, `placeholders`, `deferred` |
| A `parent` that names no section, or the section itself, refused or dropped | `unvalidated`, `parent-names-itself` |
| A scalar `run` accepted, read as one item, or split on spaces | `run-scalar`, `wrong-type` |
| A scalar `run` refused only in a step, or only in a check | `run-scalar` |
| A scalar `run` refused where a list of one `$name` must load | `acme-clone-and-run-tests`, `placeholders` |
| A bad version accepted: quoted, decimal, missing, zero read as valid or as invalid, a newer one read for other problems or accepted, a version detail missing, the last of two keys winning | the `version-` cases and `duplicate-version` |
| The plan file code for a non-template, an empty file giving the zero Section, a list of sections read as one, only the first document read, a syntax error in a later document hidden | the file cases |
| A nil list or map made empty or the reverse, a null `includeIf` made empty, a present empty `askIf` made none, a null `isSettled` made false, a null `default` made empty text, an empty List Value made none | `null-and-empty` |
| A null item dropped from a list, a null `isOneOf` item reported as an empty item | `null-items` |
| A plain text or a list where a clause or an answer test belongs crashes the loader or is skipped without a problem, in `includeIf` or in `askIf`, or a plain text where `all` belongs | `wrong-type-items` |
| A map, or a list holding a map, scanned for more problems or read item by item | `wrong-type-values`, `wrong-type` |
| A plain `true` accepted in a string field, a quoted `"true"` accepted as a boolean, a wrong type in a list item skipped | `wrong-type`, `wrong-type-values`, `many-problems` |
| A list sorted, reversed, or stripped of repeats, in `after`, `options`, `steps`, or `variables` | `list-order` and the ACME cases |
| Problems sorted by path, in the order found, or with the order at one place reversed | `many-problems`, `same-place`, `unknown-field-repeated` |
| Only the first problem reported, or a partly filled Section returned with problems | every problem case |
| A problem of severity `warning`, with no message, or with no `path` | every problem case |
| The input bytes changed, including a byte order mark stripped in place or carriage returns replaced | the inputs-unchanged test, the Windows line endings test |
| State kept between calls | the results-share-nothing test |
