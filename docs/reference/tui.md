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

Decided. There are four screens and a help page. Every screen has a header line with the tab strip, a progress rail, a body, a status line, and a key hint line, in that order. The hint line is always the last line.

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
- `w`, `o`, `r`, and `s` open the Walk, Overview, Review, and Sections from any screen while the person is not typing. The tab strip names them, see Screen tabs.
- `p`, left arrow, or shift+tab goes back one question. `n` or right arrow skips forward without changing anything.
- `d` defers the question and moves on. `x` clears the answer and stays, so the person can answer again.
- Escape on the Walk returns to the Overview or Review, whichever sent the person there. Escape on Overview or Review returns to the Walk.
- From Overview or Review, enter opens the chosen question in the Walk. `w` on Review returns to the Walk at the first unanswered question, and `w` on Overview or Sections returns to the question the person left.
- `q` quits from the Overview, Review, and Walk, except while typing. The Sections screen takes `q` as back. Ctrl+c always quits.

## Layout

Decided. A designer judged the layout against the people in Intent, and it replaces the first draft, which put a question list beside every question. Bubble Tea sends the window size at the start and on every resize, and the layout follows it. Nothing has a fixed width. The design target is a window 120 to 140 columns wide.

### What the layout does

- The Walk shows one question and nothing else that competes with it. The answer column holds the section title, the position ("4 of 11, required, not answered yet"), the question, the way to answer, and any problem.
- A progress rail sits under the header: one mark per question, grouped by section, with a diamond at the current question. The rail replaces the question list. It answers "where am I and how much is left" in one quiet line, and it never asks for a read. It is hidden when it does not fit.
- The notes sit beside the answer at 120 columns and wider, and below it when narrower. They hold the example, the type and options, who answers, the name, and the suggestions. They are faint, because they are for a glance and never for action.
- The Overview is one list, grouped by section, with each row showing its mark, its question, and its answer. At 120 columns and wider a faint preview of the chosen question sits beside it.
- The Review keeps two panes, because it is a master and detail task: the items on the left, the document preview on the right. Below 120 columns, tab switches between them.

| Width | Walk | Overview | Review |
| -- | -- | -- | -- |
| 120 or more | Answer column about three fifths of the width, notes beside it | List about three fifths, preview beside it | Items about two fifths, document beside it |
| 60 to 119 | One column up to 100 wide, notes below the answer if the height allows | One column | One column. Tab switches between the items and the document |
| Under 60 wide or under 12 high | A plain message asking for a larger window |  |  |

Rules that hold at any width:

- Widths come from the window. Text wraps at the column width, and a list scrolls by the real height around the chosen row.
- A line is never wider than the window. Rows that do not fit are cut with an ellipsis, and the key hint line drops its least useful hints before it drops `? help`.
- The answer column comes first in reading order and sits at the left, so the eye starts on the task.
- Leftover height stays empty. Breathing room is cheaper than a filled screen.

### Layouts considered

| Layout | For the first-time person (9 to 40 questions) | For the returning person | For the onboardee |
| -- | -- | -- | -- |
| A. Two panes, a question list on the left, the detail on the right (the first draft) | The bright list pulls the eye off the question, and every answer repaints a row beside the question | Good for jumping | A list of questions they will never answer reads like a form |
| B. One focused column, a progress rail on top, notes beside or below (chosen) | One thing to read and one place to type. The rail gives position without a read | The Overview is the list, one key away, and the Review lists what is waiting | A short rail and a welcome line. Nothing else on screen |
| C. One column only, a plain "4 of 11" and no rail or notes | Calmest, but there is no sense of the whole and no place for the example | Needs the Overview for everything | Fine, but bare |
| D. The answer wide on the left, and the document preview always beside it | The document changes under every answer and draws the eye | Useful for checking | A document full of gaps is discouraging at the start |

### Rejected

