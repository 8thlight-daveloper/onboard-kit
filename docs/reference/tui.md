# The terminal walkthrough

Status: living. Each point says whether it is decided, a direction, or open, and any of them can change as the screens are built and used. This page is not a contract. Unit tests drive the screens through key and window-size messages, and a person tries them in a real terminal.

The code is `internal/tui`, opened by `onboard-kit` with no argument, `onboard-kit FILE`, `onboard-kit open FILE`, `onboard-kit new`, `onboard-kit walk`, and `onboard-kit start`. The toolkit is Bubble Tea, see [language](../stack/language.md). The people it serves, and what each may answer, are in [0055](../decisions/0055-variable-audience.md) and [0060](../decisions/0060-personal-plan.md).

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
| Anyone | `onboard-kit`, or `obk` | The front door: the plans they opened before, and what they can do next | The plan list, and a new plan file after `y` |
| The onboarder | `walk --plan FILE [--root DIR]` | Onboarder questions, section switches, the assembled document, the people's preparations, and who is ready for an invitation | Onboarder answers and switches, one person's preparation after a reviewed save, and one new invitation file after `y` |
| The onboardee | `start --org-plan FILE` | Onboardee questions and their own document | Onboardee answers only |

Decided: the onboardee mode hides section switches. Only the onboarder decides what applies. Built: `start` makes the newcomer's own plan file from the org plan and opens `walk` for the onboardee audience. Refresh from a changed org plan, a URL, and the locks are not built, see 0060.

Decided: the onboardee does not get a separate list of the organization's answers. The document preview already shows them in context, and a second list would repeat it.

## Entry and roles

Decided by [0610](../decisions/0610-people-work-in-screens-and-flags-serve-scripts.md). Partly built. Each point says which.

- **One entry.** Built. `obk`, an alias of `onboard-kit`, opens the terminal
  interface. `just build` writes `output/obk` beside `output/onboard-kit`. The release
  archives and the Homebrew formula do not ship the alias yet. Flags serve scripts
  and agents.
- **The file decides.** Built. `obk FILE` opens a plan in the plan editor, exactly as
  `walk --plan FILE` does. Any other file is refused with one sentence that names
  what was expected, and the exit code is 1. `obk open FILE` is the explicit
  spelling for scripts, with the same behavior. Neither asks anything before it
  opens. A command name always wins over a file of the same name.
- **The invitation, interim.** `obk FILE` with an invitation, and Begin my
  onboarding, run the typed flow of `start --invitation FILE --root DIR` after the
  screen closes. `DIR` is the folder the tool was started in, and the screen and
  the first printed line say so before it hands over. This stays until the
  newcomer's screens are built.
- **No terminal.** Built. With no argument and no terminal, the tool prints its
  usage and one line that names the scripting commands, and exits 2. A plan or an
  invitation named with no terminal is refused the same way, naming
  `assemble --plan FILE` or `start --invitation FILE --root DIR`.
- **The front door.** Built, see Front door below. With no known plans the first
  screen asks "What would you like to do?" and offers "Build an onboarding plan"
  and "Begin my onboarding". With known plans it shows a picker, last used first,
  with "New plan", "Open a file", and "Begin my onboarding".
- **A new plan, interim.** `obk new` and Build an onboarding plan write the core
  set of [0530](../decisions/0530-new-plan-from-core-set.md) after the person sees
  the path and presses `y`. The reviewed creation draft of 0530, where a person
  removes and adds sections before the save, is not built. The person switches
  sections off on the Sections tab afterwards.
- **The header names the role.** Built, see Screen tabs. The onboarder sees
  "Building: Acme plan", with the tabs Answer, Overview, Review, Sections,
  Contacts, Preparation, and Invitation. The newcomer sees "Your onboarding at
  Acme", with the tabs Your questions, Overview, and Review. Acme is the plan's
  `organization` answer. Until it is answered, the onboarder sees "Building an
  onboarding plan" and the newcomer sees "Your onboarding". The Answer screen
  below is the one screen behind both Answer and Your questions. 0610 names the
  newcomer's third tab Document. It stays Review, because the newcomer's Review
  holds their blockers, deferred questions, and problems beside the document
  preview, not the document alone.
- **Tell before looking.** Not built. Before the tool looks beyond the file in the
  command, it says what it will read and why, and warns when the system may show
  a permission prompt. Looking for an invitation in the Downloads folder is the
  first case. Today the person types the invitation's path, so the tool looks
  nowhere else.
- **About.** Built. A screen lists every folder and file the tool uses, what each
  holds, and whether the tool reads it, writes it, or both, with the version and
  the template set. The front door opens it with `a`, and every help page that
  has it offers `a`.

## Screens

Decided. There are seven screens and a help page. Every screen has a header line with the tab strip, a progress rail, a body, a status line, and a key hint line, in that order. The hint line is always the last line.

| Screen | Holds | Opened by |
| -- | -- | -- |
| Answer | One question: its position, its section and group, what it asks, whether it is required, its current answer, the way to answer, suggestions, problems | The start, enter on an overview or review row, finishing a question |
| Overview | Every question in order with a mark for its state. The cursor jumps to any of them | `o` |
| Review | Blockers, deferred questions, validator problems, the document preview, and a quiet line when answers are left over, see Left over answers | `r`, finishing the last question, or the start when the plan holds answers |
| Sections | Section switches. The onboarder only | `s` |
| Contacts | The organization's contacts, roles, and teams, to add, change, and remove. The onboarder only | `c` |
| Preparation | The people who have a preparation, and one person's recipient details and arrangements, staged and saved after a review. The onboarder only, see Preparation | `t` |
| Invitation | The people who have a preparation and whether each is ready, the help for one person's invitation beside their first steps guide, and the written file with the line for the newcomer. The onboarder only, see Invitation | `i` |
| Help | Every key of the current screen | `?`, or `f1` while typing |

How a person moves:

