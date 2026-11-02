# Security policy

## Report a problem privately

Do not open a public issue for a security problem. Use the **Report a vulnerability** button on this repository's Security tab, or [open a private report](../../security/advisories/new) directly. Only the maintainers can read it. Include what you found, the steps to reproduce it, and the version or commit you tested.

We reply in the report, so you do not need to share an email address. We will confirm that we received it and keep you informed as we look into it.

## Supported versions

onboard-kit has no tagged release yet. We fix problems on `main`. Once releases exist, this section will name the supported ones.

## What counts as a security problem

onboard-kit runs on a person's own machine, for teams inside one organization. At this stage we protect against mistakes more than attackers: a lost or half-written file, an edit saved over a newer one, an edit that lands in the wrong file, and personal data written where others can read it. [Decision 0580](docs/decisions/0580-security-scope-at-this-stage.md) records the reasoning.

These are in scope:

- Credentials and secrets stored in a plan, record, or invitation.
- Network access the tool makes.
- Authentication.
- Running commands that come from plan content.
- Personal data handling, as [decision 0160](docs/decisions/0160-personal-data-and-preparation-approval.md) describes it.
- Overwriting or exposing a file the user did not choose.

These are out of scope for now:

- A local user who can write to a directory above a plan file.
- A hostile process running as the same user.

A shared or multi-user host, wider distribution, or data beyond onboarding plans would bring these back into scope. If you think one of them applies to your use, tell us through the private channel above.
