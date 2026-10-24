package reviewer_test

import (
	"maps"
	"reflect"
	"strings"
	"testing"

	contract "onboard-kit/contracts/session/reviewer"
	"onboard-kit/internal/model"
)

func requireOwnError(t *testing.T, diagnostics []contract.Diagnostic, stage, code, message, path string) {
	t.Helper()
	errors := fatal(diagnostics)
	var detail map[string]string
	if path != "" {
		detail = map[string]string{"path": path}
	}
	if len(errors) != 1 || errors[0].Stage != stage || errors[0].Problem.Code != code || errors[0].Problem.Message != message ||
		errors[0].Problem.Question != "" || errors[0].Problem.Section != "" || !maps.Equal(errors[0].Problem.Detail, detail) {
		t.Fatalf("got diagnostics %#v", diagnostics)
	}
}

func TestExplicitContextAndCompleteIdentityBatch(t *testing.T) {
	for _, change := range []func(*contract.Source){
		func(input *contract.Source) { input.Account = "" },
		func(input *contract.Source) { input.Account = " guessed account " },
		func(input *contract.Source) { input.PlanID = "" },
	} {
		input := source(accessPlan)
		change(&input)
		input.Plan = []byte("not a plan")
		got, diagnostics := newReviewer().Inspect(input)
		if !reflect.DeepEqual(got, contract.Session{}) {
			t.Fatal("context refusal returned usable session")
		}
		requireOwnError(t, diagnostics, "context", "session.context", "Select an explicit valid account and plan identity.", "context")
	}
	input := source(accessPlan)
	input.Plan = []byte(strings.Replace(accessPlan, "id: plan-one", "id: other-plan", 1))
	input.Record = []byte(strings.Replace(strings.Replace(recordBefore, "account: newcomer", "account: Newcomer", 1), "id: plan-one", "id: other-record-plan", 1))
	got, diagnostics := newReviewer().Inspect(input)
	if !reflect.DeepEqual(got, contract.Session{}) {
		t.Fatal("identity refusal returned usable session")
	}
	want := []contract.Diagnostic{
		{Stage: "identity", Problem: model.Problem{Code: "session.plan", Severity: model.SeverityError, Message: "The shared plan does not match this session.", Detail: map[string]string{"path": "plan.id"}}},
		{Stage: "identity", Problem: model.Problem{Code: "session.record-account", Severity: model.SeverityError, Message: "The record account does not match this session.", Detail: map[string]string{"path": "record.account"}}},
		{Stage: "identity", Problem: model.Problem{Code: "session.record-plan", Severity: model.SeverityError, Message: "The record plan does not match this session.", Detail: map[string]string{"path": "record.plan.id"}}},
	}
	if !reflect.DeepEqual(fatal(diagnostics), want) {
		t.Fatalf("identity order/batch %#v", diagnostics)
	}
	input = source(accessPlan)
	input.Record = []byte(strings.Replace(recordBefore, "  id: plan-one\n", "", 1))
	_, diagnostics = newReviewer().Inspect(input)
	requireOwnError(t, diagnostics, "identity", "session.record-plan", "The record plan does not match this session.", "record.plan.id")
}

func TestActualCodecAndDeclarationRefusals(t *testing.T) {
	for _, role := range []string{"plan", "record", "preparation"} {
		input := source(accessPlan)
		switch role {
		case "plan":
			input.Plan = []byte(strings.Replace(accessPlan, "schemaVersion: 1", "schemaVersion: 2", 1))
		case "record":
			input.Record = []byte(strings.Replace(recordBefore, "schemaVersion: 1", "schemaVersion: 2", 1))
		case "preparation":
			input.Preparation = []byte(strings.Replace(preparation, "schemaVersion: 1", "schemaVersion: 2", 1))
		}
		got, diagnostics := newReviewer().Inspect(input)
		errors := fatal(diagnostics)
		if !reflect.DeepEqual(got, contract.Session{}) || len(errors) != 1 || errors[0].Stage != role+".decode" || errors[0].Problem.Code != "version.too-new" || errors[0].Problem.Detail["path"] != "schemaVersion" {
			t.Fatalf("%s codec refusal %#v %#v", role, got, diagnostics)
		}
	}
	input := source(strings.Replace(strings.Replace(answerPlan, "title: Personal choices", "title: \" \"", 1), "type: string", "type: text", 1))
	got, diagnostics := newReviewer().Inspect(input)
	errors := fatal(diagnostics)
	if !reflect.DeepEqual(got, contract.Session{}) || len(errors) != 2 {
		t.Fatalf("authoring batch %#v %#v", got, diagnostics)
	}
	for i, want := range []struct{ code, path, question string }{
		{"authoring.required-prose", "sections[0].title", ""},
		{"authoring.variable-type", "sections[0].variables[0].type", "personal.choice"},
	} {
		if errors[i].Stage != "plan.authoring" || errors[i].Problem.Code != want.code || errors[i].Problem.Question != want.question || errors[i].Problem.Section != "personal" ||
			!maps.Equal(errors[i].Problem.Detail, map[string]string{"path": want.path}) || errors[i].Problem.Message == "" {
			t.Fatalf("authoring[%d] %#v", i, errors[i])
		}
	}
}

