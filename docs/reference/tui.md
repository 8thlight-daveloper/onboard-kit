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

Decided. There are five screens and a help page. Every screen has a header line with the tab strip, a progress rail, a body, a status line, and a key hint line, in that order. The hint line is always the last line.

| Screen | Holds | Opened by |
| -- | -- | -- |
| Walk | One question: its position, its section and group, what it asks, whether it is required, its current answer, the way to answer, suggestions, problems | The start, enter on an overview or review row, finishing a question |
| Overview | Every question in order with a mark for its state. The cursor jumps to any of them | `o` |
| Review | Blockers, deferred questions, validator problems, the document preview, and a quiet line when answers are left over, see Left over answers | `r`, finishing the last question, or the start when the plan holds answers |
| Sections | Section switches. The onboarder only | `s` |
| Contacts | The organization's contacts, roles, and teams, to add, change, and remove. The onboarder only | `c` |
| Help | Every key of the current screen | `?`, or `f1` while typing |

How a person moves:

- Start: if the plan holds an answer to any question of this audience, open Review. Otherwise open Walk at the first question. A plan with no active questions opens Review.
- Enter accepts the answer on the screen and moves to the next question in order. This was a choice between the next question and the next unanswered one. The next in order is predictable, and a person who wants an unanswered one opens Review, where the blockers are listed.
- Past the last question, the person lands on Review.
- `w`, `o`, `r`, `s`, and `c` open the Walk, Overview, Review, Sections, and Contacts from any screen while the person is not typing. The tab strip names them, see Screen tabs.
- `p`, left arrow, or shift+tab goes back one question. `n` or right arrow skips forward without changing anything.
- `d` defers the question and moves on. `x` clears the answer and stays, so the person can answer again.
- Escape on the Walk returns to the Overview or Review, whichever sent the person there. Escape on Overview or Review returns to the Walk.
- From Overview or Review, enter opens the chosen question in the Walk. `w` on Review returns to the Walk at the first unanswered question, and `w` on Overview or Sections returns to the question the person left.
- `q` quits from the Overview, Review, and Walk, except while typing. The Sections and Contacts screens take `q` as back, and escape goes back too, to the screen the person came from. Ctrl+c always quits.

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
| Contacts, list | The list and the list names at the top | The detail of the chosen entry |
| Contacts, form and removal | The whole body |  |

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
| Review, with answers left over | [review-left-over](../../internal/tui/testdata/screens/review-left-over.80.txt) | [review-left-over](../../internal/tui/testdata/screens/review-left-over.120.txt) | [review-left-over](../../internal/tui/testdata/screens/review-left-over.140.txt) |
| Review, removing left over answers | [review-prune](../../internal/tui/testdata/screens/review-prune.80.txt) | [review-prune](../../internal/tui/testdata/screens/review-prune.120.txt) | [review-prune](../../internal/tui/testdata/screens/review-prune.140.txt) |
| Sections | [sections](../../internal/tui/testdata/screens/sections.80.txt) | [sections](../../internal/tui/testdata/screens/sections.120.txt) | [sections](../../internal/tui/testdata/screens/sections.140.txt) |
| The onboardee's walk | [onboardee](../../internal/tui/testdata/screens/onboardee.80.txt) | [onboardee](../../internal/tui/testdata/screens/onboardee.120.txt) | [onboardee](../../internal/tui/testdata/screens/onboardee.140.txt) |
| The onboardee, nothing to pick | [onboardee empty](../../internal/tui/testdata/screens/pick-empty-onboardee.80.txt) | [onboardee empty](../../internal/tui/testdata/screens/pick-empty-onboardee.120.txt) | [onboardee empty](../../internal/tui/testdata/screens/pick-empty-onboardee.140.txt) |
| Contacts, contacts | [list](../../internal/tui/testdata/screens/contacts-list.80.txt) | [list](../../internal/tui/testdata/screens/contacts-list.120.txt) | [list](../../internal/tui/testdata/screens/contacts-list.140.txt) |
| Contacts, roles | [roles](../../internal/tui/testdata/screens/contacts-roles.80.txt) | [roles](../../internal/tui/testdata/screens/contacts-roles.120.txt) | [roles](../../internal/tui/testdata/screens/contacts-roles.140.txt) |
| Contacts, teams | [teams](../../internal/tui/testdata/screens/contacts-teams.80.txt) | [teams](../../internal/tui/testdata/screens/contacts-teams.120.txt) | [teams](../../internal/tui/testdata/screens/contacts-teams.140.txt) |
| Contacts, nothing yet | [empty](../../internal/tui/testdata/screens/contacts-empty.80.txt) | [empty](../../internal/tui/testdata/screens/contacts-empty.120.txt) | [empty](../../internal/tui/testdata/screens/contacts-empty.140.txt) |
| Contacts, new contact | [form](../../internal/tui/testdata/screens/contacts-form.80.txt) | [form](../../internal/tui/testdata/screens/contacts-form.120.txt) | [form](../../internal/tui/testdata/screens/contacts-form.140.txt) |
| Contacts, picking holders of a role | [holders](../../internal/tui/testdata/screens/contacts-holders.80.txt) | [holders](../../internal/tui/testdata/screens/contacts-holders.120.txt) | [holders](../../internal/tui/testdata/screens/contacts-holders.140.txt) |
| Contacts, a team's holders for one role | [override](../../internal/tui/testdata/screens/contacts-override.80.txt) | [override](../../internal/tui/testdata/screens/contacts-override.120.txt) | [override](../../internal/tui/testdata/screens/contacts-override.140.txt) |
| Contacts, removing a contact | [remove](../../internal/tui/testdata/screens/contacts-remove.80.txt) | [remove](../../internal/tui/testdata/screens/contacts-remove.120.txt) | [remove](../../internal/tui/testdata/screens/contacts-remove.140.txt) |
| Contacts, removing a role | [remove role](../../internal/tui/testdata/screens/contacts-remove-role.80.txt) | [remove role](../../internal/tui/testdata/screens/contacts-remove-role.120.txt) | [remove role](../../internal/tui/testdata/screens/contacts-remove-role.140.txt) |
| Contacts, editing a team | [team-form](../../internal/tui/testdata/screens/contacts-team-form.80.txt) | [team-form](../../internal/tui/testdata/screens/contacts-team-form.120.txt) | [team-form](../../internal/tui/testdata/screens/contacts-team-form.140.txt) |
| Contacts, picking a team's contact | [team-contact](../../internal/tui/testdata/screens/contacts-team-contact.80.txt) | [team-contact](../../internal/tui/testdata/screens/contacts-team-contact.120.txt) | [team-contact](../../internal/tui/testdata/screens/contacts-team-contact.140.txt) |
| Contacts, removing a team's contact | [remove-team-contact](../../internal/tui/testdata/screens/contacts-remove-team-contact.80.txt) | [remove-team-contact](../../internal/tui/testdata/screens/contacts-remove-team-contact.120.txt) | [remove-team-contact](../../internal/tui/testdata/screens/contacts-remove-team-contact.140.txt) |
| Window too small | [too small](../../internal/tui/testdata/screens/too-small.80.txt) | [too small](../../internal/tui/testdata/screens/too-small.120.txt) | [too small](../../internal/tui/testdata/screens/too-small.140.txt) |

