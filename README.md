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
