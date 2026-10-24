# Independent fixture sources

These are agent-authored literal oracles, not human approvals. No component
implementation produced expected bytes, hashes, diagnostic meaning or model
projections. Standard-library ZIP framing is only input transport construction.
Actual public dependency contracts define the composed behavior. Execution and
wrong-variant evidence will be recorded separately by the root agent; this draft
does not claim a passing suite.

| Fixture or constructed input | Source and independent expectation |
| -- | -- |
| minimal.plan.yaml | ADR0380 accepted authoring profile;0055 onboardee ownership;0040 literal variable substitution; plan decoder and authoring validator contracts. One required question is genuinely used in summary. |
| minimal.seed.yaml |0380 identity-only seed;0240 envelope account;0090 record reference. Raw minimal.plan.yaml SHA256 is `00d6473bd698e40caca271f61af7b87288f0c864924008bd24375d6b19b3c6a4`, independently calculated over stored source bytes including LF. |
| minimal.preparation.yaml |0140 separation; actual preparation decoder/validator permit incomplete identity-only preparation. |
| first-steps.md |0380 nonempty UTF8 document data, not a regenerated session or consent. |
| minimal-section.canonical.json | Section hasher contract's fixed canonical key order/defaults; independently handwritten string. Its storage LF is excluded from hash input. SHA256 is `5631a3d34e963b37019b9ae53b12907c3bdef3d17c6642f1462784d178c3f4ad`. |
| begun.record.yaml |0230 supplied first-use event; record encoder key order and inherited plan encoder scalar rules. Timestamp and digit-leading hash are quoted. No actual encoder is an expected-byte oracle. |
| retained.record.yaml |0380 semantic no-op retains original bytes and comments, bypassing canonical encoding; unknown root notes remain held. No existing event is backfilled. |
| choices.plan.yaml |0055 used onboardee questions;0120 declared default is only a suggestion;0370 permits a scalar choice default and command List default, not multiple choice. Actual finder deep list/choice/default data and actual resolver markers. Nested unknown data tests ownership without making it executable. |
| choices.reconciled.record.yaml | Actual reconciler contract removes invalid scalar choice red, but preserves hidden removed scope. Record encoder fixed key order, empty-map omission and plain letter-leading hash. The independent raw choices.plan.yaml SHA256 is `cb1f0be905bb0c1d742c2dea3c314786d5cf7f3199bbadf89b55c742e2e6b787`. |
| stale.preparation.yaml |0380 finite stale diagnostic/code/path/reference profile,0140 support separation; repeated helpers are held, not deduplicated. Nonmatching known references make support unavailable without discarding source or blocking the personal session. |
| support.preparation.yaml |0140 all known recipient/arrangement fields and repeated contacts;0160 no automatic acceptance of preparation as personal answers;0180 planned dates remain goals. Unknown private-note stays in raw source, not Support. |
| Strict seed field mutations |0380 ordered independent guard paths, actual record decoder's accepted false/Deferred entries, empty maps/lists. No synthetic typed None is presented as decoder-reachable evidence. |
| Historical/new event tables |0380 historical UTC preservation differs intentionally from bounded new event input. Gregorian leap-year and exact zone/case/fraction controls are handwritten. |
| Expanded unknown-string pair | Record encoder inherits plan encoder Scalars: raw YAML escape `\0` occupies2bytes and canonical `\x00` occupies4bytes.300000 decoded NUL characters therefore have600000 escaped input bytes and1200000 canonical escaped bytes, without needing a product encoder to establish that the latter exceeds1048576. Fixed surrounding source remains below the input bound. Equal reference must retain; changed reference must refuse oversized encoded output. |
| Cache-tamper and ownership tests |0380 source re-derivation with complete projection equality;0020 immutable pure core;0015 no state. Reflection copies preserve nil/empty and nested shapes rather than using JSON serialization as the ownership oracle. |
