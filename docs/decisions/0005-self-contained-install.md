# 0005. Self-contained install

Status: accepted

Relates to [#24](https://github.com/8thlight-daveloper/onboard-kit/issues/24), [#4](https://github.com/8thlight-daveloper/onboard-kit/issues/4).

## Context

Consultants run this tool at the start of an engagement, often on a new machine with limited setup. Every extra install step is a reason not to use it. The tool also stays internal, so we control how it is delivered, and hosting ([#4](https://github.com/8thlight-daveloper/onboard-kit/issues/4)) is deferred.

## Decision

Installing the tool means getting one file and running it.

- No interpreter, runtime, package manager, or container is required on the consultant's machine.
- No system library beyond the baseline of each supported operating system.
- No second download at first run. Modules are embedded in the binary.
- It runs offline.

This is a hard constraint on every other choice. A language, library, or tool that breaks it is out, however convenient it is otherwise.

### Verification

CI checks the constraint on every release build:

- Build the release artifact for each supported platform.
- Run it in a clean environment with no language runtimes on the path, and run the acceptance suite against it.
- Inspect the linked libraries and fail if any are outside the operating system baseline.

## Consequences

- The language and library choice in `docs/stack.md` must satisfy this. A compiled language that produces one static binary does so natively. An interpreted language qualifies only if a bundled build passes the verification above.
- Libraries that need a native dependency must link it statically.
- Updating the tool means replacing the file. Update delivery stays in [#4](https://github.com/8thlight-daveloper/onboard-kit/issues/4).

## Alternatives considered

- **Interpreter-based install** (a Python tool through `uv` or `pipx`). Needs a runtime or tool manager on the machine, which is a secondary install.
- **Container image.** Needs a container runtime.
- **Installer or package manager as the only route.** Needs the manager. A package-manager channel can exist later as a convenience, as long as the standalone file stays the base.

## Open questions

- Which Mac architectures do we support first? macOS is the first platform, since 8th Light provides Macs. Linux and Windows may follow, and CI verification would then run for each.
- Does a downloaded binary need code signing or notarization to run without a warning on each platform?
