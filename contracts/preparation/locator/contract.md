# Preparation locator contract

Component: `preparation/locator`. Agent-authored contract; not human-reviewed.

## Behavior and interface

`Locate(resultsPath, account) -> Paths, problems` derives the relative record
and preparation filenames for one explicit account. The [API](api.go) owns
Paths with Record and Preparation strings; it adds no domain-model field.
The [preparation format](../../../docs/reference/preparation-format.md) supplies
the fixed sibling filename and retains the configured record filename.
[0090](../../../docs/decisions/0090-plan-and-onboarding-record.md) supplies the
record convention, and [0140](../../../docs/decisions/0140-person-specific-preparation.md)
separates each person's preparation from newcomer-owned work.

Reads: the supplied resultsPath and account strings only. Writes: none. Returns
only relative slash-separated paths or problems; reads or changes no plan,
preparation body, record, or unknown-field data. Every call is deterministic,
leaves arguments and previous results unchanged, and holds no state between
calls. No filesystem, root discovery, environment, clock, process, network,
archive inspection, file creation, authorization, or identity inference.

The path restrictions below are agent-decided preparation-flow policy in
[0210](../../../docs/decisions/0210-person-specific-preparation-paths.md), under
[0100](../../../docs/decisions/0100-behavior-judgment.md)'s P2, P3, P4, and P7
and the rejection direction of [0110](../../../docs/decisions/0110-reject-ambiguous-plans.md).
They do not change the interpretation of legacy configuration by other flows,
move an existing record, or certify filesystem safety. A driver must separately
choose and validate its root and reject filesystem symlinks before any write.

## Resolution phases

Stop at the first failed phase. A failure returns zero Paths, containing two
empty strings, and exactly one error. Never return a partial path or substitute
the convention after rejecting a nonempty configured value.

1. Account must match the complete ASCII pattern
   `[A-Za-z0-9][A-Za-z0-9._-]*`, exactly as in the
   [preparation validator](../validator/contract.md). Do not trim, case-fold,
   infer, or substitute an account. Invalid account yields preparation.account
   at account before inspecting resultsPath.
2. If resultsPath is exactly empty, use the path convention
   `records/{account}/record.obk.yaml`. This is storage configuration, not an
   answer or an accepted variable default under
   [0120](../../../docs/decisions/0120-defaults-require-acceptance.md).
   Otherwise inspect the configured value exactly as supplied. Reject invalid
   syntax below with preparation.results-path at resultsPath.
3. Require at least one literal `{account}` token in a directory component,
   before the last slash. Otherwise return preparation.account-directory at
   resultsPath. An account only in the filename cannot isolate the fixed
   preparation sibling for different newcomers. A legacy flat path therefore
   fails preparation rather than sharing preparation or moving the record.
4. Both expanded paths must fit an application budget of 4096 UTF-8 bytes each.
   Before constructing either expanded string, safely calculate the sizes of
   the substituted record path and its fixed preparation sibling. Reject either
   excess with preparation.path-too-long at resultsPath. Count bytes, not runes;
   never truncate. Avoid multiplication or addition overflow. A long valid
   account repeated in a template must not cause a large expanded allocation.
   This is an application resource budget, not an operating-system path limit.
5. Replace every literal `{account}` token with the explicit account, including
   repeated tokens and tokens in the filename. Reject a resulting final
   filename equal to `preparation.obk.yaml` under ASCII case-insensitive
   comparison, with preparation.path-collision at resultsPath. This catches
   the actual sibling collision on the initial macOS platform; do not otherwise
   fold case or normalize text.
6. Return the substituted path as Record. Replace its final component with
   exactly `preparation.obk.yaml` to obtain Preparation. Preserve every other
   character and directory component exactly. Do not force a record extension,
   append directories, clean the path, or invent another record location.

Repeated tokens are ordinary substitutions, not ambiguous identities. A token
may occupy part or all of a component. For example, `people/person-{account}`
is an account-specific directory. An account such as `A_1.b-2` remains exact.

## Path syntax

The configured template must be valid UTF-8 and a relative slash path with a
nonempty filename. Split on the literal `/`. Every component must be nonempty,
must not equal `.` or `..`, and must not begin or end with ordinary ASCII space.
This rejects leading or repeated slashes, trailing slashes, and traversal
segments anywhere rather than normalizing them. A component such as `a..b`
is ordinary text and remains allowed.

