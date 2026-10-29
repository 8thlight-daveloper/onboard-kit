# Dependency check

Part of the [stack](README.md). The rules are data in `architecture.yaml`, as decided in [0020](../decisions/0020-binary-structure-and-regenerative-components.md). This file holds the extractor that reads the code, and the packages each banned category maps to.

## Current

| Choice | Current | Why | Since | Replace when |
| -- | -- | -- | -- | -- |
| Dependency check | Rules in `architecture.yaml`. Extractor to decide, spiking ArchUnitGo | Needed for the dependency rules in 0020. The rules are data and the extractor is an adapter. |  |  |

## Banned categories in core components

Core components may not use these categories, see [0015](../decisions/0015-functional-core-and-imperative-shell.md). The packages below are candidates and not final. Deciding the list is part of the spike.

| Category | Candidate Go packages |
| -- | -- |
| Filesystem | `os`, `io/fs`, `io/ioutil`, `path/filepath` |
| Network | `net`, `net/http` |
| Process | `os/exec`, `os/signal`, `syscall` |
| Terminal | `golang.org/x/term`, the TUI toolkit |
| Randomness | `math/rand`, `crypto/rand` |
| Clock | `time`, restricted to its clock calls. The package also holds types the model may need, so an import check alone is too coarse. The extractor may have to see calls |
| Environment | `os` |

## Retired

| Choice | Was | Since | Until | Replaced by | Why it changed |
| -- | -- | -- | -- | -- | -- |

## Considered and not chosen

| Choice | Option | Status | Why | Revisit when |
| -- | -- | -- | -- | -- |
| Dependency check | `depguard` in golangci-lint | Not chosen | One rule per package duplicates the dependency table, and it does not express the `wire_test.go` exception or the folder checks. Not verified in depth | Rules can be generated from `architecture.yaml` |
| Dependency check | `go-arch-lint`, `arch-go` | Not chosen | Not evaluated in depth, and not verified for maintenance | A spike of ArchUnitGo fails |
| Dependency check | ArchUnitGo | Evaluating | Rules are Go tests, folder globs fit our components, and it detects cycles. It is a pre-1.0 preview (v0.0.1, API may move) with one main maintainer, and 3 stars on 2026-10-15. MIT license. Standard library import bans are undocumented, so we spike them | It stays pre-1.0 and the API moves, or it cannot express the dependency rule in 0020 |
| Dependency check | Tree-sitter extractor | Not chosen now | One language-neutral extractor, with a small query file per language. It is syntactic, so it needs import-to-component rules per language and cannot see build tags or aliases. We have one language and the language's own tooling is exact | The language changes, or a second language joins |
| Dependency check | Custom architecture test only | Fallback | Covers every rule, including folder naming, at the cost of code we own | The spike succeeds |
