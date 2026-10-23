# First steps assembler contract

Status: agent-authored contract; implementation pending.
Component: `firststeps/assembler`. Package: `assembler`.

Assemble one deterministic Markdown guide around the actual bootstrap fragment.
The guide explains the three separately owned invitation artifacts before access
exists. It does not open an invitation, authorize an import, or certify readiness.
Authority: ADRs 0090, 0140, 0150, 0180, 0240, 0250, and 0260; the current first-steps
reference; the document assembler, preparation validator, and locator contracts.
[0280](../../../docs/decisions/0280-first-steps-help-and-sources.md) fixes the source and dependency boundary.

## Interface and dependencies

`Assemble(context, originalPlan, preparation) -> (Markdown, problems)`.
`New()` has no arguments. See `api.go`; the draft binding returns empty results.
Position 5 permits model, this contract, and the actual document assembler at
position 4, preparation validator at position 1, and preparation locator at
position 1. No record, joiner, encoder, hasher, action runner, or driver dependency.
No alternative dependency injection or caller-authored Markdown fragment.

Reads: Context.Account and GeneralHelp; Plan.ID, URL, ResultsPath,
Provenance.GeneratedOn, Directory, the settled Text organization answer; all
plan fields read by the actual Bootstrap assembler; and Preparation fields read
by the actual preparation validator. Of validated preparation, display only
Recipient.OnboardingOwner and use Recipient.Team for help-role overrides.
No recipient name, company email, planned date, work location, time zone,
arrangement, unknown data, or record content is copied into standard help.
Writes: no argument, plan field, artifact, environment, clock, process, network,
file, answer, outcome, hash, first-use event, or outside state. Returns text and
fresh problems only. No saved state, including concurrent calls on one instance.

## Explicit context and trust

Context.Account is the explicit requested invitation account supplied by the
driver, never inferred from preparation or an address. Account confirmation
remains a later import action: rendering this account is no confirmation.

GeneralHelp is an ordered list explicitly selected by the onboarder for this
guide. Each entry identifies exactly one directory contact or role by Kind
`contact` or `role` and exact ID. A caller must obtain this selection from an
explicit trusted input, not an arbitrary unknown field, the first directory
entry, every directory entry, a title match, or a guess. There is currently no
saved model binding for this selection. A future driver must either supply
explicit input or leave the list empty and show the missing-help wording.
This projection adds no stored schema, synchronization or persistence rule.

Plan directory identities must already be unambiguous under full authoring
validation. This component is not an alternative to the full plan validator.
Bootstrap still enforces its own complete structural refusal on every call.
No plan URL or contact route is fetched, trusted as executable, or made into an
instruction to download or run anything.

## Phases and problems

Stop at the first failed phase; return exactly empty text and no partial guide.

1. Call actual preparation validation using a context derived from original
   Plan.ID, explicit Context.Account, all declared contact/team/section IDs, and
   no record identity. Return its full diagnostics unchanged on failure.
2. Validate GeneralHelp in held order. For each distinct (Kind,ID), report only
   its first index when Kind is unknown or the exact ID is absent from that
   namespace. One `firststeps.help-reference` error per invalid distinct entry;
   empty Question and Section; Detail exactly `field: generalHelp`, `item: N`.
   Message: `Choose an existing contact or role for this help item, or remove it.`
   Do not echo the supplied ID. Valid repeats are ordinary data, not identities.
3. Call actual preparation locator with original Plan.ResultsPath and explicit
   Context.Account. Return its full diagnostics unchanged on failure. No
   cleaning, partial destination, fallback after rejecting a configured path,
   root inference, or filesystem guarantee.
4. Call actual `AssembleBootstrap` with the complete original plan unchanged.
   If it returns any error, return exactly empty text and all its diagnostics
   unchanged. Do not prepend wrapper gaps, warnings, or help on a refused call.
   Otherwise retain the exact fragment bytes and all returned warnings in their
   order. Never call ordinary Assemble and slice its Markdown, duplicate its
   renderer, filter the plan, inject preparation as answers, or resolve omitted
   sections. Wrapper omissions are labeled in prose and add no warning problems.
