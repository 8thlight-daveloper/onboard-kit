package reviewer_test

import (
	"bytes"
	"maps"
	"reflect"
	"strings"
	"testing"

	contract "onboard-kit/contracts/session/reviewer"
	"onboard-kit/internal/model"
)

func TestOutcomePreviewIsUncommittedAndFinalizeUsesActualEvent(t *testing.T) {
	input := source(accessPlan)
	input.Plan = []byte(accessPlan + "    source:\n      template: base\n      hash: old-source\n")
	input.Record = []byte(strings.Replace(recordBefore, "  - section: a-removed", "  - section: access\n    outcome: failed\n    on: old\n    sectionHash: old\n    note: Clear this\n    differently: Clear this too\n  - section: a-removed", 1))
	reviewer := newReviewer()
	proposal, diagnostics := reviewer.Propose(input, contract.Intent{Operation: contract.OperationOutcome, Section: "access", Kind: model.OutcomeVerified})
	success(t, diagnostics)
	if !proposal.NeedsEvent || len(proposal.AfterRecord.Outcomes) != 3 {
		t.Fatalf("outcome proposal %#v", proposal)
	}
	preview := outcome(t, proposal.After, "access")
	if preview.Outcome.Kind != model.OutcomeVerified || preview.Outcome.On != "" || preview.Outcome.SectionHash != accessHash || preview.CurrentHash != accessHash || preview.Comparison != "same" {
		t.Fatalf("preview must not claim an event: %#v", preview)
	}
	if proposal.AfterRecord.Outcomes[1].On != "" || proposal.AfterRecord.Outcomes[1].Note != "" || proposal.AfterRecord.Outcomes[1].Differently != "" ||
		proposal.AfterRecord.Outcomes[0].On != "older" || proposal.AfterRecord.Outcomes[2].On != "legacy-on" {
		t.Fatalf("target replacement/history %#v", proposal.AfterRecord.Outcomes)
	}
	publication, diagnostics := reviewer.Finalize(proposal, event)
	success(t, diagnostics)
	if !publication.Changed || publication.Event != event || !bytes.Equal(publication.Record, []byte(recordVerified)) {
		t.Fatalf("real event/literal outcome publication %#v\n%s", publication, publication.Record)
	}
	if proposal.AfterRecord.FirstOpenedAt != "held-uninterpreted-event" || proposal.AfterRecord.StartDate != "legacy-date" || proposal.AfterRecord.Plan != heldRecord().Plan {
		t.Fatal("outcome preview altered unrelated legacy metadata")
	}
	got, diagnostics := reviewer.Finalize(proposal, "not-a-time")
	if !reflect.DeepEqual(got, contract.Publication{}) {
		t.Fatalf("invalid event produced publication %#v", got)
	}
	errors := fatal(diagnostics)
	if len(errors) != 1 || errors[0].Stage != "outcome" || errors[0].Problem.Code != "outcome.timestamp-invalid" ||
		errors[0].Problem.Severity != model.SeverityError || errors[0].Problem.Question != "" || errors[0].Problem.Section != "" ||
		!maps.Equal(errors[0].Problem.Detail, map[string]string{"field": "on"}) || errors[0].Problem.Message != "Supply a valid UTC outcome timestamp." {
		t.Fatalf("actual timestamp diagnostics %#v", diagnostics)
	}
}

func TestOnlyTargetPreviewEventIsBlanked(t *testing.T) {
	input := source(accessPlan)
	input.Record = []byte(strings.Replace(recordBefore, "on: legacy-on", "on: \"2000-01-01T00:00:00Z\"", 1))
	proposal, diagnostics := newReviewer().Propose(input, contract.Intent{Operation: contract.OperationOutcome, Section: "access", Kind: model.OutcomeFailed})
	success(t, diagnostics)
	if outcome(t, proposal.After, "z-removed").Outcome.On != "2000-01-01T00:00:00Z" || outcome(t, proposal.After, "access").Outcome.On != "" {
		t.Fatal("preview blanked unrelated held event by matching sentinel text")
	}
}