## Colour

Decided, see [0560](../decisions/0560-use-the-terminal-colors.md). The screens use the terminal's own colours and none of ours: the default foreground and background, and the sixteen colours the person's theme defines. One accent, the terminal's green in bold, marks "you are here", see Emphasis. Two other colours mean status and nothing else: the terminal's yellow for a warning and its red for a problem. Everything quiet is drawn with the terminal's faint style, so it works on a light or a dark background and the tool never asks the terminal for its background colour. The styles are built when a screen is drawn, never at package level.

- No screen uses a hexadecimal, 256-colour, or 24-bit value. `NO_COLOR` and a terminal with no colour get plain text.
- Colour never carries meaning alone. The chosen row starts with `>`, and the current tab is in brackets. States are words and marks: `[x]` answered, `[~]` deferred, `[ ]` not answered, `*` required, and in the rail `x`, `~`, `.`, and a diamond for the current question. The position and state are also words under the section title. Problems start with `! problem:` and warnings with `warning:`. The focused pane of the Review shows `>` before its title.
- Tests render every screen with no colour and check that these marks and the key hints are still there. Other tests check, with colour on, that the focused pane is bright, the other is faint, the accent is not on a title, and no screen uses a fixed colour value.

## The example

Decided. A question may carry one `Example`, an illustration of a good answer. It is shown as faint placeholder text in the answer box, after "e.g.", and as an "Example:" line in the notes. It is never an answer and never a default: enter on an empty box does not save it, and it disappears when the person types.

