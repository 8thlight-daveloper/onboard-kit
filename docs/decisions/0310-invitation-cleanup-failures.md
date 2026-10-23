# 0310. Invitation cleanup failures

Status: accepted by agent judgment

Clarifies the publication and cleanup requirements in
[0300](0300-explicit-invitation-export.md).

## Context

An invitation can be completely published while the operating system prevents
removing its temporary directory entry. Returning success hides residual private
data. Reporting that nothing was written is also false and encourages a confusing
retry. An output-writer failure is separate from this filesystem failure.

## Decision

Check owned temporary cleanup on success and failure. If cleanup is prevented,
return exit1 and explain the remaining repair need without exposing paths,
temporary names, account values or raw operating-system errors.

Before publication, use status refused and invite.cleanup-refused. Preserve
the absent destination or anything independently created there. After publication,
retain the complete archive, attempt the honest written stdout line, and use
status written and invite.cleanup-written. Keep the command's five-key report
and dependency warnings. Cleanup failure takes priority over a secondary reporting
failure whenever stderr remains writable; an unwritable stderr cannot be promised
a diagnostic. Never remove a committed archive to make a refusal report true.

An already absent owned temporary entry counts as cleaned. Do not change directory
permissions, strip inherited access controls, alter unrelated entries or retry
indefinitely to force removal. Setting the temporary file's ordinary private mode
does not authorize changing its inherited access controls. No saved format, public
API, crash-durability guarantee or universal hostile-host guarantee is added.

## Rationale

P3 of [0100](0100-behavior-judgment.md) requires an actionable next step. P4 favors
preserving complete committed work, and P7 requires honest handling of residual
personal data. Keeping the output while reporting cleanup failure serves all
three. Ignoring cleanup, silently stripping access controls, or rolling back the
complete archive does not.

## Verification boundary

Independent tests use an owned unprivileged Darwin directory with an inheritable
deny-delete ACL. A separate capability control must first establish successful
creation, private-mode setting, writing, closing and hard-link publication while
temporary unlink is denied. The public command then receives the same environment.
Tests inspect literal ZIP payloads, metadata, unchanged inputs and the retained
temporary entry, and require the written-cleanup error with ordinary and failing
stdout writers. No product filesystem hook supplies the fault or expected result.

Unsupported platforms or demonstrated unsupported capability are skipped fault
coverage, not proof of cleanup behavior. Once the control succeeds, wrong product
behavior must fail. Prepublication cleanup refusal needs a separate independently
established filesystem fixture; the postpublication control does not verify it.
