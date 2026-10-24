package reviewer

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"unicode/utf8"

	invitationdecoderapi "onboard-kit/contracts/invitation/decoder"
	api "onboard-kit/contracts/invitation/reviewer"
	authoringapi "onboard-kit/contracts/plan/authoringvalidator"
	plandecoderapi "onboard-kit/contracts/plan/decoder"
	planvalidatorapi "onboard-kit/contracts/plan/validator"
	preparationdecoderapi "onboard-kit/contracts/preparation/decoder"
	locatorapi "onboard-kit/contracts/preparation/locator"
	preparationapi "onboard-kit/contracts/preparation/validator"
	recorddecoderapi "onboard-kit/contracts/record/decoder"
	encoderapi "onboard-kit/contracts/record/encoder"
	joinerapi "onboard-kit/contracts/record/joiner"
	reconcilerapi "onboard-kit/contracts/record/reconciler"
	walkerapi "onboard-kit/contracts/section/walker"
	invitationdecoder "onboard-kit/internal/invitation/decoder"
	"onboard-kit/internal/model"
	"onboard-kit/internal/plan/authoringvalidator"
	plandecoder "onboard-kit/internal/plan/decoder"
	planvalidator "onboard-kit/internal/plan/validator"
	preparationdecoder "onboard-kit/internal/preparation/decoder"
	"onboard-kit/internal/preparation/locator"
	preparationvalidator "onboard-kit/internal/preparation/validator"
	recorddecoder "onboard-kit/internal/record/decoder"
	"onboard-kit/internal/record/encoder"
	"onboard-kit/internal/record/joiner"
	"onboard-kit/internal/record/reconciler"
	"onboard-kit/internal/section/walker"
)

type reviewer struct {
	invitationDecoder    invitationdecoderapi.Decoder
	planDecoder          plandecoderapi.Decoder
	authoringValidator   authoringapi.Validator
	planValidator        planvalidatorapi.Validator
	preparationDecoder   preparationdecoderapi.Decoder
	preparationValidator preparationapi.Validator
	locator              locatorapi.Locator
	recordDecoder        recorddecoderapi.Decoder
	reconciler           reconcilerapi.Reconciler
	joiner               joinerapi.Joiner
	encoder              encoderapi.Encoder
	walker               walkerapi.Walker
}

type inspected struct {
	incoming    api.Incoming
	plan        model.Plan
	preparation model.Preparation
	record      model.Record
}

func New() api.Reviewer {
	return reviewer{invitationDecoder: invitationdecoder.New(), planDecoder: plandecoder.New(), authoringValidator: authoringvalidator.New(), planValidator: planvalidator.New(), preparationDecoder: preparationdecoder.New(), preparationValidator: preparationvalidator.New(), locator: locator.New(), recordDecoder: recorddecoder.New(), reconciler: reconciler.New(), joiner: joiner.New(), encoder: encoder.New(), walker: walker.New()}
}

func (receiver reviewer) Inspect(archive []byte) (api.Incoming, []api.Diagnostic) {
	value, problems := receiver.inspect(archive)
	return value.incoming, problems
}

func (receiver reviewer) inspect(archive []byte) (inspected, []api.Diagnostic) {
	files, problems := receiver.invitationDecoder.Decode(archive)
	if fatal(problems) {
		return inspected{}, wrap("inspect.envelope", problems)
	}
	warnings := wrap("inspect.envelope", problems)
	plan, problems := receiver.planDecoder.Decode(files.Plan)
	if fatal(problems) {
		return inspected{}, wrap("inspect.plan-decode", problems)
	}
	warnings = append(warnings, wrap("inspect.plan-decode", problems)...)
	if !admit(plan, model.Record{}, nil, len(archive), true) {
		return inspected{}, resource("inspect.plan-admission", "plan")
	}
	problems = receiver.authoringValidator.Validate(plan)
	if fatal(problems) {
		return inspected{}, wrap("inspect.plan-authoring", problems)
	}
	warnings = append(warnings, wrap("inspect.plan-authoring", problems)...)
	problems = receiver.planValidator.Validate(plan)
	if fatal(problems) {
		return inspected{}, wrap("inspect.plan-validation", problems)
	}
	warnings = append(warnings, wrap("inspect.plan-validation", problems)...)
	if plan.ID == "" {
		return inspected{}, own("inspect.plan-identity", "plan-id", "plan.id")
	}
	preparation, problems := receiver.preparationDecoder.Decode(files.Preparation)
	if fatal(problems) {
		return inspected{}, wrap("inspect.preparation-decode", problems)
	}
	warnings = append(warnings, wrap("inspect.preparation-decode", problems)...)
	record, problems := receiver.recordDecoder.Decode(files.Record)
	if fatal(problems) {
		return inspected{}, wrap("inspect.record-decode", problems)
	}
	warnings = append(warnings, wrap("inspect.record-decode", problems)...)
	if !admit(plan, record, &preparation, len(archive), true) {
		return inspected{}, resource("inspect.payload-admission", "invitation")
	}
	hash := sha256.Sum256(files.Plan)
	reference := model.PlanReference{ID: plan.ID, Hash: hex.EncodeToString(hash[:]), URL: plan.URL}
	if diagnostics := seedProblems(record, files.Account, reference); len(diagnostics) > 0 {
		return inspected{}, diagnostics
	}
	paths, problems := receiver.locator.Locate(plan.ResultsPath, files.Account)
	if fatal(problems) {
		return inspected{}, wrap("inspect.locator", problems)
	}
	problems = receiver.preparationValidator.Validate(preparationContext(plan, files.Account), preparation)
	if fatal(problems) {
		return inspected{}, wrap("inspect.preparation-validation", problems)
	}
	if len(files.FirstSteps) == 0 || !utf8.Valid(files.FirstSteps) {
		return inspected{}, own("inspect.first-steps", "first-steps", "firstSteps")
	}
	incoming := api.Incoming{Archive: archive, Files: files, PlanID: plan.ID, Reference: reference, Paths: api.Paths{Record: paths.Record, Preparation: paths.Preparation}, Warnings: warnings}
	return inspected{incoming: owned(incoming), plan: plan, preparation: preparation, record: record}, nil
}

