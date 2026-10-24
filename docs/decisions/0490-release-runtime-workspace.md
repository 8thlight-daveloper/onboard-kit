# 0490. Release runtime workspace

Status: agent-selected local tooling decision, not human approval.

## Context

Release verification originally placed its runtime files inside the checkout.
The real artifact passed59of60 scenarios there: catalog editing refused an
ancestor ACL in the checkout's home-directory hierarchy. This is a native
profile limitation, not permission to remove application safeguards, strip a
home ACL or skip a scenario.

Moving runtime files to the system temporary hierarchy exposed another real
condition: new directories inherited its group rather than the process's
effective group. Existing catalog publication requires effective ownership of
its new private temporary files. Ordinary0700 mode alone does not prove that.

## Decision

Create each verifier runtime workspace under the actual system temporary
hierarchy, outside the source checkout and independent of inherited temporary
overrides pointing into it. Resolve its physical absolute path and verify that
it is an empty private nonsymlink directory outside the selected repository.
Each invocation owns a distinct newly created directory.

Establish the effective group on that new empty owned directory before any
caller data or child objects exist. Never change a parent, source file or
preexisting runtime object's permissions, ownership or ACL. Do not adopt another
workspace. Creation or metadata failures remain real verification failures.

Keep the exact artifact, all scenario assertions, signature and library checks,
empty PATH and isolated child environment. This changes verifier placement,
not the application's filesystem profile. Supporting ACL-bearing ancestors
requires a separate contract decision and independent application tests.

## Evidence and limits

Independent real-filesystem tests failed the old checkout-local helper, then
passed external placement. A separate native ownership test exposed the actual
inherited-group mismatch before caller writes. The repaired helper passes both
sets without modifying their assertions or repository fixtures.

These tests do not prove absent ACLs or universal native compatibility. The
complete real-artifact release run must separately pass in the selected runtime
workspace. Local verification does not publish, sign for distribution, notarize
or prove browser-download quarantine behavior.
