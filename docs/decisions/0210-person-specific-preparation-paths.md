---
category: behavior
status: accepted
approval: agent
affects: [command/prepare, preparation/locator]
builds-on: ['0090', '0110', '0140']
---

# 0210. Person-specific preparation paths

Builds on [0090](0090-plan-and-onboarding-record.md),
[0110](0110-reject-ambiguous-plans.md), and
[0140](0140-person-specific-preparation.md).

## Context

Preparation lives beside one person's record with the fixed name
preparation.obk.yaml. A legacy results path with every record in one directory
cannot provide that sibling name independently for each account. Guessing another
destination would change the agreed convention or risk overwriting another person.
Path interpolation must also stay bounded even when untrusted input repeats a
long account many times.

## Decision

An empty resultsPath uses records/{account}/record.obk.yaml. This is a storage
convention, not a default answer to a question. A nonempty invalid configuration
never silently falls back to that convention.

Preparation workflows require a literal {account} placeholder in the directory
portion of the record path. Repeated placeholders are allowed and all are
replaced exactly. The account follows the existing ASCII identity rule; do not
normalize its spelling or infer it from another field. Preserve all valid path
text and case.

Paths are relative slash-separated names, not URLs, shell expressions, or home
shortcuts. Never clean traversal into another meaning, decode percent escapes,
expand a dollar expression or tilde, or interpret arbitrary placeholders. Reject
absolute paths, empty components, dot and dot-dot components, trailing slashes,
backslashes, colons, and braces other than the exact account token.

Allow valid UTF-8 printable names, including Unicode letters and internal ordinary
ASCII spaces. Reject nonprintable characters, other Unicode whitespace, and
leading or trailing spaces within a component rather than trimming them.

Each expanded relative path has an application budget of 4096 UTF-8 bytes. Check
both paths using bounded length arithmetic before constructing the expansion;
repeated placeholders must not amplify a supplied account into a huge allocation.
This is a resource policy, not a promise about operating-system filename limits.
Drivers may still receive filesystem errors for an otherwise valid relative path.

Derive preparation from the record's directory and fixed sibling name. Reject a
record filename equal to preparation.obk.yaml, including ASCII case variants:
the initial macOS platform may use a case-insensitive filesystem. Do not case-fold
or rename any output to resolve that collision.

## Compatibility and boundaries

Legacy plan strings and existing files remain untouched. A flat custom record
layout may remain usable by older workflows; preparation support reports that
it needs a per-account directory rather than moving records or guessing names.
Any future move must be an explicit separate action that preserves existing work.

The locator is pure and lexical. It does not establish filesystem containment,
check symlinks, create directories, confirm identity, or write files. Drivers
must resolve an explicitly chosen root, enforce their filesystem safety boundary,
and check the account and plan of existing files before writing. Deriving a path
does not authorize overwriting another person's preparation or record.

The [locator contract](../../contracts/preparation/locator/contract.md) pins
default and custom layouts, exact interpolation, invalid-path rejection, bounds,
and collision handling. Driver acceptance scenarios must separately verify writes
remain local and preserve the other two artifacts.
