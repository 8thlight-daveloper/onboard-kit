# The terminal walkthrough

Status: living. Each point says whether it is decided, a direction, or open, and any of them can change as the screens are built and used. This page is not a contract. Unit tests drive the screens through key and window-size messages, and a person tries them in a real terminal.

The code is `internal/tui`, opened by `onboard-kit walk` and `onboard-kit start`. The toolkit is Bubble Tea, see [language](../stack/language.md). The people it serves, and what each may answer, are in [0055](../decisions/0055-variable-audience.md) and [0060](../decisions/0060-personal-plan.md).

## Intent

Decided.

|  | Onboarder | Onboardee |
| -- | -- | -- |
| Who | A consultant who builds the plan for an organization | A newcomer who answers a few questions about themselves |
| What they try to do | Answer every required question, decide which sections apply, and hand over a plan with nothing missing | Answer their own questions once, quickly, and see their own onboarding document |
| What they should feel | In control. They always know what is left | Welcome and unhurried. Nothing looks like a form to fear, and nothing they type can break the plan |

Rules that follow from this:

- One question at a time, in order, with a count. A person never has to wonder what to do next.
- Nothing is lost. Every accepted answer is saved at once, so quitting is always safe.
- A person can step back, skip, defer, and see the whole list at any time.
- The newcomer sees only their own questions. The onboarder's answers are in the document, and they cannot be changed from here.

## Who uses it

| Person | Command | Sees | Changes |
| -- | -- | -- | -- |
| The onboarder | `walk --plan FILE` | Onboarder questions, section switches, the assembled document | Onboarder answers and switches |
| The onboardee | `start --org-plan FILE` | Onboardee questions and their own document | Onboardee answers only |

Decided: the onboardee mode hides section switches. Only the onboarder decides what applies. Built: `start` makes the newcomer's own plan file from the org plan and opens `walk` for the onboardee audience. Refresh from a changed org plan, a URL, and the locks are not built, see 0060.

Decided: the onboardee does not get a separate list of the organization's answers. The document preview already shows them in context, and a second list would repeat it.

## Screens

Decided. There are four screens and a help page. Every screen has a header line, a body, a status line, and a key hint line, in that order. The hint line is always the last line.

| Screen | Holds | Opened by |
| -- | -- | -- |
| Walk | One question: its position, its section and group, what it asks, whether it is required, its current answer, the way to answer, suggestions, problems | The start, enter on an overview or review row, finishing a question |
| Overview | Every question in order with a mark for its state. The cursor jumps to any of them | `o` |
| Review | Blockers, deferred questions, validator problems, and the document preview | `r`, finishing the last question, or the start when the plan holds answers |
| Sections | Section switches. The onboarder only | `s` |
| Help | Every key of the current screen | `?`, or `f1` while typing |

How a person moves:

- Start: if the plan holds an answer to any question of this audience, open Review. Otherwise open Walk at the first question. A plan with no active questions opens Review.
- Enter accepts the answer on the screen and moves to the next question in order. This was a choice between the next question and the next unanswered one. The next in order is predictable, and a person who wants an unanswered one opens Review, where the blockers are listed.
- Past the last question, the person lands on Review.
- `p`, left arrow, or shift+tab goes back one question. `n` or right arrow skips forward without changing anything.
- `d` defers the question and moves on. `x` clears the answer and stays, so the person can answer again.
- Escape on the Walk returns to the Overview or Review, whichever sent the person there. Escape on Overview or Review returns to the Walk.
- From Overview or Review, enter opens the chosen question in the Walk. `w` on Review returns to the Walk at the first unanswered question.
- `q` quits from the Overview, Review, and Walk, except while typing. The Sections screen takes `q` as back. Ctrl+c always quits.

## Layout

Decided. Bubble Tea sends the window size at the start and on every resize, and the layout follows it. The panes are separated by a vertical bar.

| Width | Walk | Overview | Review |
| -- | -- | -- | -- |
| 140 or more | The question list on the left, about a third of the width and at most 52 columns. The question and its detail on the right | The list on the left. The detail of the chosen question on the right | Items on the left. The document on the right |
| 100 to 139 | The same, with the list at about a third of the width, at least 30 columns | The same | The same |
| 60 to 99 | One column: the question, then its detail if the height allows | One column. Each row shows its answer when there is room | One column. Tab switches between the items and the document |
| Under 60 wide or under 12 high | A plain message asking for a larger window |  |  |

At 140 and more the right pane has room to wrap less. Rules that hold at any width:

- Widths come from the window. Text wraps at the pane width, and a list scrolls by the real height around the chosen row.
- A line is never wider than the window.
- Rows that do not fit are cut with an ellipsis, and the key hint line drops its least useful hints before it drops `? help`.

## Wireframes

Decided. No wireframe tool fits a terminal screen, so the screens are drawn by the code itself. A test renders the real view from a fixed plan with no colour and compares it to a text file. The files are the wireframes, and they cannot drift from the screen. Regenerate them on purpose with `go test ./internal/tui -run Wireframes -update`, and review the diff.

| Screen | 80 columns | 140 columns |
| -- | -- | -- |
| Walk, choosing | [walk](../../internal/tui/testdata/screens/walk.80.txt) | [walk](../../internal/tui/testdata/screens/walk.140.txt) |
| Walk, typing | [edit](../../internal/tui/testdata/screens/edit.80.txt) | [edit](../../internal/tui/testdata/screens/edit.140.txt) |
| Overview | [overview](../../internal/tui/testdata/screens/overview.80.txt) | [overview](../../internal/tui/testdata/screens/overview.140.txt) |
| Review | [review](../../internal/tui/testdata/screens/review.80.txt) | [review](../../internal/tui/testdata/screens/review.140.txt) |
| Sections | [sections](../../internal/tui/testdata/screens/sections.80.txt) | [sections](../../internal/tui/testdata/screens/sections.140.txt) |
| The onboardee's walk | [onboardee](../../internal/tui/testdata/screens/onboardee.80.txt) | [onboardee](../../internal/tui/testdata/screens/onboardee.140.txt) |
| Window too small | [too small](../../internal/tui/testdata/screens/too-small.80.txt) | [too small](../../internal/tui/testdata/screens/too-small.140.txt) |

