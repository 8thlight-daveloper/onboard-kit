# ACME Corp onboarding: Platform team

Generated 2026-10-15 from template set `example`. A person must review this before it is sent.

## Gaps to fill

| Gap | Who fills it | By when |
| --- | --- | --- |
| Who grants work tracker access? | Engineering manager | Before day one |

## Access

Request the accounts and permissions you need, and record who grants each.

- [ ] **Source control.** Owner: IT service desk. Open: https://access.acme.example/request. Check: `git ls-remote git@git.acme.example:platform/widget-service.git` exits with code 0.
- [ ] **CI dashboard.** Owner: Platform team lead. Sign in as `<account_name>`. Check: Can you open the CI dashboard?
- [ ] **Work tracker.** Owner: _deferred_. Check: Can you open the work tracker?

Request each system on day minus five if you can. Record who grants it and when
you asked.

Done when: You can open the repository, the CI dashboard, and the work tracker.

Owner: Engineering manager. When: before day one.

## Clone the project and run the tests

Get a working copy of the project and a passing test run on your machine.

- [ ] **Clone the repository.** Run: `git clone git@git.acme.example:platform/widget-service.git widget-service`. Check: `test -d widget-service/.git` exits with code 0.
- [ ] **Install dependencies.** Directory: `widget-service`. Run: `make setup`.
- [ ] **Run the tests.** Directory: `widget-service`. Run: `make test`.

If a test fails on a fresh clone, write down the failure and ask the team before
changing anything.

Done when: make test passes on a fresh clone.

Owner: newcomer. When: week one.

## If you are stuck

Who to ask when something is blocked and nobody named in a section can help.

If a step is blocked and the people named in its section cannot help, ask these people
first: Alex Rivera, alex.rivera@acme.example, @alex, then Sam Okafor, sam.okafor@acme.example, @sam.
Say what you tried and what you need. They will point you to someone else to try.