- Start: if the plan holds an answer to any question of this audience, open Review. Otherwise open the Answer screen at the first question. A plan with no active questions opens Review.
- Enter accepts the answer on the screen and moves to the next question in order. This was a choice between the next question and the next unanswered one. The next in order is predictable, and a person who wants an unanswered one opens Review, where the blockers are listed.
- Past the last question, the person lands on Review.
- `w`, `o`, `r`, `s`, `c`, `t`, and `i` open the Answer, Overview, Review, Sections, Contacts, Preparation, and Invitation screens from any screen while the person is not typing. The tab strip names them, see Screen tabs.
- `p`, left arrow, or shift+tab goes back one question. `n` or right arrow skips forward without changing anything.
- `d` defers the question and moves on. `x` clears the answer and stays, so the person can answer again.
- Escape on the Answer screen returns to the Overview or Review, whichever sent the person there. Escape on Overview or Review returns to the Answer screen.
- From Overview or Review, enter opens the chosen question on the Answer screen. `w` on Review returns to the Answer screen at the first unanswered question, and `w` on Overview or Sections returns to the question the person left.
- `q` quits from the Overview, Review, and Answer screens, except while typing. The Sections, Contacts, Preparation, and Invitation screens take `q` as back, and escape goes back too, to the screen the person came from. Ctrl+c always quits.

## Layout

Decided. A designer judged the layout against the people in Intent, and it replaces the first draft, which put a question list beside every question. Bubble Tea sends the window size at the start and on every resize, and the layout follows it. Nothing has a fixed width. The design target is a window 120 to 140 columns wide.

### What the layout does

- The Answer screen shows one question and nothing else that competes with it. The answer column holds the section title, the position ("4 of 11, required, not answered yet"), the question, the way to answer, and any problem.
- A progress rail sits under the header: one mark per question, grouped by section, with a diamond at the current question. The rail replaces the question list. It answers "where am I and how much is left" in one quiet line, and it never asks for a read. It is hidden when it does not fit.
- The notes sit beside the answer at 120 columns and wider, and below it when narrower. They hold the example, the type and options, who answers, the name, and the suggestions. They are faint, because they are for a glance and never for action.
- The Overview is one list, grouped by section, with each row showing its mark, its question, and its answer. At 120 columns and wider a faint preview of the chosen question sits beside it.
- The Review keeps two panes, because it is a master and detail task: the items on the left, the document preview on the right. Below 120 columns, tab switches between them.

| Width | Answer | Overview | Review |
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

- The question list beside the Answer screen's question (A). The product owner found it distracting, and the reasons hold: it is the brightest thing on screen and the person is not acting in it. The rail keeps what the list gave, a sense of place, at a tenth of the weight.
- The document preview beside the Answer screen's question (D). It is the right tool for the Review, where the person judges the whole, and noise on the Answer screen, where the person answers one thing.
- No rail and no notes (C). Fewer parts, but a person answering 40 questions loses their place, and the example needs a home.
- A layout fixed at 80 columns. Modern terminals are wider. The 120 column break is where the notes fit beside a comfortable answer column. Below it the notes move under the answer, and the tool is still usable at 80.
- A vertical rule down the whole screen. The rule stops where the content stops.

Direction: the onboardee sees the same screens with fewer parts: no section switches, a welcome line while nothing is answered, and "3 of 5 to answer" in the header.

## Emphasis

Decided. The part of the screen where the person is acting is drawn bright, and every other part is drawn faint. Emphasis follows focus.

| Screen | Bright | Faint |
| -- | -- | -- |
| Answer | The question, the answer, the problems | The rail, the notes, the status |
| Overview | The list | The rail, the preview |
| Review | The pane with focus, items or document | The other pane |
| Sections, Help | The whole body |  |
| Contacts, list | The list and the list names at the top | The detail of the chosen entry |
| Contacts, form and removal | The whole body |  |
| Preparation, list and person | The list or the person's rows | The notices are body text, and the not saved notice is a warning |
| Preparation, forms, confirmations, review | The whole body, or the review pane with focus | The other review pane |

- The accent marks "you are here" and nothing else: the `>` before the chosen row or the focused pane, the diamond in the rail, the prompt of the answer box, and the current tab. Titles, headings, and key names are never in the accent. A title is bold, and a key name is bold.
- Bold and faint are enough to carry the hierarchy, so it holds on a light or dark background and with no colour.

## Wireframes

Decided. No wireframe tool fits a terminal screen, so the screens are drawn by the code itself. A test renders the real view from a fixed plan with no colour and compares it to a text file. The files are the wireframes, and they cannot drift from the screen. Regenerate them on purpose with `go test ./internal/tui -run Wireframes -update`, and review the diff.

