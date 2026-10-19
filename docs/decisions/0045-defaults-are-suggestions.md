# 0045. Defaults are suggestions

Status: accepted

Builds on [0035](0035-settled-and-deferred-answers.md). It changes parts of [0030](0030-self-contained-plan-file.md) and [0010](0010-plan-editor.md), which are named under Consequences.

## Context

A variable can declare a `default`. The docs call it "the value used when none is given". Three places act on that wording today.

- Filling a variable reads the section's own answer, then the section's default ([0030](0030-self-contained-plan-file.md)).
- A section is active when its `enabled` answer, or its default, is true (0030, [optional sections](../reference/section-template-schema/optional-sections.md)). A separate decision takes this one out of the question model, see Not decided.
- The editor reports each question "with its current answer or default" (0010).

The conditions reference already goes the other way. A condition counts only an explicit answer, never the default.

So a default quietly stands in for a decision nobody made. The plan file does not show it. A later change to a template default changes what an old plan means. And a person cannot tell from the plan whether they chose a value or never looked at the question.

## Decision

A default is a suggestion. It is never an answer, and nothing acts on it unless a person accepts it.

- **The plan holds only answers.** An answer is something a person gave: they typed it, picked it, deferred it, or accepted a suggestion. Accepting a suggestion stores the value as a settled answer.
- **No component fills in a default on its own.** The section resolver, the section selector, the plan validator, and the document assembler read answers only.
- **A user interface offers the default.** It shows the default as the suggested entry. The person accepts it with one action. An action such as "accept every default" is allowed, because a person chose it.
- **A question with a default and no answer is unanswered.** It follows the rules in 0035: a required one blocks, and a variable that text uses becomes a gap.
- **A new plan has no answers.** The plan instantiator copies no default into `answers`. This is what the instantiator contract already says.

## Why

- A person is the only source of a decision. A default is a template author's guess.
- The plan file shows everything that shaped the output. Nothing depends on a template default that can change later, which suits the self-contained plan in 0030.
- Settled and deferred depend on stored answers alone, so the rules in 0035 need no exception for defaults.
- The condition rule and the filling rule now agree.

## Cost

A person answers, or accepts, more questions than before. A walk through the questions that offers each default with one action keeps this short. The review screen can list every question that has a suggestion and no answer.

## Consequences

- **Parts of 0030 and 0010 change.** The lines on filling from the default and on the finder returning a question with "its current answer or default" now say the default is offered as a suggestion. A new ADR supersedes the old lines, so 0030 and 0010 get a pointer to this one.
- **Schema docs.** `variables.md` (the `default` row), `optional-sections.md` (the "No answer" line), and the schema README (the section-active line) change.
- **Contracts**, each its own small issue:
  - Section resolver: rule 3 (a default fills an unanswered variable) is removed. An unanswered variable is a gap, or `<name>` for an onboarding variable.
  - Section selector: no change from this ADR. Switching a section on or off is not a variable, and a separate decision covers it.
  - Plan instantiator: no change.
  - Question finder, plan editor, and plan validator, not yet drafted: each takes the rule from the start.
- **The ACME example.** Its plan holds no answer for `output_style`, which only has a default today, so its assembled document needs an answer or shows a gap. A person decides which.
- **Versioning.** As in 0035, nothing exists outside this repository, so the schema version stays 1. After the first release, a change like this needs a version bump.
- **The hasher** still hashes declared defaults. A default is part of what the template says.

## Not decided

- How a section is switched on or off. It should not be a variable or a question. See [0050](0050-section-switch.md).
- Whether the TUI shows a default as a pre-filled entry or as a hint.
- Whether a plan can record that a person accepted every default at once, or each answer stands alone.
