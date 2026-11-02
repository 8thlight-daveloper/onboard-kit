# Onboarding with onboard-kit

onboard-kit helps a team onboard a new person. Someone who knows the team, the onboarder, answers questions once. The tool turns those answers into a plan: a single YAML file that holds the onboarding steps, who to ask, and where things live. The new person, the onboardee, opens an invitation and works through the plan in a terminal. They record what worked and what did not, so the onboarder learns what to fix.

The tool is a single command-line binary, `onboard-kit`. It is an early release. It builds for macOS arm64 only, and there is no packaged download yet, so you build it from source.

## How it works

1. The onboarder starts a plan from the built-in section templates and answers its questions in a terminal screen. Every accepted answer is saved to the plan file at once.
2. The onboarder prepares one invitation per newcomer and exports it as a ZIP file.
3. The onboardee opens the invitation, confirms who they are, and walks through the plan. Their answers and outcomes go in their own onboarding record.
4. The plan can also be turned into a Markdown document.

Assembly is deterministic. The same plan and templates always produce the same document, and no model runs during assembly. The design notes are in [`docs/architecture`](docs/architecture/README.md), and each choice has a short record in [`docs/decisions`](docs/decisions/).

## Build and try it

You need [Nix](https://nixos.org/download) with flakes enabled. The dev shell supplies Go and every other tool at fixed versions.

```sh
nix develop                                   # or `direnv allow`, to enter the shell on cd
just build                                    # writes output/onboard-kit
just demo-init                                # creates a sample plan from the Acme example
just demo-plan                                # opens it in the terminal walkthrough
```

The [manual testing guide](docs/reference/manual-testing.md) covers the full demo: preparing an invitation, opening it as a newcomer, and viewing results. To generate a document without the walkthrough:

```sh
output/onboard-kit walk --plan plan.obk.yaml --templates docs/reference/examples/acme/templates
output/onboard-kit assemble --plan plan.obk.yaml --out onboarding.md
```

## Commands

| Command | Does |
| -- | -- |
| `start --org-plan FILE` | Legacy newcomer flow. Copies the organization's plan into their own plan file in a per-user folder (`OBK_HOME` or `--home` change it) and opens the onboardee walkthrough. Running it again reopens their file and refreshes from a changed organization plan |
| `start --invitation FILE --root DIR [--home DIR]` | Reviews an invitation, confirms identity and publishes approved files only after begin, then displays a read-only record-backed session. See [invitation opening](docs/reference/invitation-start.md) |
| `start --resume-import TOKEN --root DIR [--home DIR]` | Reviews and resumes or cleans an interrupted invitation opening in the same context. This is not a general onboarding resume command |
| `walk --plan FILE [--templates DIR] [--audience onboarder\|onboardee]` | Opens the terminal walkthrough. It opens `FILE`, or starts a new plan from `DIR` when `FILE` does not exist. Every accepted answer is saved to `FILE` at once. Keys: enter answers a question, `d` defers it, `x` clears it, `s` switches sections on and off, `r` reviews the blockers and the document, `q` quits |
| `assemble --plan FILE [--out FILE]` | Writes a Markdown document from a plan. Gaps go on standard error and in the document's gap table |
| `new --templates DIR [--out FILE]` | Writes a new plan from a folder of section templates, and an optional `plan-variables.yaml` |
| `prepare <operation> --plan FILE --account ID [--root DIR]` | Creates or edits the onboarder's preparation for one person, or shows it beside recorded outcomes without changing the plan or record. See [preparation commands](docs/reference/preparation-command.md) |
| `invite export --plan FILE --preparation FILE --account ID --out FILE --preaccess-help-contact ID` | Publishes a new invitation ZIP with the original plan, preparation, identity-only record seed and first-steps guide. See [invitation export](docs/reference/invitation-export.md) |
| `catalog list`, `catalog show --id ID`, `catalog export --id ID` | Lists, displays or exports all eighty compiled catalog entries without reading or writing files. See [the catalog](docs/reference/onboarding-catalog.md#using-the-embedded-catalog) |
| `catalog create --id ID --out FILE` | Selects embedded section templates, assigns permanent activity IDs and shared facts, then creates a new ordinary plan only after a complete review and save. Existing files are never overwritten |
| `catalog edit --plan FILE` | Stages additions, shared-answer changes, retirement or reactivation in an existing supported private plan. It displays the complete changes and saves only after review. Personal files are not read or changed |
| `session edit --root ABS --home ABS --session KEY` | Reviews one explicit personal answer, follow-plan choice or outcome, then saves only the selected registered record after separate consent. See [record editing](docs/reference/session-edit.md) |
| `stdio (--plan FILE \| --templates DIR)` | Answers line-delimited JSON requests about a plan, for scripts and tests |

Variable names must be unique within their scope, and nonempty step IDs within their section. Validation reports ambiguous declarations instead of choosing a winner. See [the ambiguity rule](docs/decisions/0110-reject-ambiguous-plans.md).

## Onboarding checklists

This repo also holds plain-Markdown onboarding checklists that do not need the tool. Copy them, fill in the specifics for the organization you are joining, and keep the result in that organization's own system.

| Folder | Purpose |
| -- | -- |
| `guides/` | How to turn these files into an onboarding doc for one organization |
| `core/` | Checklists and plans that apply to every engagement |
| `variants/team-state/` | What changes when joining an established team versus being the first consultants |
| `variants/discipline/` | What changes for software development versus design work |

Start with [`guides/how-to-build-your-onboarding-doc.md`](guides/how-to-build-your-onboarding-doc.md). Pick one team-state variant and one discipline variant, then combine them with `core/`. These files are first drafts, and the design variant is an outline.

## Contributing

Run `just check` before you commit. It formats and lints the Markdown, checks links, applies the prose rules, and runs the Go tests and acceptance scenarios. `just hook` installs a git hook that runs it before every commit. `just` lists the other recipes.

Read [`CONTRIBUTING.md`](CONTRIBUTING.md) first, especially the rules about what must stay out of this repo. Plan files and generated documents can carry details about a real organization, so they never belong here. Report security problems as described in [`SECURITY.md`](SECURITY.md).

### GoLand

The IDE does not see the dev shell's Go on its own. Run `just ide` once. It links the dev shell's Go at `.direnv/go`, a path that does not change. In GoLand, open Settings, Go, GOROOT, add a local SDK, and choose `.direnv/go/share/go` in this repo. After a flake update, run `just ide` again, and the IDE follows.

## License

Copyright © 2026 8th Light, LLC. Licensed under the MIT License. See [LICENSE](LICENSE).

8th Light® and the 8th Light logo are trademarks of 8th Light, LLC and are not covered by the MIT License.

<a href="https://8thlight.com"><img src="images/8th-light-logo.svg" alt="8th Light" width="64"></a>