| Screen | 80 columns | 120 columns | 140 columns |
| -- | -- | -- | -- |
| Answer, choosing | [answer](../../internal/tui/testdata/screens/answer.80.txt) | [answer](../../internal/tui/testdata/screens/answer.120.txt) | [answer](../../internal/tui/testdata/screens/answer.140.txt) |
| Answer, typing | [edit](../../internal/tui/testdata/screens/edit.80.txt) | [edit](../../internal/tui/testdata/screens/edit.120.txt) | [edit](../../internal/tui/testdata/screens/edit.140.txt) |
| Answer, with an example | [example](../../internal/tui/testdata/screens/example.80.txt) | [example](../../internal/tui/testdata/screens/example.120.txt) | [example](../../internal/tui/testdata/screens/example.140.txt) |
| Answer, picking one | [pick](../../internal/tui/testdata/screens/pick.80.txt) | [pick](../../internal/tui/testdata/screens/pick.120.txt) | [pick](../../internal/tui/testdata/screens/pick.140.txt) |
| Answer, picking several | [pick many](../../internal/tui/testdata/screens/pick-many.80.txt) | [pick many](../../internal/tui/testdata/screens/pick-many.120.txt) | [pick many](../../internal/tui/testdata/screens/pick-many.140.txt) |
| Answer, nothing to pick | [pick empty](../../internal/tui/testdata/screens/pick-empty.80.txt) | [pick empty](../../internal/tui/testdata/screens/pick-empty.120.txt) | [pick empty](../../internal/tui/testdata/screens/pick-empty.140.txt) |
| Overview | [overview](../../internal/tui/testdata/screens/overview.80.txt) | [overview](../../internal/tui/testdata/screens/overview.120.txt) | [overview](../../internal/tui/testdata/screens/overview.140.txt) |
| Overview with sub-sections | [subsections](../../internal/tui/testdata/screens/subsections.80.txt) | [subsections](../../internal/tui/testdata/screens/subsections.120.txt) | [subsections](../../internal/tui/testdata/screens/subsections.140.txt) |
| Review | [review](../../internal/tui/testdata/screens/review.80.txt) | [review](../../internal/tui/testdata/screens/review.120.txt) | [review](../../internal/tui/testdata/screens/review.140.txt) |
| Review, with answers left over | [review-left-over](../../internal/tui/testdata/screens/review-left-over.80.txt) | [review-left-over](../../internal/tui/testdata/screens/review-left-over.120.txt) | [review-left-over](../../internal/tui/testdata/screens/review-left-over.140.txt) |
| Review, removing left over answers | [review-prune](../../internal/tui/testdata/screens/review-prune.80.txt) | [review-prune](../../internal/tui/testdata/screens/review-prune.120.txt) | [review-prune](../../internal/tui/testdata/screens/review-prune.140.txt) |
| Sections | [sections](../../internal/tui/testdata/screens/sections.80.txt) | [sections](../../internal/tui/testdata/screens/sections.120.txt) | [sections](../../internal/tui/testdata/screens/sections.140.txt) |
| The onboardee's questions | [onboardee](../../internal/tui/testdata/screens/onboardee.80.txt) | [onboardee](../../internal/tui/testdata/screens/onboardee.120.txt) | [onboardee](../../internal/tui/testdata/screens/onboardee.140.txt) |
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
| Preparation, people | [people](../../internal/tui/testdata/screens/preparation-people.80.txt) | [people](../../internal/tui/testdata/screens/preparation-people.120.txt) | [people](../../internal/tui/testdata/screens/preparation-people.140.txt) |
| Preparation, adding a person | [add](../../internal/tui/testdata/screens/preparation-add.80.txt) | [add](../../internal/tui/testdata/screens/preparation-add.120.txt) | [add](../../internal/tui/testdata/screens/preparation-add.140.txt) |
| Preparation, one person | [person](../../internal/tui/testdata/screens/preparation-person.80.txt) | [person](../../internal/tui/testdata/screens/preparation-person.120.txt) | [person](../../internal/tui/testdata/screens/preparation-person.140.txt) |
| Preparation, recipient form | [recipient](../../internal/tui/testdata/screens/preparation-recipient.80.txt) | [recipient](../../internal/tui/testdata/screens/preparation-recipient.120.txt) | [recipient](../../internal/tui/testdata/screens/preparation-recipient.140.txt) |
| Preparation, picking a manager | [pick](../../internal/tui/testdata/screens/preparation-pick.80.txt) | [pick](../../internal/tui/testdata/screens/preparation-pick.120.txt) | [pick](../../internal/tui/testdata/screens/preparation-pick.140.txt) |
| Preparation, new arrangement | [arrangement](../../internal/tui/testdata/screens/preparation-arrangement.80.txt) | [arrangement](../../internal/tui/testdata/screens/preparation-arrangement.120.txt) | [arrangement](../../internal/tui/testdata/screens/preparation-arrangement.140.txt) |
| Preparation, picking helpers | [helpers](../../internal/tui/testdata/screens/preparation-helpers.80.txt) | [helpers](../../internal/tui/testdata/screens/preparation-helpers.120.txt) | [helpers](../../internal/tui/testdata/screens/preparation-helpers.140.txt) |
| Preparation, a refused form | [problem](../../internal/tui/testdata/screens/preparation-problem.80.txt) | [problem](../../internal/tui/testdata/screens/preparation-problem.120.txt) | [problem](../../internal/tui/testdata/screens/preparation-problem.140.txt) |
| Preparation, removing an arrangement | [remove](../../internal/tui/testdata/screens/preparation-remove.80.txt) | [remove](../../internal/tui/testdata/screens/preparation-remove.120.txt) | [remove](../../internal/tui/testdata/screens/preparation-remove.140.txt) |
| Preparation, dropping staged changes | [drop](../../internal/tui/testdata/screens/preparation-drop.80.txt) | [drop](../../internal/tui/testdata/screens/preparation-drop.120.txt) | [drop](../../internal/tui/testdata/screens/preparation-drop.140.txt) |
| Preparation, review | [review](../../internal/tui/testdata/screens/preparation-review.80.txt) | [review](../../internal/tui/testdata/screens/preparation-review.120.txt) | [review](../../internal/tui/testdata/screens/preparation-review.140.txt) |
| Preparation, review with the guide in focus | [guide](../../internal/tui/testdata/screens/preparation-guide.80.txt) | [guide](../../internal/tui/testdata/screens/preparation-guide.120.txt) | [guide](../../internal/tui/testdata/screens/preparation-guide.140.txt) |
| Window too small | [too small](../../internal/tui/testdata/screens/too-small.80.txt) | [too small](../../internal/tui/testdata/screens/too-small.120.txt) | [too small](../../internal/tui/testdata/screens/too-small.140.txt) |

## Colour

Decided, see [0560](../decisions/0560-use-the-terminal-colors.md). The screens use the terminal's own colours and none of ours: the default foreground and background, and the sixteen colours the person's theme defines. One accent, the terminal's green in bold, marks "you are here", see Emphasis. Two other colours mean status and nothing else: the terminal's yellow for a warning and its red for a problem. Everything quiet is drawn with the terminal's faint style, so it works on a light or a dark background and the tool never asks the terminal for its background colour. The styles are built when a screen is drawn, never at package level.

- No screen uses a hexadecimal, 256-colour, or 24-bit value. `NO_COLOR` and a terminal with no colour get plain text.
- Colour never carries meaning alone. The chosen row starts with `>`, and the current tab is in brackets. States are words and marks: `[x]` answered, `[~]` deferred, `[ ]` not answered, `*` required, and in the rail `x`, `~`, `.`, and a diamond for the current question. The position and state are also words under the section title. Problems start with `! problem:` and warnings with `warning:`. The focused pane of the Review shows `>` before its title.
- Tests render every screen with no colour and check that these marks and the key hints are still there. Other tests check, with colour on, that the focused pane is bright, the other is faint, the accent is not on a title, and no screen uses a fixed colour value.

