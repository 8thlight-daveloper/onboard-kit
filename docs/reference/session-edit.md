# Record-backed session editing

The release candidate provides an explicit record-only editing route:

```sh
onboard-kit session edit --root /absolute/onboarding --home /absolute/state --session KEY
```

Use the session key registered by invitation start. All three arguments are
required; the command does not guess identity, discover a home or import an
invitation. Pending or interrupted imports must be resolved first. The native
contract suite and the invitation-to-record binary scenario pass.

## Review and save

Confirm the displayed identity exactly, including the account's double quotes:

```text
confirm "newcomer"
```

The command then displays current activities, questions, recorded outcomes and
support. Choose one explicit change:

```text
answer personal.choice text "my answer"
follow-plan personal.choice
outcome access verified "" ""
```

Question and activity IDs come from that display. Answers also support flag,
list and deferred forms. Outcomes support verified, failed and deferred; their
two JSON strings record notes and differing instructions. No instruction, check
or command is executed. Verified is the user's statement, not an automated test.

Review the complete proposed result, then enter save or cancel. Exiting, cancelling
or incomplete consent writes no record. One invocation makes at most one change.
Defaults and suggestions are never accepted merely by opening this view.

## Ownership

An explicit answer becomes a personal override even when it equals a shared
answer. Follow-plan removes that override: it can reveal a shared stored answer,
not necessarily a blank. Later shared-plan changes do not overwrite an override.

Only the selected record is replaced. The shared plan, onboarder preparation,
registration, firstOpenedAt and historical plan reference remain unchanged.
Outcomes retain their event time and definition hash; a new outcome samples the
current time only after save consent and freshness checks. Planning dates are
goals, not failure deadlines.

## Native safeguards and limits

The command checks the registered root and participant identities, pending state,
bounded metadata and selected bytes before use and again before publication.
Cooperative root/home locks are released while waiting for input. Changed records
use canonical YAML; a no-op retains the original bytes and inode.

Atomic replacement preserves inspected record permissions, ownership and native
attributes, or refuses without replacing it. Unsupported controls can therefore
require local repair. A private sibling stage shields unpublished bytes. Unknown
or nonempty interrupted staging is preserved and refused, never replayed or
automatically deleted. Records, terminal scrollback and staging can contain
personal information.

This is not protection against a hostile owner or a power-loss guarantee. If
publication succeeds but later cleanup or reporting fails, the result truthfully
reports saved with repair required; it does not roll back or replay the change.
See [0460](../decisions/0460-record-backed-session.md) and the
[driver contract](../../contracts/session/editor/contract.md) for exact bounds.
