# Onboarding plans

Generic onboarding material for consultants joining a new engagement. Copy them, fill in the specifics for the organization you are joining, and hand the result to the client so it outlives the engagement.

## What is here

| Folder | Purpose |
| -- | -- |
| `guides/` | How to turn these files into an onboarding doc for one organization |
| `core/` | Checklists and plans that apply to every engagement |
| `variants/team-state/` | What changes when joining an established team versus being the first consultants |
| `variants/discipline/` | What changes for software development versus design work |

## How to use it

1. Read `guides/how-to-build-your-onboarding-doc.md`.
2. Pick one team-state variant and one discipline variant.
3. Combine them with `core/` into a single doc for the organization.
4. Store that doc in the client's own system, not in this repo.

## Development

Tools come from a Nix dev shell, so everyone has the same versions.

```sh
nix develop    # or `direnv allow`, to enter the shell on cd
just check     # format, linters, and tests
just hook      # optional: run `just check` before every commit
```

`just` lists the other recipes. Run `just check` before committing.

### GoLand

The IDE does not see the dev shell's Go on its own. Run `just ide` once. It links the dev shell's Go at `.direnv/go`, a path that does not change. In GoLand, open Settings, Go, GOROOT, add a local SDK, and choose `.direnv/go/share/go` in this repo. After a flake update, run `just ide` again, and the IDE follows.

## The tool

For isolated manual testing, start with just demo-init and just demo-plan.
The [stage recipes](docs/reference/manual-testing.md) cover preparation,
invitations, newcomer activity and results, with recoverable per-stage resets.

`onboard-kit` is a command-line tool, built from `binary/`. It is an early release candidate. It turns section templates into a plan, walks a person through the plan's questions in a terminal screen, and turns the plan into a document.

```sh
just build                                    # writes output/onboard-kit
output/onboard-kit walk --plan plan.obk.yaml --templates docs/reference/examples/acme/templates
output/onboard-kit assemble --plan plan.obk.yaml --out onboarding.md
```

| Command | Does |
| -- | -- |
| `start --org-plan FILE` | Legacy newcomer flow. Copies the organization's plan into their own plan file in a per-user folder (`OBK_HOME` or `--home` change it) and opens the onboardee walkthrough. Running it again reopens their file and refreshes from a changed organization plan |
| `start --invitation FILE --root DIR [--home DIR]` | Reviews an invitation, confirms identity and publishes approved files only after begin, then displays a read-only record-backed session. See [invitation opening](docs/reference/invitation-start.md) |
| `start --resume-import TOKEN --root DIR [--home DIR]` | Reviews and resumes or cleans an interrupted invitation opening in the same context; this is not a general onboarding resume command |
| `walk --plan FILE [--templates DIR] [--audience onboarder\|onboardee]` | Opens the terminal walkthrough. It opens `FILE`, or starts a new plan from `DIR` when `FILE` does not exist. Every accepted answer is saved to `FILE` at once. Keys: enter answers a question, `d` defers it, `x` clears it, `s` switches sections on and off, `r` reviews the blockers and the document, `q` quits |
| `assemble --plan FILE [--out FILE]` | Writes a Markdown document from a plan. Gaps go on standard error and in the document's gap table |
| `new --templates DIR [--out FILE]` | Writes a new plan from a folder of section templates, and an optional `plan-variables.yaml` |
| `prepare <operation> --plan FILE --account ID [--root DIR]` | Creates or edits the onboarder's preparation for one person, or shows it beside recorded outcomes without changing the plan or record. See [preparation commands](docs/reference/preparation-command.md) |
| `invite export --plan FILE --preparation FILE --account ID --out FILE --preaccess-help-contact ID` | Publishes a new invitation ZIP with the original plan, preparation, identity-only record seed and first-steps guide. See [invitation export](docs/reference/invitation-export.md) |
| `catalog list`, `catalog show --id ID`, `catalog export --id ID` | Lists, displays or exports all eighty compiled catalog entries without reading or writing files. See [the catalog](docs/reference/onboarding-catalog.md#using-the-embedded-catalog) |
| `catalog create --id ID --out FILE` | Selects embedded section templates, assigns permanent activity IDs and shared facts, then creates a new ordinary plan only after a complete review and save. Existing files are never overwritten |
| `catalog edit --plan FILE` | Stages additions, shared-answer changes, retirement or reactivation in an existing supported private plan; displays the complete changes and saves only after review. Personal files are not read or changed |
| `session edit --root ABS --home ABS --session KEY` | Reviews one explicit personal answer, follow-plan choice or outcome, then saves only the selected registered record after separate consent. See [record editing](docs/reference/session-edit.md) |
| `stdio (--plan FILE \| --templates DIR)` | Answers line-delimited JSON requests about a plan, for scripts and tests |

Preparation editing, explicit invitation export, embedded catalog availability and prerequisite graph validation are implemented.
Confirmed invitation opening and its read-only record-backed session are implemented.
Reviewed catalog creation supports repeated activities with independent identities.
Reviewed existing-plan edits preserve held identities and reserve retired IDs.
Record-backed answer editing and outcome publication use a separate explicit route.
The invitation-to-record binary scenario verifies ownership and first-use preservation.
The explicit
legacy start route and walk still use the plan-file flow, without outcome-verification gates.

`just acceptance` builds a development binary and runs the scenarios under
`acceptance/scenarios/` against it.

On macOS arm64, `just release-check` builds one release artifact and
runs those scenarios against that exact file without development tools on PATH.
It also checks the signature and linked libraries. See
[local release verification](docs/stack/release.md#local-verification).

Variable names must be unique within their scope, and nonempty step IDs within
their section. Validation reports ambiguous declarations instead of choosing a
winner. Assembly refuses duplicate identities without replacing an existing
document. The same name in different scopes and repeated command arguments
remain valid. See [the ambiguity rule](docs/decisions/0110-reject-ambiguous-plans.md).

## What never goes in this repo

This repo is generic on purpose. Keep out:

- Client or organization names
- URLs, tenant names, repo names, and internal tool details
- Names of people at the client
- Credentials, tokens, and anything from a contract

Filled-in docs belong to the client. See `CONTRIBUTING.md` for the sanitizing rules.

## Status

| Piece | State |
| -- | -- |
| Core | First draft |
| Established team | First draft |
| First consultants on site | Outline |
| Software development | First draft |
| Design | Outline |
