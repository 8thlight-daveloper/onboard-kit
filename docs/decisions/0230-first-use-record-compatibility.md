---
category: behavior
status: accepted
approval: agent
affects: [outcome/recorder, record/decoder, record/encoder, record/reconciler, record/splitter, session/reviewer]
builds-on: ['0180', '0100']
---

# 0230. First-use record compatibility

Implements the model direction of [0180](0180-first-onboarding-use.md) and applies
P2, P4, P7, and P8 of [0100](0100-behavior-judgment.md).

## Context

Recording when a newcomer first begins onboarding removes an unnecessary date
question. Existing records may already represent months of work, however. An
empty timestamp does not prove that a record has never been opened, and a legacy
startDate is not evidence of that event. Upgrading must not manufacture history.

## Decision

Add optional FirstOpenedAt of type Timestamp to Record, serialized as
firstOpenedAt after startDate and before plan. An absent or empty value remains
unknown and is omitted on encode. Keep schema version 1: this is an additive
optional root field, where older readers preserve unknown data.

Record codecs remain representational. Decode the field with the existing text
rules and preserve its exact text; encode it with the existing string rules.
The codecs neither validate a calendar or timestamp nor read a clock. Recognize
firstOpenedAt as a known root key, reject repeated raw keys, and reject an
unknown-fields map entry colliding with that known key rather than choosing a
winner. Do not also retain the recognized field as unknown data.

As required by the existing YAML dialect, text containing Unicode line or
paragraph separators must survive serialization without becoming invalid YAML.
Extend [0200](0200-preserve-unicode-line-separators.md)'s explicit escaping rule
to record strings and unknown keys in this contract increment. Plan codecs are
not changed by this decision.

Answer splitting and record reconciliation preserve FirstOpenedAt exactly,
including empty or unusual held text. It is not an answer, outcome, plan date,
section hash, or invitation-creation time. None of those operations records a
new event or silently repairs a stored timestamp.

## Recording the event

The invitation/start driver supplies a UTC RFC 3339 timestamp only after the
person confirms the account and destination and begins onboarding. It sets the
field only when creating their new local record at that confirmed boundary.
An invitation seed has no first-use event. Preparing, inspecting, rejecting,
or merely decoding an invitation records nothing.

For any existing record without FirstOpenedAt, leave it unknown, even if answers
and outcomes are empty. Absence of work is not proof of absence of prior use.
Do not backfill from the current clock, startDate, preparation, file metadata,
or Git. The tradeoff is deliberately missing early metrics rather than falsely
precise history. No extra question is added to resolve that uncertainty.

An existing nonempty event is preserved on retries, imports, reconciliation,
and preparation updates. A driver that needs to interpret it must first check
that it is a real UTC RFC 3339 timestamp; malformed held text is an actionable
error, not permission to replace it. The codecs still preserve such text so a
repair does not silently destroy data.

Legacy startDate remains exactly as held and is not relabeled actual start or
first use. New invitation seeds do not copy plannedStartDate into it. Planned
date remains onboarder-owned preparation; the newcomer is not asked for an
actual-start-date answer. Later target views must name their chosen anchor and
show unknowns rather than silently substitute dates.

## Verification boundaries

Independent codec fixtures pin omission, exact text, key order, known-key
collision, repeated raw keys, text type guards, and Unicode separators. Splitter
and reconciler fixtures pin preservation alongside real answer updates and
refusals. These tests do not prove an event was recorded at the right time.

Driver contracts must separately inject a clock and test confirmation before
writes, exactly-once new-record behavior, existing unknown history, retries,
identity mismatch, and zero event writes from non-onboarding commands. Binary
scenarios must exercise the actual invitation and start commands. Model and
serialization changes alone are not a completed first-use feature.