func preparationContext(plan model.Plan, account string) preparationapi.Context {
	context := preparationapi.Context{PlanID: plan.ID, Account: account}
	for _, contact := range plan.Directory.Contacts {
		context.Contacts = append(context.Contacts, contact.ID)
	}
	for _, team := range plan.Directory.Teams {
		context.Teams = append(context.Teams, team.ID)
	}
	for _, section := range plan.Sections {
		context.Sections = append(context.Sections, section.ID)
	}
	return context
}

func seedProblems(record model.Record, account string, reference model.PlanReference) []api.Diagnostic {
	checks := []struct {
		bad  bool
		path string
	}{
		{record.Account != account, "record.account"}, {record.Plan.ID != reference.ID, "record.plan.id"},
		{record.Plan.Hash != reference.Hash, "record.plan.hash"}, {record.Plan.URL != reference.URL, "record.plan.url"},
		{record.StartDate != "", "record.startDate"}, {record.FirstOpenedAt != "", "record.firstOpenedAt"},
		{len(record.Answers.Plan) > 0, "record.answers.plan"}, {hasSectionAnswers(record.Answers), "record.answers.sections"},
		{len(record.Outcomes) > 0, "record.outcomes"}, {len(record.UnknownFields) > 0, "record.unknownFields"},
	}
	var diagnostics []api.Diagnostic
	for _, check := range checks {
		if check.bad {
			diagnostics = append(diagnostics, own("inspect.seed", "seed", check.path)...)
		}
	}
	return diagnostics
}

func hasSectionAnswers(answers model.Answers) bool {
	for _, values := range answers.Sections {
		if len(values) > 0 {
			return true
		}
	}
	return false
}

func (receiver reviewer) Review(incoming api.Incoming, local api.Local, choice api.PlanChoice) (api.ReviewResult, []api.Diagnostic) {
	result, _, diagnostics := receiver.review(incoming, local, choice)
	return result, diagnostics
}

