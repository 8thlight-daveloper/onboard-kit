# Dev tooling

Part of the [stack](README.md). The tools that format, lint, and check this repo, and the CI that runs them. Run everything with `just check`, from the Nix dev shell.

## Current

| Choice | Current | Why | Since | Replace when |
| -- | -- | -- | -- | -- |
| Task runner | `just` | One command surface for local runs, the git hook, and CI | 2026-10-16 | A recipe needs logic that `just` handles badly |
| Dev environment | A Nix flake dev shell, `nixpkgs-unstable`, pinned by `flake.lock`, for `aarch64-darwin` and `x86_64-linux` | Everyone, and CI, gets the same tool versions | 2026-10-16 | A platform we support is missing, or the pin goes stale |
| Markdown format | `mdformat` with the GFM and frontmatter plugins, wrapping kept, compact tables | Formats without reflowing our prose. Compact tables keep diffs small | 2026-10-16 |  |
| Markdown lint | `markdownlint-cli2`, line length rule off | Structure checks | 2026-10-16 |  |
| Link check | `lychee`, offline, with heading anchors | Catches broken relative links and moved headings, with no custom code. Tested on a broken file, a broken anchor, and a broken cross-file anchor | 2026-10-16 | It stops checking anchors correctly |
| Voice rules | Vale, with the `OnboardKit` style in `.vale/`. YAML is read as Markdown text | One home for voice rules. Covers Markdown and prose inside template YAML | 2026-10-16 | A rule needs parsing that text matching cannot do |
| CI | GitHub Actions on `ubuntu-latest`. `nix-installer-action` v23, then `nix develop --command just check` | The same recipe as local. The first run took 52 seconds | 2026-10-16 | The job takes long enough to need caching, or a macOS job is needed |

Both actions are pinned by commit SHA.

## Trying the tool by hand

The walk saves every answer to the plan file it opens, so never point it at a file under docs/. The [manual stage recipes](../reference/manual-testing.md) keep each role's test data in output/manual-test/, which Git ignores. Stage resets archive only selected data and never delete another stage's work.

| Recipe | Does |
| -- | -- |
| `just demo-init` | Creates an unanswered Acme-template plan, or retains the existing plan |
| `just demo-plan` | Edits the shared organization plan |
| `just demo-prepare ACCOUNT` | Reviews and saves one fictional recipient's preparation |
| `just demo-invite CONTACT` | Exports with an explicitly chosen pre-access helper |
| `just demo-onboardee` | Opens the invitation in a separate personal workspace |
| `just demo-progress` | Edits one explicit personal answer or outcome |
| `just demo-results` | Shows current preparation and returned outcomes as JSON |
| `just demo-status` | Lists all selected file paths without changing data |
| `just demo-reset STAGE` | Archives only the chosen stage, or explicitly all five |

## Retired

| Choice | Was | Since | Until | Replaced by | Why it changed |
| -- | -- | -- | -- | -- | -- |

## Considered and not chosen

Evidence is dated, because it ages.

| Choice | Option | Status | Why | Revisit when |
| -- | -- | -- | -- | -- |
| Markdown format | Prettier | Not chosen | Skipped in the tooling issue. `mdformat` handles GFM tables and keeps our wrapping |  |
| Nix caching in CI | `nix-community/cache-nix-action` | Not now | The first CI run took 52 seconds, so there is little to save. Its README lists three compatible installers and does not list `nix-installer-action`. It warns that it may slow a workflow, and caches are capped at 10 GB. Checked 2026-10-16 | A cold run takes over two minutes |
| CI runner | `Yeradon/nixos-ci-image` | Not chosen | A personal repo with 2 stars, 9 commits, and no releases on 2026-10-16. Built for self-hosted runners, with Nix's sandbox and syscall filtering off. Use on GitHub-hosted runners, and private pulls, are not stated | We move to self-hosted runners and need a Nix image |
| Git hook split | A fast `pre-commit` recipe beside `check` | Merged | Both took about a second, so one recipe, `check`, does both | `check` becomes slow enough to need a fast subset |