The only placeholder is the exact literal `{account}`. After removing all
such tokens for character inspection, any remaining `{` or `}` is invalid.
Reject colon, backslash, every control or nonprintable Unicode character, and
every Unicode whitespace character except internal ordinary ASCII space.
Printable means the categories accepted by Go unicode.IsPrint; whitespace
means unicode.IsSpace. In particular, tabs, line breaks, NUL, DEL, C1 controls,
nonbreaking spaces, line separators, paragraph separators, and zero-width
formatting characters fail. Unicode letters, combining marks, printable
symbols, and internal ordinary ASCII spaces remain exact, with no Unicode
normalization.

Colon excludes drive-prefixed and URI spellings. Backslash is never a separator.
Percent encodings, dollar signs, and tildes are literal filename characters:
never decode them, expand them, run a shell, or consult the environment. Reject
unsupported brace placeholders even when the account token is also present.
Do not infer filesystem length limits, file existence, case-sensitive volumes,
or aliases from lexical text. Those are separate driver concerns.

## Problems and evidence

Each problem has severity error, an explanatory nonblank message identifying
the rule and how to repair it, empty Question and Section, and Detail containing
only the fixed path key specified above. Messages never echo the supplied
account or configured path. Empty or nil problems both mean success. Diagnostic
maps do not alias another call's result.

[cases.json](testdata/cases.json) holds hand-specified inputs, exact output paths,
and code/path expectations, with a source for each case. [locator_test.go](locator_test.go)
checks those independent expectations, determinism, diagnostic isolation,
privacy markers, and calls with changing accounts and paths. Supplemental Go
cases supply invalid UTF-8, which JSON cannot represent. [wire_test.go](wire_test.go)
binds only to the independently authored component's New constructor. A rebuild
first runs these tests against an empty stub; every output fixture must fail.

The serial allocation probe rejects a template whose expansion would exceed
32 MiB while allowing less than 8 MiB of working allocation. This independently
checks preflight rejection without asking an intentionally broken implementation
to allocate gigabytes during mutation testing.

The contract does not implement preparation editing, readiness, archive entry
selection, root selection, file ownership checks, or any filesystem write.
Those boundaries need their own contracts and binary acceptance scenarios.

## Wrong implementation probes

| Defect | Fixture that must reject it |
| -- | -- |
| Return empty paths and no problems | Every manifest fixture |
| Ignore configured path and always use the convention | custom-record-filename |
| Fall back after rejecting nonempty input | whitespace-is-not-absent, traversal-directory |
| Put preparation in a fixed top-level account directory | custom-record-filename |
| Accept a shared preparation directory | account-only-in-filename, no-placeholder |
| Replace only the first token | repeated-account-tokens |
| Reject repeated tokens or tokens within a component | repeated-account-tokens, embedded-directory-token |
| Clean dot, traversal, or empty components | dot-directory, traversal-directory, doubled-slash |
| Trim or fold input | component-leading-space, exact-account-case |
| Restrict paths to ASCII or ban every space | unicode-and-internal-spaces |
| Treat backslash as a separator or expand environment text | windows-separators, literal-shell-and-percent-text |
| Allow an unsupported token beside the known token | unsupported-placeholder-with-account |
| Check collision case-sensitively or before substitution | collision-uppercase, collision-after-substitution |
| Inspect path before rejecting account | account-guard-before-path |
| Skip the second account character or exclude valid ASCII endpoints | invalid-second-account-character, uppercase-interior-account, uppercase-last-account, last-digit-account |
| Reject a prefix or same-length nonmatching sibling name | sibling-prefix-is-not-collision, same-length-name-before-sibling, same-length-name-after-sibling |
| Accept invalid UTF-8 as replacement characters | TestInvalidUTF8 |
| Count runes, permit only one oversized sibling, or amplify before checking | TestExpandedPathBudget |
| Expand oversized paths and only then report the length error | TestRejectsAmplificationBeforeAllocation |
| Share diagnostic maps or keep a previous account | TestDiagnosticsDoNotShareState, TestCallsDoNotKeepAccountOrPath |
