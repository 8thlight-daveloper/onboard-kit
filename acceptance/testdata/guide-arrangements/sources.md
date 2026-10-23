# Guided arrangement binary fixtures

These fixtures and six new scenarios are independently hand-authored whole-binary
acceptance data. No guide implementation, implementation builder notes, draft
component tests or draft expected payloads were read. Expected YAML was written
from the format and encoder contracts, without invoking a product codec.
All identities and descriptions are fictional examples.

## Sources

- [0350](../../../docs/decisions/0350-guided-support-arrangements.md) supplies
  arrangement forms, ordered repeated helpers, permanent identity, staged
  removal, complete review and the aggregate interaction budget.
- [0340](../../../docs/decisions/0340-guided-preparation-save.md) supplies
  recipient editing, current reviewed consent, byte-preserving no-op and
  separate artifact ownership.
- [0190](../../../docs/decisions/0190-incremental-preparation-repair.md) and the
  [editor contract](../../../contracts/preparation/editor/contract.md) supply
  in-place whole replacement, scoped preservation and malformed-ID removal.
- The [preparation format](../../../docs/reference/preparation-format.md),
  [preparation encoder](../../../contracts/preparation/encoder/contract.md) and
  [plan encoder](../../../contracts/plan/encoder/contract.md) supply literal
  expected YAML: known-field order, omitted empty fields, retained list order,
  quoted dates and sorted unknown mappings.
- The [guide contract](../../../contracts/preparation/guide/contract.md) supplies
  exact line actions, ordinal grammar, prompts and complete arrangement blocks.
  Authoring used the extended prose/API in the frozen-contract scratch area.
  Its SHA256 is
  `553609cfb9254e8e5b03cd0b42209b61f412d2f2005dfbf280501a773c40a74e`.
  The integrating author confirmed that Markdown normalization changed the
  initially supplied hash, without changing the protocol or intended behavior.
- The [invitation command](../../../contracts/command/invite/contract.md) supplies
  the separate explicit export refusal for an absent pre-access helper.
- Existing acceptance runner schema and matching code supplied unchanged data
  syntax: one contiguous stdout substring per step, JSON subset comparisons,
  exact same-as file bytes and before/after unchanged file bytes.

## Fixture provenance

| Fixture | Independent construction |
| -- | -- |
| plan.yaml | Minimal plan based on the shape of the existing preparation plan fixture. Adds declared literal action-word contact IDs and retains an inactive declared section; omits unrelated questions. |
| full.yaml | All arrangement fields entered explicitly by the new scenario, including ordered save/back/save/clear helpers and a past leap-day goal. |
| stale.yaml | Whole held pairing plus a stale separate section, stale recipient manager, multiline note and nested unknown data. |
| cleared.yaml | First stale.yaml arrangement replaced at its original index with explicit clears and changed section/kind/status; all unrelated values retained. |
| helpers.yaml | full.yaml with only the first helper occurrence removed; equal save helper remains in position after back. |
| removable.yaml | Two unique invalid identities followed by a valid support arrangement and unknown data. Empty identity includes a writable note. |
| removed.yaml | removable.yaml with the explicitly selected empty and malformed entries removed in sequence; retained item and unknown data unchanged. |
| mixed.yaml | One recipient and one arrangement accepted after an earlier review is left; one final current review saves their combined value. |

The helpers scenario embeds deliberately noncanonical flow YAML with a comment
and missing final newline. Its first step stages and discards nested and enclosing
changes, and applies an unchanged helper form; the unchanged check protects exact
source bytes. The second step removes one repeated occurrence and compares the
whole canonical result against helpers.yaml.

The edit scenario includes an unsupported edit-mode id action before valid
field actions. The fixed identity and in-place replacement are checked by the
complete expected file and review block. The remove scenario rejects blank and
leading-zero selections before an explicit selection; confirmation back preserves
the item, confirmed removal remains cancellable, and a later save removes both
unique invalid identities. The collision scenario supplies a complete valid
replacement-shaped proposal, so an upsert would change the accepted value and
fail the expected unchanged result.

The mixed scenario first reviews recipient changes, returns to draft, receives a
refused partial arrangement, supplies explicit status, then tries save at Draft
before the final Review/save. Literal saved bytes and the contiguous combined
recipient/list review assertion require the final draft. A subsequent explicit
invite export refuses an unknown pre-access helper and creates no archive.
Malformed selected-person records and other-person files remain byte-identical.

The export refusal retains exactly five `directory.contact-unreachable` warnings:
each declared contact in plan.yaml lacks email, slack and slackId. The
[plan validator](../../../contracts/plan/validator/contract.md) requires one
warning per such contact. The invitation command's phase 4 retains plan warnings,
and its reporting rules carry prior warnings through the later pre-access-help
refusal. The five identical public projections deliberately assert that warning
multiplicity, without exposing contact identities. Section authoring checks are
outside this plan validator's stated scope; missing section metadata is not the
source of these warnings. This corrects the original empty-problems expectation
from the normative contracts after integration reported a count mismatch; the
plan and export boundary assertions are unchanged.

## Coverage boundary and handoff

These bounded scripts exercise real root dispatch, real component composition,
ordinary line input, current review, saved bytes and ownership. They do not prove
inode preservation, absence of arbitrary newly named files, temporary cleanup,
native ACL behavior, write failures, races, exact 64 KiB line limits or exact
1 MiB interaction boundaries. Those require the separate component reaching
fixtures; no huge binary fixture was added merely to duplicate the budget tests.
The runner's unchanged check compares bytes and existence, not inodes.

The updated help scenario is copied verbatim from the existing scenario except
for its single exact expected help prose line. Its dispatch and refusal steps
are retained.

No binary, Go, Nix, build, test or mutation command was run by this author.
Execution and old-binary red evidence belong to the integrating author. This
manifest records provenance and coverage, not execution success or human approval.