## The example

Decided. A question may carry one `Example`, an illustration of a good answer. It is shown as faint placeholder text in the answer box, after "e.g.", and as an "Example:" line in the notes. It is never an answer and never a default: enter on an empty box does not save it, and it disappears when the person types.

## Answering

Decided, and built. The answer is made on the Answer screen, under the question.

- A boolean or choice is picked from a list, with "deferred, answer later" and "clear the answer" at the end. The cursor starts on the current answer. Up and down move, enter picks.
- A contact, role, or team is picked from the organization's directory, see Pick lists. A person never types an id.
- A string takes typed text, with the current answer filled in. Tab cycles through the question's suggestions. Enter on text equal to the current answer moves on without saving. Enter on an empty box clears the answer. Ctrl+d defers.
- A command is words separated by spaces, with quotes keeping spaces together.
- Several contacts or several roles in a `multiple` question are picked as a checklist, see Pick lists.
- An answer that the validator refuses stays on the Answer screen with the problem message under the answer.
- Every accepted answer is saved to the plan file at once. There is no unsaved state.
- While a person types, letters are text, so the single-letter keys are off. Back, forward, defer, and help have control-key forms listed below.

## Pick lists

Decided, and built. A `contact`, `role`, or `team` question carries the directory's entries as `Question.Choices`, each with an id and a name, in directory order, see [0080](../decisions/0080-directory.md). The Answer screen shows the names and stores the id. The id stays out of sight. The Contacts screen is where entries are made, see Contacts.

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
- **Stale answers.** An answer that names an id the directory no longer holds counts as not answered, as the finder counts it. The Answer screen, the Overview, the rail, and the counts show it as not answered, and no id is shown. A faint "was sam-okafor" was rejected: the validator already warns about the stale id on the Review, and the person's task is to pick again. A list with at least one known id counts as answered and shows only the known names. Confirming the picks drops the unknown ids.

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
- **Key.** `h`. It is free on the Review and is not typing, so it cannot collide with text. `x` was rejected because it clears an answer on the Answer screen, and `p` because it is back. `r` is the Review itself. The hint line shows `h left over` while something is left over. With nothing left over, `h` says "No answers are left over from removed questions." and changes nothing.
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

| Key | Answer, choosing | Answer, typing | Overview | Review | Sections |
| -- | -- | -- | -- | -- | -- |
| enter | Pick, move on. On a several-answers list, confirm the picks | Accept, move on | Open the question | Open the question, or the Contacts entry of a directory problem | Switch on or off |
| up, down, `k`, `j` | Move the choice |  | Move | Move | Move |
| space | Pick or unpick on a several-answers list |  |  |  | Switch on or off |
| `p`, left, shift+tab | Back | shift+tab, ctrl+p |  |  |  |
| `n`, right | Skip forward | ctrl+n |  |  |  |
| `d` | Defer, move on | ctrl+d |  |  |  |
| `x` | Clear | Empty the box and enter |  |  |  |
| tab |  | Try a suggestion |  | Items or document |  |
| `w`, `o`, `r`, `s`, `c`, `t`, `i` | `o`, `r`, `s`, `c`, `t`, `i` | Letters are text | `w`, `r`, `s`, `c`, `t`, `i` | `w`, `o`, `s`, `c`, `t`, `i` | `w`, `o`, `r`, `c`, `t`, `i` |
| pgup, pgdn, `g`, `G` |  |  | `g`, `G` | Scroll the document |  |
| esc | Back to overview or review | Same | Answer | Answer | Back |
| `?` | Help | f1 | Help | Help | Help |
| `q` | Quit |  | Quit | Quit | Back |

On the Review, `h` shows the answers left over from questions the plan no longer asks, see Left over answers. It is free on every other screen and is not a screen key, so the tab strip does not list it. The key hint line shows `h left over` only while something is left over.

`s`, `c`, `t`, `i`, the Sections screen, the Contacts screen, the Preparation screen, and the Invitation screen do not exist in the onboardee mode. The key hint line leaves out the screen keys while the tab strip shows them, see Screen tabs.

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

Decided. The app is a Bubble Tea model, so tests send key messages and window-size messages and check the plan, the saved copies, and the view. At widths from 60 to 200 and heights from 12 to 50, a test checks that no line is wider than the window, including questions with very long descriptions, examples, and suggestions. Another checks the small-window message under the limits, and that the Answer screen has two columns at 120 and one below. The flow tests cover the guided order, the position text, back, defer and move on, finishing on the Review, jumping from the Review, and the onboardee mode. A fake Backend with hand-built questions tests the example, the pick lists, and the sub-sections. The pick tests cover one answer, several answers in order, an empty directory for each audience, same-named entries, and a stale answer. Tab tests check the strip with and without colour, and the keys it drops at narrow widths. Role header tests check the title for both people with the organization answered, not answered, deferred, blank, long, and holding control characters, the title changing as the onboarder answers, and the current tab on screen at 60, 80, 100, and 120 columns on every tab. Another checks that no wireframe and no help page says walk. The Contacts tests drive the screen through the real editor: adding, changing, and removing each kind, the id made from a name, a refusal that leaves the plan as it was, the dry run before a removal, the flags for vacant roles and unreachable contacts, the autosave of every accepted edit, and the Answer screen's pick list after returning. Style tests check the bright and faint panes in each state. The prune tests drive the Review through the real editor for both audiences: the notice only when something is left over, the dry run that changes and saves nothing, every key but `y` keeping the answers, `y` saving once and clearing the notice, the short forms of values with no long value on screen, the cut list, and `h` as text while typing. A fake Backend covers the update logic, and the width tests include the notice and the confirmation. The Preparation tests drive the screen through a fake port and the real preparation editor, reviewer, and encoder: the tab only for the onboarder with a port, the list, adding a person, typing with no extra step, the pick lists, several changes staged into one save, nothing written before `y`, the real guide and blockers in the review, replacing and removing an arrangement, a refused form, a taken id, dropping staged changes, a stale save applied again, each refusal, and control characters shown as text. The Invitation tests drive the screen through a fake port and the real preparation reviewer: the tab only for the onboarder with a port, the readiness counts, the blocked page with no form, the default file, the pick list of contacts with a route, the guide that follows the picked help, the required pre-access help, nothing written before `y`, the request with the pre-access help first, the result line, the next free name after a file appears, and letters as text in the file name. The walk command's tests run both store-backed ports on real files, and the invite command's tests run `Export` and `Check` on real files. A scripted pseudo-terminal run covers start and quit, another adds a person, types a name, picks a team, reviews, and saves with `y` through the built binary, and a third writes an invitation, then tries the same name again and takes the next free one. It must answer the terminal's background-colour query and its cursor report, or the program waits five seconds.