5. Compose the fixed guide below and the exact fragment. Nil and empty problem
   lists are equivalent. Successful output always contains all twelve topics.

Preparation validation is deliberately not a readiness check. Missing email,
owner, general help, dates and arrangements do not alone refuse this preview.
Successful output is previewable help, not permission to send or import an
invitation. Sending still requires a future driver readiness contract.

## Standard help and source values

`testdata/minimal.md` is the literal standard wording. Its twelve level-two
headings and order are normative. It has one level-one title, no review banner,
and exactly one final newline. Paragraphs and list items are not wrapped.
Every source substitution is a whole displayed value, never a template that is
evaluated again. The standard wrapper never prints executable commands or an
invented download URL. Bootstrap authored commands remain exactly the renderer's.

- Organization: settled nonempty Text at Answers.Plan["organization"], otherwise
  `Organization not specified.` Defaults, flags, lists and deferred values are
  not organizational answers. Display as `Organization: VALUE`.
- Onboarding owner: exact Recipient.OnboardingOwner contact resolved from the
  directory, otherwise `Not assigned. Ask the person who sent this invitation
  to name an onboarding owner and provide a contact route.` Repeat this same
  value in the owner line under Welcome and Who to ask. Missing owner is not
  silently replaced by a buddy, manager, first contact, or role holder.
- Account: explicit Context.Account, after validation, in an inline code span.
  Show only in the account
  line, fixed archive entry paths, and locator-derived destination paths.
- The four archive names are exactly ADR0240's names. Destination paths come
  from the locator and are distinct from archive names. Display that they are
  relative to the root explicitly selected at import and that state/root
  discovery and identity confirmation belong to the import flow. Never claim
  archive paths are directly extracted or that rendering saved anything.
- Plan location: exact nonempty Plan.URL or `Not specified. Ask the onboarding
  owner where the shared plan is kept.` Display text only; no clone command,
  rewritten URL, guessed repository root, or automatic fetch.
- Plan ID: exact nonempty Plan.ID. Preparation identity validation rejects an
  absent/mismatching plan identity before output. Generated date: held nonempty
  Provenance.GeneratedOn, otherwise `Not specified.` Never use today's date,
  plannedStartDate, legacy startDate, or a first-use guess.
- Empty successful bootstrap fragment: exactly `No pre-access activities are
  listed in this plan. Ask the onboarding owner what you need before access
  works.` A nonempty fragment replaces that paragraph byte-for-byte under
  `## Do these first`, with one blank line before and after it. Do not add an
  empty gap table, alter flat headings, deduplicate existing warnings, or infer
  work from an early target date.
- General help: display distinct (Kind,ID) entries once in first-held order,
  one list item each. Empty list is `General help contacts not specified. Ask
  the onboarding owner to provide a help route that works before company access.`

