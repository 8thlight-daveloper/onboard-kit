---
category: architecture
status: accepted
approval: unrecorded
partly-superseded-by: ['0600']
---

# 0005. Self-contained install

0600 replaces the platform list and the Nix channel clause. Everything else here still applies.

## Context

Consultants run this tool at the start of an engagement, often on a new machine with limited setup. Every extra install step is a reason not to use it. Hosting is deferred.

## Decision

Installing the tool means getting one file and running it.

- No interpreter, runtime, package manager, or container is required on the consultant's machine.
- No system library beyond the baseline of each supported operating system.
- No second download at first run. Templates are embedded in the binary.
- It runs offline.

This is a hard constraint on every other choice. A language, library, or tool that breaks it is out, however convenient it is otherwise.

### Platforms

- **First platform:** macOS on arm64 (`darwin/arm64`) only. 8th Light provides Macs, and one target keeps the release build and its CI verification simple.
- **Not supported at first:** Intel Macs. Adding them later means a second build and a second verification run, and a universal binary is one way to do it. Linux and Windows may follow, each with its own verification run.

### Package channels

Package-manager channels are conveniences beside the standalone file, which stays the base. Each installs the same self-contained binary and needs nothing at run time. They cover the same platforms as the standalone file. Priority order:

1. **Homebrew.** A tap, which is a git repo holding a formula with the release URL and a hash per platform. Most Mac users already know it.
2. **Nix.** A flake that fetches the release binary. Fetching is simpler for us than building from source, because the flake is one small file of URLs and hashes and avoids a stale `vendorHash`. It is faster for the Nix user, with no compile on update. It also installs the exact binary that CI verified.

Release CI computes the hashes and writes both channel files, so a release updates every channel from one source. A release tool such as GoReleaser can do this, which is a stack choice for [release](../stack/release.md).

### Private or public repository

The channels must work whether the repository is private or public.

- If the repository is private, release assets need authentication. Homebrew needs a token and a custom download strategy. Nix `fetchurl` cannot authenticate, so Nix needs a private cache or another way to reach the release. This is part of hosting.
- If it is public, both channels use plain release URLs and need no authentication code. Keep the release URL and the authentication in one place in each channel file, so the change is small.
- Making a repository public needs the content rules in `CONTRIBUTING.md` to hold for the whole history. Check that first.

### Verification

CI checks the constraint on every release build:

- Build the release artifact for each supported platform.
- Run it in a clean environment with no language runtimes on the path, and run the acceptance suite against it.
- Inspect the linked libraries and fail if any are outside the operating system baseline.

## Consequences

- The language and library choice in [the stack](../stack/README.md) must satisfy this. A compiled language that produces one static binary does so natively. An interpreted language qualifies only if a bundled build passes the verification above.
- Libraries that need a native dependency must link it statically.
- Updating the tool means replacing the file. Update delivery is deferred with hosting.

## Alternatives considered

- **Interpreter-based install** (a Python tool through `uv` or `pipx`). Needs a runtime or tool manager on the machine, which is a secondary install.
- **Container image.** Needs a container runtime.
- **Installer or package manager as the only route.** Needs the manager. Package-manager channels are conveniences beside the standalone file, which stays the base. Homebrew and Nix are the first two, see above.

## Open questions

- How do releases reach the Homebrew and Nix channels when the repository is private?
- Does a downloaded binary need code signing or notarization to run without a warning on each platform? Ask whoever holds 8th Light's developer signing account whether signing and notarization already exist. A file installed by Homebrew or Nix may not get the quarantine flag that a browser download gets, so signing may matter mostly for the standalone download. Confirm that, and confirm that the build carries at least the ad-hoc signature arm64 macOS requires.