func TestNoAmbiguousPlanOrOutcomeWinner(t *testing.T) {
	input := source(accessPlan + "  - id: access\n    title: Second declaration\n    summary: Do not choose a winner\n")
	got, diagnostics := newReviewer().Inspect(input)
	errors := fatal(diagnostics)
	if !reflect.DeepEqual(got, contract.Session{}) || len(errors) != 1 || errors[0].Stage != "plan.validate" || errors[0].Problem.Code != "plan.duplicate-section-id" || errors[0].Problem.Section != "access" {
		t.Fatalf("ambiguous plan %#v %#v", got, diagnostics)
	}
	input = source(accessPlan)
	input.Record = []byte(strings.Replace(recordBefore, "future:\n", "  - section: z-removed\n    outcome: failed\n    on: another-old-time\nfuture:\n", 1))
	got, diagnostics = newReviewer().Inspect(input)
	errors = fatal(diagnostics)
	if !reflect.DeepEqual(got, contract.Session{}) || len(errors) != 1 || errors[0].Stage != "walk" || errors[0].Problem.Code != "walk.duplicate-outcome" ||
		errors[0].Problem.Section != "z-removed" || errors[0].Problem.Question != "" || !maps.Equal(errors[0].Problem.Detail, map[string]string{"field": "outcomes"}) || errors[0].Problem.Message != "Keep one recorded outcome for this section." {
		t.Fatalf("ambiguous held history %#v %#v", got, diagnostics)
	}
}

func TestOwnershipAndUnknownKindGuardAllHeldEntries(t *testing.T) {
	input := source(strings.Replace(answerPlan, "answers:\n", "variables:\n  - name: organization\n    type: string\nanswers:\n", 1))
	input.Plan = []byte(strings.Replace(string(input.Plan), "audience: onboardee", "audience: onboarder", 1))
	input.Record = []byte(strings.Replace(recordBefore, "answers:\n", "answers:\n  plan:\n    aaaa: false\n    organization: deferred\n    zzzz: held\n", 1))
	input.Record = []byte(strings.Replace(string(input.Record), "      hidden: deferred\noutcomes:", "      hidden: deferred\n    personal:\n      aaaa: []\n      choice: deferred\n      zzzz: false\noutcomes:", 1))
	got, diagnostics := newReviewer().Inspect(input)
	errors := fatal(diagnostics)
	if !reflect.DeepEqual(got, contract.Session{}) || len(errors) != 2 {
		t.Fatalf("ownership batch %#v %#v", got, diagnostics)
	}
	for i, path := range []string{"answers.plan[1]", "answers.sections[1].values[1]"} {
		if errors[i].Stage != "record.ownership" || errors[i].Problem.Code != "session.record-owned-answer" || errors[i].Problem.Message != "Keep preparation answers out of the newcomer record." ||
			errors[i].Problem.Question != "" || errors[i].Problem.Section != "" || !maps.Equal(errors[i].Problem.Detail, map[string]string{"path": path}) {
			t.Fatalf("ownership[%d] %#v", i, errors[i])
		}
	}
	input = source(accessPlan)
	input.Record = []byte(strings.Replace(strings.Replace(recordBefore, "outcome: verified", "outcome: VERIFIED", 1), "outcome: deferred", "outcome: future-kind", 1))
	got, diagnostics = newReviewer().Inspect(input)
	errors = fatal(diagnostics)
	if !reflect.DeepEqual(got, contract.Session{}) || len(errors) != 2 {
		t.Fatalf("kind batch %#v %#v", got, diagnostics)
	}
	for i, path := range []string{"outcomes[0].kind", "outcomes[1].kind"} {
		if errors[i].Stage != "record.history" || errors[i].Problem.Code != "session.outcome-kind" || errors[i].Problem.Message != "Repair the unknown recorded outcome kind before editing this session." ||
			errors[i].Problem.Question != "" || errors[i].Problem.Section != "" || !maps.Equal(errors[i].Problem.Detail, map[string]string{"path": path}) {
			t.Fatalf("kind[%d] %#v", i, errors[i])
		}
	}
}

