# Going public

Status: updated, 2026-11-02.

A working brief for an agent that prepares the repository to be made public.
The agent prepares and reports. A person decides and flips visibility.

## Where it stands

- **Repository.** Private, under a personal account. The intent, from the
  open source readiness issue, is to publish under the 8th Light organization.
- **License.** MIT, with an 8th Light copyright notice, in `LICENSE`. 8th Light
  has approved the license and the use of its name.
- **History.** Rewritten on 2026-11-02. A platform vendor name was removed from
  the tree, commit messages and two issues, and every author now uses a GitHub
  noreply address. Check a fresh clone for both before the flip, because
  commits have landed since.
- **GitHub issues.** Audited, clean. Docs and contracts no longer link to
  issues, so a reader without access loses nothing.
- **CI.** `check.yml` runs `just check` in the Nix dev shell on Ubuntu and is
  green, with the install summary switched off. `mutate.yml` runs mutation
  testing on a schedule.
- **Linux.** The file checks and terminal scenarios were fixed for ext4,
  Linux signal names and the `CI` variable. The check passes on macOS arm64,
  NixOS and the Ubuntu runner.
- **Local files.** `.idea/`, `.DS_Store`, `.scratch/` and `output/` are
  ignored.
- **Third-party notices.** `THIRD_PARTY_NOTICES.md` is generated from the
  modules the binary links, and `just check` fails when it drifts.
- **Root files.** `SECURITY.md` sends reports to GitHub private reporting.
  There is no code of conduct yet, because no private enforcement contact
  exists.

## Done

1. README rewritten for readers outside 8th Light.
2. Architecture docs say the onboardee runs the tool.
3. `SECURITY.md`, scoped as in [0580](../decisions/0580-security-scope-at-this-stage.md).
4. `CONTRIBUTING.md` rewritten for outside contributors, with the sanitizing
   rules kept.
5. Third-party notices tool and drift check.
6. Linux run fixed and green.
7. History and author rewrite.
8. Primary sources linked in `docs/prior-art.md`.

## Open

- **Code of conduct.** Add the Contributor Covenant once a person names a
  private enforcement contact other than a personal address.
- **Lint paths.** `SECURITY.md` and `THIRD_PARTY_NOTICES.md` sit outside the
  `justfile` Markdown paths. Adding them is a tool config change, so a person
  decides.
- **Bare issue numbers.** `docs/stack/language.md` and one secrets contract
  fixture name issue numbers without links. Harmless to a reader, but contracts
  are a person's to change.
- **Install story.** The documented path is a build from source. No tag exists.
  Check `docs/stack/release.md` against what exists, then tag `v0.1.0`.
- **Final audits.** Rerun the history, issue and working tree scans on the
  exact commit that will be published, with the rules in `CONTRIBUTING.md`.

## Needs a person

The agent must not do these, and should list them as open at the end.

- **Transfer to the organization,** with issues and the project board. A
  transfer keeps stars and issue numbers, but check which team owns it.
- **Project board visibility.** Not yet checked.
- **Visibility change.** Only the repository owner flips it, and after
  everything above is done.
- **After the flip.** Turn on private vulnerability reporting, set branch
  protection with `check` required, and confirm private reporting survives the
  transfer.
- **Pushing.** Nothing here authorizes a push. The agent commits locally and
  asks.

## How the agent proves it is done

- `just check` passes.
- A fresh clone builds with the documented commands and passes the checks.
- The audits are recorded in a note under `output/`, which git ignores, and
  summarized for the person.
- The open-source readiness checklist is updated by the person, not the agent.
