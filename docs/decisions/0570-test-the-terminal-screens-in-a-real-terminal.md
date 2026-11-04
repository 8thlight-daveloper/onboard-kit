---
category: operations
status: accepted
approval: person
builds-on: ['0520']
---

# 0570. Test the terminal screens in a real terminal

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
exit code that told a script nothing for two signals. A second project has used
the same tool on macOS and on a Linux CI runner. Their reports are the source for
the tool and the exit rules below. Nothing here has been run against this tool
yet.

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

**Tool.** Use the tui-test Go binding, pinned to one released version (0.1.1 in
the reports) and upgraded on purpose. It is a test-only dependency of about 54 MB,
so the shipped binary does not change and
[0005](0005-self-contained-install.md) still holds. It is a 0.x tool with one
emulator, so a failure that only one emulator shows is checked by hand before it
is trusted. If the tool is dropped, only the runner's terminal step goes with it.
The first step of the work is a trial run against `walk`: launch it at a fixed
size, wait for text, quit, and assert the exit code and that the alternate screen
is off. The trial also records how this tool exits today.

What users of the tool report, to be confirmed in the trial:

- **Keys, text, resize, and waiting** work. A wait times out after 5 seconds by
  default, and its failure message does not show the screen, so the runner adds
  the screen text to its own failure output.
- **Cell colors can be read.** Each cell has its character, foreground and
  background (`default`, an index such as `2`, or a hex value), and bold, faint,
  reverse, italic, and underline. One report saw the process crash when reading
  cells repeatedly (a garbage collector problem in the binding), and avoided it by
  forcing a collection and turning the collector off around each call. The runner
  must do the same, or read cells once per screen.
- **Terminal modes can be read** after exit: alternate screen, cursor visibility,
  and bracketed paste.
- **The exit code** is read by polling the terminal state. The tool's own exit
  code assertion failed in one report.
- **Environment and window size** are set when the child starts. The color
  profile follows `NO_COLOR` and `COLORTERM`. Setting `TERM` was not tried. The
  emulator's theme, meaning its foreground, background, and sixteen colors, can
  be set, so one scenario can run under a dark and a light theme.
- **Signals through the tool are unreliable.** In one report `INT` acted like a
  Ctrl+C key press, `TERM` arrived as a hangup, and `HUP` was unsupported. See
  the exit rules.
- **Terminal queries** are answered by the emulator from its theme, and this tool
  sends none. A test should wait for a specific state before it asserts.
- **A launched program** may need the tool's ready-wait turned off.
- **Screenshots** carry a window frame, which a test crops if it keeps them.
- **Speed:** 0.2 to 0.9 seconds a test. One acceptance package of this kind took
  about 26 seconds on CI.
- **Not tried:** cleanup of child processes by the tool, traces and recordings,
  where the engine is cached, and the runner's glibc version.

**Environment.** Each terminal step fixes the window size and sets `TERM`,
`COLORTERM` and `NO_COLOR` explicitly, because the color profile comes from
them. It uses a temporary home and `OBK_HOME`, and it needs no network. A step
waits for text and never for a fixed time. Traces and failure artifacts go under
`output/`, are never committed, and are reviewed before anyone uploads them,
because they can hold screen text.

**Trial results.** A trial program ran `walk` in a pseudo-terminal with the tool,
on macOS. It confirmed the reports above and measured how this tool behaves today.

- **Screen, keys, and modes work.** The screen text read back correctly. `r`, `s`,
  `c`, and `o` moved between tabs. While running, the alternate screen is on, the
  cursor is hidden, and bracketed paste is on. After `q` or `Ctrl+C` the exit
  code is 0 and all three modes are restored.
- **Colors match 0560.** Reading every cell of the Review screen found only the
  default color and index 2 (green), with the default background. Bold was on
  for 109 cells and faint for 103. With `NO_COLOR` set, every cell is the default
  color with no bold or faint.
- **`TERM` works.** `TERM=dumb` and `TERM=vt100` give plain text with no color
  and no styling. `TERM=xterm-256color` gives the green accent and the styling.
- **Real signals** sent to the child process:

| Signal | Exit code today | Terminal restored | Target |
| -- | -- | -- | -- |
| `SIGINT` | 1 | yes | 130 |
| `SIGTERM` | 0 | yes | 143 |
| `SIGHUP` | killed by the signal, so 129 to a shell | not applicable | 129 |

The two gaps are the exit codes for `SIGINT` and `SIGTERM`. `SIGTERM` exits 0, so
a script cannot tell a stop request from a normal exit. `SIGHUP` already ends
the process with the right status. These runs made no edits, so they did not
test a signal in the middle of a save.

- **Signals sent through the tool** again did not match real ones: `INT` exited
  0, `TERM` exited 1, and `HUP` was unsupported.
- **The cell-reading crash is real and has a simple fix.** 200 repeated reads of
  the screen crashed the test process with `invalid pointer found on stack`.
  Running the tests with `GODEBUG=gcshrinkstackoff=1` made the same 200 reads
  pass, and so did forcing a collection around each read. The setting is the
  simpler fix, since it is one environment variable on the test command.

**Signals come from the runner, not the tool.** A `send signal` action is carried
out by the runner, which starts the child itself, keeps its process id, and sends
a real signal with the operating system. Signals sent through the terminal tool
do not match real ones. Reading the exit code and the terminal modes still uses
the tool.

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
- A signal says nothing reliable about intent, so the rule follows consent, not
  the signal. Before the person confirms a save, a stop request discards the
  draft and leaves the file alone. After confirmation, the save finishes, because
  the atomic replacement takes microseconds and the person asked for it, and then
  the tool exits 130 or 143. If a stop request arrives after the temporary file
  exists and before the replacement, the tool removes that temporary file and
  keeps the original. It never rolls back a completed replacement.
- The guarantee is atomic replacement, not a lock against signals. A signal at
  any instant leaves the original file or the complete reviewed one, never a
  mix, and `SIGKILL` gets the same guarantee. A scenario checks that outcome
  and does not try to hit the replacement window by timing. The only leftover
  allowed is an owned temporary file after `SIGKILL`.
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
  [0560](0560-use-the-terminal-colors.md) that every cell is the default color
  or an index from 0 to 15, run under a dark and a light emulator theme. The
  in-process test of the same rule stays, because it is faster.

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
2. **Running on the macOS runners** in addition to Linux. The first target is
   macOS, and CI runs on Linux.
3. **Whether the crash setting can live in the repository,** so the test command
   does not depend on someone remembering it, or must be set by the runner.

A signal in the middle of a save was open in the first draft. The exit rules now
settle it. The trial did not test that no file is left half written. The first
scenarios do.
