# 0600. Release platforms and channels

Status: accepted. Supersedes in part [0005](0005-self-contained-install.md).

## Context

0005 named macOS arm64 as the only first platform and chose a Nix flake that
fetches the release binary. The repository is now public, and Windows users
reach the tool through WSL, which is Linux. The file checks and terminal
scenarios pass on Linux, and the third-party notices already cover the Linux
targets. Another 8th Light tool, repo-tutor, already releases through a tagged
workflow, a Homebrew tap and a Nix flake that builds from source.

## Decision

### Platforms

- Release `darwin/arm64`, `linux/amd64` and `linux/arm64`. Linux covers WSL.
- Not released: Intel Macs and native Windows. Adding one means a build, a
  verification run and a notices check for that target.
- Every build keeps the constraints in 0005: CGO off, `-trimpath`, no runtime
  and no system library beyond the operating system baseline.

### Archives and versions

- Each target is one `.tar.gz` holding the binary, `LICENSE` and
  `THIRD_PARTY_NOTICES.md`. A `checksums.txt` lists the SHA-256 of every archive.
- A `VERSION` file holds the current version. The tag must equal it.
- A `0.x` release is marked as a pre-release, as in the
  [stack release note](../stack/release.md).
- The binary reports its version with `--version`, set at build time.

### Release verification

CI verifies each archive's binary on a runner of the same operating system and
architecture, with the checks in 0005: linked libraries inside the baseline, no
development tools on PATH, and every acceptance scenario run against the
unchanged file. The macOS check keeps `codesign --verify`. The Linux check
requires a static executable.

### Channels

- **Homebrew.** The release workflow renders a formula with a URL and hash per
  target and commits it to `8thlight/homebrew-tap`. Nobody edits the formula by
  hand.
- **Nix.** A flake package builds the binary from source with `buildGoModule`.
  This replaces 0005's flake that fetches the release binary. One expression
  covers every system, and it matches how the other 8th Light tools install. A
  workflow builds the package on every push, so a stale `vendorHash` fails before
  a tag.
- The release URL is a plain public URL. No channel carries authentication code.

### Release tooling

A release script checks the branch, the tag and the passing workflows, bumps
`VERSION`, tags and pushes. A tag workflow checks, builds, verifies, attests
build provenance, publishes the release and updates the tap. A small Go tool
renders the formula. This follows repo-tutor. GoReleaser was not evaluated.

## Consequences

- 0005's platform list and its Nix clause no longer apply. Its constraints, its
  Homebrew clause and its verification rules still do.
- The Nix package is built from source, so it is not the byte-identical file the
  release verified. Both come from the same commit with the same flags.
- The verification tool in `tools/release/verify` needs a Linux check beside its
  macOS one.
- The tap needs a deploy key with write access for this repository, and this
  repository needs the private half as the `HOMEBREW_TAP_DEPLOY_KEY` secret.
  A tap admin adds the public key after the transfer to the organization.
  The tap's README must say that formulas here are MIT, not GPL.
- Each added platform costs a runner and a verification run.

## Alternatives considered

- **macOS arm64 only.** Keeps the build small. It leaves WSL and Linux users
  with no install path.
- **Add Intel Macs.** One more target and notices row. No known user needs it.
- **Fetch-binary Nix flake.** It installs the verified file, but each release
  needs a hand-written URL and hash per system, and every other 8th Light tool
  already builds from source.
- **GoReleaser.** Writes both channels from one config. Not evaluated, and the
  repo-tutor scripts already do the same work in code we own.
