package writer_test

import (
	"reflect"
	"testing"

	contract "onboard-kit/contracts/record/writer"
	"onboard-kit/internal/model"
)

func gapPlan() model.Plan {
	settled := true
	return model.Plan{
		SchemaVersion: 1,
		ID:            "plan-one",
		Variables: []model.Variable{
			{Name: "zulu", Type: model.VariableTypeString, Audience: model.AudienceOnboardee},
			{Name: "alpha", Type: model.VariableTypeString, Audience: model.AudienceOnboardee},
			{Name: "middle", Type: model.VariableTypeString, Audience: model.AudienceOnboardee},
		},
		Answers: model.Answers{
			Plan: map[string]model.Value{"alpha": text("plan-alpha"), "middle": text("plan-middle"), "zulu": text("plan-zulu")},
			Sections: map[string]map[string]model.Value{
				"aaa": {"alpha": text("aaa-alpha"), "middle": text("aaa-middle"), "zulu": text("aaa-zulu")},
				"zzz": {"alpha": text("zzz-alpha"), "middle": text("zzz-middle"), "zulu": text("zzz-zulu")},
			},
		},
		Sections: []model.Section{
			{
				ID: "zzz", Title: "Later exact scope", Summary: "Use $alpha, $middle and $zulu.",
				Variables: []model.Variable{
					{Name: "zulu", Type: model.VariableTypeString, Audience: model.AudienceOnboardee},
					{Name: "alpha", Type: model.VariableTypeString, Audience: model.AudienceOnboardee},
					{Name: "middle", Type: model.VariableTypeString, Audience: model.AudienceOnboardee},
				},
				IncludeIf: &model.Condition{All: []model.Clause{
					{Question: "plan.alpha", Answer: model.AnswerTest{IsSettled: &settled}},
					{Question: "plan.middle", Answer: model.AnswerTest{IsSettled: &settled}},
					{Question: "plan.zulu", Answer: model.AnswerTest{IsSettled: &settled}},
				}},
			},
			{
				ID: "aaa", Title: "Earlier exact scope", Summary: "Use $alpha, $middle and $zulu.",
				Variables: []model.Variable{
					{Name: "middle", Type: model.VariableTypeString, Audience: model.AudienceOnboardee},
					{Name: "zulu", Type: model.VariableTypeString, Audience: model.AudienceOnboardee},
					{Name: "alpha", Type: model.VariableTypeString, Audience: model.AudienceOnboardee},
				},
			},
		},
	}
}

func gapPlanRecord(shape string) model.Record {
	record := heldRecord()
	switch shape {
	case "nil":
		record.Answers.Plan = nil
	case "empty":
		record.Answers.Plan = map[string]model.Value{}
	case "nonempty":
		record.Answers.Plan = map[string]model.Value{
			"hidden_before": {Kind: model.ValueKindList, List: []string{}},
			"hidden_after":  {Kind: model.ValueKindFlag, Flag: false},
		}
	}
	return record
}

func TestPlanScopeSubmitAllocatesOnlyRequiredContainer(t *testing.T) {
	for _, shape := range []string{"nil", "empty", "nonempty"} {
		t.Run(shape, func(t *testing.T) {
			before := gapPlanRecord(shape)
			want := gapPlanRecord(shape)
			if shape == "nil" {
				want.Answers.Plan = map[string]model.Value{}
			}
			want.Answers.Plan["middle"] = text("explicit-plan-choice")
			got, problems := newWriter().Write(gapPlan(), before, contract.Intent{
				Operation: contract.OperationSubmit, Question: "plan.middle", Value: text("explicit-plan-choice"),
			})
			requireNoProblems(t, problems)
			requireRecord(t, got, want)
			requireRecord(t, before, gapPlanRecord(shape))
			if !reflect.DeepEqual(got.Answers.Sections, heldRecord().Answers.Sections) {
				t.Fatal("plan submit changed unrelated section containers")
			}
			got.Answers.Plan["middle"] = text("changed-result")
			requireRecord(t, before, gapPlanRecord(shape))
		})
	}
}

func TestExactScopeAndVariableAmongSeveralDeclarations(t *testing.T) {
	for _, scope := range []string{"plan", "aaa", "zzz"} {
		for _, name := range []string{"alpha", "middle", "zulu"} {
			t.Run(scope+"."+name, func(t *testing.T) {
				before, want := heldRecord(), heldRecord()
				if scope == "plan" {
					want.Answers.Plan[name] = text("owned-exact-target")
				} else {
					want.Answers.Sections[scope] = map[string]model.Value{name: text("owned-exact-target")}
				}
				got, problems := newWriter().Write(gapPlan(), before, contract.Intent{
					Operation: contract.OperationSubmit, Question: scope + "." + name, Value: text("owned-exact-target"),
				})
				requireNoProblems(t, problems)
				requireRecord(t, got, want)
				requireRecord(t, before, heldRecord())
			})
		}
	}
}

