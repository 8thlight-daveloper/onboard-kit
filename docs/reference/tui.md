# The terminal walkthrough

Status: living. Each point says whether it is decided, a direction, or open, and any of them can change as the screens are built and used. This page is not a contract. Unit tests drive the screens through key and window-size messages, and a person tries them in a real terminal.

The code is `internal/tui`, opened by `onboard-kit walk`. The toolkit is Bubble Tea, see [language](../stack/language.md). The people it serves, and what each may answer, are in [0055](../decisions/0055-variable-audience.md) and [0060](../decisions/0060-personal-plan.md).

## Who uses it

| Person | Command | Sees | Changes |
| -- | -- | -- | -- |
| The onboarder | `walk --plan FILE` | Onboarder questions, section switches, the assembled document | Onboarder answers and switches |
| The onboardee | `start`, not built yet | Onboardee questions, the onboarder's answers read-only, their own document | Onboardee answers only |

Decided: the onboardee mode hides section switches. Only the onboarder decides what applies.

## Flow

Direction, wanted by the product owner:

1. **Start.** The onboarder opens on the review screen only when the plan already holds answers, so a returning person sees where they stand. A plan with no answers goes straight to the first question.
2. **Walk.** The person moves from question to question in order, with a clear sense of position, such as "4 of 11". Enter accepts and moves on. A question that depends on an earlier answer appears after it, as the question finder orders them.
3. **Review.** A review screen lists blockers, problems, and deferred questions, and previews the document. From it a person jumps back to any question.
4. **Overview.** The full question list stays available as an overview a person opens from the walk, and does not replace the walk.

Current state: the screen is a list with a cursor. A person jumps to a question and edits it. There is no guided order and no review-first start. This is the main change the redesign makes.

## Layout

Direction. Assume a modern display of 120 to 140 columns, and adapt below that. Bubble Tea sends the window size at the start and on every resize, so the layout follows the window.

| Width | Layout |
| -- | -- |
| 140 or more | Two panes. The left holds the walk or the list. The right holds the detail for the current question: the full description, the answer type, suggestions, the sections that use it, and problem messages. The document preview sits beside the questions on the review screen |
| 100 to 139 | Two panes, with a narrower list |
| 60 to 99 | One column. The detail shows below the list, or on its own screen |
| Under 60, or very short | A plain message asking for a larger window |

These numbers are a starting point for the design pass to refine. Rules that hold at any width:

- Widths come from the window and never from constants. Text wraps at the pane width, and a list scrolls by the real height.
- A line is never wider than the window.

Current state: a single column, a fixed 34-character description column, and truncation. The width is used only for the input box and the scroll window.

## Answering

Decided, and built:

- A boolean or choice is picked from a list, with "deferred, answer later" and "clear the answer" at the end.
- A string takes typed text. Tab cycles through the question's suggestions.
- A command is words separated by spaces, with quotes keeping spaces together.
- Several people are separated by `;`.
- An answer that the validator refuses stays on the edit screen with the problem message.
- Every accepted answer is saved to the plan file at once. There is no unsaved state.

## Keys

Built today: enter answers, `d` defers, `x` clears, `s` section switches, `r` review, `q` quits, up and down or `j` and `k` move, `g` and `G` jump to the first and last question, and escape goes back. The guided flow will change what enter does at the list, and the redesign owns the final keys.

## Testing

Decided. The app is a Bubble Tea model, so tests send key messages and window-size messages and check the plan, the saved copies, and the view. At several widths a test checks that no line is wider than the window and that the key elements are present. A scripted pseudo-terminal run covers start and quit. `.scratch/pty_walk.py` is the example, and it must answer the terminal's background-color query or the program waits five seconds.

## Open

- Splitting this page into one file per view under `docs/design/`, with design decision records (DDRs) for choices that cut across views, such as the width breakpoints. DDRs stay separate from ADRs: an ADR is for a decision that contracts or acceptance tests depend on, and a DDR is for a design choice that only the screens depend on.
- The final key set and the screens' look. A design pass with the `/ux-designer` skill takes the flow and the layout above as its brief.
- Whether the review screen is also the first screen of the onboardee's walk.
- How a person sees the questions they deferred, and how the tool reminds them.
- Whether the document preview scrolls with the walk.
- Color and the no-color case.