- The question list beside the Walk (A). The product owner found it distracting, and the reasons hold: it is the brightest thing on screen and the person is not acting in it. The rail keeps what the list gave, a sense of place, at a tenth of the weight.
- The document preview beside the Walk (D). It is the right tool for the Review, where the person judges the whole, and noise on the Walk, where the person answers one thing.
- No rail and no notes (C). Fewer parts, but a person answering 40 questions loses their place, and the example needs a home.
- A layout fixed at 80 columns. Modern terminals are wider. The 120 column break is where the notes fit beside a comfortable answer column. Below it the notes move under the answer, and the tool is still usable at 80.
- A vertical rule down the whole screen. The rule stops where the content stops.

Direction: the onboardee sees the same screens with fewer parts: no section switches, a welcome line while nothing is answered, and "3 of 5 to answer" in the header.

## Emphasis

Decided. The part of the screen where the person is acting is drawn bright, and every other part is drawn faint. Emphasis follows focus.

| Screen | Bright | Faint |
| -- | -- | -- |
| Walk | The question, the answer, the problems | The rail, the notes, the status |
| Overview | The list | The rail, the preview |
| Review | The pane with focus, items or document | The other pane |
| Sections, Help | The whole body |  |

- The accent marks "you are here" and nothing else: the `>` before the chosen row or the focused pane, the diamond in the rail, the prompt of the answer box, and the current tab. Titles, headings, and key names are never in the accent. A title is bold, and a key name is bold.
- Bold and faint are enough to carry the hierarchy, so it holds on a light or dark background and with no colour.

## Wireframes

Decided. No wireframe tool fits a terminal screen, so the screens are drawn by the code itself. A test renders the real view from a fixed plan with no colour and compares it to a text file. The files are the wireframes, and they cannot drift from the screen. Regenerate them on purpose with `go test ./internal/tui -run Wireframes -update`, and review the diff.

| Screen | 80 columns | 120 columns | 140 columns |
| -- | -- | -- | -- |
| Walk, choosing | [walk](../../internal/tui/testdata/screens/walk.80.txt) | [walk](../../internal/tui/testdata/screens/walk.120.txt) | [walk](../../internal/tui/testdata/screens/walk.140.txt) |
| Walk, typing | [edit](../../internal/tui/testdata/screens/edit.80.txt) | [edit](../../internal/tui/testdata/screens/edit.120.txt) | [edit](../../internal/tui/testdata/screens/edit.140.txt) |
| Walk, with an example | [example](../../internal/tui/testdata/screens/example.80.txt) | [example](../../internal/tui/testdata/screens/example.120.txt) | [example](../../internal/tui/testdata/screens/example.140.txt) |
| Walk, picking one | [pick](../../internal/tui/testdata/screens/pick.80.txt) | [pick](../../internal/tui/testdata/screens/pick.120.txt) | [pick](../../internal/tui/testdata/screens/pick.140.txt) |
| Walk, picking several | [pick many](../../internal/tui/testdata/screens/pick-many.80.txt) | [pick many](../../internal/tui/testdata/screens/pick-many.120.txt) | [pick many](../../internal/tui/testdata/screens/pick-many.140.txt) |
| Walk, nothing to pick | [pick empty](../../internal/tui/testdata/screens/pick-empty.80.txt) | [pick empty](../../internal/tui/testdata/screens/pick-empty.120.txt) | [pick empty](../../internal/tui/testdata/screens/pick-empty.140.txt) |
| Overview | [overview](../../internal/tui/testdata/screens/overview.80.txt) | [overview](../../internal/tui/testdata/screens/overview.120.txt) | [overview](../../internal/tui/testdata/screens/overview.140.txt) |
| Overview with sub-sections | [subsections](../../internal/tui/testdata/screens/subsections.80.txt) | [subsections](../../internal/tui/testdata/screens/subsections.120.txt) | [subsections](../../internal/tui/testdata/screens/subsections.140.txt) |
| Review | [review](../../internal/tui/testdata/screens/review.80.txt) | [review](../../internal/tui/testdata/screens/review.120.txt) | [review](../../internal/tui/testdata/screens/review.140.txt) |
| Sections | [sections](../../internal/tui/testdata/screens/sections.80.txt) | [sections](../../internal/tui/testdata/screens/sections.120.txt) | [sections](../../internal/tui/testdata/screens/sections.140.txt) |
| The onboardee's walk | [onboardee](../../internal/tui/testdata/screens/onboardee.80.txt) | [onboardee](../../internal/tui/testdata/screens/onboardee.120.txt) | [onboardee](../../internal/tui/testdata/screens/onboardee.140.txt) |
| The onboardee, nothing to pick | [onboardee empty](../../internal/tui/testdata/screens/pick-empty-onboardee.80.txt) | [onboardee empty](../../internal/tui/testdata/screens/pick-empty-onboardee.120.txt) | [onboardee empty](../../internal/tui/testdata/screens/pick-empty-onboardee.140.txt) |
| Window too small | [too small](../../internal/tui/testdata/screens/too-small.80.txt) | [too small](../../internal/tui/testdata/screens/too-small.120.txt) | [too small](../../internal/tui/testdata/screens/too-small.140.txt) |