func TestFollowPlanPreservesBothLexicalScopesAndNonTargetKeys(t *testing.T) {
	for _, scope := range []string{"plan", "aaa", "zzz"} {
		t.Run(scope, func(t *testing.T) {
			before, want := heldRecord(), heldRecord()
			if scope == "plan" {
				before.Answers.Plan = map[string]model.Value{"alpha": text("keep-alpha"), "middle": text("remove-middle"), "zulu": text("keep-zulu")}
				want.Answers.Plan = map[string]model.Value{"alpha": text("keep-alpha"), "zulu": text("keep-zulu")}
			} else {
				before.Answers.Sections[scope] = map[string]model.Value{"alpha": text("keep-alpha"), "middle": text("remove-middle"), "zulu": text("keep-zulu")}
				want.Answers.Sections[scope] = map[string]model.Value{"alpha": text("keep-alpha"), "zulu": text("keep-zulu")}
			}
			got, problems := newWriter().Write(gapPlan(), before, contract.Intent{Operation: contract.OperationFollowPlan, Question: scope + ".middle"})
			requireNoProblems(t, problems)
			requireRecord(t, got, want)
			if scope == "plan" {
				if !reflect.DeepEqual(before.Answers.Plan["middle"], text("remove-middle")) {
					t.Fatal("follow-plan removed its original plan input")
				}
			} else if !reflect.DeepEqual(before.Answers.Sections[scope]["middle"], text("remove-middle")) {
				t.Fatal("follow-plan removed its original section input")
			}
		})
	}
}

func TestFollowPlanRequiresAllZeroValuePayload(t *testing.T) {
	for _, test := range []struct {
		name  string
		value model.Value
	}{
		{"inactive-text", model.Value{Text: "held-inactive-text"}},
		{"inactive-flag", model.Value{Flag: true}},
		{"allocated-empty-list", model.Value{List: []string{}}},
		{"inactive-nonempty-list", model.Value{List: []string{"held", "", "held"}}},
		{"deferred-kind", model.Value{Kind: model.ValueKindDeferred}},
		{"empty-text-kind", model.Value{Kind: model.ValueKindText}},
		{"false-flag-kind", model.Value{Kind: model.ValueKindFlag, Flag: false}},
		{"empty-list-kind", model.Value{Kind: model.ValueKindList}},
	} {
		t.Run(test.name, func(t *testing.T) {
			intent := contract.Intent{Operation: contract.OperationFollowPlan, Question: "missing", Value: test.value}
			got, problems := newWriter().Write(gapPlan(), heldRecord(), intent)
			requireRecord(t, got, model.Record{})
			want := []model.Problem{ownProblem("record.answer.follow-value", "missing", "Follow-plan must not carry an answer.", "intent.value")}
			if !reflect.DeepEqual(problems, want) {
				t.Fatalf("follow-plan value refusal %#v, want %#v", problems, want)
			}
		})
	}
	got, problems := newWriter().Write(gapPlan(), heldRecord(), contract.Intent{Operation: contract.OperationFollowPlan, Question: "plan.middle", Value: model.Value{}})
	requireNoProblems(t, problems)
	requireRecord(t, got, heldRecord())
}

func TestExplicitUnsupportedOperationsOnBothSidesOfAcceptedSpellings(t *testing.T) {
	for _, operation := range []contract.Operation{"", "answer", "follow-pla", "follow-plan-extra", "submi", "submit-extra", "zzzz"} {
		t.Run(string(operation), func(t *testing.T) {
			intent := contract.Intent{Operation: operation, Question: "plan.middle", Value: text("otherwise-valid")}
			got, problems := newWriter().Write(gapPlan(), heldRecord(), intent)
			requireRecord(t, got, model.Record{})
			want := []model.Problem{ownProblem("record.answer.operation", "plan.middle", "Unsupported answer operation.", "intent.operation")}
			if !reflect.DeepEqual(problems, want) {
				t.Fatalf("operation refusal %#v, want %#v", problems, want)
			}
		})
	}
}

func TestMissingExactVariableCannotUseLexicalNeighbor(t *testing.T) {
	for _, question := range []string{"plan.before", "plan.nobody", "plan.zzzz", "aaa.before", "aaa.nobody", "aaa.zzzz", "zzz.before", "zzz.nobody", "zzz.zzzz"} {
		t.Run(question, func(t *testing.T) {
			got, problems := newWriter().Write(gapPlan(), heldRecord(), contract.Intent{Operation: contract.OperationSubmit, Question: question, Value: text("otherwise-valid")})
			requireRecord(t, got, model.Record{})
			want := []model.Problem{ownProblem("record.answer.question", question, "Choose one declared newcomer question.", "intent.question")}
			if !reflect.DeepEqual(problems, want) {
				t.Fatalf("missing-name refusal %#v, want %#v", problems, want)
			}
		})
	}
}

func TestPresentSectionMapKeepsOtherAnswersOnSubmit(t *testing.T) {
	for _, scope := range []string{"aaa", "zzz"} {
		t.Run(scope, func(t *testing.T) {
			before, want := heldRecord(), heldRecord()
			before.Answers.Sections[scope] = map[string]model.Value{
				"alpha": text("keep-alpha"), "middle": text("old-middle"), "zulu": text("keep-zulu"),
				"hidden_empty": {Kind: model.ValueKindList, List: []string{}},
			}
			want.Answers.Sections[scope] = map[string]model.Value{
				"alpha": text("keep-alpha"), "middle": text("new-middle"), "zulu": text("keep-zulu"),
				"hidden_empty": {Kind: model.ValueKindList, List: []string{}},
			}
			got, problems := newWriter().Write(gapPlan(), before, contract.Intent{Operation: contract.OperationSubmit, Question: scope + ".middle", Value: text("new-middle")})
			requireNoProblems(t, problems)
			requireRecord(t, got, want)
			if !reflect.DeepEqual(before.Answers.Sections[scope]["middle"], text("old-middle")) {
				t.Fatal("submit changed the original target")
			}
		})
	}
}
