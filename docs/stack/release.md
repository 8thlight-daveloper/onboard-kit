# Release

Part of the [stack](README.md). The constraints and channel order are in [0005](../decisions/0005-self-contained-install.md). This file holds the tools that build and publish a release.

## Current

| Choice | Current | Why | Since | Replace when |
| -- | -- | -- | -- | -- |
| Release build | A static build for `darwin/arm64`, verified in CI as 0005 describes | macOS arm64 is the first platform. | 2026-10-16 | Another platform is supported. |
| Version numbers | No tag until the first distribution, and the commit sha identifies every build before then. The first tag is `v0.1.0-rc.1`. Before 1.0, fixes and additions are patch bumps, and a minor bump means something an installed copy depends on broke, such as a plan file that no longer opens. `1.0.0` promises that plan files stay readable | Nobody installs the tool yet, so a number would mean nothing. The plan file's schema version, not the tool's version, decides what opens. | 2026-10-20 | A distribution channel exists, or the plan schema version changes. |
| Release tool | To decide. GoReleaser is a candidate, not yet evaluated | One tool could build the release and write the Homebrew and Nix channel files. |  |  |
| Homebrew channel | A tap holding one formula, with a release URL and hash per platform | First channel in 0005. | 2026-10-16 |  |
| Nix channel | A flake that fetches the release binary, with a URL and hash per platform | Second channel in 0005. | 2026-10-16 |  |

Release assets need authentication while the repository is private, see 0005 and [#4](https://github.com/8thlight-daveloper/onboard-kit/issues/4).

## Retired

| Choice | Was | Since | Until | Replaced by | Why it changed |
| -- | -- | -- | -- | -- | -- |

## Considered and not chosen

| Choice | Option | Status | Why | Revisit when |
| -- | -- | -- | -- | -- |
| Nix channel | Build from source | Not chosen | Needs a reproducible build and a `vendorHash` that goes stale when dependencies change. The flake would install a different artifact than the one CI verified | The Nix user needs a build from source |
