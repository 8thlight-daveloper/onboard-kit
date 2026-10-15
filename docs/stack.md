# Stack

The current language and tooling. These are implementation choices, cheap to change, so they live here and not in an ADR. Anything the contracts or acceptance tests depend on belongs in an ADR instead.

Every choice must satisfy [0005](decisions/0005-self-contained-install.md): one self-contained file, no secondary installs.

## Current

| Choice | Current | Why | Since | Replace when |
| --- | --- | --- | --- | --- |
| Language | Go | Builds one static binary, which meets 0005 with no extra work. The compiler and `internal/` visibility check regenerated components. | 2026-10-15 | A bundled build in another language passes the 0005 checks and the team maintains it more easily. Maintaining the Go code proves harder than expected. |
| TUI toolkit | Bubble Tea and Huh | Huh covers selects, multi-selects, text input, validation, and step groups. | 2026-10-15 | The walkthrough needs heavy custom rendering. |
| Test runner | `go test -race`, parallel by default | See [0020](decisions/0020-binary-structure-and-regenerative-components.md). | 2026-10-15 | |
| Dependency check | To decide | Needed for the dependency rules in 0020. | | |

## Retired

| Choice | Was | Since | Until | Replaced by | Why it changed |
| --- | --- | --- | --- | --- | --- |
| | | | | | |

## Changing a row

1. Move the old row from Current to Retired. Fill in Until, Replaced by, and Why it changed.
2. Add or edit the Current row with a new Since date.
3. Run the contract and acceptance suites. A stack change that needs a test change is probably an ADR, not a stack note.

Git holds the exact diffs. The Retired table holds the reasons, so nobody has to dig through history to learn why something was replaced.
