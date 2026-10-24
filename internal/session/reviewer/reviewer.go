package reviewer

import (
	"crypto/sha256"
	"encoding/hex"
	"reflect"

	outcomecontract "onboard-kit/contracts/outcome/recorder"
	preparationcontract "onboard-kit/contracts/preparation/validator"
	writercontract "onboard-kit/contracts/record/writer"
	contract "onboard-kit/contracts/session/reviewer"
	"onboard-kit/internal/model"
	"onboard-kit/internal/outcome/recorder"
	"onboard-kit/internal/plan/authoringvalidator"
	plandecoder "onboard-kit/internal/plan/decoder"
	planvalidator "onboard-kit/internal/plan/validator"
	preparationdecoder "onboard-kit/internal/preparation/decoder"
	preparationvalidator "onboard-kit/internal/preparation/validator"
	recorddecoder "onboard-kit/internal/record/decoder"
	recordencoder "onboard-kit/internal/record/encoder"
	"onboard-kit/internal/record/joiner"
	"onboard-kit/internal/record/writer"
	"onboard-kit/internal/section/walker"
)

type Reviewer struct{}

func New() Reviewer { return Reviewer{} }

type inspected struct {
	plan        model.Plan
	record      model.Record
	preparation model.Preparation
	effective   model.Plan
	session     contract.Session
}

func (Reviewer) Inspect(source contract.Source) (contract.Session, []contract.Diagnostic) {
	value, diagnostics := inspect(source)
	if fatal(diagnostics) {
		return contract.Session{}, diagnostics
	}
	return clone(value.session), diagnostics
}

func inspect(source contract.Source) (inspected, []contract.Diagnostic) {
	var state inspected
	diagnostics := sourceAdmission(source)
	if fatal(diagnostics) {
		return inspected{}, diagnostics
	}
	if !accountValid(source.Account) || source.PlanID == "" {
		return inspected{}, own("context", "context", "Select an explicit valid account and plan identity.", "context")
	}
	var problems []model.Problem
	state.plan, problems = plandecoder.New().Decode(source.Plan)
	diagnostics = appendProblems(diagnostics, "plan.decode", problems)
	if fatal(diagnostics) {
		return inspected{}, diagnostics
	}
	if !admitted(state.plan, model.Record{}, nil, sourceSize(source), true) {
		return inspected{}, appendResource(diagnostics, "plan.admission", "plan")
	}
	diagnostics = appendProblems(diagnostics, "plan.authoring", authoringvalidator.New().Validate(state.plan))
	if fatal(diagnostics) {
		return inspected{}, diagnostics
	}
	diagnostics = appendProblems(diagnostics, "plan.validate", planvalidator.New().Validate(state.plan))
	if fatal(diagnostics) {
		return inspected{}, diagnostics
	}
	state.record, problems = recorddecoder.New().Decode(source.Record)
	diagnostics = appendProblems(diagnostics, "record.decode", problems)
	if fatal(diagnostics) {
		return inspected{}, diagnostics
	}
	if !admitted(state.plan, state.record, nil, 0, false) {
		return inspected{}, appendResource(diagnostics, "record.admission", "record")
	}
	state.preparation, problems = preparationdecoder.New().Decode(source.Preparation)
	diagnostics = appendProblems(diagnostics, "preparation.decode", problems)
	if fatal(diagnostics) {
		return inspected{}, diagnostics
	}
	if !admitted(state.plan, state.record, &state.preparation, 0, false) {
		return inspected{}, appendResource(diagnostics, "preparation.admission", "preparation")
	}
	if state.plan.ID != source.PlanID {
		diagnostics = append(diagnostics, own("identity", "plan", "The shared plan does not match this session.", "plan.id")...)
	}
	if state.record.Account != source.Account {
		diagnostics = append(diagnostics, own("identity", "record-account", "The record account does not match this session.", "record.account")...)
	}
	if state.record.Plan.ID != source.PlanID {
		diagnostics = append(diagnostics, own("identity", "record-plan", "The record plan does not match this session.", "record.plan.id")...)
	}
	if fatal(diagnostics) {
		return inspected{}, diagnostics
	}
	diagnostics = append(diagnostics, ownership(state.plan, state.record)...)
	if fatal(diagnostics) {
		return inspected{}, diagnostics
	}
	for index, outcome := range state.record.Outcomes {
		if outcome.Kind != model.OutcomeVerified && outcome.Kind != model.OutcomeFailed && outcome.Kind != model.OutcomeDeferred {
			diagnostics = append(diagnostics, own("record.history", "outcome-kind", "Repair the unknown recorded outcome kind before editing this session.", indexed("outcomes", index)+".kind")...)
		}
	}
	if fatal(diagnostics) {
		return inspected{}, diagnostics
	}
	if !admitted(state.plan, state.record, &state.preparation, sourceSize(source), true) {
		return inspected{}, appendResource(diagnostics, "payload.admission", "session")
	}
	context := preparationcontract.Context{PlanID: source.PlanID, Account: source.Account, Record: &preparationcontract.RecordIdentity{Account: state.record.Account, PlanID: state.record.Plan.ID}}
	for _, value := range state.plan.Directory.Contacts {
		context.Contacts = append(context.Contacts, value.ID)
	}
	for _, value := range state.plan.Directory.Teams {
		context.Teams = append(context.Teams, value.ID)
	}
	for _, value := range state.plan.Sections {
		context.Sections = append(context.Sections, value.ID)
	}
	problems = preparationvalidator.New().Validate(context, state.preparation)
	support := contract.Support{Available: true, Recipient: state.preparation.Recipient, Arrangements: clone(state.preparation.Arrangements)}
	if len(problems) > 0 && staleSupport(state.preparation, problems) {
		support = contract.Support{Problems: appendProblems(nil, "preparation.validate", problems)}
	} else {
		diagnostics = appendProblems(diagnostics, "preparation.validate", problems)
		if fatal(diagnostics) {
			return inspected{}, diagnostics
		}
	}
	state.effective = joiner.New().Join(state.plan, state.record)
	if !admitted(state.effective, state.record, &state.preparation, sourceSize(source), true) {
		return inspected{}, appendResource(diagnostics, "effective.admission", "effectivePlan")
	}
	diagnostics = appendProblems(diagnostics, "effective.validate", planvalidator.New().Validate(state.effective))
	if fatal(diagnostics) {
		return inspected{}, diagnostics
	}
	view, problems := walker.New().Walk(state.effective, state.record)
	diagnostics = appendProblems(diagnostics, "walk", problems)
	if fatal(diagnostics) {
		return inspected{}, diagnostics
	}
	state.session = session(source, state.record, support, view)
	return state, diagnostics
}

