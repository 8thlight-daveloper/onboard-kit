# 0470. Preserve Unicode separators in encoded plans

Status: accepted agent decision. Verification evidence is recorded separately.

Extends0200's preparation rule and0230's record rule to plan serialization only.
Supersedes the plan encoder's literal-character and literal-block eligibility
rules only where they would emit U+2028 or U+2029 literally. Other commitments
of those decisions and the complete existing plan encoder contract remain.

## Context

The creator accepts an explicit opaque Unicode PlanID. An independent fixture
reached correct terminal quotation, but a generic YAML parser could not read the
saved plan when the ID contained Unicode line or paragraph separators. Both
independent creator builds exposed the same upstream plan encoder issue. Their
agreement and differential success did not prove the plan bytes could be reopened.

The approved plan encoder contract says `Every other character is written as it is`, does not exclude these separators from plain scalar eligibility, and
can select a literal block when the same text also holds LF.0200 deliberately
left plan codecs for a separate increment;0230 did so again. Applying their rule
to plans therefore needs this explicit contract revision, not a claim
that the old literal rule already promised the new byte spelling.

## Decision

Every plan string or mapping key containing U+2028 or U+2029 is double-quoted
and uses explicit lowercase `\u2028` or `\u2029`, respectively. Never emit these
characters literally, in a plain scalar, or inside a literal block, including
when the same text also contains LF. Preserve exact text, sequence order,
repetitions, original byte-order map sorting and all unrelated canonical rules.

This applies uniformly to typed scalar fields, string-valued enums/dates,
directory facts/holder keys, variable metadata/defaults/options/conditions,
scoped answer keys and Text/List values, section/step/check/command strings,
Source fields and unknown strings/keys at every supported nesting depth.
The encoder remains representational: no semantic validation or new refusal
of an otherwise representable Unicode value is introduced.

P2 respects the explicit supplied value; P4 preserves work instead of losing or
rewriting characters.0200 already provides the narrow successful representation
direction. Rejecting Unicode IDs in creator would put a codec defect into a
driver restriction and leave other fields broken. Deleting separators, changing
them to LF/spaces, or ASCII-escaping all unrelated Unicode would lose data or
broaden the change. None is chosen.

No model/schema/API/decoder change, normalization on reading, hash algorithm or
source-baseline recomputation is authorized. This deliberately changes canonical
plan serialization only for strings/keys containing these two characters; it
does not claim the former invalid or lossy byte spellings remain canonical.
Held Source.Hash/Origin.ContentHash fields remain exact ordinary string data.
No automatic rewrite or filesystem operation occurs in this pure component.

## Independent acceptance and evidence status

Hand-authored exact canonical templates and exact escaped payload literals cover
both separators, ordinary/quoted strings, LF combinations, typed fields, keys,
Text/List values, commands and nested unknown data. A generic YAML parser must
read the bytes into independently expected mappings and exact original text.
Unrelated printable Unicode and ordinary LF literal blocks keep their original
canonical bytes. The creator assertion remains unchanged as an integration
regression; no narrow ID refusal works around the upstream issue.

Verify assertion failure against zero output and the former actual encoder for
the new separator behavior, while proving the ordinary-Unicode control passes.
Keep fixture and implementation authors independent, rerun the complete encoder
suite, and retain the creator regression. These checks remain separate from
native publication and release verification. No ongoing authority or survivor
acceptance exception follows from this decision.
