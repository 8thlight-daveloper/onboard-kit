---
name: writing-issues
description: Write, split, and size GitHub issues for this repo so each one is small and the sizes mean the same thing. Use when drafting a new issue, when an issue feels big, when splitting an epic, or when setting Size, Priority, or Status on the project board.
---

# Writing issues

An issue is one unit of work a person can review in one sitting and accept or reject. Small issues get reviewed, finish, and unblock each other. Big ones sit in progress and hide their real size. This skill says how to cut issues, how to size them, and how to keep the board honest. The mechanics of creating one, such as the templates and `gh` flags, are in the Issues section of `AGENTS.md`.

## Rules

1. **One verb and one noun in the title.** The naming rule in `AGENTS.md` applies: a name that needs "and" is two units. "Plan file: save, reopen, and assemble" is three issues. A list in a title is a list of children.
2. **Cut on a seam, not on size alone.** A seam is a place where one slice can be done, reviewed, and used without the next. Look for these, in order:
   - **Verb:** save, open, assemble.
   - **Component:** one issue per component, never two.
   - **Stage:** a contract, then its build, then its hardening pass (the `contract-hardening` skill) are three issues.
   - **Scenario:** at most three acceptance scenarios in an integration issue.
   - **Document:** one doc, one decision, or one stack row per issue.
   - **Criterion:** more than five done criteria is a sign of more than one issue.
3. **Every leaf issue is XS, S, or M.** L and XL are not allowed on a leaf. L means "split me". XL means epic.
4. **An epic has children and no work of its own.** It names an outcome, lists its sub-issues, and closes when they all close. It gets no Size and no done criteria of its own beyond "all children closed". Use the `epic` label.
5. **Done criteria are checks a machine can run.** Exceptions: "A person accepted it" on a decision or a contract, and a review by a named kind of person on content.
6. **Link, do not copy.** Link ADRs and docs. Do not paste their text into an issue.
7. **Say what is out of scope.** The next slice belongs to a sibling issue, and the out-of-scope line names it.

## Sizes

Size is the review surface: what a person has to read to accept the work. It is not a time estimate. Our own history sets the scale.

| Size | A leaf issue of this size is | Calibration |
| -- | -- | -- |
| XS | Two files or fewer, no new behavior: a stack row, a doc fix, a config line | The `testing.md` pointer |
| S | One decision with a few options, or one contract with 15 cases or fewer, or one build under 500 lines, or one doc under 100 lines | The encoder contract (14 cases, 7 fixture files) |
| M | One contract with 50 cases or fewer, or one build under 1,200 lines, or one integration with three scenarios or fewer, or one doc under 300 lines | The encoder build (657 lines), the decoder build (1,055 lines) |
| L | Split it. Not a valid size for a leaf | The decoder contract as it was filed, with 89 cases and 94 fixture files. It should have been a draft (M) and a hardening pass (M) |
| XL | An epic | The plan file, #23 |

Two checks when the size is unclear:

- If a person could not finish reviewing it in one sitting, it is not M.
- If two reasonable people would size it differently by more than one step, the scope is vague. Narrow the scope until they agree.

## Splitting an issue that already exists

1. Decide the seams first and write the children as a list.
2. Keep the original issue number for the slice that is already in progress or most advanced. Retitle it to that slice, and move its criteria.
3. Create the other slices. Make each one a sub-issue of the same parent (`--parent`), and set `--blocked-by` for each real dependency.
4. If nothing stays in the original, make it the epic. Add the `epic` label, list the children, and drop its own criteria and Size.
5. Leave a comment on the original saying what was split and where each criterion went, so nobody loses a criterion.

## The board

- **Status:** Ready means every blocker is closed and the issue is sized XS, S, or M. Backlog is everything else. An issue is In progress when someone is on it, and In review when its change is open for a person.
- **Priority:** P0 is the critical path to the current milestone. P1 is needed for the milestone but not blocking the path. P2 is later. Set it on leaves and epics.
- **Size:** set on leaves only, by the table above.
- **Order inside Ready:** priority first, then the issue that unblocks the most others.

## Before creating an issue

- [ ] The title has one verb and one noun and no "and"
- [ ] It is a leaf and XS, S, or M, or it is an epic with its children listed
- [ ] It names its template kind: decision, contract, build, integration
- [ ] Each done criterion is a check, and there are five or fewer
- [ ] Inputs are links
- [ ] Out of scope names the sibling that takes the next slice
- [ ] `--parent` and `--blocked-by` are set, and Priority and Size are set on the board
