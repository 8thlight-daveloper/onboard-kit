---
category: ux
status: proposed
affects: [command/create]
builds-on: ['0120', '0650', '0670', '0680']
---

# 0690. Ask for a value as the person has it

Builds on [0120](0120-defaults-require-acceptance.md),
[0650](0650-the-front-door-first-build.md),
[0670](0670-layered-plans-and-where-a-plan-lives.md), and
[0680](0680-focus-zones-and-pages-on-the-answer-screen.md). It settles how a question
asks for a value, using where a plan is kept as the case.

## Context

Where a plan is kept is a repository and a file in it. It could be one question, two, or
three: host, organization, and repository. A person also has a third form, the link to the
file's page in a browser, such as `https://example.test/acme/onboarding.git/tree/master/plans/onboarding.obk.yaml`.
That link is one thing that a person could paste from one place, so "one paste from one
place" does not decide the case alone.

## Decision

**Ask for a value as one question when the person holds it as one thing and the tool can
use it as given, or split it reliably.** A repository address is the clone address from the
repository's Clone button. It is the repository's identity, and `git clone` takes it as it
is. The tool derives the parts it needs, such as the host and path that name the standard
folder in 0670.

**Do not ask the person to do the tool's parsing.** Host plus organization plus repository,
or a first and last name, are the tool's model of the value and not how the person knows it.
Splitting them breaks on cases that were not seen, such as subgroups, projects, ports,
and the ssh and https forms.

**Ask for separate values when the parts have different kinds, sources, owners, or rates of
change, or when the tool uses them separately.** The repository address and the path of the
plan file inside it are two questions. The address is an identity, and the path says which
plan. One repository holds many plans, so the address is the same for an organization plan
and every team plan. A contact's name, email, and chat handle are three fields for the same
reason.

**A link to a page is a view, not a value.** It carries one forge's page syntax, and it
cannot be used as a clone address or split without that forge's rules. A branch name with a
slash makes the split ambiguous.

**Say where to copy the value from.** The question names the source. The repository address
question says "Copy it from the repository's Clone button, not from a page of the file."

**When a common wrong form exists, answer it in layers.** The first layer is the wording
above. The second is a pattern on the question that requires the allowed clone forms,
`https://`, `ssh://`, or `git@host:path`, which also keeps an address from starting with
`-`, with a message that says what to paste. The third recognizes a known forge's file link
and offers the split, which the person accepts as 0120 requires. The first layer is built.

## Consequences

- The plan source is two optional questions in a first section of a plan, as 0650 records.
- The pattern layer arrives with the address safety checks of 0670. A pattern cannot
  exclude a substring such as `/tree/`, so the wording carries that case.
- The third layer needs a list of forges and an answer for branch names with slashes. It
  waits until someone has the problem.