## Screen tabs

Decided, and built. A designer asked for it after using the redesigned question screen, then named Walk: a bare title does not show which screen a person is on.

The header line holds the title, the tab strip, and the progress count at the right.

```text
  Building: Acme plan  Answer w  [Overview]  Review r  Sections s  Contacts c          13 of 13 answered
  Your onboarding at Acme  Your questions w  [Overview]  Review r                        1 of 3 to answer
```

- The title names whose screen it is, under [0610](../decisions/0610-people-work-in-screens-and-flags-serve-scripts.md). The onboarder reads "Building: Acme plan" and the newcomer reads "Your onboarding at Acme", where Acme is the plan's `organization` answer. With no answer, a deferred one, or only spaces, the titles are "Building an onboarding plan" and "Your onboarding". The title follows the plan, so it changes as soon as the onboarder answers the question. The name shows at most 24 columns and a longer one is cut with `…`, so a long name cannot push the tabs out. Control characters in the name show as text, such as `\x1b`.
- The strip names the screens a person can move between, in a fixed order: Answer, Overview, Review, Sections, Contacts, Preparation, Invitation. The newcomer's first tab is named Your questions. It is the same screen. A screen the audience cannot use, such as Sections, Contacts, Preparation, and Invitation for the onboardee, is left out. Preparation and Invitation also need the `walk` command to supply their ports, which it always does for the onboarder, see Preparation and Invitation. The Contacts key is lowercase `c`. It was the capital `P` when the screen was named People. `p` is back on the Answer screen, so `P` was a special case that only worked with shift. `b` is page-up on the scrolling screens. `c` is free on every screen and matches the screen name.
- The Answer key is `w`, the key of the screen when it was named Walk. It stays because it is a letter of "Answer", as `t` is a letter of "Preparation". `a`, the first letter, adds an entry on the Contacts screen and opens About from a help page. No free letter is in both "Answer" and "Your questions", so the newcomer's Your questions tab takes `w` too, and one key opens the same screen for both people.
- The current tab is bold in the accent colour, which is "you are here". With no colour it is in brackets. The other tabs are faint, each with its key.
- The strip is the one place the screen keys are shown, so the key hint line leaves them out. The Answer screen while choosing, the Overview, the Review, and the Sections screen all take `w`, `o`, `r`, and `s`, except while the person is typing.
- While a person types, letters are text, so the tabs show their names with no keys, and the hint line says `esc` leaves.
- When the header is too wide, it drops the progress count first, then the keys. With the keys gone, the hint line shows them again, so a key is always on screen somewhere. The title "Building: ACME Corp plan" is 24 columns, and "Building an onboarding plan" is 27. With the 24-column title and all seven tabs, the title and the strip with keys are 109 columns wide, so at 120 columns the progress count goes and the keys stay, and at 100 columns the keys go too. With five tabs they are 80 columns wide, so at 100 columns the keys stay and at 80 columns they go. At 60 columns the strip shows names only, and its gaps shrink to one column.
- When even the names do not fit, the strip is cut after the last tab that fits and ends with `…`. If the cut would hide the current tab, the title goes instead, so the current tab is always on screen. At 60 and 80 columns the onboarder's seven tabs need one of these two. With the seventh tab, Invitation, the strip without the title can still be too wide when a late tab is current. Then the strip shows a window of tabs around the current one, with `…` on each side that is cut, such as `… Review Sections Contacts Preparation [Invitation]`.

Rejected: the screen name beside the title, as before (`Onboarding plan  Overview`), because it shows one screen and hides that there are others. Also rejected: a tab for Help, because help is a page over the current screen and the current tab stays marked while it is open.

## Sub-sections

Decided, and built. Under [0070](../decisions/0070-sub-sections.md) a section can have a parent.

- The Overview shows a sub-section under its parent, indented two columns for each level of depth. The indent applies to the section title and to its question rows.
- The progress rail keeps the cells of a section and its sub-sections together, with the usual gap only between families of sections.
- The Answer screen shows the sub-section's own title. The document preview nests it by heading level.

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
- A cleared required question is unanswered afterwards, and the Answer screen and the Review show it so.

### Saving and returning

- Every accepted edit and every removal is saved to the plan file at once, the way an answer is, and the status says "saved". A save that fails says "could not save", and the plan stays changed in memory as it does for an answer.
- After a change the screen asks the finder again, so the Answer screen's questions, their pick lists, the counts, and the rail are current. Returning to the Answer screen with escape, `q`, or `w` opens the question the person left, with the new entries in its list.
- Escape and `q` go back to the screen the person came from, including the Answer screen, the Overview, and the Review. From the Sections screen they go to where Sections was opened from.

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
- Inline adding from the Answer screen's pick list. The Contacts screen is one key away, opens on the right list, and brings the person back to the same question. A second way to add would need the whole form inside the Answer screen.
- Letting the id be typed. The name makes it, and a person who sees `sam-okafor` is not asked to invent one.

## Preparation

Built, under [0520](../decisions/0520-preparation-screen.md). The save rules of [0340](../decisions/0340-guided-preparation-save.md) and [0350](../decisions/0350-guided-support-arrangements.md) still apply. The onboarder's screen for what is arranged for each newcomer: their recipient details and their arrangements, as the [preparation format](preparation-format.md) defines them. The onboardee has no tab and no key for it. There is no scripted form of this screen. The `prepare` commands stay for scripts.

### Where the files are

