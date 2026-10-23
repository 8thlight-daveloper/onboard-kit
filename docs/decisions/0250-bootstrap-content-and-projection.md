# 0250. Bootstrap content and projection

Status: accepted by agent judgment; implementation pending

[0260](0260-prerequisites-and-action-order.md) retains the marker and full-plan
evaluation rules, but replaces the still-planned nested action fragment with
the global prerequisite sequence and flat activity presentation.

Builds on [0085](0085-walking-sections-and-recording-outcomes.md),
[0090](0090-plan-and-onboarding-record.md), and
[0240](0240-invitation-archive-envelope.md).

## Context

0085 already chooses a section's `bootstrap` marker for activities usable before
company access exists. The model and components do not yet carry it. An invitation
needs the actual section renderer, not copied formatting or a second interpretation
of conditions. Early timing alone does not establish that instructions work
without company access.

## Decision

Add Bootstrap, a boolean, to Section. It belongs to the section definition,
including custom templates, not to answers or a plan-level list of section IDs.
It means the author has designated these instructions for the pre-access path.
It does not certify network reachability, activate a section, verify completion,
or select a suggested answer.

An absent, empty, or null marker is false. Explicit false is false; only the
canonical plain boolean true enables it. Keep the existing strict boolean
spellings, type checks, duplicate-key checks, and anchor, alias, and tag guards.
It is a known key on templates and saved sections, not on the plan root.
The encoder rejects a collision with Section.UnknownFields even when false.

Canonical YAML writes `bootstrap: true` after parent and before timing; it omits
false. The content hash writes the JSON boolean true at the same key position
and omits false. All existing hashes without a true marker remain identical.
Changing the marker changes the definition hash as 0085 requires; it does not
delete outcomes or declare that an earlier verification was false.

### Evaluate first, then project

The document assembler gains AssembleBootstrap(plan), returning a Markdown
fragment and problems. Its existing Assemble behavior stays unchanged.

1. Evaluate activity and ordering with the actual selector on the complete,
   unchanged plan.
2. Retain exactly the ordered active sections with Bootstrap true.
3. Resolve and render them using the original answers, directory, team context,
   and section identities. Do not substitute defaults for missing answers.

Do not delete nonbootstrap sections before selection: conditions may refer to
their switches, and ordering may pass through them. Excluded sections contribute
no content, gaps, or resolution problems. Reuse the assembler's real formatting;
do not slice a completed Markdown document or copy its private renderer elsewhere.

### Containment is not implicit inclusion

A marked parent does not mark a child. A marked child does not include its
unmarked parent's text. Full-plan ancestor gates still determine activity.
For display only, count retained marked ancestors to determine heading depth.
A marked child below an unmarked parent becomes a fragment root. A marked
grandchild with one retained marked ancestor has depth one, even if an unmarked
ancestor lies between them. Never rewrite Parent or the plan to achieve this.

This avoids leaking company-access-dependent parent instructions into the
pre-access document while preserving conditions and scope. Requiring all
ancestors to be marked would force unrelated content into first steps or require
unnecessary author restructuring. Preserving orphan heading levels would make
the result harder to read. These choices follow P1, P2, P3, and P5 of
[0100](0100-behavior-judgment.md).

### Fragment format

The fragment has no document title or generated-review line. Its ordinary gap
table uses a level-three heading. Sections start at level three plus projected
depth, capped at six. Keep the existing section, step, check, footer, placeholder,
gap, diagnostic-order, and vacant-role rules. Nonempty output ends in exactly
one newline; an empty projection returns exactly empty text and no problems.
The first-steps caller explains neutrally when no pre-access activities are listed.
The assembler does not read a clock, save answers, or record first use.

## Staging and remaining commitments

This increment implements the marker and rendering foundation. It does not
implement bootstrap-first walking or certify invitation readiness. 0085's
Requires graph, unknown-reference and cycle checks, bootstrap prerequisite
closure, inactive-prerequisite cascade, and verified-outcome gate remain required
before delivery of the new invitation/start runtime.

Do not silently turn existing After ordering into verification prerequisites.
Keep its current behavior during this foundation increment. A separate decision
must reconcile 0085's intended replacement with existing saved plans before
Requires contracts are drafted. This staging is not a decision to abandon that
replacement or its actual prerequisite semantics.

Standard first-steps help, explicit content sources, confirmed identity/import,
recoverable publication, record-backed saves, the full embedded catalog, stable
repeated instances, and neutral support views remain separate required work.
Do not print unimplemented commands or invent a distribution address.

## Verification

Independent contracts must pin true, false, absent, null, wrong types, forbidden
spellings, raw duplicate keys, key placement, unknown-field collisions, and
hand-computed old and marked content hashes. Preserve the marker through template
loading, instantiation, unrelated edits, refresh, record joining, and selection.

Projection cases must distinguish early timing from bootstrap, evaluate conditions
on omitted switches, preserve order through omitted sections, retain full ancestor
gates, promote display depth without leaking ancestor text, and omit excluded gaps
and diagnostics. Hand-written fragments are the oracle, not text produced by
Assemble and then filtered. Existing full-document goldens must stay unchanged.
These are planned safeguards, not evidence that the runtime flow works.
