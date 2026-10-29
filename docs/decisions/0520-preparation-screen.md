# 0520. Preparation as a full-screen flow

Status: accepted by a person. It supersedes the interaction style of
[0340](0340-guided-preparation-save.md) and
[0350](0350-guided-support-arrangements.md). Their save rules stay.

## Context

[0340](0340-guided-preparation-save.md) chose a plain text conversation for
guided preparation. It said a full-screen framework was unnecessary for the
first useful interaction, and gave ordinary keyboards, screen readers, and
independent scripted acceptance scenarios as reasons.

The plan walk, Review, Sections, and Contacts are now full-screen terminal
screens. The [terminal walkthrough](../reference/tui.md) promises keyboard-only
use and plain words, and makes no screen-reader claim that the text flow would
lose. Of the three reasons in 0340, only scripted acceptance still separates
preparation from the other onboarder screens.

A manual run of `just demo-prepare` showed the cost. The not-saved notice, the
privacy notices, the input limit, the account and destination, and the prompt
print as one block of lines with nothing to tell them apart. The exact text is
pinned by the guide contract, so a clearer layout was a contract change either
way.

The same run showed more than layout. A person cannot type a name at the field
prompt: they must first choose `change`, and a typed name is answered with
"Choose one of the displayed actions." The team, manager, owner, and buddy are
typed as raw contact IDs, though the plan's directory already lists every valid
choice and the Contacts screen already offers pick lists for them. Before the
prompt, the tool also prints eleven file paths that matter only to someone
testing the tool.

## Decision

Proposed: preparation moves into the onboarder's existing terminal screens as
a Preparation tab, not a separate full-screen program. It reuses the palette,
frame, pane, and pick-list code of `internal/tui`.

- The tab lists the people who have a preparation under the root, and a way to
  add a person.
- A person's form picks the team, manager, owner, and buddy from the plan's
  directory, as the walk does for contact questions. Typing into a field needs
  no extra step first.
- Notices sit in their own area, and the account and destination stay visible.
- Review shows the real guide beside the changes. Save needs an explicit key,
  and nothing is written before it.
- The screen needs to know the preparation root, so the walk gains a `--root`
  option. The onboardee mode does not get the tab.

The save path is the real cost. The bounded reads, safe-path checks, byte
recheck, and atomic publish of 0340 now live in the guide command's own package.
The screen must reuse them, so they would move to a component both can call.
Nothing about saving changes. Staging several changes, one reviewed save, no
partial save, and the byte recheck before publication of 0340 and 0350 all stay. The screen is another way to supply the same staged changes.

There is no scripted path for guided preparation. Once the screen ships, the
text conversation of `prepare guide` is retired, along with its contract and
acceptance scenarios, which are already removed. The one-operation `prepare` commands that 0340 left in
place stay as they are. This decision does not add a headless mode, a plain
mode, or a batch input. A later decision can add one if a real need appears.

## Further decisions

1. **Retiring the old guide.** Done. A person authorized removing the text
   guide's contract, scenarios, implementation, and command wiring. The save path
   it held (bounded reads, safe-path checks, byte recheck, atomic publish) is in
   git history at commit `4a5a7d2`, `internal/preparation/guide/files.go`. The
   screen's save component starts from it. Until the screen ships, the explicit
   `prepare recipient`, `arrangement`, and `remove` commands are the way to
   write a preparation.

2. **The input bound.** Drop the per-run input cap. It limited how much text the
   tool would read from piped input, so a script could not make it consume
   unbounded memory. A screen reads one key at a time, and only a paste could
   come near any limit. Keep the bound that protects the file: the encoded
   preparation still may not exceed 1 MiB, and each field keeps its own length
   limit.

3. **Dependencies.** The rule in
   [0020](0020-binary-structure-and-regenerative-components.md) is about the
   screen, not the retired command. The `tui` row lists only the finder, the plan
   editor, the assembler, the plan codecs, and the model. A preparation screen
   needs the preparation decoder, encoder, editor, reviewer, and locator, so the
   row grows by those. File access stays out of `tui`: the walk command supplies
   it to the screen, as it supplies the plan save today.

4. **Scope of the first increment.** Both recipient details and arrangements,
   as the text guide did, so no capability is lost.

5. **Screen readers.** The tool does not support them and has never claimed to.
   No document makes the claim, and the full-screen screens were never built
   for it. This decision does not change that. A later decision can add support.

6. **Open: who sets the manager and owner.** The form picks the team,
   manager, owner, and buddy from the directory. Some organizations assign a
   manager per team. Others, such as consultancies, give the team an account
   director and give each person a separate personal manager. The first
   increment treats each of the four as its own pick for the person. A later
   decision can say whether a team can supply a default for the manager or owner.

## Acceptance boundary

Screens are driven by key and window-size messages and compared with golden
text, as the other screens are. A scripted pseudo-terminal run covers start and
quit, as for the walk. The moved save component keeps its own contract and
byte-exact tests, so refusals, cancel, and a failed publication leave the files
as 0340 and 0350 require. These checks no longer run the screen through the
built binary with scripted input, so the binary-level scenarios for guided
preparation end with the text guide.