func session(source contract.Source, record model.Record, support contract.Support, view walker.View) contract.Session {
	digest := sha256.Sum256(source.Plan)
	result := contract.Session{Account: source.Account, PlanID: source.PlanID, Reference: record.Plan, PlanReferenceMatches: record.Plan.Hash == hex.EncodeToString(digest[:]), Support: support, Questions: view.Questions, Blockers: view.Blockers}
	for _, value := range view.Sections {
		result.Activities = append(result.Activities, contract.Activity{Section: value.Section, CurrentHash: value.CurrentHash, WaitingOn: value.WaitingOn, Escalation: value.Escalation})
	}
	for _, value := range view.Outcomes {
		result.Outcomes = append(result.Outcomes, contract.HeldOutcome{Outcome: value.Outcome, CurrentHash: value.CurrentHash, Comparison: string(value.Comparison)})
	}
	return clone(result)
}

func (Reviewer) Propose(source contract.Source, intent contract.Intent) (contract.Proposal, []contract.Diagnostic) {
	diagnostics := sourceAdmission(source)
	if fatal(diagnostics) {
		return contract.Proposal{}, diagnostics
	}
	diagnostics = intentAdmission(intent)
	if fatal(diagnostics) {
		return contract.Proposal{}, diagnostics
	}
	state, diagnostics := inspect(source)
	if fatal(diagnostics) {
		return contract.Proposal{}, diagnostics
	}
	if !validIntent(intent) {
		return contract.Proposal{}, append(diagnostics, own("intent", "intent", "Choose exactly one session change.", "")...)
	}
	candidate, problems := change(state, intent, "2000-01-01T00:00:00Z")
	stage := "answer"
	needsEvent := intent.Operation == contract.OperationOutcome
	if needsEvent {
		stage = "outcome"
	}
	diagnostics = appendProblems(diagnostics, stage, problems)
	if fatal(diagnostics) {
		return contract.Proposal{}, diagnostics
	}
	after, more := candidateSession(source, state, candidate)
	diagnostics = append(diagnostics, more...)
	if fatal(diagnostics) {
		return contract.Proposal{}, diagnostics
	}
	if needsEvent {
		for index := range candidate.Outcomes {
			if candidate.Outcomes[index].Section == intent.Section {
				candidate.Outcomes[index].On = ""
			}
		}
		for index := range after.Outcomes {
			if after.Outcomes[index].Outcome.Section == intent.Section {
				after.Outcomes[index].Outcome.On = ""
			}
		}
	}
	return clone(contract.Proposal{Source: source, Intent: intent, Before: state.session, After: after, BeforeRecord: state.record, AfterRecord: candidate, NeedsEvent: needsEvent}), diagnostics
}

