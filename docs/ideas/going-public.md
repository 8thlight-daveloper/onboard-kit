# Going public

Status: researched, 2026-11-02.

A working brief for an agent that prepares the repository to be made public.
The agent prepares and reports. A person decides and flips visibility.

## Where it stands

- **Repository.** Private, under a personal account. The intent, from the
  open source readiness issue, is to publish under the 8th Light organization.
  8th Light has not approved this.
- **License.** MIT, with an 8th Light copyright notice, in `LICENSE`. 8th Light
  has not confirmed the choice.
- **History.** Audited once on 2026-10-30 for secrets, client details and
  personal details. Nothing found. The one author email, the maintainer's work
  address, was accepted by the maintainer. Audit again before the flip, because
  commits have landed since.
- **GitHub issues.** Audited once, clean. Docs and contracts no longer link to
  issues, so a reader without access loses nothing.
- **CI.** `check.yml` runs `just check` in the Nix dev shell on Ubuntu. The
  terminal acceptance scenarios have not yet run on Linux. `mutate.yml` runs
  mutation testing on a schedule.
- **Local files.** `.idea/`, `.DS_Store`, `.scratch/` and `output/` are
  ignored. `.codex/` is untracked and not ignored.

## Gaps an agent can close

Each is a change to docs or repository files. None needs a person's approval
beyond review of the diff, except where noted.

01. **README describes the old project.** It presents the repository as
    generic consultant onboarding checklists and does not mention the tool, how
    to install it, or how to run it. Rewrite it for a reader outside 8th Light.
    Run it through the brand voice and prose checks.
02. **Architecture overview says the client never runs the tool.** The line
    `Client-facing output only` in `docs/architecture/README.md` and similar
    wording elsewhere predate the plan file and the onboardee audience. Search
    `docs/` for "internal tool" and "client never" and correct each.
03. **No `SECURITY.md`.** Write a policy that says how to report a problem
    privately. The contact address is a person's decision, so leave it as a
    marked blank for them. Read [0580](../decisions/0580-security-scope-at-this-stage.md)
    first and state the same scope.
04. **No code of conduct.** Propose the Contributor Covenant text. A person
    chooses the enforcement contact.
05. **`CONTRIBUTING.md` addresses 8th Light consultants.** It talks about
    engagements and "the client". Rewrite for outside contributors, and keep
    the sanitizing rules because plan files can carry client details.
06. **Third-party notices.** Dependencies are permissive, but the binary carries
    their code, and MIT and BSD ask that notices travel with it. Generate a
    `THIRD_PARTY_NOTICES` file for the modules in `go.mod`, and add a check
    that it matches the module list. Record the tool used in
    `docs/stack/dev-tooling.md`.
07. **One dependency has no license file.** `go-localereader` ships none in the
    module. Its upstream README states MIT. Confirm upstream, and decide with a
    person whether that is enough.
08. **Purego is Apache 2.0.** Its notice and any NOTICE text must be in the
    third-party file.
09. **`.codex/`.** Decide with a person whether to ignore or commit it.
10. **Linux run.** After the next push, read the `check` result for the
    terminal scenarios and fix what differs.
11. **Final audits.** Rerun the history audit and the issue audit on the exact
    commit that will be published. Scan the working tree for names of clients
    and people, URLs, tenants and tokens, with the rules in `CONTRIBUTING.md`.
12. **Install story.** `docs/ideas/install-script-in-invitation.md` and the
    release stack note describe distribution. A public repository needs at
    least a documented build from source and a tagged release. Check
    `docs/stack/release.md` against what exists.

## Needs a person

The agent must not do these, and should list them as open at the end.

- **8th Light sign-off,** with a named approver, and the choice of license.
- **Use of the 8th Light name and logo.** The brand identity skill covers
  visual use. Approval is not the agent's.
- **Transfer to the organization,** with issues and the project board. A
  transfer keeps stars and issue numbers, but check which team owns it.
- **Project board visibility.** Not yet checked.
- **Visibility change.** Only the repository owner flips it, and after
  everything above is done.
- **Branch protection and required checks.** The current plan does not offer
  these on a private personal repository. Set them after the move.
- **Pushing.** Nothing here authorizes a push. The agent commits locally and
  asks.

## Order

1. Items 1 to 5 in one pass, since they share the README voice.
2. Items 6 to 9, which touch files and tooling.
3. A push, once a person confirms, then item 10.
4. Item 11 on the final commit.
5. The person items, in the order sign-off, license, transfer, visibility.

## How the agent proves it is done

- `just check` passes.
- A fresh clone builds with the documented commands and passes the checks.
- The audits in item 11 are recorded in a note under `output/`, which git
  ignores, and summarized for the person.
- The open-source readiness checklist is updated by the person, not the agent.
