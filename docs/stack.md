# Stack

The current language and tooling. These are implementation choices, cheap to change, so they live here and not in an ADR. Anything the contracts or acceptance tests depend on belongs in an ADR instead.

Every choice must satisfy [0005](decisions/0005-self-contained-install.md): one self-contained file, no secondary installs.

## Current

| Choice | Current | Why | Since | Replace when |
| --- | --- | --- | --- | --- |
| Language | Go | Builds one static binary, which meets 0005 with no extra work. The compiler and `internal/` visibility check regenerated components. | 2026-10-15 | A bundled build in another language passes the 0005 checks and the team maintains it more easily. Maintaining the Go code proves harder than expected. |
| TUI toolkit | Bubble Tea and Huh | Huh covers selects, multi-selects, text input, validation, and step groups. | 2026-10-15 | The walkthrough needs heavy custom rendering. |
| Test runner | `go test -race`, parallel by default | See [0020](decisions/0020-binary-structure-and-regenerative-components.md). | 2026-10-15 | |
| Acceptance runner | A Go program that reads the scenarios and drives the binary | It must stay replaceable, see [0020](decisions/0020-binary-structure-and-regenerative-components.md). Scenarios are data, and the runner is the only language-specific part. | 2026-10-15 | The implementation language changes, or a neutral runner is worth having. |
| Dependency check | Rules in `architecture.yaml`. Extractor to decide, spiking ArchUnitGo | Needed for the dependency rules in 0020. The rules are data and the extractor is an adapter. | | |
| Mutation tool | To decide | Needed for the contract-strength check in 0020. | | |

## Retired

| Choice | Was | Since | Until | Replaced by | Why it changed |
| --- | --- | --- | --- | --- | --- |
| | | | | | |

## Considered and not chosen

Options we weighed and why they are not the current choice. Review these with the context of what we knew at the time. Evidence is dated, because it ages.

| Choice | Option | Status | Why | Revisit when |
| --- | --- | --- | --- | --- |
| Language | Python | Not chosen | The org has more Python repos than Go or Rust: 14 against 2 and 2 in a `gh repo list` count on 2026-10-15. The Go repos were last pushed in 2022 and 2018. It meets 0005 only through a bundled build, whose size, per-OS builds, and antivirus behavior we have not tested. No compiler checks regenerated components, so boundaries need a type checker and import linter to match | A bundled build passes the 0005 checks, and the team prefers maintaining it |
| Language | Rust | Not chosen | Viable, and fits dense live screens. Ratatui has no built-in form widgets. Team familiarity is no better than Go. Source: issue #2 | The walkthrough grows into screens that need heavy custom rendering |
| TUI toolkit | Ratatui | Not chosen | Follows from the Rust option, and has no built-in form widgets | Same as Rust |
| Dependency check | `depguard` in golangci-lint | Not chosen | One rule per package duplicates the dependency table, and it does not express the `wire_test.go` exception or the folder checks. Not verified in depth | Rules can be generated from `architecture.yaml` |
| Dependency check | `go-arch-lint`, `arch-go` | Not chosen | Not evaluated in depth, and not verified for maintenance | A spike of ArchUnitGo fails |
| Dependency check | ArchUnitGo | Evaluating | Rules are Go tests, folder globs fit our components, and it detects cycles. It is a pre-1.0 preview (v0.0.1, API may move) with one main maintainer, and 3 stars on 2026-10-15. MIT license. Standard library import bans are undocumented, so we spike them | It stays pre-1.0 and the API moves, or it cannot express rule 3 in 0020 |
| Dependency check | Tree-sitter extractor | Not chosen now | One language-neutral extractor, with a small query file per language. It is syntactic, so it needs import-to-component rules per language and cannot see build tags or aliases. We have one language and the language's own tooling is exact | The language changes, or a second language joins |
| Dependency check | Custom architecture test only | Fallback | Covers every rule, including folder naming, at the cost of code we own | The spike succeeds |

## Changing a row

0. Record the options you weighed in Considered and not chosen, with the date and the evidence.
1. Move the old row from Current to Retired. Fill in Until, Replaced by, and Why it changed.
2. Add or edit the Current row with a new Since date.
3. Run the contract and acceptance suites. A stack change that needs a test change is probably an ADR, not a stack note.

Git holds the exact diffs. The Retired table holds the reasons, so nobody has to dig through history to learn why something was replaced.
