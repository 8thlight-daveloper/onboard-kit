---
category: architecture
status: accepted
approval: person
builds-on: ['0100', '0160', '0450']
---

# 0580. Match protection to the threat we have

Builds on [0100](0100-behavior-judgment.md) and
[0160](0160-personal-data-and-preparation-approval.md). Narrows
[0450](0450-reviewed-existing-catalog-edit.md).

## Context

The tool is used by teams inside one organization, on their own machines. The
catalog edit guide (`catalog/editguide`) inspects every directory from the plan
up to `/` for owner, write permission, ACL and attribute changes, and inspects
them again under a lock before it saves. The check guards against another user
on the same machine changing a directory above the plan while someone edits.

That threat is rare for this tool. The check is not cheap:

- It is native code that differs on macOS and Linux.
- It refused valid saves when a shared temporary directory changed, which made
  `just check` fail intermittently until it was relaxed.
- Its files hold about 45% of the mutation survivors for the component, and
  some of those survivors cannot be killed without a seam for racing two
  processes.

0450 already says the edit is cooperative and makes no claim against a hostile
writer. The ancestor checks go further than that scope needs.

## Decision

### The new principle

Name the threat before building a defense. At this stage the threats are
mistakes, not attackers: a lost or half-written file, an edit saved over a
newer one, an edit that lands in the wrong file, and personal data written
where others can read it.

Keep a protection when it is cheap and prevents one of those. Defer a defense
when its only reach is a hostile local user or process and its cost is native
code, platform-specific checks, or tests that cannot be written without a race.

A deferred defense is recorded, so it can return:

- The ADR names the threat and says it is out of scope.
- The ADR says what would bring it back: a shared or multi-user host, wider
  distribution, or data beyond onboarding plans.

This principle does not defer protection for credentials, network access,
authentication, running commands from plan content, or the personal data rules
in 0160. Those stay in scope.

An agent does not drop a protection on its own authority. A person approves
the ADR that defers it.

### Catalog edit guide profile

Keep:

- Refuse a plan path that is a symlink, a non-regular file, or reached through
  a symlink or `..`.
- Require the selected parent directory to be owned by the effective user and
  not writable by group or others.
- Write a private temporary file in the same directory and replace the plan by
  atomic rename.
- Take the cooperative lock after consent, re-read the plan bytes, and refuse
  when they changed, so a stale edit never overwrites a newer one.
- Keep the plan file profile of 0450 unchanged, including exact preservation
  of native file attributes.

Drop:

- Inspecting the owner, mode, ACL and attributes of every directory above the
  selected parent.
- Re-inspecting those directories under the lock.
- The parent ACL check. A group or other write grant through an ACL is out of
  scope with the other hostile-local-user cases.

Out of scope until an ADR brings it back: a local user with write access to a
directory above the plan, and a hostile process under the same user.

## Consequences

- The contract, its verification plan and its tests change. They are protected,
  so a person authorizes that change. The component is then rebuilt from the
  new contract.
- `ancestor_churn_test.go` stays valid. It checks that a save succeeds while
  entries come and go above the plan.
- Fewer native checks mean fewer mutants and fewer survivors, and no
  remaining source of that flake.
- Other components use the same path walk: the session editor, preparation
  store, invitation starter, catalog creator and invite command. This ADR does
  not change them. Each gets reviewed against P9 in its own decision.
- A later shared deployment can restore the stricter profile with a new ADR.
  No stored format changes, so nothing needs migrating.
