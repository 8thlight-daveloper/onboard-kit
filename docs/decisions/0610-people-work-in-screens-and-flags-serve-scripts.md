---
category: ux
status: proposed
affects: [catalog/creator, catalog/editguide, command/catalog, command/create, command/invite, command/root, command/start, command/stdio, command/walk, invitation/starter, session/editor, tui]
builds-on: ['0320', '0430', '0520', '0530', '0540', '0560']
---

# 0610. People work in screens, and flags serve scripts

Builds on [0320](0320-embedded-catalog-availability.md),
[0430](0430-reviewed-catalog-creation.md),
[0520](0520-preparation-screen.md),
[0530](0530-new-plan-from-core-set.md),
[0540](0540-adding-and-removing-sections.md), and
[0560](0560-use-the-terminal-colors.md). Extends the reasoning of 0520 from
preparation to every flow a person does. Replaces the line in 0530 that says `new`
with `--templates` is unchanged.

## Context

The tool has a terminal interface, and a person still has to leave it to use it.

- **An onboarder runs many commands.** One plan reaches a newcomer through `new`,
  `walk`, `prepare init`, `prepare recipient`, and `invite export`, with flags for
  each. A scenario that does it end to end needs six commands.
- **A newcomer types a protocol.** `start --invitation FILE --root DIR` needs two
  flags, and then `confirm newcomer`, `begin`, and `exit` typed as lines.
- **Text conversations remain.** `catalog create`, `catalog edit`, and
  `session edit` ask for typed lines. 0520 already retired this style for
  preparation, because the screens it ran beside gave better pick lists, layout,
  and wording.
- **`walk` names one tab.** The onboarder does more than walk through questions.
  They answer, switch sections, manage contacts, and review the document.
- **`--templates` builds plans the tool then rejects.** The five example templates
  give a plan that `new` and the walk accept, and `start` refuses with two
  authoring errors.

A person should launch the tool once and make every choice inside it. Otherwise
the screens add little.

## Decision

**One entry, and flags serve scripts.** `obk`, an alias that ships beside
`onboard-kit`, opens the terminal interface. A person makes every choice there.
Flags and arguments exist for scripts and agents.

**The file decides what opens.** `obk FILE` looks at the kind of file. A plan
opens the plan editor. An invitation starts the newcomer's flow. A file of any
other kind is refused with one plain sentence that names what it expected.
`obk open FILE` is the explicit spelling for docs and scripts. It never asks
anything, and it exits with an error on a file of the wrong kind. No command
carries a role word, and no verb is needed for the newcomer.

**The front door.** `obk` with no argument shows a first screen:

- With no saved plan, it asks "What would you like to do?" and offers "Build an
  onboarding plan" and "Begin my onboarding".
- With known plans, it shows a picker with the last-used plan preselected, beside
  "New plan", "Open a file", and "Begin my onboarding".
- A registry file in the state home lists known plan paths. Plans stay where their
  team keeps them. "Remove from list" forgets a plan and never deletes the file. A
  moved or deleted plan shows as missing.

**Creating a plan.** `obk new` stages the core set of 0530, shows the path it will
write, asks for confirmation, writes the file, and opens the editor. It never
overwrites. `--templates` and `--template-set` are removed. The template set a plan
records is the build's version, as [0030](0030-self-contained-plan-file.md) says.
Custom template folders are not supported, and a later decision must run the
authoring validator before it writes a plan.

**Screens name the role in the header.**

- The onboarder sees "Building: Acme plan". Tabs are Answer, Overview, Review,
  Sections, Contacts, Preparation, and Invitation. The word "walk" retires for
  this editor.
- The newcomer sees "Your onboarding at Acme". Tabs are Your questions, Overview,
  and Document. There are no Sections or Contacts, as today.
- The Invitation tab writes the file, then shows its path and the exact line to
  give the newcomer: `obk` and the path. The tool sends nothing.
