# ACME Corp onboarding: Platform team

Generated 2026-10-15 from template set `example`. A person must review this before it
is sent.

## Gaps to fill

| Gap | Who fills it | By when |
| --- | --- | --- |
| Who grants work tracker access | Engineering manager | Before day one |

## Access

Request each system on day minus five if you can. Record who grants it and when you
asked.

- [ ] **Source control.** Granted by IT service desk. Request it at
  https://access.acme.example/request. You know it worked when
  `git ls-remote git@git.acme.example:platform/widget-service.git` succeeds.
- [ ] **CI dashboard.** Granted by Platform team lead. Sign in as `<account_name>`. You
  know it worked when you can open the CI dashboard.
- [ ] **Work tracker.** Granted by _unknown yet_. You know it worked when you can open
  the work tracker.

You know you are done when you can open the repository, the CI dashboard, and the
work tracker.

Owner: Engineering manager. When: before day one.

## Clone the project and run the tests

Get a working copy of the project and a passing test run on your machine.

- [ ] **Clone the repository.** Run
  `git clone git@git.acme.example:platform/widget-service.git widget-service`. You
  know it worked when `widget-service/.git` exists.
- [ ] **Install dependencies.** In `widget-service`, run `make setup`.
- [ ] **Run the tests.** In `widget-service`, run `make test`. You know it worked when
  the command exits with code 0.

If a test fails on a fresh clone, write down the failure and ask the team before
changing anything.

You know you are done when `make test` passes on a fresh clone.

Owner: newcomer. When: week one.

## If you are stuck

Who to ask when something is blocked and nobody named in a section can help.

If a step is blocked and the people named in its section cannot help, ask these people
first: Alex Rivera, alex.rivera@acme.example, @alex, then Sam Okafor,
sam.okafor@acme.example, @sam. Say what you tried and what you need. They will point
you to someone else to try.
