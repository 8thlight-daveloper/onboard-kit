# Composition and projection regression provenance

This independently authored public Run regression file uses existing guide_test
sandbox, exact terminal-report, budget padding and counting-reader helpers.
Only normative guide, editor, preparation validator and encoder contracts and
existing contract tests were read. No product implementation or wrong-variant
bodies were read, and no expected value came from a product codec or renderer.

The multibyte budget case uses the existing byte-sized padding builder with
Unicode enabled. Every logical line remains within 64 KiB. The complete review
line ends at exactly 1,048,576 consumed bytes while save remains queued. 0350
and the guide contract require guide.input before another reader call or any
publication. The test also asserts the review-complete prompt was displayed,
zero reads after exhaustion, the whole save line still unread and identical
filesystem snapshots. This distinguishes byte accounting from rune accounting
on the refusal side, complementing the existing Unicode exact-budget save case.

The composition case stages an explicitly valid pending support arrangement
first, then applies a recipient name. The editor operates on the current
accepted preparation, preserving previously accepted arrangements. The expected
combined YAML is hand-written from preparation encoder key order and omission
rules; the current review must display the recipient change and whole held
arrangement. This complements recipient-first binary coverage.

The diagnostic cases use actual invalid recipient references, a new arrangement
with explicit valid ID/section but absent kind/status, and held invalid recipient
references during actual review. Preparation validator order supplies manager
before buddy and kind before status. Paths are fixed field paths. The guide
contract permits only safely quoted severity/code/path triples; dependency
messages are never printed. Assertions include the preceding fixed heading and
the immediately following form/review block, so extra quoted messages (including
empty messages), omitted diagnostics and interleaved output cannot satisfy them.
Rejected private scalar values may appear in their authorized editable forms;
these tests judge the diagnostic blocks, not a fabricated global value ban.

Root reported possible focused survivors before these tests were authored;
execution remains root's responsibility. No Go, build, Nix, test or mutation
command was run by this author. Only Go formatting was performed.
