# Generic acceptance capture extension

Generic scenario-format extension selected under
[0480](../decisions/0480-bounded-acceptance-captures.md). Existing scenarios retain
their current rules unless extension-enabled as defined below. The runner has
no onboarding-specific knowledge.

## 1. Purpose, public API and effects

Keep exported runner.Parse([]byte), Load(path), Run(Scenario,Options), Scenario,
Options and Failure boundaries. Add Step.Capture and additive YAML scalar
predicates. External tests use only Parse and Run;
unsupported syntax is an assertion failure, not a compilation failure mistaken
for stub-red. The runner still executes the provided actual BinaryPath directly
for the literal onboard-kit placeholder, never a shell or another driver.

Read: scenario, bounded fixture-owned sandbox paths and explicitly referenced
repository fixtures, provided child stdout/stderr/exit. Write: existing sandbox
setup/cleanup and ordinary child I/O. Capture state is invocation-local only,
never global, persisted or exposed in logs. No env expansion, recursive discovery,
network, executable generation/build, product-specific hashing or schema knowledge.

## 2. Exact format and execution behavior

### Schema and bounded state

Optional step capture is a NONEMPTY mapping from unique names to exactly one:

```yaml
capture:
  ITEM: {unique-directory-entry: entries}
  HELD: {yaml-scalar: {file: data.yaml, path: [items, '0', value]}}
```

Names match `[A-Z][A-Z0-9_]{0,63}`; SANDBOX is reserved and cannot be captured.
Capture declarations across the WHOLE scenario must be unique, including across
steps. Reject duplicate YAML mapping keys, aliases/anchors, unknown fields,
empty/null/list/scalar capture shapes, multiple kinds, missing selector fields,
nonstring/empty file or directory, and invalid path/predicate shapes at Parse.
Explicitly reject a scenario with more than128 capture declarations at Parse,
before sandbox allocation/process execution. Preserve all earlier strict fields.
Capture maps are evaluated in bytewise name order, atomically committed only if
ALL captures in that step succeed. No same-step capture can refer to another
new capture regardless of order. Each capture value must be valid UTF8 \<=4096
bytes and contain no NUL. Empty YAML string is allowed; directory names cannot
be empty. Retained names+values \<=1048576 bytes; this is a defensive bound that
cannot currently be reached after128\*(64+4096), not a fabricated boundary test.

### Interpolation

Before EACH step's setup, expand supported fields against SANDBOX plus prior
successful captures. SANDBOX is the physical absolute sandbox path, after actual
EvalSymlinks of the freshly owned temporary directory, not logical PWD/TMPDIR.
Expand Run arguments EXCEPT Run[0], raw Stdin, Files map keys, expectation Files
map keys, Unchanged entries, new capture file/directory paths, and equals string
values inside yaml-scalars. Do NOT expand inline Files text, repository From/
SameAs references, JSON/Talk, ordinary stdout/stderr equals/contains, selector
path components, capture names or executable selection. Those remain literal.
This finite field list prevents accidentally templating arbitrary fixture data.
Parse admits these string fields without resolving/validating their reference
tokens; unknown or malformed interpolation is the Run-time phase below, never
an unsupported-schema compilation/Parse shortcut.

Only `${NAME}` is special. NAME must match the grammar above and already exist.
An opening `${` without its matching valid name/closing brace is invalid. `$HOME`,
`$(...)`, backticks and other dollar bytes are literal, never executed/expanded.
Replacement is ONE PASS: captured `${OTHER}` remains literal bytes and is NOT
recursively expanded; unknown original references refuse. There is no escape,
default expression, trimming, case folding or forward/self-reference. A reference
to a same-step capture fails before any setup/child/capture read. Each expanded
FIELD \<=1048576 bytes; capture value cap4096 is not a global stdin cap. Run[0]
must not contain `${`; existing literal executable behavior otherwise unchanged.
Two distinct map keys that expand to the same exact path are ambiguity/refusal,
not last-wins. Build an independent step value; do not mutate supplied Scenario.

### Native path safety and phases

A scenario using capture, `${` in a supported field, or new equals/absent scalar
predicates is extension-enabled for its whole invocation. All its sandbox file
setup, expectation, unchanged and capture paths must pass local containment
AFTER expansion, no absolute/empty/NUL or literal .. components; do not clean a
symlink away. Resolve relative to the actual owned sandbox capability. Never
follow ANY symlink ancestor or leaf, even one resolving inside the sandbox.
Require existing traversed directories and captured objects owned by effective
UID on supported POSIX hosts. Missing setup ancestors may be created as usual
only beneath proved safe native parents. Missing expected paths are permitted
only for their existing absence/unchanged semantics. FIFO/device/socket and
unsafe object types refuse before opening; no blocking pseudo-file read. Keep
repository reference semantics unchanged; they are not dynamically interpolated.
This is cooperative fixture safety, not hostile-process race-proof sandboxing,
ACL privacy or containment of arbitrary commands explicitly run by the fixture.
Unsupported native ownership/no-follow facts refuse, never fake a filesystem.

Order: expand entire step; verify paths/ambiguities; place Files; take unchanged
baseline; execute child; evaluate ALL ordinary expectations. If ANY failure,
return immediately without opening capture sources or running a later child.
Only after success perform this step's captures, validate all values/budgets and
commit state. A capture failure stops later steps; no partial winner/state.
Capture paths are guarded again at actual postchild use, not trusted from a
pre-child check. Ordinary process setup/timeout failures also precede capture.
For capture sources the prechild phase validates only expanded lexical syntax;
native source inspection/opening starts ONLY after ordinary expectations pass.
Sources may be legitimately created by that child. Do not inspect a capture
symlink early and hide a preceding ordinary expectation failure behind it.

