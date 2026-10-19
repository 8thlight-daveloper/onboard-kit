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
just check     # exactly what CI runs: format, linters, tests
just hook      # optional: run `just check` before every commit
```

`just` lists the other recipes. CI runs `just check` on every pull request.

### GoLand

The IDE does not see the dev shell's Go on its own. Run `just ide` once. It links the dev shell's Go at `.direnv/go`, a path that does not change. In GoLand, open Settings, Go, GOROOT, add a local SDK, and choose `.direnv/go/share/go` in this repo. After a flake update, run `just ide` again, and the IDE follows.

## The tool

`onboard-kit` is a command-line tool, built from `binary/`. It is an early release candidate: it turns section templates into a plan file and a plan file into a document. It has no interactive walkthrough yet.

```sh
just build                                    # writes output/onboard-kit
output/onboard-kit new --templates docs/reference/examples/acme/templates --out plan.obk.yaml
output/onboard-kit assemble --plan docs/reference/examples/acme/acme.obk.yaml
```

- `new` reads a folder of section template files, and an optional `plan-variables.yaml`, and writes a plan file.
- `assemble` reads a plan file and writes a Markdown document. Gaps in the plan go on standard error and in the document's gap table.
- `just acceptance` runs the scenarios under `acceptance/scenarios/` against the built binary.

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
