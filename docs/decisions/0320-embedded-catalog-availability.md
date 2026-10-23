# 0320. Embedded catalog availability

Status: accepted by agent judgment; read-only availability implemented

Implements a bounded availability increment of
[0150](0150-onboarding-library-and-preparation-flow.md), using the complete
[onboarding catalog](../reference/onboarding-catalog.md).

## Context

The full catalog must travel inside the self-contained executable. Embedding
only its list of titles would not supply usable templates. Representing all
entries as newcomer activities would also misrepresent onboarder preparation
and standard tool help. Availability must not silently select activities,
invent assignments or change existing plans.

## Decision

Embed exactly the catalog's existing eighty identities and complete generic
content. Keep identities stable and classify every entry explicitly:

- Fifty-nine section-template entries carry actual schema-version-1 Section YAML
  accepted by the existing loader and authoring validation.
- Nineteen preparation-guidance entries carry Markdown instructions for the
  onboarder. They are not preparation instances or newcomer outcomes.
- invitation-start and onboarding-resume carry non-instantiable tool-help
  Markdown. They explain supported capability and clearly label staged work.

onboarding-repository and onboarding-refresh remain actual section templates
for organization-specific sharing and update instructions. Their content must
not invent a supported update command. Standard rendered invitation wording
remains owned by the first-steps assembler; static catalog help is no second
private renderer and uses no fabricated personal context.

A separate immutable manifest supplies exact identity, title, category, planning
goal, intended result and carrier. Do not insert that metadata into unknown
template fields. Keep ordinary templates in the existing loader format, with
no Source, Enabled, accepted answers or default plan. Organization-specific
routes, commands and permissions require explicit questions, not guessed values
or provider defaults. No real personal or organization-specific data belongs
in built-in content.

Preparation guidance explains existing ownership, relevant fields, missing
information and readiness. It never emits a filled recipient or arrangement,
assumes a section instance exists, sets arranged status or verifies access.
Names such as invitation-prepare and support-assign do not create new arrangement
kinds. Targets remain informational goals, never overdue failures.

### Read-only interface

Provide a pure catalog reader with exact identity lookup and identity-sorted
summaries. Return owned containers and content bytes. Read only compiled-in
data: no caller files, environment, clock, network, optional directory or saved
state. Do not parse identities as paths, trim or case-fold them, infer aliases,
or choose a winner for duplicate manifest identities.

Expose these explicit command operations:

```text
onboard-kit catalog list
onboard-kit catalog show --id ID
onboard-kit catalog export --id ID
```

List emits machine-readable inventory with carrier classification. Show presents
the complete entry and its content. Export writes exact YAML for a section
template or exact Markdown for guidance/help, without a status banner mixed into
stdout. Help explains these differences. Unknown identities produce an empty
stdout data error without echoing an arbitrary selector. Malformed operations,
unknown flags, missing or repeated identities and positional ambiguity are usage
errors. Independent contracts pin syntax, report codes and output boundaries.

These operations write stdout/stderr only. No output directory, overwrite option,
automatic selection, plan mutation, preparation generation, account inference,
download or custom-directory fallback is added. Existing explicit custom-template
loading and historical example outputs remain unchanged.

## Rationale

P1 and P3 of [0100](0100-behavior-judgment.md) favor usable definitions over empty
inventory. P5 requires offline availability. P2, P4 and
[0140](0140-person-specific-preparation.md) preserve ownership and saved work.
P6 keeps planning goals neutral; P7 excludes real personal data. The carrier split
preserves these constraints while making every catalog definition available.

## Verification and remaining work

Independent contracts pin all eighty identities, their classifications and full
content, not merely nonempty exports or the implementation's discovered subset.
Normative content is a lasting contract input, separate from disposable reader
code and independently authored expected inventory/content digests. Test exact
lookup, ordering, ownership, repeated and concurrent calls, malformed arguments,
read-only behavior and real section loading. Verify all content against the
exact artifact outside the checkout without external template files.

Full availability does not complete 0150's runtime flow. Selection, repeated
resource identities, activity-to-preparation relationships, explicit answer
bindings, guided invitation authoring, confirmed import and record-backed
walking require their own contracts. Catalog membership creates no Requires
edge or authorization. Never bypass team policy to complete PR or deploy work.