func TestIntentIsOneExplicitOperationAndWriterErrorsPassThrough(t *testing.T) {
	for _, intent := range []contract.Intent{
		{Operation: "ANSWER", Question: "personal.choice", Value: text("new")},
		{Operation: contract.OperationAnswer, Question: "personal.choice", Value: text("new"), Section: "personal"},
		{Operation: contract.OperationAnswer, Question: "personal.choice", Value: model.Value{Kind: 99}},
		{Operation: contract.OperationFollowPlan, Question: "personal.choice", Value: model.Value{List: []string{}}},
		{Operation: contract.OperationOutcome, Section: "personal", Kind: model.OutcomeFailed, Value: model.Value{List: []string{}}},
	} {
		got, diagnostics := newReviewer().Propose(source(answerPlan), intent)
		if !reflect.DeepEqual(got, contract.Proposal{}) {
			t.Fatal("irrelevant intent fields returned usable proposal")
		}
		requireOwnError(t, diagnostics, "intent", "session.intent", "Choose exactly one session change.", "")
	}
	input := source(strings.Replace(strings.Replace(answerPlan, "type: string", "type: contact\n        multiple: true", 1), "        default: default-example\n", "", 1))
	input.Plan = []byte(strings.Replace(string(input.Plan), "answers:\n  sections:\n    personal:\n      choice: org-a\n", "", 1))
	got, diagnostics := newReviewer().Propose(input, contract.Intent{Operation: contract.OperationAnswer, Question: "personal.choice", Value: model.Value{Kind: model.ValueKindList, List: []string{"missing-one", "", "missing-two"}}})
	errors := fatal(diagnostics)
	if !reflect.DeepEqual(got, contract.Proposal{}) || len(errors) != 2 {
		t.Fatalf("actual writer batch %#v %#v", got, diagnostics)
	}
	for i, index := range []string{"0", "2"} {
		if errors[i].Stage != "answer" || errors[i].Problem.Code != "directory.unknown-id" || errors[i].Problem.Question != "personal.choice" || errors[i].Problem.Section != "personal" ||
			!maps.Equal(errors[i].Problem.Detail, map[string]string{"kind": "contact", "item": index}) || errors[i].Problem.Message == "" {
			t.Fatalf("writer[%d] %#v", i, errors[i])
		}
	}
}

func TestStaleSupportIsUnavailableNotAnAnswerAndOtherErrorsRefuse(t *testing.T) {
	input := source(accessPlan)
	input.Preparation = []byte(preparation + "  buddy: missing-contact\n")
	session, diagnostics := newReviewer().Inspect(input)
	success(t, diagnostics)
	if session.Account != "newcomer" || session.Support.Available || session.Support.Recipient != (model.PreparationRecipient{}) || session.Support.Arrangements != nil || len(session.Support.Problems) != 1 {
		t.Fatalf("stale support %#v", session)
	}
	problem := session.Support.Problems[0]
	if problem.Stage != "preparation.validate" || problem.Problem.Code != "preparation.unknown-contact" || problem.Problem.Severity != model.SeverityError || !maps.Equal(problem.Problem.Detail, map[string]string{"path": "recipient.buddy"}) {
		t.Fatalf("stale problem relabeled/lost %#v", problem)
	}
	proposal, diagnostics := newReviewer().Propose(input, contract.Intent{Operation: contract.OperationOutcome, Section: "access", Kind: model.OutcomeVerified})
	success(t, diagnostics)
	if !reflect.DeepEqual(proposal.AfterRecord.Answers, heldRecord().Answers) {
		t.Fatal("support became accepted answer")
	}
	input.Preparation = []byte(preparation + "  buddy: missing-contact\n  plannedStartDate: \"2026-02-30\"\n")
	got, diagnostics := newReviewer().Inspect(input)
	errors := fatal(diagnostics)
	if !reflect.DeepEqual(got, contract.Session{}) || len(errors) != 2 || errors[0].Stage != "preparation.validate" || errors[0].Problem.Code != "preparation.date" ||
		errors[0].Problem.Detail["path"] != "recipient.plannedStartDate" || errors[1].Problem.Code != "preparation.unknown-contact" {
		t.Fatalf("mixed invalid preparation was wrongly allowed %#v %#v", got, diagnostics)
	}
}

func TestCandidateValidationRejectsNewlyActiveInvalidHeldAnswer(t *testing.T) {
	input := source(`schemaVersion: 1
id: plan-one
answers:
  sections:
    personal:
      gate: false
      detail: false
sections:
  - id: personal
    title: Personal choices
    summary: Gate $gate; detail $detail.
    variables:
      - name: gate
        type: boolean
        audience: onboardee
      - name: detail
        type: string
        audience: onboardee
        askIf:
          all:
            - question: personal.gate
              answer:
                is: true
`)
	original := cloneSource(input)
	session, diagnostics := newReviewer().Inspect(input)
	success(t, diagnostics)
	if session.Account != "newcomer" {
		t.Fatal("original inactive stored value did not reach valid inspection")
	}
	got, diagnostics := newReviewer().Propose(input, contract.Intent{Operation: contract.OperationAnswer, Question: "personal.gate", Value: model.Value{Kind: model.ValueKindFlag, Flag: true}})
	errors := fatal(diagnostics)
	if !reflect.DeepEqual(got, contract.Proposal{}) || len(errors) != 1 || errors[0].Stage != "candidate.validate" || errors[0].Problem.Code != "answer.wrong-kind" ||
		errors[0].Problem.Question != "personal.detail" || errors[0].Problem.Section != "personal" || !maps.Equal(errors[0].Problem.Detail, map[string]string{"expected": "text", "found": "flag"}) {
		t.Fatalf("candidate full validation did not refuse %#v %#v", got, diagnostics)
	}
	if !reflect.DeepEqual(input, original) {
		t.Fatal("rejected candidate changed original source")
	}
}