const prerequisitePlan = `schemaVersion: 1
id: plan-one
sections:
  - id: access
    title: Access
    summary: Request access
    requires:
      - a
  - id: a
    title: Prerequisite
    summary: Complete prerequisite
`

func TestCurrentDirectPrerequisiteGate(t *testing.T) {
	for _, held := range []model.OutcomeKind{"", model.OutcomeFailed, model.OutcomeDeferred, model.OutcomeVerified} {
		t.Run(string(held), func(t *testing.T) {
			input := source(prerequisitePlan)
			if held != "" {
				input.Record = []byte(strings.Replace(recordBefore, "  - section: a-removed", "  - section: a\n    outcome: "+string(held)+"\n    on: old-uninterpreted\n    sectionHash: mismatch\n  - section: a-removed", 1))
			}
			for _, proposed := range []model.OutcomeKind{model.OutcomeVerified, model.OutcomeFailed, model.OutcomeDeferred} {
				proposal, diagnostics := newReviewer().Propose(input, contract.Intent{Operation: contract.OperationOutcome, Section: "access", Kind: proposed})
				if proposed == model.OutcomeVerified && held != model.OutcomeVerified {
					if !reflect.DeepEqual(proposal, contract.Proposal{}) {
						t.Fatal("waiting verified proposal was usable")
					}
					errors := fatal(diagnostics)
					if len(errors) != 1 || errors[0].Stage != "outcome" || errors[0].Problem.Code != "outcome.prerequisite-unverified" ||
						errors[0].Problem.Section != "access" || errors[0].Problem.Question != "" ||
						!maps.Equal(errors[0].Problem.Detail, map[string]string{"field": "requires", "required": "a"}) {
						t.Fatalf("actual prerequisite diagnostics %#v", diagnostics)
					}
					continue
				}
				success(t, diagnostics)
				preview := outcome(t, proposal.After, "access")
				if preview.CurrentHash != accessRequiresHash || preview.Outcome.SectionHash != accessRequiresHash || preview.Outcome.Kind != proposed || preview.Outcome.On != "" || !proposal.NeedsEvent {
					t.Fatalf("truthful attempt/current definition %#v", preview)
				}
				publication, diagnostics := newReviewer().Finalize(proposal, event)
				success(t, diagnostics)
				if !publication.Changed || publication.Event != event || !bytes.Contains(publication.Record, []byte(accessRequiresHash)) {
					t.Fatalf("valid current attempt publication %#v", publication)
				}
			}
		})
	}
}

func TestMissingQuestionsAreNotAnExtraOutcomeGate(t *testing.T) {
	input := source(answerPlan)
	input.Plan = []byte(strings.Replace(answerPlan, "answers:\n  sections:\n    personal:\n      choice: org-a\n", "", 1))
	input.Plan = []byte(strings.Replace(string(input.Plan), "        description: Your choice", "        description: Your choice\n        required: true", 1))
	proposal, diagnostics := newReviewer().Propose(input, contract.Intent{Operation: contract.OperationOutcome, Section: "personal", Kind: model.OutcomeVerified})
	success(t, diagnostics)
	if len(proposal.Before.Blockers) != 1 || proposal.Before.Blockers[0].Question != "personal.choice" || !proposal.NeedsEvent {
		t.Fatalf("missing question was hidden or gate changed %#v", proposal)
	}
	publication, diagnostics := newReviewer().Finalize(proposal, event)
	success(t, diagnostics)
	if !publication.Changed || publication.Event != event || bytes.Contains(publication.Record, []byte("choice: default-example")) {
		t.Fatal("missing answer/default became verification prerequisite or accepted answer")
	}
}

func fatal(diagnostics []contract.Diagnostic) []contract.Diagnostic {
	var errors []contract.Diagnostic
	for _, diagnostic := range diagnostics {
		if diagnostic.Problem.Severity == model.SeverityError {
			errors = append(errors, diagnostic)
		}
	}
	return errors
}
