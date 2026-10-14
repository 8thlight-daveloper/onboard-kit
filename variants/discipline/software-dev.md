# Variant: software development

Adds the technical setup and norms that core does not cover.

## Access to request

- [ ] Source control organization and the specific repositories
- [ ] Issue tracker and project boards
- [ ] CI/CD system (view first, run or edit when needed)
- [ ] Cloud console or hosting accounts, read only at first
- [ ] Logging, monitoring, and alerting tools
- [ ] Package registries and artifact stores
- [ ] Secrets manager, if your work needs it
- [ ] Test and staging environments
- [ ] On-call or paging tool, if you will take part

Record where each is granted in `core/access-checklist.md`.

## Week one setup

- [ ] Clone the main repositories and build them locally
- [ ] Run the test suite. Note failures and flaky tests
- [ ] Run the app locally. Note every undocumented step
- [ ] Trace one change from code to production by reading, then by watching
- [ ] Open the project README and fix what is wrong

## Learn the codebase

- [ ] Read the top level structure and any architecture docs or decision records
- [ ] Find the oldest and newest parts of the code, and learn why they differ
- [ ] Identify the areas the team is afraid to touch
- [ ] Read recent merged changes to learn style and review habits

## Learn the delivery process

- [ ] Branching and merge strategy
- [ ] Code review: who reviews, how long it takes, what they look for
- [ ] How releases happen and how often
- [ ] How the team handles incidents and rollbacks
- [ ] Definition of done, including tests, docs, and approvals

## Learn the standards

- [ ] Coding conventions, linting, and formatting
- [ ] Testing approach and expected coverage
- [ ] Security practices, such as dependency updates and secret handling
- [ ] Observability: logs, metrics, traces, dashboards

## First contributions

- Start with a small change you can ship in the first week: a bug fix, a doc fix, a test.
- Follow the team's review process exactly, as a newcomer would.
- Pair on the first larger change.
- Add to the onboarding doc anything you had to figure out alone.
