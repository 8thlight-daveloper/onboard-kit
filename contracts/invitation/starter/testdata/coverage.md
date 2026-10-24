# Starter fixture scope and native evidence limits

The deterministic suite specifies fresh begin, actual record-backed read-only
session, exact identity/review/help/report bytes, no-op retention, explicit plan
choice, retained stale support, disjoint contexts/registry continuity, private
fixed roles, strict pending/claims/receipt parsing, frozen ZIP-free recovery,
receipt-last cleanup, finite security transitions and input/output precedence.
Native tests use actual filesystem objects, permissions, ACLs and processes.
No product implementation, fake filesystem or semantic oracle was used to author
these expected results. Authorship is not passing evidence.

## Explicit native conditions

- Darwin filesec tests require supported little-endian native ACL operations.
  ENOTSUP/ENOSYS are named skips; an unexpectedly malformed established fixture
  fails rather than assuming absence. Linux security fixtures use actual POSIX
  ACL xattrs. Full native ACL-transition cases are Darwin-specific.
- Alternate group preservation needs another group the caller may actually set.
  A missing or refused native group fixture is an explicitly reported gap.
- Actual two-process tests require the contract test executable to run itself
  with owned pipe streams. They do not claim hostile-owner or remote-filesystem
  locking guarantees.
- Real cross-volume ROOT/HOME publication is not established by these same-volume
  fixtures. No group atomicity, power-loss durability or hostile-owner CAS claim
  follows from any case.

## Optional observed interruption layer

Run the contract test binary with OBK_STARTER_OBSERVE_DEATH=1 to request the
Darwin observation layer. This is test-only selection, not a production fault
flag. Each intended phase must be observed while the owned child is actually
stopped before killing it. Twelve attempts per fresh, canonical-reconciliation
and retained-record profile are finite and do not guarantee that every tiny
native syscall window can be reached on a host.

The layer reports every unobserved phase separately: idle, unclaimed-empty,
claimed-incomplete, pending-no-role, after-local-plan, after-preparation,
after-session-metadata, retained-record/pending-no-role,
retained-record/after-local-plan, retained-record/after-preparation,
payload-final-controls, all-roles-no-receipt, receipt-pending and receipt-only. Until actual root-run
evidence is recorded, every one remains unverified. Seeded journal tests and
ordinary writer/source errors are not substitutes for those observations.
An out-of-order applied nonretained role is never an accepted prefix. The
retained-record family explicitly observes a no-op record while companion
publication remains unfinished; record equality is not completion.

Strict reached classes compare each frozen field/action/body against this finite
fixture profile, original independently captured native preimages/parent guards,
claimed actual stage/payload identities and permitted finite controls. Receipt
classification checks exact role names, final body hashes, actual device/inode,
security hashes and namespace binding. Partial/noncanonical schema or any fact
not proved is unclassified: resume the child normally, do not kill and label it
verified. This deliberately underreports some transient cleanup/claim-update
windows rather than trusting canonical JSON alone as ownership proof.

## Not separately isolated in this first suite

The normative limits remain mandatory, including inclusive10MiB/18MiB combined
payload,16MiB metadata aggregate,2MiB per metadata,16KiB claims,64KiB native ACLs,
4096-byte paths and inclusive8MiB display. This first suite isolates selected raw
source/pending/display excess and valid128/129registry count, not every arithmetic
counter independently. Header/claim bounds can be masked by stricter earlier
shape/native controls; do not report such a refusal as isolated arithmetic proof.
Custom-resultsPath nested namespace/case-folding collisions and real cross-volume
publication need further native/profile-specific reachability controls. These
are known coverage work, not waived requirements or accepted survivors.
The independent streams fixture isolates borrowed io.Closer preservation and
readers returning n>0 together with io.EOF or a nonEOF error at ordinary
identity/begin/exit boundaries. Those cases do not establish combined total-byte
budget/known-EOF precedence or paired errors at recovery prompts. Those distinct
conditions remain first-build hardening work; ordinary strings.Reader inputs
are not evidence for simultaneous positive-count/error returns.

The independent session-gap fixtures pin the complete nonempty held step/check
block and the full stale-support session. The independent input-gap fixtures
isolate zero-progress reads, exact action spelling and incomplete EOF at all
three ordinary interaction states with bounded output observation. These do not
establish recovery-prompt or native publication-window coverage.
