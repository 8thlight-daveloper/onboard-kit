# Arrangement boundary regression provenance

This independently authored public Run regression file uses the existing
guide_test sandbox, fixture loading, snapshot and exact report helpers. It
imports no implementation package and does not inspect implementation state.
No product implementation or wrong-variant bodies were read. No product codec
or renderer supplies expected data.

The inline preparation is hand-authored from the preparation format and encoder
contract: one pending support arrangement in the existing plan.yaml fixture's
declared inactive section, with two declared helper IDs and an original note.
The noncanonical comment, flow style and missing final newline make an unintended
rewrite observable. All local proposals remain valid under the preparation
editor's scoped validation: pending support permits absent owner, reference and
helpers. Changed notes are ordinary text; the replacement helper is declared.

0350 and the extended guide contract require nested helper apply to change only
the local enclosing form, helper back to discard nested changes and enclosing
back to discard the whole local form. Four paths exercise those boundaries with
valid discarded data, then explicitly review/save the unchanged accepted draft.
They require the exact unchanged report, original bytes, same inode, whole-list
draft review and no masked editor refusal. The scalar-only back case is a
positive regression control even though root reports that existing tests already
kill the corresponding wrong variant.

The contract requires canonical ASCII decimal ordinals and the fixed invalid
action line followed immediately by the same selection prompt. The ordinal
matrix separately covers arrangement edit, arrangement removal and helper
removal. Each invalid token ends the available input. The one-byte reader records
stdout at the next read before returning EOF. Assertions require that recorded
output to end with the exact invalid-action/retry block and contain exactly two
selection prompts. No later scripted action can undo or mask an early selection.
Tokens cover blank, zero, leading zero, signs, ASCII and Unicode whitespace,
suffix truncation, decimal/exponent syntax, unsigned overflow, non-ASCII digits,
out-of-range values and ID text. Canonical selections are separately exercised
by every valid-discard path and by entry into the helper form.

Root reported full-suite survivors for early enclosing apply and two permissive
ordinal variants. That report motivated these cases; it is not execution evidence
from this author. Root must verify the original passes and the wrong variants
fail. This author ran no Go, build, Nix or test command.
