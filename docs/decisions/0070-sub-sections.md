---
category: behavior
status: accepted
approval: unrecorded
affects: [document/assembler, plan/authoringvalidator, plan/decoder, plan/editor, plan/encoder, plan/instantiator, plan/structurevalidator, question/finder, section/hasher, section/loader, section/resolver, section/selector]
builds-on: ['0050', '0065']
---

# 0070. Sub-sections

Builds on [0050](0050-section-switch.md) and [0065](0065-messages-and-evidence.md). It relaxes the rule in [conditions](../reference/section-template-schema/conditions.md) that a section's `includeIf` names only plan variables.

## Context

0065 says a section is the unit of "only if": every step in a section applies together, and a step that applies sometimes goes in its own small section. That raises a question. What decides whether the small section applies?

- A plan variable can decide it, but then a fact that belongs to one topic, such as "does a newcomer request their own source control access?", lives in the plan beside the axis variables, away from the section it concerns.
- The section's own variables cannot decide it, because a section cannot gate itself.
- `includeIf` may not name another section's questions today. The reason is circularity: whether a section is in must not depend on a question that is asked only if some section is in.

A parent and child relationship removes the circularity and keeps the fact with its topic.

## Decision

A section may name a **parent** section with `parent: <section id>`. A section with a parent is a **sub-section**.

- **A sub-section is active only if its parent is active.** The parent is active by its own `includeIf` and switch, and its own parent in turn if it has one. A sub-section is then active by its own `includeIf` and switch.
- **A sub-section's `includeIf` may name questions of its ancestors.** It names them by their ids, such as `access.self_service`. It may still name plan variables. It may not name a question of a sibling, a descendant, or an unrelated section. The relationship is a tree, so a reference can only point up it, and no cycle is possible.
- **Switches cascade.** Turning a section off turns off its sub-sections. A sub-section can be turned off alone. The cascade is computed when the plan is read and never stored: each sub-section keeps the switch a person set, so turning the parent back on restores a sub-section that was turned off alone.
- **Order.** A sub-section comes directly after its parent and its earlier siblings, in the order the selector gives, and takes its parent's timing when it has none of its own. Its questions are asked after its parent's.
- **The document nests it.** A sub-section is written one heading level below its parent.
- **Its answers stay its own.** They live under the sub-section's id, as for any section.
- **The parent is part of the section's content.** It is in the content hash, written like any other field.

An example: an `access` section asks the onboarder "Do new people request their own source control access?" A sub-section `access-request` has `parent: access` and an `includeIf` on that answer. Its steps, and the message it sends under 0065, appear only for organizations that answered yes.

## Why

- The fact stays with its topic, and the plan variables stay for facts about the whole engagement.
- A tree makes the condition rule safe. A question can depend on its ancestors and nothing else, so there is nothing to cycle.
- Switches and ordering already exist and carry over.

## Consequences

- **The domain model changes.** A Section gains an optional `parent`. This needs `slow-change`.
- **Contracts**, each its own small increment:
  - Decoder, encoder, loader: read and write `parent`.
  - Hasher: the parent is part of the hash.
  - Selector: activity follows the parent, and order and timing follow the rules above.
  - Finder: question order follows the section order, and the evaluator is given the ancestors' answers.
  - Resolver: nothing, since it fills one section at a time.
  - Assembler: heading level from depth.
  - Plan validator, in its structure checks: the parent exists, there is no cycle, and an `includeIf` reference points only to an ancestor or a plan variable.
- **Docs.** The schema docs and the domain model, and the conditions rule changes from "plan variables only" to "plan variables and ancestors".
- **The walkthrough** groups a sub-section under its parent. The overview shows it indented one level for each level of depth, under its own section title, and the progress rail keeps its cells next to its parent's. The wireframes show it.
- **ADR 0030.** Sections stay a flat list in the file, each with its own `parent` field, so the file keeps its shape and diffs stay small.

## Not decided

- A depth limit. The model allows any depth, and the first templates need one level.
- Whether a sub-section can be answered by a different audience than its parent. Nothing in this ADR forbids it.
- How the assembled document shows a sub-section that has no steps of its own.
