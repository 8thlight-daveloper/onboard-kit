# Section loader contract

Status: draft accepted by a person. A hardening pass, #65, comes before the first build. Component: `section/loader`. Issue: [#64](https://github.com/8thlight-daveloper/onboard-kit/issues/64).

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
| Template | `schemaVersion`, `id`, `title`, `summary`, `includeIf`, `timing`, `after`, `owner`, `escalation`, `verification`, `body`, `variables`, `steps` |
| Variable | `name`, `description`, `type`, `multiple`, `options`, `pattern`, `patternHint`, `required`, `default`, `phase`, `askIf` |
| Condition | `all` |
| Clause | `question`, `answer` |
| Answer test | `is`, `isNot`, `isOneOf`, `isNotOneOf`, `isSettled` |
| Step | `id`, `name`, `notes`, `owner`, `timing`, `escalation`, `run`, `open`, `directory`, `env`, `checks` |
| Check | `run`, `ask`, `env` |

The `env` of a step is a map whose keys are the variable names it sets, so any scalar key is allowed there. Its values are text.

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
| `format.unknown-field` | A map has a key that the table above does not list | The key, such as `summery`, `source`, or `steps[0].extra` |

Problems come in the order of their places in the file, and all are reported. The detail of a problem names its place in `detail.path`. A file problem is the only problem, and a version problem is the only problem. The value of an unknown key is not read, so nothing inside it is reported. A repeated unknown key is reported each time, with `format.duplicate-key` for each repeat.

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
- A `timing`, `type`, or `phase` that is not in its list. A `type` that is absent stays empty, a `timing` stays empty, and a `phase` stays empty. The loader fills no default.
- A `multiple` on a type that does not allow it, a `default` that does not fit its type, options, or pattern, and a `pattern` that is not valid RE2.
- An `enabled` variable of any type. It is an ordinary Variable, with no special reading.
- A condition that cannot do its job: a single value in `isOneOf`, tests that cannot hold together, a clause naming a question that does not exist, an `includeIf` that names a section question, and a `askIf` with no clauses.
- A check with none of `run`, `ask`, `env`, or with more than one. A step `env` name that is not uppercase.
- An `after` or `escalation` that names a section or variable that does not exist, or the section itself.
- A secret in any value. The secrets check judges it.

After loading, `includeIf` and `askIf` are `*Condition` values with their clauses in file order. `enabled` is a Variable named `enabled` in `Variables`, in file order, with the `Default` Value that the file gives. Nothing is resolved, ordered, or selected.

## Properties

These hold for every file the contract tests use.

- **Deterministic.** Two calls with the same bytes give the same result.
- **Inputs unchanged.** The bytes are the same after the call as before, with or without a byte order mark.
- **Nothing shared.** Changing a returned Section changes no other result.
- **Whole or nothing.** A problem means the zero Section. A Section means no problems.
- **No source, no unknown fields.** A loaded Section has a nil `Source` and nil `UnknownFields`.

## Not covered

- Reading files, which the driver does, and the folder layout of a template set.
- Checking that a template is valid, which is the section validator. Checking the set of templates is the plan validator on the default plan, and style rules are the template linter.
- Warnings. The loader reports errors only.
- Migrating an older version. The migrator is deferred until the first breaking change. See 0030.
- Line and column numbers in a problem. The path names the place.
- `%YAML` directives, a document end marker `...`, and escaping in a path, as in the decoder contract.
- A plain scalar that ends in `?` inside a flow list or map, such as `{ask: Done?}`. One YAML library cannot read it. Quote it.
- A number too small for a 64-bit float, as in the decoder contract.
- The rest of the decoder's edge rules are inherited by reference and are tested only where a case below names them. The hardening pass, [#65](https://github.com/8thlight-daveloper/onboard-kit/issues/65), finds which of them need a case of their own for a template.

## Open questions

Questions 2 to 4 are settled above so that the tests can run. A person decides whether the choice stands.