func change(state inspected, intent contract.Intent, event model.Timestamp) (model.Record, []model.Problem) {
	if intent.Operation == contract.OperationOutcome {
		return recorder.New().Record(state.effective, state.record, outcomecontract.Proposal{Section: intent.Section, Kind: intent.Kind, Note: intent.Note, Differently: intent.Differently}, event)
	}
	operation := writercontract.OperationSubmit
	if intent.Operation == contract.OperationFollowPlan {
		operation = writercontract.OperationFollowPlan
	}
	return writer.New().Write(state.plan, state.record, writercontract.Intent{Operation: operation, Question: intent.Question, Value: intent.Value})
}

func candidateSession(source contract.Source, state inspected, candidate model.Record) (contract.Session, []contract.Diagnostic) {
	if !admitted(state.plan, candidate, &state.preparation, 0, false) {
		return contract.Session{}, appendResource(nil, "candidate.admission", "candidate")
	}
	effective := joiner.New().Join(state.plan, candidate)
	if !admitted(effective, candidate, &state.preparation, sourceSize(source), true) {
		return contract.Session{}, appendResource(nil, "candidate.admission", "candidate")
	}
	diagnostics := appendProblems(nil, "candidate.validate", planvalidator.New().Validate(effective))
	if fatal(diagnostics) {
		return contract.Session{}, diagnostics
	}
	view, problems := walker.New().Walk(effective, candidate)
	diagnostics = appendProblems(diagnostics, "candidate.walk", problems)
	if fatal(diagnostics) {
		return contract.Session{}, diagnostics
	}
	return session(source, candidate, state.session.Support, view), diagnostics
}

func (reviewer Reviewer) Finalize(proposal contract.Proposal, event model.Timestamp) (contract.Publication, []contract.Diagnostic) {
	diagnostics := sourceAdmission(proposal.Source)
	if fatal(diagnostics) {
		return contract.Publication{}, diagnostics
	}
	diagnostics = intentAdmission(proposal.Intent)
	if fatal(diagnostics) {
		return contract.Publication{}, diagnostics
	}
	if !boundedText(string(event), 64) {
		return contract.Publication{}, appendResource(nil, "event.admission", "event")
	}
	expected, diagnostics := reviewer.Propose(proposal.Source, proposal.Intent)
	if fatal(diagnostics) {
		return contract.Publication{}, diagnostics
	}
	if !proposalEqual(expected, proposal) {
		return contract.Publication{}, append(diagnostics, own("proposal", "proposal", "Review this change again.", "")...)
	}
	if !expected.NeedsEvent && event != "" {
		return contract.Publication{}, append(diagnostics, own("event", "event", "An answer change does not use an event timestamp.", "")...)
	}
	state, inspection := inspect(expected.Source)
	if fatal(inspection) {
		return contract.Publication{}, append(diagnostics, inspection...)
	}
	candidate := expected.AfterRecord
	if expected.NeedsEvent {
		var problems []model.Problem
		candidate, problems = change(state, expected.Intent, event)
		diagnostics = appendProblems(diagnostics, "outcome", problems)
		if fatal(diagnostics) {
			return contract.Publication{}, diagnostics
		}
	}
	_, more := candidateSession(expected.Source, state, candidate)
	diagnostics = append(diagnostics, more...)
	if fatal(diagnostics) {
		return contract.Publication{}, diagnostics
	}
	if reflect.DeepEqual(candidate, state.record) {
		result := contract.Publication{Record: clone(expected.Source.Record)}
		if expected.NeedsEvent {
			result.Event = event
		}
		return result, diagnostics
	}
	encoded, problems := recordencoder.New().Encode(candidate)
	diagnostics = appendProblems(diagnostics, "record.encode", problems)
	if fatal(diagnostics) {
		return contract.Publication{}, diagnostics
	}
	if len(encoded) > 1048576 {
		return contract.Publication{}, appendResource(diagnostics, "record.encode", "record")
	}
	decoded, problems := recorddecoder.New().Decode(encoded)
	diagnostics = appendProblems(diagnostics, "record.roundtrip", problems)
	if fatal(diagnostics) {
		return contract.Publication{}, diagnostics
	}
	if !reflect.DeepEqual(recordNormal(candidate), recordNormal(decoded)) {
		return contract.Publication{}, append(diagnostics, own("record.roundtrip", "representation", "This record needs a lossless save path.", "")...)
	}
	result := contract.Publication{Record: clone(encoded), Changed: true}
	if expected.NeedsEvent {
		result.Event = event
	}
	return result, diagnostics
}

func recordNormal(record model.Record) model.Record {
	record = clone(record)
	if len(record.Answers.Plan) == 0 {
		record.Answers.Plan = nil
	}
	for scope, values := range record.Answers.Sections {
		if len(values) == 0 {
			delete(record.Answers.Sections, scope)
		}
	}
	if len(record.Answers.Sections) == 0 {
		record.Answers.Sections = nil
	}
	if len(record.Outcomes) == 0 {
		record.Outcomes = nil
	}
	return record
}