## Colour

Decided. One accent, the Branded Blue `0071CE`, marks the focus: the cursor row, the headings, and the key names in the hint line. Two other brand colours mean status and nothing else: Orange `FF8300` for a warning, Red `F32735` for a problem. Everything quiet is drawn with the terminal's faint style, so it works on a light or a dark background and the tool never asks the terminal for its background colour.

- A terminal that reports true colour gets the hex values. One with 256 colours or 16 colours gets the nearest ANSI value. `NO_COLOR` and a terminal with no colour get plain text.
- Colour never carries meaning alone. The cursor row starts with `>`. States are words and marks: `[x]` answered, `[~]` deferred, `[ ]` not answered, `*` required. Problems start with `problem:` and warnings with `warning:`. The focused pane of the Review shows `>` before its title.
- A test renders every screen with no colour and checks that these marks and the key hints are still there.

## Emphasis follows focus

Direction, from the first round of use. The first version drew the left column in bright text all the time, which pulled the eye away from the question being answered.

- The pane a person is acting in is bright, and the other pane is drawn faint.
- While typing or choosing an answer, the detail pane is bright and the list is faint. The list is context there, not a control.
- While the list is what a person is moving through, as in the overview, the list is bright and the detail pane is faint.
- The accent colour marks "you are here" in the focused pane, so there is one bright spot at a time. It comes out of titles. Whether it stays on the position counter, such as "3 of 9", and on section names, such as "If you are stuck", is left to the design pass.
- No colour still reads, because the cursor mark `>` carries the focus. Brightness only reinforces it.
- Golden wireframes are plain text and cannot show this. A style test checks which pane is drawn bright in each state.

## Answering

Decided, and built. The answer is made on the Walk, under the question.

- A boolean or choice is picked from a list, with "deferred, answer later" and "clear the answer" at the end. The cursor starts on the current answer. Up and down move, enter picks.
- A string takes typed text, with the current answer filled in. Tab cycles through the question's suggestions. Enter on text equal to the current answer moves on without saving. Enter on an empty box clears the answer. Ctrl+d defers.
- A command is words separated by spaces, with quotes keeping spaces together.
- Several people are separated by `;`.
- An answer that the validator refuses stays on the Walk with the problem message under the answer.
- Every accepted answer is saved to the plan file at once. There is no unsaved state.
- While a person types, letters are text, so the single-letter keys are off. Back, forward, defer, and help have control-key forms listed below.

## Keys

Decided. The key hint line shows the most useful keys for the screen and the width, and `?` shows all of them.

| Key | Walk, choosing | Walk, typing | Overview | Review | Sections |
| -- | -- | -- | -- | -- | -- |
| enter | Pick, move on | Accept, move on | Open the question | Open the question | Switch on or off |
| up, down, `k`, `j` | Move the choice |  | Move | Move | Move |
| `p`, left, shift+tab | Back | shift+tab, ctrl+p |  |  |  |
| `n`, right | Skip forward | ctrl+n |  |  |  |
| `d` | Defer, move on | ctrl+d |  |  |  |
| `x` | Clear | Empty the box and enter |  |  |  |
| tab |  | Try a suggestion |  | Items or document |  |
| `o`, `r`, `s` | Overview, review, sections |  | `r`, `s` | `o`, `s` |  |
| `w` |  |  |  | Walk |  |
| pgup, pgdn, `g`, `G` |  |  | `g`, `G` | Scroll the document |  |
| esc | Back to overview or review | Same | Walk | Walk | Back |
| `?` | Help | f1 | Help | Help | Help |
| `q` | Quit |  | Quit | Quit | Back |

`s` and the sections screen do not exist in the onboardee mode.

## Accessibility

Decided.

- Every action is reachable from the keyboard. There is no mouse use.
- The focus is always visible: a `>` mark, bold, and the accent colour.
- No information is carried by colour alone, see Colour.
- A short key hint line is always the last line. `?` opens the full list.
- Position is always stated in words: "4 of 11", and the count answered.
- The tool makes no motion. The text cursor does not blink, and nothing animates.
- Plain words on screen. The wording of a problem comes from the validator.

## Testing

Decided. The app is a Bubble Tea model, so tests send key messages and window-size messages and check the plan, the saved copies, and the view. At widths from 50 to 200 and heights from 10 to 50, a test checks that no line is wider than the window, that the small-window message appears under the limits, and that both panes appear at 140. A no-colour test checks the marks. A scripted pseudo-terminal run covers start and quit. It must answer the terminal's background-colour query and its cursor report, or the program waits five seconds.

## Open

- Splitting this page into one file per view under `docs/design/`, with design decision records (DDRs) for choices that cut across views, such as the width breakpoints. DDRs stay separate from ADRs: an ADR is for a decision that contracts or acceptance tests depend on, and a DDR is for a design choice that only the screens depend on.
- Whether the Walk goes to the next unanswered question when the person is revisiting from Review. Today it goes to the next in order.
- How the tool reminds a person of deferred questions after they quit.
- A mouse and a way to change the accent are not planned.
- The newcomer's view of the onboarder's answers, if the document preview proves not enough.
