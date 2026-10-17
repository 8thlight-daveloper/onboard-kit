# YAML library

Part of the [stack](README.md). Every choice must satisfy [0005](../decisions/0005-self-contained-install.md). The plan decoder reads plan files with it, see the [decoder contract](../../contracts/plan/decoder/contract.md). The library decides how strictly a file is read, so it is recorded here and not left to the build.

## Current

| Choice | Current | Why | Since | Replace when |
| -- | -- | -- | -- | -- |
| YAML library | `go.yaml.in/yaml/v3`, pure Go | Its node API shows what the contract needs and a plain decode hides: plain or quoted style, anchors, aliases, duplicate keys, and line positions. The YAML organization maintains it. A reference decoder written from the contract passed every contract test with this API. Limits found: a plain scalar ending in `?` inside `{}` or `[]` is a parse error, so the contract lists it under Not covered. UTF-16 with a byte order mark decodes without an error, so the decoder checks that the bytes are UTF-8 first. It tags `01`, `0x1F`, and `1_000` as numbers, so the decoder applies the contract's number rules to the text | 2026-10-17 | A library reads a plain `?` in flow style and keeps the node details, or the node API proves too coarse for a contract rule |

## Retired

| Choice | Was | Since | Until | Replaced by | Why it changed |
| -- | -- | -- | -- | -- | -- |

## Considered and not chosen

| Choice | Option | Status | Why | Revisit when |
| -- | -- | -- | -- | -- |
| YAML library | `gopkg.in/yaml.v3` | Not chosen | The same code and API. Its author labeled the repository unmaintained in April 2025, and the YAML organization took over as `go.yaml.in/yaml/v3`. Evidence from the maintained fork's README, read on 2026-10-17 | Never, unless the fork is abandoned |
| YAML library | `github.com/goccy/go-yaml` | Not evaluated | Named as an alternative only. We did not test whether it shows quoting style, anchors, and duplicate keys, or how it reads a plain `?` in flow style | The current choice fails a contract rule that this one could meet |
