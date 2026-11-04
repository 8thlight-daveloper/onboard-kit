---
category: behavior
status: accepted
approval: agent
note: An agent-selected integration candidate, not human approval or accepted-list authority. It is separate from the authoring work in 0440 and 0450.
affects: [record/writer, session/editor, session/reviewer]
---

# 0460. Explicit record-backed session editing

## Context

0090 separates organization-owned shared plan/preparation from newcomer-owned
record answers/outcomes. 0290 makes verified an explicit statement, never an
executed check. 0360 requires pending-import detection before a mixed session
can be used. An unconstrained --plan/--record shell cannot prove the registered
context/pending boundary. Generic record splitter omits equal preseed overrides;
that remains its existing contract, not the right abstraction for explicit
newcomer ownership. Defaults/examples remain suggestions, not stored answers.

## Decision

Preserve existing cores. Position6 record writer is a narrowly contracted pure
explicit-intent writer; position7 session reviewer composes actual codecs,
joiner, walker4, authoring/answer validators, writer6 and outcome recorder5.
Inspect/Propose/Finalize have no file or clock effects. Position8 native session
editor alone owns registered context, consent, actual decoder2+locator binding,
native guards/locks and one-record publication; position9 entrypoint owns raw
command dispatch and a real UTC clock. 0020 permits growing the entry position;
do not omit a dependency to retain an obsolete maximum8. No shared-model/API
changes beyond those separately frozen pure contracts are proposed here.

The finite route is `session edit --root ABS --home ABS --session KEY`. The
operator reviews explicit selected identity, confirms it, sees the real joined
session, proposes ONE answer/follow-plan/outcome, reviews its complete effect,
and separately saves. No hidden default, read-ahead, account guessing, automatic
reconciliation, scope/audience bypass, command/check execution or second driver.
Initial registered paths are metadata-only/unverified; actual decode/locator
occurs only after identity consent and bounded raw admission. Registry metadata
and cooperative root/home exclusion establish pending/selection before sources.

Exact matching plan and account identities retain the existing old-root
continuity requirement. Genuine absence in an unrelated registration need not
block a safe selected edit, but requires native no-follow prefix, absence and
nonalias proofs through publication. A symlink or inspection failure is never
absence. Unrelated files and registration stay untouched and undisclosed.

Before/after ownership example:

| Record state after Plan A choice=org-a | Plan B choice=org-b effective answer |
| -- | -- |
| untouched follower, no key | org-b |
| explicit submit org-a, stored personal.choice=org-a | org-a |
| prior explicit override then deliberate follow-plan, key removed | org-b |

A true variable default of org-a without an authorized stored plan answer is
still unanswered until explicit submission. Follow-plan does NOT promise blank
when a shared answer exists. No tombstone, invented previous record or generic
Split behavior change is needed. Pure writer preserves exact unrelated model
shapes; changed encoding permits ONLY documented known-empty container normal
forms. Unknown {} and [] remain actual data. Raw no-op retains original bytes.

Outcome preview labels its new time sampled-after-save. Only actual Finalize
receives one injected current event after consent and stale guards; answer and
follow-plan never sample it. Current prerequisite gate and definition hash come
from actual reviewer/walker/outcome recorder, not shell duplication. Historical
firstOpenedAt, reference and unrelated outcomes never become current implicitly.

Record-only native publication uses a private sibling stage and same-volume
atomic rename, not starter's multi-artifact transaction. Exact selected native
metadata, including ALL persistent xattrs/ACL/flags, is preserved or refused;
Darwin automatic provenance is no excuse to ignore an attribute. Preserve
unknown abandoned stages and refuse, never replay/delete them. No power-loss,
hostile-owner CAS or unreached crash-window guarantee is added. Independent
native mandatory positive fixtures must actually save before claiming delivery.

## Consequences and acceptance boundary

Read-only start and legacy walk are unchanged. The route requires a registered
session, explicit root/home/key and existing safe artifacts; it offers no repair
or automatic rebind. Plan/preparation/metadata remain organization-owned exact
artifacts. Three raw selected bodies, identity/native metadata and pending state
are rechecked; only selected record changes. Record privacy includes scrollback
and private interrupted staging. Operators get truthful saved/repair status.

Pure writer/reviewer evidence is separate from native and binary verification.
Contract native tests use actual files and a same-owned-PID native-lock process,
not fake storage. Binary proposal requires a separately reviewed generic capture
runner, and must really invoke the built binary. Source-only fixtures cannot
prove native metadata operations, crash timing, descriptor cleanup under every
failure, or binary dispatch. See provenance.md for finite verification plan.
