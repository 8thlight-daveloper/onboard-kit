# Document assembler contract

Status: draft by an agent. Not reviewed by a person. Component: `document/assembler`. Issue: [#69](https://github.com/8thlight-daveloper/onboard-kit/issues/69).

The document assembler turns a plan into a Markdown document. It asks the section selector for the active sections in reading order, asks the section resolver to fill each one, and writes a title, a review line, a gap table, and each section. It is a pure function. It reads no file, reads no clock, keeps no state between calls, and does not change its argument. See [0030](../../../docs/decisions/0030-self-contained-plan-file.md), [0010](../../../docs/decisions/0010-plan-editor.md), the [domain model](../../../docs/reference/domain-model.md), [variables](../../../docs/reference/section-template-schema/variables.md), [steps](../../../docs/reference/section-template-schema/steps.md), and [0015](../../../docs/decisions/0015-functional-core-and-imperative-shell.md).

This contract is transitional and kept thin, as [0025](../../../docs/decisions/0025-contracts-and-tests.md) asks. The gap table, the review line, and the output style exist only in the assembled document. Nothing else depends on them. The only output style is Markdown.

## Interface

`Assemble(plan) -> (document, problems)`.

- `plan` is a Plan, as stored. The assembler reads `provenance`, `answers`, and `sections`, and the `variables` a section declares. It reads no other field.
- `document` is text, in Markdown.
- `problems` is the gaps of the document, as the section resolver reported them. They are listed in the gap table and returned unchanged, in the same order. It is empty when there are none.

The assembler does not fail. Any plan gives a document. It does not check that the plan is valid or that its required questions are answered. A driver does that first, with the plan validator and the question finder, as [0010](../../../docs/decisions/0010-plan-editor.md) says.

## How the component gets the selector and the resolver

The assembler may depend on the section selector, the section resolver, and the model, and on nothing else in the project ([0020](../../../docs/decisions/0020-binary-structure-and-regenerative-components.md)). A contract folder may import only the model, the shared test code, and its own component, in its wiring test. So the constructor takes nothing, `New()`, and the component calls the real selector and the real resolver itself. The contract tests run against them and not against stand-ins. What the two do is as their own contracts say: [section selector](../../section/selector/contract.md) and [section resolver](../../section/resolver/contract.md). The assembler does not redo their work. It does not decide which sections are active, in which order, or how a `$name` is filled.

## Rules

The document is these blocks, in this order, each followed by one blank line, except the last. A line break ends the document, and no more than one.

1. The title.
2. The review line.
3. The gap table.
4. Each active section, in the order the selector gives.

Lines are not wrapped. A paragraph, a list item, and a table row is each one line, except for the line breaks held inside a prose field.

### Title

`# ` and then the title, from the plan's own answers: `answers.plan.organization` and `answers.plan.team`. An answer counts when it is a Text with at least one character. Anything else counts as absent: nothing stored, empty text, Deferred, a Flag, or a List.

| `organization` | `team` | Title |
| -- | -- | -- |
| present | present | `<organization> onboarding: <team>` |
| present | absent | `<organization> onboarding` |
| absent | present | `Onboarding: <team>` |
| absent | absent | `Onboarding` |

The text is written as held.

### Review line

`Generated <generatedOn> from template set `<templateSet>`. A person must review this before it is sent.`

`generatedOn` and `templateSet` come from `provenance` as held. The assembler reads no clock.

### Gap table

The heading `## Gaps to fill`, a blank line, and a table.

- The header row is `| Gap | Who fills it | By when |` and the second row is `| --- | --- | --- |`. The table is always there, and it has no other row when there are no gaps.
- There is one row for each problem the section resolver returned, in this order: the sections in the selector's order, and within a section the resolver's order. A deferred and an unanswered variable look alike.
- A row is `| ` and the three cells joined by ` | ` and then ` |`. A cell may be empty.
- **Gap** is the `description` of the variable the problem names, from the variables of the section. A problem's `question` is `<section id>.<name>`. A variable with a blank `description`, or one that is not found, gives the name.
- **Who fills it** is the section's `owner`, after the resolver has filled it. It is empty when the section has no owner.
- **By when** is the section's timing in words, with the first letter in capitals: see Timing in words. A section with no `timing` is `Anytime`.
- In a cell, a `|` is written `\|` and every line break is written as one space.

### Sections

Each active section is these blocks, in this order, from the section as the resolver returned it. A block with nothing in it is left out, with its blank line.

1. `## ` and the `title`.
2. The `summary`.
3. The steps, as a list, see Steps.
4. The `body`.
5. `Done when: ` and the `verification`.
6. The footer.

A section that is not active is not in the document, and its gaps are not in the table. The assembler does not group sections or put a heading for a timing. The order is the selector's.

**Footer.** The footer is up to two parts joined by one space: `Owner: <owner>.` when the section has an owner, and `When: <timing in words>.` when the section has a `timing` that is not `anytime`. With neither part, there is no footer.

### Steps

A step is one list item: `- [ ] **<name>.**` and then these parts, in this order, each when the step has it, each ending in a full stop except where noted, all joined by one space. The items follow each other with no blank line.

| Part | Written as |
| -- | -- |
| `owner` | `Owner: <owner>.` |
| `timing` | `When: <timing in words>.`, only when the step has a `timing` and it is not the section's. A step with no `timing` has its section's. A section with no `timing` has `anytime` |
| `notes` | The notes, as held, with no full stop added |
| `open` | `Open: <open>.` |
| `directory` | ``Directory: `<directory>`.`` |
| `run` | ``Run: `<items joined by one space>`.`` |
| `env` | ``Env: `<name>=<value>`, `<name>=<value>`.``, in the order of the names, compared by bytes |
| each check with `run` | ``Check: `<items joined by one space>` exits with code 0.`` |
| each check with `ask` | `Check: <ask>` |
| each check with `env` | ``Check: `<name>` is set and not empty.`` |

A step with `run` and no checks gets no check line. The checks follow each other in the step's order. A step with no part has only its name. Every line of an item after the first starts with two spaces, and a blank line in an item stays empty, so a multi-line `notes` stays in its item. `escalation` and the `after` of a section are not written.

### Timing in words

The name of the timing with each `-` written as a space: `before day one`, `week one`, `days 8 to 30`, `days 31 to 90`, `anytime`. In the gap table the first letter is a capital.

### Prose is written as held

`summary`, `body`, `verification`, `notes`, and the other text fields are written as the resolver returned them: with their line breaks, and without wrapping, trimming, or changing a character. The only change is that space and line breaks at the end of a field are removed. A field with no character other than white space is left out. A field is not read again as a template, so a `$` in it is written as it is.

### Markdown

The assembler escapes very little. A person writes Markdown in the plan and the document is Markdown. These are the whole of it.

- **A blank for an onboarding variable.** The resolver writes an onboarding variable that has no value as its name between `<` and `>`, and Markdown reads that as a tag and hides it. In the text fields `title`, `summary`, `owner`, `verification`, `body`, a step's `name`, `owner`, `notes`, `open`, and a check's `ask`, the assembler writes every occurrence of that blank, for each onboarding variable the section declares, between backticks. It does not do this in a code span the assembler writes, such as `run`, `directory`, and `env`, where the blank shows as it is. Other text between `<` and `>` is not touched. The assembler finds the blank by the name of the variable and does not look at the wording of any other marker. The resolver's markers for a deferred and an unanswered variable are written as they are.
- **The table.** A cell escapes `|` and line breaks, as above.
- **Code spans.** The template set, a command, a directory, and an environment variable are each written between single backticks, with the text as held. Text that holds a backtick is not covered.

## Properties

These hold for every plan the contract tests use.

- **Pure.** The plan is the same after the call as before. Changing the returned problems changes nothing in the plan.
- **No memory.** The same plan gives the same document and problems on a shared assembler, whatever calls came before, forward or in reverse order.
- **Same plan, same bytes.** Rotating or reversing `plan.Sections` does not change the document.
- **One row for each problem.** The gap table has as many rows as the problems returned.
- **Ends with one line break,** and holds no carriage return.
- **The zero plan** gives a document and no problem.

## Side effects and plan access

None. It reads `provenance`, `answers`, and section content. It writes nothing but the returned text and problems. This matches the row for `document/assembler` in [components](../../../docs/architecture/components.md).

## Not covered

- Whether the plan is valid: required questions, types, options, patterns, timing conflicts, cycles, and unknown references. The validators and the finder report them. A driver stops before assembling when they do.
- An empty or missing `generatedOn` or `templateSet`.
- A title in which `organization` or `team` holds a line break.
- Text that holds a backtick where the assembler writes a code span, a `|` in a heading, and a name or a title that ends in a full stop.
- A `timing` that is not one of the five names, and a step with its own `timing` that is not one of them.
- Escaping any other Markdown in a field, such as `*`, `_`, or a tag. The one exception is the blank for an onboarding variable.
- An onboarding blank that sits inside a code span a template author wrote.
- What the selector and the resolver do. Their contracts cover it, and the assembler tests rely on it. That includes how a variable with a `default` is filled, see [0045](../../../docs/decisions/0045-defaults-are-suggestions.md). The fixtures here declare no default.
- Escalation. It is not written, because it names people by variable and no document rule says how to show them.
- Output styles other than Markdown, and wrapping lines at a width.
- The format of the plan file, and the file the document is written to. Drivers do both.

## Decisions to know

The docs were silent, or could be read two ways, in these places. A choice marked `hard` changes what a template author writes, what a plan holds, or what a person sees. A choice marked `cheap` touches one fixture and nothing outside the component.

1. **The ACME document in the docs folder is not the golden. This folder has its own.** [`acme.md`](../../../docs/reference/examples/acme/acme.md) was written by hand and no single set of rules produces it. It treats like things differently, with nothing in the plan to tell them apart: the `access` section shows its `body` first and no `summary`, while the other two sections show the `summary` first and the `body` after the steps. A step owner is "Granted by", which suits a grantor and no other owner. "Install dependencies" and "Run the tests" have the same shape, and only the second says "You know it worked when the command exits with code 0". A check `run [test, -d, x]` is "`x` exists", and a check `ask` "Can you open X?" is "you can open X", which no general rule does. And "`make test` passes" has backticks that the resolver's text, `make test passes`, does not carry. The contract writes one set of rules, and `testdata/acme.md` follows them for the same plan. The plan is the ACME plan, as the decoder and encoder contracts check it. To go back, change the docs file to match `testdata/acme.md`, or the rules to match the docs file, and point the case `acme` at it. Raised by "`acme.md` ... written by hand" and "`acme.md` ... is a transitional golden". `hard`
2. **A section is its summary, steps, body, verification, and footer, in that order.** The `summary` is shown. Raised by the ACME document, where two of three sections do this. Case: `acme`, `minimal-section`. `hard`
3. **The words of the labels.** A step has `Owner:`, `When:`, `Open:`, `Directory:`, `Run:`, `Env:`, and `Check:`. The verification is `Done when:` and its text as held, with no change of letter case. The footer is `Owner:` and `When:`, as in the ACME document. Raised by the ACME document's longer sentences, which the plan data cannot fill in generally. Cases: `steps-all-fields`, `acme`, `footers`. `hard`
4. **The gap table has the owner and timing of the section, not of the step that uses the variable.** A variable used in a step's `owner`, such as `tracker_grantor`, is still a row of its section, with the section's owner and timing. The table has no column for deferred or unanswered. Raised by "It renders as a visible gap with an owner and timing". Cases: `acme`, `gaps-deferred`, `gap-timings`. `hard`
5. **The gap table is always there, and empty when there are no gaps.** Raised by "Every assembled document carries a gap table with owner and timing". Case: `empty-plan`, `minimal-section`. `hard`
6. **A gap with no owner has an empty cell.** No wording is invented. A variable with no description shows its name. Raised by the table's two columns, with no rule for a section without an owner. Case: `gaps-unanswered`. `cheap`
7. **Lines are not wrapped.** The docs file wraps at about 86 characters, but one width does not give it: one line of it is 87 characters. Markdown joins the lines of a paragraph, so a wrap only changes the file and not the page. A template's own line breaks in `body` and `notes` are kept. Raised by "Prose fields hold Markdown ... hard-wrapped", which is about the template. Case: `acme`. `hard`
8. **Markdown is escaped in three places only.** The blank of an onboarding variable goes between backticks, a table cell escapes `|` and line breaks, and nothing else is escaped, so `*`, `_`, and tags in a plan are Markdown in the document. The owner in a footer is not escaped for `|`. Raised by "The assembled document shows it as a blank, `<name>`" and the resolver contract, which leaves escaping to the assembler. Cases: `onboarding-blank`, `table-cell-escaping`, `prose-as-held`. `hard`
9. **The title is `<organization> onboarding: <team>` from the plan's answers.** An answer that is not a settled Text leaves its part out, and `Onboarding` stands when both are. The assembler does not require the plan to declare the variables. Raised by the ACME title, with no rule given. Cases: `acme`, `title-organization-only`, `title-team-only`, `title-empty-text`, `empty-plan`. `hard`
10. **The footer parts stand alone, and `anytime` is not shown.** A section with an owner and no timing has `Owner:` only, and one with a timing and no owner has `When:` only. A section with neither has no footer. Raised by the ACME document, where `stuck-general` has `anytime` and no owner and no footer. Cases: `footers`, `gap-timings`. `hard`
11. **A step shows `When:` only when its timing differs from its section's.** Raised by "A step inherits `timing` from its section, when assembling". Case: `steps-all-fields`. `cheap`
12. **A step with `run` and no checks gets no check line.** The steps reference says it passes on exit code 0, and the document does not say so. Raised by the unequal treatment of two like steps in the ACME document. Cases: `acme`. `cheap`
13. **The assembler gets the selector and resolver itself, with `New()` and no arguments.** The contract tests run against the real ones. Raised by the dependency table in 0020, which lets the assembler use both, and by the rule that a contract may import only the model, the shared test code, and its own component. `cheap`
14. **The problems returned are the resolver's, unchanged and in table order.** The assembler adds none, and no type is added to the model. A document is a string. Raised by "document text + problems", with no list of problems given. Case: all. `cheap`

## Wrong implementations tried

Each row names a plausible wrong implementation and the cases that fail it.

| Wrong implementation | Caught by |
| -- | -- |
| An inactive section is written, or its gaps are listed | `inactive-section`, `acme` |
| Sections in the plan's order, or grouped under a timing heading | `reading-order`, `footers`, `acme` |
| A `$name` is left unfilled, or the resolver is not used | `acme`, `gaps-deferred` |
| The gap table is left out when there are no gaps | `empty-plan`, `minimal-section` |
| Only deferred answers are listed, or only unanswered ones | `gaps-deferred`, `gaps-unanswered` |
| Gap rows sorted, or one row for each use of a variable | `gaps-unanswered` |
| The owner or timing in a row is the step's, or is missing, or the timing is not in words | `acme`, `gap-timings` |
| An empty owner shows a word, or the row is left out | `gaps-unanswered` |
| `|` or a line break in a cell is not escaped | `table-cell-escaping` |
| The onboarding blank is not in backticks, or `<note>` is too | `onboarding-blank` |
| The blank is also wrapped in code spans, or only in `body` | `onboarding-blank` |
| `anytime` shown in a footer, or a footer with only one part left out | `footers`, `gap-timings` |
| The summary left out, or the body before the steps | `acme`, `minimal-section` |
| `Done when:` left out, or a blank verification written | `acme`, `prose-as-held`, `onboarding-blank` |
| The body trimmed of its inner line breaks, or the final blank lines kept | `prose-as-held`, `acme` |
| Markdown in prose escaped | `prose-as-held` |
| Environment variables in map order | `steps-all-fields` |
| A step's own timing always shown, or never | `steps-all-fields` |
| A continuation line not indented, or a blank line indented | `steps-all-fields` |
| A part of a step left out, or a check worded differently | `steps-all-fields`, `acme`, `onboarding-blank` |
| A part of the title left out wrongly, or a deferred or empty answer used | `title-organization-only`, `title-team-only`, `title-empty-text`, `acme` |
| The date or the template set of the review line fixed in the code | `title-organization-only` |
| No line break at the end, or two | The test of the end of the document |
| The plan changed, or its sections sorted in place | The plan-unchanged test, the order test |
| A cache between calls | The no-memory test |
| The problems not returned, or their severity changed | `acme`, `gaps-deferred`, `gaps-unanswered`, `people-list` |
| A step with a command and no checks gets a check line | `acme` |
| The assembler panics on a plan with no sections | `empty-plan`, the zero plan test |
