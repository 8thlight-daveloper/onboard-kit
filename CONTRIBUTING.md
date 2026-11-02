# Contributing

Thanks for helping improve onboard-kit. This page says what to check before you open a pull request. Be kind to each other, as the [code of conduct](CODE_OF_CONDUCT.md) asks.

For a larger change, open an issue first so we can talk about the approach before you write code. Small fixes to docs, templates, and tests can go straight to a pull request.

## Keep real organizations out

Plan files, onboarding records, and generated documents can carry details about a real organization. Those details must never land in this repo. Before you commit, check each change against these rules:

- No client or organization names, including in examples. Use `<Organization>` and `<Team>`.
- No real URLs, repo names, or tool tenants. Use `<link to ...>`. Test data uses reserved domains such as `example.test` and `example.invalid`.
- No names of people at an organization. Use roles: "the onboarding owner", "your tech lead".
- No credentials, tokens, or anything from a contract.
- No figures, dates, or outcomes from a real engagement.

If a lesson only makes sense with an organization's details attached, state the general pattern and leave the details out.

## Writing style

Write for a newcomer who has never heard of us.

- Plain words. Short sentences. Active voice.
- Imperative for steps: "Ask your lead who approves access."
- Sentence case headings.
- Variables go in angle brackets: `<Organization>`.
- One idea per checklist item.

## Set up and check your work

Tools come from a Nix dev shell, so everyone has the same versions. See the [README](README.md#build-and-try-it) for setup.

Run `just check` before you push. It formats and lints Markdown, checks that relative links and heading anchors resolve, applies the voice rules (no em dashes, no banned filler words, sentence case headings), and runs the tests. `just fmt` fixes formatting. `just hook` runs the check before every commit. CI runs the same recipe, so a green local run means a green CI run.

## Where files go

- `output/` holds plans and documents the tool generates. Git ignores everything in it except `output/.gitkeep`. Never commit what is in there.
- `.scratch/` is for working files. Git ignores it. Create it when you need it. `.scratch/tmp/` and `.scratch/worktrees/` are safe to delete, and `just clean-scratch` and `just clean-worktrees` do it. Everything else in `.scratch/` is yours to keep or prune.

## How the code is organized

The tool is built from small components, and each one has a contract and tests. Read [`AGENTS.md`](AGENTS.md) and the [architecture docs](docs/architecture/README.md) before you change code. Two rules matter most:

- A change to a contract, an acceptance scenario, or the domain model needs a maintainer's approval. Open an issue and ask.
- Do not add a dependency the [dependency rules](docs/decisions/0020-binary-structure-and-regenerative-components.md) do not allow.

Choices that tests depend on are recorded as short decision records in [`docs/decisions`](docs/decisions/). A new decision supersedes an old one, so add a record instead of editing history.

## Adding to the checklists

The Markdown checklists under `core/` and `variants/` hold only what differs from `core/`. If a variant repeats core content, move that content to core. A new discipline, such as data or product management, is one new file in `variants/discipline/`.

## Reporting a security problem

Do not open a public issue. Follow [`SECURITY.md`](SECURITY.md).
