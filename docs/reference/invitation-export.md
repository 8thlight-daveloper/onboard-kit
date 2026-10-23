# Invitation export

`invite export` packages an existing organization plan and one person's preparation
into a new invitation ZIP. It does not send, import or open the invitation.
See [0300](../decisions/0300-explicit-invitation-export.md) and the
[command contract](../../contracts/command/invite/contract.md).

## Invocation

```sh
onboard-kit invite export \
  --plan plan.obk.yaml \
  --preparation records/newcomer/preparation.obk.yaml \
  --account newcomer \
  --out invitation.zip \
  --preaccess-help-contact owner
```

All five flags are required and each may appear only once. Values may use
`--flag=value` spelling. Optional repeated `--help-contact ID` and `--help-role ID`
select additional general help in appearance order. The designated pre-access
contact is included in that ordering. `invite --help` and `invite export --help`
print usage without inspecting supplied files.

The plan must have an identity. Preparation must match that plan and the explicit
account. It must name a known onboarding owner with a display name. The designated
helper must have a display name and a displayable Slack or business-email route;
SlackID alone is insufficient. Owner and helper may be the same person.

Designating a helper asserts that the recipient can use that route before company
access exists. The tool cannot verify this assertion. Check it before delivery;
do not assume a company-only address is reachable by the newcomer.

Required unanswered onboarder questions block export. Unaccepted defaults do not
count as answers. Explicit deferrals retain their existing semantics. Required
onboardee questions do not block handover. Optional company email, planned date,
manager, buddy and other preparation may be absent; invalid supplied values still
fail their validation rules.

## Files and preservation

Sources must be bounded local nonsymlink regular files. The output parent must
already exist and contain no symlink path components; the destination must be new.
Ambient aliases such as macOS `/var` are also rejected when explicitly present
in a selected path. Use a resolved path instead. No directory tree is created,
existing output is never replaced, and source files remain unchanged.

The archive contains four ordered payloads:

1. The canonical original organization plan.
2. The canonical preparation file.
3. A fresh identity-only record seed with account and plan reference.
4. The actual first-steps guide, including selected help.

The seed hash is SHA256 of the exact embedded canonical plan bytes, not the raw
source spelling. It is not authentication. Personal answers, outcomes and first-use
timestamps are not manufactured. Preparation is not silently converted into
question answers. See the
[archive contract](../../contracts/invitation/encoder/contract.md).

## Results and repair

Successful export prints `Invitation archive written.` and returns exit0. Deliver
the ZIP through an approved channel; export has not established access, delivery,
identity confirmation or onboarding progress. Treat its personal data accordingly.

Non-help results also write a typed JSON report to stderr. Exit2 means usage;
exit1 means another failure. Reports omit raw paths, unknown selectors, input
values and operating-system error details. Blocker reports intentionally name
declared question IDs so the onboarder can supply or deliberately defer answers.

Atomic publication is the commit point. A later display failure leaves the complete
archive in place and returns exit1 with status written when reporting is possible.
Under [0310](../decisions/0310-invitation-cleanup-failures.md), a temporary cleanup
failure also returns exit1: status written means keep the committed archive;
status refused means this export did not publish it. Temporary invitation data
may remain in the selected directory. Restore access and inspect exporter-owned
temporary data before retrying. Do not blindly remove temporary-looking files
that another export may still be using.

Guided invitation authoring, confirmed import, first-use recording, preparation
bindings and record-backed walking remain separate work.