func (receiver reviewer) review(incoming api.Incoming, local api.Local, choice api.PlanChoice) (api.ReviewResult, model.Record, []api.Diagnostic) {
	state, diagnostics := receiver.inspect(incoming.Archive)
	if len(diagnostics) > 0 {
		return api.ReviewResult{}, model.Record{}, diagnostics
	}
	if !equalIncoming(incoming, state.incoming) {
		return api.ReviewResult{}, model.Record{}, own("review.incoming", "incoming", "incoming")
	}
	if diagnostics = localProblems(local); len(diagnostics) > 0 {
		return api.ReviewResult{}, model.Record{}, diagnostics
	}
	if choice != api.PlanChoiceUndecided && choice != api.PlanChoiceUseIncoming {
		return api.ReviewResult{}, model.Record{}, own("review.choice", "plan-choice", "choice")
	}
	if local.KnownRecordMissing {
		return api.ReviewResult{}, model.Record{}, own("review.known-record", "known-record-missing", "local.record")
	}
	if local.Plan.Present && !bytes.Equal(local.Plan.Bytes, incoming.Files.Plan) && choice == api.PlanChoiceUndecided {
		return api.ReviewResult{PlanChoiceNeeded: true}, model.Record{}, nil
	}
	previous, comparison := receiver.previousPlan(local.Plan, state.plan.ID)
	preparation := state.preparation
	var preparationProblems []api.Diagnostic
	if local.Preparation.Present {
		var problems []model.Problem
		preparation, problems = receiver.preparationDecoder.Decode(local.Preparation.Bytes)
		if fatal(problems) {
			return api.ReviewResult{}, model.Record{}, wrap("review.preparation-decode", problems)
		}
		if !admit(state.plan, model.Record{}, &preparation, 0, false) {
			return api.ReviewResult{}, model.Record{}, resource("review.preparation-admission", "local.preparation")
		}
		problems = receiver.preparationValidator.Validate(preparationContext(state.plan, incoming.Files.Account), preparation)
		if len(problems) > 0 {
			for _, problem := range problems {
				if !staleReference(preparation, problem) {
					return api.ReviewResult{}, model.Record{}, wrap("review.preparation-validation", problems)
				}
			}
			preparationProblems = wrap("review.preparation-validation", problems)
		}
	}
	original := state.record
	if local.Record.Present {
		var problems []model.Problem
		original, problems = receiver.recordDecoder.Decode(local.Record.Bytes)
		if fatal(problems) {
			return api.ReviewResult{}, model.Record{}, wrap("review.record-decode", problems)
		}
		if !admit(state.plan, original, &preparation, 0, false) {
			return api.ReviewResult{}, model.Record{}, resource("review.record-admission", "local.record")
		}
		if original.Account != incoming.Files.Account {
			diagnostics = append(diagnostics, own("review.record-identity", "record-account", "record.account")...)
		}
		if original.Plan.ID != "" && original.Plan.ID != state.plan.ID {
			diagnostics = append(diagnostics, own("review.record-identity", "record-plan", "record.plan.id")...)
		}
		if len(diagnostics) > 0 {
			return api.ReviewResult{}, model.Record{}, diagnostics
		}
		if diagnostics = ownershipProblems(state.plan, original); len(diagnostics) > 0 {
			return api.ReviewResult{}, model.Record{}, diagnostics
		}
		if diagnostics = historyProblems(original); len(diagnostics) > 0 {
			return api.ReviewResult{}, model.Record{}, diagnostics
		}
	}
	proposed, reconciliation, problems := receiver.reconciler.Reconcile(original, previous, state.plan, incoming.Reference)
	if fatal(problems) {
		return api.ReviewResult{}, model.Record{}, wrap("review.reconciliation", problems)
	}
	effective := receiver.joiner.Join(state.plan, proposed)
	sourceBytes := len(incoming.Archive)
	for _, body := range []api.Body{local.Plan, local.Preparation, local.Record} {
		if body.Present {
			sourceBytes += len(body.Bytes)
		}
	}
	if !admit(effective, proposed, &preparation, sourceBytes, true) {
		return api.ReviewResult{}, model.Record{}, resource("review.effective-admission", "effectivePlan")
	}
	problems = receiver.planValidator.Validate(effective)
	if fatal(problems) {
		return api.ReviewResult{}, model.Record{}, wrap("review.effective-validation", problems)
	}
	effectiveWarnings := wrap("review.effective-validation", problems)
	view, problems := receiver.walker.Walk(effective, proposed)
	if fatal(problems) {
		return api.ReviewResult{}, model.Record{}, wrap("review.walker", problems)
	}
	session := api.Session{Questions: view.Questions, Blockers: view.Blockers}
	for _, section := range view.Sections {
		session.Activities = append(session.Activities, api.Activity{Section: section.Section, CurrentHash: section.CurrentHash, WaitingOn: section.WaitingOn, Escalation: section.Escalation})
	}
	for _, outcome := range view.Outcomes {
		session.Outcomes = append(session.Outcomes, api.Outcome{Outcome: outcome.Outcome, CurrentHash: outcome.CurrentHash, Comparison: string(outcome.Comparison)})
	}
	artifacts := []api.Artifact{{Role: api.RolePlan, Action: api.ActionInstall, Bytes: incoming.Files.Plan}, {Role: api.RolePreparation, Action: api.ActionInstall, Bytes: incoming.Files.Preparation}, {Role: api.RoleRecord, Action: api.ActionCreate}}
	if local.Plan.Present {
		artifacts[0].Action = api.ActionReplace
		if bytes.Equal(local.Plan.Bytes, incoming.Files.Plan) {
			artifacts[0].Action = api.ActionRetain
			artifacts[0].Bytes = local.Plan.Bytes
		}
	}
	if local.Preparation.Present {
		artifacts[1].Action = api.ActionRetain
		artifacts[1].Bytes = local.Preparation.Bytes
	}
	if local.Record.Present {
		if reflect.DeepEqual(original, proposed) {
			artifacts[2].Action = api.ActionRetain
			artifacts[2].Bytes = local.Record.Bytes
		} else {
			encoded, errors := receiver.encode(proposed, "review")
			if len(errors) > 0 {
				return api.ReviewResult{}, model.Record{}, errors
			}
			artifacts[2].Bytes = encoded
			artifacts[2].Action = api.ActionReconcile
			if original.Plan.ID == "" {
				artifacts[2].Action = api.ActionBind
			}
		}
	}
	support := api.Support{}
	if len(preparationProblems) == 0 {
		support = api.Support{Available: true, Recipient: preparation.Recipient, Arrangements: preparation.Arrangements}
	}
	reviewed := api.Review{Account: incoming.Files.Account, PlanID: state.plan.ID, Paths: incoming.Paths, Artifacts: artifacts, Comparison: comparison, Reconciliation: reconciliation, NeedsFirstOpenedAt: !local.Record.Present, BindsUnboundRecord: local.Record.Present && original.Plan.ID == "", PreparationProblems: preparationProblems, IncomingWarnings: state.incoming.Warnings, EffectivePlanWarnings: effectiveWarnings, SessionProblems: wrap("review.walker", problems), Session: session, Support: support, InvitationFirstSteps: incoming.Files.FirstSteps}
	proposal := owned(api.Proposal{Evidence: api.Evidence{Incoming: state.incoming, Local: local, Choice: choice}, Review: reviewed})
	return api.ReviewResult{Proposal: &proposal}, proposed, nil
}

