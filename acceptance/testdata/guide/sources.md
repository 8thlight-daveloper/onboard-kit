# Guided preparation binary fixtures

These are independently hand-written data, not generated encoder output.
The author does not implement the preparation guide driver. Root must execute
the frozen scenarios and report evidence separately.

- `new-literal-values.yaml` applies ADR 0340's literal value rule: `save` is a
  name and `q` is a time-zone value, not a global action. Explicit owner-one is a
  declared directory contact in the existing preparation plan. Field order,
  omitted empty fields and final LF come from
  `contracts/preparation/encoder/contract.md` and its scalar rules inherited
  from `contracts/plan/encoder/contract.md`.
- `new-identity-only.yaml` pins explicitly saving an absent preparation's
  identity-only draft. Missing optional recipient facts and unaccepted
  organization defaults do not turn reviewed preparation save into invitation
  export readiness. No record is created for that second account.
- `cleared-recipient.yaml` copies the semantic preparation at
  `acceptance/testdata/preparation/preparation.yaml`, deliberately removes only
  companyEmail and plannedStartDate, and retains all other recipient facts,
  arrangement fields, list order and unknown root data. The exact expected
  bytes were transcribed independently using the encoder contract.
- `repaired-recipient.yaml` applies ADR 0190's scoped repair: replace both
  stale contact IDs with explicitly selected declared IDs, retain the stale
  unrelated arrangement and unknown root data, and save without claiming full
  validity. An intermediate invalid form is not applied and stays available
  for correction. Exact YAML order and scalar spelling follow the encoder
  contract, not a generated expected output.
- Scenario plans use `acceptance/testdata/preparation/plan.yaml`, already used
  by whole-binary preparation scenarios. The fixture has no supplied answer
  for its required organization variable. Review must not accept its default.
- Inline malformed records are deliberate unreadable sentinels. The guide
  must ignore and preserve them, not parse progress or create an outcome.
- Inline unchanged preparations use comments, flow mappings or empty
  arrangements to distinguish byte-preserving no-op from canonical encoding.
  The empty arrangement is decoder-readable but encoder-unrepresentable;
  ADR 0340 explicitly permits this unchanged no-op.
- Input scripts and fixed terminal statuses come from the independently
  authored `contracts/preparation/guide/contract.md`. YAML `stdin: |-` holds
  the final `save` without LF; `stdin: |` includes LF. The existing runner
  passes stdin verbatim. No scenario runner fields or matching rules change.

The runner compares stderr JSON by subset, not exact shape. Exact three-key
diagnostic shape and detailed prompt/output ordering remain component-contract
assertions. The binary scenarios pin real root dispatch, explicit script
actions, byte equality, refusal/cancellation status and artifact ownership.