## Answering

Decided, and built. The answer is made on the Walk, under the question.

- A boolean or choice is picked from a list, with "deferred, answer later" and "clear the answer" at the end. The cursor starts on the current answer. Up and down move, enter picks.
- A contact, role, or team is picked from the organization's directory, see Pick lists. A person never types an id.
- A string takes typed text, with the current answer filled in. Tab cycles through the question's suggestions. Enter on text equal to the current answer moves on without saving. Enter on an empty box clears the answer. Ctrl+d defers.
- A command is words separated by spaces, with quotes keeping spaces together.
- Several contacts or several roles in a `multiple` question are picked as a checklist, see Pick lists.
- An answer that the validator refuses stays on the Walk with the problem message under the answer.
- Every accepted answer is saved to the plan file at once. There is no unsaved state.
- While a person types, letters are text, so the single-letter keys are off. Back, forward, defer, and help have control-key forms listed below.

## Pick lists

Decided, and built. A `contact`, `role`, or `team` question carries the directory's entries as `Question.Choices`, each with an id and a name, in directory order, see [0080](../decisions/0080-directory.md). The Walk shows the names and stores the id. The id stays out of sight. The Contacts screen is where entries are made, see Contacts.

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
- **Nothing to pick.** When the directory has no entry of the kind, a message takes the list's place and every other key still works, so a person can defer and come back. The onboarder reads "No contacts to pick from yet. Add them on the Contacts screen, press c." with "roles" or "teams" for those types, and `c` opens the Contacts screen on the list the question needs. The onboardee reads "No teams to pick from yet. Ask your onboarder to add this."
- **Names everywhere.** The Overview, its preview, and the suggestion lines in the notes show the name, and several names are separated by `;` in the order stored. The Review lists blockers, deferred questions, and validator problems, which are about questions and not about answers, so it needs no name. The document preview shows the names the assembler resolves.
- **Stale answers.** An answer that names an id the directory no longer holds counts as not answered, as the finder counts it. The Walk, Overview, rail, and counts show it as not answered, and no id is shown. A faint "was sam-okafor" was rejected: the validator already warns about the stale id on the Review, and the person's task is to pick again. A list with at least one known id counts as answered and shows only the known names. Confirming the picks drops the unknown ids.

## Left over answers

Decided, and built. For both audiences, under [0075](../decisions/0075-refreshing-a-personal-plan.md): a refresh never removes an answer whose question the plan no longer asks. The owner removes it, and the editor's `PruneHiddenAnswers` does the work. The onboarder removes them from the organization's plan they walk, and a newcomer from their own plan.

A left over answer is data, not something the person gave on purpose today. The words say so: "left over from questions that are no longer asked". They never say "your answers", because the person did not choose to keep them and may never have seen them.

```text
  > Items
  Answered 2 of 12.
  3 answers are left over from questions that are no longer asked. Press h to
  review them.
```

- **Where.** On the Review, under "Answered 2 of 12", in faint text. It is there only when something is left over, so a plan with none shows no new line. The Review is where a person checks the plan before handing it on, and the line sits next to the validator's problems about the same answers. Each of those problems, "No variable declares this question", stays as it is.
- **Key.** `h`. It is free on the Review and is not typing, so it cannot collide with text. `x` was rejected because it clears an answer on the Walk, and `p` because it is back. `r` is the Review itself. The hint line shows `h left over` while something is left over. With nothing left over, `h` says "No answers are left over from removed questions." and changes nothing.
- **Confirmation.** `h` runs `PruneHiddenAnswers` on the current plan as a dry run and shows the result in place of the Review panes, as the Contacts removal does. Each answer shows its id, such as `plan.old_tool`, because the plan no longer declares the question and has no text to show for it, and its value in a short form below. A list of more than five is cut with "and 5 more". `y` applies it. Every other key keeps the answers and says "nothing removed". `?` opens help over it. Enter does not confirm, for the reason given under Removing an entry.