`walk --plan FILE` looks for preparations under the folder that holds `FILE`. `--root DIR` names another folder, and a `DIR` that is not a folder is refused before anything opens. Each person's preparation is at the path the [preparation locator](../../contracts/preparation/locator/contract.md) derives from the plan's results path, by default `records/<account>/preparation.obk.yaml`. The `tui` package does no file work. It defines a port, `PreparationPort`, that lists the people, loads one, and saves reviewed bytes, and the walk command backs it with the [preparation store](../../contracts/preparation/store/contract.md). The port carries plan and preparation values and bytes, never store types.

### What the Preparation screen shows

- **People.** The list of people who have a preparation, by account and path, and an "Add a person" row. A file the store cannot read shows `! cannot be read`. The list reads no file body. Notices at the right, or below under 120 columns, say what a preparation is.
- **Adding a person.** `a`, or enter on "Add a person", asks for the account, which names the person's folder and never changes. Enter loads it. A new account opens an empty preparation with nothing written. An account that already has one opens it with a warning. An account the locator refuses keeps the box open with the rule.
- **One person.** The account and the path the file saves to stay at the top of every screen for this person, including the forms and the review. Below them are the recipient fields, each arrangement on one row with its id, section, kind, and status, an "Add an arrangement" row, and a "Review and save" row.
- **Notices.** A notices area sits beside the person, or below under 120 columns. It says what is not saved and how many changes are staged, that saving creates the person's folder when it is missing, that two arrangements share an id when they do, and who can read the file.

### Changing a preparation

- **Recipient.** Enter on the recipient opens a form with the name, company email, planned start, team, manager, onboarding owner, buddy, work location, and time zone. Typing goes straight into a text field. The team comes from the plan's teams, with "No team" on top. The manager, owner, and buddy come from the plan's contacts, with "Nobody" on top. Work location is "Not set", remote, on-site, or hybrid. Enter on a pick row opens its list, and enter chooses, as the team contact does on the Contacts screen.
- **Arrangements.** `a`, or enter on "Add an arrangement", opens a form with the id, section, kind, owner, helpers, reference, target date, status, and note. The id is typed, and it must not be taken. Enter on an arrangement opens it with its id fixed. The section comes from the plan's sections, with "(switched off)" after a section that is off. The kind and the status come from the lists in the preparation format. The helpers are a several-answers list, in pick order.
- **Staging.** Enter on the last field or ctrl+s stages the form through the [preparation editor](../../contracts/preparation/editor/contract.md). The editor's problems show under the form with the field's name, such as "! problem: Owner: Assign an owner before marking the arrangement arranged.", and nothing is staged. Escape cancels the form.
- **Removal.** `x` on an arrangement asks first, as a contact removal does, and `y` stages the removal.
- **Leaving.** Escape on a person with staged changes asks before it drops them, and `y` drops them. Staged changes survive a visit to another tab. Quitting drops them.

### Review and save

- "Review and save" shows the staged changes on the left and the real first steps guide from the [preparation reviewer](../../contracts/preparation/reviewer/contract.md) on the right. Below 120 columns, tab switches between them, as on the Review. The left also lists the preparation problems, the onboarder blockers in the plan, and the guide problems. When the preparation has problems, the blockers and the guide say they are not checked yet.
- Each change reads as the new value and the old one: `Name: "Alex Doe", was not set`. A new arrangement lists its fields. A removed one reads "removed".
- `y` saves. Nothing is written before it, and only the reviewed bytes are written. A preparation the encoder cannot write, or one over 1 MiB, cannot be saved, and the review says why. With no changes, `y` writes nothing and says so.
- When the plan or the preparation changed on disk since it was opened, the store refuses. The screen loads the person again, applies the staged changes to the new version, and shows the review again. A change that no longer applies is dropped, and the status says how many.
- After a save, the person is loaded again with nothing staged, and the status names the saved path. A save that leaves a temporary file behind says so, and so does a refusal.

### Preparation keys

| Key | People | Person | Form | Confirmation | Review |
| -- | -- | -- | -- | -- | -- |
| up, down, `k`, `j` | Move | Move | Move between fields, or in a pick list. `k` and `j` are text in a text field |  | Scroll the pane with focus |
| `g`, `G` | First, last | First, last |  |  |  |
| enter, `e` | Open the person, or add one | Open the row | Next field, open a pick list, choose, stage on the last field | Keep |  |
| `a` | Add a person | Add an arrangement | Letters are text | Keep |  |
| `x`, delete |  | Remove the chosen arrangement, after a confirmation | Letters are text | Keep |  |
| space |  |  | Pick or unpick in a helpers list | Keep | Scroll |
| tab, shift+tab |  |  | Next or previous field | Keep | Changes or guide |
| pgup, pgdn |  |  |  | Keep | Scroll the pane with focus |
| ctrl+s |  |  | Stage the form |  |  |
| `y` |  |  |  | Remove, or drop the staged changes | Save |
| `w`, `o`, `r`, `s`, `c`, `i` | Open that screen | Open that screen | Letters are text | Keep | Open that screen |
| esc | Back | Back to the people, after a confirmation when changes are staged | Cancel, or close a pick list | Keep | Back to the person |
| `q` | Back | Same as esc | Letters are text | Keep | Same as esc |
| `?` | Help | Help | `f1` | Help | Help |

### Why `t`

The screen needs a key that is free on every screen. `p` is back on the Answer screen, `r` is the Review, `a` and `e` act on the Contacts list, `x`, `d`, `n`, and `q` act on the Answer screen, and `i` is left for the Invitation tab that [0610](../decisions/0610-people-work-in-screens-and-flags-serve-scripts.md) plans. `t` is free on every screen and is in the name, "preparation".

### Rejected on the Preparation screen

- Picking the kind and the status as typed text. The preparation format lists both, so a pick list cannot be misspelled. The editor still judges the value.
- Showing the whole original and proposed preparation in the review, as the retired text guide did. The changes, read as new and old values, are shorter, and the guide is the larger thing to judge.
- Dropping the staged changes when a save finds the files changed. The person would lose work that a reload and a second review can keep.

## Invitation