Unique-directory-entry opens the safe owned directory, requests AT MOST2 entries
in a bounded native directory read, counting ALL entries including hidden and
nondirectories. Exactly1 entry is required, then inspect actual no-follow child:
owned directory only, not symlink/regular/special. Return its direct basename,
not full path, body or recursive descendant. No product name/pattern filtering.
An empty/multiple directory refuses regardless of read order. This2-entry request
ceiling is a contract resource obligation; black-box functional tests establish
ambiguity, not invisible syscall request size. Do not claim they prove the latter.
The ceiling is the held directory enumeration API's requested/result entries
(for example File.ReadDir(2)), not a claim that a kernel byte-buffer prefetch
contains at most2 native dirents. Unlimited os.ReadDir then sorting/counting is
not equivalent and is forbidden; internal bounded OS byte prefetch is allowed.

### Strict bounded YAML selectors

Capture YAML and every yaml-scalars file share the existing validated YAML
profile: native regular non-symlink file, read no more than1048577 bytes to prove
inclusive1048576 limit BEFORE parsing; valid UTF8, exactly one nonempty document,
no anchors/aliases ANYWHERE, duplicate decoded string keys ANYWHERE forbidden,
all mapping keys string scalars (including unselected subtrees). Reuse one actual
validated parser/profile, not a permissive capture parser or JSON roundtrip.

Path is a nonempty sequence of nonempty YAML string scalars. At a mapping each
component is an exact string key (including '00' or '0'). At a sequence require
canonical ASCII unsigned decimal index:0 or `[1-9][0-9]*`, bounded without integer
overflow. Missing mapping key/out-of-range sequence index is absence. Wrong node
kind, malformed/overflowing sequence index, malformed document or unsafe file is
an error, NOT absence. Selector never creates or normalizes data. Captures require
a present !!str scalar, including empty string; null/bool/number/timestamp/container
is not a captured string. Exact Unicode/line endings remain bytes after YAML
scalar decoding, not trim/casefold/number/string/timestamp coercion.

Each yaml-scalars item is exactly path plus ONE of:

- utc-timestamp: YAML boolean true, unchanged existing UTC timestamp profile.
- equals: YAML !!str scalar; compare exact decoded UTF8 string bytes, including
  empty and control text. Expand expected value once from PRIOR captures only.
- absent: YAML boolean true; require genuine selector absence. Present null,
  empty string, {}, [] and wrong-kind traversal all fail. An earlier missing
  key/index satisfies absence without inventing later nodes.

The yaml-scalars list remains nonempty and the ONLY file assertion. Duplicate/
unknown fields, mixed predicates, nonstring equals, false/null/string booleans,
anchors and aliases refuse at Parse. Whole-document validation occurs even for
absent predicates. The actual file itself missing is a file failure, not a
missing YAML node. Existing timestamp predicates also gain sequence selection.

### Failure privacy and exact new diagnostics

New errors return exactly one Failure, correct1-based Step, no path/value/raw
parser/OS/stdout/stderr snippets:

| Expectation | Message |
| -- | -- |
| interpolation | Scenario interpolation is invalid. |
| path safety | Scenario path is not a safe local path. |
| capture | Capture could not be established. |
| capture limit | Capture budget exceeded. |

Over4096/invalidUTF8/NUL captured value is capture limit. A missing/wrong-kind/
unsafe YAML or directory source is capture (unsafe native path is path safety).
Schema errors need nonempty sanitized Parse error, never raw captured/source
values. For extension-enabled scenarios sanitize ALL ordinary failure messages
to `Scenario expectation failed.`; retain static existing expectation categories
such as exit/stdout equals, but use `file yaml-scalars`, `file`, `unchanged`,
`setup`, `start`, `time limit` rather than expanded path-containing categories.
No captured/raw process values in Failure fields/String, timeout/start text or
debug logs. Legacy nonextension scenarios retain existing diagnostics. Do not
read a failed step's capture source to discover secrets for redaction.

## 3. Architecture and dependencies

Extend only the generic runner schema/execution context and its existing strict
YAML selector. No internal product imports or new binary mode. Invocation-local
context carries physical root, immutable successful capture strings and accounting.
Parse validates schema, whole-scenario unique names/count. Run guards current
native sources and expands a fresh step; no permanent Scenario mutation. The
existing runner subprocess path executes the real provided binary exactly as
before. Tests use a tiny owned test-executable subprocess to exercise runner
plumbing, never masquerade as product acceptance. Full product scenarios stay
language-neutral data and separately require actual binary execution.

## 4. Independent fixtures and wrong-feature verification

See the capture tests under acceptance/runner. Literals are independently authored,
not generated by the implementation. Actual native symlinks/directories and a test process
provide positive reaching facts. Every positive runner invocation has an external
owned proof file changed by the child, rejecting a runner stub that returns nil
without executing. These are runner conformance tests, not product behavior tests.
Schema refusal against the previous runner and proof-file failures against an
execution stub are distinct checks. The actual binary scenario separately checks
record-backed ownership; fixture subprocess tests alone do not prove that flow.
