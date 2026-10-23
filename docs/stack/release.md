# Release

Part of the [stack](README.md). The constraints and channel order are in [0005](../decisions/0005-self-contained-install.md). This file holds the tools that build and publish a release.

## Current

| Choice | Current | Why | Since | Replace when |
| -- | -- | -- | -- | -- |
| Release build | A CGO-disabled `darwin/arm64` binary built with `-trimpath` | macOS arm64 is the first platform. Only the explicitly allowed macOS system libraries may be linked. | 2026-10-22 | Another platform is supported. |
| Release verification | `just release-check`, run locally on arm64 | Builds one artifact, checks its architecture, linked libraries and signature, and runs every acceptance scenario against that unchanged file without development tools on PATH. No CI service is required. | 2026-10-22 | Release automation calls the same checks. |
| Version numbers | No tag until the first distribution, and the commit sha identifies every build before then. The first tag is `v0.1.0`, marked as a pre-release on the release page, with no `-rc` suffix because a `0.x` version already says it can change. Before 1.0, fixes and additions are patch bumps, and a minor bump means something an installed copy depends on broke, such as a plan file that no longer opens. `1.0.0` promises that plan files stay readable | Nobody installs the tool yet, so a number would mean nothing. The plan file's schema version, not the tool's version, decides what opens. | 2026-10-20 | A distribution channel exists, or the plan schema version changes. |
| Release tool | To decide. GoReleaser is a candidate, not yet evaluated | One tool could build the release and write the Homebrew and Nix channel files. |  |  |
| Homebrew channel | A tap holding one formula, with a release URL and hash per platform | First channel in 0005. | 2026-10-16 |  |
| Nix channel | A flake that fetches the release binary, with a URL and hash per platform | Second channel in 0005. | 2026-10-16 |  |

Release assets need authentication while the repository is private, see 0005 and [#4](https://github.com/8thlight-daveloper/onboard-kit/issues/4).

## Local verification

Run from the repository root on macOS arm64:

```sh
nix develop -c just release-check
```

This builds `output/release/onboard-kit` once with CGO disabled. It checks that
the artifact is an arm64 Mach-O executable, permits only
`/usr/lib/libSystem.B.dylib` and `/usr/lib/libresolv.9.dylib`, and runs
`codesign --verify --strict`. Any other library fails verification and needs
review, not an automatic exception. Signature verification does not establish
distribution signing, notarization, or browser-download quarantine behavior.

The development machine needs Go and the macOS `codesign` tool. The compiled
acceptance runner and the artifact run with no inherited environment, an
empty directory for PATH, and isolated temporary files and tool state. The
runner receives the artifact's absolute path and does not rebuild it. Its
SHA-256 checksum must be the same before and after the scenarios. A successful
run prints that checksum. A failure exits nonzero without claiming verification.

To check an existing artifact without rebuilding it:

```sh
nix develop -c just release-check --binary output/release/onboard-kit
```

Other hosts fail with an explanation rather than silently skipping checks.
The scenario files and their matching rules are unchanged. The checks cover
the current scenarios, including supplied templates and all eighty embedded
catalog exports; they do not claim that offline network isolation, distribution channels, release
publication, or release CI are implemented. No tag or publication is created.

## Retired

| Choice | Was | Since | Until | Replaced by | Why it changed |
| -- | -- | -- | -- | -- | -- |
| Release build | A static `darwin/arm64` build described as CI-verified | 2026-10-16 | 2026-10-22 | Local artifact verification | The existing CI workflow runs `just check`, not release verification. Make the artifact checks runnable locally before adding automation. |

## Considered and not chosen

| Choice | Option | Status | Why | Revisit when |
| -- | -- | -- | -- | -- |
| Release verification | Small Go command using the existing acceptance runner | Chosen, 2026-10-22 | Reuses unchanged scenario data and matching rules, reads Mach-O metadata with the standard library, and needs no release service or new dependency | A release automation tool can call the same command |
| Nix channel | Build from source | Not chosen | Needs a reproducible build and a `vendorHash` that goes stale when dependencies change. The flake would install a different artifact than the one CI verified | The Nix user needs a build from source |