```text
  Remove 3 left over answers?

  These belong to questions the plan no longer asks, so they are not part of
  the current plan. They stay in the file until you remove them.

    - plan.old_token
        value: "sk-l…" (30 characters)
    - plan.old_tool
        value: "Trello"
    - retired.flag
        value: yes

  Nothing is removed until you press y.
```

- **Values.** A value can be long or secret-like, so a value of up to 20 characters shows whole, in quotes, and a longer one shows its first 4 characters and its length: `"sk-l…" (30 characters)`. A list shows its first 4 characters and the number of items when it is long. A flag shows yes or no, and a deferred answer shows "deferred". Rejected: showing every value whole, because a stale token would then be on screen, and hiding every value, because a person cannot judge "plan.old_tool" without a hint of what it held. A short secret can still show whole. The tool cannot tell a short secret from a short word.
- **After `y`.** The plan the dry run returned replaces the plan, and it is saved at once, as every answer is. The status says "removed 3 left over answers. saved", or "could not save" with the reason, and the plan stays changed in memory as for an answer. The Review is rebuilt, so the line and the validator's problems go, and the cursor and focus stay where they were.
- **Rejected.** A prune on start with no question asked, because the decision says the owner prunes. A Contacts-style tab for them, because there are few and a person visits once. Listing each answer with its own key to remove it, because the reason for removing one is the same as for the rest.

## Keys

Decided. The key hint line shows the most useful keys for the screen and the width, and `?` shows all of them.

| Key | Walk, choosing | Walk, typing | Overview | Review | Sections |
| -- | -- | -- | -- | -- | -- |
| enter | Pick, move on. On a several-answers list, confirm the picks | Accept, move on | Open the question | Open the question, or the Contacts entry of a directory problem | Switch on or off |
| up, down, `k`, `j` | Move the choice |  | Move | Move | Move |
| space | Pick or unpick on a several-answers list |  |  |  | Switch on or off |
| `p`, left, shift+tab | Back | shift+tab, ctrl+p |  |  |  |
| `n`, right | Skip forward | ctrl+n |  |  |  |
| `d` | Defer, move on | ctrl+d |  |  |  |
| `x` | Clear | Empty the box and enter |  |  |  |
| tab |  | Try a suggestion |  | Items or document |  |
| `w`, `o`, `r`, `s`, `P` | `o`, `r`, `s`, `P` | Letters are text | `w`, `r`, `s`, `P` | `w`, `o`, `s`, `P` | `w`, `o`, `r`, `P` |
| pgup, pgdn, `g`, `G` |  |  | `g`, `G` | Scroll the document |  |
| esc | Back to overview or review | Same | Walk | Walk | Back |
| `?` | Help | f1 | Help | Help | Help |
| `q` | Quit |  | Quit | Quit | Back |

On the Review, `h` shows the answers left over from questions the plan no longer asks, see Left over answers. It is free on every other screen and is not a screen key, so the tab strip does not list it. The key hint line shows `h left over` only while something is left over.

`s`, `c`, the Sections screen, and the Contacts screen do not exist in the onboardee mode. The key hint line leaves out the screen keys while the tab strip shows them, see Screen tabs.

## The status line

Decided. The status line stays the row above the key hint line on every screen. It reports the result of the last action: a save, a refusal, a notice. The next key clears it.

A person who presses a key on an item at the top of the screen is looking at the item, not at the bottom row. A message there can go unseen. This was found when enter on a validator problem of the directory did nothing visible. The choice was between moving the status line, making it louder, and removing the cases that need it.

- **Remove the dead end first.** An action that cannot work should not be offered, so the status line is not the only place the person learns it. This is the chosen move.
- **Say where the item leads.** A validator problem of the directory names its entry, such as "Onboarding owner: This role has no holder.", and enter opens that entry on the Contacts screen. Escape returns to the Review. In the onboardee mode there is no Contacts screen, so enter says so in the status line. A problem that names a question still opens the question.
- **Direction, not built.** The key hint line shows `enter open` on the Review only when the chosen item opens something. A problem that names a question the plan no longer asks opens nothing.
- **Direction, not built.** A warning in the status line starts with an exclamation mark and is bold, so it is not read as a quiet note. Colour stays a second signal, see Colour.

### Rejected for the status line

