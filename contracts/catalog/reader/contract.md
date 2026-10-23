# Catalog reader contract

Component catalog/reader. New takes no arguments and uses only compiled data.

Authority: ADR0005, ADR0015, ADR0020, ADR0025, ADR0100, ADR0110, ADR0120,
ADR0140, ADR0150 and ADR0320; the complete onboarding catalog and existing
section template schema. API types belong to this contract, not the saved model.

## Interface and permanent inputs

List() returns summaries. Read(exact ID) returns an Entry and problems. Construct
the actual built-in catalog internally; no directory, filesystem, manifest bytes,
clock, environment, account, plan or substitute dependency argument exists.

The permanent normative inputs are content/manifest.json and all its declared
content files in this contract. They are lasting authored content, not judging
goldens and not disposable implementation. Builders may read and embed them.
Runtime operations read compiled data only. Do not locate these paths at runtime.
No loader, validator, encoder or other runtime project dependency is needed.

The manifest is one JSON array with exactly eighty objects. Each object has
exactly the seven string keys id, title, category, goal, result, carrier and
contentPath. Every field is nonempty. Text is the exact corresponding catalog
table metadata; goal is an informational phrase, not a Timing enum or deadline.
Entries are sorted by ascending exact ID UTF-8 bytes, with unique IDs. No alias,
title-derived identity, optional external subset, first/last winner or merging.

The independent testdata/inventory.json fixes all eighty exact summaries. Do
not derive the expected inventory by discovering content or implementation files.
There are 59 section-template, 19 preparation-guidance and 2 tool-help entries.
Only invitation-start and onboarding-resume are tool-help. onboarding-repository
and onboarding-refresh retain section-template carriers. All nineteen onboarder
preparation identities remain preparation-guidance, not newcomer activities.

Categories are invitation-and-first-steps,
identity-equipment-and-communication, team-and-coordination,
tools-repositories-and-documentation, software-work-and-delivery, design-work,
feedback-and-progress and onboarder-preparation, as pinned per entry.

contentPath is exactly sections/ID.yaml, preparation/ID.md or help/ID.md according
to carrier, relative to the permanent content root. It is not a caller path or
permission to clean an input ID into one. Files are complete UTF-8/LF content
with exactly one final LF. Export them byte-for-byte, without reformatting,
resolving variables, translating line endings or appending a second newline.
Pinned independent size/SHA256 fixtures fix content after authoring freezes.

## Results and errors

List returns all eighty exact summaries in ascending ID byte order. It returns
no content and selects no activities. Summary fields are precisely the manifest
values; no counters, generated timestamps, inferred team or display aliases.

Read of a held exact ID returns that complete Summary and complete actual
content bytes, with no problems. The section carrier is actual schema-version-1
Section YAML, not catalog Markdown, a JSON wrapper or a runtime-expanded plan.
Preparation/help carriers are their complete Markdown, never fabricated saved
Preparation or Record data. Real loader/authoring composition checks belong to
separate integration evidence; this reader does not reinterpret the content.

Every other ID, including empty text, wrong case, whitespace padding, a display
title, separators, dot/parent paths, a NUL or invalid Go UTF-8, returns exactly
the zero Entry (zero Summary and no content) and one problem:

| Field | Value |
| --- | --- |
| Code | catalog.unknown-id |
| Severity | error |
| Message | Choose an exact ID from catalog list. |
| Question | empty |
| Section | empty |
| Detail | exactly field: id |

No supplied ID appears anywhere in the diagnostic. Refusal has no partial
metadata/content, nearest match or fallback. Successful problem lists and
zero-entry content may be nil or empty. The unknown Detail has its required
one key; nil/empty are not equivalent to that required map. No other diagnostic
or runtime manifest-repair policy is introduced for immutable validated inputs.

## Ownership and side effects

No caller or outside state is read or written. No filesystem, environment,
process, network, clock, saved plan/preparation/record or model invocation.
List slices, returned Entry slots, Content backing bytes, problem slices and
Detail maps are new owned storage, independent of earlier/later calls and the
compiled source. Strings are immutable values. A caller may edit results freely.
List/Read on one instance are deterministic and may run concurrently. Unknown
calls do not poison later known results. No retained mutable selection state.

Catalog membership creates no Requires edge, selected plan, accepted default,
arrangement, verification fact, due-date failure or permission. The source text
must not be silently changed to make it look complete or executable.

## Four-part oracle and verification boundary

api.go defines the binding. wire_test.go starts with a nil stub and later changes
only its constructor call. reader_test.go judges inventory, exact lookup,
unknown diagnostics, byte digests, ownership and concurrent reuse. Inventory
and digest fixture files state their provenance. Digest values are independently
computed from frozen normative content, never by a production Reader.

A stub must fail nonempty List/known Read and every unknown-problem case by
assertion; empty/property controls may pass. No fixture-loading/compiler error
counts as successful stub-red evidence. No build/mutation/differential evidence
is implied by this authored contract.

| Plausible wrong implementation | Intended independent discriminator |
| --- | --- |
| Return titles only, examples only or a discovered subset | Exact eighty complete summaries and carrier counts |
| Sort by title/category or return manifest declaration accidents | Exact ascending ID inventory |
| Classify all content as sections or help | Exact 59/19/2 classification and mixed-purpose entries |
| Return an empty/truncated/reformatted content body | Every ID's pinned size and independent SHA256 |
| Clean/trim/case-fold IDs or echo unknown data | Exact known and unknown lookup table |
| Return metadata or bytes beside an error | Zero Entry checks on every unknown input |
| Share buffers, problem maps or mutable global state | Bidirectional result edits, reuse and concurrent known/unknown calls |
| Load source content files at runtime | Outside-checkout exact-built-artifact acceptance, separately required |
