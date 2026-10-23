# Preparation reviewer contract

Component preparation/reviewer, dependency position6. This is read-only semantic
review, not a draft editor, saved approval or invitation export-readiness check.

Authorities: ADR0015, ADR0020, ADR0025, ADR0055, ADR0090, ADR0100, ADR0110,
ADR0120, ADR0140, ADR0150, ADR0160, ADR0180, ADR0280, ADR0300 and ADR0330; preparation
format; actual preparation validator, question finder and firststeps assembler
contracts. ADR0330 supplies the durable bounded composition decision.

## API and trusted inputs

New() constructs the actual preparation validator, question finder and firststeps
assembler, with no injection or alternate algorithm. Review(context, originalPlan,
proposedPreparation) returns exactly four separate fields: PreparationProblems,
OnboarderBlockers, FirstSteps and FirstStepsProblems. API-owned context/result
types live in this contract; no saved model field is added. There is no ready,
exportable, approved or changed field; no returned preparation or edit operation.

The caller supplies an already authoring-validated, unambiguous original Plan,
one proposed Preparation and explicit Context.Account plus ordered GeneralHelp.
The plan is not a preparation-injected/effective/filtered copy. Full plan and
directory authoring validation is a caller precondition, not a job delegated to
this reviewer. This does not expand the current plan validator's documented
coverage or add omitted template checks to this leaf. Do not add a plan
validator dependency or new plan diagnostics.
Account and help are explicit caller values, never inferred from proposed data,
directory order, current login, record, question names, default or unknown fields.
HelpReference has the firststeps input meaning: exact Kind and ID, ordered and
repeated as held. There is no saved help or availability-assertion field here.

## Exact phases and field separation

1. Build a trusted preparation-validation context from explicit Account,
   original Plan.ID, all declared contact IDs, all declared team IDs and all
   declared section IDs. Include inactive/unmarked declarations; no Record
   identity. Call the real preparation validator on the proposed Preparation.
   Put its entire result unchanged in PreparationProblems. The current validator
   reports errors only. If the result is nonempty, stop: no finder/firststeps
   operation is required, and OnboarderBlockers and FirstStepsProblems are empty
   and FirstSteps is exactly empty. Do not return a usable partial review or
   inspect later input to add extra help/path/required-question diagnostics.
2. When preparation validation succeeds, obtain exactly the actual finder's
   Blockers(originalPlan, onboarder), in its own order, as OnboarderBlockers.
   Required onboardee-only questions do not enter it. A Default is not an answer;
   an intentional Deferred retains the current satisfying semantics. Do not sort,
   deduplicate, infer, synthesize or convert these blockers into model problems.
3. Independently call actual firststeps Assemble with the same explicit account,
   same ordered help, complete unchanged original Plan and proposed Preparation.
   Return its exact text as FirstSteps and all its diagnostics unchanged and in
   order as FirstStepsProblems, including errors or warnings. The Blockers list
   does not suppress this call. Firststeps failure does not erase prior blockers,
   move its errors to PreparationProblems or create a second synthetic error.

This composition preserves firststeps's own phase/refusal behavior: invalid help
or result layout gives exactly empty text with its real diagnostics; successful
Bootstrap gaps may produce nonempty guide plus warnings. Its repeated internal
preparation validation is normal composition, not permission for this wrapper
to merge duplicate diagnostics or copy its renderer. Observable call count is
not a contract. Never invoke export Run as a dry run or compute export gates.

Missing owner, company email, dates, manager, buddy, route or optional arrangements
may coexist with a valid guide containing honest missing-source wording. A useful
preview is not invitation readiness, writable output, archive-size validation,
confirmed identity, reachable contact, accepted edit, first use or verified work.
The existing export's additional gates are intentionally not duplicated here.

Preparation Team may affect firststeps general-help role lookup under its own
contract, but never answers or selects the original plan's team variable or
Bootstrap activity. No private plan copy, default acceptance, prerequisite edge,
arrangement auto-creation, resource expansion or record-answer join is allowed.

## Diagnostics and result ownership

No new model diagnostic code or message exists. Preserve each dependency's code,
severity, message, Question, Section and Detail values. Do not print private
account, name, email, note, answer, arbitrary help selector or proposed value in
synthetic diagnostics. Preparation messages and some gap messages have only
their existing nonblank/privacy guarantee, not new literal wording. Firststeps
fixed help diagnostics retain their literal message and exact detail keys.

Return new owned top-level problem/blocker slices and Detail maps independent
of input, earlier/later calls and other result fields. Immutable FirstSteps text
and strings are values. Nil/empty lists are equivalent. No preparation/plan input
slice, map, pointer, unknown data or context-help item may be modified, even
temporarily; no trimming, normalization, target-date rewrite or hidden default.
Result edits cannot change another result or a later call. Repeated calls and
concurrent valid/refused calls on one reviewer are deterministic and stateless.

## Read/write and dependencies

Reads only explicit Context, full original Plan through actual finder/firststeps
and the declaration projection needed by actual preparation validation, and
proposed Preparation through those dependencies. Writes derived Review only.
Position6 permits actual firststeps assembler5, question finder3, preparation
validator1, model and this own contract. No editor, record/joiner, export driver,
plan encoder/validator, archive, filesystem, environment, clock, network, Git,
process, action, saved session, consent marker or standard-help renderer.

## Four parts and independent oracle provenance

api.go and constructor-only wire_test.go define the language binding. Fixtures
under testdata name sources for every semantic case and every literal guide.
Most guide bytes are copied from existing firststeps normative goldens; derived
guides change only independently specified literal help/Bootstrap paragraphs.
No expected guide or diagnostic is obtained from product implementation calls.
Tests compare each of the four output fields, not just absence of an error or
agreement between two implementations. Input snapshots are independently parsed.

The nil stub must fail every case with nonempty guide, blockers or diagnostics
by assertion; empty/property controls may pass. A compiler or malformed-fixture
error is not red evidence. Contract/test and implementation authorship must stay
independent under ADR0025.

| Plausible wrong implementation | Independent intended discriminator |
| --- | --- |
| Always return zero or only preparation validity | Literal baseline, guide and blocker cases |
| Require owner/email/help route or return export-ready semantics | Missing-optional minimal guide and separate four fields |
| Run later phases despite invalid preparation or move diagnostics | Wrong-identity and invalid-arrangement precedence with would-be blockers/help errors |
| Stop at onboarder blockers or erase them on guide errors | Required-default plus valid guide and help/path-error-plus-blocker cases |
| Count defaults as answers, reject Deferred or block onboardee questions | Required default/Deferred/audience controls |
| Inject preparation team, filter unmarked context or reorder help | Original-plan blue Bootstrap with red help role and explicit order/repeats |
| Copy firststeps errors into PreparationProblems or drop warning gaps | Help/layout errors and successful Bootstrap warnings in separate fields |
| Reimplement preview or output partial guide on firststeps failure | Complete literal guide bytes and zero-text error cases |
| Normalize/repair proposed arrangements, drop unknown data or keep state | Full input snapshots, ownership edits and concurrent/repeated inputs |