- Moving it to the top of the screen. It only changes the distance, and it shifts the layout of every screen and every wireframe.
- A banner or a modal. A banner shifts the layout, and a modal asks for a decision the message does not need.
- A timed toast. The tool makes no motion and keeps no timers, see Accessibility, and the next key already clears the line.
- A faint second line under the chosen item for the reason. It is the most direct signal but needs wrapping and height rules in the list. Revisit it if the two directions above prove too quiet.

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

Decided. The app is a Bubble Tea model, so tests send key messages and window-size messages and check the plan, the saved copies, and the view. At widths from 60 to 200 and heights from 12 to 50, a test checks that no line is wider than the window, including questions with very long descriptions, examples, and suggestions. Another checks the small-window message under the limits, and that the Walk has two columns at 120 and one below. The flow tests cover the guided order, the position text, back, defer and move on, finishing on the Review, jumping from the Review, and the onboardee mode. A fake Backend with hand-built questions tests the example, the pick lists, and the sub-sections. The pick tests cover one answer, several answers in order, an empty directory for each audience, same-named entries, and a stale answer. Tab tests check the strip with and without colour, and the keys it drops at narrow widths. The Contacts tests drive the screen through the real editor: adding, changing, and removing each kind, the id made from a name, a refusal that leaves the plan as it was, the dry run before a removal, the flags for vacant roles and unreachable contacts, the autosave of every accepted edit, and the walk's pick list after returning. Style tests check the bright and faint panes in each state. The prune tests drive the Review through the real editor for both audiences: the notice only when something is left over, the dry run that changes and saves nothing, every key but `y` keeping the answers, `y` saving once and clearing the notice, the short forms of values with no long value on screen, the cut list, and `h` as text while typing. A fake Backend covers the update logic, and the width tests include the notice and the confirmation. A scripted pseudo-terminal run covers start and quit. It must answer the terminal's background-colour query and its cursor report, or the program waits five seconds.

## Screen tabs

Decided, and built. A designer asked for it after using the redesigned Walk: a bare title does not show which screen a person is on.

The header line holds the title, the tab strip, and the progress count at the right.

```text
  Onboarding plan  Walk w  [Overview]  Review r  Sections s          13 of 13 answered
```

- The strip names the screens a person can move between, in a fixed order: Walk, Overview, Review, Sections, Contacts. A screen the audience cannot use, such as Sections and Contacts for the onboardee, is left out. The Contacts key is lowercase `c`. It was the capital `P` when the screen was named People. `p` is back on the Walk, so `P` was a special case that only worked with shift. `b` is page-up on the scrolling screens. `c` is free on every screen and matches the screen name.
- The current tab is bold in the accent colour, which is "you are here". With no colour it is in brackets. The other tabs are faint, each with its key.
- The strip is the one place the screen keys are shown, so the key hint line leaves them out. The Walk while choosing, the Overview, the Review, and the Sections screen all take `w`, `o`, `r`, and `s`, except while the person is typing.
- While a person types, letters are text, so the tabs show their names with no keys, and the hint line says `esc` leaves.
- When the header is too wide, it drops the progress count first, then the keys. With the keys gone, the hint line shows them again, so a key is always on screen somewhere. The onboarder's strip with five tabs and keys is 70 columns wide, so at 80 columns the progress count goes and the keys stay. At 60 columns the strip shows names only, and its gaps shrink to one column.

Rejected: the screen name beside the title, as before (`Onboarding plan  Overview`), because it shows one screen and hides that there are others. Also rejected: a tab for Help, because help is a page over the current screen and the current tab stays marked while it is open.

## Sub-sections

Decided, and built. Under [0070](../decisions/0070-sub-sections.md) a section can have a parent.

- The Overview shows a sub-section under its parent, indented two columns for each level of depth. The indent applies to the section title and to its question rows.
- The progress rail keeps the cells of a section and its sub-sections together, with the usual gap only between families of sections.
- The Walk shows the sub-section's own title. The document preview nests it by heading level.

## Contacts

Decided, and built. The onboarder's screen for the directory in [0080](../decisions/0080-directory.md). The onboardee has no tab and no key for it. It does the same work as the editor's directory operations, shows each one's result, and leaves every rule to the editor and the validator.

### What it shows

The three lists are three sub-tabs under the header, with a count each: `Contacts 5  Roles 4  Teams 2`. The current list is bold in the accent, or in brackets with no colour, as the screen tabs are. Tab, shift+tab, and the side arrows move between them.