1. **Decided: scalar `run`.** A person decided that `run` stays a list and a scalar is refused. The source is the schema reference, [steps](../../../docs/reference/section-template-schema/steps.md), which writes `run: [$test_command]`. The ACME templates follow it.
2. **A `source` key.** A template with a `source` is `format.unknown-field`. Another reading is a code of its own, such as `template.has-source`. The first needs no new code and keeps one rule for every unknown key.
3. **The file code.** `file.not-a-template` is new. The decoder's `file.not-a-plan` names a plan.
4. **`x-` keys.** The decoder keeps them on a plan and a section. A template rejects them. A person may want a place for author notes in a template, which would be an ADR 0030 change and a version bump.

## Wrong implementations tried

While drafting, a reference loader was written as scratch work outside the repo, from this page and the decoder contract, and then broken on purpose. 120 wrong loaders were tried, each with one defect. Every one failed at least one test. The first run of the reference found one fixture error, a plain `y` in an `isNotOneOf` list that is the boolean spelling. Later fixtures were added for the defects that first got through. After a person decided that `run` is always a list, the sugar rows were dropped and `run-scalar` was added. A case that disagreed with the reference's reading of a Value showed the reference was wrong and not the case. This is a draft check only. There was no second independent build, no mutation run, and no differential run. Those belong to the hardening pass, [#65](https://github.com/8thlight-daveloper/onboard-kit/issues/65).

| Wrong implementation | Caught by |
| -- | -- |
| An unknown key kept or ignored at the top level, or only below it, or only at the top | `unknown-field-top`, `unknown-field-everywhere`, `unknown-field-repeated`, `many-problems` |
| An unknown key accepted in a variable, condition, answer test, step, or check, or `schemaVersion` known inside a step | `unknown-field-everywhere` |
| An `x-` key accepted | `unknown-field-top`, `unknown-field-repeated`, `many-problems` |
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
| Defaults filled for `type`, `timing`, or `phase` | every case with a variable or a section that omits them |
| The template validated: a missing `id` or `title`, an unknown `type` or `timing`, `id: plan`, an `after` naming the section, a bad variable name, two steps with one id, a check with two keys, an `includeIf` on a section question | `version-only`, `minimal`, `unvalidated` |
| `$$` unescaped, a `$name` resolved or refused, or a trailing newline trimmed | `placeholders`, `acme-clone-and-run-tests`, `acme-access`, `strings` |
| Text trimmed, such as an owner or a directory with spaces | `strings` |
| Quoted `"deferred"` read as Deferred, a plain one read as text in a Value, or read as Deferred or dropped in a string or a list | `deferred`, `all-fields` |
| A scalar `run` accepted, read as one item, or split on spaces | `run-scalar`, `wrong-type` |
| A scalar `run` refused only in a step, or only in a check | `run-scalar` |
| A scalar `run` refused where a list of one `$name` must load | `acme-clone-and-run-tests`, `placeholders` |
| A bad version accepted: quoted, decimal, missing, zero read as valid or as invalid, a newer one read for other problems or accepted, a version detail missing, the last of two keys winning | the `version-` cases and `duplicate-version` |
| The plan file code for a non-template, an empty file giving the zero Section, a list of sections read as one, only the first document read, a syntax error in a later document hidden | the file cases |
| A nil list or map made empty or the reverse, a null `includeIf` made empty, a present empty `askIf` made none, a null `isSettled` made false, a null `default` made empty text, an empty List Value made none | `null-and-empty` |
| A null item dropped from a list, a null `isOneOf` item reported as an empty item | `null-items` |
| A map, or a list holding a map, scanned for more problems or read item by item | `wrong-type-values`, `wrong-type` |
| A plain `true` accepted in a string field, a quoted `"true"` accepted as a boolean, a wrong type in a list item skipped | `wrong-type`, `wrong-type-values`, `many-problems` |
| A list sorted, reversed, or stripped of repeats, in `after`, `options`, `steps`, or `variables` | `list-order` and the ACME cases |
| Problems sorted by path, in the order found, or with the order at one place reversed | `many-problems`, `same-place`, `unknown-field-repeated` |
| Only the first problem reported, or a partly filled Section returned with problems | every problem case |
| A problem of severity `warning`, with no message, or with no `path` | every problem case |
| The input bytes changed, including a byte order mark stripped in place or carriage returns replaced | the inputs-unchanged test, the Windows line endings test |
| State kept between calls | the results-share-nothing test |
