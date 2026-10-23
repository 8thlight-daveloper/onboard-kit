# Catalog command contract

Component command/catalog. New takes no arguments. Run(args, stdout, stderr)
receives tokens after catalog. This read-only driver uses the real catalog reader.

Authorities: ADR0005, ADR0015, ADR0020, ADR0025, ADR0100, ADR0110, ADR0120,
ADR0140, ADR0150, ADR0320 and the catalog reader contract. No domain-model change.

## Read-only operations and syntax

The only operations are list, show --id ID, and export --id ID. list takes no
flags or positional arguments. show/export require exactly one nonempty --id
value, using --id VALUE or --id=VALUE. In separate spelling, a following token
beginning -- is a missing value; equals spelling permits that literal value,
which then undergoes exact reader lookup. No trimming or case folding.

Unknown/missing operation, unknown flag, repeated id even when identical, empty
value, positional token and bare -- are usage errors. No --out, --all, --plan,
--account, --templates, selector, state-root, overwrite or implicit operation.
Parse all tokens before lookup or writing a result.

Initial --help, or --help immediately after a recognized list/show/export
operation, takes precedence over remaining tokens. Unknown-operation --help
remains a usage error. Help succeeds with exit0, empty stderr and exactly:

```text
Usage: onboard-kit catalog list
       onboard-kit catalog show --id ID
       onboard-kit catalog export --id ID

list prints JSON inventory. show displays complete catalog content.
export prints section YAML or preparation/tool-help Markdown; it writes no files.
```

The final line ends in one LF. Help performs no reader lookup or file inspection.

## List

Call the actual reader List and emit one JSON array followed by one LF. The
array contains all eighty summaries in reader ID order, with exactly the seven
lowerCamel keys id,title,category,goal,result,carrier,contentPath per object.
Values equal the independently pinned inventory. JSON object key order and
interior JSON whitespace are free; no prefix/banner or second JSON value is
allowed. The output begins with [ and ends with ] followed by LF; no extra LF.
No content bytes, implicit active flag, dependency or account is added.

## Show

Read the exact ID through the real reader. On success write these blocks:

```text
# TITLE

ID: `ID`
Carrier: `CARRIER`
Category: `CATEGORY`
Goal: GOAL
Result: RESULT
Content: `CONTENTPATH`

## Content

CONTENT
```

Replace metadata with exact trusted Summary fields. Each metadata row is one
line, in the shown order, with no blank lines between rows. For preparation and
help, CONTENT is exact original Markdown, without indentation, extra fences,
parsing or a second final newline. For section-template, wrap exact YAML in a
backtick fence with yaml as its opening language tag. Fence length is max(3,
one plus the longest consecutive backtick run anywhere in content); the closing
fence is identical without a language tag. There is one LF after the opening
fence, the original content already ends in LF, and closing fence plus one LF.
The full show result ends in exactly one LF. No live personal context, renderer,
placeholder resolution, secret or guessed command is injected.

This pins observable output for the immutable catalog, not an injected arbitrary
content API or mandatory internal fence algorithm. If a simpler implementation
emits identical safe bytes for every frozen asset, these tests do not invent
alternate payloads merely to distinguish it. Future content changes require
independently reviewed output/digest updates before widening that profile.

## Export

Read the exact ID through the real reader and write exactly Entry.Content, with
no wrapper, BOM, reformatted YAML, status text or additional final LF. Carrier
decides whether these existing bytes are Section YAML or guidance/help Markdown.
No output file is created. Preparation guidance must not become a saved
preparation instance, and tool help must not become an executable section.

## Error and writer behavior

Normal list/show/export and help success return0 with empty stderr. Usage returns2;
unknown exact ID returns1. Both write no stdout and one private JSON object plus
LF to stderr, containing exactly code and message:

| Code | Message |
| --- | --- |
| catalog.usage | Use catalog list, show --id ID, or export --id ID. |
| catalog.unknown-id | Choose an exact ID from catalog list. |
| catalog.result-output | The catalog result could not be fully displayed. |
| catalog.help-output | The catalog help could not be fully displayed. |

No raw selector, argv, path, underlying writer error or reader Detail is echoed.
JSON key order/interior whitespace are free; no extra keys or second value.
Unknown ID maps to the fixed command report, not a printable dependency object.

Check every supplied writer error and short write. A stdout failure for a normal
result returns1 and attempts catalog.result-output on stderr; a help stdout
failure returns1 and attempts catalog.help-output. A stderr error/short write on
any refusal returns1, including usage that ordinarily returns2. No report can be
promised to an unwritable stderr. Partial output on writer failure is allowed;
silent exit0, infinite retry or retry through another output destination is not.
An unused stderr writer is not itself a failure on a successful stdout-only call.

## Access and purity boundary

Reads argv, the actual compiled catalog reader and supplied writer results only.
Writes supplied stdout/stderr only. No filesystem inspection or mutation,
directory, environment, cwd, optional template input, network, process, clock,
download, organization/team selection, plan/preparation/record state or outcome
operation. Construct dependencies internally with no injection. Repeated and
concurrent calls on one command instance retain no mutable state. Preserve the
caller's args slice and strings. Existing custom-template commands are unchanged.

## Four-part evidence and proposed wrong behavior

api.go plus constructor-only wire_test.go define the interface/binding. Fixtures
pin metadata, malformed argv, exact reports, help and complete show content.
Export digest fixtures independently pin all eighty normative content bytes;
they are not obtained from the command or reader implementation. All fixtures
cite source documents or independently authored normative content. Tests may not
import another product implementation or construct judging values by calling it.

A nil command stub must fail all meaningful nonempty output/refusal cases by
assertion. Property controls may pass. Broken variants, stub-red execution,
independent builds and shipped-artifact acceptance remain unverified until run.

| Plausible wrong implementation | Intended discriminator |
| --- | --- |
| Return no output, a subset or plain ID list | Exact full JSON inventory |
| Emit section YAML for prep/help or add export banners | All-carrier exact content size/digests |
| Omit metadata/content or use an unsafe fixed fence | Literal section/prep/help show goldens and content fence oracle |
| Infer ID, trim/case-fold it or accept duplicate scalar flags | Literal argv/refusal matrix and exact unknown inputs |
| Honor help only after parsing remaining tokens | Initial/immediate recognized help precedence controls |
| Honor help after an unknown operation | Unknown-operation usage case |
| Print private selector or OS/writer errors | Fixed exact two-key reports with private-marker inputs |
| Ignore short writes/output errors or retry forever | Deliberately short/error writers on results/help/refusals |
| Mutate args, retain state or change another call's output | Input snapshots, shared-instance repeats and concurrent calls |
| Load templates from checkout or change custom-directory loading | Exact artifact outside checkout and historical custom-template acceptance |