Built, under [0610](../decisions/0610-people-work-in-screens-and-flags-serve-scripts.md). The export rules of [0300](../decisions/0300-explicit-invitation-export.md) and the [invite contract](../../contracts/command/invite/contract.md) still apply. The onboarder's screen for writing the file a newcomer opens. It replaces running `prepare` and `invite export` by hand. The onboardee has no tab and no key for it. `invite export` stays for scripts.

### Where the invitation comes from

The people are the ones the [preparation store](../../contracts/preparation/store/contract.md) lists under the same folder as the Preparation tab. The `tui` package does no file work. It defines a second port, `InvitationPort`, that lists the people with their readiness and creates one invitation. The walk command backs it with the store and two functions of the invite command: `Check`, which reads the plan and the preparation with the export's own bounded readers and runs its own checks, and `Export`, which does everything `invite export` does after reading its arguments. So the tab and the flags cannot drift.

### What the Invitation screen shows

- **People.** Each person with a preparation, by account, with "Ready" or "! not ready, N to fix". N counts the preparation problems, the onboarder blockers in the plan, and a missing onboarding owner, the same checks the export runs. A plan or preparation that cannot be read or used reads "! not ready, cannot be checked". The list checks everyone again each time the tab opens. Notices say what an invitation holds and that the tool sends nothing.
- **What blocks.** Enter on a person who is not ready shows what blocks the invitation, grouped as on the Preparation review, and no form. The way to fix each is on the Answer or the Preparation tab.
- **The form.** Enter on a ready person opens three fields: the pre-access help, the general help, and the file. A line under the account names the whole path the file will be written to, because the field row may cut it. The first steps guide sits beside the form, or below it under 120 columns, and follows the picked help.
  - **Pre-access help.** Required, and the row says so: "required, not chosen". A pick list of the plan's contacts that have a name and an email or Slack handle, as the export requires, with that rule under the list. A Slack id alone does not count.
  - **General help.** Optional, and the row says so: "optional, nobody". A several-answers list of the plan's contacts and roles, in pick order.
  - **File.** Typed, and filled with `obk-invitation-<account>.zip` in the plan's folder. The name starts `obk-invitation` because the newcomer's flow looks for that prefix in the Downloads folder. When that file exists, the next free name is filled in instead: `obk-invitation-<account>-2.zip`, then `-3`.
- **The last step.** Enter on the file, or ctrl+s, shows what will be written beside the guide. `y` writes the file. Nothing is written before `y`, and an existing file is never replaced.
- **The result.** "Give this file to Alex Doe. They run: obk obk-invitation-alex-doe.zip" in bold, then a line that the tool sends nothing and that the pre-access help is not tested. The wording of the line for the newcomer is one constant in `internal/tui`. A path the shell would split is quoted.
- **A refusal.** When a file appeared at that name since the form opened, the form opens again with the next free name filled in, and nothing is written. Any other refusal shows its reason under the form.

### Invitation keys

| Key | People | What blocks | Form | Last step | Result |
| -- | -- | -- | -- | -- | -- |
| up, down, `k`, `j` | Move |  | Move between fields, or in a pick list | Scroll the guide |  |
| `g`, `G` | First, last |  |  |  |  |
| enter, `e` | Open the person | Back | Open a pick list, choose, or go on from the file |  | Back to the people |
| tab, shift+tab |  |  | Next or previous field |  |  |
| space |  |  | Pick or unpick in the general help list |  |  |
| pgup, pgdn |  |  | Scroll the guide | Scroll the guide |  |
| ctrl+s |  |  | Go on to the last step |  |  |
| `y` |  |  | Letters are text in the file | Write the file |  |
| `w`, `o`, `r`, `s`, `c`, `t` | Open that screen | Open that screen | Letters are text in the file |  | Open that screen |
| esc, `q` | Back | Back | Back to the people. `q` is text in the file | Back to the form |  |
| `?` | Help | Help | `f1` | Help | Help |

### Why a last step

The form already shows the guide, so `y` could write from it. But the file name is typed, so `y` is a letter there. The last step lists the file, the help, and the guide once more, and keeps `y` the one key that writes, as on the Preparation review and the Contacts removal.

### Rejected on the Invitation screen

- Opening the form for a person who is not ready. The export would refuse, so the form would be a dead end. The page of what blocks says where to fix it.
- Offering every contact for the pre-access help and refusing on write. The pick list shows only the contacts the export accepts.
- Writing a numbered name without asking when the chosen file exists. The form shows the free name and waits for `y` again.

## Front door

Built, under [0610](../decisions/0610-people-work-in-screens-and-flags-serve-scripts.md). The first screen of `onboard-kit` with no argument. The code is `internal/tui/front_door.go`. The `tui` package does no file work: a port, `FrontDoorPort`, lists the known plans, forgets one, checks a file's kind, proposes and writes a new plan, and gives the About entries. `internal/command/frontdoor` backs it. When the screen closes it says what the person chose, open a plan or begin an onboarding, and the command does it.

### What the front door shows

- **The header.** `onboard-kit` on the left and the screen's name on the right: Start, New plan, Open a file, Begin my onboarding, or About. There is no tab strip, because nothing is open yet.
- **No known plans.** "What would you like to do?" with "Build an onboarding plan" and "Begin my onboarding". A faint line under the list says what the chosen row does.
- **Known plans.** "Your plans, last used first". Each row shows the file name, "opened" and the local date and time, and the folder. The last used plan is chosen when the screen opens. Under the plans, after a blank line: "New plan", "Open a file", and "Begin my onboarding". The plans come from [the plan list](plan-list.md).
- **A missing plan.** A plan whose file is gone reads "missing" in place of the time. The hint line offers no enter on it, and enter only says it is not where it was and to press `x` to remove it from the list.
- **Remove from list.** `x` on a plan asks "Remove acme.obk.yaml from the list?" and says the file stays where it is. `y` forgets it. Esc, or any other key, keeps it, and `?` opens help over the question. The file is never touched.
- **New plan.** A file field filled with `onboarding.obk.yaml`, or `onboarding-2.obk.yaml` when that exists, in the folder the tool was started in. A line under it names the whole path. Enter, or ctrl+s, shows the last step, "Write this new plan?", with the path. `y` writes the core set, never over an existing file, adds it to the plan list, and opens it in the editor. A refusal returns to the field with its reason.
- **Open a file.** A file field. Enter checks the file's kind. A plan opens in the editor. An invitation goes on to Begin my onboarding with the path filled in. Any other file shows the same one sentence as `obk FILE`, under the field after `! problem:`. A path may start with `~/` for the home folder.
- **Begin my onboarding.** A field for the invitation file's path, with `~/Downloads/obk-invitation-your-name.zip` as the example, a line that says the records are kept in the folder the tool was started in, and a faint line that says the tool will ask the person to confirm who they are and that they type each reply. Enter on an invitation closes the screen and starts the typed flow. A plan typed here is refused with a sentence that says it is a plan.
- **A damaged plan list.** The first screen shows no plans, a line under the question that starts `! problem:` and says to press `a` to see why, and Open a file as a third choice. The line stays until the screen closes. About names the file and the reason.