```text
  [Contacts 5]  Roles 4  Teams 2                                        │ About this contact
                                                                        │
  > Alex Rivera      alex.rivera@acme.example  @alex                    │ Id: alex-rivera
    Sam Okafor       sam.okafor@acme.example   @sam                     │ Email: alex.rivera@acme.example
    IT service desk  help@acme.example         #it-help                 │ Slack: @alex
    Pat Lee          ! email looks wrong  pat.lee.acme.example          │ Holds: Platform team lead
    Front desk       ! no way to reach                                  │
```

- **Contacts** show the name, the email, and the Slack handle. **Roles** show the name and the holders' names in order, or `! vacant`. **Teams** show the name and the number of roles they override. The detail of a team starts with its contact by name and Slack handle, or email when it has no handle: "Contact: Platform team channel, #platform-team". A team with no contact reads "Contact: none".
- **Problems.** The validator's directory problems sit next to the entry in words: `! vacant`, `! no way to reach`, `! email looks wrong`, `! holder is not a contact`, `! team contact is not a contact`, `! role does not exist`, and `! id is shared`. The detail of a team contact problem says "The contact for this team is not one of the contacts. Pick another, or choose no team contact." A flagged row shows the flag before its other details, so a narrow window cuts the details and keeps the flag. The detail says what it means: "warning: Nobody holds this role, so it shows as a gap in the document." Colour adds orange, and the words and the `!` carry the meaning without it. A problem that is an error reads `! problem:`.
- **Layout.** At 120 columns and wider the list is the left three fifths and the detail of the chosen entry is faint at the right. Below 120 the list is one column and the detail sits under it, cut to the height that is left. The detail holds every field, the holders in order, and for a team the roles whose holders differ.
- **Nothing yet.** A list with no entry says "No contacts yet." and "Press a to add one."

### Why three sub-tabs

One list with three headers would show everything at once, but the add key would have to guess which kind the person means, and the cursor would cross kinds. Three groups on one screen need a focus between them, which is the same as sub-tabs with more keys. Sub-tabs keep one cursor, one add key, and one meaning for the rows. The cost is that the other two lists are one key away, and the counts in the strip say whether they hold anything. A walk question that needs a role opens the Roles list at once.

### Changing an entry

`a` adds an entry to the current list, and enter or `e` changes the chosen one. A form takes the whole body. Enter moves to the next field and saves on the last, ctrl+s saves from anywhere, and escape cancels with nothing saved. Letters are text while a form is open, so the screen keys are off and the header shows the tab names with no keys.

- **Contact.** Name, Email, Slack, and Slack id. Email, Slack, and Slack id are optional and each says what it is for in faint placeholder text.
- **Role.** A name, and the holders picked from the contacts. The pick works as a several-answers list does: space picks or unpicks, the numbers and an "Order:" line show the order, and the order is stored. A role with no holder can be saved, and the list then flags it vacant.
- **Team.** A name, a team contact, and a row for every role showing who holds it for this team. The team contact row comes first, because most teams have one contact and few have overrides. Enter on it opens a pick list of one choice, with "No team contact" on top and then the contacts, and enter on a row chooses it and closes the list. A team with no contact saves with an empty `contact`. The contact is part of the team the form sends, so changing an override or the name never drops it. A contact that is not a contact is not in the form, as for a holder, and saving drops it. Enter on a role row opens a pick list with two rows on top, "Use the role's own holders" and "Vacant for this team", then the contacts. The first stores no override. The second stores an empty override, which is different: the role is vacant for the team and does not fall back to its own holders. Picking a contact leaves the top two rows.
- **The id.** An entry's id is made from the name as a lowercase slug and shown under the title as it is typed: "Id: pat-lee, made from the name". If the id is taken in that list the form adds a number and says why: "Id: pat-lee-2, made from the name, with a number because pat-lee is taken". The id of an existing entry never changes, and the form says so. A letter outside A to Z, such as an accent, is left out of the id and stays in the name.
- **Replace whole.** A change sends the whole entry, as the editor replaces it whole. A form starts from the stored entry, so a field a person does not touch comes back as it was. A holder that is not a contact, and a team override for a role that is gone, are not in the form, and saving drops them. That clears a flag.
- **Refusal.** A problem from the editor shows under the form in words, such as "A name is needed.", with `! problem:` in front. The form stays open, the status says "not saved", and the plan is as it was. A name with no letter or digit gets "The name needs a letter or a digit, so an id can be made from it."

