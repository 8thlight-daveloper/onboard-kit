---
category: behavior
status: accepted
approval: agent
affects: [preparation/encoder, record/decoder, record/encoder]
builds-on: ['0030', '0100']
---

# 0200. Preserve Unicode line separators

Builds on [0030](0030-self-contained-plan-file.md) and
[0100](0100-behavior-judgment.md).

## Context

Two independently built preparation encoders agreed on writing Unicode line
and paragraph separators literally. A separate probe with a supplied name
containing U+2028 or U+2029 produced bytes the decoder rejected as invalid YAML.
Agreement between builds did not establish that saved text could be reopened.

## Decision

Preparation serialization preserves these characters as text. A string or
mapping key containing U+2028 or U+2029 is double-quoted and uses `\u2028` or
`\u2029`, respectively. It is never written as a plain scalar or literal block,
even when it also contains ordinary newlines. Do not delete, replace, or
normalize these characters.

This overrides the inherited encoder rule that every non-control character is
written literally for preparation serialization. It introduces no model field,
inferred answer, normalization on reading, or automatic rewrite of other files.
Existing plan and record codecs need their own contract increments before this
rule is applied there; they are not changed by the preparation build.

## Verification

Use hand-specified bytes and decoded text for names, other values, and mapping
keys containing these characters. Include a value with both an ordinary newline
and a Unicode separator. Check each codec independently against the expected
fixture, not only against the other implementation. P4 requires preserving the
supplied text, not merely having both builds agree.
