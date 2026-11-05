# Plan store driver

Status: idea, 2026-11-05.

## The problem

Writing a new plan file is done by `create.WriteNew` in `internal/command/create`,
and by the front door's port that calls it. Both are in the `command` kind, which has
no contract folder. So nothing pins what the writer does when the folder is missing,
when the file exists, or when a permission is denied. A new plan in a folder that did
not exist failed with a raw error that named a temporary file. The fix has unit tests
and an acceptance scenario, but a rebuild of those packages from the decisions would
not know to keep it.

0610 says screens and flags call shared operations, and `architecture.yaml` keeps file
access out of fronts. A writer that lives in a front breaks both.

## The idea

A driver, in the style of `preparation/store`, with its own contract. A name such as
`plan/store` follows the preparation family: it lists, loads, and saves plan files under
native guards. It would hold the behavior that the front door and `new` need today, and
the behavior that 0660 and 0670 will add.

What the contract should pin:

- **A new plan.** Create exclusively, never replace, in a folder that exists. A missing
  folder is refused unless the caller says it may create it, with the same code and
  consent shape as `store.directories-missing` in the preparation store.
- **A file where a folder should be,** and every other refusal, with a stable code and
  the plain sentence that goes with it.
- **An atomic publish.** A private temporary file in the same folder, a no-replace
  publish, and cleanup of a leftover. Nothing half-written is ever visible.
- **Saving an existing plan.** The byte recheck since the load, so a plan changed by a
  pull or another tool is reported stale and reloaded, as 0630 decides for preparations.
- **Native details.** Access controls kept or the write refused, no following of
  symlinks, and a volume check, as the other stores do.

Callers would be the front door's new-plan flow, `new --out`, the plan editor's autosave
in `walk`, and the refresh from a plan repository in 0670.

## Related work

- Records need the same treatment. 0670 moves a person's record to the tool's data
  folder, and its writes are the same pattern. `record/store` in the rebuild of
  `session/*` would share the design.
- The plan list file is written by the front door's registry code, also in a `command`
  package. It is a smaller case of the same problem.
- Once this exists, `internal/command/create/exclusive.go` goes away.

## Not verified

- Whether `catalog/creator`'s existing publication code can be reused as the base, since
  its contract already covers same-parent publication and a missing parent.
- The exact set of refusal codes, which the contract draft would settle.
