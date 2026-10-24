# Independent history example and integration obligation

The executable core fixtures hold repository-old with local definition edits,
the exact legacy-original-id-baseline Source hash, shared and newcomer scoped
facts, a child relationship and a dependent prerequisite. The source-less
landing/dependent definitions remain source-less. Retired keeps its present
empty Source and disabled switch; no baseline is repaired.

Separately held record example, NOT an input to catalog editor:

```yaml
schemaVersion: 1
kind: record
account: example-account
firstOpenedAt: '2026-01-02T03:04:05Z'
outcomes:
  - section: repository-old
    outcome: verified
    on: '2026-01-03T04:05:06Z'
    sectionHash: held-definition-reference
    note: Original independent resource verified.
  - section: retired
    outcome: deferred
    on: '2026-01-04T05:06:07Z'
    sectionHash: legacy-unknown-basis
    note: Preserve this history.
```

Explicit operations: Add repository-access as repository-new with its own shared
repository_scope fact; Retire repository-old; Add document-reading as
document-new with an explicit Deferred reading_material. All original sections
remain in their held order, repository-old becomes Enabled false, and the two
new sections append in that operation order. Original source hashes, all old
facts and the dependent Requires target remain unchanged. Each new Source.Hash
is standard SHA256 of its independently authored canonical literal, excluding
the literal fixture's final LF. It is NOT an old baseline or outcome hash.

Reversing the held section order changes only that retained prefix. Reactivate
repository-old later addresses the same held identity and keeps its facts and
baseline. A new independent resource may never Add under repository-old,
retired or a hidden answer scope key. Omission never means retire or delete.

Future shell acceptance MUST start with literal original plan, preparation and
record files, review an explicit edit batch and save only the reviewed plan.
Compare record/preparation exact bytes after reorder/add/retire/reactivation;
new activities must have no transferred outcomes. Cancel, partial consent,
stale context and publication failure preserve all original files. No shell
grammar, save implementation or passing whole-binary scenario is supplied by
this pure core proposal; merely holding a record variable beside a core call
would not independently prove any I/O preservation property.
