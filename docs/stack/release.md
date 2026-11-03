# Release

Part of the [stack](README.md). The constraints and channel order are in [0005](../decisions/0005-self-contained-install.md). This file holds the tools that build and publish a release.

## Current

| Choice | Current | Why | Since | Replace when |
| -- | -- | -- | -- | -- |
| Release build | CGO-disabled `darwin/arm64`, `linux/amd64` and `linux/arm64` binaries built with `-trimpath`, each in a `.tar.gz` with `LICENSE` and `THIRD_PARTY_NOTICES.md` | See [0600](../decisions/0600-release-platforms-and-channels.md). Linux covers WSL. Only the explicitly allowed system libraries may be linked. | 2026-11-02 | Another platform is supported. |
| Release verification | `just release-check`, run locally on macOS arm64. The release workflow runs the same checks for each target on a runner of the same system and architecture, with a Linux check beside the macOS one | Builds one artifact, checks its architecture, linked libraries and signature, and runs every acceptance scenario against that unchanged file without development tools on PATH. | 2026-10-22 | The release workflow replaces the local run. |
| Version numbers | No tag until the first distribution, and the commit sha identifies every build before then. The first tag is `v0.1.0`, marked as a pre-release on the release page, with no `-rc` suffix because a `0.x` version already says it can change. Before 1.0, fixes and additions are patch bumps, and a minor bump means something an installed copy depends on broke, such as a plan file that no longer opens. `1.0.0` promises that plan files stay readable | Nobody installs the tool yet, so a number would mean nothing. The plan file's schema version, not the tool's version, decides what opens. | 2026-10-20 | A distribution channel exists, or the plan schema version changes. |
| Release tool | A `just release` script, a tag-triggered release workflow, and a small Go tool that renders the Homebrew formula | Follows repo-tutor. The script checks the branch and the passing workflows, bumps `VERSION`, tags and pushes. The workflow checks, builds, verifies, attests provenance, publishes and updates the tap. | 2026-11-02 | A release tool such as GoReleaser does the same work with less code of ours. |
| Homebrew channel | `8thlight/homebrew-tap` holds one formula, with a release URL and hash per platform. The release workflow commits it over a deploy key | First channel in 0005. | 2026-10-16 |  |
| Nix channel | A flake package built from source with `buildGoModule`. A package workflow builds it on every push | One expression covers every system. A stale `vendorHash` fails before a tag. See [0600](../decisions/0600-release-platforms-and-channels.md). | 2026-11-02 | The Nix user needs the exact verified binary. |

The repository is public, so release URLs need no authentication.

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

Runtime workspaces use the physical system temporary hierarchy outside the
checkout, not an inherited temporary override pointing into it. The verifier
establishes effective group ownership only on its own new empty private directory
before writing anything inside it. Existing parent controls and application
safety checks are unchanged. See [0490](../decisions/0490-release-runtime-workspace.md).

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
| Release build | A CGO-disabled `darwin/arm64` binary built with `-trimpath` | 2026-10-22 | 2026-11-02 | Three targets with notices in each archive | macOS arm64 was the first platform. Linux and WSL users need an install path. |
| Nix channel | A flake that fetches the release binary, with a URL and hash per platform | 2026-10-16 | 2026-11-02 | A flake package built from source | One expression for every system, matching other 8th Light tools. See 0600. |

## Considered and not chosen

| Choice | Option | Status | Why | Revisit when |
| -- | -- | -- | -- | -- |
| Release verification | Small Go command using the existing acceptance runner | Chosen, 2026-10-22 | Reuses unchanged scenario data and matching rules, reads Mach-O metadata with the standard library, and needs no release service or new dependency | A release automation tool can call the same command |
| Nix channel | Build from source | Chosen, 2026-11-02 | The package workflow catches a stale `vendorHash` on every push, and one expression covers every system | The Nix user needs the exact verified binary |
| Nix channel | Fetch the release binary | Not chosen, 2026-11-02 | It installs the verified file, but each release needs a URL and hash per system | The package workflow cannot keep the hash fresh |
| Release tool | GoReleaser | Not chosen, 2026-11-02 | Not evaluated. The repo-tutor scripts already do the same work in code we own | The scripts become costly to keep |