## Colour

Decided. One accent, the Branded Blue `0071CE`, marks "you are here", see Emphasis. Two other brand colours mean status and nothing else: Orange `FF8300` for a warning, Red `F32735` for a problem. Everything quiet is drawn with the terminal's faint style, so it works on a light or a dark background and the tool never asks the terminal for its background colour. The styles are built when a screen is drawn, never at package level.

- A terminal that reports true colour gets the hex values. One with 256 colours or 16 colours gets the nearest ANSI value. `NO_COLOR` and a terminal with no colour get plain text.
- Colour never carries meaning alone. The chosen row starts with `>`, and the current tab is in brackets. States are words and marks: `[x]` answered, `[~]` deferred, `[ ]` not answered, `*` required, and in the rail `x`, `~`, `.`, and a diamond for the current question. The position and state are also words under the section title. Problems start with `! problem:` and warnings with `warning:`. The focused pane of the Review shows `>` before its title.
- Tests render every screen with no colour and check that these marks and the key hints are still there. Other tests check, with colour on, that the focused pane is bright, the other is faint, and the accent is not on a title.

## The example

Decided. A question may carry one `Example`, an illustration of a good answer. It is shown as faint placeholder text in the answer box, after "e.g.", and as an "Example:" line in the notes. It is never an answer and never a default: enter on an empty box does not save it, and it disappears when the person types.

## Answering

Decided, and built. The answer is made on the Walk, under the question.

- A boolean or choice is picked from a list, with "deferred, answer later" and "clear the answer" at the end. The cursor starts on the current answer. Up and down move, enter picks.
- A contact, role, or team is picked from the organization's directory, see Pick lists. A person never types an id.
- A string takes typed text, with the current answer filled in. Tab cycles through the question's suggestions. Enter on text equal to the current answer moves on without saving. Enter on an empty box clears the answer. Ctrl+d defers.
- A command is words separated by spaces, with quotes keeping spaces together.
- Several people of the `person` type are separated by `;`.
- An answer that the validator refuses stays on the Walk with the problem message under the answer.
- Every accepted answer is saved to the plan file at once. There is no unsaved state.
- While a person types, letters are text, so the single-letter keys are off. Back, forward, defer, and help have control-key forms listed below.

## Pick lists

Decided, and built. A `contact`, `role`, or `team` question carries the directory's entries as `Question.Choices`, each with an id and a name, in directory order, see [0080](../decisions/0080-directory.md). The Walk shows the names and stores the id. The id stays out of sight.

```text
  Which role owns onboarding here? It appears as the owner of the Access
  section.

  > Engineering manager
    IT contact
    Platform team lead
    deferred, answer later
    clear the answer
```

