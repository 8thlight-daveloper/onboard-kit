# Additive acceptance YAML scalar expectations

This is an acceptance-runner capability, not a product record decoder.
It adds no product-model imports, clock override, scenario actions or file writes.

## Syntax and existing public API

```yaml
expect:
  files:
    records/newcomer/record.obk.yaml:
      yaml-scalars:
        - path: [firstOpenedAt]
          utc-timestamp: true
```

The existing runner.Parse and runner.Run public API remain the binding. Add an
optional yaml-scalars assertion to FileExpectation's serialized shape. Tests
must use that existing public API without naming an implementation type or new
exported symbol: the old runner must compile and fail the explicit valid-Parse
control, rather than fail to compile. The implementation's private Go storage
types are not part of this format contract.

yaml-scalars is a nonempty sequence of mappings. Every mapping has exactly two
required keys: path and utc-timestamp. path is a nonempty sequence of nonempty
YAML string scalars; utc-timestamp is exactly the YAML boolean true, not false,
null, text, a number or a timestamp. All path items are literal mapping keys:
periods, slashes, brackets, tildes, whitespace and Unicode have no special path,
JSON-pointer, trimming or normalization interpretation. Repeated paths are
allowed and both assertions are evaluated. No sequence indexing is introduced.

Reject unknown, missing, duplicate or wrong-kind keys/values in this assertion.
Reject anchors and aliases anywhere in this new file-assertion mapping/subtree.
When yaml-scalars is present, the FileExpectation mapping has exactly that one
key: it cannot be combined with same-as, json, zip or absent, even absent:false,
regardless of mapping order. Repeated outer yaml-scalars keys are invalid even
when equal. This new exclusivity does not alter any legacy-only combinations.
An empty FileExpectation {} continues to mean only that the file exists.

## Bounded data inspection

Dispatch this assertion before the legacy generic unlimited file read. Open
only a regular file in the existing scenario-root path boundary and perform
a capped read at1048576 bytes inclusive, with at most one detection byte beyond
the remaining allowance. Do not allocate based on an unchecked file size or
read an unlimited body and then compare its length. No extraction, subprocess,
network, file mutation, product-model decode or environment lookup is needed.

The inspected bytes must be valid UTF8 and contain exactly one nonempty YAML
document. Explicit document delimiters for that one document are allowed, but
a second document, including an empty one, refuses. YAML parser errors refuse.
Reject every anchor or alias globally, including on unselected nodes and unused
anchors. Reject duplicate mapping keys globally before converting any mapping
into a representation that discards one occurrence. Every mapping key anywhere
in the document must be an exact YAML !!str scalar. Numeric, boolean, null,
timestamp and complex keys are unsupported and refuse, even when unrelated to
an asserted path. Compare exact decoded strings for duplicates: plain, quoted
and escaped spellings of the same string collide. No trimming, case folding or
key coercion is allowed. Unselected values may remain any scalar or collection
kind subject to the global document rules; no product field whitelist applies.

Walk each asserted path from the document root. Each intermediate node must
be a mapping and each selected key must be an exact !!str scalar equal to that
path item. Missing keys, a sequence/scalar/null intermediate, or a non-string
selected value fail the assertion. The final node must be exactly a YAML !!str
scalar, nonempty, with the timestamp syntax and calendar rules below. The YAML
tag is checked, not a string produced by generic conversion or a permissive
decoder. In particular an unquoted YAML !!timestamp value is NOT accepted even
when it looks like the wanted instant. Quoted text or explicit !!str is valid.

## Exact timestamp predicate

The decoded string must be exactly:

```text
YYYY-MM-DDTHH:MM:SS[.FRACTION]Z
```

Digits are ASCII only. Year has exactly four digits and is0001 through9999;
month/day/hour/minute/second have exactly two digits in their displayed places.
Uppercase T and final Z are required. Fraction is either absent or a period
followed by one through nine ASCII digits. No surrounding whitespace, offset,
comma fraction, lowercase separator, extended year, single-digit hour or leap
second is accepted. Calendar/time must be real in the proleptic Gregorian
calendar: month1..12, valid day including century leap rules, hour0..23,
minute0..59, second0..59. No normalization repairs invalid text.

This checks a UTC event representation, not recency, authentication, the actual
employment start date, clock sampling count or relationship to the test machine's
time. Use independent literal positive and negative values. A new-event binary
scenario combines this predicate on the first created record with a subsequent
exact unchanged snapshot; neither preservation alone nor a printed event proves
the first stored value existed.

## Failures and compatibility

A content/predicate failure names its step and `file PATH yaml-scalars`; a missing,
unopenable or nonregular file uses `file PATH` instead. A selected-path failure may identify
the authored predicate ordinal/path. Exact English wording is not frozen. Never
dump held file bytes, rejected scalar values, parser excerpts or unselected keys/
values into any Failure.Message or String. Report a fixed useful reason instead,
for example invalid YAML, duplicated key, missing path, wrong scalar kind or
invalid UTC timestamp, without echoing input. An authored scenario filename/path
is not a license to append a parser's raw error text.

Checking does not rewrite the inspected file. The existing unchanged snapshot,
same-as, json, absent, ZIP, output, exit and sequential-step behavior remain
unchanged. The new predicate is conjunctive across its own list: every item must
pass, in supplied order. It is exclusive only within that file's FileExpectation;
other files in the same step may use all their existing assertion kinds.

## Independent fixture sources and wrong implementations

The external tests bind only existing runner.Parse/Run and the already present
runner_test fake-binary helpers. Payloads are fabricated literal YAML; no expected
value comes from a product encoder, decoder, model, timestamp producer or this
new matcher. A positive control must parse and run. An ignored predicate must
fail multiple explicit negative controls; reject-all must fail positive controls.

| Wrong implementation | Independent test family |
| -- | -- |
| Ignore unknown new assertion or always pass it | TestYAMLScalarValueAndShapeFailures |
| Reject every new assertion | TestYAMLScalarPublicParse and TestYAMLScalarLiteralPathsAndConjunction |
| Coerce YAML timestamps/numbers into text | TestYAMLScalarValueAndShapeFailures |
| Use only a permissive standard timestamp parser | TestYAMLScalarTimestampProfile |
| Keep first/last duplicate or validate only selected subtree | TestYAMLScalarGlobalDocumentValidation |
| Resolve aliases or ignore unused anchors | TestYAMLScalarGlobalDocumentValidation |
| Check just the first document or truncate overlong file | TestYAMLScalarGlobalDocumentValidation and TestYAMLScalarInclusiveByteLimit |
| Treat dots/slashes as a path language or only evaluate first item | TestYAMLScalarLiteralPathsAndConjunction |
| Rewrite inspected data or break old assertions | TestYAMLScalarCheckingPreservesBytesAndOtherAssertions |
| Echo rejected private data through errors | TestYAMLScalarFailuresDoNotExposeHeldText |

Independent source authors must not change the judging tests. Execution and
mutation outcomes are review evidence, not part of this format definition.
