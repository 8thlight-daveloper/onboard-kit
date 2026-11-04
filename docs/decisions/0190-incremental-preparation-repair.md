---
category: behavior
status: accepted
approval: agent
affects: [command/prepare, preparation/editor]
builds-on: ['0100', '0110', '0140']
---

# 0190. Incremental preparation repair

Builds on [0100](0100-behavior-judgment.md),
[0110](0110-reject-ambiguous-plans.md), and
[0140](0140-person-specific-preparation.md).

## Context

A preparation file may become invalid when a helper leaves or a section is
removed from the organization plan. Requiring the whole file to validate after
every edit would prevent correcting one stale arrangement while another still
needs repair. Silently discarding stale data would lose the onboarder's work.

## Decision

The preparation editor validates the requested change, not readiness of the
whole file. It first checks schema, kind, account, plan identity, and duplicate
arrangement identities. It then validates the supplied recipient or arrangement
and preserves unrelated invalid details unchanged. A successful edit is not a
claim that the preparation is ready to send. Delivery must separately validate
the complete preparation and invitation requirements.

Recipient edits replace the recipient value. Arrangement edits replace an exact
permanent id in place or append a new id. Empty supplied fields deliberately
clear previous values; they are not omitted updates. No edit changes the shared
plan or newcomer record, accepts defaults, or records newcomer completion.

An id-only operation cannot distinguish repeated arrangement ids, including
repeated empty or malformed ids. Reject these before every edit rather than
choosing a winner. Removing a unique malformed or empty id is allowed by explicit
exact match so invalid hand-edited data can be removed. A command must distinguish
that explicit choice from a missing argument. Missing targets are errors, not
inferred or silently successful removals.

P3 favors repair that a person can complete one change at a time. P4 favors
preserving unrelated work. Neither overrides P2 ownership or 0110's rejection
of ambiguity. Full validation remains the authority for whole-file validity.

## Verification

The [editor contract](../../contracts/preparation/editor/contract.md) requires
independent expected results for correcting one stale reference while preserving
another, rejecting duplicate identities before edits, preserving earlier values
when edits fork, and returning the original value after an invalid change.
These checks do not establish invitation readiness, file authorization, archive
safety, or preservation across driver writes; those have separate boundaries.