Contact display is `Name, Slack`, else `Name, Email`, else `Name` as the assembler's
directory convention. A blank Name means empty or consisting entirely of
Unicode White_Space characters, as tested by Go unicode.IsSpace. It is displayed
as `Name not supplied` rather
than an empty label. No route gives the suffix ` (contact route not supplied)`;
SlackID is not an inferred route. Roles use Recipient.Team's explicit known team
holder override when present, including an empty override; otherwise their own
holders. Empty holder set is `Role name (vacant)`. Named holders are resolved
as contacts in held order and joined with `; `. A role's blank Name is `Unnamed
role`. Full authoring validation owns invalid directory holder references and
duplicate identities; inputs outside that precondition are not given a winner.
General-help selection never changes bootstrap team resolution: its assembler
still receives the original plan answers, directory, and variables.

Wrapper source values are literal display text: escape backslash and the ASCII
Markdown punctuation backtick, asterisk, underscore, braces, brackets, parentheses,
angle brackets, hash, plus, minus, full stop, exclamation and pipe with backslash;
replace CRLF, CR, LF, U+2028 and U+2029 with one ordinary space per separator;
leave other valid UTF-8 bytes exact. Apply this to each held source field before
adding fixed labels, separators, missing-value text and vacancy/route suffixes;
those fixed words and punctuation remain as shown in the goldens. A nonblank
Name keeps its whitespace except for the specified line-separator replacement.
For displayed paths and the account, use literal inline code instead of prose
escaping: flatten the specified line separators but keep held punctuation and
backticks without adding backslashes. Use a fence longer than any held backtick
run and pad inside the fence when the text begins/ends in a backtick or space.
This rule applies only to wrapper
values, never to the actual bootstrap fragment. Wrapper inputs and directory
fields otherwise have their existing authoring validity preconditions. Route
choice uses a nonblank Slack, else a nonblank Email, with blank defined by the
same Unicode whitespace rule; do not trim held values.

## Properties and exclusions

Input values/backing storage stay unchanged. Reversing or rotating section
declarations and unselected directory declarations does not change output.
Changing returned problems/details cannot change arguments or later results.
Calls on a shared instance are deterministic and safe concurrently. Changing
preparation recipient fields not displayed, arrangements, record-shaped unknown
data, previous calls, or an unaccepted organization default does not alter
standard help when preparation remains valid.

No identity-only record seed is generated here. This component writes neither
archive bytes nor files, fetches no location, installs nothing, opens no path,
confirms nobody, performs no Git operation, records no first use, selects no
default, and claims no activity verified. Planned dates are not employment start
or observed first use. Import/start/update/outcome commands, distribution,
readiness, recovery, approval, and the complete invitation runtime remain staged.

## Fixture and wrong-adapter boundary

All expected Markdown files and diagnostic expectations are literal hand-authored
data with sources in `testdata/cases.json`; tests do not compute expected text by
calling another assembler. The first null stub must fail every nonempty guide
and every error-output case; empty controls may pass. The wrong implementations table below specifies intended rejection boundaries. The drafting author does not implement
the component in this run or later repair code judged by these new fixtures.

## Wrong implementations

| One plausible defect | Independent intended rejection |
| -- | -- |
| Always return no guide/problems | Every nonempty guide and refusal oracle |
| Print only Bootstrap without standard help | Three complete guide goldens |
| Invent a download address or pending start command | Complete literal standard wording |
| Omit preparation, or call it newcomer-owned | Four entry/ownership bullets in every guide |
| Put configured record paths into ZIP entry explanations | custom-vacant.md exact fixed archive names |
| Ignore configured destinations or normalize account case | custom-vacant.md exact locator paths |
| Fall back after invalid resultsPath | invalid-results-path-does-not-fall-back |
| Copy account from preparation before comparing identity | identity-refusal-precedes-help-and-path |
| Select the first contact as missing owner | original-plan-real-bootstrap-renderer puts helper first; missing-sources wording remains fixed |
| Show every directory contact as general help | Explicit selected helper; custom-vacant omits global support |
| Count repeated selected help references twice | bootstrap.md has one helper item |
| Echo an invalid help ID into diagnostics | help-refusal-first-distinct-index-no-leak |
| Ignore an explicitly empty team role override | custom-vacant.md shows Support (vacant) |
| Accept an organization default as an answer | unaccepted-organization-default-is-not-fact |
| Print recipient email, date, names or unknown values | non-displayed-personal-data-stays-out |
| Call planned/generated date employment start or first use | Literal help and About this plan labels |
| Filter Bootstrap before full-plan selection | Original ancestor-answer/omitted-switch fixture |
| Render ordinary Assemble and slice Markdown | Literal flat Within context and exact step/check fragment |
| Reimplement grouped selector order rather than scheduler | action-order-and-exact-parent-context-kept |
| Drop Bootstrap warning problems or rewrite placeholders | bootstrap-warning-and-onboardee-placeholder-propagate |
| Leak an unmarked ancestor instruction or private gap | PRIVATE markers and literal Bootstrap oracle |
| Treat before-day-one as Bootstrap | early-timing-does-not-create-bootstrap-work |
| Return wrapper/partial work despite inactive structural error | inactive-unmarked-structural-error-refuses-whole-guide |
| Insert wrapper values as raw Markdown | Literal heading/link/newline escaping fixture |
| Always use one backtick for paths | wrapper-path-backtick-fence-is-literal |
| Mutate input, cache diagnostics, share returned detail maps | Input snapshots, reordered/shared/concurrent calls and result-mutation checks |
