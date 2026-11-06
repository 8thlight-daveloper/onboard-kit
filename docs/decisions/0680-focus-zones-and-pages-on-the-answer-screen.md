---
category: ux
status: proposed
affects: [tui]
builds-on: ['0520', '0610', '0630', '0640']
---

# 0680. Focus zones and pages on the Answer screen

Builds on [0520](0520-preparation-screen.md),
[0610](0610-people-work-in-screens-and-flags-serve-scripts.md),
[0630](0630-the-preparation-tab-saves-what-was-reviewed.md), and
[0640](0640-the-invitation-tab-writes-an-invitation.md). It records two changes to how a
person answers questions, and the choices made while building them. The person who owns
the decisions chose the design and approved the build. The details an agent chose are
proposals until that person accepts them.

## Context

Answering a text question locked the screen. A text field has to take every printable key,
so the one-letter tab keys stopped working, and only `ctrl` chords and `esc` got a person
out. The actions defer, skip, and clear existed for choice questions as list items and hid
behind chords for text questions. A person also pressed shift+tab to change tabs, and it
meant "back one question". Nothing was learned yet, because the tool is unreleased, so key
meanings could change.

An answer with several parts, such as a repository address and a file path, had to be
entered as one question per screen. The tool had no way to show them together.

## Decision

**The Answer screen has three focus zones, moved with arrows.** The tab strip, the answer,
and the options. Up and down walk one order: the tab strip, then each row's field, then
that row's options, then the next row. Up from the first row reaches the tab strip. Left
and right choose a tab there, and `enter` opens it.

**`esc` stops typing, and it must be unmistakable.** The caret goes, the field dims and
keeps its text and shows a label ("Not saved yet: Glob", "Saved: Globex", or "Nothing typed
yet"), the chosen list row gets a highlight bar, the hint line starts "Not typing.", and the
status line says the text is kept. A second `esc` goes to the Overview. While not typing,
the one-letter keys work again: `p n d x q ?` and the tab letters.

**The options are a vertical list under the field.** It holds "Edit the answer", shown only
once typing stops, one "Use the suggestion: X" row for each suggestion, "deferred, answer
later", "skip, change nothing", and "clear the answer". These are the items a choice
question already showed, so both kinds of question behave the same. Suggestions are rows,
so no key is needed to try one.

**Tab and shift+tab move between screens.** They wrap at the ends, and they work while
typing. Inside a form with several fields they move between fields, and the tab strip is
reached with up from the first field. Back one question is `p` or left when not typing, and
`ctrl+p` while typing. Skip is `n` or right, and `ctrl+n` while typing. Left and right
switch the two panes of the Review, the Preparation review, and the Contacts lists.

**A page is one scope.** The plan's own questions together, or one section's questions
together. Each question is a row with its description, an example, a mark for required, and
its answer. The focused row is an input. A row not in focus shows its state: "Saved: X",
"Not saved yet: X", "Deferred, answer later", "Required, not answered yet", or "Not answered
yet". A choice, contact, role, or team row shows its pick widget under the row while
focused. `enter` saves a row and moves to the next, and after the last row it opens the
next page. A page of one row looks like the old screen. The question's `Group` is not a
page. It means the same variable name and type in different scopes, and it feeds the
suggestions.

**Rows have a plan-wide number,** and the position line reads "Question 7 of 8". The count
line reads "1 of 4 answered on this page". A page that does not fit scrolls from a row
boundary, with "N more questions above" and "below". A row whose `askIf` changes appears or
leaves in its place, and focus stays on the same question. The newcomer's pages have the
same rows.

**A draft is kept while the person is on the page.** Typed text stays in its row, including
across a visit to another screen. Defer saves the deferral and drops that row's draft. Skip
and back save nothing and drop it. Clear empties the box and the saved answer and returns to
typing. Leaving the page drops unsaved drafts, and the status line starts with "warning:"
and names the questions.

## Consequences

- Many wireframes and a few scenarios changed: the answer-screen hint text and position
  line, the progress count, and one `Tab` key press. The scenarios no longer wait on a
  count that follows catalog content.
- The status line read "saved. saved" after every `enter`, and that is fixed.
- The terminal page, `docs/reference/tui.md`, describes the keys and the pages.

## Not decided

1. **Asking before dropping drafts.** `enter` on the last row of a page, and a move to
   another page, drop unsaved text in other rows with only a warning. Asking first is
   safer and is a change of behavior.
2. **Quit with a draft.** Quit drops a draft without a warning.
3. **A key for the newcomer's Your questions screen.** `w` opens it for both audiences.
   `u` is the candidate.
4. **Esc in the Contacts, Preparation, and Invitation forms.** It still cancels the form
   and does not stop typing first, and the Invitation form cannot reach the tab strip,
   because reopening that tab drops it.
