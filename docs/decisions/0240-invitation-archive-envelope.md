---
category: architecture
status: accepted
approval: agent
builds-on: ['0090', '0140', '0210', '0230']
---

# 0240. Invitation archive envelope

Builds on [0090](0090-plan-and-onboarding-record.md),
[0140](0140-person-specific-preparation.md),
[0210](0210-person-specific-preparation-paths.md), and
[0230](0230-first-use-record-compatibility.md).

## Context

The invitation must carry enough local material to begin before company access
exists. Reading an archive must not authorize its paths, overwrite existing work,
or claim that its contents are suitable onboarding data. Separate the bounded
delivery envelope from artifact decoding and the confirmed import operation.
This follows P2, P4, P5, and P7 of [0100](0100-behavior-judgment.md).

## Decision

Use separate pure invitation encoder and decoder components over InvitationFiles.
This transport value holds an explicit Account string and four byte sequences:
Plan, Preparation, Record, and FirstSteps. It is not a joined plan, a fourth
source of truth, an authorization, or a new saved YAML schema. Payload bytes are
opaque at this boundary, including empty or non-text bytes. The importing driver
must separately decode, validate, and explain them before any write.

The archive has exactly four regular-file entries:

1. plan.obk.yaml
2. records/<account>/preparation.obk.yaml
3. records/<account>/record.obk.yaml
4. first-steps.md

These are archive names, independent of the plan's configured resultsPath.
Neither component extracts files, opens paths, reads an environment or clock,
or alters its input. The decoder returns owned payload bytes. A driver derives
actual destinations through the preparation locator and enforces filesystem
safety separately. Never copy archive entry names into extraction operations.

Account follows the complete ASCII pattern `[A-Za-z0-9][A-Za-z0-9._-]*`, with exact
case and no trimming, inference from payloads, or normalization. Both account
paths must name the same account. Each full entry name must fit the 4096-byte
application path budget of 0210; check lengths before constructing names.
Reject duplicate names, extra entries, missing entries, directory entries,
symlinks, unsupported file types, encryption, and unsupported compression.
Reject alternate spellings rather than cleaning traversal or choosing a winner.
Case variants of a fixed name are not valid aliases.

Keep 0090's aggregate limit of 1 MiB uncompressed across all four payloads.
Also cap the supplied ZIP bytes at 2 MiB before parsing. This defensive input
budget accommodates the canonical writer's bounded headers and payloads while
preventing large containers full of irrelevant metadata. It is a new envelope
limit, not an operating-system limit or permission to embed large help files.
Check declared sizes without overflow, then bound actual decompression and
require successful checksum verification. Do not allocate from an unchecked
archive size or trust a header as proof that a stream is small.

The writer uses the listed entry order, Store compression, fixed private regular
file permissions, and a fixed 1980-01-01 ZIP timestamp, never the current time.
Its output is deterministic. The reader also accepts Deflate, does not require
writer order, and does not require byte-for-byte canonical writer output.
Contract details specify the harmless metadata it accepts and the exact
diagnostic precedence. Archive names are routing data only, never executable
instructions. Neither ZIP checksums nor a plan hash establish authenticity.

## Import boundaries

Envelope success does not establish that payloads decode, share a plan or account,
contain an identity-only seed, or provide usable bootstrap instructions. Those
checks belong to the invitation driver contract before identity confirmation
and publication. A new seed contains no legacy startDate, first-use event,
newcomer answers, or outcomes. The confirmed start driver records first use as
0230 specifies; the archive components never do.

Existing-record reconciliation remains required by 0090. Import conflict rules
for different plan or preparation copies and recoverable multi-file writes need
a separate driver contract before implementation. An archive timestamp is not
evidence that its preparation is newer. This decision does not approve silent
replacement, automatic Git operations, network access, or an extraction command.

## Verification

Independent contracts must cover both size boundaries, malformed and truncated
containers, declared versus actual lengths, checksums, each forbidden entry
class, identity namespace mismatches, and valid Store and Deflate controls.
Writer assertions use independent expected metadata and payloads, not merely
agreement with the new reader. Tests must also pin determinism, input and prior
result preservation, privacy-safe diagnostics, and no partial result on failure.

These are planned safeguards, not evidence that an invitation command works.
Binary scenarios must later cover pre-access use, confirmed identity, refusal
without writes, and preservation of existing work through import and retries.