### Removing an entry

`x` on an entry does not remove it. The screen runs the editor's removal on the current plan as a dry run, and shows the result before anything is kept.

```text
  Remove the contact Sam Okafor?

  This also clears 1 answer and removes Sam Okafor from 1 role.

  Answers cleared
    - Who can anyone ask for help with anything?

  Holders removed
    - Engineering manager role

  Left with nobody
    - Engineering manager role has nobody left

  Nothing is removed until you press y.
```

- The sentence counts the cleared answers, and the roles and team overrides the entry leaves. A contact that is the contact of teams adds "and removes it as the contact of 1 team", and the places list "Platform team, as the team contact". The team stays and has no contact afterwards, as the editor reports in `team-contact:<team id>`. Below it are the questions, by their text and not their ids, and the places. "Left with nobody" names a role or a team override that the removal empties, so a person sees the vacancy before they cause it. A list of more than five is cut with "and 3 more".
- `y` applies it, with the plan the dry run returned, and saves. Every other key keeps the entry and says "nothing removed". The key is `y` and not enter because enter is the key a person presses to get on, and a removal needs a deliberate one.
- A cleared required question is unanswered afterwards, and the walk and the Review show it so.

### Saving and returning

- Every accepted edit and every removal is saved to the plan file at once, the way an answer is, and the status says "saved". A save that fails says "could not save", and the plan stays changed in memory as it does for an answer.
- After a change the screen asks the finder again, so the walk's questions, their pick lists, the counts, and the rail are current. Returning to the walk with escape, `q`, or `w` opens the question the person left, with the new entries in its list.
- Escape and `q` go back to the screen the person came from, including the Walk, the Overview, and the Review. From the Sections screen they go to where Sections was opened from.

### Contacts keys

| Key | List | Form | Removal |
| -- | -- | -- | -- |
| up, down, `k`, `j` | Move | Move between fields, or in a pick list |  |
| tab, shift+tab, left, right | Next or previous list | Next or previous field |  |
| `a` | Add to this list |  |  |
| enter, `e` | Change the chosen entry | Next field, save on the last, open the team contact or a role row, and choose a row in the team contact list | Keep |
| `x`, delete | Remove, after the dry run |  |  |
| space |  | Pick or unpick | Keep |
| ctrl+s |  | Save |  |
| `y` |  |  | Remove |
| `w`, `o`, `r`, `s` | Open that screen | Letters are text | Keep |
| esc | Back | Cancel, or close a pick list | Keep |
| `q` | Back | Letters are text | Keep |
| `?` | Help | `f1` | Help |

### Rejected on this screen

- A separate window for each kind. The three lists share one cursor model, so one screen is less to learn.
- A key to move a holder up or down. The numbered pick order is enough for the two or three holders a role has, as for any several-answers list.
- Enter to confirm a removal. See above.
- Inline adding from the walk's pick list. The Contacts screen is one key away, opens on the right list, and brings the person back to the same question. A second way to add would need the whole form inside the walk.
- Letting the id be typed. The name makes it, and a person who sees `sam-okafor` is not asked to invent one.

## Open

- Splitting this page into one file per view under `docs/design/`, with design decision records (DDRs) for choices that cut across views, such as the width breakpoints. DDRs stay separate from ADRs: an ADR is for a decision that contracts or acceptance tests depend on, and a DDR is for a design choice that only the screens depend on.
- A key to move a pick up or down in a several-answers list.
- Importing contacts from a directory service. It is not planned, see 0080.
- A key to move a holder up or down in the Contacts forms.
- Accents in ids. The slug keeps only A to Z and digits, so a name such as "José" makes the id `jos`.
- Whether the progress count should outrank the tab keys at 80 columns now that the strip has five tabs.
- Whether the Walk goes to the next unanswered question when the person is revisiting from Review. Today it goes to the next in order.
- How the tool reminds a person of deferred questions after they quit.
- Whether the rail should show the section titles when the window is wide and the plan is small.
- Whether the notes should hide while a person is typing, to leave only the answer.
- A mouse and a way to change the accent are not planned.
- The newcomer's view of the onboarder's answers, if the document preview proves not enough.
