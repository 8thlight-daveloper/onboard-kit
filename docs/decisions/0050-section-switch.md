---
category: behavior
status: accepted
approval: unrecorded
affects: [condition/evaluator, document/assembler, plan/decoder, plan/editor, plan/instantiator, plan/refresher, plan/validator, question/finder, section/hasher, section/loader, section/resolver, section/selector]
builds-on: ['0030', '0045', '0010']
---

# 0050. Section switch

Builds on [0030](0030-self-contained-plan-file.md) and [0045](0045-defaults-are-suggestions.md). It replaces the `enabled` variable described in [optional sections](../reference/section-template-schema/optional-sections.md) and in [0010](0010-plan-editor.md).

## Context

A template can declare a variable named `enabled`, so that a section can be switched off. It is an ordinary boolean variable. That makes it a question with the id `<section>.enabled`, stored among the answers, able to block when it is required, and able to fall back to a default.

Switching a section off is not an answer to a question about the engagement. It is a setting that an organization changes to fit its own case. A template that was added is meant to be used, so it starts on.

## Decision

A section in a plan has a switch. It is a setting on the section, outside the questions and the answers.

- **A section is on unless the plan says otherwise.** The plan section carries an optional boolean `enabled`, true by default. A new section is on, and when the field is absent the section is on. The switch is optimistic: it is named for the state we expect. A template never declares the switch, and a template section is always on when it is first added.

- **Only a person turns a section off.** The editor offers it as an action on a section, for example a list of sections with a toggle. It is not a question, it never blocks, it cannot be deferred, and it has no default.

- **A template can no longer declare a variable named `enabled`.** The name stays reserved.

- **Turning a section off keeps its answers.** They stay stored and are not asked while the section is off. Turning it on again brings them back, so an accidental toggle loses nothing.

- **The section selector reads the switch.** A section is active when its `includeIf` holds, or it has none, and its switch is on.

- **A condition can test whether a section is on.** A question's `askIf` and a section's `includeIf` may depend on another section being on, so a follow-up question can depend on it. The test is its own clause, separate from a question clause, because the switch is not a question:

  ```yaml
  askIf:
    all:
      - section: design-review
        enabled: true
  ```

  - `section` is the id of another section in the plan. `enabled` is `true` or `false`.
  - The clause tests the switch only. It does not ask whether that section's `includeIf` holds, so a chain of sections cannot make a cycle through the switches.
  - A clause with `section` has no `question` or `answer`. A clause with both kinds of key is a format problem.
  - The condition evaluator receives the switches along with the answers, and reads nothing else about the plan.

- **The hasher leaves the switch out.** It is the plan's own setting and not part of what a template says, so flipping it does not change the content hash, the same as `source`.

## Why

- A switch is a configuration choice, so questions, deferral, defaults, and blockers have no meaning for it.
- The plan file shows which sections an organization turned off, and that is all it shows. Nothing else depends on a template default.
- Keeping the answers makes the switch safe to flip.

## Consequences

- **The domain model changes.** The Section record in a plan gains an optional `enabled`. The section's answers no longer hold an `enabled` entry. This needs `slow-change`.
- **Docs.** `optional-sections.md`, the clause syntax in `conditions.md`, `conditions.md`, the schema README, `reserved-names.md`, the domain model, the glossary, the architecture docs, the ACME example and its `review-design-system` template, 0030, and 0010 change.
- **Contracts**, each its own small increment:
  - Plan decoder and encoder: read and write the optional `enabled` on a plan section.
  - Section loader: a template with a variable named `enabled` is an error, and a template section has no `enabled`.
  - Section hasher: the switch is left out of the hash.
  - Section selector: reads the switch. The `enabled` variable, its answer, and its default logic go. The selector fixtures named for `enabled` change.
  - Condition: the model gains the section clause. The plan decoder and encoder read and write it, the section loader reads it, the hasher hashes it as part of a condition, and the condition evaluator tests it.
  - Plan instantiator: every section starts on, so nothing is written.
  - Question finder and plan editor, not yet drafted: the editor offers the toggle, and the finder never lists `enabled`.
- **Versioning.** As in 0035, nothing exists outside this repository, so the schema version stays 1.

## Settled in review

- The model holds `enabled`, true by default, and not a flag for "off". How a language keeps an absent field true is a code detail.
- The plan validator reports a `section` clause that names an unknown section, as it does for an unknown question.
