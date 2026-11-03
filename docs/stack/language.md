# Language and toolkit

Part of the [stack](README.md). Every choice must satisfy [0005](../decisions/0005-self-contained-install.md).

## Current

| Choice | Current | Why | Since | Replace when |
| -- | -- | -- | -- | -- |
| Language | Go | Builds one static binary, which meets 0005 with no extra work. The compiler and `internal/` visibility check regenerated components. | 2026-10-15 | A bundled build in another language passes the 0005 checks and the team maintains it more easily. Maintaining the Go code proves harder than expected. |
| TUI toolkit | Bubble Tea and Huh | Huh covers selects, multi-selects, text input, validation, and step groups. | 2026-10-15 | The walkthrough needs heavy custom rendering, or structured records need a pick-or-add flow that Huh cannot compose. |

## Retired

| Choice | Was | Since | Until | Replaced by | Why it changed |
| -- | -- | -- | -- | -- | -- |

## Considered and not chosen

Options we weighed and why they are not the current choice. Review these with the context of what we knew at the time. Evidence is dated, because it ages.

| Choice | Option | Status | Why | Revisit when |
| -- | -- | -- | -- | -- |
| Language | Python | Not chosen | The org has more Python repos than Go or Rust: 14 against 2 and 2 in a `gh repo list` count on 2026-10-15. The Go repos were last pushed in 2022 and 2018. It meets 0005 only through a bundled build, whose size, per-OS builds, and antivirus behavior we have not tested. No compiler checks regenerated components, so boundaries need a type checker and import linter to match | A bundled build passes the 0005 checks, and the team prefers maintaining it |
| Language | Rust | Not chosen | Viable, and fits dense live screens. Ratatui has no built-in form widgets. Team familiarity is no better than Go. Source: [issue 2](https://github.com/8thlight/onboard-kit/issues/2) | The walkthrough grows into screens that need heavy custom rendering |
| TUI toolkit | Ratatui | Not chosen | Follows from the Rust option, and has no built-in form widgets | Same as Rust |
