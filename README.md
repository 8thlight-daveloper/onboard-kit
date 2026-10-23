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

`onboard-kit` is a command-line tool, built from `binary/`. It is an early release candidate. It turns section templates into a plan, walks a person through the plan's questions in a terminal screen, and turns the plan into a document.

```sh
just build                                    # writes output/onboard-kit
output/onboard-kit walk --plan plan.obk.yaml --templates docs/reference/examples/acme/templates
output/onboard-kit assemble --plan plan.obk.yaml --out onboarding.md
```

| Command | Does |
| -- | -- |
| `start --org-plan FILE` | Legacy newcomer flow. Copies the organization's plan into their own plan file in a per-user folder (`OBK_HOME` or `--home` change it) and opens the onboardee walkthrough. Running it again reopens their file and refreshes from a changed organization plan |
| `walk --plan FILE [--templates DIR] [--audience onboarder\|onboardee]` | Opens the terminal walkthrough. It opens `FILE`, or starts a new plan from `DIR` when `FILE` does not exist. Every accepted answer is saved to `FILE` at once. Keys: enter answers a question, `d` defers it, `x` clears it, `s` switches sections on and off, `r` reviews the blockers and the document, `q` quits |
| `assemble --plan FILE [--out FILE]` | Writes a Markdown document from a plan. Gaps go on standard error and in the document's gap table |
| `new --templates DIR [--out FILE]` | Writes a new plan from a folder of section templates, and an optional `plan-variables.yaml` |
| `prepare <operation> --plan FILE --account ID [--root DIR]` | Creates or edits the onboarder's preparation for one person, or shows it beside recorded outcomes without changing the plan or record. See [preparation commands](docs/reference/preparation-command.md) |
| `stdio (--plan FILE \| --templates DIR)` | Answers line-delimited JSON requests about a plan, for scripts and tests |

Preparation editing is implemented. Invitation ZIP import and the record-backed
walkthrough remain planned; start and walk still use the legacy plan-file flow.

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
