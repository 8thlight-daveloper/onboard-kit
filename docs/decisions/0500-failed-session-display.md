---
category: ux
status: accepted
approval: agent
note: An agent-selected bounded clarification under this session's delegation, not human approval or accepted-survivor authority.
affects: [session/editor]
---

# 0500. Failed inspection is not an empty session

## Context

[0460](0460-record-backed-session.md) requires actual source review before a
registered record can be edited. A fatal reviewer result has no usable Session.
Two independent drivers disagreed about its display: one showed only safe
diagnostics; the other rendered the zero return as an empty session and offered
an action menu before refusing without reading an action.

Both refused without saving, but zero activity counts and unavailable support
would describe facts that a failed inspection never established. An action menu
that cannot accept an action adds no useful repair guidance.

## Decision

After a fatal initial reviewer inspection, display the complete safe Session
problems block and fixed repair instruction, then refuse. Do not display a
current-session header, derived session fields, reference comparison or action
menu. Consume no action input, sample no event and change no files.

All actual method diagnostics retain the existing safe projection. If finalizing
an approved change refuses, display its actual diagnostics without reopening a
session or reporting success. Existing event sampling and no-write rules remain.

This clarifies the existing conditional terminal grammar and fatal-result rule.
It changes no validation, ownership, exit code, diagnostic privacy, save consent
or retry behavior. A valid session with only the separate allowed stale-support
problems still displays its real editable activities and unavailable support.

The [behavior principles](0100-behavior-judgment.md) favor useful honest repair
guidance and preserved ownership over a fabricated empty state. A failed result
is not a successful result containing no work.

## Acceptance boundary

Independent fixtures cover real codec, preparation identity, held ownership,
effective-answer and mixed fatal/stale failures. They require safe diagnostics,
identity-only input consumption, no clock and unchanged native objects. A
stale-support-only positive control prevents hiding every session with a support
problem. The driver contract and terminal grammar carry the exact rule.

This does not establish crash recovery, hostile-owner safety or equivalence for
unreached states. The common differential run and exact release verification
must be repeated after a source repair.
