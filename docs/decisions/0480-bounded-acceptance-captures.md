---
category: operations
status: accepted
approval: agent
note: An agent-selected integration candidate, not human approval.
---

# 0480. Bounded acceptance captures

## Context

An invitation opening creates a session registration and a first-use timestamp.
Whole-binary tests must use those actual values in later commands without guessing
them, adding test flags to the product or importing product implementations.
Existing static scenarios cannot express this sequence.

## Decision

Extend the generic acceptance format with invocation-local named captures,
single-pass interpolation of prior captures, and exact YAML string equality or
node absence. Keep existing scenario behavior and diagnostics unchanged when the
extension is not used. This is test tooling, not a runtime product dependency.

A capture reads either one direct directory entry or one strict YAML string
scalar. It has no knowledge of accounts, sessions, hashes or onboarding records.
The physical owned sandbox is available as SANDBOX. Substitution never executes
a shell, expands environment variables, recursively expands captured text or
selects the executable. Undefined and same-step references refuse.

Limits are explicit:128 unique names,4096 UTF8 bytes per value,1MiB per YAML file
and expanded field. Directory capture requests at most two entries and refuses
anything except one actual owned directory. All dynamic local paths are checked
after expansion, with native ownership and no-follow traversal. Captures occur
only after every ordinary expectation succeeds. No state persists between runs.

String equality, timestamps and absence use the same strict YAML parser and
mapping/sequence selector. Null, empty containers and malformed traversal are not
absence. Duplicate decoded keys, anchors and aliases are rejected throughout the
document, including unselected data.

Extension-enabled failures retain a useful static category and step number but
never include captured values, expanded paths or child output. This applies to
the whole invocation, including steps before a capture. It follows0100 P3 and P7;
redaction does not read failed-step capture sources to discover secrets.

The complete independent format contract is the
[capture reference](../reference/acceptance-captures.md).

## Verification and limits

Independent tests use an actual tiny owned subprocess and an external execution
proof file. A runner returning success without executing cannot satisfy positive
cases. This proves runner plumbing, not product behavior. Product acceptance
must separately run the actual binary through invitation opening and subsequent
record changes, preserving first use, registration and companion files.

Native functional tests cannot establish an invisible directory-read request
size; source inspection separately checks that ceiling. Cooperative fixture
safety does not contain arbitrary commands or promise hostile-race isolation.
The retained1MiB aggregate capture bound is defensive and unreachable under the
earlier per-name/count bounds; no fabricated reaching test is required.
