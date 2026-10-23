# 0350. Guided support arrangements

Status: accepted by agent judgment; contract and implementation pending

Builds on [0140](0140-person-specific-preparation.md),
[0150](0150-onboarding-library-and-preparation-flow.md),
[0160](0160-personal-data-and-preparation-approval.md),
[0190](0190-incremental-preparation-repair.md), and
[0340](0340-guided-preparation-save.md).

## Context

The recipient guide can stage details and save one reviewed preparation. An
onboarder still needs separate explicit commands to assign a partner, arrange
access or supply a deployment opportunity. These commands already have approved
data and editor rules. A guided form should use those rules rather than invent
another arrangement format or claim that assignments verify newcomer work.

## Decision

Extend prepare guide with an arrangements action, alongside recipient, review
and cancel. This extends 0340's recipient-only scope and input limits.
Keep its explicit account/root, dependency boundary, privacy notice, source
snapshots, reviewed publication, ACL protection, cleanup and reporting rules.
No constructor, saved format, dependency position or artifact owner changes.

### Explicit arrangement choices

Offer new, edit, remove and back. Show the held arrangements in their current
order with quoted permanent IDs, section references, kind and preparation
status. Numbered menu positions identify an explicit selection in this open
form only. Never store a list position as an arrangement or activity identity.
An empty menu line selects nothing.

Creation requires a supplied permanent ID. Refuse an already used ID rather
than silently editing that item; the person can choose edit explicitly.
This is a driver-local creation guard, not a change to the editor's upsert API.
Existing IDs cannot be renamed in
place; correcting an invalid ID means explicitly removing it and creating the
replacement. Editing starts with the whole selected arrangement, not a sparse
update. Fields cover section, kind, accountable owner, assigned contacts,
reference, target date, preparation status and note. Keep, change and clear have
the same deliberate meaning as the recipient form. Back discards the local
form. A failed apply keeps its proposed corrections available.

Show declared section IDs and titles, directory contact IDs and names, and
supported kind/status spellings where relevant. A declared inactive activity
may still have preparation. Listing it does not claim that it is active,
authorized or ready. Require explicit exact references and explicit kind/status;
do not guess from catalog text, role names, contact order or a date.

Assigned contacts are an ordered list, not a set. Provide explicit add, remove,
clear, apply and back actions in a local helper-list form. Each added ID is
literal text at its value prompt, including menu-looking words. Removing one
listed occurrence does not remove another equal ID. Preserve repeats and order
under the existing format. Apply copies the proposed helper list into the local
arrangement form only. Back discards helper-list changes and restores that local
form's previous helper list. Neither action applies or saves the enclosing
arrangement. Back from the enclosing form discards all its local changes.

Apply a completed arrangement through the real editor. Its scoped validation
remains authoritative: preserve unrelated invalid arrangements and recipient
facts. Do not replace a failed operation with a partial edit or infer pending
for an absent status. An arranged item must satisfy the existing completeness
rules. Arranged is not verified, and a passed target never fails an activity.

### Removal is a staged change

Select the exact existing item visibly, then explicitly confirm remove or go
back. This permits removal of one uniquely identified malformed or empty ID;
missing input is not selection of an empty ID. Duplicate permanent IDs still
refuse at the existing startup ambiguity guard, before a menu can choose a winner.

Use the real RemoveArrangement operation. Removal changes the in-memory draft
only and preserves the remaining order. Cancel before save must preserve every
original byte, including a staged removal. Do not delete an activity, contact,
answer, outcome or another person's file as a side effect.

### One complete review and one save

Review includes the complete original and proposed arrangement lists as separate
clearly labeled blocks, alongside all existing recipient changes and actual
reviewer results. Display every known arrangement field, ordered helpers and
intentional absence; quote values with the existing safe display rules. A removed
item remains visible in the original block. Preserve multiline held text as
literal display, and do not print unknown preparation data.

An accepted arrangement change invalidates earlier review. Only save at the
current complete Review prompt publishes the final draft. All forms are unsaved;
no local apply or removal confirmation writes files. Existing semantic no-op
preserves original bytes and inode. Changed drafts still require the real encoder
and the existing 1 MiB output budget, without an added readiness gate.

### Bound the interaction

The richer flow can accumulate many arrangements and helpers rather than a fixed
number of scalar fields. Keep memory bounded even for a scripted input stream.
In addition to the 64 KiB logical-line bound, cap consumed interaction input at
1 MiB per Run, including menu actions and LF/CRLF bytes. Help and usage still
consume no input. Explain the fixed interaction limit before offering forms.
A complete save inside that budget remains valid; if another
read is needed after the budget is consumed, refuse guide.input before reading
or publishing further. Do not prefetch consent. Unconsumed trailing input after
an exit does not count as consumed data. This is a resource limit, not a stored
format limit or permission to truncate a value. At exhaustion this refusal takes
precedence over an unread EOF or cancel; EOF below the budget retains ordinary
cancellation behavior. A person can reopen and review
the original preparation after refusal; there is no hidden partial save.

### Keep later work separate

This increment edits explicitly supplied arrangements only. It does not infer
activity-to-preparation relationships, create repeated activity instances,
select templates, bind recipient facts into answers, choose export help routes
or send an invitation. Existing invite export remains separate and explicit.
Those later steps keep their own decision and contract boundaries.

## Rationale and alternatives

P1 and P3 of [0100](0100-behavior-judgment.md) favor helping the onboarder supply
known support through the same flow. P2 and P4 require explicit choices, stable
identities and preservation of unrelated work. P6 keeps targets separate from
completion. P7 constrains free text and readership, not a claim that free text
is automatically scrubbed. P8 favors reusing the existing editor and plain-text
interaction rather than a second framework or another file.

Keeping only explicit write commands remains usable, but makes one reviewed
recipient-and-support change require separate publication steps. Automatic
arrangements inferred from activity names would hide missing product bindings
and guess support. A separate wizard would duplicate the save and safety boundary.
The selected direction adds a bounded form to the existing driver instead.

## Verification and delivery

An independent contract author specifies exact prompts and selection framing,
literal payloads, diagnostic projection, limits and whole-list review before a
different author changes implementation. Include append, in-place replacement,
ordered repeated helpers, helper-form back, empty versus clear, invalid whole
form correction, removal of a unique empty ID, removal cancellation, duplicate
guard, stale references elsewhere and preservation of unknown data/records.

Pin actual editor composition and literal saved bytes, not a replacement editor
or an implementation-generated expected file. Broken variants must demonstrate
loss of a helper, defaulted status, replacement by append, unintended rename,
early removal publication, lost unrelated stale work and save outside Review.
Whole-binary scenarios must exercise the new root path with existing recipient
operations and the unchanged separate export boundary. First run the expanded
contract against the old recipient-only implementation to show the new behavior
is missing; a second independent implementation and differential checks follow.

Input-budget tests need exact-boundary saves and refusal before an additional
read when the budget ends at a complete preceding line, leaving queued consent
untouched. Also test a budget ending partway through a line: a prefix may already
have been consumed, but no further byte may be read and no partial save is
consent. Both refusals preserve original files. Re-run source recheck,
no-op, real ACL inheritance, failed display and reporting safeguards. Execution
evidence and agent judgments remain distinct from human approval. This ADR
records direction, not a delivered or verified arrangement editor.
