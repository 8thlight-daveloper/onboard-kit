# 0570. Test the terminal screens in a real terminal

Status: proposed. A person has not yet approved it.

Builds on [0520](0520-preparation-screen.md), which promises a scripted
pseudo-terminal run for the Preparation tab. It extends the acceptance
scenarios described in the [acceptance scenarios reference](../reference/acceptance-scenarios.md).

## Context

The terminal screens are tested in process. A test sends key and window-size
messages to the Bubble Tea model and compares the string its view returns with
golden text. The acceptance scenarios drive the built binary, but only through
arguments, standard input and standard output, so they cannot reach a screen.

Nothing runs a screen in a terminal. Nothing checks how real key presses decode,
whether the alternate screen is left on exit, what a closed terminal window does
to a draft, or whether a signal leaves a half-written file. The tool will soon
hold more state on screen, such as the Preparation tab's unsaved changes and a
draft plan, so these gaps matter more.

A sibling project made the same decision for its own screens and found two real
gaps when it measured them: an agent process left running after a hangup, and an
exit code that told a script nothing for two signals. Its report is the source
for the tool and the exit rules below. Nothing here has been run against this
tool yet.

## Decision

Add a layer of terminal scenarios that run the real binary in a pseudo-terminal
behind a terminal emulator, and state the exit rules the scenarios check.

**What the layer is for.** The model tests stay the fast and exact layer. A
terminal scenario covers only what a model test cannot see: key decoding, the
rendered screen, the alternate screen, resizing, process exit, signals, and the
files left behind. A check that a model test can make stays a model test.

**Scenarios are data.** A terminal scenario is a file under
`acceptance/scenarios/`, like the others. It adds a third kind of step beside the
command step and the stdio step: a terminal step. A terminal step names a
terminal size, the environment, a list of actions (send keys, send a signal, wait
for text, resize), and what must be true afterward: the exit code, text on the
screen, the terminal modes after exit, and the files. The runner does the
driving. No scenario imports an implementation package. The exact fields go in
the acceptance scenarios reference when the runner is built.

**Tool.** Use the tui-test Go binding, pinned to one released version and
upgraded on purpose. It is a test-only dependency, so the shipped binary does not
change and [0005](0005-self-contained-install.md) still holds. It is a 0.x tool
with one emulator, so a failure that only one emulator shows is checked by hand
before it is trusted. If the tool is dropped, only the runner's terminal step
goes with it. The first step of the work is a trial run of the tool against
`walk`, to confirm what the sibling project reported and to record how this tool
exits today.

**Environment.** Each terminal step fixes the window size and sets `TERM`,
`COLORTERM` and `NO_COLOR` explicitly, because the color profile comes from
them. It uses a temporary home and `OBK_HOME`, and it needs no network. A step
waits for text and never for a fixed time. Traces and failure artifacts go under
`output/`, are never committed, and are reviewed before anyone uploads them,
because they can hold screen text.

**Exit rules.** The tool exits cleanly on every way a person or the system can
end it. These are the targets the scenarios check. What the tool does today is
measured in the trial run, and each gap is its own small change.

- `q` and `Ctrl+C` behave as the [terminal walkthrough](../reference/tui.md)
  says: `q` quits except while typing, and `Ctrl+C` always quits.
- `SIGINT` and `SIGTERM` restore the terminal and exit 130 and 143, the shell
  convention of 128 plus the signal number.
- `SIGHUP`, sent when a terminal window closes, saves what is saved by an
  explicit action and nothing else, and exits 129. No restore is needed because
  the terminal is gone.
- No exit leaves a half-written file. The plan, preparation and record writes
  are already atomic, and a scenario checks it after a signal.
- Unsaved edits are lost on exit by design, as the screens say. A scenario does
  not expect them to survive.
- `SIGQUIT` keeps its default Go behavior and is out of scope.

After each exit a scenario starts the same session again and checks that it
resumes from the saved files. The restore check reads the terminal modes: the
alternate screen is off, the cursor is visible, and bracketed paste is off.

**First scenarios.** One file covers these first, and a scenario is added when a
bug in this layer is fixed:

- the walk: start, answer a question, defer one, and quit;
- the tab keys, and the Sections switch saving a change;
- the onboardee mode, which has no Sections or Preparation tab;
- the Preparation tab, once it exists: add a person, pick a team, review, and
  the explicit save key, with nothing written before it;
- a narrow window and a resize while running;
- `NO_COLOR` and a forced color profile, and the rule of
  [0560](0560-use-the-terminal-colors.md) that no screen uses a fixed color,
  if the tool can read the colors of the cells. The tool's documentation was not
  checked for that. If it cannot, the rule stays tested in process.

**Out of scope.** Tests against a real model or network, whole-screen baseline
comparison across emulators, and Windows. Windows runs the tool under WSL, which
is Linux.

**CI.** The scenarios run in `just check` and in the `check` workflow. The
native engine needs a recent glibc, and the first run on CI confirms the runner
has one.

## Consequences

- The screens get a test at the layer people touch, and the exit behavior becomes
  a stated rule with tests. The promise in 0520 is met.
- The cost is a large test-only module download, a native engine extracted to
  the user cache, and an upgrade chore while the tool is below 1.0.
- Changing the acceptance runner and the scenario reference is a change to
  `acceptance/` and needs a person to authorize it. Accepting this decision
  authorizes the terminal step and nothing else in `acceptance/`.
- Likely code changes follow from the exit rules, such as a handler for
  `SIGHUP` and exit codes for `SIGINT` and `SIGTERM`. Each is its own small
  change with its own scenario.

## Not decided

1. **The scenario fields** for a terminal step. They are defined when the runner
   is built and recorded in the acceptance scenarios reference.
2. **Whether the tool can read cell colors.** Checked in the trial run.
3. **Running on the macOS runners** in addition to Linux. The first target is
   macOS, and CI runs on Linux.
