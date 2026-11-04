---
category: behavior
status: accepted
approval: agent
affects: [catalog/materializer, document/assembler, outcome/recorder, record/reconciler, section/walker, session/reviewer]
---

# 0270. Outcome definition hashes

Reconciles [0030](0030-self-contained-plan-file.md) with
[0085](0085-walking-sections-and-recording-outcomes.md). Supersedes 0085's use
of the held Source.Hash for outcome recording and its claim that every mismatch
proves a change after verification. Preserves the verified-kind prerequisite
rule of [0260](0260-prerequisites-and-action-order.md).

## Context

Source.Hash records the original template baseline. Editing a saved section
does not update it: template updates need that baseline to distinguish local
changes from upstream changes. Custom sections may have no Source at all.

The record reconciler currently compares outcomes to Source.Hash and skips
source-less sections. Consequently a local instruction edit can escape its
changed-outcome flag. Treating the baseline as today's definition is not a
sound way to compare the newcomer's record with the current instructions.

Legacy outcome hashes have no basis or algorithm marker. Some writers may have
recorded the template baseline after instructions had already changed. A new
comparison cannot reconstruct that history from dates or Git metadata.

## Decision

### Hash the held definition

For each new explicit verified, failed, or deferred outcome, compute the actual
current Section hash with the section hasher. Store it in Outcome.SectionHash.
Use the validated section identified by its exact ID, before resolution or
rendering. The driver supplies the timestamp; the core reads no clock.

When comparing a held nonempty outcome hash with a section still present,
compute the same actual Section hash. Source may be absent, stale, or unrelated;
it supplies no fallback. Compare exact held text, without case folding, repair,
truncation, or an Enabled-state exemption. Report mismatching section IDs once,
in sorted order, through the existing ChangedOutcomes result.

An empty held hash means comparison unknown. A removed section has no current
definition to compare. Keep both kinds of history, but do not add them to the
mismatch list. Absence from that list is not proof of freshness. An arbitrary
nonempty legacy hash is compared as held, not silently corrected.

### Preserve both histories

Never rewrite Source.Hash to obtain agreement. Never migrate a held outcome to
today's hash during inspection, reconciliation, refresh, or invitation assembly.
Outcome kind, timestamp, hash, note, differently text, and first use remain held
facts. Only an explicit authorized outcome operation replaces an outcome.

Show a mismatch as "current definition differs from recorded hash", not as proof
that editing happened after verification. Explain that older records may refer
to an original template baseline. A canonicalization change can also produce a
mismatch without an authored edit. Do not invent a hash-version field or infer
one from event dates. Stronger historical claims need a separate format decision.

The hash covers the section definition under the existing hasher contract, not
all rendered instructions or onboarding context. Answers, preparation, directory
lookups, Source, and the section switch do not participate. Equality is neither
authenticity nor proof of current access authorization.

### Verification is a separate fact

A unique stored verified outcome still satisfies the verified-kind prerequisite
gate when its hash mismatches or comparison is unknown. Display the distinction;
do not hide a freshness requirement inside hashing, reset a dependent, or erase
notes. Failed, deferred, and absent outcomes remain insufficient.

Reject duplicate section identities and duplicate outcome identities at the
runtime validation boundary. This increment does not change the reconciler's
defensive treatment of unchecked duplicate input into a supported runtime rule.

## Rationale and alternatives

Principles P2 and P4 of [0100](0100-behavior-judgment.md) protect ownership and
held work. P3 favors a useful, truthful definition comparison. P5 includes
custom self-contained sections, and P8 favors an honest sufficient comparison
without another question. P7 argues against expanding the hash to personal
answers or directory information without a real need.

Keeping the current source comparison temporarily would be honest only if the
product did not promise local-definition detection. Rewriting Source would
destroy the template-update baseline; rewriting outcomes would manufacture
agreement. Hashing resolved Markdown tracks a different question. Passing a
caller-supplied hash map would move the original mistake into composition and
change the public API. Prefer the actual pure hasher dependency.

## Component boundary

Move the record reconciler from dependency position 1 to 2 and permit its direct
dependency on section hasher at position 1, alongside model and its own contract.
Keep its public API and constructor. It reads actual current section content,
but writes no plan field. Its existing record-answer reconciliation and report
outputs remain unchanged except for the explicit comparison basis.

The section walker already has a hasher dependency under 0260. The outcome
recorder may also depend directly on the hasher at position 1, in addition to
walker and model, to pin the actual section at the write boundary. Architecture
tables and independent contracts must reflect these reads before building.

## Verification and delivery

Independent fixtures must pin literal canonical strings and expected digests,
not compute expected values by calling the implementation under test. Cover a
locally edited section with an unchanged baseline, a custom source-less section,
changes only to Source or Enabled, matching and empty hashes, removed sections,
and complete preservation of outcomes and first use. Apply wrong source-baseline
and source-presence variants to prove the tests reach these gaps.

Runtime scenarios must display mismatches truthfully while allowing the held
verified-kind gate. Refused and read-only operations must leave files unchanged.
Old bare hashes stay bare; no saved-format migration is introduced.

The reconciler comparison contract and implementation now use actual definition
hashes. Recording and runtime scenarios remain planned; this decision alone is
not runtime evidence. The new walkthrough must verify its integration before
claiming current definition reporting. Requires format work proceeds independently.
