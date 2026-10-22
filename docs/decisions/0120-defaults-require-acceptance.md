# 0120. Defaults require acceptance

Status: accepted

Clarifies [0045](0045-defaults-are-suggestions.md) and principle P2 in
[0100](0100-behavior-judgment.md). Supersedes the directory-default fallback
in [0080](0080-directory.md).

## Context

0045 makes defaults suggestions, but 0080 allows the resolver to use a valid
directory default when an answer names an unknown id. That exception silently
turns a suggestion into an answer. The user confirmed that no variable type
should have this exception.

## Principle

A default is a suggestion, not an answer. Only a person's explicit acceptance
makes it an answer. The system must never silently use a default as an answer,
regardless of variable type, audience, or scope.

## Decision

- Offer a default for a person to select. Selecting it stores an answer under
  the normal ownership rules. An explicit action accepting several suggestions
  is allowed by 0045; merely opening, saving, or assembling a plan is not
  acceptance.
- Resolve content and evaluate conditions from stored answers, never from
  unaccepted defaults. A missing, deferred, or invalid answer does not authorize
  substitution of a default. Apply the existing rules for that answer state.
- Apply the same rule to contact, role, and team variables. In particular,
  an unknown directory id must not cause the resolver to select a different id
  from the variable's default. Keep the existing directory validation and
  multiple-answer rules.

A known answer supplied by an authorized person is still an answer, even when
it happens to equal a default. This principle does not require the newcomer to
approve organizational facts again or change who may supply an answer.

Directory lookup is not default acceptance. Looking up a selected role's
holders, including the organization-wide holders when no team override applies,
still follows 0080. The selected role must come from an answer, not a suggestion.

## Verification and delivery

Pin scenarios where an unanswered variable has a valid default, a directory
answer names an unknown id while its default names a known id, and an answer
is deferred while a default is available. None may use the default. Cover both
audiences and ordinary and directory variable types. As a control, explicitly
accepting a suggestion must store and use that answer; changing the declared
default later must not change it.

This records product authority, not completed enforcement. Align component
contracts, independent tests, and implementation separately under
[0025](0025-contracts-and-tests.md). It changes no stored format, ownership rule,
or accepted mutation list.