- The newcomer's flow is one session. An identity screen reads "Is this you?"
  with name, team, and onboarding owner, and then a short "what to expect", and
  then their own questions. Running it again says "Welcome back" and resumes.

**Tell the person before touching a place they did not name.** Before the tool
looks anywhere beyond the file in the command, it says what it will look at and
why. If the system may ask permission, it says so, and names what the prompt will
read. Looking for an invitation is the first case:

- After "Begin my onboarding", a screen says it will look in the Downloads folder
  and shows the path. It adds that the system may ask whether the terminal can
  access Downloads, and that Allow lets the tool find the file. The choices are
  "Look in Downloads" and "Type a path instead".
- The Downloads folder is the `XDG_DOWNLOAD_DIR` setting, or `~/Downloads`. The
  tool reads only the top level, and only file names that start `obk-invitation`.
  The Invitation tab proposes that name, and browsers may add a number.
- It lists the candidates the registry does not record as loaded, with their
  dates. "Not now" and "Never for this file" are remembered in the state home. It
  opens an archive only after the person says yes, as
  [0390](0390-invitation-resource-admission.md) requires.
- Scanning does not run at launch. A first launch should not meet a system prompt
  before the person knows why.

**An About screen lists every place the tool uses.** It shows the state home and
how it was chosen (`OBK_HOME`, `XDG_STATE_HOME`, or the default), the plan
registry, the pending-import and session folders, the records root, and the
Downloads folder with a note that it is read only when the person chooses it.
Each line says what the place holds and whether the tool reads it, writes it, or
both. It also shows the version and the template set. The front door and the help
page reach it. Code that opens a new place adds it to the list the screen renders.

**Scripts and agents keep their own surface.** `new --out FILE`, `assemble`,
`catalog list|show|export`, `prepare`, `invite`, and `stdio` stay, listed under
"for scripts" in `--help`. With no terminal, an interactive entry refuses and names
the scripting command to use. It never falls back to typed lines.

**Text conversations retire as a human interface.** `catalog create`,
`catalog edit`, `session edit`, `start` with its typed commands, and `walk` become
screens. The operations behind them stay, and the screens call them, so a flag and
a screen cannot drift. Which scripted forms survive is decided command by command
when each screen is built.

**Structure.** Screens, flags, and `stdio` are fronts over shared operations. No
front imports another front, and an operation imports no terminal toolkit. These
become rules in `architecture.yaml` when the operations layer exists.

## Consequences

- Building this is a long run of changes. It starts with the Preparation and
  Invitation tabs of 0520, then `new` as a Sections-screen creation, then the
  newcomer's identity screen, then the front door, the registry, and the About
  screen.
- About ten scenarios that pass `--templates` change, and so do the scenarios of
  the typed protocols: ten for `start`, five for `invite export`, and four for
  `catalog`. Typed ones are replaced by terminal scenarios or kept as tests of the
  scripting surface. Scenarios live in `acceptance/`, so a person authorizes each
  change, and retiring a contract is a person's decision.
- The five example templates move out of `docs/reference/examples` into fixture
  data under `acceptance/testdata`, and their two authoring errors get fixed in
  the move. The source path that about a dozen contract fixtures cite changes with
  them.
- The README quick start and command table change to match.
- `obk` needs a check for a clash with a common tool before the name ships.
- On macOS the Downloads prompt appears the first time a terminal reads that
  folder. The notice before scanning is how a person learns to expect it.

## Not decided

1. **Scripted forms of retired commands.** `catalog edit` adds, retires, and
   reactivates catalog sections in a plan. `session edit` changes a newcomer's own
   record. Both take typed lines today, and screens will replace them for people.
   `stdio` already lets an agent answer questions, toggle sections, edit the
   directory, and save a plan. It cannot add or retire a catalog section or edit a
   newcomer's record. The question is whether a script or an agent still needs
   either operation, and if so whether it belongs in `stdio` or as a flag command.
