# Bare command and first run

Status: idea, 2026-10-29.

## Idea

Typing `onboard-kit` with no arguments should do the useful thing for the
current state. Today it prints usage and exits with code 2. The person should
not have to say they want a new plan when the tool can see none exists.

## Shape

The tool looks for a workspace in this order: an explicit option, the
`OBK_HOME` setting, then the plan it opened last. What it finds decides the
first screen.

| Found | First screen |
| -- | -- |
| No plan | "No plan found." Offers to create one from the built-in library, or to open an invitation file. |
| One plan | Opens it directly, with the organization name in the header. No picker. |
| Several plans | A picker with the last-used plan preselected, showing organization, last opened, and progress. Includes "New plan" and "Remove from list". |
| An invitation file as an argument | The onboardee flow. |

- The `new` subcommand stays for scripts and for an explicit create. It never
  overwrites an existing file.
- Creating from a screen shows the path it will write and asks for
  confirmation. A bare run in the wrong folder must not write files silently.
- "Remove from list" forgets a plan and never deletes the file.
- A registry file in the state home lists known plan paths, so plans stay where
  people keep them, for example in a team repository.
- The first-run "open an invitation file" choice lets a newcomer who has the
  binary but does not know the argument rule begin.
- `start` stays the explicit onboardee entry. The bare command is the
  onboarder's.

## Binary names

Ship one file named `onboard-kit` and an alias `obk`, a symlink made by the
Homebrew formula and the Nix package. The state home already uses `obk`. Both
names behave the same. The invitation and docs use the full name. Check `obk`
for a collision with common tools before committing to it.

## Open questions

- Where does a plan live: a registry of paths, or files managed under the state
  home. The registry is the current preference.
- Whether first run is a screen, and how the plan id and file name are
  suggested. [0540](../decisions/0540-adding-and-removing-sections.md) decides
  the suggestion for the plan id.
- What the picker shows for a plan that has moved or was deleted.

## Related

[0530](../decisions/0530-new-plan-from-core-set.md) decides what a new plan
starts with. The first-run screen calls that creation path.