- **One answer.** A single question looks and works like a choice question: up and down move, enter picks, and the cursor starts on the current answer, marked `(current)`. The list ends with "deferred, answer later" and "clear the answer".
- **Several answers.** A `multiple` question is a checklist. Space picks or unpicks the row under the cursor, and each picked row shows the number it was picked in. Enter confirms the picks as one ordered list. The cursor starts on the first current pick.

```text
  > [2] Alex Rivera
    [1] Sam Okafor
    [ ] IT service desk

  Space picks or unpicks. Enter confirms. Pick order is kept.
  Order: Sam Okafor, then Alex Rivera
```

- **Order.** The order of the picks is the order stored, because an escalation list is read in order. The numbers and the "Order:" line show it. There is no key to move a pick, and that was rejected: unpick and pick again costs two keys for a list of two or three, and a move key is one more key to learn and to show. Revisit it if lists grow.
- **Enter on no picks.** With nothing picked and nothing answered, enter moves on without saving, as on an empty text box. With nothing picked and an answer saved, enter clears it.
- **Same names.** Two entries with the same name show their ids in brackets, `Sam Okafor (sam-1)` and `Sam Okafor (sam-2)`, and only those two. A name that is unique shows no id.
- **Defer and move on.** `d` defers, `x` clears, `n` and `p` move, as for every other question. A several-answers list has no rows for defer and clear, because enter confirms the picks.
- **Long lists.** The list scrolls around the cursor to fit the window. A cut list adds a faint "Showing 12 to 19 of 40" line. The help lines under a several-answers list go first when there is no room.
- **Nothing to pick.** When the directory has no entry of the kind, a message takes the list's place and every other key still works, so a person can defer and come back. The onboarder reads "No contacts to pick from yet. Add them to the organization's directory first." with "roles" or "teams" for those types. The onboardee reads "No teams to pick from yet. Ask your onboarder to add this." The People screen will replace the first wording with a pointer to it.
- **Names everywhere.** The Overview, its preview, and the suggestion lines in the notes show the name, and several names are separated by `;` in the order stored. The Review lists blockers, deferred questions, and validator problems, which are about questions and not about answers, so it needs no name. The document preview shows the names the assembler resolves.
- **Stale answers.** An answer that names an id the directory no longer holds counts as not answered, as the finder counts it. The Walk, Overview, rail, and counts show it as not answered, and no id is shown. A faint "was sam-okafor" was rejected: the validator already warns about the stale id on the Review, and the person's task is to pick again. A list with at least one known id counts as answered and shows only the known names. Confirming the picks drops the unknown ids.

## Keys

Decided. The key hint line shows the most useful keys for the screen and the width, and `?` shows all of them.

| Key | Walk, choosing | Walk, typing | Overview | Review | Sections |
| -- | -- | -- | -- | -- | -- |
| enter | Pick, move on. On a several-answers list, confirm the picks | Accept, move on | Open the question | Open the question | Switch on or off |
| up, down, `k`, `j` | Move the choice |  | Move | Move | Move |
| space | Pick or unpick on a several-answers list |  |  |  | Switch on or off |
| `p`, left, shift+tab | Back | shift+tab, ctrl+p |  |  |  |
| `n`, right | Skip forward | ctrl+n |  |  |  |
| `d` | Defer, move on | ctrl+d |  |  |  |
| `x` | Clear | Empty the box and enter |  |  |  |
| tab |  | Try a suggestion |  | Items or document |  |
| `w`, `o`, `r`, `s` | `o`, `r`, `s` | Letters are text | `w`, `r`, `s` | `w`, `o`, `s` | `w`, `o`, `r` |
| pgup, pgdn, `g`, `G` |  |  | `g`, `G` | Scroll the document |  |
| esc | Back to overview or review | Same | Walk | Walk | Back |
| `?` | Help | f1 | Help | Help | Help |
| `q` | Quit |  | Quit | Quit | Back |

`s` and the sections screen do not exist in the onboardee mode. The key hint line leaves out the screen keys while the tab strip shows them, see Screen tabs.

## Accessibility

Decided.