### Front door keys

| Key | Start | New plan, Open a file, Begin my onboarding | Last step of New plan | About |
| -- | -- | -- | -- | -- |
| up, down, `k`, `j` | Move |  |  | Scroll |
| `g`, `G` | First, last | Letters are text |  | Top, end |
| pgup, pgdn |  |  |  | Scroll a page |
| enter | Open the plan, or choose what to do | Check and go on |  | Back |
| ctrl+s |  | Go on to the last step, on New plan |  |  |
| `x` | Remove the plan from the list, after `y` | Letters are text |  |  |
| `y` | Confirms a removal | Letters are text | Write the plan |  |
| `a` | About | Letters are text |  | Back |
| esc | Nothing, so pressing it twice to leave a field never quits | Back to Start. When `obk new` opened New plan, quit | Back to the field | Back |
| `q` | Quit | Letters are text |  | Back |
| `?` | Help | `f1` | Help | Help |
| ctrl+c | Quit | Quit | Quit | Quit |

Help lists the keys of the screen it was opened from. `a` on any help page opens About, and any other key closes the help.

### `obk new`

With a terminal, `onboard-kit new [--out FILE]` opens New plan directly, with `FILE` or the proposed name in the field. Esc quits, because there is no first screen behind it. `--id` and `--generated-on` set the plan's id and date. With no terminal, `new --out FILE` writes the plan with no question and never over a file, and `new` with no `--out` writes it to standard output. Neither adds it to the plan list. `new --templates DIR` keeps its old behavior for the scenarios that use it, and `--template-set` is refused without `--templates`, because a new plan records the version of the tool as its template set.

### About

A list of every place the tool uses, rendered from data in `internal/command/places`. Code that opens a new place adds one entry there. Each entry shows the path, whether the tool reads it, writes it, or both, what it holds, and for the state home how it was chosen. The places are the state home, the plan list, the sessions folder, the imports folder, the plan folder, the records root, and the Downloads folder, `XDG_DOWNLOAD_DIR` or `~/Downloads`, which reads "Not used yet." From the front door, the plan folder reads "the folder of each plan you open", and the records root is the folder the tool was started in. From the editor's help, both are the open plan's own. Above the places are the version and the template set, and above those any problem: a damaged plan list, or a state home that cannot be found. When the list is longer than the window, its last line says which lines show, such as "Lines 1 to 18 of 32", and the arrows scroll.

### Why these keys

`x` and `y` remove a plan as they remove a contact. `a` is free on the first screen and on every help page, because help closes on any other key. Esc means back everywhere, so on Start, with nothing behind it, it does nothing, and `q` quits. On New plan opened by `obk new` esc quits, because `q` is text there and there is no first screen behind it.

### Rejected on the front door

- Reading each listed plan to show its organization's name. It would decode every plan at launch, and the file name already tells plans apart.
- Opening the editor for a missing plan, or removing it without asking. Enter explains, and `x` asks first, as everywhere else.
- Writing the new plan from the field with enter. The name is typed, so `y` cannot be its key there. The last step keeps `y` the one key that writes, as on the Invitation tab.

## Open

- Whether a team can supply a default manager or onboarding owner for a person, as [0520](../decisions/0520-preparation-screen.md) leaves open. Today each of the four is its own pick.

- Whether the Preparation review should show the guide under the changes below 120 columns, instead of behind tab.

- Whether the line for the newcomer should name only the file, not the onboarder's path. The newcomer saves the file in their own folder, often Downloads, so the onboarder's path may not be theirs.

- Whether Start with no known plans should also offer Open a file. 0610 lists two choices there, so a person whose team already keeps a plan runs `obk FILE` the first time.

- Whether the plan picker should show the organization name from each plan, not the file name.

- Whether the Invitation form should remember its values across a visit to another tab. Today opening the tab checks everyone again and starts at the list.

- Whether the guide beside the Invitation form should scroll to the help it names after a pick, so the effect of the pick is on screen.

- Splitting this page into one file per view under `docs/design/`, with design decision records (DDRs) for choices that cut across views, such as the width breakpoints. DDRs stay separate from ADRs: an ADR is for a decision that contracts or acceptance tests depend on, and a DDR is for a design choice that only the screens depend on.

- A key to move a pick up or down in a several-answers list.

- Importing contacts from a directory service. It is not planned, see 0080.

- A key to move a holder up or down in the Contacts forms.

- Accents in ids. The slug keeps only A to Z and digits, so a name such as "José" makes the id `jos`.

- Whether the progress count should outrank the tab keys at 80 columns now that the strip has five tabs.

- Whether the newcomer's Review tab should be named Document, as 0610 says. It holds more than the document today, see Entry and roles.

- Whether the newcomer's tabs should take their own keys, such as `u` for Your questions, which is free on every screen. Today both people use `w`, `o`, and `r`.

- Whether the Answer screen goes to the next unanswered question when the person is revisiting from Review. Today it goes to the next in order.

- How the tool reminds a person of deferred questions after they quit.

- Whether the rail should show the section titles when the window is wide and the plan is small.

- Whether the notes should hide while a person is typing, to leave only the answer.

- A mouse and a way to change the accent are not planned.

- The newcomer's view of the onboarder's answers, if the document preview proves not enough.
