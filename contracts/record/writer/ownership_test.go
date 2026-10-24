package writer_test

import (
	"reflect"
	"testing"

	contract "onboard-kit/contracts/record/writer"
	"onboard-kit/internal/model"
)

func commandPlan() model.Plan {
	plan := sharedPlan()
	plan.Sections[0].Variables[0].Type = model.VariableTypeCommand
	plan.Sections[0].Variables[0].Default = model.Value{}
	delete(plan.Answers.Sections, "personal")
	return plan
}

func commandRecord() model.Record {
	record := heldRecord()
	record.Answers.Sections["personal"] = map[string]model.Value{
		"choice": {Kind: model.ValueKindList, List: []string{"go", "", "test"}},
	}
	return record
}

func TestPreservationAndOwnership(t *testing.T) {
	plan, before := commandPlan(), heldRecord()
	intent := contract.Intent{Operation: contract.OperationSubmit, Question: "personal.choice", Value: model.Value{Kind: model.ValueKindList, List: []string{"go", "", "test"}}}
	writer := newWriter()
	got, problems := writer.Write(plan, before, intent)
	requireNoProblems(t, problems)
	requireRecord(t, got, commandRecord())
	second, problems := writer.Write(plan, before, intent)
	requireNoProblems(t, problems)
	got.Answers.Plan["unrelated"] = text("changed")
	got.Answers.Sections["gone"]["hidden"].List[0] = "changed"
	got.Answers.Sections["personal"]["choice"].List[0] = "changed"
	got.Outcomes[0].Note = "changed"
	got.UnknownFields["future"].(map[string]any)["items"].([]any)[1].(map[string]any)["nested"] = false
	got.UnknownFields["future"].(map[string]any)["emptyMap"].(map[string]any)["added"] = true
	requireRecord(t, before, heldRecord())
	requireRecord(t, second, commandRecord())
	if !reflect.DeepEqual(plan, commandPlan()) || !reflect.DeepEqual(intent.Value.List, []string{"go", "", "test"}) {
		t.Fatal("result aliases plan or intent")
	}
	before.Answers.Sections["gone"]["hidden"].List[0] = "input changed"
	before.Outcomes[1].Note = "input changed"
	before.UnknownFields["future"].(map[string]any)["emptyMap"].(map[string]any)["input"] = true
	intent.Value.List[0] = "intent changed"
	plan.Sections[0].Variables[0].Example = "plan changed"
	requireRecord(t, second, commandRecord())
	third, problems := writer.Write(commandPlan(), heldRecord(), contract.Intent{Operation: contract.OperationSubmit, Question: "personal.choice", Value: model.Value{Kind: model.ValueKindList, List: []string{"go", "", "test"}}})
	requireNoProblems(t, problems)
	requireRecord(t, third, commandRecord())
}

func TestNoOpAndUninterpretedHistoryRemainOwned(t *testing.T) {
	makeRecord := func() model.Record {
		record := heldRecord()
		record.Outcomes[1].Kind = "future-kind"
		record.Outcomes = append(record.Outcomes, model.Outcome{Section: "z-removed", Kind: model.OutcomeFailed, On: "duplicate-old"})
		record.Answers.Plan["organization"] = text("held-onboarder-key")
		return record
	}
	before := makeRecord()
	got, problems := newWriter().Write(sharedPlan(), before, contract.Intent{Operation: contract.OperationFollowPlan, Question: "personal.choice"})
	requireNoProblems(t, problems)
	requireRecord(t, got, makeRecord())
	got.Outcomes[0].Note = "changed"
	got.Answers.Plan["organization"] = text("changed")
	got.Answers.Sections["empty-scope"]["new"] = text("changed")
	got.UnknownFields["future"].(map[string]any)["emptyMap"].(map[string]any)["new"] = true
	requireRecord(t, before, makeRecord())
}

func TestMinimalContainerChanges(t *testing.T) {
	for _, allocated := range []bool{false, true} {
		before := model.Record{SchemaVersion: 1, Account: "newcomer"}
		if allocated {
			before.Answers.Sections = map[string]map[string]model.Value{}
		}
		intent := contract.Intent{Operation: contract.OperationFollowPlan, Question: "personal.choice"}
		got, problems := newWriter().Write(sharedPlan(), before, intent)
		requireNoProblems(t, problems)
		requireRecord(t, got, before)
		intent.Operation, intent.Value = contract.OperationSubmit, text("owned")
		got, problems = newWriter().Write(sharedPlan(), before, intent)
		requireNoProblems(t, problems)
		want := model.Record{SchemaVersion: 1, Account: "newcomer", Answers: model.Answers{
			Sections: map[string]map[string]model.Value{"personal": {"choice": text("owned")}},
		}}
		requireRecord(t, got, want)
		cleared, problems := newWriter().Write(sharedPlan(), got, contract.Intent{Operation: contract.OperationFollowPlan, Question: "personal.choice"})
		requireNoProblems(t, problems)
		want.Answers.Sections = map[string]map[string]model.Value{}
		requireRecord(t, cleared, want)
	}
	plan := sharedPlan()
	plan.Variables = []model.Variable{{Name: "choice", Type: model.VariableTypeString, Audience: model.AudienceOnboardee}}
	plan.Answers.Plan = map[string]model.Value{"choice": text("shared")}
	settled := true
	plan.Sections[0].IncludeIf = &model.Condition{All: []model.Clause{{Question: "plan.choice", Answer: model.AnswerTest{IsSettled: &settled}}}}
	before := model.Record{Answers: model.Answers{Plan: map[string]model.Value{"choice": text("owned")}}}
	got, problems := newWriter().Write(plan, before, contract.Intent{Operation: contract.OperationFollowPlan, Question: "plan.choice"})
	requireNoProblems(t, problems)
	requireRecord(t, got, model.Record{Answers: model.Answers{Plan: map[string]model.Value{}}})
}

func TestRefusalDiagnosticsAndInputsAreIndependent(t *testing.T) {
	writer := newWriter()
	plan, before := sharedPlan(), heldRecord()
	intent := contract.Intent{Operation: contract.OperationFollowPlan, Question: "missing", Value: text("not-empty")}
	got, first := writer.Write(plan, before, intent)
	requireRecord(t, got, model.Record{})
	want := []model.Problem{ownProblem("record.answer.follow-value", "missing", "Follow-plan must not carry an answer.", "intent.value")}
	if !reflect.DeepEqual(first, want) {
		t.Fatalf("got %#v, want %#v", first, want)
	}
	first[0].Detail["path"] = "changed"
	first[0].Message = "changed"
	_, second := writer.Write(plan, before, intent)
	if !reflect.DeepEqual(second, want) {
		t.Fatalf("later diagnostics aliased: %#v", second)
	}
	requireRecord(t, before, heldRecord())
	if !reflect.DeepEqual(plan, sharedPlan()) || !reflect.DeepEqual(intent.Value, text("not-empty")) {
		t.Fatal("refusal changed inputs")
	}
}
