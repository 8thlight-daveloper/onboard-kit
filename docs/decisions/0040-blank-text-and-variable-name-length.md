---
category: behavior
status: accepted
approval: unrecorded
affects: [plan/authoringvalidator, plan/editor, plan/validator, section/resolver]
builds-on: ['0035']
---

# 0040. Blank text and variable name length

Builds on [0035](0035-settled-and-deferred-answers.md). The name rule it changes is in the [domain model](../reference/domain-model.md).

## Context

Two small text rules were left open, and the section resolver build showed both.

- Text with only spaces is not told apart from missing text in some places and is in others. A title of three spaces passes a "has at least one character" check and shows as nothing in the document. The model has no absent text, only the empty string, so absent and empty are already the same value.
- A variable name may be one letter, as in `$x`. The resolver reads `$` and the longest run of name characters, so a one-letter name sits next to prose such as `$5` and `$a` in a price or a regex. Short names also give weak signals to a person reading a template. No template under `docs/` uses a name shorter than three characters. Many contract fixtures do, as short placeholders such as `a` and `v`.

## Decision

### Blank text

**Blank** means nothing, or only whitespace. Whitespace is what Unicode calls space, the same as the secrets check. A person never has to write the word "empty" for this.

Blank applies to prose fields. These are the fields a person reads as sentences.

| Place | Fields |
| -- | -- |
| Section | `title`, `summary`, `body`, `verification` |
| Step | `name`, `notes` |
| Variable | `description`, `patternHint` |

Rules:

- A required prose field that is blank is treated as missing. `title` and `summary` are the required ones.
- An optional prose field that is blank is treated as absent. The decoder keeps the text as written. The checks and the hasher read it as absent.
- The hasher already normalizes prose whitespace, so a blank prose field and an absent one give one hash.

Blank does not apply where the exact text carries meaning:

- `run` items. An empty string is a real argument.
- `env` values. An empty value is how a step clears a variable.
- A variable `default`, an answer, and `pattern`. A single space is a legitimate answer or separator. A settled text answer still needs at least one character, as in 0035.
- Ids, names, `owner`, `open`, and `directory`. Their own rules reject the empty string and whitespace already.

### Variable name length

A variable name is a lowercase letter followed by at least two more lowercase letters, digits, or underscores. The pattern becomes `[a-z][a-z0-9_]{2,}`.

- The rule applies to a variable declaration in a section template or a plan. The lint and the section validator report a shorter name.
- A `$` followed by a one or two character run, such as `$x` or `$ab`, is not a variable. The resolver leaves it as written, with no problem, as it does for `$5`. A longer run is read as before, and the name still takes every character it can.
- Ids, section ids, and step ids keep their own slug rule. This ADR does not change them.
- Reserved names stay reserved. Any reserved name shorter than three characters becomes unreachable and drops out of the list.

## Why

- **Blank.** The failure is silent: a person enters a space and a section shows an empty title. Treating it as missing gives the person a message at the point of entry. The exclusions keep the cases where whitespace is data.
- **Three characters.** A rule that removes `$x` and `$ab` costs a template author little, since a name is a label for a person. It removes the confusion with `$5` and with shell and regex text, where one and two letter names are common.
- **Versioning.** As in 0035, nothing exists outside this repository, so the schema version stays 1. After the first release, a rule like this needs a version bump and a step in the plan migrator.

## Consequences

- The domain model changes: the name pattern, a definition of blank, and the list of prose fields. This ADR authorizes that edit.
- Contract fixtures that use a name shorter than three characters change wherever a component checks names: the section loader, the section resolver, and the validators. The plan decoder, plan encoder, and section hasher do not check names, so their fixtures keep short names.
- Contracts that take small increments, each its own issue:
  - Decoder: no change. It keeps text as written.
  - Section loader: a variable name of one or two characters is `format.invalid-name`, and a blank required prose field is a missing field.
  - Section hasher: a blank prose field hashes as an absent one.
  - Section resolver: `$x` and `$ab` are left as written, and a one or two character declared name is outside what it must handle.
  - Plan validator and section validator: both rules apply, once those contracts exist.
- The schema reference in `docs/reference/section-template-schema/` takes the new pattern and the definition of blank.

## Not decided

- Whether the TUI blocks a blank entry while typing or reports it on save.
- Whether a maximum name length is wanted.
