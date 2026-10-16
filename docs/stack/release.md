# Release

Part of the [stack](README.md). The constraints and channel order are in [0005](../decisions/0005-self-contained-install.md). This file holds the tools that build and publish a release.

## Current

| Choice | Current | Why | Since | Replace when |
| --- | --- | --- | --- | --- |
| Release build | A static build for `darwin/arm64`, verified in CI as 0005 describes | macOS arm64 is the first platform. | 2026-10-16 | Another platform is supported. |
| Release tool | To decide. GoReleaser is a candidate, not yet evaluated | One tool could build the release and write the Homebrew and Nix channel files. | | |
| Homebrew channel | A tap holding one formula, with a release URL and hash per platform | First channel in 0005. | 2026-10-16 | |
| Nix channel | A flake that fetches the release binary, with a URL and hash per platform | Second channel in 0005. | 2026-10-16 | |

Release assets need authentication while the repository is private, see 0005 and [#4](https://github.com/8thlight-daveloper/onboard-kit/issues/4).

## Retired

| Choice | Was | Since | Until | Replaced by | Why it changed |
| --- | --- | --- | --- | --- | --- |
| | | | | | |

## Considered and not chosen

| Choice | Option | Status | Why | Revisit when |
| --- | --- | --- | --- | --- |
| Nix channel | Build from source | Not chosen | Needs a reproducible build and a `vendorHash` that goes stale when dependencies change. The flake would install a different artifact than the one CI verified | The Nix user needs a build from source |
