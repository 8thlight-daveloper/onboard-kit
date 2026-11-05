# The plan list

The plan list is the plan registry of [0610](../decisions/0610-people-work-in-screens-and-flags-serve-scripts.md). It records the plans a person opened or created at a terminal, so the first screen can offer them again, last used first. Plans stay where their team keeps them. The list holds only their paths. The code is `internal/command/registry`.

## Where it is

The file is `plans.txt` in the state home. The state home is chosen in this order:

1. `OBK_HOME`, when it is set and not empty.
2. On Linux, `XDG_STATE_HOME` followed by `obk`, when it is set and not empty.
3. The default: `~/Library/Application Support/obk` on macOS, and `~/.local/state/obk` on Linux.

`start --invitation` chooses its home the same way, so the list sits beside the `sessions` and `import-pending` folders. The About screen shows the state home, how it was chosen, and the list's path.

## Format

A plain UTF-8 text file. The first line is exactly:

```text
onboard-kit plans 1
```

Each later line is one plan: the time it was last opened, a tab, and its full path. The time is RFC 3339 in UTC, to the second. In the example, `<TAB>` stands for one tab character.

```text
onboard-kit plans 1
2026-11-04T10:15:00Z<TAB>/Users/sam/work/acme/acme.obk.yaml
2026-11-02T08:40:12Z<TAB>/Users/sam/work/globex/platform-team.obk.yaml
```

- The tool writes the lines newest first. A reader sorts them by time anyway.
- A path is absolute, at most 4096 bytes, and holds no tab, newline, or other control character. A path that breaks this rule is not recorded.
- A path appears once. Opening a listed plan moves it to the top with the new time.
- Blank lines are skipped. A missing newline at the end is accepted.
- The list keeps the newest 100 plans and drops older ones.

## When it changes

| Event | Change |
| -- | -- |
| A plan opens in the editor: `onboard-kit FILE`, `onboard-kit open FILE`, the first screen, or `walk --plan FILE` | The plan's full path moves to the top with the current time |
| `onboard-kit new` writes a plan at a terminal | The new plan is added at the top |
| Remove from list on the first screen | The plan's line is removed. The plan file is never touched |
| `new --out FILE` or `new` with no terminal | Nothing. Scripts and agents do not fill a person's list |

The tool writes the whole file again each time: a temporary file in the state home, then a rename over `plans.txt`. A reader never sees half a list. The state home is created with mode `0700` when it does not exist.

## A damaged list

The tool reads at most 256 KiB. A list is damaged when it is larger, is not UTF-8, has a different first line, holds a line that is not a time, a tab, and a full path, names a plan twice, or holds more than 100 plans.

A damaged list is never changed and never stops the tool. The first screen shows no plans and says "The list of your plans cannot be read. Press a to see why." The About screen names the file and the reason. Plans opened while the list is damaged are not added. Removing the file starts a new list.

## Missing plans

The first screen checks each listed path when it opens. A path that no longer exists shows as missing. Enter on it says so, and Remove from list forgets it.