func (receiver reviewer) previousPlan(body api.Body, id string) (model.Plan, api.Comparison) {
	comparison := api.Comparison{Reason: "absent"}
	if !body.Present {
		return model.Plan{}, comparison
	}
	plan, problems := receiver.planDecoder.Decode(body.Bytes)
	comparison.Problems = wrap("previous.plan-decode", problems)
	if fatal(problems) {
		comparison.Reason = "decode"
		return model.Plan{}, comparison
	}
	if !admit(plan, model.Record{}, nil, len(body.Bytes), true) {
		comparison.Reason = "resource"
		comparison.Problems = append(comparison.Problems, own("previous.plan-admission", "previous-resource", "previousPlan")...)
		return model.Plan{}, comparison
	}
	for _, phase := range []struct {
		stage, reason string
		validate      func(model.Plan) []model.Problem
	}{{"previous.plan-authoring", "authoring", receiver.authoringValidator.Validate}, {"previous.plan-validation", "validation", receiver.planValidator.Validate}} {
		problems = phase.validate(plan)
		comparison.Problems = append(comparison.Problems, wrap(phase.stage, problems)...)
		if fatal(problems) {
			comparison.Reason = phase.reason
			return model.Plan{}, comparison
		}
	}
	switch {
	case plan.ID == "":
		comparison.Reason = "empty-id"
	case plan.ID != id:
		comparison.Reason = "different-id"
	case len(plan.Variables) == 0 && len(plan.Sections) == 0:
		comparison.Reason = "empty-context"
	default:
		comparison.Reason = "available"
		comparison.Available = true
		return plan, comparison
	}
	return model.Plan{}, comparison
}

func (receiver reviewer) Begin(proposal api.Proposal, event model.Timestamp) (api.Publication, []api.Diagnostic) {
	result, record, diagnostics := receiver.review(proposal.Evidence.Incoming, proposal.Evidence.Local, proposal.Evidence.Choice)
	if len(diagnostics) > 0 {
		return api.Publication{}, diagnostics
	}
	if result.Proposal == nil || !equalProposal(*result.Proposal, proposal) {
		return api.Publication{}, own("begin.proposal", "proposal", "proposal")
	}
	newRecord := result.Proposal.Review.NeedsFirstOpenedAt
	if (newRecord && !validTimestamp(string(event), true)) || (!newRecord && event != "") {
		return api.Publication{}, own("begin.event", "event", "event")
	}
	publication := api.Publication{OriginalReview: owned(result.Proposal.Review), Artifacts: owned(result.Proposal.Review.Artifacts), FirstOpenedAt: record.FirstOpenedAt}
	if newRecord {
		record.FirstOpenedAt = event
		encoded, problems := receiver.encode(record, "begin")
		if len(problems) > 0 {
			return api.Publication{}, problems
		}
		publication.Artifacts[2].Bytes = owned(encoded)
		publication.FirstOpenedAt = event
	}
	return publication, nil
}

func (receiver reviewer) encode(record model.Record, phase string) ([]byte, []api.Diagnostic) {
	encoded, problems := receiver.encoder.Encode(record)
	if len(problems) > 0 {
		return nil, wrap(phase+".record-encode", problems)
	}
	if len(encoded) > 1048576 {
		return nil, own(phase+".record-size", "output-size", "record")
	}
	return encoded, nil
}