- Every action is reachable from the keyboard. There is no mouse use.
- The focus is always visible: a `>` mark and bold text, and the rest of the screen is faint.
- No information is carried by colour alone, see Colour.
- A short key hint line is always the last line. `?` opens the full list.
- Position is always stated in words: "4 of 11", and the count answered.
- A window under 60 columns or 12 rows gets a plain message in place of the screen.
- The tool makes no motion. The text cursor does not blink, and nothing animates.
- Plain words on screen. The wording of a problem comes from the validator.

## Testing

Decided. The app is a Bubble Tea model, so tests send key messages and window-size messages and check the plan, the saved copies, and the view. At widths from 60 to 200 and heights from 12 to 50, a test checks that no line is wider than the window, including questions with very long descriptions, examples, and suggestions. Another checks the small-window message under the limits, and that the Walk has two columns at 120 and one below. The flow tests cover the guided order, the position text, back, defer and move on, finishing on the Review, jumping from the Review, and the onboardee mode. A fake Backend with hand-built questions tests the example, the pick lists, and the sub-sections. The pick tests cover one answer, several answers in order, an empty directory for each audience, same-named entries, and a stale answer. Tab tests check the strip with and without colour, and the keys it drops at narrow widths. Style tests check the bright and faint panes in each state. A scripted pseudo-terminal run covers start and quit. It must answer the terminal's background-colour query and its cursor report, or the program waits five seconds.

## Screen tabs

Decided, and built. A designer asked for it after using the redesigned Walk: a bare title does not show which screen a person is on.

The header line holds the title, the tab strip, and the progress count at the right.

```text
  Onboarding plan  Walk w  [Overview]  Review r  Sections s          13 of 13 answered
```

- The strip names the screens a person can move between, in a fixed order: Walk, Overview, Review, Sections. A screen the audience cannot use, such as Sections for the onboardee, is left out.
- The current tab is bold in the accent colour, which is "you are here". With no colour it is in brackets. The other tabs are faint, each with its key.
- The strip is the one place the screen keys are shown, so the key hint line leaves them out. The Walk while choosing, the Overview, the Review, and the Sections screen all take `w`, `o`, `r`, and `s`, except while the person is typing.
- While a person types, letters are text, so the tabs show their names with no keys, and the hint line says `esc` leaves.
- When the header is too wide, it drops the progress count first, then the keys. With the keys gone, the hint line shows them again, so a key is always on screen somewhere. At 60 columns the strip shows names only.

Rejected: the screen name beside the title, as before (`Onboarding plan  Overview`), because it shows one screen and hides that there are others. Also rejected: a tab for Help, because help is a page over the current screen and the current tab stays marked while it is open.

## Sub-sections

Decided, and built. Under [0070](../decisions/0070-sub-sections.md) a section can have a parent.

- The Overview shows a sub-section under its parent, indented two columns for each level of depth. The indent applies to the section title and to its question rows.
- The progress rail keeps the cells of a section and its sub-sections together, with the usual gap only between families of sections.
- The Walk shows the sub-section's own title. The document preview nests it by heading level.

## Open

- Splitting this page into one file per view under `docs/design/`, with design decision records (DDRs) for choices that cut across views, such as the width breakpoints. DDRs stay separate from ADRs: an ADR is for a decision that contracts or acceptance tests depend on, and a DDR is for a design choice that only the screens depend on.
- A key to move a pick up or down in a several-answers list.
- Adding a contact, role, or team from the pick list. Under 0080 the onboarder adds one inline, and the People screen is where that arrives.
- Whether the Walk goes to the next unanswered question when the person is revisiting from Review. Today it goes to the next in order.
- How the tool reminds a person of deferred questions after they quit.
- Whether the rail should show the section titles when the window is wide and the plan is small.
- Whether the notes should hide while a person is typing, to leave only the answer.
- A mouse and a way to change the accent are not planned.
- The newcomer's view of the onboarder's answers, if the document preview proves not enough.
