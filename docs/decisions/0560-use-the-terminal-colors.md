---
category: ux
status: accepted
approval: person
note: Replaces the brand colors in the Colour section of the terminal walkthrough.
affects: [tui]
---

# 0560. Use the terminal's colors

## Context

The terminal screens draw with three fixed brand colors: a blue accent, an
orange warning, and a red problem, each a 24-bit value with a 256-color and a
16-color fallback. A person's terminal may be dark, light, tinted, or themed,
and we cannot know which. Colors chosen for one background can fail on another.

The contrast ratios below are computed from the three colors against some
common backgrounds. A ratio of 4.5 is the usual bar for normal text.

| Color | Black | Dark gray | Solarized dark | White | Solarized light |
| -- | -- | -- | -- | -- | -- |
| Blue `0071CE` | 4.25 | 3.38 | 3.04 | 4.94 | 4.58 |
| Orange `FF8300` | 8.5 | 6.75 | 6.08 | 2.47 | 2.29 |
| Red `F32735` | 5.15 | 4.09 | 3.68 | 4.08 | 3.78 |

The orange warning is close to unreadable on a light terminal. A test against a
background we picked proves nothing about the one a person uses, and a person
already chose a theme that reads well for them. Brand colors in a terminal tool
add little, since the tool is not a place where a brand needs to show.

## Decision

Draw every screen with the terminal's own colors and none of ours.

- **Palette.** The terminal's default foreground and background, and the sixteen
  indexed colors the person's theme defines. No hexadecimal color and no
  256-color or 24-bit value appears on any screen. The tool never asks the
  terminal for its background color and never computes contrast.
- **Body text** is unpainted, so it takes the default foreground.
- **Quiet text and separators** use the faint attribute, never color 8, which
  some themes draw in the background color.
- **The accent** marks "you are here": the current tab, the `>` marker, the
  position in the progress rail, and the input prompt. It is the terminal's
  green, in bold.
- **A warning** is the terminal's yellow, and a **problem** is the terminal's
  red. Each always carries words, as the Colour section already requires:
  `warning:` and `! problem:`.
- **With `NO_COLOR` set** or on a terminal with no color, the screens are plain
  text. There is no bold or faint either, so the marks, brackets, and words carry
  all the meaning. This is today's behavior, measured in a real terminal.
- **Meaning never rests on color alone.** The marks, brackets, and words stay as
  they are.

There is no setting for the accent in this decision. A later decision can let a
person choose it from the terminal's colors.

## Testing

A test visits every screen under a 24-bit color profile and checks that no
24-bit or 256-color escape appears, so only the default color and indexes 0 to 15
remain. The tests that check the accent, the warning, and the problem expect the
terminal's green, yellow, and red. Contrast is not tested, because the person's
theme decides it.

## Consequences

The screens read as well as the person's theme does and follow it when the
person changes it. The brand blue, orange, and red are gone from the terminal,
and the palette code gets simpler. A theme with an unreadable green or yellow
shows that in this tool, as it does in `git` and `ls`.
