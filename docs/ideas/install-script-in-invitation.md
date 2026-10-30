# Install script in the invitation

Status: idea, 2026-10-29.

## Idea

A newcomer should not need the binary installed before they start. The
invitation package carries a script that installs the tool and launches it with
the invitation file. A URL cannot do this: a custom URL handler needs the binary
installed first, and a web link can only lead to a page with instructions.

## Shape

- The package holds `install-and-start.sh` and the plain invitation file. The
  package stays usable without the script.
- The onboarder's own binary generates the script. It knows its release version
  and the expected hash of the release file for each platform, because release
  CI already computes those for the Homebrew and Nix channels. The script
  cannot drift from the release.
- The script starts by saying what it will do: download version X for this
  platform, check the hash, put the file in a named folder, then launch. It
  asks for permission. When declined it prints how to install by hand and the
  exact launch command. That text must carry the instructions itself, not only a
  link, because a link works only once the repository is public.
- The script names the install route it will use before it asks, such as
  Homebrew, and not only the download. A newcomer who uses Nix opts out of the
  automatic install and installs through Nix by hand. The decline text covers
  that route too.
- Platforms: macOS and Linux. Windows is supported only through WSL, which is
  Linux. The script is POSIX `sh`. Native Windows gets a clear refusal that
  points to WSL.

## Open questions

- The newcomer has no company access, so the download needs a public location.
  A public repository, a public releases location, or a separate releases-only
  repository would each do.
- macOS quarantine and notarization. A file fetched by a script may not get the
  quarantine flag that a browser download gets. This is open in
  [0005](../decisions/0005-self-contained-install.md).
- Which route the script uses when it installs: Homebrew where it exists, the
  plain release file, or a choice the newcomer makes. How a newcomer opts out
  without answering "no" to everything.
- Whether the script also installs related tools, such as a codebase tutor.

## Related

[0005](../decisions/0005-self-contained-install.md) and
[0240](../decisions/0240-invitation-archive-envelope.md).
